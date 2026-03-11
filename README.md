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
| **v5.0.0-beta.x** | `v0.1.x` | `v5.0-legacy` | <i class="fas fa-archive"></i> Maintenance |
| **v5.1.0+** | `v0.2.x` | `main` | <i class="fas fa-rocket"></i> Active |

## <i class="fas fa-cubes"></i> Middleware Registry

| Module | Purpose | Status | Backend Support |
| :--- | :--- | :--- | :--- |
| **[`echox/cache`](cache/README.md)** | RFC-compliant HTTP caching | `Stable` | <i class="fas fa-check-circle" style="color:green"></i> Memory <br> <i class="fas fa-exclamation-triangle" style="color:orange"></i> Redis (Alpha) |
| **[`echox/idempotency`](cache/README.md)** | Safe POST/PATCH retries | `In-Dev` | <i class="fas fa-vial"></i> Internal Testing |
| **`ratelimit`** | Traffic shaping & limiting | `Planned` | <i class="fas fa-hourglass-start"></i> Researching |



## <i class="fas fa-microchip"></i> Core Features

* **<i class="fas fa-shield-alt"></i> Anti-Stampede Protection:** Atomic locking prevents "thundering herd" issues on cache misses.
* **<i class="fas fa-exchange-alt"></i> HTTP/1.1 ETag Support:** Native validation of `If-None-Match` for bandwidth optimization.
* **<i class="fas fa-database"></i> Pluggable Storage:** Unified `Store` interface allows sharing a single Redis/Memory instance across multiple middlewares.
* **<i class="fas fa-code-branch"></i> Echo v5 Optimized:** Full support for `*echo.Context` and `slog` structured logging.

## <i class="fas fa-terminal"></i> Requirements

* **Go:** 1.24+
* **Echo:** v5.0.0-beta.x or higher

## <i class="fas fa-play-circle"></i> Quick Start (Cache)

```go
package main

import (
	"time"

	"[github.com/labstack/echo/v5](https://github.com/labstack/echo/v5)"
	"[github.com/its-ernest/echox/cache](https://github.com/its-ernest/echox/cache)"
	"[github.com/its-ernest/echox/internal/store](https://github.com/its-ernest/echox/internal/store)"
)

func main() {
	e := echo.New()

	// 1. Initialize the storage backend
	memStore := store.NewMemoryStore()

	// 2. Register Cache Middleware
	e.Use(cache.New(cache.Config{
		Store: memStore,
		TTL:   10 * time.Minute,
	}))

	e.GET("/data", func(c *echo.Context) error {
		return c.String(200, "Cached in Echo v5")
	})

	e.Start(":8080")
}
```

##  License

Distributed under the MIT License.
