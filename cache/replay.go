package cache

import (
	"net/http"

	"github.com/its-ernest/echox/internal/store"
	"github.com/labstack/echo/v5"
)

// replay sends the cached response to the client
func replay(c echo.Context, entry *store.Entry) error {
	clientETag := c.Request().Header.Get("If-None-Match")
	serverETag := entry.Header.Get("Etag")

	if serverETag != "" && clientETag == serverETag {
		return c.NoContent(http.StatusNotModified)
	}

	for k, v := range entry.Header {
		c.Response().Header()[k] = v
	}
	c.Response().Header().Set("X-Cache", "HIT")

	contentType := entry.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(entry.Body)
	}
	return c.Blob(entry.Status, contentType, entry.Body)
}