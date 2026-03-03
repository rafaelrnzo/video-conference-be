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
	DeleteRecording(ctx context.Context, id uint) error
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
		"COMPLETED",
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
	prefix := "" // Scan everything inside bucket

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

		if strings.HasSuffix(object.Key, "/") {
			continue
		}

		isMP4 := strings.HasSuffix(object.Key, ".mp4")
		isHLS := strings.HasSuffix(object.Key, "index.m3u8")

		if !isMP4 && !isHLS {
			continue
		}

		base := strings.TrimRight(utility.Config.MinioEndpoint, "/")
		bucketName := strings.Trim(utility.Config.MinioBucket, "/")
		link := fmt.Sprintf("%s/%s/%s", base, bucketName, object.Key)

		exists, err := s.recordSvc.Exists(ctx, link)
		if err != nil {
			log.Printf("[SYNC] error checking existence for %s: %v", link, err)
			continue
		}

		if exists {
			continue
		}

		// Try to parse room name from key
		parts := strings.Split(object.Key, "/")
		roomName := "Unknown"
		
		// If it's something like recordings/roomName/timestamp/index.m3u8
		if len(parts) >= 2 && parts[0] == "recordings" {
			roomName = parts[1]
		} else if isMP4 {
			// If it's something like roomName-timestamp.mp4 from recorder
			filename := object.Key
			if len(parts) > 0 {
				filename = parts[len(parts)-1]
			}
			
			// Try to extract roomName by finding the last dash before .mp4
			// e.g. "myroom-123456789.mp4" -> "myroom"
			nameParts := strings.Split(strings.TrimSuffix(filename, ".mp4"), "-")
			if len(nameParts) > 1 {
				// The last part is likely the timestamp, join the rest for roomName
				roomName = strings.Join(nameParts[:len(nameParts)-1], "-")
			} else {
				roomName = strings.TrimSuffix(filename, ".mp4")
			}
		}

		name := fmt.Sprintf("%s Recording", roomName)
		if isHLS {
			name = fmt.Sprintf("%s (HLS)", roomName)
		}

		egressID := object.Key // Unique identifier since it's the filename itself

		_, err = s.recordSvc.Create(ctx, roomName, name, link, egressID, "COMPLETED")
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

func (s *recordingService) DeleteRecording(ctx context.Context, id uint) error {
	rec, err := s.recordSvc.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get record: %w", err)
	}

	if s.minioClient != nil {
		bucketName := strings.Trim(utility.Config.MinioBucket, "/")
		
		// Determine the object key from the link or egress ID
		// The link is usually http://endpoint/bucket/object-key
		// Let's try to extract object key from link
		objectKey := rec.Link
		parts := strings.SplitN(rec.Link, bucketName+"/", 2)
		if len(parts) == 2 {
			objectKey = parts[1]
		} else {
			// Fallback: use egressID if it looks like a path
			if strings.Contains(rec.EgressID, "/") || strings.HasSuffix(rec.EgressID, ".mp4") {
				objectKey = rec.EgressID
			}
		}

		if objectKey != "" && objectKey != rec.Link {
			isHLS := strings.HasSuffix(objectKey, "index.m3u8")
			if isHLS {
				// We need to delete the entire directory
				dirPrefix := strings.TrimSuffix(objectKey, "index.m3u8")
				objectsCh := s.minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
					Prefix:    dirPrefix,
					Recursive: true,
				})
				for object := range objectsCh {
					if object.Err == nil {
						if err := s.minioClient.RemoveObject(ctx, bucketName, object.Key, minio.RemoveObjectOptions{}); err != nil {
							log.Printf("[DELETE RECORDING] inline failed to delete object %s: %v", object.Key, err)
						}
					}
				}
				log.Printf("[DELETE RECORDING] deleted HLS directory: %s", dirPrefix)
			} else {
				if err := s.minioClient.RemoveObject(ctx, bucketName, objectKey, minio.RemoveObjectOptions{}); err != nil {
					log.Printf("[DELETE RECORDING] failed to delete object %s: %v", objectKey, err)
				} else {
					log.Printf("[DELETE RECORDING] deleted object: %s", objectKey)
				}
			}
		}
	}

	return s.recordSvc.Delete(ctx, id)
}
