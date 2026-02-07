package mouser

import (
	"testing"
	"time"
)

// TestMemoryCacheSet tests basic cache set operation.
func TestMemoryCacheSet(t *testing.T) {
	cache := NewMemoryCache(5 * time.Minute)
	defer cache.Close()

	key := "test:key"
	value := []byte("test value")

	cache.Set(key, value, 1*time.Minute)

	if cache.Size() != 1 {
		t.Errorf("expected cache size 1, got %d", cache.Size())
	}
}

// TestMemoryCacheGet tests basic cache get operation.
func TestMemoryCacheGet(t *testing.T) {
	cache := NewMemoryCache(5 * time.Minute)
	defer cache.Close()

	key := "test:key"
	value := []byte("test value")

	cache.Set(key, value, 1*time.Minute)

	retrieved, ok := cache.Get(key)
	if !ok {
		t.Fatal("expected to find value in cache")
	}

	if string(retrieved) != string(value) {
		t.Errorf("expected value %s, got %s", value, retrieved)
	}
}

// TestMemoryCacheGetMissing tests cache get for missing key.
func TestMemoryCacheGetMissing(t *testing.T) {
	cache := NewMemoryCache(5 * time.Minute)
	defer cache.Close()

	_, ok := cache.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss for nonexistent key")
	}
}

// TestMemoryCacheDelete tests cache delete operation.
func TestMemoryCacheDelete(t *testing.T) {
	cache := NewMemoryCache(5 * time.Minute)
	defer cache.Close()

	key := "test:key"
	cache.Set(key, []byte("value"), 1*time.Minute)

	if cache.Size() != 1 {
		t.Errorf("expected cache size 1 after set")
	}

	cache.Delete(key)

	if cache.Size() != 0 {
		t.Errorf("expected cache size 0 after delete")
	}

	_, ok := cache.Get(key)
	if ok {
		t.Fatal("expected cache miss after delete")
	}
}

// TestMemoryCacheTTL tests that expired entries are not returned.
func TestMemoryCacheTTL(t *testing.T) {
	cache := NewMemoryCache(5 * time.Minute)
	defer cache.Close()

	key := "test:key"
	cache.Set(key, []byte("value"), 100*time.Millisecond)

	// Should be available immediately
	_, ok := cache.Get(key)
	if !ok {
		t.Fatal("expected value in cache immediately after set")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	_, ok = cache.Get(key)
	if ok {
		t.Fatal("expected cache miss after TTL expiration")
	}
}

// TestMemoryCacheDefaultTTL tests that default TTL is used when zero is passed.
func TestMemoryCacheDefaultTTL(t *testing.T) {
	cache := NewMemoryCache(100 * time.Millisecond)
	defer cache.Close()

	key := "test:key"
	// Pass 0 as TTL to use default
	cache.Set(key, []byte("value"), 0)

	// Should be available immediately
	_, ok := cache.Get(key)
	if !ok {
		t.Fatal("expected value in cache immediately after set")
	}

	// Wait for default TTL expiration
	time.Sleep(150 * time.Millisecond)

	_, ok = cache.Get(key)
	if ok {
		t.Fatal("expected cache miss after default TTL expiration")
	}
}

// TestMemoryCacheMultipleEntries tests cache with multiple entries.
func TestMemoryCacheMultipleEntries(t *testing.T) {
	cache := NewMemoryCache(5 * time.Minute)
	defer cache.Close()

	entries := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
	}

	for key, value := range entries {
		cache.Set(key, value, 1*time.Minute)
	}

	if cache.Size() != 3 {
		t.Errorf("expected cache size 3, got %d", cache.Size())
	}

	for key, expectedValue := range entries {
		value, ok := cache.Get(key)
		if !ok {
			t.Errorf("expected to find key %s in cache", key)
			continue
		}
		if string(value) != string(expectedValue) {
			t.Errorf("expected value %s for key %s, got %s", expectedValue, key, value)
		}
	}
}

// TestMemoryCacheClear tests clearing all cache entries.
func TestMemoryCacheClear(t *testing.T) {
	cache := NewMemoryCache(5 * time.Minute)
	defer cache.Close()

	// Add multiple entries
	for i := 0; i < 5; i++ {
		cache.Set("key"+string(rune('0'+i)), []byte("value"), 1*time.Minute)
	}

	if cache.Size() == 0 {
		t.Fatal("expected cache to have entries")
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("expected cache size 0 after clear, got %d", cache.Size())
	}
}

// TestMemoryCacheSize tests the Size method.
func TestMemoryCacheSize(t *testing.T) {
	cache := NewMemoryCache(5 * time.Minute)
	defer cache.Close()

	if cache.Size() != 0 {
		t.Errorf("expected initial cache size 0, got %d", cache.Size())
	}

	for i := 0; i < 10; i++ {
		cache.Set("key"+string(rune('0'+i)), []byte("value"), 1*time.Minute)
		expected := i + 1
		if cache.Size() != expected {
			t.Errorf("expected cache size %d, got %d", expected, cache.Size())
		}
	}
}

// TestMemoryCacheOverwrite tests overwriting existing cache entries.
func TestMemoryCacheOverwrite(t *testing.T) {
	cache := NewMemoryCache(5 * time.Minute)
	defer cache.Close()

	key := "test:key"

	cache.Set(key, []byte("value1"), 1*time.Minute)
	if cache.Size() != 1 {
		t.Fatal("expected cache size 1 after first set")
	}

	// Overwrite with new value
	cache.Set(key, []byte("value2"), 1*time.Minute)
	if cache.Size() != 1 {
		t.Fatal("expected cache size 1 after overwrite (should not duplicate)")
	}

	value, ok := cache.Get(key)
	if !ok {
		t.Fatal("expected to find value in cache")
	}

	if string(value) != "value2" {
		t.Errorf("expected new value value2, got %s", value)
	}
}

