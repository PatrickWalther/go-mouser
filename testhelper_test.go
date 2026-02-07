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

// TestSearchAllByManufacturerMock tests multi-page iteration.
func TestSearchAllByManufacturerMock(t *testing.T) {
	page := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		w.Header().Set("Content-Type", "application/json")

		var parts string
		if page == 1 {
			// Return MaxRecords (50) to signal more pages
			items := make([]string, MaxRecords)
			for i := range items {
				items[i] = `{"MouserPartNumber": "P` + string(rune('A'+i%26)) + `"}`
			}
			parts = ""
			for i, item := range items {
				if i > 0 {
					parts += ","
				}
				parts += item
			}
		} else {
			// Second page: fewer than MaxRecords, signals end
			parts = `{"MouserPartNumber": "LAST"}`
		}

		_, _ = w.Write([]byte(`{"Errors":[],"SearchResults":{"NumberOfResult":51,"Parts":[` + parts + `]}}`))
	})

	client := newTestClient(t, handler)

	var collected []string
	err := client.SearchAllByManufacturer(context.Background(),
		KeywordAndManufacturerSearchOptions{
			Keyword:          "test",
			ManufacturerName: "TestMfr",
		},
		func(p Part) bool {
			collected = append(collected, p.MouserPartNumber)
			return true
		})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(collected) != 51 {
		t.Errorf("expected 51 parts, got %d", len(collected))
	}
	if page != 2 {
		t.Errorf("expected 2 pages fetched, got %d", page)
	}
}

// TestSearchAllByManufacturerEarlyStopMock tests callback returning false.
func TestSearchAllByManufacturerEarlyStopMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Errors":[],"SearchResults":{"NumberOfResult":100,"Parts":[
			{"MouserPartNumber":"P1"},{"MouserPartNumber":"P2"},{"MouserPartNumber":"P3"}
		]}}`))
	})

	client := newTestClient(t, handler)

	count := 0
	err := client.SearchAllByManufacturer(context.Background(),
		KeywordAndManufacturerSearchOptions{Keyword: "test", ManufacturerName: "TestMfr"},
		func(p Part) bool {
			count++
			return count < 2 // stop after first
		})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected callback called 2 times (stop on 2nd), got %d", count)
	}
}

// TestSearchAllByManufacturerEmptyMock tests empty results.
func TestSearchAllByManufacturerEmptyMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Errors":[],"SearchResults":{"NumberOfResult":0,"Parts":[]}}`))
	})

	client := newTestClient(t, handler)

	count := 0
	err := client.SearchAllByManufacturer(context.Background(),
		KeywordAndManufacturerSearchOptions{Keyword: "nonexistent", ManufacturerName: "Nobody"},
		func(p Part) bool {
			count++
			return true
		})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 callback calls for empty results, got %d", count)
	}
}

// --- Mock tests for existing search endpoints ---

func searchResponseJSON(n int) string {
	parts := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			parts += ","
		}
		parts += `{"MouserPartNumber":"MOCK-` + string(rune('0'+i)) + `","Description":"Mock Part"}`
	}
	return `{"Errors":[],"SearchResults":{"NumberOfResult":` + json.Number(string(rune('0'+n))).String() + `,"Parts":[` + parts + `]}}`
}

