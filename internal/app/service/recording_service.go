package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"video-conference-be/pkg/utility"

	livekit "github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type RecordingService interface {
	StartRoomRecording(ctx context.Context, roomName string) (*livekit.EgressInfo, error)
	StopRoomRecording(ctx context.Context, roomName string) (*livekit.EgressInfo, error)
	SyncFromMinio(ctx context.Context) error
}

type recordingService struct {
	egressClient *lksdk.EgressClient
	minioClient  *minio.Client
	recordSvc    RecordService

	mu           sync.Mutex
	activeByRoom map[string]string
}

func NewRecordingService(recordSvc RecordService) RecordingService {
	c := utility.Config

	client := lksdk.NewEgressClient(
		c.LivekitURL,
		c.LivekitAPIKey,
		c.LivekitSecret,
	)

	// Initialize MinIO client
	minioEndpoint := strings.TrimPrefix(c.MinioEndpoint, "http://")
	minioEndpoint = strings.TrimPrefix(minioEndpoint, "https://")
	
	mClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.MinioAccess, c.MinioSecret, ""),
		Secure: strings.HasPrefix(c.MinioEndpoint, "https"),
	})
	if err != nil {
		log.Printf("[WARNING] Failed to init MinIO client: %v", err)
	}

	return &recordingService{
		egressClient: client,
		minioClient:  mClient,
		recordSvc:    recordSvc,
		activeByRoom: make(map[string]string),
	}
}

func (s *recordingService) StartRoomRecording(ctx context.Context, roomName string) (*livekit.EgressInfo, error) {
	ts := time.Now().Format("20060102-150405")
	prefix := fmt.Sprintf("recordings/%s/%s", roomName, ts)

	req := &livekit.RoomCompositeEgressRequest{
		RoomName: roomName,
		Layout:   "grid",
		Options: &livekit.RoomCompositeEgressRequest_Preset{
			Preset: livekit.EncodingOptionsPreset_H264_720P_30,
		},
	}
	req.SegmentOutputs = []*livekit.SegmentedFileOutput{
		{
			FilenamePrefix: prefix,
			PlaylistName:   "index.m3u8",
		},
	}

	info, err := s.egressClient.StartRoomCompositeEgress(ctx, req)
	if err != nil {
		return nil, err
	}

	log.Printf("[EGRESS START] id=%s room=%s status=%s error=%s",
		info.EgressId, info.RoomName, info.Status.String(), info.Error)

	s.mu.Lock()
	s.activeByRoom[roomName] = info.EgressId
	s.mu.Unlock()

	return info, nil
}

func (s *recordingService) StopRoomRecording(ctx context.Context, roomName string) (*livekit.EgressInfo, error) {
	s.mu.Lock()
	egressID, ok := s.activeByRoom[roomName]
	s.mu.Unlock()

	if !ok || egressID == "" {
		return nil, fmt.Errorf("no active egress for room %s", roomName)
	}

	info, err := s.egressClient.StopEgress(ctx, &livekit.StopEgressRequest{
		EgressId: egressID,
	})
	if err != nil {
		return nil, err
	}

	log.Printf("[EGRESS STOP] id=%s room=%s status=%s error=%s",
		info.EgressId, info.RoomName, info.Status.String(), info.Error)

	if info.Status == livekit.EgressStatus_EGRESS_COMPLETE {
		s.handleEgressComplete(ctx, info)
	}

	return info, nil
}

func (s *recordingService) handleEgressComplete(ctx context.Context, info *livekit.EgressInfo) {
	if info == nil {
		return
	}

	var filePath string

	if len(info.SegmentResults) > 0 {
		filePath = strings.TrimSpace(info.SegmentResults[0].PlaylistName)
	} else if len(info.FileResults) > 0 {
		filePath = strings.TrimSpace(info.FileResults[0].Filename)
	}

	if filePath == "" {
		log.Printf("[EGRESS COMPLETE] no file results for egress %s", info.EgressId)
		return
	}

	filePath = strings.TrimLeft(filePath, "/")

	base := strings.TrimRight(utility.Config.MinioEndpoint, "/")
	bucket := strings.Trim(utility.Config.MinioBucket, "/")

	if base == "" || bucket == "" {
		log.Printf("[EGRESS COMPLETE] MinioEndpoint/Bucket kosong, skip save record")
		return
	}

	link := fmt.Sprintf("%s/%s/%s", base, bucket, filePath)

	defaultName := fmt.Sprintf("%s - %s", info.RoomName, time.Now().Format("2006-01-02 15:04:05"))

	_, err := s.recordSvc.Create(
		ctx,
		info.RoomName,
		defaultName,
		link,
		info.EgressId,
	)
	if err != nil {
		log.Printf("[EGRESS COMPLETE] failed to save record: %v", err)
		return
	}

	log.Printf("[EGRESS COMPLETE] record saved: room=%s egress=%s path=%s link=%s",
		info.RoomName, info.EgressId, filePath, link)
}

