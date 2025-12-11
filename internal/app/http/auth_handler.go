package http

import (
	"net/http"
	"video-conference-be/internal/app/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(a service.AuthService) *AuthHandler {
	return &AuthHandler{authService: a}
}

type registerReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.authService.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "user registered",
		"user":    u.Username,
		"role":    u.Role,
	})
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, u, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"username": u.Username,
		"role":     u.Role,
		"user_id":  u.ID,
	})
}

func (h *AuthHandler) Public(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "public endpoint"})
}

func (h *AuthHandler) Protected(c *gin.Context) {
	username := c.GetString("username")
	role := c.GetString("role")
	c.JSON(http.StatusOK, gin.H{
		"message":  "protected endpoint",
		"username": username,
		"role":     role,
	})
}
