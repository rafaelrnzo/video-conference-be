package http

import (
	"net/http"
	"video-conference-be/internal/app/service"
	"video-conference-be/internal/pkg/rbac"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	svc service.RoleService
}

func NewRoleHandler(svc service.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

type PermissionReq struct {
	Role   string `json:"role" binding:"required"`
	Object string `json:"object" binding:"required"`
	Action string `json:"action" binding:"required"`
}

type RoleReq struct {
	Role string `json:"role" binding:"required"`
}

// ListRoles returns all roles from the database
func (h *RoleHandler) ListRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list roles"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// GetRolePermissions returns permissions for a specific role
func (h *RoleHandler) GetRolePermissions(c *gin.Context) {
	role := c.Param("role")
	// Get permissions for the role (filtered policy)
	// Get permissions for the role including inherited ones
	// p, role, obj, act
	policy, err := rbac.Enforcer.GetImplicitPermissionsForUser(role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get permissions"})
		return
	}

	// Format: check model.conf. p = sub, obj, act
	// result is [[role, obj, act], ...]
	var perms []map[string]string
	for _, p := range policy {
		if len(p) >= 3 {
			perms = append(perms, map[string]string{
				"object": p[1],
				"action": p[2],
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"role": role, "permissions": perms})
}

// AddPermission adds a policy rule (p, role, obj, act)
func (h *RoleHandler) AddPermission(c *gin.Context) {
	var req PermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure role exists in DB
	if _, err := h.svc.CreateRole(req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to ensure role exists"})
		return
	}

	added, err := rbac.Enforcer.AddPolicy(req.Role, req.Object, req.Action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add permission"})
		return
	}

	if !added {
		c.JSON(http.StatusConflict, gin.H{"message": "permission already exists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permission added"})
}

// RemovePermission removes a policy rule
func (h *RoleHandler) RemovePermission(c *gin.Context) {
	var req PermissionReq
	// Allow passing via query params as well for DELETE, or JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		// Try query params
		req.Role = c.Query("role")
		req.Object = c.Query("object")
		req.Action = c.Query("action")

		if req.Role == "" || req.Object == "" || req.Action == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
	}

	removed, err := rbac.Enforcer.RemovePolicy(req.Role, req.Object, req.Action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove permission"})
		return
	}

	if !removed {
		c.JSON(http.StatusNotFound, gin.H{"message": "permission not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permission removed"})
}

	// CreateRole adds role to DB
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req RoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := h.svc.CreateRole(req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role created"})
}

// ListSystemPermissions returns all available system permissions
func (h *RoleHandler) ListSystemPermissions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"permissions": rbac.SystemPermissions})
}
