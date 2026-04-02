## Using Redis Instead of MemoryStore

`MemoryStore` is excellent for single-server development. However, for **production environments** with multiple application nodes, a shared distributed cache is required to ensure consistent cache hits across the cluster.

### 1. The `store.Store` Interface

To maintain plug-and-play compatibility with your middleware, the `RedisStore` must satisfy the core interface:

```go
type Store interface {
    Get(ctx context.Context, key string) (*Entry, error)
    Save(ctx context.Context, key string, entry *Entry, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Lock(ctx context.Context, key string, ttl time.Duration) (bool, error)
    Unlock(ctx context.Context, key string) error
}

```

---

### 2. Modern RedisStore Implementation

We recommend to use `github.com/redis/go-redis/v9` for its native support for context and modern Redis features.

```go
package store

import (
    "context"
    "encoding/json"
    "errors"
    "time"

    "github.com/redis/go-redis/v9"
)

var ErrNotFound = errors.New("cache: entry not found")

type RedisStore struct {
    client *redis.Client
}

func NewRedisStore(addr string) *RedisStore {
    return &RedisStore{
        client: redis.NewClient(&redis.Options{
            Addr: addr,
            // Configure pool for production:
            PoolSize:     10,
            MinIdleConns: 5,
        }),
    }
}

func (r *RedisStore) Get(ctx context.Context, key string) (*Entry, error) {
    val, err := r.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil, ErrNotFound
    }
    
    var e Entry
    if err := json.Unmarshal([]byte(val), &e); err != nil {
        return nil, err
    }
    return &e, nil
}

func (r *RedisStore) Save(ctx context.Context, key string, entry *Entry, ttl time.Duration) error {
    data, _ := json.Marshal(entry)
    return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisStore) Delete(ctx context.Context, key string) error {
    return r.client.Del(ctx, key).Err()
}

// Lock implements a distributed mutex using the SET NX PX pattern.
func (r *RedisStore) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
    // Atomic: Sets if Not Exists with an expiration (to prevent deadlocks)
    return r.client.SetNX(ctx, "lock:"+key, "1", ttl).Result()
}

func (r *RedisStore) Unlock(ctx context.Context, key string) error {
    return r.client.Del(ctx, "lock:"+key).Err()
}

```

---

### 3. Integration in `main.go`

Switching from `MemoryStore` to `RedisStore` is a zero-friction change.

```go
// Replace: memStore := store.NewMemoryStore()
rStore := store.NewRedisStore("localhost:6379")

e.Use(cache.New(cache.Config{
    Store:       rStore,
    TTL:         5 * time.Minute,
    MaxBodySize: 1 * 1024 * 1024, // 1MB limit for Redis efficiency
}))

```

---

### 4. Recommendations

* **Fail-Soft Logic:** In a real production environment, you might want to wrap `RedisStore` so that if Redis is down, the middleware simply skips the cache instead of returning an error to the user.
