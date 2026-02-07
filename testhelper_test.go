package mouser

import (
	"context"
	"encoding/json"
	"io"
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

// TestPartNumberSearchSendsPartSearchOptions verifies the JSON body contains
// "partSearchOptions" (not "searchOptions") for the part number search endpoint.
func TestPartNumberSearchSendsPartSearchOptions(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(body, &raw); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}

		var inner map[string]json.RawMessage
		if err := json.Unmarshal(raw["SearchByPartRequest"], &inner); err != nil {
			t.Fatalf("failed to parse SearchByPartRequest: %v", err)
		}

		if _, ok := inner["partSearchOptions"]; !ok {
			t.Errorf("expected 'partSearchOptions' in request body, got keys: %v", inner)
		}
		if _, ok := inner["searchOptions"]; ok {
			t.Errorf("unexpected 'searchOptions' key — should be 'partSearchOptions'")
		}

		// Verify the value
		var val string
		if err := json.Unmarshal(inner["partSearchOptions"], &val); err == nil {
			if val != "Exact" {
				t.Errorf("expected partSearchOptions=Exact, got %s", val)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Errors":[],"SearchResults":{"NumberOfResult":0,"Parts":[]}}`))
	})

	client := newTestClient(t, handler)
	_, _ = client.PartNumberSearch(context.Background(), PartNumberSearchOptions{
		PartNumber:       "TEST-123",
		PartSearchOption: PartSearchOptionExact,
	})
}

// TestPartNumberSearchOnlySendsExpectedFields verifies that only mouserPartNumber
// and partSearchOptions are sent in the request body (no records, startingRecord, etc.).
func TestPartNumberSearchOnlySendsExpectedFields(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var raw map[string]json.RawMessage
		json.Unmarshal(body, &raw)

		var inner map[string]json.RawMessage
		json.Unmarshal(raw["SearchByPartRequest"], &inner)

		// Only mouserPartNumber and partSearchOptions should be present
		allowed := map[string]bool{"mouserPartNumber": true, "partSearchOptions": true}
		for key := range inner {
			if !allowed[key] {
				t.Errorf("unexpected field %q in SearchByPartRequest", key)
			}
		}

		if _, ok := inner["records"]; ok {
			t.Error("records field should not be sent in part number search request")
		}
		if _, ok := inner["startingRecord"]; ok {
			t.Error("startingRecord field should not be sent in part number search request")
		}
		if _, ok := inner["searchWithYourSignUpLanguage"]; ok {
			t.Error("searchWithYourSignUpLanguage field should not be sent in part number search request")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Errors":[],"SearchResults":{"NumberOfResult":0,"Parts":[]}}`))
	})

	client := newTestClient(t, handler)
	_, _ = client.PartNumberSearch(context.Background(), PartNumberSearchOptions{
		PartNumber:                   "TEST-123",
		Records:                      10,
		StartingRecord:               5,
		SearchWithYourSignUpLanguage: true,
		PartSearchOption:             PartSearchOptionExact,
	})
}

// TestPartDeserializationNewFields verifies all new Part fields deserialize correctly.
func TestPartDeserializationNewFields(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Errors": [],
			"SearchResults": {
				"NumberOfResult": 1,
				"Parts": [{
					"MouserPartNumber": "TEST-001",
					"ActualMfrName": "Actual Corp",
					"AvailableOnOrder": "500",
					"PID": "PID-123",
					"REACH-SVHC": ["Lead", "Cadmium"],
					"RTM": "rtm-val",
					"SField": "sfield-val",
					"SalesMaximumOrderQty": "10000",
					"SurchargeMessages": [
						{"code": "SC01", "message": "Environmental surcharge"}
					],
					"VNum": "vnum-val",
					"ProductAttributes": [
						{"AttributeName": "Size", "AttributeValue": "0402", "AttributeCost": "0.01"}
					]
				}]
			}
		}`))
	})

	client := newTestClient(t, handler)
	result, err := client.KeywordSearch(context.Background(), SearchOptions{Keyword: "test", Records: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(result.Parts))
	}

	p := result.Parts[0]
	if p.ActualMfrName != "Actual Corp" {
		t.Errorf("ActualMfrName = %q, want %q", p.ActualMfrName, "Actual Corp")
	}
	if p.AvailableOnOrder != "500" {
		t.Errorf("AvailableOnOrder = %q, want %q", p.AvailableOnOrder, "500")
	}
	if p.PID != "PID-123" {
		t.Errorf("PID = %q, want %q", p.PID, "PID-123")
	}
	if len(p.REACH_SVHC) != 2 || p.REACH_SVHC[0] != "Lead" || p.REACH_SVHC[1] != "Cadmium" {
		t.Errorf("REACH_SVHC = %v, want [Lead Cadmium]", p.REACH_SVHC)
	}
	if p.RTM != "rtm-val" {
		t.Errorf("RTM = %q, want %q", p.RTM, "rtm-val")
	}
	if p.SField != "sfield-val" {
		t.Errorf("SField = %q, want %q", p.SField, "sfield-val")
	}
	if p.SalesMaximumOrderQty != "10000" {
		t.Errorf("SalesMaximumOrderQty = %q, want %q", p.SalesMaximumOrderQty, "10000")
	}
	if len(p.SurchargeMessages) != 1 || p.SurchargeMessages[0].Code != "SC01" || p.SurchargeMessages[0].Message != "Environmental surcharge" {
		t.Errorf("SurchargeMessages = %v, unexpected", p.SurchargeMessages)
	}
	if p.VNum != "vnum-val" {
		t.Errorf("VNum = %q, want %q", p.VNum, "vnum-val")
	}
	if len(p.ProductAttributes) != 1 || p.ProductAttributes[0].AttributeCost != "0.01" {
		t.Errorf("ProductAttributes[0].AttributeCost = %q, want %q", p.ProductAttributes[0].AttributeCost, "0.01")
	}
}
