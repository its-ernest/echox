## Using Redis Instead of MemoryStore

`MemoryStore` is great for a single-server, in-memory cache, but in **production with multiple servers**, you usually want **shared distributed caching**. Redis is perfect for this.

### 1. Implement the `store.Store` Interface

Your Redis store must implement the same methods:

```go
type Store interface {
    Get(ctx context.Context, key string) (*Entry, error)
    Save(ctx context.Context, key string, entry *Entry, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Lock(ctx context.Context, key string, ttl time.Duration) (bool, error)
    Unlock(ctx context.Context, key string) error
}
```

* `Get/Save/Delete` interact with Redis keys and TTLs.
* `Lock/Unlock` can use `SET key value NX PX ttl` for atomic locking.

---

### 2. Example RedisStore (simplified)

```go
package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr, password string, db int) *RedisStore {
	return &RedisStore{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

func (r *RedisStore) Get(ctx context.Context, key string) (*Entry, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	var e Entry
	if err := json.Unmarshal([]byte(val), &e); err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *RedisStore) Save(ctx context.Context, key string, entry *Entry, ttl time.Duration) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisStore) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Lock/Unlock using SET NX PX
func (r *RedisStore) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	ok, err := r.client.SetNX(ctx, key, "1", ttl).Result()
	return ok, err
}

func (r *RedisStore) Unlock(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
```

---

### 3. Replace MemoryStore in Middleware

```go
rstore := store.NewRedisStore("localhost:6379", "", 0)

e.Use(cache.New(cache.Config{
    Store:       rstore,
    TTL:         2 * time.Minute,
    MaxBodySize: 2 * 1024 * 1024,
}))
```

* Everything else in the middleware works **without changes**.
* Locks in Redis are distributed, so multiple servers **won’t stampede**.
* TTL for cache entries and locks is enforced at Redis level.

---

### 4. Notes for Production

* For high concurrency, consider **Redlock** algorithm for distributed locks.
* Set **MaxBodySize** according to your Redis memory constraints.
* Use **connection pooling** and **timeouts** to prevent Redis latency from slowing your requests.
* Middleware logging and ETag handling remain the same.
