package http

import (
	"github.com/gin-gonic/gin"
)

// ConfigureRoleRoutes helper to keep router.go clean
func ConfigureRoleRoutes(r *gin.RouterGroup, handler *RoleHandler) {
	roles := r.Group("/roles")
	roles.GET("", RequirePermission("role:read"), handler.ListRoles)
	roles.POST("", RequirePermission("role:create"), handler.CreateRole)
	roles.PATCH("/:id", RequirePermission("role:update"), handler.UpdateRole)
	roles.DELETE("/:id", RequirePermission("role:delete"), handler.DeleteRole)

	roles.POST("/:id/permissions", RequirePermission("role:update"), handler.AssignPermission)
	roles.DELETE("/:id/permissions/:permID", RequirePermission("role:update"), handler.RevokePermission)

	roles.GET("/permissions", RequirePermission("role:read"), handler.ListPermissions)
	roles.POST("/init-defaults", RequirePermission("role:update"), handler.InitDefaults)
}
