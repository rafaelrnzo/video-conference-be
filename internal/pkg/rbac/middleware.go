package rbac

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Middleware enforces RBAC policies directly using Casbin
// obj: resource (e.g., "rooms")
// act: action (e.g., "create")
func Middleware(obj, act string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role from context (set by JWTAuthMiddleware)
		// We trust the JWT middleware to have validated the token and extracted the role.
		role := c.GetString("role")
		if role == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no role found in context"})
			return
		}

		// Enforce policy using Casbin
		// Sub: role, Obj: obj, Act: act
		ok, err := Enforcer.Enforce(role, obj, act)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "rbac enforce error"})
			return
		}

		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("permission denied: %s %s", act, obj)})
			return
		}

		c.Next()
	}
}
