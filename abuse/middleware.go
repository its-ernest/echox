// Package abuse provides behavioral analysis middleware for Echo v5.
// It tracks client reputation via a scoring system to mitigate scrapers and scanners.
package abuse

import (
	"net/http"
	"time"

	"github.com/its-ernest/echox/store"
	"github.com/labstack/echo/v5"
)

type (
	// Rule defines the "heat" points added for specific path patterns.
	// Pattern supports suffix wildcards like "/admin*".
	Rule struct {
		// Path is the URL path to match (supports wildcards via Echo's path matching)
		Path string

		// Method is the HTTP method to match (GET, POST, etc.). Empty means match all.
		Method string

		// Score is the amount of *heat* to add when this rule matches
		Score int
	}

	// Config defines the settings for the abuse detection middleware.
	Config struct {
		Store     store.Counter
		Threshold int
		Rules     []Rule
		Cooldown  time.Duration
		// EnableDelay adds latency to suspicious requests instead of a hard block.
		EnableDelay  bool
		ErrorHandler func(c *echo.Context, score int) error
	}
)

var DefaultConfig = Config{
	Threshold: 100,
	Cooldown:  30 * time.Minute,
	ErrorHandler: func(c *echo.Context, score int) error {
		return echo.NewHTTPError(http.StatusForbidden, "Access denied due to suspicious activity")
	},
}

// New returns an Echo middleware that monitors and limits abusive behavior.
// It should be placed early in the middleware stack to protect downstream resources.
func New(config Config) echo.MiddlewareFunc {
	// 1. Panic if store is missing
	if config.Store == nil {
		panic("echox/abuse: Store is required")
	}

	// 2. Apply defaults
	if config.Cooldown == 0 {
		config.Cooldown = DefaultConfig.Cooldown
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = DefaultConfig.ErrorHandler
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx := c.Request().Context()
			clientIP := c.RealIP()
			key := "abuse:heat:" + clientIP

			// 1. Fetch current heat
			currentScore, _ := config.Store.Get(ctx, key)

			// 2. Tarpit Logic: Slow client down if getting "warm"
			if config.EnableDelay && currentScore > (config.Threshold/2) {
				// Linear delay: more heat = more wait
				// e.g., if threshold is 100 and score is 75, wait 750ms
				delay := time.Duration(currentScore) * 10 * time.Millisecond
				time.Sleep(delay)
			}

			// 3. Immediate Block: If toasted, don't even look at rules
			if currentScore >= config.Threshold {
				return config.ErrorHandler(c, currentScore)
			}

			// 4. Pre-Handler Path Analysis
			requestPath := c.Request().URL.Path
			penalty := 0
			for _, rule := range config.Rules {

				// check and match exact request method if any is defined in rule
				if rule.Method != "" && rule.Method != c.Request().Method {
					continue
				}

				if matchPath(rule.Path, requestPath) {
					penalty += rule.Score
				}
			}

			// 5. Atomic Update and Final Enforcement
			if penalty > 0 {
				newScore, err := config.Store.Increment(ctx, key, penalty, config.Cooldown)
				if err == nil && newScore >= config.Threshold {
					return config.ErrorHandler(c, newScore)
				}
			}

			return next(c)
		}
	}
}
