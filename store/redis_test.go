package store

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupRedisTest(t *testing.T) (*RedisStore, *miniredis.Miniredis) {
	// Start a mock Redis server
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	// Create RedisStore pointing to the mock server
	store := &RedisStore{
		client: redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		}),
	}

	return store, mr
}

func TestRedisStore_SaveAndGet(t *testing.T) {
	store, mr := setupRedisTest(t)
	defer mr.Close()
	defer store.Close()

	ctx := context.Background()
	entry := &Entry{
		Status: 200,
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: []byte(`{"message":"hello"}`),
	}

	// Save an entry
	err := store.Save(ctx, "test-key", entry, time.Minute)
	assert.NoError(t, err)

	// Retrieve the entry
	retrieved, err := store.Get(ctx, "test-key")
	assert.NoError(t, err)
	assert.Equal(t, entry.Status, retrieved.Status)
	assert.Equal(t, entry.Body, retrieved.Body)
}

func TestRedisStore_GetNotFound(t *testing.T) {
	store, mr := setupRedisTest(t)
	defer mr.Close()
	defer store.Close()

	ctx := context.Background()

	// Try to get a non-existent key
	_, err := store.Get(ctx, "non-existent")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestRedisStore_Delete(t *testing.T) {
	store, mr := setupRedisTest(t)
	defer mr.Close()
	defer store.Close()

	ctx := context.Background()
	entry := &Entry{
		Status: 200,
		Body:   []byte("test"),
	}

	// Save and then delete
	err := store.Save(ctx, "test-key", entry, time.Minute)
	assert.NoError(t, err)

	err = store.Delete(ctx, "test-key")
	assert.NoError(t, err)

	// Verify deletion
	_, err = store.Get(ctx, "test-key")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestRedisStore_TTL(t *testing.T) {
	store, mr := setupRedisTest(t)
	defer mr.Close()
	defer store.Close()

	ctx := context.Background()
	entry := &Entry{
		Status: 200,
		Body:   []byte("test"),
	}

	// Save with a short TTL
	err := store.Save(ctx, "test-key", entry, 100*time.Millisecond)
	assert.NoError(t, err)

	// Fast-forward miniredis time
	mr.FastForward(150 * time.Millisecond)

	// Entry should have expired
	_, err = store.Get(ctx, "test-key")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestRedisStore_LockAndUnlock(t *testing.T) {
	store, mr := setupRedisTest(t)
	defer mr.Close()
	defer store.Close()

	ctx := context.Background()

	// Acquire lock
	acquired, err := store.Lock(ctx, "resource-1", time.Minute)
	assert.NoError(t, err)
	assert.True(t, acquired)

	// Try to acquire same lock again (should fail)
	acquired, err = store.Lock(ctx, "resource-1", time.Minute)
	assert.NoError(t, err)
	assert.False(t, acquired)

	// Release lock
	err = store.Unlock(ctx, "resource-1")
	assert.NoError(t, err)

	// Try to acquire lock again (should succeed)
	acquired, err = store.Lock(ctx, "resource-1", time.Minute)
	assert.NoError(t, err)
	assert.True(t, acquired)
}

func TestRedisStore_LockExpiration(t *testing.T) {
	store, mr := setupRedisTest(t)
	defer mr.Close()
	defer store.Close()

	ctx := context.Background()

	// Acquire lock with short TTL
	acquired, err := store.Lock(ctx, "resource-1", 100*time.Millisecond)
	assert.NoError(t, err)
	assert.True(t, acquired)

	// Fast-forward time to expire the lock
	mr.FastForward(150 * time.Millisecond)

	// Lock should have expired, can acquire again
	acquired, err = store.Lock(ctx, "resource-1", time.Minute)
	assert.NoError(t, err)
	assert.True(t, acquired)
}

func TestNewRedisStoreWithOptions(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	// Test with custom options
	store := NewRedisStoreWithOptions(RedisStoreOptions{
		Addr:         mr.Addr(),
		PoolSize:     5,
		MinIdleConns: 2,
		Password:     "",
		DB:           0,
	})
	defer store.Close()

	ctx := context.Background()
	entry := &Entry{
		Status: 200,
		Body:   []byte("test"),
	}

	// Verify it works
	err = store.Save(ctx, "test-key", entry, time.Minute)
	assert.NoError(t, err)

	retrieved, err := store.Get(ctx, "test-key")
	assert.NoError(t, err)
	assert.Equal(t, entry.Body, retrieved.Body)
}