package store

import (
	"context"
	"sync"
	"time"
)

type item struct {
	entry      *Entry
	expiresAt  time.Time
}

type MemoryStore struct {
	data  sync.Map
	locks sync.Map
}

type lockItem struct {
	acquiredAt time.Time
	ttl        time.Duration
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

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

func (m *MemoryStore) Save(_ context.Context, key string, entry *Entry, ttl time.Duration) error {
	m.data.Store(key, item{
		entry:     entry,
		expiresAt: time.Now().Add(ttl),
	})
	return nil
}

func (m *MemoryStore) Delete(_ context.Context, key string) error {
	m.data.Delete(key)
	return nil
}

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

func (m *MemoryStore) Unlock(_ context.Context, key string) error {
	m.locks.Delete(key)
	return nil
}

func (m *MemoryStore) StartEvictor(interval time.Duration, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.data.Range(func(k, v interface{}) bool {
					item := v.(item)
					if time.Now().After(item.expiresAt) {
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