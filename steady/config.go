package steady

import (
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// Config defines the configuration for the steady middleware.
type Config struct {
	// Skipper defines a function to skip middleware.
	Skipper middleware.Skipper

	// MaxConcurrent is the maximum number of concurrent requests allowed.
	// Default is 100.
	MaxConcurrent int

	// WaitTimeout is the maximum time a request will wait for a processing slot.
	// Default is 10 seconds.
	WaitTimeout time.Duration

	// ErrorHandler is called when a request exceeds the limit.
	// Default returns 503 Service Unavailable with Retry-After header.
	ErrorHandler func(c *echo.Context) error
}

// DefaultConfig is the default configuration for the steady middleware.
var DefaultConfig = Config{
	Skipper:       middleware.DefaultSkipper,
	MaxConcurrent: 100,
	WaitTimeout:   10 * time.Second,
}
