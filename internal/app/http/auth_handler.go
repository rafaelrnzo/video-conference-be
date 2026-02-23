package http

import (
	"net/http"
	"time"
	"video-conference-be/internal/app/service"
	"video-conference-be/pkg/utility"

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

type ssoLoginReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (h *AuthHandler) SSOLogin(c *gin.Context) {
	var req ssoLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	u, err := h.authService.SyncUserFromSSO(c.Request.Context(), req.Username, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	roleName := "user"
	if u.Role != nil {
		roleName = u.Role.Name
	}
	token, err := utility.GenerateJWT(u.Username, roleName, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

func (h *AuthHandler) Me(c *gin.Context) {
	username := c.GetString("username")

	_, err := h.authService.SyncUserFromSSO(c.Request.Context(), username, "") // Or a new method GetUserByUsername. SyncUserFromSSO effectively finds or creates.
	profile, err := h.authService.GetProfile(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       profile.ID,
		"username": profile.Username,
		"role":     profile.Role,
	})
}
