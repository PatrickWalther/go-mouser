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

- **Search API** — Keyword, part number, manufacturer filtering, and list retrieval (V1 & V2)
- **Cart API** — Create, update, remove items, and manage scheduled releases
- **Order History API** — Query orders by date filter, date range, or order number
- **Order API** — Create orders, query options, get currencies and countries
- In-memory response caching with configurable TTL
- Automatic retries with exponential backoff
- Rate limiting (30 requests/minute, 1000 requests/day)
- Comprehensive error handling
- No external dependencies (stdlib only)
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

    // Search for parts by keyword
    result, err := client.KeywordSearch(context.Background(), mouser.SearchOptions{
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

// With custom options
client, err := mouser.NewClient(
    os.Getenv("MOUSER_API_KEY"),
    mouser.WithCache(&customCache),
    mouser.WithRetryConfig(mouser.RetryConfig{MaxRetries: 5}),
)
```

### Keyword Search

```go
result, err := client.KeywordSearch(ctx, mouser.SearchOptions{
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
result, err := client.PartNumberSearch(ctx, mouser.PartNumberSearchOptions{
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
result, err := client.KeywordAndManufacturerSearch(ctx,
    mouser.KeywordAndManufacturerSearchOptions{
        Keyword: "microcontroller",
        ManufacturerName: "STMicroelectronics",
        Records: 10,
    })
```

### Iterate All Results

```go
// V1 keyword search pagination
err := client.SearchAll(ctx, mouser.SearchOptions{
    Keyword: "resistor",
}, func(part mouser.Part) bool {
    fmt.Println(part.Description)
    return true  // continue iterating
})

// V2 keyword+manufacturer pagination
err := client.SearchAllByManufacturer(ctx,
    mouser.KeywordAndManufacturerSearchOptions{
        Keyword: "capacitor",
        ManufacturerName: "Murata",
    }, func(part mouser.Part) bool {
        fmt.Println(part.Description)
        return true
    })
```

### Cart Operations

```go
// Insert items into a cart
resp, err := client.InsertCartItems(ctx, mouser.CartItemRequestBody{
    CartItems: []mouser.CartItemRequest{
        {MouserPartNumber: "595-TMS320F28335PGFA", Quantity: 5},
    },
}, "US", "USD")

// Get cart contents
cart, err := client.GetCart(ctx, resp.CartKey, "US", "USD")

// Update item quantity
_, err = client.UpdateCartItems(ctx, mouser.CartItemRequestBody{
    CartKey: resp.CartKey,
    CartItems: []mouser.CartItemRequest{
        {MouserPartNumber: "595-TMS320F28335PGFA", Quantity: 10},
    },
}, "US", "USD")

// Remove an item
_, err = client.RemoveCartItem(ctx, resp.CartKey, "595-TMS320F28335PGFA", "US", "USD")
```

### Order History

```go
// Query by date filter
history, err := client.GetOrderHistoryByDateFilter(ctx, mouser.DateFilterThisMonth)

// Query by date range
history, err := client.GetOrderHistoryByDateRange(ctx, "2025-01-01", "2025-06-30")

// Get order details
detail, err := client.GetOrderBySalesOrderNumber(ctx, "12345678")
```

### Order Operations

```go
// Get available currencies and countries
currencies, err := client.GetCurrencies(ctx, "US")
countries, err := client.GetCountries(ctx, "")

// Query order options for a cart
options, err := client.QueryOrderOptions(ctx, mouser.OrderOptionsRequest{
    CartKey: cartKey,
    CurrencyCode: "USD",
})

// Create an order (use SubmitOrder: false to validate first)
order, err := client.CreateOrder(ctx, mouser.CreateOrderRequest{
    CartKey:      cartKey,
    CurrencyCode: "USD",
    SubmitOrder:  false,
})
```

## Configuration

### Environment Variables

```bash
export MOUSER_API_KEY="your-api-key"
```

### Client Options

```go
client, _ := mouser.NewClient(apiKey,
    mouser.WithHTTPClient(&http.Client{Timeout: 60 * time.Second}),
    mouser.WithRateLimiter(mouser.NewRateLimiter(30, 1000)),
    mouser.WithCacheConfig(mouser.CacheConfig{
        Enabled: true,
        SearchTTL: 5 * time.Minute,
        DetailsTTL: 10 * time.Minute,
        ManufacturersTTL: 24 * time.Hour,
        CurrenciesTTL: 24 * time.Hour,
        CountriesTTL: 24 * time.Hour,
    }),
    mouser.WithRetryConfig(mouser.RetryConfig{
        MaxRetries: 3,
        InitialBackoff: 500 * time.Millisecond,
        MaxBackoff: 30 * time.Second,
        Multiplier: 2.0,
        Jitter: 0.1,
    }),
    mouser.WithoutCache(),
    mouser.WithoutRetry(),
)
```

## Caching

The client caches responses to reduce API usage:

| Data | Default TTL | Cached? |
|------|-------------|---------|
| Search results | 5 minutes | Yes |
| Part details | 10 minutes | Yes |
| Manufacturer list | 24 hours | Yes |
| Currencies | 24 hours | Yes |
| Countries | 24 hours | Yes |
| Cart operations | — | No (mutations) |
| Order operations | — | No (mutations) |
| Order history | — | No (user-specific) |

```go
// Disable caching
client, _ := mouser.NewClient(apiKey, mouser.WithoutCache())

// Clear cache
client.ClearCache()
```

## Rate Limiting

The client enforces Mouser API rate limits:

```go
stats := client.RateLimitStats()
fmt.Printf("Minute remaining: %d\n", stats.MinuteRemaining)
fmt.Printf("Daily remaining: %d\n", stats.DayRemaining)
```

## Error Handling

```go
result, err := client.KeywordSearch(ctx, opts)
if err != nil {
    if errors.Is(err, mouser.ErrNotFound) {
        // Part not found
    }
    if errors.Is(err, mouser.ErrRateLimitExceeded) {
        // Rate limited
    }

    if mouserErr, ok := err.(*mouser.MouserError); ok {
        fmt.Printf("HTTP %d: %s\n", mouserErr.StatusCode, mouserErr.Message)
    }
}
```

## API Coverage

### Search API (5 endpoints)

| Endpoint | Method | Function |
|----------|--------|----------|
| `/search/keyword` | POST | `KeywordSearch` |
| `/search/partnumber` | POST | `PartNumberSearch` |
| `/search/keywordandmanufacturer` | POST | `KeywordAndManufacturerSearch` |
| `/search/partnumberandmanufacturer` | POST | `PartNumberAndManufacturerSearch` |
| `/search/manufacturerlist` | GET | `GetManufacturerList` |

### Cart API (8 endpoints)

| Endpoint | Method | Function |
|----------|--------|----------|
| `/cart` | GET | `GetCart` |
| `/cart` | POST | `UpdateCart` |
| `/cart/items/insert` | POST | `InsertCartItems` |
| `/cart/items/update` | POST | `UpdateCartItems` |
| `/cart/item/remove` | POST | `RemoveCartItem` |
| `/cart/insert/schedule` | POST | `InsertCartSchedule` |
| `/cart/update/schedule` | POST | `UpdateCartSchedule` |
| `/cart/deleteall/schedule` | POST | `DeleteAllCartSchedules` |

### Order History API (4 endpoints)

| Endpoint | Method | Function |
|----------|--------|----------|
| `/orderhistory/ByDateFilter` | GET | `GetOrderHistoryByDateFilter` |
| `/orderhistory/ByDateRange` | GET | `GetOrderHistoryByDateRange` |
| `/orderhistory/salesOrderNumber` | GET | `GetOrderBySalesOrderNumber` |
| `/orderhistory/webOrderNumber` | GET | `GetOrderByWebOrderNumber` |

### Order API (7 endpoints)

| Endpoint | Method | Function |
|----------|--------|----------|
| `/order/options/query` | POST | `QueryOrderOptions` |
| `/order/currencies` | GET | `GetCurrencies` |
| `/order/countries` | GET | `GetCountries` |
| `/order` | POST | `CreateOrder` |
| `/order/CreateFromOrder` | POST | `CreateOrderFromPrevious` |
| `/order/{orderNumber}` | GET | `GetOrderDetails` |
| `/order/item/CreateCartFromOrder` | POST | `CreateCartFromOrder` |

### Convenience Methods

| Function | Description |
|----------|-------------|
| `GetPartDetails` | Exact part number lookup (single part) |
| `GetPartDetailsWithManufacturer` | Part lookup with manufacturer filter |
| `SearchAll` | Paginated keyword search iterator |
| `SearchAllByManufacturer` | Paginated keyword+manufacturer iterator |

## Testing

### Setup

```bash
cp .env.example .env
# Edit .env and add your MOUSER_API_KEY
```

### Run Tests

```bash
# Run all tests (integration tests skip if MOUSER_API_KEY not set)
go test -v ./...

# Run with API key from environment
MOUSER_API_KEY=your-key go test -v ./...
```

### Test Coverage

- **195 total tests** across 10 test files
- **~170 mock tests** run without API key (httptest-based)
- **~25 integration tests** use real Mouser API endpoints (auto-skip without key)

### Test Files

| File | Type |
|------|------|
| testhelper_test.go | Mock server infrastructure + search mock tests |
| client_test.go | Client configuration unit + integration tests |
| products_test.go | Search integration tests |
| cart_test.go | Cart mock + integration tests |
| orderhistory_test.go | Order history mock + integration tests |
| order_test.go | Order mock + integration tests |
| cache_test.go | Cache unit tests |
| ratelimit_test.go | Rate limiter unit + integration tests |
| retry_test.go | Retry logic unit tests |
| errors_test.go | Error handling unit tests |

## Performance

- Caching reduces redundant API calls
- Rate limiting prevents throttling
- Exponential backoff with jitter for retries
- Thread-safe for concurrent use
- No external dependencies (faster builds)

## License

MIT License - see [LICENSE](LICENSE) for details.
