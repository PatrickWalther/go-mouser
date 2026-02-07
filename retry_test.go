package mouser

import (
	"testing"
	"time"
)

// TestDefaultRetryConfig tests default retry configuration.
func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("expected MaxRetries 3, got %d", config.MaxRetries)
	}
	if config.InitialBackoff != 500*time.Millisecond {
		t.Errorf("expected InitialBackoff 500ms, got %v", config.InitialBackoff)
	}
	if config.MaxBackoff != 30*time.Second {
		t.Errorf("expected MaxBackoff 30s, got %v", config.MaxBackoff)
	}
	if config.Multiplier != 2.0 {
		t.Errorf("expected Multiplier 2.0, got %f", config.Multiplier)
	}
}

// TestNoRetry tests disabled retry configuration.
func TestNoRetry(t *testing.T) {
	config := NoRetry()

	if config.MaxRetries != 0 {
		t.Errorf("expected MaxRetries 0, got %d", config.MaxRetries)
	}
}

// TestCalculateBackoff tests backoff calculation.
func TestCalculateBackoff(t *testing.T) {
	config := RetryConfig{
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
		Multiplier:     2.0,
		Jitter:         0.0, // Disable jitter for predictability
	}

	backoff1 := config.calculateBackoff(0)
	backoff2 := config.calculateBackoff(1)
	backoff3 := config.calculateBackoff(2)

	if backoff1 != 100*time.Millisecond {
		t.Errorf("expected first backoff 100ms, got %v", backoff1)
	}

	if backoff2 != 200*time.Millisecond {
		t.Errorf("expected second backoff 200ms, got %v", backoff2)
	}

	if backoff3 != 400*time.Millisecond {
		t.Errorf("expected third backoff 400ms, got %v", backoff3)
	}
}

// TestCalculateBackoffWithJitter tests backoff with jitter.
func TestCalculateBackoffWithJitter(t *testing.T) {
	config := RetryConfig{
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
		Multiplier:     2.0,
		Jitter:         0.1,
	}

	backoff1 := config.calculateBackoff(0)
	backoff2 := config.calculateBackoff(0)

	// Both should be around 100ms but might differ due to jitter
	if backoff1 < 50*time.Millisecond || backoff1 > 150*time.Millisecond {
		t.Errorf("backoff with jitter out of range: %v", backoff1)
	}

	if backoff2 < 50*time.Millisecond || backoff2 > 150*time.Millisecond {
		t.Errorf("backoff with jitter out of range: %v", backoff2)
	}
}

// TestCalculateBackoffMaxCap tests backoff max cap.
func TestCalculateBackoffMaxCap(t *testing.T) {
	config := RetryConfig{
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     2 * time.Second,
		Multiplier:     2.0,
		Jitter:         0.0,
	}

	backoff1 := config.calculateBackoff(0)
	backoff2 := config.calculateBackoff(1)
	backoff3 := config.calculateBackoff(2)
	backoff4 := config.calculateBackoff(3)

	if backoff1 != 1*time.Second {
		t.Errorf("expected 1s, got %v", backoff1)
	}
	if backoff2 != 2*time.Second {
		t.Errorf("expected 2s, got %v", backoff2)
	}
	if backoff3 != 2*time.Second {
		t.Errorf("expected capped 2s, got %v", backoff3)
	}
	if backoff4 != 2*time.Second {
		t.Errorf("expected capped 2s, got %v", backoff4)
	}
}

// TestShouldRetryRateLimited tests shouldRetry for 429.
func TestShouldRetryRateLimited(t *testing.T) {
	if !shouldRetry(nil, 429) {
		t.Error("expected shouldRetry to return true for 429")
	}
}

// TestShouldRetryServerError tests shouldRetry for 500.
func TestShouldRetryServerError(t *testing.T) {
	if !shouldRetry(nil, 500) {
		t.Error("expected shouldRetry to return true for 500")
	}
}

// TestShouldRetryBadGateway tests shouldRetry for 502.
func TestShouldRetryBadGateway(t *testing.T) {
	if !shouldRetry(nil, 502) {
		t.Error("expected shouldRetry to return true for 502")
	}
}

// TestShouldRetryServiceUnavailable tests shouldRetry for 503.
func TestShouldRetryServiceUnavailable(t *testing.T) {
	if !shouldRetry(nil, 503) {
		t.Error("expected shouldRetry to return true for 503")
	}
}

// TestShouldRetryGatewayTimeout tests shouldRetry for 504.
func TestShouldRetryGatewayTimeout(t *testing.T) {
	if !shouldRetry(nil, 504) {
		t.Error("expected shouldRetry to return true for 504")
	}
}

// TestShouldRetryClientError tests shouldRetry for 400.
func TestShouldRetryClientError(t *testing.T) {
	if shouldRetry(nil, 400) {
		t.Error("expected shouldRetry to return false for 400")
	}
}

