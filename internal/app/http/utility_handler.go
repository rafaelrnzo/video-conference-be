package http

import (
	"net/http"
	"video-conference-be/pkg/utility"

	"github.com/gin-gonic/gin"
)

type UtilityHandler struct{}

func NewUtilityHandler() *UtilityHandler {
	return &UtilityHandler{}
}

func (h *UtilityHandler) GetLinkMeta(c *gin.Context) {
	urlStr := c.Query("url")
	if urlStr == "" {
		respondError(c, http.StatusBadRequest, "url query parameter is required")
		return
	}

	meta, err := utility.ScrapeOG(urlStr)
	if err != nil {
		// Log but assume it's just a failed scrape (e.g. 404 or bad host)
		// We return a limited object or error. 
		// For the frontend, it's better to return 200 with empty fields or 422?
		// Let's return 200 but check if we got anything useful.
		respondError(c, http.StatusUnprocessableEntity, "failed to scrape url: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, meta)
}
