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
)

type RecordingService interface {
	StartRoomRecording(ctx context.Context, roomName string) (*livekit.EgressInfo, error)
	StopRoomRecording(ctx context.Context, roomName string) (*livekit.EgressInfo, error)
}

type recordingService struct {
	egressClient *lksdk.EgressClient
	recordSvc    RecordService

	mu           sync.Mutex
	activeByRoom map[string]string // roomName -> egressID
}

func NewRecordingService(recordSvc RecordService) RecordingService {
	c := utility.Config

	client := lksdk.NewEgressClient(
		c.LivekitURL,
		c.LivekitAPIKey,
		c.LivekitSecret,
	)

	return &recordingService{
		egressClient: client,
		recordSvc:    recordSvc,
		activeByRoom: make(map[string]string),
	}
}

func (s *recordingService) StartRoomRecording(ctx context.Context, roomName string) (*livekit.EgressInfo, error) {
	// filename prefix: recordings/{room}/{timestamp}
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

	// hanya kalau COMPLETE baru kita catat ke DB
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
