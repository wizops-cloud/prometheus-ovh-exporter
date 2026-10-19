package ovh

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type cacheEntry struct {
	data      []byte
	expiresAt time.Time
}

// CachedClient wraps a Client and caches GET responses for a configurable TTL.
type CachedClient struct {
	inner Client
	ttl   time.Duration
	mu    sync.Mutex
	cache map[string]cacheEntry
	now   func() time.Time
}

// NewCachedClient creates a caching wrapper around a Client.
func NewCachedClient(inner Client, ttl time.Duration) *CachedClient {
	return &CachedClient{
		inner: inner,
		ttl:   ttl,
		cache: make(map[string]cacheEntry),
		now:   time.Now,
	}
}

// Get returns a cached response if available and not expired, otherwise delegates to the inner client.
func (c *CachedClient) Get(path string, result interface{}) error {
	return c.GetWithContext(context.Background(), path, result)
}

// GetWithContext returns a cached response if available and not expired, otherwise delegates to the inner client.
func (c *CachedClient) GetWithContext(ctx context.Context, path string, result interface{}) error {
	c.mu.Lock()
	entry, found := c.cache[path]
	c.mu.Unlock()

	if found && c.now().Before(entry.expiresAt) {
		return json.Unmarshal(entry.data, result)
	}

	if err := c.inner.GetWithContext(ctx, path, result); err != nil {
		return err
	}

	data, _ := json.Marshal(result)
	c.mu.Lock()
	c.cache[path] = cacheEntry{data: data, expiresAt: c.now().Add(c.ttl)}
	c.mu.Unlock()

	return nil
}
