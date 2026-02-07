package mouser

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func cartSuccessResponse() string {
	return `{
		"Errors": [],
		"CartKey": "abc-123",
		"CurrencyCode": "USD",
		"CartItems": [{
			"MouserPartNumber": "TEST-001",
			"MfrPartNumber": "MFR-001",
			"Description": "Test Part",
			"Quantity": 10,
			"UnitPrice": 1.50,
			"ExtendedPrice": 15.00,
			"LifeCycle": "Active",
			"Manufacturer": "TestMfr",
			"MouserATS": "100",
			"PartsPerReel": 5000,
			"SalesMultipleQty": "1",
			"SalesMinimumOrderQty": "1",
			"SalesMaximumOrderQty": "10000",
			"AdditionalFees": [{"Amount": 0.10, "ExtendedAmount": 1.00, "Code": "ENV"}]
		}],
		"TotalItemCount": 1,
		"AdditionalFeesTotal": 1.00,
		"MerchandiseTotal": 15.00
	}`
}

// TestGetCartMock tests GetCart with a mock server.
func TestGetCartMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/cart" {
			t.Errorf("expected path /cart, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("cartKey"); got != "abc-123" {
			t.Errorf("expected cartKey=abc-123, got %s", got)
		}
		if got := r.URL.Query().Get("countryCode"); got != "US" {
			t.Errorf("expected countryCode=US, got %s", got)
		}
		if got := r.URL.Query().Get("currencyCode"); got != "USD" {
			t.Errorf("expected currencyCode=USD, got %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cartSuccessResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.Cart.Get(context.Background(), "abc-123", "US", "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CartKey != "abc-123" {
		t.Errorf("expected CartKey=abc-123, got %s", resp.CartKey)
	}
	if resp.TotalItemCount != 1 {
		t.Errorf("expected TotalItemCount=1, got %d", resp.TotalItemCount)
	}
	if len(resp.CartItems) != 1 {
		t.Fatalf("expected 1 cart item, got %d", len(resp.CartItems))
	}
	item := resp.CartItems[0]
	if item.MouserPartNumber != "TEST-001" {
		t.Errorf("expected MouserPartNumber=TEST-001, got %s", item.MouserPartNumber)
	}
	if item.UnitPrice != 1.50 {
		t.Errorf("expected UnitPrice=1.50, got %f", item.UnitPrice)
	}
	if len(item.AdditionalFees) != 1 || item.AdditionalFees[0].Code != "ENV" {
		t.Errorf("unexpected AdditionalFees: %+v", item.AdditionalFees)
	}
}

// TestUpdateCartMock tests UpdateCart with a mock server.
func TestUpdateCartMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/cart" {
			t.Errorf("expected path /cart, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req CartItemRequestBody
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request: %v", err)
		}
		if req.CartKey != "abc-123" {
			t.Errorf("expected CartKey=abc-123, got %s", req.CartKey)
		}
		if len(req.CartItems) != 1 {
			t.Errorf("expected 1 cart item, got %d", len(req.CartItems))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cartSuccessResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.Cart.Update(context.Background(), CartItemRequestBody{
		CartKey: "abc-123",
		CartItems: []CartItemRequest{
			{MouserPartNumber: "TEST-001", Quantity: 10},
		},
	}, "US", "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CartKey != "abc-123" {
		t.Errorf("expected CartKey=abc-123, got %s", resp.CartKey)
	}
}

// TestInsertCartItemsMock tests InsertCartItems with a mock server.
func TestInsertCartItemsMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/cart/items/insert" {
			t.Errorf("expected path /cart/items/insert, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cartSuccessResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.Cart.InsertItems(context.Background(), CartItemRequestBody{
		CartItems: []CartItemRequest{
			{MouserPartNumber: "TEST-001", Quantity: 5},
		},
	}, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CartKey != "abc-123" {
		t.Errorf("expected CartKey=abc-123, got %s", resp.CartKey)
	}
}

// TestUpdateCartItemsMock tests UpdateCartItems with a mock server.
func TestUpdateCartItemsMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/cart/items/update" {
			t.Errorf("expected path /cart/items/update, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cartSuccessResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.Cart.UpdateItems(context.Background(), CartItemRequestBody{
		CartKey: "abc-123",
		CartItems: []CartItemRequest{
			{MouserPartNumber: "TEST-001", Quantity: 20},
		},
	}, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CartKey != "abc-123" {
		t.Errorf("expected CartKey=abc-123, got %s", resp.CartKey)
	}
}

