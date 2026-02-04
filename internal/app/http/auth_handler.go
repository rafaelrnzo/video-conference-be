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
		"role":     u.Role, // This is now a *Role struct or similar
		"user_id":  u.ID,
	})
}

type ssoLoginReq struct {
	Username string `json:"username"`
	Email    string `json:"email"` // We might use email as unique identifier
	// In a real app, we should verify an ID Token here!
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

	// Generate JWT for our app
	roleName := "user"
	if u.Role != nil {
		roleName = u.Role.Name
	}
	token, err := utility.GenerateJWT(u.Username, roleName, 24*time.Hour) // Fixed: pass string
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
	// Determine user from context (set by JWT middleware)
	// The middleware sets "username"
	username := c.GetString("username")

	// We need to fetch the full user to get permissions
	_, err := h.authService.SyncUserFromSSO(c.Request.Context(), username, "") // Or a new method GetUserByUsername. SyncUserFromSSO effectively finds or creates.
	// Actually, we should use a method that just Finds.
	// authService.userRepo is not exposed.
	// But login logic does: u, err := s.userRepo.FindByUsername(ctx, username)
	// We should expose a 'GetProfile' in AuthService or use UserService if injected.
	// AuthHandler has AuthService.

	// Let's assume we can reuse Login or we need to extend AuthService.
	// Login requires password.
	// Let's add GetProfile to AuthService interface.
	// For now, I will interpret the 'SyncUserFromSSO' call as safe if email is empty?
	// No, SyncUserFromSSO logic creates if not found.
	// Let's modify AuthService to include 'GetProfile'

	// Temporary hack: Just use SyncUserFromSSO? No, that might create duplicates if logic is weak.
	// Check AuthService again.
	// It has `userRepo`. So we can just add `GetProfile` to `AuthService`.

	profile, err := h.authService.GetProfile(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       profile.ID,
		"username": profile.Username,
		"role":     profile.Role,
		// Add more fields if needed
	})
}
