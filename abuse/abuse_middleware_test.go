package abuse

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/its-ernest/echox/store"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestAbuseMiddleware(t *testing.T) {
	e := echo.New()

	// 1. Setup the Store and Config
	counter := store.NewMemoryCounter()
	config := Config{
		Store:     counter,
		Threshold: 10,
		Rules: []Rule{
			{Path: "/evil", Score: 10},
			{Path: "/suspicious*", Score: 5},
		},
		Cooldown: time.Second * 1,
	}

	mw := New(config)
	handler := mw(func(c *echo.Context) error {
		return c.String(http.StatusOK, "passed")
	})

	t.Run("Clean request should pass", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/safe", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Hitting /evil should trigger immediate block on next request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/evil", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// First hit adds 10 points and blocks (since newScore >= Threshold)
		err := handler(c)
		assert.Error(t, err)
		he := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusForbidden, he.Code)
	})

	t.Run("Cooldown should reset score", func(t *testing.T) {
		// Wait for the 1s cooldown defined in config
		time.Sleep(time.Second * 1)

		req := httptest.NewRequest(http.MethodGet, "/safe", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestTarpitDelay(t *testing.T) {
	e := echo.New()
	counter := store.NewMemoryCounter()

	config := Config{
		Store:       counter,
		Threshold:   100,
		EnableDelay: true,
		Rules:       []Rule{{Path: "/bad", Score: 60}}, // 60 is > 50% of threshold
	}

	mw := New(config)
	handler := mw(func(c *echo.Context) error {
		return c.String(http.StatusOK, "passed")
	})

	t.Run("Should apply delay when heat is high", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/bad", nil)
		c := e.NewContext(req, httptest.NewRecorder())

		start := time.Now()
		_ = handler(c) // First hit adds 60 points

		// Second hit should be delayed
		_ = handler(c)
		duration := time.Since(start)

		//  60 * 10ms = 600ms delay.
		// check if it took at least 500ms.
		assert.True(t, duration >= 500*time.Millisecond, "Request should have been delayed by tarpit")
	})
}
