package mouser

import (
	"bufio"
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

var rateLimitTestAPIKey string

func rateLimitTestInit() {
	if rateLimitTestAPIKey != "" {
		return
	}
	// Try environment first
	if key := os.Getenv("MOUSER_API_KEY"); key != "" {
		rateLimitTestAPIKey = key
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
			rateLimitTestAPIKey = strings.TrimSpace(strings.TrimPrefix(line, "MOUSER_API_KEY="))
			return
		}
	}
}

// TestNewRateLimiter tests rate limiter creation.
func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(30, 1000)

	if rl == nil {
		t.Fatal("expected non-nil rate limiter")
	}

	stats := rl.Stats()
	if stats.MinuteRemaining != 30 {
		t.Errorf("expected 30 minute tokens, got %d", stats.MinuteRemaining)
	}
	if stats.DailyRemaining != 1000 {
		t.Errorf("expected 1000 daily tokens, got %d", stats.DailyRemaining)
	}
}

// TestRateLimiterDefaultLimits tests default rate limits.
func TestRateLimiterDefaultLimits(t *testing.T) {
	rl := NewRateLimiter(DefaultRequestsPerMinute, DefaultRequestsPerDay)

	stats := rl.Stats()
	if stats.MinuteRemaining != DefaultRequestsPerMinute {
		t.Errorf("expected minute limit %d, got %d", DefaultRequestsPerMinute, stats.MinuteRemaining)
	}
	if stats.DailyRemaining != DefaultRequestsPerDay {
		t.Errorf("expected daily limit %d, got %d", DefaultRequestsPerDay, stats.DailyRemaining)
	}
}

// TestRateLimiterWaitSuccess tests successful wait.
func TestRateLimiterWaitSuccess(t *testing.T) {
	rl := NewRateLimiter(5, 100)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		err := rl.Wait(ctx)
		if err != nil {
			t.Errorf("request %d should be allowed, got error: %v", i+1, err)
		}
	}
}

// TestRateLimiterWaitMinuteLimit tests minute limit enforcement.
func TestRateLimiterWaitMinuteLimit(t *testing.T) {
	rl := NewRateLimiter(2, 100)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Use up minute limit
	_ = rl.Wait(ctx)
	_ = rl.Wait(ctx)

	// Next request should block due to minute limit and then timeout from context
	err := rl.Wait(ctx)
	if err == nil {
		t.Fatal("expected error on 3rd request")
	}

	// Should be context timeout (blocked waiting for minute to reset)
	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded when blocked, got %v", err)
	}
}

// TestRateLimiterTryAcquire tests non-blocking acquire.
func TestRateLimiterTryAcquire(t *testing.T) {
	rl := NewRateLimiter(5, 100)

	for i := 0; i < 5; i++ {
		ok, err := rl.TryAcquire()
		if !ok {
			t.Errorf("request %d should be allowed", i+1)
		}
		if err != nil {
			t.Errorf("request %d should not have error, got %v", i+1, err)
		}
	}

	// Next request should fail
	ok, err := rl.TryAcquire()
	if ok {
		t.Fatal("expected rate limit on 6th request")
	}
	if err != ErrRateLimitExceeded {
		t.Errorf("expected ErrRateLimitExceeded, got %v", err)
	}
}

// TestRateLimiterStats tests stats retrieval.
func TestRateLimiterStats(t *testing.T) {
	rl := NewRateLimiter(10, 100)

	stats1 := rl.Stats()
	if stats1.MinuteRemaining != 10 {
		t.Errorf("expected 10 minute remaining, got %d", stats1.MinuteRemaining)
	}

	// Make some requests
	_, _ = rl.TryAcquire()
	_, _ = rl.TryAcquire()

	stats2 := rl.Stats()
	if stats2.MinuteRemaining != 8 {
		t.Errorf("expected 8 minute remaining after 2 requests, got %d", stats2.MinuteRemaining)
	}
	if stats2.DailyRemaining != 98 {
		t.Errorf("expected 98 daily remaining, got %d", stats2.DailyRemaining)
	}
}

// TestRateLimiterUpdateFromResponse tests updating limits from server response.
func TestRateLimiterUpdateFromResponse(t *testing.T) {
	rl := NewRateLimiter(10, 100)
	ctx := context.Background()

	// Use some requests
	for i := 0; i < 5; i++ {
		_ = rl.Wait(ctx)
	}

	// Server indicates rate limit with 60 second retry-after
	rl.UpdateFromResponse(60)

	// Should now be blocked
	ok, err := rl.TryAcquire()
	if ok {
		t.Error("expected to be blocked after server rate limit")
	}
	if err != ErrRateLimitExceeded {
		t.Errorf("expected ErrRateLimitExceeded after server rate limit, got %v", err)
	}

	stats := rl.Stats()
	if stats.BlockedUntil.IsZero() {
		t.Error("expected BlockedUntil to be set")
	}
}

