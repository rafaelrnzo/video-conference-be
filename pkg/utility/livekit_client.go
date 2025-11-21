package utility

import (
	"context"
	"strings"
	"time"

	"github.com/livekit/protocol/auth"
	livekit "github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

type LivekitClient struct {
	apiKey    string
	apiSecret string
	serverURL string
	svc       *lksdk.RoomServiceClient
}

func NewLivekitClient() *LivekitClient {
	c := &LivekitClient{
		apiKey:    Config.LivekitAPIKey,
		apiSecret: Config.LivekitSecret,
		serverURL: Config.LivekitURL,
	}

	// Sama seperti main.go lama:
	// svc := lksdk.NewRoomServiceClient(cfg.ServerURL, cfg.APIKey, cfg.APISecret)
	c.svc = lksdk.NewRoomServiceClient(c.serverURL, c.apiKey, c.apiSecret)

	return c
}

func (c *LivekitClient) normalizeWS(u string) string {
	if u == "" {
		return u
	}
	trimmed := strings.TrimRight(u, "/")
	if strings.HasPrefix(trimmed, "http://") {
		return "ws://" + strings.TrimPrefix(trimmed, "http://")
	}
	if strings.HasPrefix(trimmed, "https://") {
		return "wss://" + strings.TrimPrefix(trimmed, "https://")
	}
	return trimmed
}

// GenerateToken – persis seperti generateToken() di main.go lama
func (c *LivekitClient) GenerateToken(roomName, identity string, ttl time.Duration) (string, string, error) {
	at := auth.NewAccessToken(c.apiKey, c.apiSecret)

	canPub := true
	canSub := true
	canData := true

	grant := &auth.VideoGrant{
		RoomJoin:       true,
		Room:           roomName,
		CanPublish:     &canPub,
		CanSubscribe:   &canSub,
		CanPublishData: &canData,
	}

	at.SetVideoGrant(grant).
		SetIdentity(identity).
		SetValidFor(ttl)

	jwt, err := at.ToJWT()
	if err != nil {
		return "", "", err
	}

	return jwt, c.normalizeWS(c.serverURL), nil
}

// ==== WRAPPER ROOM SERVICE (sama pola dengan kode lama) ====

// ctx diambil dari handler Gin (c.Request.Context())
func (c *LivekitClient) CreateRoom(ctx context.Context, req *livekit.CreateRoomRequest) (*livekit.Room, error) {
	return c.svc.CreateRoom(ctx, req)
}

func (c *LivekitClient) ListRooms(ctx context.Context) ([]*livekit.Room, error) {
	resp, err := c.svc.ListRooms(ctx, &livekit.ListRoomsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Rooms, nil
}

func (c *LivekitClient) DeleteRoom(ctx context.Context, name string) error {
	_, err := c.svc.DeleteRoom(ctx, &livekit.DeleteRoomRequest{Room: name})
	return err
}

func (c *LivekitClient) ListParticipants(ctx context.Context, room string) ([]*livekit.ParticipantInfo, error) {
	resp, err := c.svc.ListParticipants(ctx, &livekit.ListParticipantsRequest{Room: room})
	if err != nil {
		return nil, err
	}
	return resp.Participants, nil
}

func (c *LivekitClient) RemoveParticipant(ctx context.Context, room, identity string) error {
	_, err := c.svc.RemoveParticipant(ctx, &livekit.RoomParticipantIdentity{
		Room:     room,
		Identity: identity,
	})
	return err
}
