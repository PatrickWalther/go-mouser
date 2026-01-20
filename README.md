# go-mouser

A Go client library for the [Mouser Electronics API](https://www.mouser.com/api-hub/).

## Features

- **Full API Coverage**: Keyword search, part number search, manufacturer-filtered search, manufacturer list
- **Automatic Rate Limiting**: Respects Mouser's 30/min and 1000/day limits
- **Caching**: Built-in memory cache with configurable TTLs to reduce API quota usage
- **Retry with Backoff**: Automatic retries for transient failures with exponential backoff and jitter
- **Retry-After Support**: Honors server-indicated backoff periods

## Installation

```bash
go get github.com/PatrickWalther/go-mouser
```

## Usage

### Basic Search

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/PatrickWalther/go-mouser"
)

func main() {
    client, err := mouser.NewClient("your-api-key")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    result, err := client.KeywordSearch(ctx, mouser.SearchOptions{
        Keyword: "STM32F4",
        Records: 10,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d results\n", result.NumberOfResult)
    for _, part := range result.Parts {
        fmt.Printf("%s - %s - %s\n",
            part.ManufacturerPartNumber,
            part.Manufacturer,
            part.Description,
        )
    }
}
```

### Part Number Search

```go
result, err := client.PartNumberSearch(ctx, mouser.PartNumberSearchOptions{
    PartNumber: "STM32F407VGT6",
    Records:    10,
})
```

### Get Part Details

```go
part, err := client.GetPartDetails(ctx, "STM32F407VGT6")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Manufacturer: %s\n", part.Manufacturer)
fmt.Printf("Description: %s\n", part.Description)
fmt.Printf("Availability: %s\n", part.Availability)
for _, pb := range part.PriceBreaks {
    fmt.Printf("  %d+: %s %s\n", pb.Quantity, pb.Price, pb.Currency)
}
```

### Keyword and Manufacturer Search (V2)

```go
result, err := client.KeywordAndManufacturerSearch(ctx, mouser.KeywordAndManufacturerSearchOptions{
    Keyword:          "microcontroller",
    ManufacturerName: "STMicroelectronics",
    Records:          10,
})
```

### Get Manufacturer List

```go
manufacturers, err := client.GetManufacturerList(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Found %d manufacturers\n", manufacturers.Count)
```

### Iterate All Search Results

```go
err = client.SearchAll(ctx, mouser.SearchOptions{Keyword: "capacitor 10uF"}, func(part mouser.Part) bool {
    fmt.Println(part.ManufacturerPartNumber)
    return true // return false to stop iteration
})
```

### Custom Configuration

```go
// Custom HTTP client
httpClient := &http.Client{
    Timeout: 60 * time.Second,
}

client, err := mouser.NewClient("your-api-key",
    mouser.WithHTTPClient(httpClient),
)

// Custom rate limiter
rateLimiter := mouser.NewRateLimiter(20, 500) // 20/min, 500/day
client, err := mouser.NewClient("your-api-key",
    mouser.WithRateLimiter(rateLimiter),
)

// Custom retry configuration
client, err := mouser.NewClient("your-api-key",
    mouser.WithRetryConfig(mouser.RetryConfig{
        MaxRetries:     5,
        InitialBackoff: time.Second,
        MaxBackoff:     time.Minute,
        Multiplier:     2.0,
        Jitter:         0.1,
    }),
)

// Custom cache configuration
client, err := mouser.NewClient("your-api-key",
    mouser.WithCacheConfig(mouser.CacheConfig{
        Enabled:          true,
        SearchTTL:        10 * time.Minute,
        DetailsTTL:       30 * time.Minute,
        ManufacturersTTL: 48 * time.Hour,
    }),
)

// Disable caching
client, err := mouser.NewClient("your-api-key",
    mouser.WithoutCache(),
)

// Disable retries
client, err := mouser.NewClient("your-api-key",
    mouser.WithoutRetry(),
)
```

## Rate Limiting

The Mouser API has the following rate limits:
- 30 requests per minute
- 1000 requests per day

This client automatically handles rate limiting and will wait when limits are approached. You can check remaining capacity:

```go
stats := client.RateLimitStats()
fmt.Printf("Minute remaining: %d\n", stats.MinuteRemaining)
fmt.Printf("Daily remaining: %d\n", stats.DailyRemaining)
```

## Caching

By default, the client caches responses to reduce API quota usage:

| Data Type | Default TTL |
|-----------|-------------|
| Search results | 5 minutes |
| Part details | 10 minutes |
| Manufacturer list | 24 hours |

Clear the cache manually:

```go
client.ClearCache()
```

## Error Handling

```go
result, err := client.KeywordSearch(ctx, opts)
if err != nil {
    if errors.Is(err, mouser.ErrRateLimitExceeded) {
        // Handle rate limiting
    }
    if errors.Is(err, mouser.ErrNotFound) {
        // Handle not found
    }
    if mouserErr, ok := err.(*mouser.MouserError); ok {
        fmt.Printf("Status: %d, Endpoint: %s\n", mouserErr.StatusCode, mouserErr.Endpoint)
    }
}
```

## API Key

Get your API key from the [Mouser API Hub](https://www.mouser.com/api-hub/).

## License

MIT License - see [LICENSE](LICENSE) for details.