func (s *recordingService) SyncFromMinio(ctx context.Context) error {
	if s.minioClient == nil {
		return fmt.Errorf("minio client not initialized")
	}

	bucket := utility.Config.MinioBucket
	prefix := "recordings/"

	// List objects
	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}

	objectCh := s.minioClient.ListObjects(ctx, bucket, opts)
	count := 0

	for object := range objectCh {
		if object.Err != nil {
			log.Printf("[SYNC] error listing object: %v", object.Err)
			continue
		}

		// Skip directories (if any) or non-relevant files
		if strings.HasSuffix(object.Key, "/") || (!strings.HasSuffix(object.Key, ".mp4") && !strings.HasSuffix(object.Key, ".m3u8")) {
			continue
		}
		// If it's HLS, we usually main playlist index.m3u8.
		// If our recording strategy produces playlists for segments, we might want to only sync the master playlist.
		// Based on StartRoomRecording: FilenamePrefix: prefix, PlaylistName: "index.m3u8"
		// The path is recordings/{roomName}/{timestamp}/index.m3u8
		// If flat file: recordings/{roomName}/{timestamp}.mp4 ? (need to check standard output)
		// For HLS, we want the index.m3u8.
		
		isHLS := strings.HasSuffix(object.Key, "index.m3u8")
		isMP4 := strings.HasSuffix(object.Key, ".mp4")

		if !isHLS && !isMP4 {
			continue
		}

		// Construct Link
		// Config.MinioEndpoint might imply external URL, but here we construct what we store in DB.
		// existing logic: link := fmt.Sprintf("%s/%s/%s", base, bucket, filePath)
		base := strings.TrimRight(utility.Config.MinioEndpoint, "/")
		bucketName := strings.Trim(utility.Config.MinioBucket, "/")
		link := fmt.Sprintf("%s/%s/%s", base, bucketName, object.Key)

		// Check if exists
		exists, err := s.recordSvc.Exists(ctx, link)
		if err != nil {
			log.Printf("[SYNC] error checking existence for %s: %v", link, err)
			continue
		}

		if exists {
			continue
		}

		// Create record
		// Try to parse room name from key
		// key format expected: recordings/roomName/timestamp/index.m3u8 OR recordings/roomName-timestamp.mp4 ??
		// Let's assume standard "recordings/roomName/..."
		parts := strings.Split(object.Key, "/")
		roomName := "unknown"
		if len(parts) >= 2 {
			// parts[0] is "recordings"
			roomName = parts[1]
		}

		name := fmt.Sprintf("Imported - %s", roomName)
		if isHLS {
			// try to make specific name
			name = fmt.Sprintf("%s (HLS)", roomName)
		} else {
			name = fmt.Sprintf("%s (MP4)", roomName)
		}
		// Add timestamp if available
		if len(parts) >= 3 {
             // maybe parts[2] is timestamp
             name = fmt.Sprintf("%s - %s", name, parts[2])
        }

		// EgressID is unknown for manual sync usually
		egressID := "sync-" + fmt.Sprintf("%d", time.Now().UnixNano())

		_, err = s.recordSvc.Create(ctx, roomName, name, link, egressID)
		if err != nil {
			log.Printf("[SYNC] failed to create record for %s: %v", link, err)
		} else {
			log.Printf("[SYNC] imported record: %s", link)
			count++
		}
	}

	log.Printf("[SYNC] finished. imported %d new recordings", count)
	return nil
}