// TestShouldRetryUnauthorized tests shouldRetry for 401.
func TestShouldRetryUnauthorized(t *testing.T) {
	if shouldRetry(nil, 401) {
		t.Error("expected shouldRetry to return false for 401")
	}
}

// TestShouldRetryForbidden tests shouldRetry for 403.
func TestShouldRetryForbidden(t *testing.T) {
	if shouldRetry(nil, 403) {
		t.Error("expected shouldRetry to return false for 403")
	}
}

// TestShouldRetryNotFound tests shouldRetry for 404.
func TestShouldRetryNotFound(t *testing.T) {
	if shouldRetry(nil, 404) {
		t.Error("expected shouldRetry to return false for 404")
	}
}

// TestShouldRetrySuccess tests shouldRetry for 200.
func TestShouldRetrySuccess(t *testing.T) {
	if shouldRetry(nil, 200) {
		t.Error("expected shouldRetry to return false for 200")
	}
}

// TestShouldRetryTimeoutError tests shouldRetry for timeout errors.
func TestShouldRetryTimeoutError(t *testing.T) {
	timeoutErr := &timeoutError{}
	if !shouldRetry(timeoutErr, 0) {
		t.Error("expected shouldRetry to return true for timeout error")
	}
}

// TestShouldRetryNilError tests shouldRetry with nil error.
func TestShouldRetryNilError(t *testing.T) {
	// Nil error with success code should not retry
	if shouldRetry(nil, 200) {
		t.Error("expected false for nil error with 200")
	}

	// Nil error with server error should retry
	if !shouldRetry(nil, 500) {
		t.Error("expected true for nil error with 500")
	}
}

// Helper timeout error for testing
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

// TestPow tests the pow helper function.
func TestPow(t *testing.T) {
	tests := []struct {
		base     float64
		exp      float64
		expected float64
	}{
		{2.0, 0, 1.0},
		{2.0, 1, 2.0},
		{2.0, 2, 4.0},
		{2.0, 3, 8.0},
		{3.0, 2, 9.0},
		{10.0, 0, 1.0},
		{1.0, 10, 1.0},
	}

	for _, test := range tests {
		result := pow(test.base, test.exp)
		if result != test.expected {
			t.Errorf("pow(%f, %f) = %f, expected %f", test.base, test.exp, result, test.expected)
		}
	}
}

// TestIsTimeoutError tests isTimeoutError function.
func TestIsTimeoutError(t *testing.T) {
	if !isTimeoutError(&timeoutError{}) {
		t.Error("expected isTimeoutError to recognize timeout error")
	}

	if isTimeoutError(nil) {
		t.Error("expected isTimeoutError to return false for nil")
	}

	if isTimeoutError(ErrRateLimitExceeded) {
		t.Error("expected isTimeoutError to return false for non-timeout error")
	}
}

// TestIsTemporaryNetworkError tests isTemporaryNetworkError function.
func TestIsTemporaryNetworkError(t *testing.T) {
	timeoutErr := &timeoutError{}

	if !isTemporaryNetworkError(timeoutErr) {
		t.Error("expected isTemporaryNetworkError to recognize timeout error")
	}

	if isTemporaryNetworkError(nil) {
		t.Error("expected isTemporaryNetworkError to return false for nil")
	}
}

// TestRetryConfigDefaults tests that custom config can override defaults.
func TestRetryConfigDefaults(t *testing.T) {
	customConfig := RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 200 * time.Millisecond,
		MaxBackoff:     10 * time.Second,
		Multiplier:     1.5,
		Jitter:         0.05,
	}

	if customConfig.MaxRetries != 5 {
		t.Errorf("expected MaxRetries 5, got %d", customConfig.MaxRetries)
	}
	if customConfig.InitialBackoff != 200*time.Millisecond {
		t.Errorf("expected InitialBackoff 200ms, got %v", customConfig.InitialBackoff)
	}
}

// TestBackoffProgression tests the progression of backoffs.
func TestBackoffProgression(t *testing.T) {
	config := RetryConfig{
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         0.0,
	}

	backoffs := make([]time.Duration, 0)
	for i := 0; i < 10; i++ {
		backoff := config.calculateBackoff(i)
		backoffs = append(backoffs, backoff)
	}

	// Each backoff should be >= previous (with our multiplier > 1)
	for i := 1; i < len(backoffs); i++ {
		if backoffs[i] < backoffs[i-1] {
			t.Errorf("backoff progression should be non-decreasing, but %v < %v", backoffs[i], backoffs[i-1])
		}
	}

	// All should be <= MaxBackoff
	for i, backoff := range backoffs {
		if backoff > 100*time.Millisecond {
			t.Errorf("backoff %d (%v) exceeds max", i, backoff)
		}
	}
}
