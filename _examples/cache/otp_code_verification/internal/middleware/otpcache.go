package middleware

import (
	"time"

	"github.com/its-ernest/echox/cache"
	"github.com/its-ernest/echox/store"
	"github.com/labstack/echo/v5"
)

// OTPCache uses my echox MemoryStore to prevent OTP spamming
func OTPCache(store store.Store) echo.MiddlewareFunc {
	return cache.New(cache.Config{
		Store:       store,
		TTL:         5 * time.Minute, // expire after 5 mins
		MaxBodySize: 1024 * 10,       // 10KB safety limit
		// key-generator for cache values
		// this one uses a generator specific to a phone number.
		//for e.g, 'otp_lock:+23324512XXXX', 'otp_lock:+23350348XXXX'
		// each of the keys above is assigned a distinct generated values from other parts of the code
		KeyGenerator: func(c *echo.Context) string {
			// generic helper to extract the phone number
			phone, _ := echo.FormValue[string](c, "phone")
			return "otp_lock:" + phone
		},
	})
}