// TestKeywordSearchMock tests KeywordSearch without an API key.
func TestKeywordSearchMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/search/keyword" {
			t.Errorf("expected path /search/keyword, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var raw map[string]json.RawMessage
		json.Unmarshal(body, &raw)

		var inner map[string]json.RawMessage
		json.Unmarshal(raw["SearchByKeywordRequest"], &inner)

		var keyword string
		json.Unmarshal(inner["keyword"], &keyword)
		if keyword != "capacitor" {
			t.Errorf("expected keyword=capacitor, got %s", keyword)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Errors": [],
			"SearchResults": {
				"NumberOfResult": 2,
				"Parts": [
					{"MouserPartNumber": "CAP-001", "Description": "Ceramic Capacitor"},
					{"MouserPartNumber": "CAP-002", "Description": "Electrolytic Capacitor"}
				]
			}
		}`))
	})

	client := newTestClient(t, handler)
	result, err := client.KeywordSearch(context.Background(), SearchOptions{
		Keyword:      "capacitor",
		Records:      10,
		SearchOption: SearchOptionInStock,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NumberOfResult != 2 {
		t.Errorf("expected 2 results, got %d", result.NumberOfResult)
	}
	if len(result.Parts) != 2 {
		t.Errorf("expected 2 parts, got %d", len(result.Parts))
	}
}

// TestPartNumberSearchMock tests PartNumberSearch (verifies bug fix from Phase 1).
func TestPartNumberSearchMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/search/partnumber" {
			t.Errorf("expected path /search/partnumber, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var raw map[string]json.RawMessage
		json.Unmarshal(body, &raw)

		var inner map[string]json.RawMessage
		json.Unmarshal(raw["SearchByPartRequest"], &inner)

		// Verify bug fix: partSearchOptions not searchOptions
		if _, ok := inner["partSearchOptions"]; !ok {
			t.Error("missing partSearchOptions in request")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Errors": [],
			"SearchResults": {
				"NumberOfResult": 1,
				"Parts": [{"MouserPartNumber": "LM386N-1", "ManufacturerPartNumber": "LM386N-1/NOPB"}]
			}
		}`))
	})

	client := newTestClient(t, handler)
	result, err := client.PartNumberSearch(context.Background(), PartNumberSearchOptions{
		PartNumber:       "LM386",
		PartSearchOption: PartSearchOptionExact,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NumberOfResult != 1 {
		t.Errorf("expected 1 result, got %d", result.NumberOfResult)
	}
}

// TestKeywordAndManufacturerSearchMock tests V2 keyword+manufacturer search.
func TestKeywordAndManufacturerSearchMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/search/keywordandmanufacturer" {
			t.Errorf("expected path /search/keywordandmanufacturer, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var raw map[string]json.RawMessage
		json.Unmarshal(body, &raw)

		var inner map[string]json.RawMessage
		json.Unmarshal(raw["SearchByKeywordMfrNameRequest"], &inner)

		var mfr string
		json.Unmarshal(inner["manufacturerName"], &mfr)
		if mfr != "Texas Instruments" {
			t.Errorf("expected manufacturerName=Texas Instruments, got %s", mfr)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Errors": [],
			"SearchResults": {"NumberOfResult": 1, "Parts": [{"MouserPartNumber": "TI-001"}]}
		}`))
	})

	client := newTestClient(t, handler)
	result, err := client.KeywordAndManufacturerSearch(context.Background(), KeywordAndManufacturerSearchOptions{
		Keyword:          "microcontroller",
		ManufacturerName: "Texas Instruments",
		Records:          10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NumberOfResult != 1 {
		t.Errorf("expected 1 result, got %d", result.NumberOfResult)
	}
}

// TestPartNumberAndManufacturerSearchMock tests V2 part number+manufacturer search.
func TestPartNumberAndManufacturerSearchMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search/partnumberandmanufacturer" {
			t.Errorf("expected path /search/partnumberandmanufacturer, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var raw map[string]json.RawMessage
		json.Unmarshal(body, &raw)

		var inner map[string]json.RawMessage
		json.Unmarshal(raw["SearchByPartMfrNameRequest"], &inner)

		var mfr string
		json.Unmarshal(inner["manufacturerName"], &mfr)
		if mfr != "Vishay" {
			t.Errorf("expected manufacturerName=Vishay, got %s", mfr)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Errors": [],
			"SearchResults": {"NumberOfResult": 1, "Parts": [{"MouserPartNumber": "V-001"}]}
		}`))
	})

	client := newTestClient(t, handler)
	result, err := client.PartNumberAndManufacturerSearch(context.Background(), PartNumberAndManufacturerSearchOptions{
		PartNumber:       "RN73H",
		ManufacturerName: "Vishay",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NumberOfResult != 1 {
		t.Errorf("expected 1 result, got %d", result.NumberOfResult)
	}
}

// TestGetManufacturerListMock tests GetManufacturerList without an API key.
func TestGetManufacturerListMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/search/manufacturerlist" {
			t.Errorf("expected path /search/manufacturerlist, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Errors": [],
			"MouserManufacturerList": {
				"Count": 3,
				"ManufacturerList": [
					{"ManufacturerName": "Texas Instruments"},
					{"ManufacturerName": "STMicroelectronics"},
					{"ManufacturerName": "Microchip"}
				]
			}
		}`))
	})

	client := newTestClient(t, handler)
	result, err := client.GetManufacturerList(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Count != 3 {
		t.Errorf("expected Count=3, got %d", result.Count)
	}
	if len(result.ManufacturerList) != 3 {
		t.Errorf("expected 3 manufacturers, got %d", len(result.ManufacturerList))
	}
}

// TestGetPartDetailsMock tests GetPartDetails without an API key.
func TestGetPartDetailsMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Errors": [],
			"SearchResults": {
				"NumberOfResult": 1,
				"Parts": [{
					"MouserPartNumber": "STM32F407VGT6",
					"ManufacturerPartNumber": "STM32F407VGT6",
					"Manufacturer": "STMicroelectronics",
					"Description": "ARM Microcontroller",
					"PriceBreaks": [{"Quantity": 1, "Price": "$15.00", "Currency": "USD"}]
				}]
			}
		}`))
	})

	client := newTestClient(t, handler)
	part, err := client.GetPartDetails(context.Background(), "STM32F407VGT6")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if part.MouserPartNumber != "STM32F407VGT6" {
		t.Errorf("expected part number STM32F407VGT6, got %s", part.MouserPartNumber)
	}
	if len(part.PriceBreaks) != 1 {
		t.Errorf("expected 1 price break, got %d", len(part.PriceBreaks))
	}
}

