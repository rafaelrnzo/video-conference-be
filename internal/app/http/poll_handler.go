package http

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PollHandler struct{}

func NewPollHandler() *PollHandler {
	return &PollHandler{}
}

func (h *PollHandler) SavePoll(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In a real implementation, you would map this to a domain model
	// and save it using a repository.
	// For now, we log it as requested.
	log.Printf("[POLL] Saving poll result to Database: %+v", req)

	c.JSON(http.StatusOK, gin.H{
		"message": "Poll result saved successfully",
		"data":    req,
	})
}
