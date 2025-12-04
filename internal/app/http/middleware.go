package http

import (
	"log"
	"net/http"
	"strings"

	dUser "video-conference-be/internal/domain/user"
	"video-conference-be/pkg/utility"

	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header"})
			return
		}

		tokenStr := strings.TrimSpace(parts[1])

		log.Printf("[JWT] incoming token (first 30 chars): %s\n", safePrefix(tokenStr, 30))

		claims, err := utility.ParseJWT(tokenStr)
		if err != nil {
			log.Println("[JWT] parse error:", err) // << penting
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		log.Printf("[JWT] OK username=%s role=%s\n", claims.Username, claims.Role)

		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func safePrefix(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "role missing"})
			return
		}
		role, _ := roleVal.(dUser.Role)
		if role != dUser.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin only"})
			return
		}
		c.Next()
	}
}
