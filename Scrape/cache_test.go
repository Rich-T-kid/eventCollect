package scrape

import (
	"context"
	"sync"
	"testing"
	"time"
)

// Helper function to set up the test cache
func setupTestCache() *redCache {
	cache := newCache()

	// Assert that the returned object is of type *redCache
	redCacheInstance, ok := cache.(*redCache)
	if !ok {
		panic("newCache did not return a *redCache instance")
	}

	return redCacheInstance
}

func TestRedCache(t *testing.T) {
	cache := setupTestCache()
	ctx := context.Background()

	// Cleanup Redis before running tests
	if err := cache.client.FlushAll(ctx).Err(); err != nil {
		t.Fatalf("Failed to flush Redis: %v", err)
	}

	// Test: Put and Get
	t.Run("Put and Get", func(t *testing.T) {
		err := cache.Put("testKey", "testValue")
		if err != nil {
			t.Fatalf("Put failed: %v", err)
		}

		val, found := cache.Get("testKey")
		if !found || val != "testValue" {
			t.Fatalf("Expected 'testValue', got '%v', found: %v", val, found)
		}
	})

	// Test: Exist
	t.Run("Exist", func(t *testing.T) {
		if !cache.Exist("testKey") {
			t.Fatalf("Exist failed for existing key")
		}
		if cache.Exist("missingKey") {
			t.Fatalf("Exist returned true for non-existent key")
		}
	})

	// Test: Delete
	t.Run("Delete", func(t *testing.T) {
		err := cache.Delete("testKey")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		if cache.Exist("testKey") {
			t.Fatalf("Delete did not remove the key")
		}
	})

	// Test: IncreaseTTL
	t.Run("IncreaseTTL", func(t *testing.T) {
		err := cache.Put("ttlKey", "ttlValue")
		if err != nil {
			t.Fatalf("Put failed: %v", err)
		}
		cache.SetTTl("ttlKey", 5*time.Second)

		err = cache.IncreaseTTL("ttlKey", 5*time.Second)
		if err != nil {
			t.Fatalf("IncreaseTTL failed: %v", err)
		}

		ttl, err := cache.client.TTL(ctx, "ttlKey").Result()
		if err != nil {
			t.Fatalf("TTL retrieval failed: %v", err)
		}
		if ttl < 9*time.Second || ttl > 11*time.Second {
			t.Fatalf("Expected TTL between 9-11 seconds, got %v", ttl)
		}
	})

	// Test: SetTTL
	t.Run("SetTTL", func(t *testing.T) {
		err := cache.Put("setTtlKey", "setTtlValue")
		if err != nil {
			t.Fatalf("Put failed: %v", err)
		}

		err = cache.SetTTl("setTtlKey", 10*time.Second)
		if err != nil {
			t.Fatalf("SetTTl failed: %v", err)
		}

		ttl, err := cache.client.TTL(ctx, "setTtlKey").Result()
		if err != nil {
			t.Fatalf("TTL retrieval failed: %v", err)
		}
		if ttl < 9*time.Second || ttl > 11*time.Second {
			t.Fatalf("Expected TTL between 9-11 seconds, got %v", ttl)
		}
	})

	// Test: Save (no-op in this case)
	t.Run("Save", func(t *testing.T) {
		err := cache.Save()
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	})

	// Test: Concurrency
	t.Run("Concurrency", func(t *testing.T) {
		var wg sync.WaitGroup
		err := cache.Put("concurrentKey", "value1")
		if err != nil {
			t.Fatalf("Put failed: %v", err)
		}

		wg.Add(2)
		go func() {
			defer wg.Done()
			_, _ = cache.Get("concurrentKey")
		}()
		go func() {
			defer wg.Done()
			_ = cache.Put("concurrentKey", "value2")
		}()
		wg.Wait()

		val, found := cache.Get("concurrentKey")
		if !found || (val != "value1" && val != "value2") {
			t.Fatalf("Concurrency test failed, got '%v'", val)
		}
	})

	// Edge Case: Empty Key
	t.Run("Empty Key", func(t *testing.T) {
		err := cache.Put("", "emptyKey")
		if err == nil {
			t.Fatalf("Expected error for empty key, got nil")
		}
	})

	// Edge Case: Large Value
	t.Run("Large Value", func(t *testing.T) {
		largeValue := make([]byte, 10*1024*1024) // 10 MB value
		for i := range largeValue {
			largeValue[i] = 'a'
		}

		err := cache.Put("largeKey", string(largeValue))
		if err != nil {
			t.Fatalf("Put failed for large value: %v", err)
		}

		val, found := cache.Get("largeKey")
		if !found || len(val) != len(largeValue) {
			t.Fatalf("Expected large value of size %d, got %d", len(largeValue), len(val))
		}
	})
}
