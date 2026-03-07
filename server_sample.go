package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/its-ernest/echox/cache"
	"github.com/its-ernest/echox/cache/internal/store"
)

func main() {
	e := echo.New()

	// Create MemoryStore
	memStore := store.NewMemoryStore()

	// Start periodic eviction
	stopEvictor := make(chan struct{})
	memStore.StartEvictor(1*time.Minute, stopEvictor)

	// Add cache middleware
	e.Use(cache.New(cache.Config{
		Store:       memStore,
		TTL:         2 * time.Minute,
		MaxBodySize: 2 * 1024 * 1024, // 2MB
	}))

	// Example handler
	e.GET("/news", func(c echo.Context) error {
		body := []byte("Latest news content...")
		return c.Blob(http.StatusOK, "text/plain", body)
	})

	// Start server
	go func() {
		if err := e.Start(":8080"); err != nil {
			log.Fatal(err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	close(stopEvictor)
	e.Shutdown(context.Background())
}