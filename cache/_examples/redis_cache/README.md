# Redis Cache Example

This example demonstrates how to use Redis as a distributed cache backend for the echox middleware.

## Prerequisites

- Redis server running on `localhost:6379` (or configure your own address)
- Go 1.25+

## Running the Example

```bash
# Start Redis (using Docker)
docker run -d -p 6379:6379 redis:alpine

# Run the example
go run main.go
```

## Usage

```go
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

    // Use Redis for distributed caching
    redisStore := store.NewRedisStore("localhost:6379")
    
    // Optional: Check connection
    if err := redisStore.Ping(nil); err != nil {
        panic("Failed to connect to Redis: " + err.Error())
    }

    // Apply cache middleware with Redis backend
    e.Use(cache.New(cache.Config{
        Store: redisStore,
        TTL:   10 * time.Minute,
    }))

    e.GET("/api/data", func(c *echo.Context) error {
        return c.String(http.StatusOK, "Hello from Redis cache!")
    })

    e.Start(":8080")
}
```

## Production Configuration

For production environments, use `NewRedisStoreWithConfig` for fine-grained control:

```go
redisStore := store.NewRedisStoreWithConfig(store.RedisStoreConfig{
    Addr:         "redis-cluster.example.com:6379",
    Password:      "your-password",
    DB:           0,
    PoolSize:     50,
    MinIdleConns: 10,
})
```

## Combining with Abuse Detection

You can use the same Redis instance for both caching and abuse detection:

```go
redisStore := store.NewRedisStore("localhost:6379")
redisCounter := store.NewRedisCounter("localhost:6379")

// Cache middleware
e.Use(cache.New(cache.Config{
    Store: redisStore,
    TTL:   5 * time.Minute,
}))

// Abuse detection middleware
e.Use(abuse.New(abuse.Config{
    Store:     redisCounter,
    Threshold: 100,
    Rules: []abuse.Rule{
        {Path: "/.env", Score: 100},
        {Path: "/api/v1/login", Method: "POST", Score: 20},
    },
}))
```