# go-mouser

[![Go Reference](https://pkg.go.dev/badge/github.com/PatrickWalther/go-mouser.svg)](https://pkg.go.dev/github.com/PatrickWalther/go-mouser)
[![Go Report Card](https://goreportcard.com/badge/github.com/PatrickWalther/go-mouser)](https://goreportcard.com/report/github.com/PatrickWalther/go-mouser)
[![Coverage](https://img.shields.io/badge/coverage-80.2%25-brightgreen)](https://github.com/PatrickWalther/go-mouser)
[![Tests](https://github.com/PatrickWalther/go-mouser/actions/workflows/test.yml/badge.svg)](https://github.com/PatrickWalther/go-mouser/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go client library for the [Mouser Electronics](https://www.mouser.com) API.

## Requirements

- **Go 1.22+**
- **Mouser API key** from [https://www.mouser.com/api-hub/](https://www.mouser.com/api-hub/)
- No external dependencies beyond stdlib

## Features

- Keyword and part number search (V1 & V2 APIs)
- Manufacturer filtering and list retrieval
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
    Records: 1,
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
err := client.SearchAll(ctx, mouser.SearchOptions{
    Keyword: "resistor",
}, func(part mouser.Part) bool {
    fmt.Println(part.Description)
    return true  // continue iterating
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

```go
// Default cache configuration
client := mouser.NewClient(apiKey)

// Disable caching
client := mouser.NewClient(apiKey, mouser.WithoutCache())

// Custom cache
client := mouser.NewClient(apiKey,
    mouser.WithCacheConfig(mouser.CacheConfig{
        SearchTTL: 10 * time.Minute,
        DetailsTTL: 20 * time.Minute,
    }),
)

// Clear cache
client.ClearCache()
```

## Rate Limiting

The client enforces Mouser API rate limits:

```go
// Check rate limit stats
stats := client.RateLimitStats()
fmt.Printf("Minute: %d/%d\n", stats.MinuteUsed, stats.MinuteLimit)
fmt.Printf("Day: %d/%d\n", stats.DayUsed, stats.DayLimit)
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

| Endpoint | Method |
|----------|--------|
| `/search/keyword` | POST |
| `/search/partnumber` | POST |
| `/search/keywordandmanufacturer` | POST |
| `/search/partnumberandmanufacturer` | POST |
| `/search/manufacturerlist` | GET |

## Testing

### Setup

```bash
cp .env.example .env
# Edit .env and add your MOUSER_API_KEY
```

### Run Tests

```bash
# Run all tests (API tests skip if MOUSER_API_KEY not set)
go test -v ./...

# Run with API key from environment
MOUSER_API_KEY=your-key go test -v ./...
```

### Test Coverage

- **131 total tests** organized by concern (cache, search, rate limiting, retry, errors)
- **12 tests** use real Mouser API endpoints
- **119 tests** run without API key
- Real API tests automatically skip if `MOUSER_API_KEY` environment variable not set

### Test Files

| File | Tests | Type |
|------|-------|------|
| client_test.go | 21 | Unit + API |
| products_test.go | 13 | Unit + API |
| cache_test.go | 19 | Unit only |
| ratelimit_test.go | 24 | Unit + API |
| retry_test.go | 20 | Unit only |
| errors_test.go | 34 | Unit + API |

## Performance

- Caching reduces redundant API calls
- Rate limiting prevents throttling
- Exponential backoff with jitter for retries
- Thread-safe for concurrent use
- No external dependencies (faster builds)

## License

MIT License - see [LICENSE](LICENSE) for details.
