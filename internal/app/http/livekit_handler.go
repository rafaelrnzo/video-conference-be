package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"video-conference-be/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/livekit/protocol/livekit"
)

type LivekitHandler struct {
	lk       service.LivekitService
	roomSvc  service.RoomService
	groupSvc service.GroupService
}

func NewLivekitHandler(lk service.LivekitService, roomSvc service.RoomService, groupSvc service.GroupService) *LivekitHandler {
	return &LivekitHandler{
		lk:       lk,
		roomSvc:  roomSvc,
		groupSvc: groupSvc,
	}
}

func stringInSlice(s string, list []string) bool {
	for _, v := range list {
		if s == v {
			return true
		}
	}
	return false
}

func respondError(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": msg,
	})
}

func logAndRespondError(c *gin.Context, status int, msg string, err error) {
	if err != nil {
		_ = c.Error(err)
	}
	respondError(c, status, msg)
}

func (h *LivekitHandler) Health(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (h *LivekitHandler) GenerateToken(c *gin.Context) {
	var body struct {
		RoomCode string `json:"room_code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.RoomCode == "" {
		logAndRespondError(c, http.StatusBadRequest, "room_code is required", err)
		return
	}

	dbRoom, err := h.roomSvc.GetRoomByCode(body.RoomCode)
	if err != nil {
		logAndRespondError(c, http.StatusNotFound, "room not found or invalid code", err)
		return
	}

	identity := c.GetString("username")
	
	if len(dbRoom.BannedUsers) > 0 && stringInSlice(identity, dbRoom.BannedUsers) {
		respondError(c, http.StatusForbidden, "you banned from this room")
		return
	}

	now := time.Now()
	if now.Before(dbRoom.StartDate) {
		msg := fmt.Sprintf("Meeting hasn't started yet. Starts at: %s", dbRoom.StartDate.Format(time.RFC1123))
		respondError(c, http.StatusForbidden, msg)
		return
	}
	if now.After(dbRoom.EndDate) {
		msg := fmt.Sprintf("Meeting has ended at: %s", dbRoom.EndDate.Format(time.RFC1123))
		respondError(c, http.StatusForbidden, msg)
		return
	}

	userID := c.GetUint("user_id")

	isOnline, onlineRoom, err := h.lk.IsUserOnline(c.Request.Context(), identity)
	if err != nil {
		logAndRespondError(c, http.StatusInternalServerError, "failed to check user presence", err)
		return
	}
	if isOnline {
		if onlineRoom != dbRoom.RoomCode {
			respondError(c, http.StatusConflict, fmt.Sprintf("user is currently in another room: %s", onlineRoom))
			return
		}
	}

	isAllowed := false

	if len(dbRoom.AssignedTo) > 0 {
		if stringInSlice(identity, dbRoom.AssignedTo) {
			isAllowed = true
		}
	} else {
		if dbRoom.GroupID == nil || *dbRoom.GroupID == 0 {
			isAllowed = true
		}
	}

	if !isAllowed && dbRoom.GroupID != nil && *dbRoom.GroupID > 0 {
		isMember, err := h.groupSvc.IsMember(c.Request.Context(), *dbRoom.GroupID, userID)
		if err != nil {
			logAndRespondError(c, http.StatusInternalServerError, "failed to check group membership", err)
			return
		}
		if isMember {
			isAllowed = true
		}
	}

	if !isAllowed {
		respondError(c, http.StatusForbidden, "you do not have access to this room (group/assigned restricted)")
		return
	}

	isWaiting := false

	role := c.GetString("role")
	isAdmin := role == "admin"
	isCreator := dbRoom.CreatedByID == userID


	liveRooms, err := h.lk.ListRooms(c.Request.Context())
	waitingRoomEnabled := true

	if err == nil {
		for _, r := range liveRooms {
			if r.Name == dbRoom.RoomCode {
				if r.Metadata != "" {
					var metaMap map[string]interface{}
					if json.Unmarshal([]byte(r.Metadata), &metaMap) == nil {
						if val, ok := metaMap["waiting_room_enabled"]; ok {
							if boolVal, ok := val.(bool); ok {
								waitingRoomEnabled = boolVal
							}
						}
					}
				}
				break
			}
		}
	}

	if !isAdmin && !isCreator {
		if (dbRoom.GroupID == nil || *dbRoom.GroupID == 0) && len(dbRoom.AssignedTo) == 0 {
			if waitingRoomEnabled {
				isWaiting = true
			}
		}
	}

	token, host, err := h.lk.GenerateUserToken(c.Request.Context(), dbRoom.RoomCode, identity, isWaiting)
	if err != nil {
		logAndRespondError(c, http.StatusInternalServerError, "failed to generate token", err)
		return
	}

	_ = h.lk.SetUserOnline(c.Request.Context(), userID, identity, dbRoom.RoomCode, 2*time.Minute)

	c.JSON(http.StatusOK, gin.H{
		"identity":   identity,
		"room":       dbRoom.RoomCode,
		"room_name":  dbRoom.Name,
		"token":      token,
		"host":       host,
		"is_waiting": isWaiting,
	})
}

func (h *LivekitHandler) ListActiveRooms(c *gin.Context) {
	rooms, err := h.lk.ListRooms(c.Request.Context())
	if err != nil {
		logAndRespondError(c, http.StatusInternalServerError, "failed to list active rooms", err)
		return
	}

	if rooms == nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	c.JSON(http.StatusOK, rooms)
}

func (h *LivekitHandler) DeleteActiveRoom(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		respondError(c, http.StatusBadRequest, "name is required")
		return
	}

	if err := h.lk.DeleteRoom(c.Request.Context(), name); err != nil {
		logAndRespondError(c, http.StatusInternalServerError, "failed to delete room", err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *LivekitHandler) ListParticipants(c *gin.Context) {
	room := c.Query("room")
	if room == "" {
		respondError(c, http.StatusBadRequest, "room is required")
		return
	}

	participants, err := h.lk.ListParticipants(c.Request.Context(), room)
	if err != nil {
		logAndRespondError(c, http.StatusInternalServerError, "failed to list participants", err)
		return
	}

	c.JSON(http.StatusOK, participants)
}

func (h *LivekitHandler) RemoveParticipant(c *gin.Context) {
	room := c.Query("room")
	identity := c.Query("identity")
	if room == "" || identity == "" {
		respondError(c, http.StatusBadRequest, "room and identity are required")
		return
	}

	if err := h.lk.RemoveParticipant(c.Request.Context(), room, identity); err != nil {
		logAndRespondError(c, http.StatusInternalServerError, "failed to remove participant", err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *LivekitHandler) Webhook(c *gin.Context) {
	var event struct {
		Event       string                   `json:"event"`
		Room        *livekit.Room            `json:"room"`
		Participant *livekit.ParticipantInfo `json:"participant"`
	}

	if err := c.ShouldBindJSON(&event); err != nil {
		fmt.Printf("webhook error: %v\n", err)
		c.String(http.StatusOK, "ok")
		return
	}

	fmt.Printf("LiveKit Webhook: %s, Room: %s, Identity: %s\n", event.Event, event.Room.Name, event.Participant.Identity)

	ctx := c.Request.Context()

	if event.Event == "participant_joined" {
		if err := h.lk.SetUserOnline(ctx, 0, event.Participant.Identity, event.Room.Name, 24*time.Hour); err != nil {
			fmt.Printf("failed to set user online: %v\n", err)
		}
	} else if event.Event == "participant_left" {
		if err := h.lk.SetUserOffline(ctx, event.Participant.Identity); err != nil {
			fmt.Printf("failed to set user offline: %v\n", err)
		}
	}

	c.String(http.StatusOK, "ok")
}

func (h *LivekitHandler) LeaveRoom(c *gin.Context) {
	identity := c.GetString("username")
	if identity == "" {
		respondError(c, http.StatusUnauthorized, "invalid identity")
		return
	}

	if err := h.lk.SetUserOffline(c.Request.Context(), identity); err != nil {
		logAndRespondError(c, http.StatusInternalServerError, "failed to set user offline", err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *LivekitHandler) KickParticipant(c *gin.Context) {
	var body struct {
		RoomCode string `json:"room_code"`
		Identity string `json:"identity"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.RoomCode == "" || body.Identity == "" {
		respondError(c, http.StatusBadRequest, "room_code and identity are required")
		return
	}

	userID := c.GetUint("user_id")
	role := c.GetString("role")

	room, err := h.roomSvc.GetRoomByCode(body.RoomCode)
	if err != nil {
		logAndRespondError(c, http.StatusNotFound, "room not found", err)
		return
	}

	isCreator := room.CreatedByID == userID
	isAdmin := role == "admin"

	if !isCreator && !isAdmin {
		respondError(c, http.StatusForbidden, "you are not authorized to kick participants")
		return
	}

	if err := h.lk.RemoveParticipant(c.Request.Context(), body.RoomCode, body.Identity); err != nil {
		logAndRespondError(c, http.StatusInternalServerError, "failed to kick participant", err)
		return
	}

	_ = h.lk.SetUserOffline(c.Request.Context(), body.Identity)

	c.Status(http.StatusNoContent)
}

func (h *LivekitHandler) MuteAll(c *gin.Context) {
	var body struct {
		RoomCode  string `json:"room_code"`
		MuteAudio bool   `json:"mute_audio"`
		MuteVideo bool   `json:"mute_video"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.lk.MuteAllParticipants(c.Request.Context(), body.RoomCode, body.MuteAudio, body.MuteVideo); err != nil {
		logAndRespondError(c, http.StatusInternalServerError, "failed to mute participants", err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *LivekitHandler) AdmitParticipant(c *gin.Context) {
	var body struct {
		RoomCode string `json:"room_code"`
		Identity string `json:"identity"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := c.GetUint("user_id")
	role := c.GetString("role")

	room, err := h.roomSvc.GetRoomByCode(body.RoomCode)
	if err != nil {
		logAndRespondError(c, http.StatusNotFound, "room not found", err)
		return
	}

	isCreator := room.CreatedByID == userID
	isAdmin := role == "admin"

	if !isCreator && !isAdmin {
		respondError(c, http.StatusForbidden, "you are not authorized to admit participants")
		return
	}

	canPub := true
	canSub := true
	canData := true
	permission := &livekit.ParticipantPermission{
		CanPublish:     canPub,
		CanSubscribe:   canSub,
		CanPublishData: canData,
	}
	metadata := `{"status":"active"}`

	if err := h.lk.UpdateParticipant(c.Request.Context(), body.RoomCode, body.Identity, metadata, permission); err != nil {
		logAndRespondError(c, http.StatusInternalServerError, "failed to admit participant", err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *LivekitHandler) UpdateRoomPermissions(c *gin.Context) {
	var body struct {
		RoomCode string `json:"room_code"`
		Metadata string `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.lk.UpdateRoomMetadata(c.Request.Context(), body.RoomCode, body.Metadata); err != nil {
		logAndRespondError(c, http.StatusInternalServerError, "failed to update room metadata", err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *LivekitHandler) MuteParticipant(c *gin.Context) {
	var body struct {
		RoomCode  string `json:"room_code"`
		Identity  string `json:"identity"`
		MuteAudio bool   `json:"mute_audio"`
		MuteVideo bool   `json:"mute_video"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.RoomCode == "" || body.Identity == "" {
		respondError(c, http.StatusBadRequest, "room_code and identity are required")
		return
	}

	userID := c.GetUint("user_id")
	role := c.GetString("role")

	room, err := h.roomSvc.GetRoomByCode(body.RoomCode)
	if err != nil {
		logAndRespondError(c, http.StatusNotFound, "room not found", err)
		return
	}

	isCreator := room.CreatedByID == userID
	isAdmin := role == "admin"

	if !isCreator && !isAdmin {
		respondError(c, http.StatusForbidden, "you are not authorized to mute participants")
		return
	}

	if err := h.lk.MuteParticipant(c.Request.Context(), body.RoomCode, body.Identity, body.MuteAudio, body.MuteVideo); err != nil {
		logAndRespondError(c, http.StatusInternalServerError, "failed to mute participant", err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *LivekitHandler) BanParticipant(c *gin.Context) {
	var body struct {
		RoomCode string `json:"room_code"`
		Identity string `json:"identity"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := c.GetUint("user_id")
	role := c.GetString("role")

	room, err := h.roomSvc.GetRoomByCode(body.RoomCode)
	if err != nil {
		logAndRespondError(c, http.StatusNotFound, "room not found", err)
		return
	}

	isCreator := room.CreatedByID == userID
	isAdmin := role == "admin"

	if !isCreator && !isAdmin {
		respondError(c, http.StatusForbidden, "not authorized")
		return
	}

	if err := h.roomSvc.BanUser(body.RoomCode, body.Identity); err != nil {
		logAndRespondError(c, http.StatusInternalServerError, "failed to ban user", err)
		return
	}

	_ = h.lk.RemoveParticipant(c.Request.Context(), body.RoomCode, body.Identity)
	_ = h.lk.SetUserOffline(c.Request.Context(), body.Identity)

	c.Status(http.StatusOK)
}

func (h *LivekitHandler) UnbanParticipant(c *gin.Context) {
	var body struct {
		RoomCode string `json:"room_code"`
		Identity string `json:"identity"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := c.GetUint("user_id")
	role := c.GetString("role")

	room, err := h.roomSvc.GetRoomByCode(body.RoomCode)
	if err != nil {
		logAndRespondError(c, http.StatusNotFound, "room not found", err)
		return
	}

	isCreator := room.CreatedByID == userID
	isAdmin := role == "admin"

	if !isCreator && !isAdmin {
		respondError(c, http.StatusForbidden, "not authorized")
		return
	}

	if err := h.roomSvc.UnbanUser(body.RoomCode, body.Identity); err != nil {
		logAndRespondError(c, http.StatusInternalServerError, "failed to unban user", err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *LivekitHandler) GetPublicRoom(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		respondError(c, http.StatusBadRequest, "room code required")
		return
	}
	dbRoom, err := h.roomSvc.GetRoomByCode(code)
	if err != nil {
		respondError(c, http.StatusNotFound, "room not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":                  dbRoom.Name,
		"description":           dbRoom.Description,
		"is_password_protected": dbRoom.Password != "",
		"start_date":            dbRoom.StartDate,
		"end_date":              dbRoom.EndDate,
		"is_public":             (dbRoom.GroupID == nil || *dbRoom.GroupID == 0) && len(dbRoom.AssignedTo) == 0,
	})
}

func (h *LivekitHandler) JoinPublicRoom(c *gin.Context) {
	var body struct {
		RoomCode string `json:"room_code"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, "invalid body")
		return
	}

	dbRoom, err := h.roomSvc.GetRoomByCode(body.RoomCode)
	if err != nil {
		respondError(c, http.StatusNotFound, "room not found")
		return
	}

	now := time.Now()
	if now.Before(dbRoom.StartDate) {
		respondError(c, http.StatusForbidden, "meeting has not started")
		return
	}
	if now.After(dbRoom.EndDate) {
		respondError(c, http.StatusForbidden, "meeting has ended")
		return
	}

	isPublic := (dbRoom.GroupID == nil || *dbRoom.GroupID == 0) && len(dbRoom.AssignedTo) == 0
	if !isPublic {
		respondError(c, http.StatusForbidden, "this room is private, please login")
		return
	}

	if dbRoom.Password != "" {
		if dbRoom.Password != body.Password {
			respondError(c, http.StatusForbidden, "invalid password")
			return
		}
	}

	identity := body.Username
	if identity == "" {
		respondError(c, http.StatusBadRequest, "username is required")
		return
	}

	isWaiting := false
	waitingRoomEnabled := true
	
	liveRooms, err := h.lk.ListRooms(c.Request.Context())
	if err == nil {
		for _, r := range liveRooms {
			if r.Name == dbRoom.RoomCode {
				if r.Metadata != "" {
					var metaMap map[string]interface{}
					if json.Unmarshal([]byte(r.Metadata), &metaMap) == nil {
						if val, ok := metaMap["waiting_room_enabled"]; ok {
							if boolVal, ok := val.(bool); ok {
								waitingRoomEnabled = boolVal
							}
						}
					}
				}
				break
			}
		}
	}

	if waitingRoomEnabled {
		isWaiting = true
	}

	token, host, err := h.lk.GenerateUserToken(c.Request.Context(), dbRoom.RoomCode, identity, isWaiting)
	if err != nil {
		logAndRespondError(c, http.StatusInternalServerError, "failed to generate token", err)
		return
	}

	_ = h.lk.SetUserOnline(c.Request.Context(), 0, identity, dbRoom.RoomCode, 2*time.Minute)

	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"host":       host,
		"identity":   identity,
		"room_name":  dbRoom.Name,
		"is_waiting": isWaiting,
	})
}
