package utility

import (
	"context"
	"fmt"
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

	svc    *lksdk.RoomServiceClient
	egress *lksdk.EgressClient
}

func NewLivekitClient() *LivekitClient {
	c := &LivekitClient{
		apiKey:    Config.LivekitAPIKey,
		apiSecret: Config.LivekitSecret,
		serverURL: Config.LivekitURL,
	}

	c.svc = lksdk.NewRoomServiceClient(c.serverURL, c.apiKey, c.apiSecret)
	c.egress = lksdk.NewEgressClient(c.serverURL, c.apiKey, c.apiSecret)

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

func (c *LivekitClient) StartRoomRecording(ctx context.Context, roomName, filenamePrefix string) (*livekit.EgressInfo, error) {
	if filenamePrefix == "" {
		filenamePrefix = fmt.Sprintf("%s/%s", roomName, roomName)
	}

	req := &livekit.RoomCompositeEgressRequest{
		RoomName: roomName,
		Layout:   "grid",
		Options: &livekit.RoomCompositeEgressRequest_Preset{
			Preset: livekit.EncodingOptionsPreset_H264_720P_30,
		},
	}

	req.SegmentOutputs = []*livekit.SegmentedFileOutput{
		{
			FilenamePrefix:   filenamePrefix,
			PlaylistName:     "index.m3u8",
			LivePlaylistName: "",
			SegmentDuration:  2,
			Output: &livekit.SegmentedFileOutput_S3{
				S3: &livekit.S3Upload{
					AccessKey:      Config.MinioAccess,
					Secret:         Config.MinioSecret,
					Endpoint:       Config.MinioEndpoint,
					Bucket:         Config.MinioBucket,
					Region:         Config.MinioRegion,
					ForcePathStyle: true,
				},
			},
		},
	}

	info, err := c.egress.StartRoomCompositeEgress(ctx, req)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (c *LivekitClient) StopRoomRecording(ctx context.Context, roomName string) (*livekit.EgressInfo, error) {
	listResp, err := c.egress.ListEgress(ctx, &livekit.ListEgressRequest{
		RoomName: roomName,
	})
	if err != nil {
		return nil, err
	}

	if len(listResp.Items) == 0 {
		return nil, fmt.Errorf("no egress found for room %s", roomName)
	}

	var target *livekit.EgressInfo
	for _, info := range listResp.Items {
		if info.Status == livekit.EgressStatus_EGRESS_ACTIVE ||
			info.Status == livekit.EgressStatus_EGRESS_STARTING {
			target = info
			break
		}
	}
	if target == nil {
		target = listResp.Items[len(listResp.Items)-1]
	}

	if target.Status == livekit.EgressStatus_EGRESS_ABORTED ||
		target.Status == livekit.EgressStatus_EGRESS_COMPLETE ||
		target.Status == livekit.EgressStatus_EGRESS_FAILED {
		return target, nil
	}

	info, err := c.egress.StopEgress(ctx, &livekit.StopEgressRequest{
		EgressId: target.EgressId,
	})
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "EGRESS_ABORTED") ||
			strings.Contains(msg, "EGRESS_COMPLETE") ||
			strings.Contains(strings.ToLower(msg), "cannot be stopped") {
			return target, nil
		}
		return nil, err
	}

	return info, nil
}

func (c *LivekitClient) UpdateRoomMetadata(ctx context.Context, room, metadata string) error {
	_, err := c.svc.UpdateRoomMetadata(ctx, &livekit.UpdateRoomMetadataRequest{
		Room:     room,
		Metadata: metadata,
	})
	return err
}

func (c *LivekitClient) MutePublishedTrack(ctx context.Context, room, identity, trackSid string, muted bool) error {
	_, err := c.svc.MutePublishedTrack(ctx, &livekit.MuteRoomTrackRequest{
		Room:     room,
		Identity: identity,
		TrackSid: trackSid,
		Muted:    muted,
	})
	return err
}
