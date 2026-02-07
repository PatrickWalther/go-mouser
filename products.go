package mouser

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	// MaxRecords is the maximum number of records per search request.
	MaxRecords = 50
)

// KeywordSearch searches for parts by keyword.
// This uses the V1-compatible endpoint for broad keyword searches.
func (c *Client) KeywordSearch(ctx context.Context, opts SearchOptions) (*SearchResult, error) {
	// Validate and set defaults
	if opts.Records <= 0 {
		opts.Records = 10
	}
	if opts.Records > MaxRecords {
		opts.Records = MaxRecords
	}

	req := keywordSearchRequest{
		SearchByKeywordRequest: searchByKeywordRequest{
			Keyword:                      opts.Keyword,
			Records:                      opts.Records,
			StartingRecord:               opts.StartingRecord,
			SearchOptions:                string(opts.SearchOption),
			SearchWithYourSignUpLanguage: opts.SearchWithYourSignUpLanguage,
		},
	}

	// Check cache
	cacheKey := cacheKeyForSearch("keyword", req)
	if cached, ok := c.getCached(cacheKey); ok {
		var result SearchResult
		if err := json.Unmarshal(cached, &result); err == nil {
			return &result, nil
		}
	}

	var resp searchResponse
	path := "/search/keyword"
	if err := c.doRequest(ctx, "POST", path, req, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	// Cache the result
	if data, err := json.Marshal(resp.SearchResults); err == nil {
		c.setCache(cacheKey, data, c.cacheConfig.SearchTTL)
	}

	return &resp.SearchResults, nil
}

// PartNumberSearch searches for parts by part number.
// This uses the V1-compatible endpoint. For manufacturer-specific search, use PartNumberAndManufacturerSearch.
func (c *Client) PartNumberSearch(ctx context.Context, opts PartNumberSearchOptions) (*SearchResult, error) {
	// Validate and set defaults
	if opts.Records <= 0 {
		opts.Records = 10
	}
	if opts.Records > MaxRecords {
		opts.Records = MaxRecords
	}

	req := partNumberSearchRequest{
		SearchByPartRequest: searchByPartRequest{
			MouserPartNumber:             opts.PartNumber,
			Records:                      opts.Records,
			StartingRecord:               opts.StartingRecord,
			PartSearchOptions:            string(opts.PartSearchOption),
			SearchWithYourSignUpLanguage: opts.SearchWithYourSignUpLanguage,
		},
	}

	// Check cache
	cacheKey := cacheKeyForSearch("partnumber", req)
	if cached, ok := c.getCached(cacheKey); ok {
		var result SearchResult
		if err := json.Unmarshal(cached, &result); err == nil {
			return &result, nil
		}
	}

	var resp searchResponse
	path := "/search/partnumber"
	if err := c.doRequest(ctx, "POST", path, req, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	// Cache the result
	if data, err := json.Marshal(resp.SearchResults); err == nil {
		c.setCache(cacheKey, data, c.cacheConfig.SearchTTL)
	}

	return &resp.SearchResults, nil
}

// KeywordAndManufacturerSearch searches for parts by keyword and manufacturer.
// This is a V2 endpoint that supports pagination via PageNumber.
func (c *Client) KeywordAndManufacturerSearch(ctx context.Context, opts KeywordAndManufacturerSearchOptions) (*SearchResult, error) {
	// Validate and set defaults
	if opts.Records <= 0 {
		opts.Records = 10
	}
	if opts.Records > MaxRecords {
		opts.Records = MaxRecords
	}
	if opts.PageNumber <= 0 {
		opts.PageNumber = 1
	}

	req := keywordAndManufacturerSearchRequest{
		SearchByKeywordMfrNameRequest: searchByKeywordMfrNameRequest{
			Keyword:                      opts.Keyword,
			ManufacturerName:             opts.ManufacturerName,
			Records:                      opts.Records,
			PageNumber:                   opts.PageNumber,
			SearchOptions:                string(opts.SearchOption),
			SearchWithYourSignUpLanguage: opts.SearchWithYourSignUpLanguage,
		},
	}

	// Check cache
	cacheKey := cacheKeyForSearch("keyword+mfr", req)
	if cached, ok := c.getCached(cacheKey); ok {
		var result SearchResult
		if err := json.Unmarshal(cached, &result); err == nil {
			return &result, nil
		}
	}

	var resp searchResponse
	path := "/search/keywordandmanufacturer"
	if err := c.doRequest(ctx, "POST", path, req, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	// Cache the result
	if data, err := json.Marshal(resp.SearchResults); err == nil {
		c.setCache(cacheKey, data, c.cacheConfig.SearchTTL)
	}

	return &resp.SearchResults, nil
}

// PartNumberAndManufacturerSearch searches for parts by part number and manufacturer.
// This is a V2 endpoint that provides more precise matching.
func (c *Client) PartNumberAndManufacturerSearch(ctx context.Context, opts PartNumberAndManufacturerSearchOptions) (*SearchResult, error) {
	req := partNumberAndManufacturerSearchRequest{
		SearchByPartMfrNameRequest: searchByPartMfrNameRequest{
			MouserPartNumber:  opts.PartNumber,
			ManufacturerName:  opts.ManufacturerName,
			PartSearchOptions: string(opts.PartSearchOption),
		},
	}

	// Check cache
	cacheKey := cacheKeyForSearch("partnumber+mfr", req)
	if cached, ok := c.getCached(cacheKey); ok {
		var result SearchResult
		if err := json.Unmarshal(cached, &result); err == nil {
			return &result, nil
		}
	}

	var resp searchResponse
	path := "/search/partnumberandmanufacturer"
	if err := c.doRequest(ctx, "POST", path, req, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	// Cache the result
	if data, err := json.Marshal(resp.SearchResults); err == nil {
		c.setCache(cacheKey, data, c.cacheConfig.SearchTTL)
	}

	return &resp.SearchResults, nil
}

// GetManufacturerList returns the list of all manufacturers in the Mouser catalog.
// This result is heavily cached as it rarely changes.
func (c *Client) GetManufacturerList(ctx context.Context) (*ManufacturerListResult, error) {
	// Check cache first (manufacturer list is mostly static)
	cacheKey := cacheKeyForManufacturers()
	if cached, ok := c.getCached(cacheKey); ok {
		var result ManufacturerListResult
		if err := json.Unmarshal(cached, &result); err == nil {
			return &result, nil
		}
	}

	var resp manufacturerListResponse
	path := "/search/manufacturerlist"
	if err := c.doRequest(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	// Cache the result with longer TTL
	if data, err := json.Marshal(resp.MouserManufacturerList); err == nil {
		c.setCache(cacheKey, data, c.cacheConfig.ManufacturersTTL)
	}

	return &resp.MouserManufacturerList, nil
}

// GetPartDetails retrieves detailed information for a specific part.
// This is a convenience method that uses PartNumberSearch with Records=1.
func (c *Client) GetPartDetails(ctx context.Context, partNumber string) (*Part, error) {
	// Check cache
	cacheKey := cacheKeyForDetails(partNumber)
	if cached, ok := c.getCached(cacheKey); ok {
		var result Part
		if err := json.Unmarshal(cached, &result); err == nil {
			return &result, nil
		}
	}

	result, err := c.PartNumberSearch(ctx, PartNumberSearchOptions{
		PartNumber:       partNumber,
		Records:          1,
		PartSearchOption: PartSearchOptionExact,
	})
	if err != nil {
		return nil, err
	}

	if len(result.Parts) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, partNumber)
	}

	part := result.Parts[0]

	// Cache the result
	if data, err := json.Marshal(part); err == nil {
		c.setCache(cacheKey, data, c.cacheConfig.DetailsTTL)
	}

	return &part, nil
}

// GetPartDetailsWithManufacturer retrieves detailed information for a specific part from a specific manufacturer.
// This provides more precise matching than GetPartDetails.
func (c *Client) GetPartDetailsWithManufacturer(ctx context.Context, partNumber, manufacturerName string) (*Part, error) {
	// Check cache
	cacheKey := cacheKeyForDetails(manufacturerName + ":" + partNumber)
	if cached, ok := c.getCached(cacheKey); ok {
		var result Part
		if err := json.Unmarshal(cached, &result); err == nil {
			return &result, nil
		}
	}

	result, err := c.PartNumberAndManufacturerSearch(ctx, PartNumberAndManufacturerSearchOptions{
		PartNumber:       partNumber,
		ManufacturerName: manufacturerName,
		PartSearchOption: PartSearchOptionExact,
	})
	if err != nil {
		return nil, err
	}

	if len(result.Parts) == 0 {
		return nil, fmt.Errorf("%w: %s (%s)", ErrNotFound, partNumber, manufacturerName)
	}

	part := result.Parts[0]

	// Cache the result
	if data, err := json.Marshal(part); err == nil {
		c.setCache(cacheKey, data, c.cacheConfig.DetailsTTL)
	}

	return &part, nil
}

// SearchAll iterates through all pages of search results, calling the callback for each part.
// The callback should return true to continue iterating, or false to stop.
// This is useful for processing large result sets without manually managing pagination.
func (c *Client) SearchAll(ctx context.Context, opts SearchOptions, callback func(Part) bool) error {
	opts.Records = MaxRecords
	opts.StartingRecord = 0

	for {
		result, err := c.KeywordSearch(ctx, opts)
		if err != nil {
			return err
		}

		for _, part := range result.Parts {
			if !callback(part) {
				return nil
			}
		}

		// Check if we've retrieved all results
		if len(result.Parts) < MaxRecords || opts.StartingRecord+len(result.Parts) >= result.NumberOfResult {
			break
		}

		opts.StartingRecord += len(result.Parts)
	}

	return nil
}
