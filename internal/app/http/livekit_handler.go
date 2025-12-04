package http

import (
	"net/http"

	"video-conference-be/internal/app/service"

	"github.com/gin-gonic/gin"
)

type LivekitHandler struct {
	lk      service.LivekitService
	roomSvc service.RoomService
}

func NewLivekitHandler(lk service.LivekitService, roomSvc service.RoomService) *LivekitHandler {
	return &LivekitHandler{
		lk:      lk,
		roomSvc: roomSvc,
	}
}

func (h *LivekitHandler) Health(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (h *LivekitHandler) GenerateToken(c *gin.Context) {
	var body struct {
		Room string `json:"room"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Room == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room is required"})
		return
	}

	dbRoom, err := h.roomSvc.GetRoomByName(body.Room)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found in database"})
		return
	}

	identity := c.GetString("username")
	if identity == "" {
		identity = "anonymous"
	}

	token, host, err := h.lk.GenerateUserToken(c.Request.Context(), dbRoom.Name, identity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"identity": identity,
		"room":     dbRoom.Name,
		"token":    token,
		"host":     host,
	})
}

func (h *LivekitHandler) ListActiveRooms(c *gin.Context) {
	rooms, err := h.lk.ListRooms(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rooms)
}

func (h *LivekitHandler) DeleteActiveRoom(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if err := h.lk.DeleteRoom(c.Request.Context(), name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *LivekitHandler) ListParticipants(c *gin.Context) {
	room := c.Query("room")
	if room == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room is required"})
		return
	}
	participants, err := h.lk.ListParticipants(c.Request.Context(), room)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, participants)
}

func (h *LivekitHandler) RemoveParticipant(c *gin.Context) {
	room := c.Query("room")
	identity := c.Query("identity")
	if room == "" || identity == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room and identity are required"})
		return
	}
	if err := h.lk.RemoveParticipant(c.Request.Context(), room, identity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
