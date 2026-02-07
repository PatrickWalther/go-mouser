package mouser

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func orderOptionsResponse() string {
	return `{
		"Errors": [],
		"CurrencyCode": "USD",
		"BillingAddress": {"CountryCode": "US", "City": "Austin"},
		"ShippingAddress": {"CountryCode": "US", "City": "Dallas"},
		"Shipping": {
			"Methods": [{"Method": "FedEx Ground", "Rate": 12.50, "Code": 1}],
			"FreightAccounts": [{"Number": "FA-001", "Type": "FedEx"}]
		},
		"Payment": {
			"PaymentTypes": ["CreditCard", "PurchaseOrder"],
			"TaxCertificates": ["TC-001"],
			"VatAccounts": []
		},
		"Languages": ["en", "de", "fr"]
	}`
}

func currenciesResponse() string {
	return `{
		"Errors": [],
		"Currencies": [
			{"CurrencyCode": "USD", "CurrencyName": "US Dollar"},
			{"CurrencyCode": "EUR", "CurrencyName": "Euro"},
			{"CurrencyCode": "GBP", "CurrencyName": "British Pound"}
		]
	}`
}

func countriesResponse() string {
	return `{
		"Errors": [],
		"Countries": [
			{
				"CountryName": "United States",
				"CountryCode": "US",
				"States": [
					{"StateName": "Texas", "StateCode": "TX"},
					{"StateName": "California", "StateCode": "CA"}
				]
			}
		]
	}`
}

func orderResponse() string {
	return `{
		"Errors": [],
		"OrderNumber": "ORD-001",
		"CartKey": "abc-123",
		"CurrencyCode": "USD",
		"OrderLines": [{
			"Quantity": 10,
			"UnitPrice": 5.00,
			"ExtPrice": 50.00,
			"ProductInfo": {"MouserPartNumber": "TEST-001"}
		}],
		"BillingAddress": {"City": "Austin"},
		"ShippingAddress": {"City": "Dallas"},
		"SummaryDetail": {"MerchandiseTotal": 50.00, "OrderTotal": 55.00, "AdditionalFeesTotal": 5.00}
	}`
}

// TestQueryOrderOptionsMock tests QueryOrderOptions with a mock server.
func TestQueryOrderOptionsMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/order/options/query" {
			t.Errorf("expected path /order/options/query, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req orderOptionsRequestWrapper
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request: %v", err)
		}
		if req.OrderOptionsRequest.CartKey != "abc-123" {
			t.Errorf("expected CartKey=abc-123, got %s", req.OrderOptionsRequest.CartKey)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(orderOptionsResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.QueryOrderOptions(context.Background(), OrderOptionsRequest{
		CartKey:      "abc-123",
		CurrencyCode: "USD",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CurrencyCode != "USD" {
		t.Errorf("expected CurrencyCode=USD, got %s", resp.CurrencyCode)
	}
	if len(resp.Shipping.Methods) != 1 {
		t.Fatalf("expected 1 shipping method")
	}
	if resp.Shipping.Methods[0].Method != "FedEx Ground" {
		t.Errorf("expected shipping method=FedEx Ground")
	}
	if resp.Shipping.Methods[0].Rate != 12.50 {
		t.Errorf("expected Rate=12.50, got %f", resp.Shipping.Methods[0].Rate)
	}
	if len(resp.Payment.PaymentTypes) != 2 {
		t.Errorf("expected 2 payment types")
	}
}

// TestGetCurrenciesMock tests GetCurrencies with a mock server.
func TestGetCurrenciesMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/order/currencies" {
			t.Errorf("expected path /order/currencies, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("shippingCountryCode"); got != "US" {
			t.Errorf("expected shippingCountryCode=US, got %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(currenciesResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.GetCurrencies(context.Background(), "US")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Currencies) != 3 {
		t.Fatalf("expected 3 currencies, got %d", len(resp.Currencies))
	}
	if resp.Currencies[0].CurrencyCode != "USD" {
		t.Errorf("expected first currency=USD, got %s", resp.Currencies[0].CurrencyCode)
	}
}

// TestGetCurrenciesCachedMock tests that GetCurrencies responses are cached.
func TestGetCurrenciesCachedMock(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(currenciesResponse()))
	})

	client := newTestClientCached(t, handler)

	_, _ = client.GetCurrencies(context.Background(), "US")
	_, _ = client.GetCurrencies(context.Background(), "US")

	if callCount != 1 {
		t.Errorf("expected 1 server call (cached), got %d", callCount)
	}
}

