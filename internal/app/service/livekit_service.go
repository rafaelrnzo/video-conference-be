package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"video-conference-be/pkg/utility"

	livekit "github.com/livekit/protocol/livekit"
	"github.com/redis/go-redis/v9"
)

type LivekitService interface {
	GenerateUserToken(ctx context.Context, room, identity string) (string, string, error)
	CreateRoom(ctx context.Context, req *livekit.CreateRoomRequest) (*livekit.Room, error)
	ListRooms(ctx context.Context) ([]*livekit.Room, error)
	DeleteRoom(ctx context.Context, name string) error
	ListParticipants(ctx context.Context, room string) ([]*livekit.ParticipantInfo, error)
	RemoveParticipant(ctx context.Context, room, identity string) error
	StartRoomRecording(ctx context.Context, roomName, filenamePrefix string) (*livekit.EgressInfo, error)
	StopRoomRecording(ctx context.Context, roomName string) (*livekit.EgressInfo, error)
	SetUserOnline(ctx context.Context, userID uint, identity string, room string, ttl time.Duration) error
	SetUserOffline(ctx context.Context, identity string) error
	IsUserOnline(ctx context.Context, identity string) (bool, string, error)
	UpdateRoomMetadata(ctx context.Context, room string, metadata string) error
	MuteAllParticipants(ctx context.Context, room string, muteAudio, muteVideo bool) error
}

type livekitService struct {
	client *utility.LivekitClient
}

func NewLivekitService(client *utility.LivekitClient) LivekitService {
	return &livekitService{client: client}
}

func (s *livekitService) GenerateUserToken(ctx context.Context, room, identity string) (string, string, error) {
	return s.client.GenerateToken(room, identity, 2*time.Hour)
}

func (s *livekitService) CreateRoom(ctx context.Context, req *livekit.CreateRoomRequest) (*livekit.Room, error) {
	return s.client.CreateRoom(ctx, req)
}

func (s *livekitService) ListRooms(ctx context.Context) ([]*livekit.Room, error) {
	return s.client.ListRooms(ctx)
}

func (s *livekitService) DeleteRoom(ctx context.Context, name string) error {
	return s.client.DeleteRoom(ctx, name)
}

func (s *livekitService) ListParticipants(ctx context.Context, room string) ([]*livekit.ParticipantInfo, error) {
	return s.client.ListParticipants(ctx, room)
}

func (s *livekitService) RemoveParticipant(ctx context.Context, room, identity string) error {
	return s.client.RemoveParticipant(ctx, room, identity)
}

func (s *livekitService) StartRoomRecording(ctx context.Context, roomName, filenamePrefix string) (*livekit.EgressInfo, error) {
	return s.client.StartRoomRecording(ctx, roomName, filenamePrefix)
}

func (s *livekitService) StopRoomRecording(ctx context.Context, roomName string) (*livekit.EgressInfo, error) {
	return s.client.StopRoomRecording(ctx, roomName)
}

func (s *livekitService) SetUserOnline(ctx context.Context, userID uint, identity string, room string, ttl time.Duration) error {
	key := fmt.Sprintf("user_presence:%s", identity)
	value := fmt.Sprintf("%d:%s:%s", userID, room, identity)
	return utility.RedisClient.Set(ctx, key, value, ttl).Err()
}

func (s *livekitService) SetUserOffline(ctx context.Context, identity string) error {
	key := fmt.Sprintf("user_presence:%s", identity)
	return utility.RedisClient.Del(ctx, key).Err()
}

func (s *livekitService) IsUserOnline(ctx context.Context, identity string) (bool, string, error) {
	key := fmt.Sprintf("user_presence:%s", identity)
	val, err := utility.RedisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}

	parts := strings.Split(val, ":")
	if len(parts) >= 2 {
		return true, parts[1], nil
	}

	return true, val, nil
}

func (s *livekitService) UpdateRoomMetadata(ctx context.Context, room string, metadata string) error {
	return s.client.UpdateRoomMetadata(ctx, room, metadata)
}

func (s *livekitService) MuteAllParticipants(ctx context.Context, room string, muteAudio, muteVideo bool) error {
	participants, err := s.client.ListParticipants(ctx, room)
	if err != nil {
		return err
	}

	for _, p := range participants {
		for _, t := range p.Tracks {
			if muteAudio && t.Type == livekit.TrackType_AUDIO {
				_ = s.client.MutePublishedTrack(ctx, room, p.Identity, t.Sid, true)
			}
			if muteVideo && t.Type == livekit.TrackType_VIDEO {
				_ = s.client.MutePublishedTrack(ctx, room, p.Identity, t.Sid, true)
			}
		}
	}
	return nil
}