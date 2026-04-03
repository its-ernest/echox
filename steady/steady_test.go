package steady

import (
	"context"
	"net/http"
	"net/http/httptest"

	//"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestSteady(t *testing.T) {
	e := echo.New()

	t.Run("within limits", func(t *testing.T) {
		h := New(Config{MaxConcurrent: 2})(func(c *echo.Context) error {
			return c.String(http.StatusOK, "OK")
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		assert.NoError(t, h(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())
	})

	t.Run("exceeding limits with wait and succeed", func(t *testing.T) {
		barrier := make(chan struct{})
		h := New(Config{
			MaxConcurrent: 1,
			WaitTimeout:   100 * time.Millisecond,
		})(func(c *echo.Context) error {
			<-barrier // Hold the request
			return c.String(http.StatusOK, "OK")
		})

		// First request (will hold the slot)
		go func() {
			req1 := httptest.NewRequest(http.MethodGet, "/", nil)
			rec1 := httptest.NewRecorder()
			c1 := e.NewContext(req1, rec1)
			_ = h(c1)
		}()

		// Give it a moment to take the slot
		time.Sleep(20 * time.Millisecond)

		// Second request (will wait)
		done := make(chan struct{})
		go func() {
			req2 := httptest.NewRequest(http.MethodGet, "/", nil)
			rec2 := httptest.NewRecorder()
			c2 := e.NewContext(req2, rec2)
			err := h(c2)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec2.Code)
			close(done)
		}()

		// Release first request
		close(barrier)

		select {
		case <-done:
			// Success
		case <-time.After(200 * time.Millisecond):
			t.Fatal("second request timed out")
		}
	})

	t.Run("exceeding limits and timeout", func(t *testing.T) {
		h := New(Config{
			MaxConcurrent: 1,
			WaitTimeout:   50 * time.Millisecond,
		})(func(c *echo.Context) error {
			time.Sleep(200 * time.Millisecond) // Longer than wait timeout
			return c.String(http.StatusOK, "OK")
		})

		// First request (will hold the slot)
		go func() {
			req1 := httptest.NewRequest(http.MethodGet, "/", nil)
			rec1 := httptest.NewRecorder()
			c1 := e.NewContext(req1, rec1)
			_ = h(c1)
		}()

		// Give it a moment to take the slot
		time.Sleep(10 * time.Millisecond)

		// Second request (will timeout)
		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)
		err := h(c2)

		assert.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusServiceUnavailable, he.Code)
		assert.Equal(t, "0", rec2.Header().Get("Retry-After")) // 50ms wait timeout rounds down to 0 seconds
	})

	t.Run("client cancellation", func(t *testing.T) {
		h := New(Config{
			MaxConcurrent: 1,
			WaitTimeout:   1 * time.Second,
		})(func(c *echo.Context) error {
			time.Sleep(200 * time.Millisecond)
			return c.String(http.StatusOK, "OK")
		})

		// Hold slot
		go func() {
			req1 := httptest.NewRequest(http.MethodGet, "/", nil)
			rec1 := httptest.NewRecorder()
			c1 := e.NewContext(req1, rec1)
			_ = h(c1)
		}()

		time.Sleep(10 * time.Millisecond)

		// Wait in line
		ctx, cancel := context.WithCancel(context.Background())
		req2 := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)

		// Cancel immediately
		cancel()

		err := h(c2)
		assert.Equal(t, context.Canceled, err)
	})
}
