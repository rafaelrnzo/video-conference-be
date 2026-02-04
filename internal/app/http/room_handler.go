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

func (h *RoomHandler) GetRoomByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room code is required"})
		return
	}

	room, err := h.svc.GetRoomByCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}
	c.JSON(http.StatusOK, room)
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

func (h *RoomHandler) UploadPresentation(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	path, err := h.svc.UploadPresentation(c.Request.Context(), uint(id64), fileHeader)
	if err != nil {
		log.Printf("[ERROR] UploadPresentation failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"path": path,
	})
}

func (h *RoomHandler) ProxyPresentation(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 1. Get Room to find the path
	// We can use svc.GetRoomById if available, or fetch manually if needed.
	// Since RoomService interface has GetRoomByCode but not ID exposed easily (only Update/Delete use ID),
	// We might need to extend Service or rely on repo.
	// Actually, UpdateRoom uses ID. Let's look at service: UpdateRoom(req room.Room).
	// We will add a method GetRoomByID to service or just use raw DB in handler if forced (bad practice).
	// Let's assume we add GetRoomByID to service. I'll add that next.

	// For now, I will assume h.svc has GetRoomByID. I will invoke a new tool to add it if missing.
	// Wait, I can see the file `room_service.go` in previous steps. It DOES NOT have GetRoomById.
	// It has GetRoomByCode.
	// I should add GetRoomByID to RoomService interface and implementation.

	// TEMPORARY HACK: Use reflection/raw access? No, let's do it properly.
	// I will add GetRoomByID in the next step.
	// For this chunk, I will write the handler assuming the service method exists.

	r, err := h.svc.GetRoomByID(uint(id64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "presentation not found"})
		return
	}

	if r.PresentationPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "no presentation uploaded"})
		return
	}

	// 2. Stream from MinIO
	// The Service needs to expose a Stream method, or we handle it here if service exposes one.
	// Current service has `UploadPresentation`.
	// I should add `DownloadPresentation(ctx, roomID)` that returns reader, contentType, error.

	reader, contentType, err := h.svc.DownloadPresentation(c.Request.Context(), r.PresentationPath)
	if err != nil {
		log.Printf("[ERROR] DownloadPresentation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to download file"})
		return
	}
	defer reader.Close()

	extraHeaders := map[string]string{
		"Content-Disposition": `inline; filename="presentation.pdf"`,
	}
	c.DataFromReader(http.StatusOK, -1, contentType, reader, extraHeaders)
}
