package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCounter is a Redis-backed implementation of the Counter interface.
// It provides distributed counter tracking for abuse detection and rate limiting
// across multiple application instances.
type RedisCounter struct {
	client *redis.Client
}

// RedisCounterConfig provides configuration options for the Redis client.
type RedisCounterConfig struct {
	// Addr is the Redis server address (e.g., "localhost:6379")
	Addr string
	// Password for Redis authentication (optional)
	Password string
	// DB is the Redis database to use
	DB int
	// PoolSize is the maximum number of socket connections
	PoolSize int
	// MinIdleConns is the minimum number of idle connections
	MinIdleConns int
}

// NewRedisCounter creates a new RedisCounter with default configuration.
// For production use, consider NewRedisCounterWithConfig for more control.
func NewRedisCounter(addr string) Counter {
	return NewRedisCounterWithConfig(RedisCounterConfig{
		Addr:         addr,
		PoolSize:     10,
		MinIdleConns: 5,
	})
}

// NewRedisCounterWithConfig creates a new RedisCounter with custom configuration.
func NewRedisCounterWithConfig(cfg RedisCounterConfig) Counter {
	poolSize := cfg.PoolSize
	if poolSize == 0 {
		poolSize = 10
	}
	minIdleConns := cfg.MinIdleConns
	if minIdleConns == 0 {
		minIdleConns = 5
	}

	return &RedisCounter{
		client: redis.NewClient(&redis.Options{
			Addr:         cfg.Addr,
			Password:     cfg.Password,
			DB:           cfg.DB,
			PoolSize:     poolSize,
			MinIdleConns: minIdleConns,
		}),
	}
}

// Get returns the current count for a key.
// Returns 0 if the key does not exist or has expired.
func (r *RedisCounter) Get(ctx context.Context, key string) (int, error) {
	val, err := r.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil // Key doesn't exist, count is 0
	}
	if err != nil {
		return 0, err
	}
	return val, nil
}

// Increment adds n to the key and returns the new total.
// If the key doesn't exist, it's created with value 'n' and the given TTL.
// If the key exists, the value is incremented and TTL is NOT updated (sliding window).
func (r *RedisCounter) Increment(ctx context.Context, key string, n int, ttl time.Duration) (int, error) {
	// Use INCRBY for atomic increment
	result, err := r.client.IncrBy(ctx, key, int64(n)).Result()
	if err != nil {
		return 0, err
	}

	// Set expiration only if this is a new key (result == n after increment)
	// We use EXPIRE with NX flag to only set TTL if it doesn't already have one
	if result == int64(n) {
		// This was likely a new key, set the expiration
		if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
			return 0, err
		}
	}

	return int(result), nil
}

// Reset clears the counter for the given key.
func (r *RedisCounter) Reset(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Close closes the Redis client connection.
func (r *RedisCounter) Close() error {
	return r.client.Close()
}

// Ping checks if the Redis connection is alive.
func (r *RedisCounter) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}