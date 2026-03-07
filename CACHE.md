# EchoX Cache Middleware

A **high-performance, production-ready caching middleware** for [Echo](https://echo.labstack.com/) in Go, with:

* In-memory cache store (`MemoryStore`)
* TTL-aware locking to prevent cache stampedes
* ETag support and HTTP 304 handling
* Max body size protection
* Automatic background eviction of expired entries
* Optional custom cache key generation and skip logic

---

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
go get github.com/its-ernest/echox/cache
```

---

## Basic Usage

```go
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
	"github.com/its-ernest/echox/store"
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

## Production Considerations

* **Memory Usage:** `MemoryStore` is in-memory; consider limits for very large payloads.
* **Concurrency:** Tested with high concurrent GET requests; stampede prevention enabled.
* **Persistence:** [For multi-server deployments, replace `MemoryStore` with a distributed store (Redis, Memcached)](redis.md).
* **TTL Enforcement:** Lock TTL must match cache TTL strategy to prevent stale locks.
* **Evictor Interval:** Adjust based on traffic, memory footprint, and TTLs.

---

## License

MIT © 2026