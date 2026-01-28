package http

import (
    "net/http"
    "github.com/gin-gonic/gin"
)


func CheckPermission(requiredPerm string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Check if super admin (optional, for backward compatibility or simple override)
        roleVal, exists := c.Get("role")
        if exists && roleVal.(string) == "admin" {
            c.Next()
            return
        }

        // 2. Check permissions map
        permsVal, exists := c.Get("permissions")
        if !exists {
             c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no permissions found"})
             return
        }
        
        perms, ok := permsVal.(map[string]bool)
        if !ok {
             c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "permission type mismatch"})
             return
        }
        
        if _, has := perms[requiredPerm]; !has {
             c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permission denied: " + requiredPerm})
             return
        }
        
        c.Next()
    }
}
