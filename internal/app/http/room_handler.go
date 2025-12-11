package http

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"video-conference-be/internal/app/service"
	"video-conference-be/internal/domain/room"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type RoomHandler struct {
	svc service.RoomService
}

func NewRoomHandler(svc service.RoomService) *RoomHandler {
	return &RoomHandler{svc: svc}
}

func (h *RoomHandler) ListRooms(c *gin.Context) {
	userID := c.GetUint("user_id")
	role := c.GetString("role")
	username := c.GetString("username")

	log.Printf("[ListRooms] userID=%d username=%s role=%s\n", userID, username, role)

	rooms, err := h.svc.ListRooms(userID, username, role)
	if err != nil {
		log.Printf("[ListRooms] ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rooms)
}

type createRoomReq struct {
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	MaxParticipants int       `json:"max_participants"`
	AssignedTo      []string  `json:"assigned_to"`
	GroupID         *uint     `json:"group_id"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var body createRoomReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
		return
	}

	createdByID := c.GetUint("user_id")

	req := room.Room{
		Name:            body.Name,
		Description:     body.Description,
		MaxParticipants: body.MaxParticipants,
		AssignedTo:      pq.StringArray(body.AssignedTo),
		GroupID:         body.GroupID,
		StartDate:       body.StartDate,
		EndDate:         body.EndDate,
		CreatedByID:     createdByID,
	}

	createdRoom, err := h.svc.CreateRoom(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, createdRoom)
}

type updateRoomReq struct {
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	MaxParticipants int       `json:"max_participants"`
	AssignedTo      []string  `json:"assigned_to"`
	GroupID         *uint     `json:"group_id"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
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
		AssignedTo:      pq.StringArray(body.AssignedTo),
		GroupID:         body.GroupID,
		StartDate:       body.StartDate,
		EndDate:         body.EndDate,
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
