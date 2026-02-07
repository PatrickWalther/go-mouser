// Package mouser provides a Go client for the Mouser Electronics API
// with 100% endpoint coverage.
//
// The Mouser API allows you to search for electronic components, manage
// shopping carts, place orders, and query order history.
//
// # API Coverage
//
// This client covers all four Mouser API groups:
//   - Search API (5 endpoints): keyword search, part number search, manufacturer filtering, manufacturer list
//   - Cart API (8 endpoints): create, update, remove items, manage scheduled releases
//   - Order History API (4 endpoints): query by date filter, date range, or order number
//   - Order API (7 endpoints): create orders, query options, get currencies and countries
//
// # Authentication
//
// The Mouser API requires an API key, which can be obtained from the
// Mouser API portal at https://www.mouser.com/api-hub/
//
// # Rate Limiting
//
// The API has the following rate limits:
//   - 30 requests per minute
//   - 1000 requests per day
//
// This client automatically handles rate limiting and will wait when
// limits are approached. It also respects Retry-After headers from the server.
//
// # Caching
//
// By default, the client caches API responses to reduce quota usage:
//   - Search results: 5 minutes
//   - Part details: 10 minutes
//   - Manufacturer list: 24 hours
//   - Currencies and countries: 24 hours
//   - Cart, order, and order history operations are never cached
//
// Caching can be disabled with WithoutCache() or customized with WithCacheConfig().
//
// # Retries
//
// The client automatically retries failed requests with exponential backoff
// for transient errors (network timeouts, 429, 5xx status codes).
// Retries can be disabled with WithoutRetry() or customized with WithRetryConfig().
//
// # API Version
//
// This client supports both V1 and V2 Mouser API endpoints:
//   - Search.KeywordSearch, Search.PartNumberSearch: V1-compatible endpoints
//   - Search.KeywordAndManufacturerSearch, Search.PartNumberAndManufacturerSearch: V2 endpoints
//   - Search.ManufacturerList: V2 endpoint
//   - Cart, OrderHistory, and Order endpoints: V1 endpoints
//
// # Example Usage
//
//	client, err := mouser.NewClient("your-api-key")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Search by keyword
//	result, err := client.Search.KeywordSearch(ctx, mouser.SearchOptions{
//	    Keyword: "STM32F4",
//	    Records: 10,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, part := range result.Parts {
//	    fmt.Printf("%s - %s\n", part.ManufacturerPartNumber, part.Description)
//	}
//
//	// Get part details
//	part, err := client.Search.PartDetails(ctx, "STM32F407VGT6")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Price breaks: %v\n", part.PriceBreaks)
//
//	// Iterate all search results
//	err = client.Search.All(ctx, mouser.SearchOptions{Keyword: "capacitor"}, func(part mouser.Part) bool {
//	    fmt.Println(part.ManufacturerPartNumber)
//	    return true // continue iterating
//	})
package mouser
