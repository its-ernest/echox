package main

import (
	"fmt"
	"log"
	"os"

	//context"

	//internals
	"otp-backend/internal/auth"
	"otp-backend/internal/service"

	"github.com/its-ernest/echox/internal/store"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	e := echo.New()

	//ctx := context.Background()

	// grab JWT secret key
	jwtSecretStr := os.Getenv("JWT_SECRET")
	if jwtSecretStr == "" {
		log.Println("\033[33m[WARN]\033[0m JWT_SECRET not set, using insecure default")
		jwtSecretStr = "church-default-secret-2026"
	}
	jwtSecret := []byte(jwtSecretStr)

	// echo v5 logger middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:  true,
		LogMethod:  true,
		LogURI:     true,
		LogLatency: true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			statusColor := "\033[32m" // Green
			if v.Status >= 400 {
				statusColor = "\033[33m"
			}
			if v.Status >= 500 {
				statusColor = "\033[31m"
			}
			reset := "\033[0m"

			log.Printf("%s %s %s| %s%d%s | %10v | %s",
				"\033[34m[API]\033[0m", v.Method, v.URI,
				statusColor, v.Status, reset, v.Latency, c.Request().RemoteAddr)
			return nil
		},
	}))

	// initialize the MemoryStore from my echox
	memStore := store.NewMemoryStore()

	// use in-memory store as cache temp for echox
	authService := service.NewAuthService(memStore)
	authHandler := auth.NewHandler(authService, jwtSecret)

	// auth routes groups
	authGroup := e.Group("/auth")
	authHandler.Register(authGroup, memStore)

	log.Println("OTP backend starting on :8080")

	//print all routes
	for _, route := range e.Router().Routes() {
		fmt.Printf("\033[36m[ROUTE]\033[0m %-6s %-15s -> %s\n", route.Method, route.Path, route.Name)
	}

	if err := e.Start(":8080"); err != nil {
		log.Fatalf("Server failed: %s", err)
	}
}
