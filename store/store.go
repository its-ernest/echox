package store

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrNotFound is returned when a requested key does not exist in the store
	// or has already expired. Middleware should typically handle this by
	// proceeding to the next handler to fetch fresh data.
	ErrNotFound = errors.New("store: item not found")

	// ErrLockFailed is returned when a distributed lock cannot be acquired
	// within the allowed timeframe or because it is already held by another
	// process. This is primarily used by the idempotency middleware.
	ErrLockFailed = errors.New("store: could not acquire lock")
)

// Entry represents the data that actually persists
type Entry struct {
	Status int                 `json:"status"`
	Header map[string][]string `json:"header"`
	Body   []byte              `json:"body"`
}

// Store defines the contract for any backend (Memory, Redis, etc.)
type Store interface {
	// Get retrieves an entry by key
	Get(ctx context.Context, key string) (*Entry, error)

	// Save stores an entry with a specific TTL
	Save(ctx context.Context, key string, entry *Entry, ttl time.Duration) error

	// Delete removes an entry
	Delete(ctx context.Context, key string) error

	// Lock attempts to acquire a distributed lock for a key (crucial for Idempotency)
	Lock(ctx context.Context, key string, ttl time.Duration) (bool, error)

	// Unlock releases the lock
	Unlock(ctx context.Context, key string) error
}

// Counter defines a simpler interface for numeric tracking (Abuse, Rate Limiting)
type Counter interface {
	// Get returns the current count for a key
	Get(ctx context.Context, key string) (int, error)

	// Increment adds n to the key and returns the new total.
	// if the key doesn't exist, it's created with value 'n' and the given TTL.
	Increment(ctx context.Context, key string, n int, ttl time.Duration) (int, error)

	// Reset clears the counter
	Reset(ctx context.Context, key string) error
}
