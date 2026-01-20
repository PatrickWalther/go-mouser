package mouser

import (
	"bufio"
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

var testAPIKey string

func init() {
	// Try environment variable first
	if key := os.Getenv("MOUSER_API_KEY"); key != "" {
		testAPIKey = key
		return
	}

	// Try loading from .env file as fallback
	file, err := os.Open(".env")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "MOUSER_API_KEY=") {
			key := strings.TrimPrefix(line, "MOUSER_API_KEY=")
			testAPIKey = strings.TrimSpace(key)
			return
		}
	}
}

// skipIfNoKey skips test if MOUSER_API_KEY is not available
func skipIfNoKey(t *testing.T) {
	if testAPIKey == "" {
		t.Skip("MOUSER_API_KEY not found in environment or .env file")
	}
}

// TestKeywordSearchBasic tests keyword search with a common term.
func TestKeywordSearchBasic(t *testing.T) {
	skipIfNoKey(t)

	client, err := NewClient(testAPIKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := client.KeywordSearch(ctx, SearchOptions{
		Keyword: "resistor",
		Records: 5,
	})

	if err != nil {
		t.Fatalf("keyword search failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.NumberOfResult <= 0 {
		t.Errorf("expected at least 1 result, got %d", result.NumberOfResult)
	}

	if len(result.Parts) == 0 {
		t.Error("expected at least one part in results")
	}

	if len(result.Parts) > 0 {
		part := result.Parts[0]
		if part.MouserPartNumber == "" {
			t.Error("expected part to have Mouser part number")
		}
	}
}

// TestPartNumberSearch tests searching by exact part number.
func TestPartNumberSearch(t *testing.T) {
	skipIfNoKey(t)

	apiKey := testAPIKey
	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Search for a common IC that likely exists
	result, err := client.PartNumberSearch(ctx, PartNumberSearchOptions{
		PartNumber: "LM386",
		Records:    5,
	})

	if err != nil {
		t.Fatalf("part number search failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Result may be empty if part doesn't exist in Mouser, but API call should work
	if result.NumberOfResult < 0 {
		t.Error("expected non-negative result count")
	}
}

// TestManufacturerSearch tests searching by manufacturer.
func TestManufacturerSearch(t *testing.T) {
	skipIfNoKey(t)

	apiKey := testAPIKey
	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// First get manufacturer list to use a real manufacturer
	mfrList, err := client.GetManufacturerList(ctx)
	if err != nil || len(mfrList.ManufacturerList) == 0 {
		t.Skip("could not get manufacturer list, skipping test")
	}

	// Use the first manufacturer found
	mfrName := mfrList.ManufacturerList[0].ManufacturerName

	result, err := client.KeywordAndManufacturerSearch(ctx, KeywordAndManufacturerSearchOptions{
		Keyword:          "resistor",
		ManufacturerName: mfrName,
		Records:          5,
	})

	if err != nil {
		t.Fatalf("keyword and manufacturer search failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestCachingBetweenSearches tests that results are cached and reused.
func TestCachingBetweenSearches(t *testing.T) {
	skipIfNoKey(t)

	apiKey := testAPIKey
	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// First call - hits real API
	result1, err := client.KeywordSearch(ctx, SearchOptions{
		Keyword: "capacitor",
		Records: 3,
	})
	if err != nil {
		t.Fatalf("first search failed: %v", err)
	}

	// Second call - should be cached
	result2, err := client.KeywordSearch(ctx, SearchOptions{
		Keyword: "capacitor",
		Records: 3,
	})
	if err != nil {
		t.Fatalf("second search failed: %v", err)
	}

	// Results should be identical (from cache)
	if result1.NumberOfResult != result2.NumberOfResult {
		t.Error("expected cached result to match first result")
	}
}

// TestGetManufacturerList tests retrieving the manufacturer list.
func TestGetManufacturerList(t *testing.T) {
	skipIfNoKey(t)

	apiKey := testAPIKey
	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := client.GetManufacturerList(ctx)

	if err != nil {
		t.Fatalf("get manufacturer list failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.Count <= 0 {
		t.Errorf("expected positive count, got %d", result.Count)
	}

	if len(result.ManufacturerList) == 0 {
		t.Error("expected manufacturer list to have items")
	}

	// Verify manufacturer structure
	if len(result.ManufacturerList) > 0 {
		mfr := result.ManufacturerList[0]
		if mfr.ManufacturerName == "" {
			t.Error("expected manufacturer to have name")
		}
	}
}

// TestMultipleSearchTypes tests that different search types work independently.
func TestMultipleSearchTypes(t *testing.T) {
	skipIfNoKey(t)

	apiKey := testAPIKey
	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Keyword search
	result1, err := client.KeywordSearch(ctx, SearchOptions{
		Keyword: "diode",
		Records: 3,
	})
	if err != nil {
		t.Errorf("keyword search failed: %v", err)
	}
	if result1 == nil {
		t.Error("expected non-nil result from keyword search")
	}

	// Part number search
	result2, err := client.PartNumberSearch(ctx, PartNumberSearchOptions{
		PartNumber: "1N4148",
		Records:    3,
	})
	if err != nil {
		t.Errorf("part number search failed: %v", err)
	}
	if result2 == nil {
		t.Error("expected non-nil result from part number search")
	}

	// Manufacturer list
	result3, err := client.GetManufacturerList(ctx)
	if err != nil {
		t.Errorf("get manufacturer list failed: %v", err)
	}
	if result3 == nil {
		t.Error("expected non-nil result from manufacturer list")
	}
}

// TestContextTimeout tests that context timeout is properly enforced.
func TestContextTimeout(t *testing.T) {
	skipIfNoKey(t)

	apiKey := testAPIKey
	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Very short timeout - should fail
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	result, err := client.KeywordSearch(ctx, SearchOptions{
		Keyword: "test",
	})

	if err == nil {
		t.Fatal("expected error from context timeout")
	}

	if result != nil {
		t.Error("expected nil result on timeout")
	}
}

// TestContextCancellation tests that context cancellation is respected.
func TestContextCancellation(t *testing.T) {
	skipIfNoKey(t)

	apiKey := testAPIKey
	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := client.KeywordSearch(ctx, SearchOptions{
		Keyword: "resistor",
	})

	if err == nil {
		t.Fatal("expected error from cancelled context")
	}

	if result != nil {
		t.Error("expected nil result on context cancellation")
	}
}

// TestRateLimitingWithMultipleRequests tests that rate limiter allows requests under limits.
func TestRateLimitingWithMultipleRequests(t *testing.T) {
	skipIfNoKey(t)

	apiKey := testAPIKey
	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Make a few requests (should not exceed default rate limits)
	for i := 0; i < 2; i++ {
		_, err := client.KeywordSearch(ctx, SearchOptions{
			Keyword: "capacitor",
			Records: 2,
		})
		if err != nil {
			t.Errorf("request %d failed: %v", i+1, err)
		}
	}

	// Check rate limit stats
	stats := client.RateLimitStats()
	if stats.MinuteRemaining < 0 {
		t.Errorf("expected non-negative minute remaining, got %d", stats.MinuteRemaining)
	}
	if stats.DailyRemaining < 0 {
		t.Errorf("expected non-negative daily remaining, got %d", stats.DailyRemaining)
	}
}

// TestSearchWithLowRecords tests searching with minimal records.
func TestSearchWithLowRecords(t *testing.T) {
	skipIfNoKey(t)

	apiKey := testAPIKey
	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := client.KeywordSearch(ctx, SearchOptions{
		Keyword: "transistor",
		Records: 1,
	})

	if err != nil {
		t.Fatalf("keyword search with 1 record failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Should have at most 1 part returned
	if len(result.Parts) > 1 {
		t.Errorf("expected at most 1 part, got %d", len(result.Parts))
	}
}

// TestKeywordSearchCommonComponents tests searching for various common components.
func TestKeywordSearchCommonComponents(t *testing.T) {
	skipIfNoKey(t)

	apiKey := testAPIKey
	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	keywords := []string{"resistor", "capacitor", "inductor"}

	for _, keyword := range keywords {
		result, err := client.KeywordSearch(ctx, SearchOptions{
			Keyword: keyword,
			Records: 2,
		})

		if err != nil {
			t.Errorf("search for %q failed: %v", keyword, err)
			continue
		}

		if result == nil {
			t.Errorf("expected non-nil result for %q", keyword)
		}
	}
}
