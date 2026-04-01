# echox Abuse Middleware

[![Docs](https://img.shields.io/badge/docs-api_reference-blue?style=flat-square)](./DOCS.md)
[![Status](https://img.shields.io/badge/status-production--beta-cyan?style=flat-square)](#)

A **smart behavioral firewall** for [Echo v5](https://echo.labstack.com/). `echox/abuse` tracks client reputation using a **Heat** scoring system to identify and neutralize scrapers, scanners, and malicious actors.

[**→ View API eference**](./DOCS.md)

---

## How It Works:

Unlike standard rate-limiters that reset every minute, think of this like a **heat** meter. While normal limits reset quickly, this middleware remembers a user's actions over time. Every bad request adds points that only disappear after the user stops for a while.

1.  **Passive Monitoring:** Every request is checked against a list of `Rules`.
2.  **Penalty Scoring:** If a client hits a sensitive path (e.g., `/.env` or `/phpmyadmin`), their "Heat" score increases.
3.  **Tarpit delay (Optional):** As the score approaches the threshold, the middleware artificially delays responses, wasting the attacker's time and resources.
4.  **The Block:** Once the `Threshold` is met, all further requests are rejected with a `403 Forbidden` until the `Cooldown` period expires.


---

## Installation

```bash
go get github.com/its-ernest/echox/abuse
```

---

## Basic Usage

```go
package main

import (
	"time"
	"github.com/labstack/echo/v5"
	"github.com/its-ernest/echox/abuse"
	"github.com/its-ernest/echox/store"
)

func main() {
	e := echo.New()

	// 1. Setup a Counter store (separate from Cache store for performance)
	abuseStore := store.NewMemoryCounter()

	// 2. Configure the Firewall
	e.Use(abuse.New(abuse.Config{
		Store:     abuseStore,
		Threshold: 100,        // Max heat before a total ban
		Cooldown:  1 * time.Hour, // Until forgetting about a ban
		EnableDelay: true,     // Slow down warm IPs
		Rules: []abuse.Rule{
			{Path: "/.env", Score: 100},       // Instant ban
			{Path: "/.git*", Score: 100},      // Instant ban
			{Path: "/wp-admin*", Score: 50},    // Two hits and the IP is banned
			{Path: "/api/v1/login", Score: 5},  // Penalize brute-force attempts
		},
	}))

	e.Start(":8080")
}
```

---

## Configuration Options

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `Store` | `store.Counter` | **Required** | Atomic counter backend (Memory or Redis). |
| `Threshold` | `int` | `100` | The score at which a client is fully blocked. |
| `Rules` | `[]Rule` | `nil` | List of paths and their associated penalty scores. |
| `Cooldown` | `time.Duration` | `30m` | Time before a client's heat starts to reset. |
| `EnableDelay` | `bool` | `false` | If true, applies a linear delay as heat rises. |
| `ErrorHandler`| `func(*echo.Context, int)` | `Default` | Custom logic for when a user is blocked. |

---

## delay effect

When `EnableDelay` is active, the middleware calculates a sleep duration based on the current score:
$$Delay = CurrentScore \times 10ms$$

If a user has **80 points** of heat, every request they make will be artificially slowed down by **800ms**. This makes your API an expensive and frustrating target for automated crawlers without affecting legitimate users.

---

## Recommended Practices

* **Order Matters:** Place this middleware **above** your Logger and **above** your Cache. There is no point in logging or caching a request from a known malicious actor.
* **Real IP:** Ensure your `echo.IPExtractor` is configured correctly if you are running behind Nginx, Cloudflare, or AWS ALB.
* **Gradual Penalties:** Use small scores (2-5) for general API endpoints to catch high scrapers, and large scores (50-100) for sensitive system files.
