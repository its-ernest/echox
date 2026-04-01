package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore is a Redis-backed implementation of the Store interface.
// It provides distributed caching and locking capabilities for multi-instance
// deployments, ensuring consistent cache hits across the cluster.
type RedisStore struct {
	client *redis.Client
}

// RedisStoreConfig provides configuration options for the Redis client.
type RedisStoreConfig struct {
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

// NewRedisStore creates a new RedisStore with default configuration.
// For production use, consider NewRedisStoreWithConfig for more control.
func NewRedisStore(addr string) *RedisStore {
	return NewRedisStoreWithConfig(RedisStoreConfig{
		Addr:         addr,
		PoolSize:     10,
		MinIdleConns: 5,
	})
}

// NewRedisStoreWithConfig creates a new RedisStore with custom configuration.
func NewRedisStoreWithConfig(cfg RedisStoreConfig) *RedisStore {
	poolSize := cfg.PoolSize
	if poolSize == 0 {
		poolSize = 10
	}
	minIdleConns := cfg.MinIdleConns
	if minIdleConns == 0 {
		minIdleConns = 5
	}

	return &RedisStore{
		client: redis.NewClient(&redis.Options{
			Addr:         cfg.Addr,
			Password:     cfg.Password,
			DB:           cfg.DB,
			PoolSize:     poolSize,
			MinIdleConns: minIdleConns,
		}),
	}
}

// Get retrieves a cached Entry by its key from Redis.
// Returns ErrNotFound if the key does not exist or has expired.
func (r *RedisStore) Get(ctx context.Context, key string) (*Entry, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var e Entry
	if err := json.Unmarshal([]byte(val), &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// Save persists an Entry to Redis with a specific time-to-live (TTL).
func (r *RedisStore) Save(ctx context.Context, key string, entry *Entry, ttl time.Duration) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes a specific entry from Redis immediately.
func (r *RedisStore) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Lock attempts to acquire a distributed lock for a given key using Redis SET NX.
// This implements a non-blocking lock pattern: if the lock is successfully acquired,
// it returns true; otherwise, it returns false.
// The lock will automatically expire after the specified TTL.
func (r *RedisStore) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	lockKey := "lock:" + key
	return r.client.SetNX(ctx, lockKey, "1", ttl).Result()
}

// Unlock releases the lock for the given key, allowing other processes to acquire it.
func (r *RedisStore) Unlock(ctx context.Context, key string) error {
	lockKey := "lock:" + key
	return r.client.Del(ctx, lockKey).Err()
}

// Close closes the Redis client connection.
func (r *RedisStore) Close() error {
	return r.client.Close()
}

// Ping checks if the Redis connection is alive.
func (r *RedisStore) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}