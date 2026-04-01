package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: These tests require a running Redis server.
// For CI/CD, you can use:
//   - Docker: docker run -d -p 6379:6379 redis:alpine
//   - Or mock tests for unit testing without Redis

const testRedisAddr = "localhost:6379"

// skipIfNoRedis skips the test if Redis is not available
func skipIfNoRedis(t *testing.T, store *RedisStore) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := store.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping test")
	}
}

func TestRedisStore_NewRedisStore(t *testing.T) {
	store := NewRedisStore(testRedisAddr)
	assert.NotNil(t, store)
	assert.NotNil(t, store.client)
}

func TestRedisStore_NewRedisStoreWithConfig(t *testing.T) {
	store := NewRedisStoreWithConfig(RedisStoreConfig{
		Addr:         testRedisAddr,
		Password:     "",
		DB:           0,
		PoolSize:     20,
		MinIdleConns: 10,
	})
	assert.NotNil(t, store)
	assert.NotNil(t, store.client)
}

func TestRedisStore_SaveAndGet(t *testing.T) {
	store := NewRedisStore(testRedisAddr)
	skipIfNoRedis(t, store)
	defer store.Close()

	ctx := context.Background()
	key := "test:save_and_get"
	entry := &Entry{
		Status: 200,
		Header: map[string][]string{"Content-Type": {"application/json"}},
		Body:   []byte(`{"message":"hello"}`),
	}

	// Clean up before test
	store.Delete(ctx, key)

	// Test Save
	err := store.Save(ctx, key, entry, 10*time.Second)
	require.NoError(t, err)

	// Test Get
	got, err := store.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, entry.Status, got.Status)
	assert.Equal(t, entry.Body, got.Body)
	assert.Equal(t, entry.Header["Content-Type"], got.Header["Content-Type"])

	// Clean up
	store.Delete(ctx, key)
}

func TestRedisStore_Get_NotFound(t *testing.T) {
	store := NewRedisStore(testRedisAddr)
	skipIfNoRedis(t, store)
	defer store.Close()

	ctx := context.Background()
	key := "test:not_found"

	// Clean up before test
	store.Delete(ctx, key)

	// Test Get non-existent key
	_, err := store.Get(ctx, key)
	assert.Equal(t, ErrNotFound, err)
}

func TestRedisStore_Delete(t *testing.T) {
	store := NewRedisStore(testRedisAddr)
	skipIfNoRedis(t, store)
	defer store.Close()

	ctx := context.Background()
	key := "test:delete"
	entry := &Entry{
		Status: 200,
		Header: map[string][]string{},
		Body:   []byte("test"),
	}

	// Save entry
	err := store.Save(ctx, key, entry, 10*time.Second)
	require.NoError(t, err)

	// Delete entry
	err = store.Delete(ctx, key)
	require.NoError(t, err)

	// Verify deleted
	_, err = store.Get(ctx, key)
	assert.Equal(t, ErrNotFound, err)
}

func TestRedisStore_TTL(t *testing.T) {
	store := NewRedisStore(testRedisAddr)
	skipIfNoRedis(t, store)
	defer store.Close()

	ctx := context.Background()
	key := "test:ttl"
	entry := &Entry{
		Status: 200,
		Body:   []byte("test"),
	}

	// Save with short TTL
	err := store.Save(ctx, key, entry, 500*time.Millisecond)
	require.NoError(t, err)

	// Should exist immediately
	_, err = store.Get(ctx, key)
	require.NoError(t, err)

	// Wait for TTL to expire
	time.Sleep(600 * time.Millisecond)

	// Should be expired
	_, err = store.Get(ctx, key)
	assert.Equal(t, ErrNotFound, err)
}

func TestRedisStore_Lock(t *testing.T) {
	store := NewRedisStore(testRedisAddr)
	skipIfNoRedis(t, store)
	defer store.Close()

	ctx := context.Background()
	key := "test:lock"

	// Clean up before test
	store.Unlock(ctx, key)

	// First lock should succeed
	acquired, err := store.Lock(ctx, key, 5*time.Second)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Second lock should fail (already locked)
	acquired, err = store.Lock(ctx, key, 5*time.Second)
	require.NoError(t, err)
	assert.False(t, acquired)

	// Unlock
	err = store.Unlock(ctx, key)
	require.NoError(t, err)

	// Should be able to lock again
	acquired, err = store.Lock(ctx, key, 5*time.Second)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Clean up
	store.Unlock(ctx, key)
}

func TestRedisStore_Lock_Expiration(t *testing.T) {
	store := NewRedisStore(testRedisAddr)
	skipIfNoRedis(t, store)
	defer store.Close()

	ctx := context.Background()
	key := "test:lock_expiration"

	// Clean up before test
	store.Unlock(ctx, key)

	// Acquire lock with short TTL
	acquired, err := store.Lock(ctx, key, 500*time.Millisecond)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Wait for lock to expire
	time.Sleep(600 * time.Millisecond)

	// Should be able to acquire lock after expiration
	acquired, err = store.Lock(ctx, key, 5*time.Second)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Clean up
	store.Unlock(ctx, key)
}

func TestRedisStore_Ping(t *testing.T) {
	store := NewRedisStore(testRedisAddr)
	defer store.Close()

	ctx := context.Background()
	err := store.Ping(ctx)
	// This test will fail if Redis is not running, which is expected
	// In a real environment, you'd want Redis running
	if err != nil {
		t.Logf("Ping failed (expected if Redis not running): %v", err)
	}
}