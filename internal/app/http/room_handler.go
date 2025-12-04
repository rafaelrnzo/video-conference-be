package http

import (
	"net/http"
	"strconv"

	"video-conference-be/internal/app/service"
	"video-conference-be/internal/domain/room"

	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	svc service.RoomService
}

func NewRoomHandler(svc service.RoomService) *RoomHandler {
	return &RoomHandler{svc: svc}
}

func (h *RoomHandler) ListRooms(c *gin.Context) {
	rooms, err := h.svc.ListRooms()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rooms)
}

type createRoomReq struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	MaxParticipants int    `json:"max_participants"`
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var body createRoomReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	req := room.Room{
		Name:            body.Name,
		Description:     body.Description,
		MaxParticipants: body.MaxParticipants,
	}

	createdRoom, err := h.svc.CreateRoom(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, createdRoom)
}

type updateRoomReq struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	MaxParticipants int    `json:"max_participants"`
}

func (h *RoomHandler) UpdateRoom(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var body updateRoomReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	req := room.Room{
		ID:              uint(id64),
		Name:            body.Name,
		Description:     body.Description,
		MaxParticipants: body.MaxParticipants,
	}

	updatedRoom, err := h.svc.UpdateRoom(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updatedRoom)
}

func (h *RoomHandler) DeleteRoom(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.svc.DeleteRoom(uint(id64)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
