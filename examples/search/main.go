// Example CLI tool for searching Mouser components.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/PatrickWalther/go-mouser"
)

func main() {
	var (
		apiKey     = flag.String("api-key", os.Getenv("MOUSER_API_KEY"), "Mouser API key (or set MOUSER_API_KEY env var)")
		keyword    = flag.String("keyword", "", "Search keyword")
		partNumber = flag.String("part", "", "Part number to search")
		mfrName    = flag.String("mfr", "", "Manufacturer name (use with -keyword)")
		records    = flag.Int("records", 10, "Number of results to return (max 50)")
	)
	flag.Parse()

	if *apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: API key is required. Use -api-key flag or set MOUSER_API_KEY environment variable.")
		os.Exit(1)
	}

	if *keyword == "" && *partNumber == "" {
		fmt.Fprintln(os.Stderr, "Error: Either -keyword or -part is required.")
		flag.Usage()
		os.Exit(1)
	}

	client, err := mouser.NewClient(*apiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	var result *mouser.SearchResult

	if *partNumber != "" {
		result, err = client.PartNumberSearch(ctx, mouser.PartNumberSearchOptions{
			PartNumber: *partNumber,
			Records:    *records,
		})
	} else if *mfrName != "" {
		result, err = client.KeywordAndManufacturerSearch(ctx, mouser.KeywordAndManufacturerSearchOptions{
			Keyword:          *keyword,
			ManufacturerName: *mfrName,
			Records:          *records,
		})
	} else {
		result, err = client.KeywordSearch(ctx, mouser.SearchOptions{
			Keyword: *keyword,
			Records: *records,
		})
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error searching: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d results\n\n", result.NumberOfResult)

	for i, part := range result.Parts {
		fmt.Printf("%d. %s\n", i+1, part.ManufacturerPartNumber)
		fmt.Printf("   Mouser PN:    %s\n", part.MouserPartNumber)
		fmt.Printf("   Manufacturer: %s\n", part.Manufacturer)
		fmt.Printf("   Description:  %s\n", part.Description)
		fmt.Printf("   Availability: %s\n", part.Availability)
		fmt.Printf("   Min Qty:      %s\n", part.Min)
		if len(part.PriceBreaks) > 0 {
			fmt.Printf("   Price:        %s %s (qty %d)\n",
				part.PriceBreaks[0].Price,
				part.PriceBreaks[0].Currency,
				part.PriceBreaks[0].Quantity,
			)
		}
		if part.DataSheetUrl != "" {
			fmt.Printf("   Datasheet:    %s\n", part.DataSheetUrl)
		}
		fmt.Println()
	}
}
