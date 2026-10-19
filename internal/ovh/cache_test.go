package ovh

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCachedClient_CachesResponse(t *testing.T) {
	fake := NewFakeClient()
	fake.Responses["/test"] = "hello"

	cached := NewCachedClient(fake, 5*time.Minute)

	var result1, result2 string
	require.NoError(t, cached.Get("/test", &result1))
	require.NoError(t, cached.Get("/test", &result2))

	assert.Equal(t, "hello", result1)
	assert.Equal(t, "hello", result2)
	assert.Len(t, fake.Calls, 1, "second call should use cache")
}

func TestCachedClient_ExpiresAfterTTL(t *testing.T) {
	fake := NewFakeClient()
	fake.Responses["/test"] = "hello"

	now := time.Now()
	cached := NewCachedClient(fake, 5*time.Minute)
	cached.now = func() time.Time { return now }

	var result string
	require.NoError(t, cached.Get("/test", &result))

	// Advance time past TTL.
	cached.now = func() time.Time { return now.Add(6 * time.Minute) }

	require.NoError(t, cached.Get("/test", &result))
	assert.Len(t, fake.Calls, 2, "expired cache should trigger new call")
}

func TestCachedClient_PropagatesError(t *testing.T) {
	fake := NewFakeClient()
	fake.Errors["/test"] = fmt.Errorf("api down")

	cached := NewCachedClient(fake, 5*time.Minute)

	var result string
	err := cached.Get("/test", &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api down")
}

func TestCachedClient_DifferentPathsDifferentCache(t *testing.T) {
	fake := NewFakeClient()
	fake.Responses["/a"] = "alpha"
	fake.Responses["/b"] = "bravo"

	cached := NewCachedClient(fake, 5*time.Minute)

	var a, b string
	require.NoError(t, cached.Get("/a", &a))
	require.NoError(t, cached.Get("/b", &b))

	assert.Equal(t, "alpha", a)
	assert.Equal(t, "bravo", b)
	assert.Len(t, fake.Calls, 2)
}

func TestCachedClient_CacheStillValidBeforeTTL(t *testing.T) {
	fake := NewFakeClient()
	fake.Responses["/test"] = "hello"

	now := time.Now()
	cached := NewCachedClient(fake, 5*time.Minute)
	cached.now = func() time.Time { return now }

	var result string
	require.NoError(t, cached.Get("/test", &result))

	// Advance time but still within TTL.
	cached.now = func() time.Time { return now.Add(4 * time.Minute) }

	require.NoError(t, cached.Get("/test", &result))
	assert.Len(t, fake.Calls, 1, "should still use cache")
}
