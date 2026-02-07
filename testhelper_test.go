package mouser

import (
	"context"
	"net/http"
	"net/http/httptest"
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
