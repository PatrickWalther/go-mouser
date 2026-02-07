package mouser

import (
	"net/http"
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

	common       service
	Search       *SearchService
	Cart         *CartService
	OrderHistory *OrderHistoryService
	Order        *OrderService
}

type service struct {
	client *Client
}

// SearchService handles search-related API endpoints.
type SearchService service

// CartService handles cart-related API endpoints.
type CartService service

// OrderHistoryService handles order history API endpoints.
type OrderHistoryService service

// OrderService handles order-related API endpoints.
type OrderService service

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

	// Initialize services
	c.common.client = c
	c.Search = (*SearchService)(&c.common)
	c.Cart = (*CartService)(&c.common)
	c.OrderHistory = (*OrderHistoryService)(&c.common)
	c.Order = (*OrderService)(&c.common)

	return c, nil
}

// Close releases resources held by the client.
func (c *Client) Close() error {
	if mc, ok := c.cache.(*MemoryCache); ok {
		return mc.Close()
	}
	return nil
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