// TestRateLimiterUpdateFromResponseZero tests UpdateFromResponse with zero.
func TestRateLimiterUpdateFromResponseZero(t *testing.T) {
	rl := NewRateLimiter(10, 100)

	before := rl.Stats()

	rl.UpdateFromResponse(0)

	after := rl.Stats()

	if before.MinuteRemaining != after.MinuteRemaining {
		t.Error("stats should not change with zero retry-after")
	}
}

// TestRateLimiterUpdateFromResponseNegative tests UpdateFromResponse with negative value.
func TestRateLimiterUpdateFromResponseNegative(t *testing.T) {
	rl := NewRateLimiter(10, 100)

	before := rl.Stats()

	rl.UpdateFromResponse(-10)

	after := rl.Stats()

	if before.MinuteRemaining != after.MinuteRemaining {
		t.Error("stats should not change with negative retry-after")
	}
}

// TestRateLimiterRemainingMinute tests RemainingMinute method.
func TestRateLimiterRemainingMinute(t *testing.T) {
	rl := NewRateLimiter(10, 100)

	remaining := rl.RemainingMinute()
	if remaining != 10 {
		t.Errorf("expected 10 remaining, got %d", remaining)
	}

	_, _ = rl.TryAcquire()
	remaining = rl.RemainingMinute()
	if remaining != 9 {
		t.Errorf("expected 9 remaining after 1 request, got %d", remaining)
	}
}

// TestRateLimiterRemainingDaily tests RemainingDaily method.
func TestRateLimiterRemainingDaily(t *testing.T) {
	rl := NewRateLimiter(10, 100)

	remaining := rl.RemainingDaily()
	if remaining != 100 {
		t.Errorf("expected 100 remaining, got %d", remaining)
	}

	_, _ = rl.TryAcquire()
	remaining = rl.RemainingDaily()
	if remaining != 99 {
		t.Errorf("expected 99 remaining after 1 request, got %d", remaining)
	}
}

// TestRateLimiterDailyLimitExceeded tests daily limit enforcement.
func TestRateLimiterDailyLimitExceeded(t *testing.T) {
	rl := NewRateLimiter(100, 2) // High per-minute, low daily limit

	// Use up daily limit
	_, _ = rl.TryAcquire()
	_, _ = rl.TryAcquire()

	// Next request should fail daily limit
	ok, err := rl.TryAcquire()
	if ok {
		t.Fatal("expected to be denied when daily limit exceeded")
	}

	if err != ErrDailyLimitExceeded {
		t.Errorf("expected ErrDailyLimitExceeded, got %v", err)
	}
}

