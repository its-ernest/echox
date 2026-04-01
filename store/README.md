# Store Package

The `store` package provides persistence and locking mechanisms for the echox middleware suite. It defines two core interfaces: `Store` and `Counter`, with multiple backend implementations.

## Overview

### Store Interface

The `Store` interface provides caching and distributed locking capabilities:

```go
type Store interface {
    Get(ctx context.Context, key string) (*Entry, error)
    Save(ctx context.Context, key string, entry *Entry, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Lock(ctx context.Context, key string, ttl time.Duration) (bool, error)
    Unlock(ctx context.Context, key string) error
}
```

### Counter Interface

The `Counter` interface provides simple numeric tracking for rate limiting and abuse detection:

```go
type Counter interface {
    Get(ctx context.Context, key string) (int, error)
    Increment(ctx context.Context, key string, n int, ttl time.Duration) (int, error)
    Reset(ctx context.Context, key string) error
}
```

## Implementations

### MemoryStore (Development)

`MemoryStore` is an in-memory implementation ideal for single-instance applications:

```go
memStore := store.NewMemoryStore()

// Optional: Start background eviction
stop := make(chan struct{})
memStore.StartEvictor(1*time.Minute, stop)
// Later: close(stop) to stop the evictor
```

**Features:**
- Thread-safe using sync.Map
- Simulated distributed locking
- Automatic TTL expiration
- Background eviction support

### MemoryCounter (Development)

`MemoryCounter` is an in-memory counter for single-instance applications:

```go
counter := store.NewMemoryCounter()
```

### RedisStore (Production)

`RedisStore` provides distributed caching for multi-instance deployments:

```go
// Simple initialization
redisStore := store.NewRedisStore("localhost:6379")

// Production configuration
redisStore := store.NewRedisStoreWithConfig(store.RedisStoreConfig{
    Addr:         "redis-cluster.example.com:6379",
    Password:     "your-password",
    DB:           0,
    PoolSize:     50,
    MinIdleConns: 10,
})

// Check connection
if err := redisStore.Ping(context.Background()); err != nil {
    log.Fatal("Redis connection failed:", err)
}
defer redisStore.Close()
```

**Features:**
- Distributed caching across multiple instances
- Atomic locking using Redis SET NX pattern
- JSON serialization for Entry storage
- Connection pooling
- Configurable pool size

### RedisCounter (Production)

`RedisCounter` provides distributed counting for multi-instance deployments:

```go
// Simple initialization
counter := store.NewRedisCounter("localhost:6379")

// Production configuration
counter := store.NewRedisCounterWithConfig(store.RedisCounterConfig{
    Addr:         "redis-cluster.example.com:6379",
    Password:     "your-password",
    DB:           0,
    PoolSize:     50,
    MinIdleConns: 10,
})
```

**Features:**
- Distributed counting across multiple instances
- Atomic INCRBY operations
- TTL support for automatic expiration
- Connection pooling

## Usage Examples

### Caching with Middleware

```go
import (
    "github.com/its-ernest/echox/cache"
    "github.com/its-ernest/echox/store"
)

// Development: Memory store
e.Use(cache.New(cache.Config{
    Store: store.NewMemoryStore(),
    TTL:   10 * time.Minute,
}))

// Production: Redis store
e.Use(cache.New(cache.Config{
    Store: store.NewRedisStore("localhost:6379"),
    TTL:   5 * time.Minute,
}))
```

### Abuse Detection with Middleware

```go
import (
    "github.com/its-ernest/echox/abuse"
    "github.com/its-ernest/echox/store"
)

// Development: Memory counter
e.Use(abuse.New(abuse.Config{
    Store:     store.NewMemoryCounter(),
    Threshold: 100,
    Rules: []abuse.Rule{
        {Path: "/.env", Score: 100},
        {Path: "/api/v1/login", Method: "POST", Score: 20},
    },
}))

// Production: Redis counter
e.Use(abuse.New(abuse.Config{
    Store:     store.NewRedisCounter("localhost:6379"),
    Threshold: 100,
    Rules: []abuse.Rule{
        {Path: "/.env", Score: 100},
        {Path: "/api/v1/login", Method: "POST", Score: 20},
    },
}))
```

### Manual Store Operations

```go
ctx := context.Background()

// Save an entry
entry := &store.Entry{
    Status: 200,
    Header: map[string][]string{"Content-Type": {"application/json"}},
    Body:   []byte(`{"data":"value"}`),
}
err := store.Save(ctx, "cache:key", entry, 10*time.Minute)

// Get an entry
entry, err := store.Get(ctx, "cache:key")

// Delete an entry
err := store.Delete(ctx, "cache:key")

// Distributed locking
acquired, err := store.Lock(ctx, "resource:id", 30*time.Second)
if acquired {
    // Do work
    defer store.Unlock(ctx, "resource:id")
}
```

### Manual Counter Operations

```go
ctx := context.Background()

// Increment counter
val, err := counter.Increment(ctx, "rate:user:123", 1, 1*time.Hour)

// Get current count
val, err := counter.Get(ctx, "rate:user:123")

// Reset counter
err := counter.Reset(ctx, "rate:user:123")
```

## Production Considerations

### Connection Pooling

For high-throughput applications, configure appropriate pool sizes:

```go
redisStore := store.NewRedisStoreWithConfig(store.RedisStoreConfig{
    Addr:         "redis.example.com:6379",
    PoolSize:     100,    // Max connections
    MinIdleConns: 20,     // Minimum idle connections
})
```

### Fail-Soft Logic

For resilience, consider wrapping Redis operations:

```go
type FailSoftStore struct {
    redis *store.RedisStore
    mem   *store.MemoryStore
}

func (f *FailSoftStore) Get(ctx context.Context, key string) (*Entry, error) {
    entry, err := f.redis.Get(ctx, key)
    if err != nil {
        // Fallback to memory store
        return f.mem.Get(ctx, key)
    }
    return entry, nil
}
```

### Key Naming

Use consistent key naming conventions:

```go
// Caching
key := "cache:" + requestID

// Locking
lockKey := "lock:" + resourceID

// Rate limiting
rateKey := "rate:" + userID + ":" + endpoint

// Abuse detection
abuseKey := "abuse:" + clientIP
```

## Testing

### Memory Store Tests

Run memory store tests without external dependencies:

```bash
go test ./store -run Memory
```

### Redis Store Tests

Redis tests require a running Redis server:

```bash
# Start Redis
docker run -d -p 6379:6379 redis:alpine

# Run tests
go test ./store -run Redis
```

## Dependencies

- **Memory implementations:** Standard library only
- **Redis implementations:** `github.com/redis/go-redis/v9`

## Migration from Memory to Redis

Switching from MemoryStore to RedisStore is seamless:

```go
// Before (development)
var cacheStore store.Store = store.NewMemoryStore()

// After (production)
var cacheStore store.Store = store.NewRedisStore("localhost:6379")

// Middleware configuration remains the same
e.Use(cache.New(cache.Config{
    Store: cacheStore,
    TTL:   5 * time.Minute,
}))
```