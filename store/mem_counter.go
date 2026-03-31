package store

import (
	"context"
	"sync"
	"time"
)

type memoryCounter struct {
	mu       sync.RWMutex
	counts   map[string]int
	expiries map[string]time.Time
}

func NewMemoryCounter() Counter {
	return &memoryCounter{
		counts:   make(map[string]int),
		expiries: make(map[string]time.Time),
	}
}

func (m *memoryCounter) Increment(ctx context.Context, key string, n int, ttl time.Duration) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. Check if expired or missing
	if exp, ok := m.expiries[key]; ok && time.Now().After(exp) {
		delete(m.counts, key)
		delete(m.expiries, key)
	}

	// 2. Increment
	m.counts[key] += n

	// 3. Update Expiry (only if it's a new key or you want to "slide" the window)
	if ttl > 0 {
		m.expiries[key] = time.Now().Add(ttl)
	}

	return m.counts[key], nil
}

func (m *memoryCounter) Reset(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.counts, key)
	delete(m.expiries, key)
	return nil
}

func (m *memoryCounter) Get(_ context.Context, key string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// check if expired
	if exp, ok := m.expiries[key]; ok && time.Now().After(exp) {
		return 0, nil // reset if expired
	}

	return m.counts[key], nil
}
