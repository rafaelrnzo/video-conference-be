package http

import (
	"net/http"
	"strconv"

	"video-conference-be/internal/app/service"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	svc service.GroupService
}

func NewGroupHandler(svc service.GroupService) *GroupHandler {
	return &GroupHandler{svc: svc}
}

func (h *GroupHandler) ListGroups(c *gin.Context) {
	groups, err := h.svc.ListGroups(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

func (h *GroupHandler) GetGroup(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	g, err := h.svc.GetGroup(c.Request.Context(), uint(id64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, g)
}

type createGroupReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var body createGroupReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	g, err := h.svc.CreateGroup(c.Request.Context(), body.Name, body.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, g)
}

func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var body createGroupReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	g, err := h.svc.UpdateGroup(c.Request.Context(), uint(id64), body.Name, body.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, g)
}

func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.svc.DeleteGroup(c.Request.Context(), uint(id64)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type memberReq struct {
	UserID uint `json:"user_id"`
}

func (h *GroupHandler) AddMember(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	var body memberReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if err := h.svc.AddMember(c.Request.Context(), uint(groupID), body.UserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member added"})
}

func (h *GroupHandler) RemoveMember(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.svc.RemoveMember(c.Request.Context(), uint(groupID), uint(userID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
