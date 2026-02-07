package mouser

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func orderHistoryListResponse() string {
	return `{
		"Errors": [],
		"NumberOfOrders": 2,
		"OrderHistoryItems": [
			{
				"DateCreated": "2025-01-15",
				"SalesOrderNumber": "SO-001",
				"WebOrderNumber": "WO-001",
				"PoNumber": "PO-001",
				"BuyerName": "John Doe",
				"OrderStatusDisplay": "Shipped"
			},
			{
				"DateCreated": "2025-02-01",
				"SalesOrderNumber": "SO-002",
				"WebOrderNumber": "WO-002",
				"PoNumber": "PO-002",
				"BuyerName": "Jane Doe",
				"OrderStatusDisplay": "Processing"
			}
		]
	}`
}

func orderDetailResponse() string {
	return `{
		"Errors": [],
		"SalesOrderId": "SO-001",
		"WebOrderId": "WO-001",
		"OrderStatus": 3,
		"OrderStatusName": "Shipped",
		"OrderDate": "2025-01-15",
		"CurrencyCode": "USD",
		"IsScheduled": false,
		"IsPendingOrder": false,
		"BuyerName": "John Doe",
		"BillingAddress": {
			"CountryCode": "US",
			"CompanyName": "Acme Corp",
			"AddressOne": "123 Main St",
			"City": "Austin",
			"StateOrProvince": "TX",
			"PostalCode": "78701"
		},
		"ShippingAddress": {
			"CountryCode": "US",
			"CompanyName": "Acme Corp",
			"AddressOne": "456 Oak Ave",
			"City": "Austin",
			"StateOrProvince": "TX",
			"PostalCode": "78702"
		},
		"PaymentDetail": {
			"PoNumber": "PO-001",
			"PaymentMethodName": "Credit Card"
		},
		"DeliveryDetail": {
			"ShippingMethodName": "FedEx Ground",
			"TrackingDetails": [{"Number": "1234567890", "Link": "https://fedex.com/track/1234567890"}]
		},
		"SummaryDetail": {
			"MerchandiseTotal": 150.00,
			"OrderTotal": 165.50,
			"AdditionalFeesTotal": 15.50
		},
		"OrderLines": [{
			"Quantity": 10,
			"UnitPrice": 15.00,
			"ExtPrice": 150.00,
			"FormattedUnitPrice": "$15.00",
			"FormattedExtendedPrice": "$150.00",
			"ProductInfo": {
				"MouserPartNumber": "TEST-001",
				"ManufacturerName": "TestMfr",
				"ManufacturerPartNumber": "MFR-001",
				"PartDescription": "Test Part"
			},
			"Activities": [{"InvoiceNumber": "INV-001", "Date": "2025-01-16"}]
		}]
	}`
}

// TestGetOrderHistoryByDateFilterMock tests the date filter endpoint.
func TestGetOrderHistoryByDateFilterMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/orderhistory/ByDateFilter" {
			t.Errorf("expected path /orderhistory/ByDateFilter, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("dateFilter"); got != "All" {
			t.Errorf("expected dateFilter=All, got %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(orderHistoryListResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.GetOrderHistoryByDateFilter(context.Background(), DateFilterAll)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.NumberOfOrders != 2 {
		t.Errorf("expected 2 orders, got %d", resp.NumberOfOrders)
	}
	if len(resp.OrderHistoryItems) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.OrderHistoryItems))
	}
	if resp.OrderHistoryItems[0].SalesOrderNumber != "SO-001" {
		t.Errorf("expected SalesOrderNumber=SO-001, got %s", resp.OrderHistoryItems[0].SalesOrderNumber)
	}
}

