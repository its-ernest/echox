package steady

import (
	"context"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
)

// New returns a steady middleware that limits concurrent requests.
func New(config Config) echo.MiddlewareFunc {
	// Defaults
	if config.MaxConcurrent <= 0 {
		config.MaxConcurrent = DefaultConfig.MaxConcurrent
	}
	if config.WaitTimeout <= 0 {
		config.WaitTimeout = DefaultConfig.WaitTimeout
	}
	if config.Skipper == nil {
		config.Skipper = DefaultConfig.Skipper
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = func(c *echo.Context) error {
			c.Response().Header().Set("Retry-After", strconv.Itoa(int(config.WaitTimeout.Seconds())))
			return echo.NewHTTPError(http.StatusServiceUnavailable, "Server is reaching capacity. Please try again shortly.")
		}
	}

	// The Semaphore
	sem := make(chan struct{}, config.MaxConcurrent)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			// Create a timeout context for the wait period
			ctx, cancel := context.WithTimeout(c.Request().Context(), config.WaitTimeout)
			defer cancel()

			select {
			case sem <- struct{}{}:
				// Acquired a slot!
				defer func() { <-sem }() // Release slot when done
				return next(c)

			case <-ctx.Done():
				// We waited too long or the client disconnected
				if ctx.Err() == context.DeadlineExceeded {
					return config.ErrorHandler(c)
				}
				// Client disconnected, just return the context error
				return ctx.Err()
			}
		}
	}
}
