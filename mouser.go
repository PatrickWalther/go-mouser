// Package mouser provides a Go client for the Mouser Electronics API.
//
// The Mouser API allows you to search for electronic components, retrieve
// product information, and access pricing and availability data.
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
//   - KeywordSearch, PartNumberSearch: V1-compatible endpoints
//   - KeywordAndManufacturerSearch, PartNumberAndManufacturerSearch: V2 endpoints
//   - GetManufacturerList: V2 endpoint
//
// # Example Usage
//
//	client, err := mouser.NewClient("your-api-key")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Search by keyword
//	result, err := client.KeywordSearch(ctx, mouser.SearchOptions{
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
//	part, err := client.GetPartDetails(ctx, "STM32F407VGT6")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Price breaks: %v\n", part.PriceBreaks)
//
//	// Iterate all search results
//	err = client.SearchAll(ctx, mouser.SearchOptions{Keyword: "capacitor"}, func(part mouser.Part) bool {
//	    fmt.Println(part.ManufacturerPartNumber)
//	    return true // continue iterating
//	})
package mouser
