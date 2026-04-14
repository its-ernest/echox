package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore implements the Store interface using Redis as the backend.
// It is suitable for distributed production environments where multiple
// application instances need to share a consistent cache.
type RedisStore struct {
	client *redis.Client
}

// RedisStoreOptions provides configuration options for RedisStore.
type RedisStoreOptions struct {
	// Addr is the Redis server address (e.g., "localhost:6379")
	Addr string

	// PoolSize is the maximum number of socket connections.
	// Default is 10 connections per CPU.
	PoolSize int

	// MinIdleConns is the minimum number of idle connections.
	// Default is 0.
	MinIdleConns int

	// Password for authentication. Empty string means no password.
	Password string

	// DB is the Redis database to use. Default is 0.
	DB int
}

// NewRedisStore creates a new RedisStore with the provided address.
// This is a convenience constructor for simple setups.
func NewRedisStore(addr string) *RedisStore {
	return NewRedisStoreWithOptions(RedisStoreOptions{
		Addr:         addr,
		PoolSize:     10,
		MinIdleConns: 5,
	})
}

// NewRedisStoreWithOptions creates a new RedisStore with custom configuration.
func NewRedisStoreWithOptions(opts RedisStoreOptions) *RedisStore {
	return &RedisStore{
		client: redis.NewClient(&redis.Options{
			Addr:         opts.Addr,
			PoolSize:     opts.PoolSize,
			MinIdleConns: opts.MinIdleConns,
			Password:     opts.Password,
			DB:           opts.DB,
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

// Save persists an Entry into Redis with a specific time-to-live (TTL).
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

// Lock attempts to acquire a distributed lock for a given key using SET NX.
// This is atomic and prevents race conditions in distributed environments.
// Returns true if the lock was successfully acquired.
func (r *RedisStore) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	// Use SET NX (Set if Not eXists) with expiration for atomic locking
	return r.client.SetNX(ctx, "lock:"+key, "1", ttl).Result()
}

// Unlock releases the lock for the given key.
func (r *RedisStore) Unlock(ctx context.Context, key string) error {
	return r.client.Del(ctx, "lock:"+key).Err()
}

// Close closes the Redis connection pool.
// It is important to call this when shutting down the application.
func (r *RedisStore) Close() error {
	return r.client.Close()
}
