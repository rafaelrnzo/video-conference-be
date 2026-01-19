package http

import (
	"log"
	"net/http"

	"video-conference-be/internal/app/service"
	"video-conference-be/internal/domain/poll"

	"github.com/gin-gonic/gin"
)

type PollHandler struct {
	service *service.PollService
}

func NewPollHandler(service *service.PollService) *PollHandler {
	return &PollHandler{service: service}
}

func (h *PollHandler) SavePoll(c *gin.Context) {
	var req poll.Poll
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SavePoll(c.Request.Context(), &req); err != nil {
		log.Printf("[POLL] Failed to save poll: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save poll"})
		return
	}

	log.Printf("[POLL] Poll saved: %s", req.PollID)
	c.JSON(http.StatusOK, gin.H{"message": "Poll saved successfully"})
}