// TestRemoveCartItemMock tests RemoveCartItem with a mock server.
func TestRemoveCartItemMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/cart/item/remove" {
			t.Errorf("expected path /cart/item/remove, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("cartKey"); got != "abc-123" {
			t.Errorf("expected cartKey=abc-123, got %s", got)
		}
		if got := r.URL.Query().Get("mouserPartNumber"); got != "TEST-001" {
			t.Errorf("expected mouserPartNumber=TEST-001, got %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Errors":[],"CartKey":"abc-123","CartItems":[],"TotalItemCount":0}`))
	})

	client := newTestClient(t, handler)
	resp, err := client.Cart.RemoveItem(context.Background(), "abc-123", "TEST-001", "US", "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalItemCount != 0 {
		t.Errorf("expected TotalItemCount=0 after removal, got %d", resp.TotalItemCount)
	}
}

// TestInsertCartScheduleMock tests InsertCartSchedule with a mock server.
func TestInsertCartScheduleMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/cart/insert/schedule" {
			t.Errorf("expected path /cart/insert/schedule, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req ScheduleCartItemsRequestBody
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request: %v", err)
		}
		if req.CartKey != "abc-123" {
			t.Errorf("expected CartKey=abc-123, got %s", req.CartKey)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cartSuccessResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.Cart.InsertSchedule(context.Background(), ScheduleCartItemsRequestBody{
		CartKey: "abc-123",
		ScheduleCartItems: []ScheduleReleaseRequest{
			{
				MouserPartNumber: "TEST-001",
				ScheduledReleases: []ScheduleRelease{
					{Key: "2025-06-01", Value: 5},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CartKey != "abc-123" {
		t.Errorf("expected CartKey=abc-123, got %s", resp.CartKey)
	}
}

// TestUpdateCartScheduleMock tests UpdateCartSchedule with a mock server.
func TestUpdateCartScheduleMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/cart/update/schedule" {
			t.Errorf("expected path /cart/update/schedule, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cartSuccessResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.Cart.UpdateSchedule(context.Background(), ScheduleCartItemsRequestBody{
		CartKey: "abc-123",
		ScheduleCartItems: []ScheduleReleaseRequest{
			{
				MouserPartNumber: "TEST-001",
				ScheduledReleases: []ScheduleRelease{
					{Key: "2025-06-01", Value: 10},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CartKey != "abc-123" {
		t.Errorf("expected CartKey=abc-123, got %s", resp.CartKey)
	}
}

// TestDeleteAllCartSchedulesMock tests DeleteAllCartSchedules with a mock server.
func TestDeleteAllCartSchedulesMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/cart/deleteall/schedule" {
			t.Errorf("expected path /cart/deleteall/schedule, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("cartKey"); got != "abc-123" {
			t.Errorf("expected cartKey=abc-123, got %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cartSuccessResponse()))
	})

	client := newTestClient(t, handler)
	resp, err := client.Cart.DeleteAllSchedules(context.Background(), "abc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CartKey != "abc-123" {
		t.Errorf("expected CartKey=abc-123, got %s", resp.CartKey)
	}
}

// TestCartErrorHandlingMock tests that cart endpoints properly return API errors.
func TestCartErrorHandlingMock(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Errors":[{"Id":1,"Code":"InvalidCartKey","Message":"Cart not found"}]}`))
	})

	client := newTestClient(t, handler)
	_, err := client.Cart.Get(context.Background(), "bad-key", "", "")
	if err == nil {
		t.Fatal("expected error for invalid cart key")
	}
}

// TestCartNotCachedMock tests that cart responses are not cached.
func TestCartNotCachedMock(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cartSuccessResponse()))
	})

	client := newTestClientCached(t, handler)

	// Make two identical requests
	_, _ = client.Cart.Get(context.Background(), "abc-123", "US", "USD")
	_, _ = client.Cart.Get(context.Background(), "abc-123", "US", "USD")

	// Both should hit the server (no caching for cart)
	if callCount != 2 {
		t.Errorf("expected 2 server calls (no caching), got %d", callCount)
	}
}

