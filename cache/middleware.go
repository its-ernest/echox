package cache

import (
	"crypto/sha256"
	"encoding/hex"
	//"log"
	"net/http"
	"time"

	"github.com/its-ernest/echox/internal"
	"github.com/its-ernest/echox/internal/store" // Crucial for store.Entry
	"github.com/labstack/echo/v5"
)

func New(config Config) echo.MiddlewareFunc {
	// Defaults & sanity checks
	if config.Store == nil {
		panic("echox: cache middleware requires a store")
	}
	if config.TTL == 0 {
		config.TTL = 5 * time.Minute
	}
	if config.KeyGenerator == nil {
		config.KeyGenerator = DefaultKeyGenerator
	}
	if config.MaxBodySize == 0 {
		config.MaxBodySize = 1024 * 1024 // 1MB
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 20 * time.Millisecond
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 5
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if config.Skipper != nil && config.Skipper(c) {
				return next(c)
			}
			if c.Request().Method != http.MethodGet {
				return next(c)
			}

			key := config.KeyGenerator(c)
			lockKey := key + ":lock"
			ctx := c.Request().Context()

			entry, err := config.Store.Get(ctx, key)
			if err == nil && entry != nil {
				return replay(c, entry)
			}

			locked, unlockFn := acquireLockWithTTL(ctx, config.Store, lockKey, 10*time.Second)
			if !locked {
				// wait/retry logic...
				return next(c)
			}
			defer unlockFn()

			originalWriter := c.Response()
			recorder := internal.NewResponseRecorder(originalWriter)
			c.SetResponse(recorder)

			if err := next(c); err != nil {
				return err
			}

			if recorder.Status == http.StatusOK && recorder.Body.Len() <= config.MaxBodySize {
				hash := sha256.Sum256(recorder.Body.Bytes())
				etag := hex.EncodeToString(hash[:]) // CORRECTED METHOD

				newEntry := &store.Entry{
					Status: recorder.Status,
					Header: recorder.Header().Clone(),
					Body:   recorder.Body.Bytes(),
				}

				// CAST map to http.Header to use .Set()
				entryHeader := http.Header(newEntry.Header)
				entryHeader.Set("Etag", etag)
				entryHeader.Set("X-Cache", "MISS")

				_ = config.Store.Save(ctx, key, newEntry, config.TTL)

				c.Response().Header().Set("Etag", etag)
				c.Response().Header().Set("X-Cache", "MISS")
			}

			return nil
		}
	}
}