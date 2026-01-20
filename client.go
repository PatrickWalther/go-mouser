package mouser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// DefaultBaseURL is the default Mouser API base URL.
	DefaultBaseURL = "https://api.mouser.com/api/v2"

	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second
)

// Client is a Mouser API client.
type Client struct {
	httpClient  *http.Client
	apiKey      string
	baseURL     string
	rateLimiter *RateLimiter
	retryConfig RetryConfig
	cache       Cache
	cacheConfig CacheConfig
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBaseURL sets a custom base URL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithRateLimiter sets a custom rate limiter.
func WithRateLimiter(rateLimiter *RateLimiter) ClientOption {
	return func(c *Client) {
		c.rateLimiter = rateLimiter
	}
}

// WithRetryConfig sets the retry configuration.
func WithRetryConfig(config RetryConfig) ClientOption {
	return func(c *Client) {
		c.retryConfig = config
	}
}

// WithoutRetry disables retries.
func WithoutRetry() ClientOption {
	return func(c *Client) {
		c.retryConfig = NoRetry()
	}
}

// WithCache sets a custom cache implementation.
func WithCache(cache Cache) ClientOption {
	return func(c *Client) {
		c.cache = cache
	}
}

// WithCacheConfig sets the cache configuration.
func WithCacheConfig(config CacheConfig) ClientOption {
	return func(c *Client) {
		c.cacheConfig = config
	}
}

// WithoutCache disables caching.
func WithoutCache() ClientOption {
	return func(c *Client) {
		c.cacheConfig.Enabled = false
	}
}

// NewClient creates a new Mouser API client.
func NewClient(apiKey string, opts ...ClientOption) (*Client, error) {
	if apiKey == "" {
		return nil, ErrNoAPIKey
	}

	cacheConfig := DefaultCacheConfig()

	c := &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		apiKey:      apiKey,
		baseURL:     DefaultBaseURL,
		rateLimiter: NewRateLimiter(DefaultRequestsPerMinute, DefaultRequestsPerDay),
		retryConfig: DefaultRetryConfig(),
		cacheConfig: cacheConfig,
	}

	for _, opt := range opts {
		opt(c)
	}

	// Initialize default cache if caching is enabled and no custom cache was provided
	if c.cacheConfig.Enabled && c.cache == nil {
		c.cache = NewMemoryCache(c.cacheConfig.DetailsTTL)
	}

	return c, nil
}

// RateLimiter returns the client's rate limiter.
func (c *Client) RateLimiter() *RateLimiter {
	return c.rateLimiter
}

// RateLimitStats returns current rate limit usage statistics.
func (c *Client) RateLimitStats() RateLimitStats {
	return c.rateLimiter.Stats()
}

// ClearCache clears all cached responses.
func (c *Client) ClearCache() {
	if mc, ok := c.cache.(*MemoryCache); ok {
		mc.Clear()
	}
}

// buildURL constructs a URL with the API key as a query parameter.
func (c *Client) buildURL(path string) (string, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return "", fmt.Errorf("mouser: invalid URL: %w", err)
	}

	q := u.Query()
	q.Set("apiKey", c.apiKey)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// doRequest performs an HTTP request with rate limiting, retries, and error handling.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	return c.doWithRetry(ctx, method, path, body, result)
}

// doWithRetry performs an HTTP request with retry logic.
func (c *Client) doWithRetry(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var lastErr error
	maxAttempts := c.retryConfig.MaxRetries + 1

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			backoff := c.retryConfig.calculateBackoff(attempt - 1)
			if err := sleep(ctx, backoff); err != nil {
				return err
			}
		}

		err, statusCode, retryAfter := c.doOnce(ctx, method, path, body, result)
		if err == nil {
			return nil
		}

		lastErr = err

		// Update rate limiter if we got a Retry-After header
		if retryAfter > 0 {
			c.rateLimiter.UpdateFromResponse(retryAfter)
		}

		// Check if we should retry
		if !shouldRetry(err, statusCode) {
			return err
		}

		// Don't retry on last attempt
		if attempt >= maxAttempts-1 {
			return err
		}
	}

	return lastErr
}

// doOnce performs a single HTTP request attempt.
// Returns (error, statusCode, retryAfterSeconds).
func (c *Client) doOnce(ctx context.Context, method, path string, body interface{}, result interface{}) (error, int, int) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return err, 0, 0
	}

	// Build URL with API key
	reqURL, err := c.buildURL(path)
	if err != nil {
		return err, 0, 0
	}

	// Marshal request body
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("mouser: failed to marshal request: %w", err), 0, 0
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return fmt.Errorf("mouser: failed to create request: %w", err), 0, 0
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Perform request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("mouser: request failed: %w", err), 0, 0
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("mouser: failed to read response: %w", err), resp.StatusCode, 0
	}

	// Parse Retry-After header
	retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))

	// Handle rate limiting (429)
	if resp.StatusCode == http.StatusTooManyRequests {
		return &MouserError{
			StatusCode:  resp.StatusCode,
			Message:     "rate limit exceeded",
			Details:     string(respBody),
			Endpoint:    path,
			RetryAfter:  retryAfter,
			IsRetryable: true,
		}, resp.StatusCode, retryAfter
	}

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &MouserError{
			StatusCode:  resp.StatusCode,
			Message:     http.StatusText(resp.StatusCode),
			Details:     string(respBody),
			Endpoint:    path,
			IsRetryable: shouldRetry(nil, resp.StatusCode),
		}, resp.StatusCode, retryAfter
	}

	// Unmarshal response
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("mouser: failed to parse response: %w", err), resp.StatusCode, 0
		}
	}

	return nil, resp.StatusCode, 0
}

// getCached retrieves a cached response if available.
func (c *Client) getCached(key string) ([]byte, bool) {
	if c.cache == nil || !c.cacheConfig.Enabled {
		return nil, false
	}
	return c.cache.Get(key)
}

// setCache stores a response in the cache.
func (c *Client) setCache(key string, data []byte, ttl time.Duration) {
	if c.cache == nil || !c.cacheConfig.Enabled {
		return
	}
	c.cache.Set(key, data, ttl)
}
