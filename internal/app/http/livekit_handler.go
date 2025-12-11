package http

import (
	"fmt"
	"net/http"
	"time"

	"video-conference-be/internal/app/service"

	"github.com/gin-gonic/gin"
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

// respondError bikin format error JSON konsisten: { "error": "..." }
func respondError(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": msg,
	})
}

// logAndRespondError log error ke gin (biar keliatan di middleware / logger) lalu kirim response ke client
func logAndRespondError(c *gin.Context, status int, msg string, err error) {
	if err != nil {
		// simpan error ke gin context (bisa diproses middleware logging)
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
		// kita anggap error di sini = room tidak ditemukan / invalid,
		// kalau nanti ada ErrNotFound khusus tinggal pakai errors.Is
		logAndRespondError(c, http.StatusNotFound, "room not found or invalid code", err)
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

	identity := c.GetString("username")
	userID := c.GetUint("user_id")

	isAllowed := false

	// 1. Cek AssignedTo
	if len(dbRoom.AssignedTo) > 0 {
		if stringInSlice(identity, dbRoom.AssignedTo) {
			isAllowed = true
		}
	} else {
		// Jika AssignedTo kosong, cek apakah Room Publik (GroupID nil/0)
		if dbRoom.GroupID == nil || *dbRoom.GroupID == 0 {
			isAllowed = true
		}
	}

	// 2. Cek Group Membership
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

	token, host, err := h.lk.GenerateUserToken(c.Request.Context(), dbRoom.RoomCode, identity)
	if err != nil {
		logAndRespondError(c, http.StatusInternalServerError, "failed to generate token", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"identity":  identity,
		"room":      dbRoom.RoomCode,
		"room_name": dbRoom.Name,
		"token":     token,
		"host":      host,
	})
}

func (h *LivekitHandler) ListActiveRooms(c *gin.Context) {
	rooms, err := h.lk.ListRooms(c.Request.Context())
	if err != nil {
		logAndRespondError(c, http.StatusInternalServerError, "failed to list active rooms", err)
		return
	}

	// Jika rooms nil (tidak ada room aktif), return array kosong []
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
