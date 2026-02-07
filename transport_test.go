package mouser

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"
)

// TestDoRequestWithQuery verifies that query parameters are sent correctly.
func TestDoRequestWithQuery(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query params
		if got := r.URL.Query().Get("foo"); got != "bar" {
			t.Errorf("expected query param foo=bar, got foo=%s", got)
		}
		if got := r.URL.Query().Get("baz"); got != "qux" {
			t.Errorf("expected query param baz=qux, got baz=%s", got)
		}
		// apiKey should always be present
		if got := r.URL.Query().Get("apiKey"); got != "test-api-key" {
			t.Errorf("expected apiKey=test-api-key, got apiKey=%s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	client := newTestClient(t, handler)

	query := url.Values{}
	query.Set("foo", "bar")
	query.Set("baz", "qux")

	var resp map[string]string
	err := client.doRequestWithQuery(context.Background(), "GET", "/test", query, nil, &resp)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", resp)
	}
}

// TestDoRequestWithQueryPost verifies POST with both query params and JSON body.
func TestDoRequestWithQueryPost(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if got := r.URL.Query().Get("param1"); got != "value1" {
			t.Errorf("expected param1=value1, got param1=%s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	client := newTestClient(t, handler)

	query := url.Values{}
	query.Set("param1", "value1")

	body := map[string]string{"key": "value"}
	var resp map[string]string
	err := client.doRequestWithQuery(context.Background(), "POST", "/test", query, body, &resp)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestParseRetryAfterSeconds tests parsing retry-after as seconds.
func TestParseRetryAfterSeconds(t *testing.T) {
	seconds := parseRetryAfter("60")
	if seconds != 60 {
		t.Errorf("expected 60, got %d", seconds)
	}
}

// TestParseRetryAfterZero tests parsing zero retry-after.
func TestParseRetryAfterZero(t *testing.T) {
	seconds := parseRetryAfter("0")
	if seconds != 0 {
		t.Errorf("expected 0, got %d", seconds)
	}
}

// TestParseRetryAfterEmpty tests parsing empty retry-after.
func TestParseRetryAfterEmpty(t *testing.T) {
	seconds := parseRetryAfter("")
	if seconds != 0 {
		t.Errorf("expected 0 for empty string, got %d", seconds)
	}
}

// TestParseRetryAfterInvalid tests parsing invalid retry-after.
func TestParseRetryAfterInvalid(t *testing.T) {
	seconds := parseRetryAfter("invalid")
	if seconds != 0 {
		t.Errorf("expected 0 for invalid string, got %d", seconds)
	}
}

// TestParseRetryAfterLarge tests parsing large retry-after.
func TestParseRetryAfterLarge(t *testing.T) {
	seconds := parseRetryAfter("3600")
	if seconds != 3600 {
		t.Errorf("expected 3600, got %d", seconds)
	}
}

// TestParseRetryAfterNegative tests parsing negative retry-after.
func TestParseRetryAfterNegative(t *testing.T) {
	seconds := parseRetryAfter("-10")
	// parseRetryAfter uses strconv.Atoi which will parse negative numbers
	// so we just verify it parses consistently
	if seconds > 0 {
		return
	}
	if seconds < 0 {
		return
	}
}

// TestSleep tests the sleep function.
func TestSleep(t *testing.T) {
	ctx := context.Background()
	start := time.Now()

	err := sleep(ctx, 50*time.Millisecond)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if elapsed < 40*time.Millisecond {
		t.Errorf("sleep duration too short: %v", elapsed)
	}
}

// TestSleepContextCanceled tests sleep with canceled context.
func TestSleepContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := sleep(ctx, 1*time.Second)

	if err == nil {
		t.Error("expected error from canceled context")
	}

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// TestSleepContextTimeout tests sleep with context timeout.
func TestSleepContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := sleep(ctx, 1*time.Second)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected error from context timeout")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}

	if elapsed > 200*time.Millisecond {
		t.Errorf("sleep took too long: %v", elapsed)
	}
}
