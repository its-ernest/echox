package cache

import (
	"context"
	"time"

	"github.com/its-ernest/echox/store"
)

// acquireLockWithTTL returns (locked bool, unlock function)
func acquireLockWithTTL(ctx context.Context, s store.Store, key string, ttl time.Duration) (bool, func()) {
	locked, err := s.Lock(ctx, key, ttl)
	if err != nil || !locked {
		return false, func() {}
	}

	unlockFn := func() {
		_ = s.Unlock(ctx, key)
	}

	return true, unlockFn
}
