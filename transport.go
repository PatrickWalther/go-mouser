package mouser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

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
	return c.doWithRetry(ctx, method, path, nil, body, result)
}

// doRequestWithQuery performs an HTTP request with additional URL query parameters.
func (c *Client) doRequestWithQuery(ctx context.Context, method, path string, query url.Values, body interface{}, result interface{}) error {
	return c.doWithRetry(ctx, method, path, query, body, result)
}

// doWithRetry performs an HTTP request with retry logic.
func (c *Client) doWithRetry(ctx context.Context, method, path string, query url.Values, body interface{}, result interface{}) error {
	var lastErr error
	maxAttempts := c.retryConfig.MaxRetries + 1

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			backoff := c.retryConfig.calculateBackoff(attempt - 1)
			if err := sleep(ctx, backoff); err != nil {
				return err
			}
		}

		statusCode, retryAfter, err := c.doOnce(ctx, method, path, query, body, result)
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
// Returns (statusCode, retryAfterSeconds, error).
func (c *Client) doOnce(ctx context.Context, method, path string, query url.Values, body interface{}, result interface{}) (int, int, error) {
	// Check rate limiter (non-blocking)
	if err := c.rateLimiter.Allow(); err != nil {
		return 0, 0, err
	}

	// Build URL with API key
	reqURL, err := c.buildURL(path)
	if err != nil {
		return 0, 0, err
	}

	// Merge additional query parameters
	if len(query) > 0 {
		u, err := url.Parse(reqURL)
		if err != nil {
			return 0, 0, fmt.Errorf("mouser: invalid URL: %w", err)
		}
		q := u.Query()
		for k, vs := range query {
			for _, v := range vs {
				q.Set(k, v)
			}
		}
		u.RawQuery = q.Encode()
		reqURL = u.String()
	}

	// Marshal request body
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return 0, 0, fmt.Errorf("mouser: failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return 0, 0, fmt.Errorf("mouser: failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Perform request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("mouser: request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, 0, fmt.Errorf("mouser: failed to read response: %w", err)
	}

	// Sync rate limiter from response headers on every response.
	c.rateLimiter.UpdateFromHeaders(resp.Header)

	// Parse Retry-After header
	retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))

	// Handle rate limiting (429)
	if resp.StatusCode == http.StatusTooManyRequests {
		return resp.StatusCode, retryAfter, &MouserError{
			StatusCode:  resp.StatusCode,
			Message:     "rate limit exceeded",
			Details:     string(respBody),
			Endpoint:    path,
			RetryAfter:  retryAfter,
			IsRetryable: true,
		}
	}

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, retryAfter, &MouserError{
			StatusCode:  resp.StatusCode,
			Message:     http.StatusText(resp.StatusCode),
			Details:     string(respBody),
			Endpoint:    path,
			IsRetryable: shouldRetry(nil, resp.StatusCode),
		}
	}

	// Unmarshal response
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return resp.StatusCode, 0, fmt.Errorf("mouser: failed to parse response: %w", err)
		}
	}

	return resp.StatusCode, 0, nil
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

// parseRetryAfter parses the Retry-After header value.
// Returns the number of seconds to wait, or 0 if not parseable.
func parseRetryAfter(header string) int {
	if header == "" {
		return 0
	}

	// Try parsing as seconds
	if seconds, err := strconv.Atoi(header); err == nil {
		return seconds
	}

	// Try parsing as HTTP-date
	if t, err := time.Parse(time.RFC1123, header); err == nil {
		seconds := int(time.Until(t).Seconds())
		if seconds > 0 {
			return seconds
		}
	}

	return 0
}

// sleep waits for the specified duration, respecting context cancellation.
func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
