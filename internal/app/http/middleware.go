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
			log.Println("[JWT] parse error:", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		var u dUser.User
		var userID uint
		if err := utility.DB.Where("username = ?", claims.Username).First(&u).Error; err != nil {
			log.Printf("[JWT] WARNING: user with username=%s not found, user_id=0. err=%v\n", claims.Username, err)
			userID = 0
		} else {
			userID = u.ID
		}

		log.Printf("[JWT] OK user_id=%d username=%s role=%s\n", userID, claims.Username, claims.Role)

		c.Set("user_id", userID)
		c.Set("username", claims.Username)
		c.Set("role", string(claims.Role))

		c.Next()
	}
}

func safePrefix(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
