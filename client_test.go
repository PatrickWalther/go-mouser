package mouser

import (
	"bufio"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var clientTestAPIKey string

func clientTestInit() {
	if clientTestAPIKey != "" {
		return
	}
	// Try environment first
	if key := os.Getenv("MOUSER_API_KEY"); key != "" {
		clientTestAPIKey = key
		return
	}
	// Try .env file
	file, err := os.Open(".env")
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "MOUSER_API_KEY=") {
			clientTestAPIKey = strings.TrimSpace(strings.TrimPrefix(line, "MOUSER_API_KEY="))
			return
		}
	}
}

// TestNewClient tests client creation with API key.
func TestNewClient(t *testing.T) {
	client, err := NewClient("test-api-key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer client.Close()

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.apiKey != "test-api-key" {
		t.Errorf("expected API key 'test-api-key', got %s", client.apiKey)
	}

	if client.baseURL != DefaultBaseURL {
		t.Errorf("expected base URL %s, got %s", DefaultBaseURL, client.baseURL)
	}

	if client.httpClient == nil {
		t.Fatal("expected non-nil HTTP client")
	}

	if client.rateLimiter == nil {
		t.Fatal("expected non-nil rate limiter")
	}
}

// TestNewClientNoAPIKey tests that client creation fails without API key.
func TestNewClientNoAPIKey(t *testing.T) {
	client, err := NewClient("")
	if err == nil {
		t.Fatal("expected error for empty API key")
	}

	if client != nil {
		t.Error("expected nil client when API key is empty")
	}

	if err != ErrNoAPIKey {
		t.Errorf("expected ErrNoAPIKey, got %v", err)
	}
}

// TestNewClientWithCustomHTTPClient tests client creation with custom HTTP client.
func TestNewClientWithCustomHTTPClient(t *testing.T) {
	customHTTPClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	client, err := NewClient("test-key", WithHTTPClient(customHTTPClient))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer client.Close()

	if client.httpClient != customHTTPClient {
		t.Error("expected same HTTP client instance")
	}
}

// TestNewClientWithRetryConfig tests client creation with custom retry config.
func TestNewClientWithRetryConfig(t *testing.T) {
	config := RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 100 * time.Millisecond,
	}

	client, err := NewClient("test-key", WithRetryConfig(config))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer client.Close()

	if client.retryConfig.MaxRetries != 5 {
		t.Errorf("expected max retries 5, got %d", client.retryConfig.MaxRetries)
	}
}

// TestNewClientWithCache tests client creation with custom cache.
func TestNewClientWithCache(t *testing.T) {
	cache := NewMemoryCache(5 * time.Minute)
	client, err := NewClient("test-key", WithCache(cache))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer client.Close()

	if client.cache == nil {
		t.Fatal("expected non-nil cache")
	}

	if client.cache != cache {
		t.Error("expected same cache instance")
	}
}

// TestNewClientWithCacheDisabled tests client creation with cache disabled.
func TestNewClientWithCacheDisabled(t *testing.T) {
	client, err := NewClient("test-key", WithoutCache())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer client.Close()

	if client.cacheConfig.Enabled {
		t.Error("expected cache to be disabled")
	}
}

// TestNewClientWithRetryDisabled tests client creation with retry disabled.
func TestNewClientWithRetryDisabled(t *testing.T) {
	client, err := NewClient("test-key", WithoutRetry())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer client.Close()

	if client.retryConfig.MaxRetries != 0 {
		t.Errorf("expected max retries 0, got %d", client.retryConfig.MaxRetries)
	}
}

// TestNewClientWithBaseURL tests client creation with custom base URL.
func TestNewClientWithBaseURL(t *testing.T) {
	customURL := "https://custom.example.com/api"
	client, err := NewClient("test-key", WithBaseURL(customURL))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer client.Close()

	if client.baseURL != customURL {
		t.Errorf("expected base URL %s, got %s", customURL, client.baseURL)
	}
}

// TestNewClientWithRateLimiter tests client creation with custom rate limiter.
func TestNewClientWithRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(50, 500)
	client, err := NewClient("test-key", WithRateLimiter(limiter))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer client.Close()

	if client.rateLimiter != limiter {
		t.Error("expected same rate limiter instance")
	}
}

// TestNewClientWithCacheConfig tests client creation with custom cache config.
func TestNewClientWithCacheConfig(t *testing.T) {
	config := CacheConfig{
		Enabled:          true,
		SearchTTL:        10 * time.Minute,
		DetailsTTL:       5 * time.Minute,
		ManufacturersTTL: 12 * time.Hour,
	}
	client, err := NewClient("test-key", WithCacheConfig(config))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer client.Close()

	if client.cacheConfig.SearchTTL != 10*time.Minute {
		t.Errorf("expected search TTL 10m, got %v", client.cacheConfig.SearchTTL)
	}
}

// TestRateLimitStats tests rate limit stats retrieval.
func TestRateLimitStats(t *testing.T) {
	client, _ := NewClient("test-key")
	defer client.Close()

	stats := client.RateLimitStats()
	if stats.MinuteRemaining <= 0 {
		t.Errorf("expected positive minute remaining, got %d", stats.MinuteRemaining)
	}
	if stats.DailyRemaining <= 0 {
		t.Errorf("expected positive daily remaining, got %d", stats.DailyRemaining)
	}
}

// TestClearCache tests cache clearing.
func TestClearCache(t *testing.T) {
	client, _ := NewClient("test-key")
	defer client.Close()

	// Should not panic
	client.ClearCache()

	// Verify cache is actually cleared
	cache := client.cache.(*MemoryCache)
	if cache.Size() != 0 {
		t.Errorf("expected cache size 0 after clear, got %d", cache.Size())
	}
}

