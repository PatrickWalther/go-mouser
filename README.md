# go-mouser

[![Go Reference](https://pkg.go.dev/badge/github.com/PatrickWalther/go-mouser.svg)](https://pkg.go.dev/github.com/PatrickWalther/go-mouser)
[![Go Report Card](https://goreportcard.com/badge/github.com/PatrickWalther/go-mouser)](https://goreportcard.com/report/github.com/PatrickWalther/go-mouser)
[![Tests](https://github.com/PatrickWalther/go-mouser/actions/workflows/test.yml/badge.svg)](https://github.com/PatrickWalther/go-mouser/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go client library for the [Mouser Electronics](https://www.mouser.com) API with **100% endpoint coverage**.

## Requirements

- **Go 1.22+**
- **Mouser API key** from [https://www.mouser.com/api-hub/](https://www.mouser.com/api-hub/)
- No external dependencies beyond stdlib

## Features

- All 24 Mouser API endpoints across 4 API groups
- **Service-based architecture** — endpoints grouped by domain (`client.Search`, `client.Cart`, `client.OrderHistory`, `client.Order`)
- In-memory response caching with configurable TTL
- Automatic retries with exponential backoff for transient errors
- Rate limiting (30 requests/minute, 1000 requests/day)
- No external dependencies beyond stdlib
- Thread-safe for concurrent use

## Installation

```bash
go get github.com/PatrickWalther/go-mouser
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/PatrickWalther/go-mouser"
)

func main() {
    client, err := mouser.NewClient(os.Getenv("MOUSER_API_KEY"))
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Search for parts by keyword
    result, err := client.Search.KeywordSearch(ctx, mouser.SearchOptions{
        Keyword: "STM32F4",
        Records: 10,
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, part := range result.Parts {
        fmt.Printf("%s - %s\n", part.ManufacturerPartNumber, part.Description)
    }
}
```

## Usage

### Creating a Client

```go
// Basic client
client, err := mouser.NewClient(os.Getenv("MOUSER_API_KEY"))
defer client.Close()

// Client with custom options
client, err := mouser.NewClient(
    os.Getenv("MOUSER_API_KEY"),
    mouser.WithHTTPClient(&http.Client{Timeout: 60 * time.Second}),
    mouser.WithRetryConfig(mouser.RetryConfig{MaxRetries: 5}),
)
```

### Keyword Search

```go
result, err := client.Search.KeywordSearch(ctx, mouser.SearchOptions{
    Keyword: "capacitor",
    Records: 20,
    SearchOption: mouser.SearchOptionInStock,
})

for _, part := range result.Parts {
    fmt.Printf("Part: %s, Stock: %s\n", part.MouserPartNumber, part.Availability)
}
```

### Part Number Search

```go
result, err := client.Search.PartNumberSearch(ctx, mouser.PartNumberSearchOptions{
    PartNumber: "STM32F407VGT6",
    PartSearchOption: mouser.PartSearchOptionExact,
})

if len(result.Parts) > 0 {
    part := result.Parts[0]
    fmt.Printf("Price breaks: %v\n", part.PriceBreaks)
}
```

### Search with Manufacturer Filter

```go
result, err := client.Search.KeywordAndManufacturerSearch(ctx,
    mouser.KeywordAndManufacturerSearchOptions{
        Keyword:          "microcontroller",
        ManufacturerName: "STMicroelectronics",
        Records:          10,
    })
```

### Part Details

```go
details, err := client.Search.PartDetails(ctx, "STM32F407VGT6")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Part: %s\n", details.Parts[0].ManufacturerPartNumber)
fmt.Printf("Price breaks: %v\n", details.Parts[0].PriceBreaks)
```

### Manufacturer List

```go
mfrs, err := client.Search.ManufacturerList(ctx)
for _, mfr := range mfrs.ManufacturerList {
    fmt.Printf("%s (ID: %s)\n", mfr.ManufacturerName, mfr.ManufacturerID)
}
```

### Iterate All Results

```go
// V1 keyword search pagination
err := client.Search.All(ctx, mouser.SearchOptions{
    Keyword: "resistor",
}, func(part mouser.Part) bool {
    fmt.Println(part.Description)
    return true // continue iterating
})

// V2 keyword+manufacturer pagination
err := client.Search.AllByManufacturer(ctx,
    mouser.KeywordAndManufacturerSearchOptions{
        Keyword:          "capacitor",
        ManufacturerName: "Murata",
    }, func(part mouser.Part) bool {
        fmt.Println(part.Description)
        return true
    })
```

### Cart Operations

```go
// Insert items into a cart
resp, err := client.Cart.InsertItems(ctx, mouser.CartItemRequestBody{
    CartItems: []mouser.CartItemRequest{
        {MouserPartNumber: "595-TMS320F28335PGFA", Quantity: 5},
    },
}, "US", "USD")

// Get cart contents
cart, err := client.Cart.Get(ctx, resp.CartKey, "US", "USD")

// Update item quantity
_, err = client.Cart.UpdateItems(ctx, mouser.CartItemRequestBody{
    CartKey: resp.CartKey,
    CartItems: []mouser.CartItemRequest{
        {MouserPartNumber: "595-TMS320F28335PGFA", Quantity: 10},
    },
}, "US", "USD")

// Remove an item
_, err = client.Cart.RemoveItem(ctx, resp.CartKey, "595-TMS320F28335PGFA", "US", "USD")
```

### Order History

```go
// Query by date filter
history, err := client.OrderHistory.ByDateFilter(ctx, mouser.DateFilterThisMonth)

// Query by date range
history, err := client.OrderHistory.ByDateRange(ctx, "2025-01-01", "2025-06-30")

// Get order details
detail, err := client.OrderHistory.BySalesOrderNumber(ctx, "12345678")
```

### Order Operations

```go
// Get available currencies and countries
currencies, err := client.Order.Currencies(ctx, "US")
countries, err := client.Order.Countries(ctx, "")

// Query order options for a cart
options, err := client.Order.QueryOptions(ctx, mouser.OrderOptionsRequest{
    CartKey:      cartKey,
    CurrencyCode: "USD",
})

// Create an order (use SubmitOrder: false to validate first)
order, err := client.Order.Create(ctx, mouser.CreateOrderRequest{
    CartKey:      cartKey,
    CurrencyCode: "USD",
    SubmitOrder:  false,
})
```

### Rate Limit Monitoring

```go
stats := client.RateLimitStats()
fmt.Printf("Minute: %d/%d remaining\n", stats.MinuteRemaining, stats.MinuteLimit)
fmt.Printf("Day: %d/%d remaining\n", stats.DayRemaining, stats.DayLimit)
```

### Error Handling

```go
import "errors"

result, err := client.Search.KeywordSearch(ctx, opts)
if err != nil {
    if errors.Is(err, mouser.ErrRateLimitExceeded) {
        // Wait and retry
    }
    if errors.Is(err, mouser.ErrUnauthorized) {
        // Check API key
    }
    if errors.Is(err, mouser.ErrNotFound) {
        // Part not found
    }

    // Check for HTTP error details
    var mouserErr *mouser.MouserError
    if errors.As(err, &mouserErr) {
        fmt.Printf("HTTP %d: %s\n", mouserErr.StatusCode, mouserErr.Message)
    }

    // Check for rate limit error details
    var rlErr *mouser.RateLimitError
    if errors.As(err, &rlErr) {
        fmt.Printf("Rate limited (%s): resets at %v\n", rlErr.Type, rlErr.ResetAt)
    }
}
```

## API Coverage

### Search API (5 endpoints)

| Endpoint | Method | Service Call |
|----------|--------|-------------|
| `/search/keyword` | POST | `client.Search.KeywordSearch()` |
| `/search/partnumber` | POST | `client.Search.PartNumberSearch()` |
| `/search/keywordandmanufacturer` | POST | `client.Search.KeywordAndManufacturerSearch()` |
| `/search/partnumberandmanufacturer` | POST | `client.Search.PartNumberAndManufacturerSearch()` |
| `/search/manufacturerlist` | GET | `client.Search.ManufacturerList()` |

### Cart API (8 endpoints)

| Endpoint | Method | Service Call |
|----------|--------|-------------|
| `/cart` | GET | `client.Cart.Get()` |
| `/cart` | POST | `client.Cart.Update()` |
| `/cart/items/insert` | POST | `client.Cart.InsertItems()` |
| `/cart/items/update` | POST | `client.Cart.UpdateItems()` |
| `/cart/item/remove` | POST | `client.Cart.RemoveItem()` |
| `/cart/insert/schedule` | POST | `client.Cart.InsertSchedule()` |
| `/cart/update/schedule` | POST | `client.Cart.UpdateSchedule()` |
| `/cart/deleteall/schedule` | POST | `client.Cart.DeleteAllSchedules()` |

### Order History API (4 endpoints)

| Endpoint | Method | Service Call |
|----------|--------|-------------|
| `/orderhistory/ByDateFilter` | GET | `client.OrderHistory.ByDateFilter()` |
| `/orderhistory/ByDateRange` | GET | `client.OrderHistory.ByDateRange()` |
| `/orderhistory/salesOrderNumber` | GET | `client.OrderHistory.BySalesOrderNumber()` |
| `/orderhistory/webOrderNumber` | GET | `client.OrderHistory.ByWebOrderNumber()` |

### Order API (7 endpoints)

| Endpoint | Method | Service Call |
|----------|--------|-------------|
| `/order/options/query` | POST | `client.Order.QueryOptions()` |
| `/order/currencies` | GET | `client.Order.Currencies()` |
| `/order/countries` | GET | `client.Order.Countries()` |
| `/order` | POST | `client.Order.Create()` |
| `/order/CreateFromOrder` | POST | `client.Order.CreateFromPrevious()` |
| `/order/{orderNumber}` | GET | `client.Order.Details()` |
| `/order/item/CreateCartFromOrder` | POST | `client.Order.CartFromOrder()` |

### Convenience Methods

| Service Call | Description |
|-------------|-------------|
| `client.Search.PartDetails()` | Exact part number lookup (single part) |
| `client.Search.PartDetailsWithManufacturer()` | Part lookup with manufacturer filter |
| `client.Search.All()` | Paginated keyword search iterator |
| `client.Search.AllByManufacturer()` | Paginated keyword+manufacturer iterator |

**24 endpoints + 4 convenience methods**

## Configuration

### Environment Variables

| Variable | Description |
|----------|-------------|
| `MOUSER_API_KEY` | Mouser API key from [api-hub](https://www.mouser.com/api-hub/) |

### Client Options

| Option | Description |
|--------|-------------|
| `WithHTTPClient` | Custom HTTP client |
| `WithBaseURL` | Custom base URL (for testing) |
| `WithRateLimiter` | Custom rate limiter |
| `WithCache` | Custom cache implementation |
| `WithCacheConfig` | Configure cache TTLs |
| `WithoutCache` | Disable caching |
| `WithRetryConfig` | Custom retry configuration |
| `WithoutRetry` | Disable retries |

### Services

| Service | Methods |
|---------|---------|
| `client.Search` | `KeywordSearch()`, `PartNumberSearch()`, `KeywordAndManufacturerSearch()`, `PartNumberAndManufacturerSearch()`, `ManufacturerList()`, `PartDetails()`, `PartDetailsWithManufacturer()`, `All()`, `AllByManufacturer()` |
| `client.Cart` | `Get()`, `Update()`, `InsertItems()`, `UpdateItems()`, `RemoveItem()`, `InsertSchedule()`, `UpdateSchedule()`, `DeleteAllSchedules()` |
| `client.OrderHistory` | `ByDateFilter()`, `ByDateRange()`, `BySalesOrderNumber()`, `ByWebOrderNumber()` |
| `client.Order` | `QueryOptions()`, `Currencies()`, `Countries()`, `Create()`, `CreateFromPrevious()`, `Details()`, `CartFromOrder()` |

### Client Methods

| Method | Description |
|--------|-------------|
| `Close()` | Release resources (always call with `defer`) |
| `RateLimitStats()` | Get current rate limit usage |
| `ClearCache()` | Clear all cached responses |

## Caching

The client includes in-memory caching with configurable TTL:

```go
// Default: caching enabled
client, err := mouser.NewClient(apiKey)
defer client.Close()

// Custom cache configuration
client, err := mouser.NewClient(apiKey,
    mouser.WithCacheConfig(mouser.CacheConfig{
        Enabled:          true,
        SearchTTL:        5 * time.Minute,
        DetailsTTL:       10 * time.Minute,
        ManufacturersTTL: 24 * time.Hour,
        CurrenciesTTL:    24 * time.Hour,
        CountriesTTL:     24 * time.Hour,
    }),
)

// Disable caching
client, err := mouser.NewClient(apiKey, mouser.WithoutCache())

// Clear all cached data
client.ClearCache()
```

**Caching behavior by endpoint:**
- **Cached (SearchTTL):** `client.Search.KeywordSearch`, `client.Search.PartNumberSearch`, `client.Search.KeywordAndManufacturerSearch`, `client.Search.PartNumberAndManufacturerSearch`
- **Cached (DetailsTTL):** `client.Search.PartDetails`, `client.Search.PartDetailsWithManufacturer`
- **Cached (ManufacturersTTL):** `client.Search.ManufacturerList`
- **Cached (CurrenciesTTL):** `client.Order.Currencies`
- **Cached (CountriesTTL):** `client.Order.Countries`
- **Not cached:** Cart, Order, and OrderHistory endpoints (mutations and user-specific data)

## Retries

The client automatically retries failed requests with exponential backoff:

- Retries on: 429 (rate limit), 500, 502, 503, 504, network timeouts
- Does not retry: 400, 401, 403, 404
- Default: 3 retries with 500ms initial backoff, 2x multiplier

```go
// Custom retry configuration
client, err := mouser.NewClient(apiKey,
    mouser.WithRetryConfig(mouser.RetryConfig{
        MaxRetries:     5,
        InitialBackoff: time.Second,
        MaxBackoff:     time.Minute,
        Multiplier:     2.0,
        Jitter:         0.1,
    }),
)

// Disable retries
client, err := mouser.NewClient(apiKey, mouser.WithoutRetry())
```

## Rate Limits

Mouser API enforces the following rate limits:

- **Per Minute**: 30 requests
- **Per Day**: 1000 requests

The client tracks these limits locally and returns `*RateLimitError` (wrapping `ErrRateLimitExceeded` or `ErrDailyLimitExceeded`) before making requests that would exceed them. It also respects `Retry-After` headers from the server.

## Breaking Changes

All endpoint methods moved from flat `Client` methods to service-based accessors:

| Before | After |
|--------|-------|
| `client.KeywordSearch()` | `client.Search.KeywordSearch()` |
| `client.PartNumberSearch()` | `client.Search.PartNumberSearch()` |
| `client.KeywordAndManufacturerSearch()` | `client.Search.KeywordAndManufacturerSearch()` |
| `client.PartNumberAndManufacturerSearch()` | `client.Search.PartNumberAndManufacturerSearch()` |
| `client.GetManufacturerList()` | `client.Search.ManufacturerList()` |
| `client.GetPartDetails()` | `client.Search.PartDetails()` |
| `client.GetPartDetailsWithManufacturer()` | `client.Search.PartDetailsWithManufacturer()` |
| `client.SearchAll()` | `client.Search.All()` |
| `client.SearchAllByManufacturer()` | `client.Search.AllByManufacturer()` |
| `client.GetCart()` | `client.Cart.Get()` |
| `client.UpdateCart()` | `client.Cart.Update()` |
| `client.InsertCartItems()` | `client.Cart.InsertItems()` |
| `client.UpdateCartItems()` | `client.Cart.UpdateItems()` |
| `client.RemoveCartItem()` | `client.Cart.RemoveItem()` |
| `client.InsertCartSchedule()` | `client.Cart.InsertSchedule()` |
| `client.UpdateCartSchedule()` | `client.Cart.UpdateSchedule()` |
| `client.DeleteAllCartSchedules()` | `client.Cart.DeleteAllSchedules()` |
| `client.GetOrderHistoryByDateFilter()` | `client.OrderHistory.ByDateFilter()` |
| `client.GetOrderHistoryByDateRange()` | `client.OrderHistory.ByDateRange()` |
| `client.GetOrderBySalesOrderNumber()` | `client.OrderHistory.BySalesOrderNumber()` |
| `client.GetOrderByWebOrderNumber()` | `client.OrderHistory.ByWebOrderNumber()` |
| `client.QueryOrderOptions()` | `client.Order.QueryOptions()` |
| `client.GetCurrencies()` | `client.Order.Currencies()` |
| `client.GetCountries()` | `client.Order.Countries()` |
| `client.CreateOrder()` | `client.Order.Create()` |
| `client.CreateOrderFromPrevious()` | `client.Order.CreateFromPrevious()` |
| `client.GetOrderDetails()` | `client.Order.Details()` |
| `client.CreateCartFromOrder()` | `client.Order.CartFromOrder()` |

`RateLimitStats` fields renamed: `DailyRemaining` → `DayRemaining`. New fields added: `MinuteLimit`, `MinuteUsed`, `MinuteResetAt`, `DayLimit`, `DayUsed`, `DayResetAt`.

## Testing

### Unit Tests (Fast, No API Calls)

Run fast unit tests that don't require API credentials:

```bash
# Run all unit tests (integration tests auto-skip without key)
go test -v -short ./...

# Run only mock server tests
go test -v -short -run "Mock|Unit|TestNew|TestWith" ./...
```

### Integration Tests (Real API Calls)

Run against real Mouser API with actual credentials:

#### 1. Setup Credentials

Create a `.env` file (copy from `.env.example`):

```bash
cp .env.example .env
# Edit .env and add your MOUSER_API_KEY
```

#### 2. Run Integration Tests Locally

```bash
# Load credentials from .env and run all tests
MOUSER_API_KEY=your-key go test -v ./...
```

#### 3. GitHub Actions Integration Tests

Set up GitHub repository secrets:

1. Go to: **Settings** > **Secrets and variables** > **Actions** > **New repository secret**
2. Add secret: `MOUSER_API_KEY`

Integration tests run automatically on push to main branch.

## License

MIT License - see [LICENSE](LICENSE) for details.
