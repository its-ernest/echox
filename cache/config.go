package cache

import (
	"fmt"
	"time"

	"github.com/its-ernest/echox/store"
	"github.com/labstack/echo/v5"
)

type Config struct {
	Store        store.Store
	TTL          time.Duration
	Skipper      func(c *echo.Context) bool
	KeyGenerator func(c *echo.Context) string

	// MaxBodySize prevents caching of massive responses (default 1MB)
	MaxBodySize  int
	
	RetryDelay   time.Duration
	MaxRetries   int
}

// DefaultKeyGenerator generates a cache key based on method + URL
func DefaultKeyGenerator(c *echo.Context) string {
	return fmt.Sprintf("cache:%s:%s", c.Request().Method, c.Request().URL.String())
}