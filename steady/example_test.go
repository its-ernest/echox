package steady_test

import (
	"net/http"
	"time"

	"github.com/its-ernest/echox/steady"
	"github.com/labstack/echo/v5"
)

func ExampleNew() {
	e := echo.New()

	// Configure the steady middleware with a maximum of 10 concurrent requests
	// and a 5-second wait timeout for a slot.
	config := steady.Config{
		MaxConcurrent: 10,
		WaitTimeout:   5 * time.Second,
	}

	// Add the middleware to the Echo instance
	e.Use(steady.New(config))

	e.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.Start(":8080")
}
