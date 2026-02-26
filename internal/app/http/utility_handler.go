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
		respondError(c, http.StatusUnprocessableEntity, "failed to scrape url: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, meta)
}
