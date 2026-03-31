package cache

import (
	"net/http"

	"github.com/its-ernest/echox/store"
	"github.com/labstack/echo/v5"
)

func replay(c *echo.Context, entry *store.Entry) error {
	// 1.cast map to http.Header for .Get() capability
	headers := http.Header(entry.Header)

	clientETag := c.Request().Header.Get("If-None-Match")
	serverETag := headers.Get("Etag")

	if serverETag != "" && clientETag == serverETag {
		return c.NoContent(http.StatusNotModified)
	}

	// Transfer headers to the actual response
	for k, v := range entry.Header {
		for _, val := range v {
			c.Response().Header().Add(k, val)
		}
	}

	c.Response().Header().Set("X-Cache", "HIT")

	// Ensure we provide a Content-Type or use the cached one
	contentType := headers.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return c.Blob(entry.Status, contentType, entry.Body)
}
