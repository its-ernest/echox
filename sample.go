package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/its-ernest/echox/cache" 
	"github.com/its-ernest/echox/store"
)

func main() {
	e := echo.New()

	// MINI CUSTOM LOGGER
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			err := next(c)
			stop := time.Now()
			fmt.Printf("[%s] %s %s (%s)\n", 
				stop.Format("15:04:05"),
				c.Request().Method, 
				c.Request().URL.Path,
				stop.Sub(start),
			)
			return err
		}
	})

	// setup local store
	memStore := store.NewMemoryStore()

	// apply echox cache middleware
	e.Use(cache.New(cache.Config{
		Store: memStore,
		TTL:   10 * time.Second,
	}))

	// Echo V5 handler
	e.GET("/test", func(c *echo.Context) error {
		fmt.Println("  ↳ [HANDLER] Generating fresh content...")
		timestamp := time.Now().Format(time.RFC3339Nano)
		return c.String(http.StatusOK, fmt.Sprintf("Generated at: %s", timestamp))
	})

	fmt.Println("Starting Dev Server on :8080...")
	if err := e.Start(":8080"); err != nil {
		fmt.Printf("Server stopped: %v\n", err)
	}
}