package http

import (
	"github.com/gin-gonic/gin"
)

// ConfigureRoleRoutes helper to keep router.go clean
func ConfigureRoleRoutes(r *gin.RouterGroup, handler *RoleHandler) {
	roles := r.Group("/roles")
	// Base Role Management (Admin only or manage:roles permission)
	// We can use CheckPermission("role:read") etc.
	// For now assuming the parent group has AdminOnly or we add specific perms here.

	// For specific endpoints:
	roles.GET("", CheckPermission("role:read"), handler.ListRoles)
	roles.POST("", CheckPermission("role:create"), handler.CreateRole)
	roles.PATCH("/:id", CheckPermission("role:update"), handler.UpdateRole)
	roles.DELETE("/:id", CheckPermission("role:delete"), handler.DeleteRole)

	roles.POST("/:id/permissions", CheckPermission("role:update"), handler.AddPermission)
	roles.DELETE("/:id/permissions/:permID", CheckPermission("role:update"), handler.RemovePermission)

	roles.GET("/permissions", CheckPermission("role:read"), handler.ListPermissions)
	roles.POST("/init-defaults", CheckPermission("role:update"), handler.InitDefaults)
}
