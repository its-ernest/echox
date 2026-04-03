# echox Steady Middleware

A **high-performance concurrency limiter** for [Echo](https://echo.labstack.com/) v5. 

[→ View Full API Documentation](DOCS.md)

`steady` ensures your server remains stable under heavy load by limiting the number of requests processed simultaneously. Instead of immediately rejecting excess traffic, it allows requests to wait in a queue for a configurable amount of time before returning a "Service Unavailable" response.

## Features

* **Concurrency Control:** Strictly limit the number of active goroutines handling requests.
* **Smart Queueing:** Requests wait for a slot instead of being dropped immediately.
* **Configurable Timeouts:** Define how long a request should wait before giving up.
* **Custom Error Handling:** Fully customize the response sent when the limit is exceeded.
* **Context Awareness:** Automatically stops waiting if the client cancels the request.
* **Echo v5 Ready:** Designed specifically for the latest Echo *pointer-based* context.

---

## Installation

```bash
go get github.com/its-ernest/echox/steady
```

---

## Basic Usage

```go
package main

import (
	"net/http"
	"time"

	"github.com/its-ernest/echox/steady"
	"github.com/labstack/echo/v5"
)

func main() {
	e := echo.New()

	// Apply steady middleware with custom limits
	e.Use(steady.New(steady.Config{
		MaxConcurrent: 50,               // Allow 50 simultaneous requests
		WaitTimeout:   10 * time.Second, // Wait up to 10s for a slot
	}))

	e.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Stable and Steady!")
	})

	e.Start(":8080")
}
```

---

## Configuration Options

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `MaxConcurrent` | `int` | `100` | Maximum number of requests to process at once. |
| `WaitTimeout` | `time.Duration` | `10s` | How long a request will wait in the queue for a slot. |
| `Skipper` | `middleware.Skipper` | `DefaultSkipper` | Function to skip this middleware for certain routes. |
| `ErrorHandler` | `func(*echo.Context) error` | *Internal* | Custom logic to run when a request times out in the queue. |

---

## How it Works

1. **Slot Acquisition:** When a request arrives, the middleware tries to acquire a "slot" from an internal semaphore.
2. **Immediate Execution:** If a slot is available, the request proceeds immediately.
3. **Queueing:** If the server is at `MaxConcurrent` capacity, the request waits until:
    - A slot becomes free (as another request finishes).
    - The `WaitTimeout` is reached.
    - The client closes the connection (Context cancellation).
4. **Timeout:** If the `WaitTimeout` is reached before a slot is freed, the `ErrorHandler` is triggered (defaulting to `503 Service Unavailable` with a `Retry-After` header).

---

## Why use Steady?

Standard rate limiters often focus on requests-per-second (RPS). However, some requests are "heavier" than others (e.g., intensive DB queries or image processing). `steady` focuses on **concurrency**, ensuring that no matter how slow your handlers are, you never exceed the resource limits (CPU/Memory) of your server.
