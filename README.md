# echox

[![Go Reference](https://pkg.go.dev/badge/github.com/its-ernest/echox.svg)](https://pkg.go.dev/github.com/its-ernest/echox)
[![Go Report Card](https://goreportcard.com/badge/github.com/its-ernest/echox)](https://goreportcard.com/report/github.com/its-ernest/echox)
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
| **[`echox/cache`](cache/README.md)** | RFC-compliant HTTP caching | `Stable` | <i class="fas fa-check-circle" style="color:green"></i> Memory. <br> Redis (`in progress`). |
| **[`abuse`](abuse/README.md)** | API abuse detection | `Beta` | <i class="fas fa-hourglass-start"></i> Memory. <br> Redis (`in progress`). |


## <i class="fas fa-terminal"></i> Requirements

* **Go:** 1.25+
* **Echo:** v5.0.0 or higher

## MINI PROJECTS EXAMPLES: 

* **OTP verification caching (MemoryStore)**: [_examples/cache/otp_code_verification](_examples/cache/otp_code_verification/README.md)


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

### Abuse Detection examples
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
	// 1. Setup both stores
	cacheStore := store.NewMemoryStore()
	abuseStore := store.NewMemoryCounter() // returns the Counter interface

	// 2. Apply Abuse detection first (like a mini firewall)
	e.Use(abuse.New(abuse.Config{
		Store:     abuseStore,
		Threshold: 100,
		Rules: []abuse.Rule{
			{Path: "/.env", Score: 100},
		},
	}))

	// 3. Apply Cache second
	e.Use(cache.New(cache.Config{
		Store: cacheStore,
		TTL:   10 * time.Second,
	}))
	// rest of code...
}
```

##  License

under the MIT License.
