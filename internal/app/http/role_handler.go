package http

import (
	"net/http"
	"strconv"
	"video-conference-be/internal/app/service"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	svc service.RoleService
}

func NewRoleHandler(svc service.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

func (h *RoleHandler) ListRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, roles)
}

type createRoleReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *RoleHandler) CreateRole(c *gin.Context) {
	var body createRoleReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	r, err := h.svc.CreateRole(c.Request.Context(), body.Name, body.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, r)
}

func (h *RoleHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body createRoleReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	r, err := h.svc.UpdateRole(c.Request.Context(), uint(id64), body.Name, body.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, r)
}

func (h *RoleHandler) DeleteRole(c *gin.Context) {
    idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
    if err := h.svc.DeleteRole(c.Request.Context(), uint(id64)); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.Status(http.StatusNoContent)
}

type permAssignReq struct {
    PermissionID uint `json:"permission_id"`
}

func (h *RoleHandler) AssignPermission(c *gin.Context) {
    idStr := c.Param("id")
    roleID, err := strconv.ParseUint(idStr, 10, 64)
     if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}
    var body permAssignReq
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
        return
    }
    if err := h.svc.AssignPermission(c.Request.Context(), uint(roleID), body.PermissionID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "permission added"})
}

func (h *RoleHandler) RevokePermission(c *gin.Context) {
    idStr := c.Param("id")
    roleID, err := strconv.ParseUint(idStr, 10, 64)
     if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}
    permIdStr := c.Param("permID")
    permID, err := strconv.ParseUint(permIdStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission id"})
		return
    }
     if err := h.svc.RevokePermission(c.Request.Context(), uint(roleID), uint(permID)); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "permission revoked"})
}

func (h *RoleHandler) ListPermissions(c *gin.Context) {
    perms, err := h.svc.ListPermissions(c.Request.Context())
    if err != nil {
         c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
         return
    }
    c.JSON(http.StatusOK, perms)
}

// Ensure defaults
func (h *RoleHandler) InitDefaults(c *gin.Context) {
    if err := h.svc.InitDefaultRoles(c.Request.Context()); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "defaults initialized"})
}
