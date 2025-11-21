package http

import (
	"net/http"
	"video-conference-be/internal/app/service"

	"github.com/gin-gonic/gin"
	livekit "github.com/livekit/protocol/livekit"
)

type LivekitHandler struct {
	lk service.LivekitService
}

func NewLivekitHandler(lk service.LivekitService) *LivekitHandler {
	return &LivekitHandler{lk: lk}
}

func (h *LivekitHandler) Health(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

// USER: POST /api/livekit/token
// Body: { "room": "room1" }
func (h *LivekitHandler) GenerateToken(c *gin.Context) {
	var body struct {
		Room string `json:"room"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Room == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room is required"})
		return
	}

	// diasumsikan JWT middleware sudah set "username" di context
	identity := c.GetString("username")
	if identity == "" {
		// fallback: kalau belum ada, bisa pakai anonymous atau error
		identity = "anonymous"
	}

	token, host, err := h.lk.GenerateUserToken(c.Request.Context(), body.Room, identity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"identity": identity,
		"room":     body.Room,
		"token":    token,
		"host":     host,
	})
}

// ADMIN: POST /admin/livekit/rooms
// Body: { "name": "room1", "emptyTimeout": 0, "maxParticipants": 10, "metadata": "..." }
func (h *LivekitHandler) CreateRoom(c *gin.Context) {
	var body struct {
		Name            string `json:"name"`
		EmptyTimeout    int32  `json:"emptyTimeout"`
		MaxParticipants uint32 `json:"maxParticipants"`
		Metadata        string `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload (need name)"})
		return
	}

	room, err := h.lk.CreateRoom(c.Request.Context(), &livekit.CreateRoomRequest{
		Name:            body.Name,
		EmptyTimeout:    uint32(body.EmptyTimeout),
		MaxParticipants: body.MaxParticipants,
		Metadata:        body.Metadata,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, room)
}

// ADMIN: GET /admin/livekit/rooms
func (h *LivekitHandler) ListRooms(c *gin.Context) {
	rooms, err := h.lk.ListRooms(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rooms)
}

// ADMIN: DELETE /admin/livekit/rooms/:name
func (h *LivekitHandler) DeleteRoom(c *gin.Context) {
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

// ADMIN: GET /admin/livekit/participants?room=room1
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

// ADMIN: DELETE /admin/livekit/participants?room=room1&identity=user1
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
