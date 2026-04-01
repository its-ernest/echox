package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testRedisCounterAddr = "localhost:6379"

// skipIfNoRedisForCounter skips the test if Redis is not available
func skipIfNoRedisForCounter(t *testing.T, counter *RedisCounter) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := counter.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping test")
	}
}

func TestRedisCounter_NewRedisCounter(t *testing.T) {
	counter := NewRedisCounter(testRedisCounterAddr)
	assert.NotNil(t, counter)

	rc := counter.(*RedisCounter)
	assert.NotNil(t, rc.client)
}

func TestRedisCounter_NewRedisCounterWithConfig(t *testing.T) {
	counter := NewRedisCounterWithConfig(RedisCounterConfig{
		Addr:         testRedisCounterAddr,
		Password:     "",
		DB:           0,
		PoolSize:     20,
		MinIdleConns: 10,
	})
	assert.NotNil(t, counter)

	rc := counter.(*RedisCounter)
	assert.NotNil(t, rc.client)
}

func TestRedisCounter_Increment(t *testing.T) {
	counter := NewRedisCounter(testRedisCounterAddr).(*RedisCounter)
	skipIfNoRedisForCounter(t, counter)
	defer counter.Close()

	ctx := context.Background()
	key := "test:counter:increment"

	// Clean up before test
	counter.Reset(ctx, key)

	// First increment
	val, err := counter.Increment(ctx, key, 5, 10*time.Second)
	require.NoError(t, err)
	assert.Equal(t, 5, val)

	// Second increment
	val, err = counter.Increment(ctx, key, 3, 10*time.Second)
	require.NoError(t, err)
	assert.Equal(t, 8, val)

	// Clean up
	counter.Reset(ctx, key)
}

func TestRedisCounter_Get(t *testing.T) {
	counter := NewRedisCounter(testRedisCounterAddr).(*RedisCounter)
	skipIfNoRedisForCounter(t, counter)
	defer counter.Close()

	ctx := context.Background()
	key := "test:counter:get"

	// Clean up before test
	counter.Reset(ctx, key)

	// Get non-existent key should return 0
	val, err := counter.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, 0, val)

	// Increment
	counter.Increment(ctx, key, 10, 10*time.Second)

	// Get should return 10
	val, err = counter.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, 10, val)

	// Clean up
	counter.Reset(ctx, key)
}

func TestRedisCounter_Reset(t *testing.T) {
	counter := NewRedisCounter(testRedisCounterAddr).(*RedisCounter)
	skipIfNoRedisForCounter(t, counter)
	defer counter.Close()

	ctx := context.Background()
	key := "test:counter:reset"

	// Increment
	counter.Increment(ctx, key, 100, 10*time.Second)

	// Verify value
	val, err := counter.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, 100, val)

	// Reset
	err = counter.Reset(ctx, key)
	require.NoError(t, err)

	// Verify reset
	val, err = counter.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, 0, val)
}

func TestRedisCounter_TTL(t *testing.T) {
	counter := NewRedisCounter(testRedisCounterAddr).(*RedisCounter)
	skipIfNoRedisForCounter(t, counter)
	defer counter.Close()

	ctx := context.Background()
	key := "test:counter:ttl"

	// Clean up before test
	counter.Reset(ctx, key)

	// Increment with short TTL
	val, err := counter.Increment(ctx, key, 5, 500*time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, 5, val)

	// Verify value exists
	val, err = counter.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, 5, val)

	// Wait for TTL to expire
	time.Sleep(600 * time.Millisecond)

	// Value should be gone (or 0)
	val, err = counter.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, 0, val)
}

func TestRedisCounter_MultipleKeys(t *testing.T) {
	counter := NewRedisCounter(testRedisCounterAddr).(*RedisCounter)
	skipIfNoRedisForCounter(t, counter)
	defer counter.Close()

	ctx := context.Background()
	key1 := "test:counter:multi1"
	key2 := "test:counter:multi2"

	// Clean up
	counter.Reset(ctx, key1)
	counter.Reset(ctx, key2)

	// Increment different keys
	val1, err := counter.Increment(ctx, key1, 10, 10*time.Second)
	require.NoError(t, err)
	assert.Equal(t, 10, val1)

	val2, err := counter.Increment(ctx, key2, 20, 10*time.Second)
	require.NoError(t, err)
	assert.Equal(t, 20, val2)

	// Verify they're independent
	val1, err = counter.Get(ctx, key1)
	require.NoError(t, err)
	assert.Equal(t, 10, val1)

	val2, err = counter.Get(ctx, key2)
	require.NoError(t, err)
	assert.Equal(t, 20, val2)

	// Clean up
	counter.Reset(ctx, key1)
	counter.Reset(ctx, key2)
}

func TestRedisCounter_Ping(t *testing.T) {
	counter := NewRedisCounter(testRedisCounterAddr).(*RedisCounter)
	defer counter.Close()

	ctx := context.Background()
	err := counter.Ping(ctx)
	// This test will fail if Redis is not running, which is expected
	if err != nil {
		t.Logf("Ping failed (expected if Redis not running): %v", err)
	}
}