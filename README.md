# echox

[![Go Reference](https://pkg.go.dev/badge/github.com/its-ernest/echox.svg)](https://pkg.go.dev/github.com/its-ernest/echox)
[![Go Report Card](https://goreportcard.com/badge/github.com/its-ernest/echox)](https://goreportcard.com/report/github.com/its-ernest/echox)
[![Dagger Docs](https://github.com/its-ernest/echox/actions/workflows/docs.yml/badge.svg)](https://github.com/its-ernest/echox/actions/workflows/docs.yml)
[![Build Status](https://github.com/its-ernest/echox/actions/workflows/go.yml/badge.svg)](https://github.com/its-ernest/echox/actions/workflows/go.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**High-performance middleware suite for the Echo v5 ecosystem.**

`echox` is a collection of production-ready middlewares engineered specifically for Echo v5's struct-pointer architecture. By leveraging Go 1.24+ features and standard library interfaces, `echox` provides robust solutions for distributed caching, idempotency, and request lifecycle management.

## <i class="fas fa-layer-group"></i> Framework Compatibility

To ensure stability across the evolving Echo v5 landscape, we maintain specific versions mapped to upstream Echo releases.

| Echo Framework | Recommended echox Version | Git Tag / Branch | Status |
| :--- | :--- | :--- | :--- |
| **v5.0.0-beta.x** | `v0.0.x` | `v5.0.x`/ `main` | <i class="fas fa-archive"></i> Active |
| **v5.1.0+** | `v0.1.x` | `v5.1.x` | <i class="fas fa-rocket"></i> Planned |

## <i class="fas fa-cubes"></i> Middleware Registry

| Module | Purpose | Status | Backend Support |
| :--- | :--- | :--- | :--- |
| **[`echox/cache`](cache/README.md)** | RFC-compliant HTTP caching | `Stable` | <i class="fas fa-check-circle" style="color:green"></i> Memory. <br> <i class="fas fa-check-circle" style="color:green"></i> Redis. |
| **[`echox/abuse`](abuse/README.md)** | API abuse detection | `Beta` | <i class="fas fa-check-circle" style="color:green"></i> Memory. <br> <i class="fas fa-check-circle" style="color:green"></i> Redis. |


## <i class="fas fa-terminal"></i> Requirements

* **Go:** 1.25+
* **Echo:** v5.0.0 or higher

## MINI PROJECTS EXAMPLES: 

* **OTP verification caching (MemoryStore)**: [_examples/cache/otp_code_verification](_examples/cache/otp_code_verification/README.md)
* **Redis distributed caching**: [_examples/cache/redis_cache](_examples/cache/redis_cache/README.md)


## <i class="fas fa-play-circle"></i> Quick Start examples
### 1. Caching examples

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
		fmt.Println(" [HANDLER] Generating fresh content...")
		timestamp := time.Now().Format(time.RFC3339Nano)
		return c.String(http.StatusOK, fmt.Sprintf("Generated at: %s", timestamp))
	})

	fmt.Println("Starting Dev Server on :8080...")
	if err := e.Start(":8080"); err != nil {
		fmt.Printf("Server stopped: %v\n", err)
	}
}
```

### 2. Redis Caching (Production)

For production environments with multiple application nodes, use Redis for distributed caching:

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

	// Use Redis for distributed caching across multiple instances
	redisStore := store.NewRedisStore("localhost:6379")

	// For production with custom configuration:
	// redisStore := store.NewRedisStoreWithConfig(store.RedisStoreConfig{
	//     Addr:         "redis-cluster.example.com:6379",
	//     Password:     "your-password",
	//     DB:           0,
	//     PoolSize:     50,
	//     MinIdleConns: 10,
	// })

	// apply echox cache middleware with Redis backend
	e.Use(cache.New(cache.Config{
		Store: redisStore,
		TTL:   5 * time.Minute,
	}))

	// Echo V5 handler
	e.GET("/api/data", func(c *echo.Context) error {
		fmt.Println(" [HANDLER] Generating fresh content...")
		return c.String(http.StatusOK, "Cached response")
	})

	fmt.Println("Starting Production Server on :8080...")
	if err := e.Start(":8080"); err != nil {
		fmt.Printf("Server stopped: %v\n", err)
	}
}
```

### 3. Abuse Detection examples
```go
import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/its-ernest/echox/cache" 
	"github.com/its-ernest/echox/abuse"
	"github.com/its-ernest/echox/store"
)

func main() {
	e := echo.New()

	abuseStore := store.NewMemoryCounter()
	// Instant-ban on sensitive files, warn on API spam
	e.Use(abuse.New(abuse.Config{
		Store:     abuseStore,
		Threshold: 100,
		Rules: []abuse.Rule{
			// immediate ban for ANY attempt on .env
			{Path: "/.env", Score: 100},

			// heavy score only for POST requests to login (brute force protection)
			{Path: "/api/v1/login", Method: "POST", Score: 20},

			// light score for general API exploration
			{Path: "/api/v1/*", Method: "GET", Score: 2},
			{Path: "/api/v1/users", Method: "PUT", Score: 20},
		},
	}))

	fmt.Println("Starting Dev Server on :8080...")
	if err := e.Start(":8080"); err != nil {
		fmt.Printf("Server stopped: %v\n", err)
	}
}
```

##  License

under the MIT License.
