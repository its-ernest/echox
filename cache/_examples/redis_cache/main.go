package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/its-ernest/echox/cache"
	"github.com/its-ernest/echox/store"
	"github.com/labstack/echo/v5"
)

func main() {
	e := echo.New()

	// Custom logger middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			err := next(c)
			stop := time.Now()
			statusColor := "\033[32m" // Green
			if c.Response().Status >= 400 {
				statusColor = "\033[33m" // Yellow
			}
			if c.Response().Status >= 500 {
				statusColor = "\033[31m" // Red
			}
			reset := "\033[0m"
			log.Printf("%s %s %s| %s%d%s | %10v | %s",
				"\033[34m[API]\033[0m", c.Request().Method, c.Request().URL.Path,
				statusColor, c.Response().Status, reset, stop.Sub(start), c.Request().RemoteAddr)
			return err
		}
	})

	// Redis configuration - use environment variable or default
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	// Initialize Redis store
	redisStore := store.NewRedisStore(redisAddr)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisStore.Ping(ctx); err != nil {
		log.Printf("\033[33m[WARN]\033[0m Failed to connect to Redis at %s: %v", redisAddr, err)
		log.Printf("Falling back to memory store...")
		redisStore = nil
	} else {
		log.Printf("\033[32m[OK]\033[0m Connected to Redis at %s", redisAddr)
	}

	// Use memory store as fallback if Redis is unavailable
	var cacheStore store.Store
	if redisStore != nil {
		cacheStore = redisStore
	} else {
		cacheStore = store.NewMemoryStore()
	}

	// Apply cache middleware
	e.Use(cache.New(cache.Config{
		Store: cacheStore,
		TTL:   10 * time.Second,
	}))

	// Example endpoint that benefits from caching
	e.GET("/api/time", func(c *echo.Context) error {
		log.Println("  \033[36m[HANDLER]\033[0m Generating fresh content...")
		timestamp := time.Now().Format(time.RFC3339Nano)
		return c.String(http.StatusOK, fmt.Sprintf("Generated at: %s", timestamp))
	})

	// Example endpoint with dynamic data
	e.GET("/api/user/:id", func(c *echo.Context) error {
		id := c.Param("id")
		log.Printf("  \033[36m[HANDLER]\033[0m Fetching user %s...", id)
		time.Sleep(100 * time.Millisecond) // Simulate DB lookup
		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":    id,
			"name":  fmt.Sprintf("User %s", id),
			"email": fmt.Sprintf("user%s@example.com", id),
		})
	})

	log.Println("Redis Cache Example starting on :8080")
	log.Printf("Redis address: %s", redisAddr)

	// Print routes
	for _, route := range e.Router().Routes() {
		fmt.Printf("\033[36m[ROUTE]\033[0m %-6s %-15s -> %s\n", route.Method, route.Path, route.Name)
	}

	if err := e.Start(":8080"); err != nil {
		log.Fatalf("Server failed: %s", err)
	}
}