// TestCartModelRoundtrip tests JSON marshal/unmarshal for cart models.
func TestCartModelRoundtrip(t *testing.T) {
	original := CartItemRequestBody{
		CartKey:                    "key-123",
		MouserPaysCustomsAndDuties: true,
		CartItems: []CartItemRequest{
			{
				MouserPartNumber:   "TEST-001",
				Quantity:           10,
				CustomerPartNumber: "CUST-001",
				PackagingChoice:    PackagingChoiceCutTape,
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CartItemRequestBody
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.CartKey != original.CartKey {
		t.Errorf("CartKey mismatch: %s vs %s", decoded.CartKey, original.CartKey)
	}
	if len(decoded.CartItems) != 1 || decoded.CartItems[0].PackagingChoice != PackagingChoiceCutTape {
		t.Errorf("CartItems mismatch: %+v", decoded.CartItems)
	}
}

// TestScheduleCartModelRoundtrip tests JSON marshal/unmarshal for schedule cart models.
func TestScheduleCartModelRoundtrip(t *testing.T) {
	original := ScheduleCartItemsRequestBody{
		CartKey: "key-123",
		ScheduleCartItems: []ScheduleReleaseRequest{
			{
				MouserPartNumber: "TEST-001",
				ScheduledReleases: []ScheduleRelease{
					{Key: "2025-06-01", Value: 100},
					{Key: "2025-07-01", Value: 200},
				},
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ScheduleCartItemsRequestBody
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.CartKey != original.CartKey {
		t.Errorf("CartKey mismatch")
	}
	if len(decoded.ScheduleCartItems) != 1 {
		t.Fatalf("expected 1 schedule item")
	}
	if len(decoded.ScheduleCartItems[0].ScheduledReleases) != 2 {
		t.Errorf("expected 2 scheduled releases")
	}
}

// Integration tests - gated by MOUSER_API_KEY

// TestIntegrationCartInsertAndGet tests inserting items into a cart and retrieving the cart.
func TestIntegrationCartInsertAndGet(t *testing.T) {
	skipIfNoKey(t)

	client, err := NewClient(testAPIKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Insert items
	insertResp, err := client.Cart.InsertItems(ctx, CartItemRequestBody{
		CartItems: []CartItemRequest{
			{MouserPartNumber: "595-TMS320F28335PGFA", Quantity: 1},
		},
	}, "", "")
	if err != nil {
		t.Fatalf("InsertCartItems failed: %v", err)
	}
	if insertResp.CartKey == "" {
		t.Fatal("expected non-empty CartKey")
	}

	// Get the cart
	getResp, err := client.Cart.Get(ctx, insertResp.CartKey, "", "")
	if err != nil {
		t.Fatalf("GetCart failed: %v", err)
	}
	if getResp.CartKey != insertResp.CartKey {
		t.Errorf("CartKey mismatch: %s vs %s", getResp.CartKey, insertResp.CartKey)
	}
}

// TestIntegrationCartUpdateItems tests updating cart item quantities.
func TestIntegrationCartUpdateItems(t *testing.T) {
	skipIfNoKey(t)

	client, err := NewClient(testAPIKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Insert initial item
	insertResp, err := client.Cart.InsertItems(ctx, CartItemRequestBody{
		CartItems: []CartItemRequest{
			{MouserPartNumber: "595-TMS320F28335PGFA", Quantity: 1},
		},
	}, "", "")
	if err != nil {
		t.Fatalf("InsertCartItems failed: %v", err)
	}

	// Update quantity
	_, err = client.Cart.UpdateItems(ctx, CartItemRequestBody{
		CartKey: insertResp.CartKey,
		CartItems: []CartItemRequest{
			{MouserPartNumber: "595-TMS320F28335PGFA", Quantity: 5},
		},
	}, "", "")
	if err != nil {
		t.Fatalf("UpdateCartItems failed: %v", err)
	}
}

// TestIntegrationCartRemoveItem tests removing an item from a cart.
func TestIntegrationCartRemoveItem(t *testing.T) {
	skipIfNoKey(t)

	client, err := NewClient(testAPIKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Insert item
	insertResp, err := client.Cart.InsertItems(ctx, CartItemRequestBody{
		CartItems: []CartItemRequest{
			{MouserPartNumber: "595-TMS320F28335PGFA", Quantity: 1},
		},
	}, "", "")
	if err != nil {
		t.Fatalf("InsertCartItems failed: %v", err)
	}

	// Remove the item
	_, err = client.Cart.RemoveItem(ctx, insertResp.CartKey, "595-TMS320F28335PGFA", "", "")
	if err != nil {
		t.Fatalf("RemoveCartItem failed: %v", err)
	}
}
