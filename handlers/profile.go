package handlers

import (
	"net/http"
	"video-conference-be/middleware"

	"github.com/gin-gonic/gin"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User profile fetched successfully",
		"user":    user,
	})
}

func (h *UserHandler) AdminEndpoint(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome Admin!",
	})
}