// TestGetPartDetailsNotFoundMock tests GetPartDetails with no results.
func TestGetPartDetailsNotFoundMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Errors":[],"SearchResults":{"NumberOfResult":0,"Parts":[]}}`))
	})

	client := newTestClient(t, handler)
	_, err := client.GetPartDetails(context.Background(), "NONEXISTENT-PART")
	if err == nil {
		t.Fatal("expected error for not found part")
	}
}

// TestSearchAllMock tests the SearchAll paginated iterator with mock data.
func TestSearchAllMock(t *testing.T) {
	page := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		w.Header().Set("Content-Type", "application/json")

		if page == 1 {
			// Return MaxRecords to signal more pages
			parts := ""
			for i := 0; i < MaxRecords; i++ {
				if i > 0 {
					parts += ","
				}
				parts += `{"MouserPartNumber":"P-` + string(rune('A'+i%26)) + `"}`
			}
			_, _ = w.Write([]byte(`{"Errors":[],"SearchResults":{"NumberOfResult":55,"Parts":[` + parts + `]}}`))
		} else {
			_, _ = w.Write([]byte(`{"Errors":[],"SearchResults":{"NumberOfResult":55,"Parts":[
				{"MouserPartNumber":"P-LAST1"},
				{"MouserPartNumber":"P-LAST2"},
				{"MouserPartNumber":"P-LAST3"},
				{"MouserPartNumber":"P-LAST4"},
				{"MouserPartNumber":"P-LAST5"}
			]}}`))
		}
	})

	client := newTestClient(t, handler)

	var collected int
	err := client.SearchAll(context.Background(), SearchOptions{Keyword: "test"}, func(p Part) bool {
		collected++
		return true
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if collected != 55 {
		t.Errorf("expected 55 parts, got %d", collected)
	}
	if page != 2 {
		t.Errorf("expected 2 pages, got %d", page)
	}
}

// TestSearchAPIErrorMock tests that API errors are properly returned.
func TestSearchAPIErrorMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Errors":[{"Id":1,"Code":"InvalidKeyword","Message":"Keyword is required"}],"SearchResults":{"NumberOfResult":0,"Parts":[]}}`))
	})

	client := newTestClient(t, handler)
	_, err := client.KeywordSearch(context.Background(), SearchOptions{Keyword: ""})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

// TestSearchHTTPErrorMock tests that HTTP errors are properly returned.
func TestSearchHTTPErrorMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
	})

	client := newTestClient(t, handler)
	_, err := client.KeywordSearch(context.Background(), SearchOptions{Keyword: "test"})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