// TestGetCountriesMock tests GetCountries with a mock server.
func TestGetCountriesMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/order/countries" {
			t.Errorf("expected path /order/countries, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("countryCode"); got != "US" {
			t.Errorf("expected countryCode=US, got %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(countriesResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.GetCountries(context.Background(), "US")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Countries) != 1 {
		t.Fatalf("expected 1 country, got %d", len(resp.Countries))
	}
	if resp.Countries[0].CountryName != "United States" {
		t.Errorf("expected CountryName=United States")
	}
	if len(resp.Countries[0].States) != 2 {
		t.Errorf("expected 2 states, got %d", len(resp.Countries[0].States))
	}
}

// TestGetCountriesCachedMock tests that GetCountries responses are cached.
func TestGetCountriesCachedMock(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(countriesResponse()))
	})

	client := newTestClientCached(t, handler)

	_, _ = client.GetCountries(context.Background(), "US")
	_, _ = client.GetCountries(context.Background(), "US")

	if callCount != 1 {
		t.Errorf("expected 1 server call (cached), got %d", callCount)
	}
}

// TestCreateOrderMock tests CreateOrder with a mock server.
func TestCreateOrderMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/order" {
			t.Errorf("expected path /order, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req createOrderRequestWrapper
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request: %v", err)
		}
		if req.CreateOrderRequest.CartKey != "abc-123" {
			t.Errorf("expected CartKey=abc-123, got %s", req.CreateOrderRequest.CartKey)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(orderResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.CreateOrder(context.Background(), CreateOrderRequest{
		CartKey:      "abc-123",
		CurrencyCode: "USD",
		SubmitOrder:  false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.OrderNumber != "ORD-001" {
		t.Errorf("expected OrderNumber=ORD-001, got %s", resp.OrderNumber)
	}
}

// TestCreateOrderFromPreviousMock tests CreateOrderFromPrevious with a mock server.
func TestCreateOrderFromPreviousMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/order/CreateFromOrder" {
			t.Errorf("expected path /order/CreateFromOrder, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("orderNumber"); got != "ORD-001" {
			t.Errorf("expected orderNumber=ORD-001, got %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(orderResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.CreateOrderFromPrevious(context.Background(), "ORD-001", "US", "USD", CreateOrderRequest{
		SubmitOrder: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.OrderNumber != "ORD-001" {
		t.Errorf("expected OrderNumber=ORD-001, got %s", resp.OrderNumber)
	}
}

// TestGetOrderDetailsMock tests GetOrderDetails with a mock server.
func TestGetOrderDetailsMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/order/ORD-001" {
			t.Errorf("expected path /order/ORD-001, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(orderResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.GetOrderDetails(context.Background(), "ORD-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.OrderNumber != "ORD-001" {
		t.Errorf("expected OrderNumber=ORD-001, got %s", resp.OrderNumber)
	}
	if resp.SummaryDetail.OrderTotal != 55.00 {
		t.Errorf("expected OrderTotal=55.00, got %f", resp.SummaryDetail.OrderTotal)
	}
}

// TestCreateCartFromOrderMock tests CreateCartFromOrder with a mock server.
func TestCreateCartFromOrderMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/order/item/CreateCartFromOrder" {
			t.Errorf("expected path /order/item/CreateCartFromOrder, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("orderNumber"); got != "ORD-001" {
			t.Errorf("expected orderNumber=ORD-001, got %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cartSuccessResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.CreateCartFromOrder(context.Background(), "ORD-001", "US", "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CartKey != "abc-123" {
		t.Errorf("expected CartKey=abc-123, got %s", resp.CartKey)
	}
}

// TestOrderErrorHandlingMock tests error handling for order endpoints.
func TestOrderErrorHandlingMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Errors":[{"Id":1,"Code":"InvalidOrder","Message":"Order not found"}]}`))
	})

	client := newTestClient(t, handler)
	_, err := client.GetOrderDetails(context.Background(), "BAD-ORDER")
	if err == nil {
		t.Fatal("expected error for invalid order")
	}
}

// TestOrderMutationsNotCachedMock tests that mutation endpoints are not cached.
func TestOrderMutationsNotCachedMock(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(orderResponse()))
	})

	client := newTestClientCached(t, handler)

	_, _ = client.CreateOrder(context.Background(), CreateOrderRequest{CartKey: "abc"})
	_, _ = client.CreateOrder(context.Background(), CreateOrderRequest{CartKey: "abc"})

	if callCount != 2 {
		t.Errorf("expected 2 server calls for mutations (no caching), got %d", callCount)
	}
}