// TestRateLimiterContextCancelled tests Wait with cancelled context.
func TestRateLimiterContextCancelled(t *testing.T) {
	rl := NewRateLimiter(1, 100) // Only 1 per minute
	ctx, cancel := context.WithCancel(context.Background())

	// Use the single request
	_ = rl.Wait(ctx)

	// Cancel context
	cancel()

	// Next Wait should fail with context error
	err := rl.Wait(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// TestRateLimiterContextTimeout tests Wait with context timeout.
func TestRateLimiterContextTimeout(t *testing.T) {
	rl := NewRateLimiter(1, 100)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Use the single request
	_ = rl.Wait(ctx)

	// Wait should return due to context timeout
	start := time.Now()
	err := rl.Wait(ctx)

	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}

	if time.Since(start) > 500*time.Millisecond {
		t.Errorf("wait took too long: %v", time.Since(start))
	}
}

// TestRateLimiterRetryAfterCap tests that retry-after is capped.
func TestRateLimiterRetryAfterCap(t *testing.T) {
	rl := NewRateLimiter(10, 100)

	// Try to set a very long retry-after (should be capped at 5 minutes)
	rl.UpdateFromResponse(3600) // 1 hour

	stats := rl.Stats()
	duration := time.Until(stats.BlockedUntil)

	// Should be capped to 5 minutes (300 seconds)
	if duration > 6*time.Minute {
		t.Errorf("expected retry-after to be capped, got %v", duration)
	}
}

// TestRateLimiterMultipleRequests tests sequence of requests.
func TestRateLimiterMultipleRequests(t *testing.T) {
	rl := NewRateLimiter(5, 100)

	for i := 0; i < 5; i++ {
		ok, err := rl.TryAcquire()
		if !ok {
			t.Errorf("request %d failed: %v", i+1, err)
		}
	}

	// 6th should fail
	ok, err := rl.TryAcquire()
	if ok {
		t.Fatal("expected 6th request to fail due to minute limit")
	}
	if err != ErrRateLimitExceeded {
		t.Errorf("expected ErrRateLimitExceeded, got %v", err)
	}
}

// TestRateLimiterStatsValues tests that stats values are reasonable.
func TestRateLimiterStatsValues(t *testing.T) {
	rl := NewRateLimiter(30, 1000)

	stats := rl.Stats()

	// MinuteRemaining should be between 0 and MinuteLimit
	if stats.MinuteRemaining < 0 || stats.MinuteRemaining > 30 {
		t.Errorf("MinuteRemaining %d out of bounds", stats.MinuteRemaining)
	}

	// DailyRemaining should be between 0 and DailyLimit
	if stats.DailyRemaining < 0 || stats.DailyRemaining > 1000 {
		t.Errorf("DailyRemaining %d out of bounds", stats.DailyRemaining)
	}

	// BlockedUntil should be in the future or zero
	if !stats.BlockedUntil.IsZero() && stats.BlockedUntil.Before(time.Now()) {
		t.Errorf("BlockedUntil should be in future or zero, got %v", stats.BlockedUntil)
	}
}

// TestRateLimiterBlockedUntilReset tests that BlockedUntil expires.
func TestRateLimiterBlockedUntilReset(t *testing.T) {
	rl := NewRateLimiter(10, 100)

	// Set a short block
	rl.UpdateFromResponse(1) // 1 second

	// Should be blocked now
	ok, err := rl.TryAcquire()
	if ok {
		t.Fatal("expected to be blocked")
	}
	if err != ErrRateLimitExceeded {
		t.Errorf("expected ErrRateLimitExceeded, got %v", err)
	}

	// Wait for block to expire
	time.Sleep(1100 * time.Millisecond)

	// Should be allowed now
	ok, err = rl.TryAcquire()
	if !ok {
		t.Fatalf("expected to be allowed after block expires, got error: %v", err)
	}
}

// TestRateLimiterEdgeCases tests edge cases.
func TestRateLimiterEdgeCases(t *testing.T) {
	// Very low limits
	rl := NewRateLimiter(1, 1)

	// Use the single daily request
	_ = rl.Wait(context.Background())

	// Both limits should be exhausted
	ok, err := rl.TryAcquire()
	if ok {
		t.Fatal("expected to be rate limited")
	}
	if err == nil {
		t.Fatal("expected error when limits exhausted")
	}
}

// TestRateLimiterZeroLimits tests behavior with zero limits.
func TestRateLimiterZeroLimits(t *testing.T) {
	rl := NewRateLimiter(0, 0)

	// Should fail immediately
	ok, err := rl.TryAcquire()
	if ok {
		t.Fatal("expected to fail with zero limits")
	}
	if err != ErrDailyLimitExceeded {
		t.Errorf("expected ErrDailyLimitExceeded, got %v", err)
	}
}

// TestRateLimiterHighLimits tests behavior with very high limits.
func TestRateLimiterHighLimits(t *testing.T) {
	rl := NewRateLimiter(10000, 1000000)

	// Should allow many requests without exhaustion
	for i := 0; i < 100; i++ {
		ok, err := rl.TryAcquire()
		if !ok {
			t.Errorf("request %d failed: %v", i+1, err)
		}
	}

	stats := rl.Stats()
	if stats.MinuteRemaining <= 0 {
		t.Error("expected minute remaining to be positive with high limits")
	}
}

// skipIfNoAPIKeyRateLimit skips test if MOUSER_API_KEY is not set
func skipIfNoAPIKeyRateLimit(t *testing.T) {
	rateLimitTestInit()
	if rateLimitTestAPIKey == "" {
		t.Skip("MOUSER_API_KEY not found in environment or .env file")
	}
}

// TestRateLimitingWithRealAPI tests that rate limiter works with real API calls.
func TestRateLimitingWithRealAPI(t *testing.T) {
	skipIfNoAPIKeyRateLimit(t)

	apiKey := rateLimitTestAPIKey
	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Make requests and verify rate limiting doesn't block
	initialStats := client.RateLimitStats()

	_, err = client.KeywordSearch(ctx, SearchOptions{
		Keyword: "resistor",
		Records: 2,
	})

	if err != nil {
		t.Errorf("first API call failed: %v", err)
	}

	// Check that rate limit was decremented
	afterStats := client.RateLimitStats()
	if afterStats.MinuteRemaining >= initialStats.MinuteRemaining {
		t.Logf("Warning: minute limit not decremented (before: %d, after: %d)",
			initialStats.MinuteRemaining, afterStats.MinuteRemaining)
	}

	if afterStats.DailyRemaining >= initialStats.DailyRemaining {
		t.Logf("Warning: daily limit not decremented (before: %d, after: %d)",
			initialStats.DailyRemaining, afterStats.DailyRemaining)
	}
}
