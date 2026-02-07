package mouser

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

// Cache defines the interface for caching API responses.
type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte, ttl time.Duration)
	Delete(key string)
}

// MemoryCache is a simple in-memory cache with TTL support.
type MemoryCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	ttl     time.Duration
	done    chan struct{}
}

type cacheEntry struct {
	value     []byte
	expiresAt time.Time
}

// NewMemoryCache creates a new in-memory cache with the specified default TTL.
func NewMemoryCache(defaultTTL time.Duration) *MemoryCache {
	c := &MemoryCache{
		entries: make(map[string]*cacheEntry),
		ttl:     defaultTTL,
		done:    make(chan struct{}),
	}
	go c.cleanupLoop()
	return c
}

// Get retrieves a value from the cache.
func (c *MemoryCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return entry.value, true
}

// Set stores a value in the cache with the specified TTL.
// If ttl is 0, the default TTL is used.
func (c *MemoryCache) Set(key string, value []byte, ttl time.Duration) {
	if ttl == 0 {
		ttl = c.ttl
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

// Delete removes a value from the cache.
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// cleanupLoop periodically removes expired entries.
func (c *MemoryCache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.done:
			return
		}
	}
}

// Close stops the cleanup goroutine and releases resources.
func (c *MemoryCache) Close() error {
	close(c.done)
	return nil
}

func (c *MemoryCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, key)
		}
	}
}

// Size returns the number of entries in the cache.
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Clear removes all entries from the cache.
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*cacheEntry)
}

// CacheConfig configures caching behavior.
type CacheConfig struct {
	Enabled          bool
	SearchTTL        time.Duration // TTL for search results
	DetailsTTL       time.Duration // TTL for product details
	ManufacturersTTL time.Duration // TTL for manufacturer list (longer, mostly static)
}

// DefaultCacheConfig returns the default cache configuration.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Enabled:          true,
		SearchTTL:        5 * time.Minute,
		DetailsTTL:       10 * time.Minute,
		ManufacturersTTL: 24 * time.Hour,
	}
}

// cacheKeyForSearch generates a cache key for a search request.
func cacheKeyForSearch(method string, req interface{}) string {
	data, _ := json.Marshal(req)
	hash := sha256.Sum256(data)
	return "search:" + method + ":" + hex.EncodeToString(hash[:8])
}

// cacheKeyForDetails generates a cache key for product details.
func cacheKeyForDetails(partNumber string) string {
	return "details:" + partNumber
}

// cacheKeyForManufacturers generates a cache key for the manufacturer list.
func cacheKeyForManufacturers() string {
	return "manufacturers:list"
}
