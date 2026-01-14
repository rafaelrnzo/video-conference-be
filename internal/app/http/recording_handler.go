package http

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"

	"video-conference-be/internal/app/service"
	"video-conference-be/pkg/utility"

	"github.com/gin-gonic/gin"
	livekit "github.com/livekit/protocol/livekit"
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

	info, err := h.lk.StartRoomRecording(c.Request.Context(), req.RoomName, req.FilenamePrefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to start recording: " + err.Error(),
		})
		return
	}

	log.Printf("[EGRESS START] id=%s room=%s status=%s error=%s\n",
		info.EgressId, info.RoomName, info.Status.String(), info.Error)

	c.JSON(http.StatusOK, gin.H{
		"message":   "recording started",
		"egress_id": info.EgressId,
		"room_name": info.RoomName,
		"status":    info.Status.String(),
		"error":     info.Error,
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

	info, err := h.lk.StopRoomRecording(c.Request.Context(), req.RoomName)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "no egress found") {
			c.JSON(http.StatusConflict, gin.H{
				"error":  "no recording egress found for this room (maybe never started or already cleaned up)",
				"detail": msg,
				"room":   req.RoomName,
				"egress": nil,
				"status": "NONE",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stop recording: " + msg})
		return
	}

	log.Printf("[EGRESS STOP] id=%s room=%s status=%s error=%s\n",
		info.EgressId, info.RoomName, info.Status.String(), info.Error)

	var recLink string
	var recName string

	if info.Status == livekit.EgressStatus_EGRESS_COMPLETE {
		recName, recLink = buildRecordingLinkFromInfo(info)
		if recLink != "" {
			roomID := info.RoomName
			if roomID == "" {
				roomID = req.RoomName
			}
			if recName == "" {
				recName = fmt.Sprintf("%s recording", roomID)
			}

			if _, err := h.recordSvc.Create(
				c.Request.Context(),
				roomID,
				recName,
				recLink,
				info.EgressId,
			); err != nil {
				log.Printf("[RECORD SAVE] failed to save record: %v", err)
			} else {
				log.Printf("[RECORD SAVE] saved record for room=%s link=%s", roomID, recLink)
			}
		} else {
			log.Printf("[RECORD SAVE] EGRESS_COMPLETE but no file path found for egress=%s", info.EgressId)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "recording stopped (or already finished/aborted)",
		"egress_id":   info.EgressId,
		"room_name":   info.RoomName,
		"status":      info.Status.String(),
		"error":       info.Error,
		"record_url":  recLink,
		"record_name": recName,
	})
}

func buildRecordingLinkFromInfo(info *livekit.EgressInfo) (name string, url string) {
	if info == nil {
		return "", ""
	}

	var objectPath string

	if len(info.SegmentResults) > 0 {
		s := info.SegmentResults[0]
		playlist := strings.Trim(s.PlaylistName, "/")

		if playlist != "" {
			// fixed build error
			objectPath = playlist
		}
	}

	if objectPath == "" && len(info.FileResults) > 0 {
		f := info.FileResults[0]
		objectPath = strings.TrimLeft(f.Filename, "/")
	}

	if objectPath == "" {
		return "", ""
	}

	endpoint := strings.TrimRight(utility.Config.MinioEndpoint, "/")
	bucket := strings.TrimRight(utility.Config.MinioBucket, "/")
	key := strings.TrimLeft(objectPath, "/")

	url = fmt.Sprintf("%s/%s/%s", endpoint, bucket, key)
	name = objectPath
	return name, url
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

func (h *RecordingHandler) DeleteRecord(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.recordSvc.Delete(c.Request.Context(), uint(id64)); err != nil {
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