// TestGetOrderHistoryByDateFilterEnumsMock tests that all date filter enum values are accepted.
func TestGetOrderHistoryByDateFilterEnumsMock(t *testing.T) {
	filters := []DateFilterType{
		DateFilterNone, DateFilterAll, DateFilterToday, DateFilterYesterday,
		DateFilterThisWeek, DateFilterLastWeek, DateFilterThisMonth, DateFilterLastMonth,
		DateFilterThisQuarter, DateFilterLastQuarter, DateFilterThisYear, DateFilterLastYear,
		DateFilterYearToDate,
	}

	for _, filter := range filters {
		t.Run(string(filter), func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("dateFilter"); got != string(filter) {
					t.Errorf("expected dateFilter=%s, got %s", filter, got)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"Errors":[],"NumberOfOrders":0,"OrderHistoryItems":[]}`))
			})

			client := newTestClient(t, handler)
			_, err := client.GetOrderHistoryByDateFilter(context.Background(), filter)
			if err != nil {
				t.Errorf("unexpected error for filter %s: %v", filter, err)
			}
		})
	}
}

// TestGetOrderHistoryByDateRangeMock tests the date range endpoint.
func TestGetOrderHistoryByDateRangeMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/orderhistory/ByDateRange" {
			t.Errorf("expected path /orderhistory/ByDateRange, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("startDate"); got != "2025-01-01" {
			t.Errorf("expected startDate=2025-01-01, got %s", got)
		}
		if got := r.URL.Query().Get("endDate"); got != "2025-02-01" {
			t.Errorf("expected endDate=2025-02-01, got %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(orderHistoryListResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.GetOrderHistoryByDateRange(context.Background(), "2025-01-01", "2025-02-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.NumberOfOrders != 2 {
		t.Errorf("expected 2 orders, got %d", resp.NumberOfOrders)
	}
}

// TestGetOrderBySalesOrderNumberMock tests the sales order number endpoint.
func TestGetOrderBySalesOrderNumberMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/orderhistory/salesOrderNumber" {
			t.Errorf("expected path /orderhistory/salesOrderNumber, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("salesOrderNumber"); got != "SO-001" {
			t.Errorf("expected salesOrderNumber=SO-001, got %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(orderDetailResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.GetOrderBySalesOrderNumber(context.Background(), "SO-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SalesOrderId != "SO-001" {
		t.Errorf("expected SalesOrderId=SO-001, got %s", resp.SalesOrderId)
	}
	if resp.OrderStatusName != "Shipped" {
		t.Errorf("expected OrderStatusName=Shipped, got %s", resp.OrderStatusName)
	}
	if resp.BillingAddress.City != "Austin" {
		t.Errorf("expected BillingAddress.City=Austin, got %s", resp.BillingAddress.City)
	}
	if len(resp.DeliveryDetail.TrackingDetails) != 1 {
		t.Fatalf("expected 1 tracking detail")
	}
	if resp.DeliveryDetail.TrackingDetails[0].Number != "1234567890" {
		t.Errorf("unexpected tracking number")
	}
	if resp.SummaryDetail.OrderTotal != 165.50 {
		t.Errorf("expected OrderTotal=165.50, got %f", resp.SummaryDetail.OrderTotal)
	}
	if len(resp.OrderLines) != 1 {
		t.Fatalf("expected 1 order line")
	}
	if resp.OrderLines[0].ProductInfo.MouserPartNumber != "TEST-001" {
		t.Errorf("expected MouserPartNumber=TEST-001")
	}
}

// TestGetOrderByWebOrderNumberMock tests the web order number endpoint.
func TestGetOrderByWebOrderNumberMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/orderhistory/webOrderNumber" {
			t.Errorf("expected path /orderhistory/webOrderNumber, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("webOrderNumber"); got != "WO-001" {
			t.Errorf("expected webOrderNumber=WO-001, got %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(orderDetailResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.GetOrderByWebOrderNumber(context.Background(), "WO-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SalesOrderId != "SO-001" {
		t.Errorf("expected SalesOrderId=SO-001, got %s", resp.SalesOrderId)
	}
}

// TestOrderHistoryErrorHandlingMock tests error handling for order history endpoints.
func TestOrderHistoryErrorHandlingMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Errors":[{"Id":1,"Code":"Unauthorized","Message":"Invalid API key"}]}`))
	})

	client := newTestClient(t, handler)
	_, err := client.GetOrderHistoryByDateFilter(context.Background(), DateFilterAll)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

// TestOrderHistoryModelRoundtrip tests JSON marshal/unmarshal for order history models.
func TestOrderHistoryModelRoundtrip(t *testing.T) {
	original := OrderHistoryItem{
		DateCreated:        "2025-01-15",
		SalesOrderNumber:   "SO-001",
		WebOrderNumber:     "WO-001",
		PoNumber:           "PO-001",
		BuyerName:          "John Doe",
		OrderStatusDisplay: "Shipped",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded OrderHistoryItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.SalesOrderNumber != original.SalesOrderNumber {
		t.Errorf("SalesOrderNumber mismatch")
	}
}

// TestOrderDetailModelRoundtrip tests JSON marshal/unmarshal for order detail models.
func TestOrderDetailModelRoundtrip(t *testing.T) {
	original := OrderDetailResponse{
		SalesOrderId:    "SO-001",
		OrderStatusName: "Shipped",
		BillingAddress:  Address{City: "Austin", StateOrProvince: "TX"},
		DeliveryDetail: Delivery{
			ShippingMethodName: "FedEx",
			TrackingDetails:    []Tracking{{Number: "123", Link: "https://track.test"}},
		},
		SummaryDetail: OrderDetailSummary{OrderTotal: 100.50},
		OrderLines: []OrderDetailLine{
			{
				Quantity:  10,
				UnitPrice: 10.05,
				ProductInfo: OrderLineProduct{
					MouserPartNumber: "TEST-001",
				},
				Activities: []OrderLineActivity{{InvoiceNumber: "INV-001", Date: "2025-01-16"}},
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded OrderDetailResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.SalesOrderId != "SO-001" {
		t.Errorf("SalesOrderId mismatch")
	}
	if decoded.BillingAddress.City != "Austin" {
		t.Errorf("BillingAddress.City mismatch")
	}
	if decoded.DeliveryDetail.TrackingDetails[0].Number != "123" {
		t.Errorf("TrackingDetails mismatch")
	}
	if decoded.SummaryDetail.OrderTotal != 100.50 {
		t.Errorf("OrderTotal mismatch")
	}
	if decoded.OrderLines[0].Activities[0].InvoiceNumber != "INV-001" {
		t.Errorf("Activities mismatch")
	}
}

// Integration tests - gated by MOUSER_API_KEY

// TestIntegrationOrderHistoryByDateFilter tests order history query by date filter.
func TestIntegrationOrderHistoryByDateFilter(t *testing.T) {
	skipIfNoKey(t)

	client, err := NewClient(testAPIKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	resp, err := client.GetOrderHistoryByDateFilter(context.Background(), DateFilterAll)
	if err != nil {
		t.Fatalf("GetOrderHistoryByDateFilter failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	// NumberOfOrders can be 0 if the account has no orders
	t.Logf("Found %d orders", resp.NumberOfOrders)
}

// TestIntegrationOrderHistoryByDateRange tests order history query by date range.
func TestIntegrationOrderHistoryByDateRange(t *testing.T) {
	skipIfNoKey(t)

	client, err := NewClient(testAPIKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	resp, err := client.GetOrderHistoryByDateRange(context.Background(), "2024-01-01", "2025-12-31")
	if err != nil {
		t.Fatalf("GetOrderHistoryByDateRange failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	t.Logf("Found %d orders in date range", resp.NumberOfOrders)
}
