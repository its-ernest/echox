# echox Cache Middleware

A **high-performance, production-ready caching middleware** for [Echo](https://echo.labstack.com/) in Go.
[→ View Full API Documentation](DOCS.md)

## Features

* **Concurrent-safe:** Uses `sync.Map` for high-concurrency environments.
* **Cache stampede prevention:** Only one request computes the cache per key; others wait or retry.
* **ETag & Not-Modified:** Supports client-side caching validation.
* **Configurable:** TTL, max body size, retry behavior, skip functions, key generators.
* **Automatic cleanup:** Periodic eviction of expired entries prevents memory bloat.
* **Production-ready logging:** Tracks hits, misses, and cache events.

---

## Installation

```bash
# since this is a new push
GOPROXY=direct go get github.com/its-ernest/echox/cache@latest

# in the long run
go get github.com/its-ernest/echox/cache
```

---

## Basic Usage

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

	// SIMPLE CUSTOM LOGGER
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			err := next(c)
			stop := time.Now()
			fmt.Printf("[%s] %d %s %s (%s)\n", 
				stop.Format("15:04:05"),
				c.Request().Method, 
				c.Request().URL.Path,
				stop.Sub(start),
			)
			return err
		}
	})

	// Setup local store
	memStore := store.NewMemoryStore()

	// Apply cache middleware
	e.Use(cache.New(cache.Config{
		Store: memStore,
		TTL:   20 * time.Second,
	}))

	// Echo V5 Handler
	e.GET("/test", func(c *echo.Context) error {
		fmt.Println(" [HANDLER] Generating fresh content...")
		timestamp := time.Now().Format(time.RFC3339Nano)
		return c.String(http.StatusOK, fmt.Sprintf("Generated at: %s", timestamp))
	})

	// Start
	fmt.Println("Starting Dev Server on :8080...")
	if err := e.Start(":8080"); err != nil {
		fmt.Printf("Server stopped: %v\n", err)
	}
}
```

---

## Configuration Options

| Field          | Type                        | Default               | Description                                        |
| -------------- | --------------------------- | --------------------- | -------------------------------------------------- |
| `Store`        | `store.Store`               | **Required**          | Backend cache store (`MemoryStore` or custom).     |
| `TTL`          | `time.Duration`             | `5 * time.Minute`     | Time-to-live for cached entries.                   |
| `Skipper`      | `func(echo.Context) bool`   | `nil`                 | Function to skip caching for specific requests.    |
| `KeyGenerator` | `func(echo.Context) string` | `DefaultKeyGenerator` | Function to generate cache keys.                   |
| `MaxBodySize`  | `int`                       | 1MB                   | Maximum response size to cache.                    |
| `RetryDelay`   | `time.Duration`             | 20ms                  | Delay between retries if cache is being generated. |
| `MaxRetries`   | `int`                       | 5                     | Max retry attempts for waiting requests.           |

---

## Middleware Behavior

1. **Cache Check:** Middleware first checks the cache.
2. **Cache Hit:** Returns cached response with `X-Cache: HIT`. ETag validated; returns 304 if client has latest version.
3. **Cache Miss:** Acquires lock for this key (TTL enforced). Only one request builds the cache.
4. **Other Concurrent Requests:** Retry loop waits for cache to be ready; fallback to origin handler if lock persists.
5. **Cache Save:** Stores response with TTL and generates ETag, sets `X-Cache: MISS`.
6. **Eviction:** Expired items removed automatically by the background evictor.

---

## `MemoryStore` Advanced Features

* **TTL-aware locks:** Locks expire automatically to prevent deadlocks.
* **Background eviction:** Periodic cleanup of expired cache entries.
* **Concurrent-safe operations:** All methods (`Get`, `Save`, `Delete`, `Lock`, `Unlock`) are goroutine-safe.
* **Optional singleflight / GetOrCompute** can be added for fully atomic cache population.

---

## HTTP Headers

* `X-Cache`: `HIT` or `MISS`
* `Etag`: SHA256 hash of response body
* `Content-Type`: Respected from handler or auto-detected

---

## Logging

* `Cache HIT` / `Cache MISS` events logged via `log.Printf`.
* Fallbacks due to lock contention are logged.
* Can be replaced with a structured logger in production.

---

## Recommended practices:

* **Memory Usage:** `MemoryStore` is in-memory; consider limits for very large payloads.
* **Concurrency:** Tested with high concurrent GET requests; stampede prevention enabled.
* **Persistence:** [For multi-server deployments, replace `MemoryStore` with a distributed store (Redis, Memcached)](cache/redis.md).
* **TTL Enforcement:** Lock TTL must match cache TTL strategy to prevent stale locks.
* **Evictor Interval:** Adjust based on traffic, memory footprint, and TTLs.