// TestGetCached tests cache retrieval.
func TestGetCached(t *testing.T) {
	client, _ := NewClient("test-key")
	defer client.Close()
	cache := client.cache.(*MemoryCache)

	key := "test:key"
	expectedData := []byte("cached data")

	cache.Set(key, expectedData, 1*time.Minute)

	data, ok := client.getCached(key)
	if !ok {
		t.Fatal("expected to retrieve cached data")
	}

	if string(data) != string(expectedData) {
		t.Errorf("expected %s, got %s", expectedData, data)
	}
}

// TestGetCachedDisabled tests that getCached returns nothing when cache is disabled.
func TestGetCachedDisabled(t *testing.T) {
	client, _ := NewClient("test-key", WithoutCache())
	defer client.Close()

	data, ok := client.getCached("test:key")
	if ok {
		t.Error("expected cache miss when cache is disabled")
	}

	if data != nil {
		t.Error("expected nil data when cache is disabled")
	}
}

// TestSetCache tests cache storage.
func TestSetCache(t *testing.T) {
	client, _ := NewClient("test-key")
	defer client.Close()
	cache := client.cache.(*MemoryCache)

	key := "test:key"
	data := []byte("cached data")

	client.setCache(key, data, 1*time.Minute)

	retrieved, ok := cache.Get(key)
	if !ok {
		t.Fatal("expected cached data")
	}

	if string(retrieved) != string(data) {
		t.Errorf("expected %s, got %s", data, retrieved)
	}
}

// TestSetCacheDisabled tests that setCache does nothing when cache is disabled.
func TestSetCacheDisabled(t *testing.T) {
	client, _ := NewClient("test-key", WithoutCache())
	defer client.Close()

	client.setCache("test:key", []byte("data"), 1*time.Minute)

	// Verify nothing was cached (should remain empty)
	if client.cache != nil {
		cache := client.cache.(*MemoryCache)
		if cache.Size() > 0 {
			t.Error("expected cache to remain empty when disabled")
		}
	}
}

// TestDefaultTimeouts tests default HTTP client timeout.
func TestDefaultTimeouts(t *testing.T) {
	client, _ := NewClient("test-key")
	defer client.Close()

	if client.httpClient.Timeout != DefaultTimeout {
		t.Errorf("expected timeout %v, got %v", DefaultTimeout, client.httpClient.Timeout)
	}
}

// TestBuildURL tests URL construction with API key.
func TestBuildURL(t *testing.T) {
	client, _ := NewClient("my-api-key")
	defer client.Close()

	url, err := client.buildURL("/search/keyword")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !contains(url, "apiKey=my-api-key") {
		t.Errorf("expected URL to contain API key parameter, got %s", url)
	}

	if !contains(url, "/search/keyword") {
		t.Errorf("expected URL to contain path, got %s", url)
	}
}

// TestBuildURLInvalidURL tests URL construction error handling.
func TestBuildURLInvalidURL(t *testing.T) {
	client, _ := NewClient("test-key", WithBaseURL("ht!tp://invalid"))
	defer client.Close()

	_, err := client.buildURL("/test")
	if err == nil {
		t.Fatal("expected error for invalid base URL")
	}
}

// TestNewClientAllOptions tests creating client with multiple options.
func TestNewClientAllOptions(t *testing.T) {
	customHTTPClient := &http.Client{Timeout: 30 * time.Second}
	customCache := NewMemoryCache(10 * time.Minute)
	customLimiter := NewRateLimiter(60, 2000)

	client, err := NewClient("test-key",
		WithHTTPClient(customHTTPClient),
		WithCache(customCache),
		WithRateLimiter(customLimiter),
		WithBaseURL("https://custom.com"),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer client.Close()

	if client.httpClient != customHTTPClient {
		t.Error("expected custom HTTP client")
	}

	if client.cache != customCache {
		t.Error("expected custom cache")
	}

	if client.rateLimiter != customLimiter {
		t.Error("expected custom rate limiter")
	}

	if client.baseURL != "https://custom.com" {
		t.Error("expected custom base URL")
	}
}

// TestRateLimiterGetter tests RateLimiter getter method.
func TestRateLimiterGetter(t *testing.T) {
	limiter := NewRateLimiter(50, 500)
	client, _ := NewClient("test-key", WithRateLimiter(limiter))
	defer client.Close()

	retrieved := client.RateLimiter()
	if retrieved != limiter {
		t.Error("expected same rate limiter instance")
	}
}

// skipIfNoCredentials skips the test if MOUSER_API_KEY is not set.
func skipIfNoAPIKey(t *testing.T) {
	clientTestInit()
	if clientTestAPIKey == "" {
		t.Skip("MOUSER_API_KEY not found in environment or .env file")
	}
}

// TestIntegrationClientSetupWithRealAPI tests client setup works with real API credentials.
func TestIntegrationClientSetupWithRealAPI(t *testing.T) {
	skipIfNoAPIKey(t)

	apiKey := clientTestAPIKey
	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("failed to create client with real API key: %v", err)
	}
	defer client.Close()

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.apiKey != apiKey {
		t.Error("API key not properly set")
	}

	// Basic stats check
	stats := client.RateLimitStats()
	if stats.MinuteRemaining <= 0 || stats.DailyRemaining <= 0 {
		t.Logf("Warning: rate limits may be exhausted (minute: %d, daily: %d)",
			stats.MinuteRemaining, stats.DailyRemaining)
	}
}
