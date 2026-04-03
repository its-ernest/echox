# echox

[![Go Reference](https://pkg.go.dev/badge/github.com/its-ernest/echox.svg)](https://pkg.go.dev/github.com/its-ernest/echox)
[![Go Report Card](https://goreportcard.com/badge/github.com/its-ernest/echox)](https://goreportcard.com/report/github.com/its-ernest/echox)
[![Dagger Docs](https://github.com/its-ernest/echox/actions/workflows/docs.yml/badge.svg)](https://github.com/its-ernest/echox/actions/workflows/docs.yml)
[![Build Status](https://github.com/its-ernest/echox/actions/workflows/go.yml/badge.svg)](https://github.com/its-ernest/echox/actions/workflows/go.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**High-performance middleware suite for the Echo v5 ecosystem.**

`echox` is a collection of production-ready middlewares engineered specifically for Echo v5's struct-pointer architecture. By leveraging Go 1.24+ features and standard library interfaces, `echox` provides robust solutions for distributed caching, idempotency, and request lifecycle management.


## Installation

```bash
# latest versions
GOPROXY=direct go get github.com/its-ernest/echox@latest
GOPROXY=direct go get github.com/its-ernest/echox/cache@latest
GOPROXY=direct go get github.com/its-ernest/echox/abuse@latest

# in the long run
go get github.com/its-ernest/echox/cache
```

## Framework Compatibility

To ensure stability across the evolving Echo v5 landscape, we maintain specific versions mapped to upstream Echo releases.

| Echo Framework | Recommended echox Version | Git Tag / Branch | Status |
| :--- | :--- | :--- | :--- |
| **v5.0.0-beta.x** | `v0.0.x` | `v5.0.x`/ `main` | <i class="fas fa-archive"></i> Active |
| **v5.1.0+** | `v0.1.x` | `v5.1.x` | <i class="fas fa-rocket"></i> Planned |

## <i class="fas fa-cubes"></i> Middleware Registry

| Module | Purpose | Status | Backend Support |
| :--- | :--- | :--- | :--- |
| **[`echox/cache`](cache/README.md)** | RFC-compliant HTTP caching | `Stable` | - `Memory`. <br> - `Redis`. |
| **[`echox/abuse`](abuse/README.md)** | API abuse detection | `Stable` | - `Memory`. <br> `Redis`. |
| **[`echox/steady`](steady/README.md)** | Concurrency limit and backpressure | `Beta` | `Channel-based`|


## Requirements

* **Go:** 1.25+
* **Echo:** v5.0.0 or higher


## To Contribute

- Read [CONTRIBUTION.md](CONTRIBUTING.md)
- Contributing on cache middleware: [cache/README.md](cache/README.md)
- Contributing on API abuse middleware: [abuse/README.md](abuse/README.md)
- Contributing on Steady/Semaphore middleware: [steady/README.md](steady/README.md)

## MINI PROJECTS EXAMPLES: 

* **OTP verification caching (MemoryStore)**: [_examples/cache/otp_code_verification](_examples/cache/otp_code_verification/README.md)


## Quick Start examples
### 1. Caching examples

```go

	// ... 
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
	// rest of code...
```

### 2. Using Redis for Distributed Caching

For production environments with multiple application instances, use RedisStore:

```go
	// ...
	e := echo.New()

	// setup Redis store for distributed caching
	redisStore := store.NewRedisStore("localhost:6379")
	defer redisStore.Close()

	// or with custom options:
	// redisStore := store.NewRedisStoreWithOptions(store.RedisStoreOptions{
	// 	Addr:         "localhost:6379",
	// 	PoolSize:     10,
	// 	MinIdleConns: 5,
	// })

	// apply echox cache middleware
	e.Use(cache.New(cache.Config{
		Store:       redisStore,
		TTL:         5 * time.Minute,
		MaxBodySize: 1 * 1024 * 1024, // 1MB limit for Redis efficiency
	}))

	// Echo V5 handler
	e.GET("/test", func(c *echo.Context) error {
		fmt.Println(" [HANDLER] Generating fresh content...")
		timestamp := time.Now().Format(time.RFC3339Nano)
		return c.String(http.StatusOK, fmt.Sprintf("Generated at: %s", timestamp))
	})
	// rest ofo code...
```


### Abuse Detection examples
```go
	// ...
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
	// rest of code...
```

### Concurrency Limiting examples
```go
	// ...
	e := echo.New()

	e.Use(steady.New(steady.Config{
		MaxConcurrent: 50,               // Allow 50 simultaneous requests
		WaitTimeout:   10 * time.Second, // Wait up to 10s for a slot
	}))

	e.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Stable and Steady!")
	})
	// rest of code...
```


## Contributors

<a href="https://github.com/its-ernest/echox/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=its-ernest/echox" />
</a>

Made with [contrib.rocks](https://contrib.rocks).


##  License

under the MIT License.
