// Package store provides the underlying persistence and locking mechanisms
// for the echox middleware suite.
package store

import (
	"context"
	"sync"
	"time"
)

// item represents a cached value and its expiration deadline.
type item struct {
	entry     *Entry
	expiresAt time.Time
}

// MemoryStore is an in-memory implementation of the Store interface.
// It utilizes sync.Map for thread-safe data operations and distributed-style
// locking simulations, making it ideal for single-instance applications.
type MemoryStore struct {
	data  sync.Map
	locks sync.Map
}

// lockItem stores metadata about an active mutex-like lock.
type lockItem struct {
	acquiredAt time.Time
	ttl        time.Duration
}

// NewMemoryStore initializes and returns a new pointer to a MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

// Get retrieves a cached Entry by its key. If the entry exists but has
// expired, it is deleted and ErrNotFound is returned.
func (m *MemoryStore) Get(_ context.Context, key string) (*Entry, error) {
	val, ok := m.data.Load(key)
	if !ok {
		return nil, ErrNotFound
	}

	i := val.(item)
	if time.Now().After(i.expiresAt) {
		m.data.Delete(key)
		return nil, ErrNotFound
	}
	return i.entry, nil
}

// Save persists an Entry into memory with a specific time-to-live (TTL).
func (m *MemoryStore) Save(_ context.Context, key string, entry *Entry, ttl time.Duration) error {
	m.data.Store(key, item{
		entry:     entry,
		expiresAt: time.Now().Add(ttl),
	})
	return nil
}

// Delete removes a specific entry from the memory store immediately.
func (m *MemoryStore) Delete(_ context.Context, key string) error {
	m.data.Delete(key)
	return nil
}

// Lock attempts to acquire a non-blocking lock for a given key.
// It returns true if the lock was successfully acquired or if a previous
// lock has already expired. This is essential for preventing "thundering herd"
// issues in idempotency middleware.
func (m *MemoryStore) Lock(_ context.Context, key string, ttl time.Duration) (bool, error) {
	now := time.Now()

	val, loaded := m.locks.LoadOrStore(key, lockItem{
		acquiredAt: now,
		ttl:        ttl,
	})

	if loaded {
		existing := val.(lockItem)
		if now.Sub(existing.acquiredAt) > existing.ttl {
			// lock expired, overwrite
			m.locks.Store(key, lockItem{acquiredAt: now, ttl: ttl})
			return true, nil
		}
		return false, nil
	}
	return true, nil
}

// Unlock releases the lock for the given key, allowing other processes
// to acquire it.
func (m *MemoryStore) Unlock(_ context.Context, key string) error {
	m.locks.Delete(key)
	return nil
}

// StartEvictor launches a background goroutine that periodically scans the
// store and removes expired entries. The goroutine stops when the provided
// stop channel is closed.
func (m *MemoryStore) StartEvictor(interval time.Duration, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.data.Range(func(k, v interface{}) bool {
					i := v.(item)
					if time.Now().After(i.expiresAt) {
						m.data.Delete(k)
					}
					return true
				})
			case <-stop:
				return
			}
		}
	}()
}
