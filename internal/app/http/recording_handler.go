package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"video-conference-be/internal/app/service"
	"video-conference-be/pkg/utility"

	"github.com/gin-gonic/gin"
)

type RecordingHandler struct {
	lk        service.LivekitService
	recordSvc service.RecordService
	recordingSvc service.RecordingService
}

func NewRecordingHandler(lk service.LivekitService, recordSvc service.RecordService, recordingSvc service.RecordingService) *RecordingHandler {
	return &RecordingHandler{
		lk:        lk,
		recordSvc: recordSvc,
		recordingSvc: recordingSvc,
	}
}

func (h *RecordingHandler) StartRecording(c *gin.Context) {
	var req struct {
		RoomName       string `json:"room_name" binding:"required"`
		FilenamePrefix string `json:"filename_prefix"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url := fmt.Sprintf("%s/start", utility.Config.RecorderSvcURL)
	payload := map[string]string{
		"roomId":   req.RoomName,
		"roomCode": req.RoomName,
	}
	bodyBytes, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to contact recorder service: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		c.JSON(resp.StatusCode, gin.H{"error": errResp["error"]})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "recording started via recorder-service",
		"room_name": req.RoomName,
		"status":    "STARTED",
	})
}

func (h *RecordingHandler) StopRecording(c *gin.Context) {
	var req struct {
		RoomName string `json:"room_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url := fmt.Sprintf("%s/stop", utility.Config.RecorderSvcURL)
	payload := map[string]string{"roomId": req.RoomName}
	bodyBytes, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stop recording via recorder-service: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		c.JSON(resp.StatusCode, gin.H{"error": errResp["error"]})
		return
	}

	var successResp struct {
		Message    string `json:"message"`
		OutputName string `json:"outputName"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&successResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode recorder response"})
		return
	}

	endpoint := strings.TrimRight(utility.Config.MinioEndpoint, "/")
	bucket := strings.TrimRight(utility.Config.MinioBucket, "/")
	key := strings.TrimLeft(successResp.OutputName, "/")

	recLink := fmt.Sprintf("%s/%s/%s", endpoint, bucket, key)
	recName := successResp.OutputName

	if _, err := h.recordSvc.Create(
		c.Request.Context(),
		req.RoomName,
		recName,
		recLink,
		successResp.OutputName, // using outputName as egressId
		"PROCESSING",
	); err != nil {
		log.Printf("[RECORD SAVE] failed to save record: %v", err)
	} else {
		log.Printf("[RECORD SAVE] saved record for room=%s link=%s", req.RoomName, recLink)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "recording stopped. processing in background",
		"egress_id":   successResp.OutputName,
		"room_name":   req.RoomName,
		"status":      "PROCESSING",
		"record_url":  recLink,
		"record_name": recName,
	})
}

func (h *RecordingHandler) ListRecords(c *gin.Context) {
	roomID := c.Query("room_id")

	var (
		result interface{}
		err    error
	)

	if roomID != "" {
		result, err = h.recordSvc.ListByRoomID(c.Request.Context(), roomID)
	} else {
		result, err = h.recordSvc.ListAll(c.Request.Context())
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *RecordingHandler) UpdateRecordName(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var body struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	rec, err := h.recordSvc.UpdateName(c.Request.Context(), uint(id64), body.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rec)
}

func (h *RecordingHandler) UpdateRecordStatus(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Status) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}

	rec, err := h.recordSvc.UpdateStatus(c.Request.Context(), uint(id64), body.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rec)
}

func (h *RecordingHandler) DeleteRecord(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.recordingSvc.DeleteRecording(c.Request.Context(), uint(id64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *RecordingHandler) Sync(c *gin.Context) {
	if err := h.recordingSvc.SyncFromMinio(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sync triggered"})
}
