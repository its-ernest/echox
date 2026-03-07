package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"github.com/its-ernest/echox/cache/internal"
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
		config.MaxBodySize = 1024 * 1024
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 20 * time.Millisecond
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 5
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper != nil && config.Skipper(c) {
				return next(c)
			}
			if c.Request().Method != http.MethodGet {
				return next(c)
			}

			key := config.KeyGenerator(c)
			lockKey := key + ":lock"
			ctx := c.Request().Context()

			// attempt cache hit 
			entry, err := config.Store.Get(ctx, key)
			if err == nil && entry != nil {
				log.Printf("Cache HIT for key %s", key)
				return replay(c, entry)
			}

			// cache stampede protection 
			locked, unlockFn := acquireLockWithTTL(ctx, config.Store, lockKey, 10*time.Second)
			if !locked {
				// wait and retry until cache is available or retries exhausted
				for i := 0; i < config.MaxRetries; i++ {
					time.Sleep(config.RetryDelay)
					entry, err := config.Store.Get(ctx, key)
					if err == nil && entry != nil {
						log.Printf("Cache HIT after wait for key %s", key)
						return replay(c, entry)
					}
				}
				log.Printf("Lock busy, fallback to handler for key %s", key)
				return next(c)
			}
			defer unlockFn()

			// cache miss: record response 
			originalWriter := c.Response().Writer
			recorder := internal.NewResponseRecorder(originalWriter)
			c.Response().Writer = recorder

			if err := next(c); err != nil {
				return err
			}

			//  persist cache 
			if recorder.Status == http.StatusOK && recorder.Body.Len() <= config.MaxBodySize {
				hash := sha256.Sum256(recorder.Body.Bytes())
				etag := hex.EncodeToString(hash[:])

				newEntry := &store.Entry{
					Status: recorder.Status,
					Header: recorder.Header().Clone(),
					Body:   recorder.Body.Bytes(),
				}

				newEntry.Header.Set("Etag", etag)
				newEntry.Header.Set("X-Cache", "MISS")

				if err := config.Store.Save(ctx, key, newEntry, config.TTL); err != nil {
					log.Printf("Cache save failed for key %s: %v", key, err)
				} else {
					log.Printf("Cache MISS saved for key %s", key)
				}

				// apply headers for current response
				c.Response().Header().Set("Etag", etag)
				c.Response().Header().Set("X-Cache", "MISS")
				if c.Response().Header().Get("Content-Type") == "" {
					c.Response().Header().Set("Content-Type", http.DetectContentType(recorder.Body.Bytes()))
				}
			}

			return nil
		}
	}
}