// TestMemoryCacheEmptyValue tests storing empty values.
func TestMemoryCacheEmptyValue(t *testing.T) {
	cache := NewMemoryCache(5 * time.Minute)
	defer cache.Close()

	key := "test:key"
	cache.Set(key, []byte(""), 1*time.Minute)

	value, ok := cache.Get(key)
	if !ok {
		t.Fatal("expected to find empty value in cache")
	}

	if len(value) != 0 {
		t.Errorf("expected empty value, got %v", value)
	}
}

// TestCacheKeyForSearch tests cache key generation for search.
func TestCacheKeyForSearch(t *testing.T) {
	req := keywordSearchRequest{
		SearchByKeywordRequest: searchByKeywordRequest{
			Keyword: "test",
		},
	}

	key1 := cacheKeyForSearch("keyword", req)
	key2 := cacheKeyForSearch("keyword", req)

	if key1 != key2 {
		t.Error("expected same cache key for same request")
	}

	// Different method should produce different key
	key3 := cacheKeyForSearch("partnumber", req)
	if key1 == key3 {
		t.Error("expected different cache key for different method")
	}
}

// TestCacheKeyForDetails tests cache key generation for details.
func TestCacheKeyForDetails(t *testing.T) {
	key1 := cacheKeyForDetails("ABC123")
	key2 := cacheKeyForDetails("ABC123")

	if key1 != key2 {
		t.Error("expected same cache key for same part number")
	}

	key3 := cacheKeyForDetails("XYZ789")
	if key1 == key3 {
		t.Error("expected different cache key for different part number")
	}

	if !contains(key1, "ABC123") {
		t.Errorf("expected cache key to contain part number, got %s", key1)
	}
}

// TestCacheKeyForManufacturers tests cache key generation for manufacturers.
func TestCacheKeyForManufacturers(t *testing.T) {
	key1 := cacheKeyForManufacturers()
	key2 := cacheKeyForManufacturers()

	if key1 != key2 {
		t.Error("expected same cache key for manufacturers list")
	}

	if !contains(key1, "manufacturers") {
		t.Errorf("expected cache key to mention manufacturers, got %s", key1)
	}
}

// TestCacheInterface tests that MemoryCache implements Cache interface.
func TestCacheInterface(t *testing.T) {
	var _ Cache = (*MemoryCache)(nil)
}

// TestDefaultCacheConfig tests default cache configuration.
func TestDefaultCacheConfig(t *testing.T) {
	config := DefaultCacheConfig()

	if !config.Enabled {
		t.Error("expected cache to be enabled by default")
	}

	if config.SearchTTL == 0 {
		t.Error("expected non-zero search TTL")
	}

	if config.DetailsTTL == 0 {
		t.Error("expected non-zero details TTL")
	}

	if config.ManufacturersTTL == 0 {
		t.Error("expected non-zero manufacturers TTL")
	}

	// Manufacturers TTL should be longer than search TTL
	if config.ManufacturersTTL <= config.SearchTTL {
		t.Errorf("expected manufacturers TTL (%v) > search TTL (%v)", config.ManufacturersTTL, config.SearchTTL)
	}
}

// TestMemoryCacheCleanup tests that expired entries are cleaned up.
func TestMemoryCacheCleanup(t *testing.T) {
	cache := NewMemoryCache(5 * time.Minute)
	defer cache.Close()

	// Add an entry with very short TTL
	cache.Set("short-lived", []byte("value"), 50*time.Millisecond)

	// Wait for it to expire
	time.Sleep(100 * time.Millisecond)

	// Entry should no longer be retrievable
	_, ok := cache.Get("short-lived")
	if ok {
		t.Fatal("expected expired entry to be gone")
	}

	// After manual cleanup, size should be 0
	cache.cleanup()
	if cache.Size() != 0 {
		t.Errorf("expected cache size 0 after cleanup, got %d", cache.Size())
	}
}

// TestMemoryCacheConcurrentAccess tests concurrent reads and writes.
func TestMemoryCacheConcurrentAccess(t *testing.T) {
	cache := NewMemoryCache(5 * time.Minute)
	defer cache.Close()

	// This test ensures no panics during concurrent access
	done := make(chan bool, 3)

	// Goroutine 1: writes
	go func() {
		for i := 0; i < 10; i++ {
			cache.Set("key"+string(rune('0'+(i%10))), []byte("value"), 1*time.Minute)
		}
		done <- true
	}()

	// Goroutine 2: reads
	go func() {
		for i := 0; i < 20; i++ {
			_, _ = cache.Get("key" + string(rune('0'+(i%10))))
		}
		done <- true
	}()

	// Goroutine 3: deletes
	go func() {
		for i := 0; i < 5; i++ {
			cache.Delete("key" + string(rune('0'+i)))
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done
}

// TestMemoryCacheLargeValue tests storing large values.
func TestMemoryCacheLargeValue(t *testing.T) {
	cache := NewMemoryCache(5 * time.Minute)
	defer cache.Close()

	// Create a large value (1MB)
	largeValue := make([]byte, 1024*1024)
	for i := range largeValue {
		largeValue[i] = byte(i % 256)
	}

	cache.Set("large", largeValue, 1*time.Minute)

	retrieved, ok := cache.Get("large")
	if !ok {
		t.Fatal("expected to retrieve large value")
	}

	if len(retrieved) != len(largeValue) {
		t.Errorf("expected value size %d, got %d", len(largeValue), len(retrieved))
	}
}

// Helper function to check if string contains substring
func contains(s, substring string) bool {
	for i := 0; i <= len(s)-len(substring); i++ {
		if s[i:i+len(substring)] == substring {
			return true
		}
	}
	return false
}
