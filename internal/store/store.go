package store

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound   = errors.New("store: item not found")
	ErrLockFailed = errors.New("store: could not acquire lock")
)

// Entry represents the data we actually persist
type Entry struct {
	Status  int               `json:"status"`
	Header  map[string][]string `json:"header"`
	Body    []byte            `json:"body"`
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