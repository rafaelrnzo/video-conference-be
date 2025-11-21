package service

import (
	"context"
	"time"

	"video-conference-be/pkg/utility"

	livekit "github.com/livekit/protocol/livekit"
)

type LivekitService interface {
	GenerateUserToken(ctx context.Context, room, identity string) (string, string, error)
	CreateRoom(ctx context.Context, req *livekit.CreateRoomRequest) (*livekit.Room, error)
	ListRooms(ctx context.Context) ([]*livekit.Room, error)
	DeleteRoom(ctx context.Context, name string) error
	ListParticipants(ctx context.Context, room string) ([]*livekit.ParticipantInfo, error)
	RemoveParticipant(ctx context.Context, room, identity string) error
}

type livekitService struct {
	client *utility.LivekitClient
}

func NewLivekitService(client *utility.LivekitClient) LivekitService {
	return &livekitService{client: client}
}

func (s *livekitService) GenerateUserToken(ctx context.Context, room, identity string) (string, string, error) {
	// ctx belum kepakai di client, tapi pattern-nya tetap dibawa biar konsisten
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
