package mouser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// newTestClient creates a *Client wired to an httptest.Server running the given handler.
// The server is registered for cleanup via t.Cleanup so callers don't need to close it.
func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := NewClient("test-api-key",
		WithBaseURL(server.URL),
		WithoutRetry(),
		WithoutCache(),
		WithRateLimiter(NewRateLimiter(10000, 100000)),
	)
	if err != nil {
		t.Fatalf("newTestClient: failed to create client: %v", err)
	}
	t.Cleanup(func() { client.Close() })

	return client
}

// newTestClientCached is like newTestClient but with caching enabled.
func newTestClientCached(t *testing.T, handler http.Handler) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := NewClient("test-api-key",
		WithBaseURL(server.URL),
		WithoutRetry(),
		WithRateLimiter(NewRateLimiter(10000, 100000)),
	)
	if err != nil {
		t.Fatalf("newTestClientCached: failed to create client: %v", err)
	}
	t.Cleanup(func() { client.Close() })

	return client
}

// TestNewTestClientRoundtrip verifies the mock server helper works end-to-end.
func TestNewTestClientRoundtrip(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"Errors": [],
			"SearchResults": {
				"NumberOfResult": 1,
				"Parts": [{"MouserPartNumber": "TEST-001", "Description": "Mock Part"}]
			}
		}`))
	})

	client := newTestClient(t, handler)

	result, err := client.KeywordSearch(context.Background(), SearchOptions{
		Keyword: "test",
		Records: 1,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.NumberOfResult != 1 {
		t.Errorf("expected 1 result, got %d", result.NumberOfResult)
	}
	if len(result.Parts) != 1 || result.Parts[0].MouserPartNumber != "TEST-001" {
		t.Errorf("unexpected parts: %+v", result.Parts)
	}
}

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