// TestOrderModelRoundtrip tests JSON marshal/unmarshal for order models.
func TestOrderModelRoundtrip(t *testing.T) {
	original := CreateOrderRequest{
		ShippingAddress: &OrderAddress{
			AddressLocationTypeID: AddressLocationCommercial,
			CountryCode:           "US",
			FirstName:             "John",
			LastName:              "Doe",
			Company:               "Acme",
			AddressOne:            "123 Main St",
			City:                  "Austin",
			StateOrProvince:       "TX",
			PostalCode:            "78701",
			PhoneNumber:           "555-1234",
			EmailAddress:          "john@acme.com",
		},
		OrderType:       OrderTypeRush,
		PrimaryShipping: 1,
		CurrencyCode:    "USD",
		CartKey:         "abc-123",
		SubmitOrder:     false,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CreateOrderRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.CartKey != "abc-123" {
		t.Errorf("CartKey mismatch")
	}
	if decoded.ShippingAddress.City != "Austin" {
		t.Errorf("ShippingAddress.City mismatch")
	}
	if decoded.OrderType != OrderTypeRush {
		t.Errorf("OrderType mismatch")
	}
}

// TestCurrencyModelRoundtrip tests JSON marshal/unmarshal for currency models.
func TestCurrencyModelRoundtrip(t *testing.T) {
	original := CurrenciesResponse{
		Currencies: []Currency{
			{CurrencyCode: "USD", CurrencyName: "US Dollar"},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CurrenciesResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(decoded.Currencies) != 1 || decoded.Currencies[0].CurrencyCode != "USD" {
		t.Errorf("Currency mismatch: %+v", decoded.Currencies)
	}
}

// TestCountryModelRoundtrip tests JSON marshal/unmarshal for country models.
func TestCountryModelRoundtrip(t *testing.T) {
	original := CountriesResponse{
		Countries: []Country{
			{
				CountryName: "United States",
				CountryCode: "US",
				States: []State{
					{StateName: "Texas", StateCode: "TX"},
				},
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CountriesResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(decoded.Countries) != 1 || decoded.Countries[0].CountryCode != "US" {
		t.Errorf("Country mismatch")
	}
	if len(decoded.Countries[0].States) != 1 || decoded.Countries[0].States[0].StateCode != "TX" {
		t.Errorf("State mismatch")
	}
}

// Integration tests - only for safe read-only endpoints

// TestIntegrationGetCurrencies tests the real currencies endpoint.
func TestIntegrationGetCurrencies(t *testing.T) {
	skipIfNoKey(t)

	client, err := NewClient(testAPIKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	resp, err := client.GetCurrencies(context.Background(), "")
	if err != nil {
		t.Fatalf("GetCurrencies failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.Currencies) == 0 {
		t.Error("expected non-empty currencies list")
	}
	t.Logf("Found %d currencies", len(resp.Currencies))
}

// TestIntegrationGetCountries tests the real countries endpoint.
func TestIntegrationGetCountries(t *testing.T) {
	skipIfNoKey(t)

	client, err := NewClient(testAPIKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	resp, err := client.GetCountries(context.Background(), "")
	if err != nil {
		t.Fatalf("GetCountries failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.Countries) == 0 {
		t.Error("expected non-empty countries list")
	}
	t.Logf("Found %d countries", len(resp.Countries))
}
