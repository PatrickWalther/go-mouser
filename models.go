package mouser

// SearchOptions contains options for keyword search requests.
type SearchOptions struct {
	// Keyword is the search term.
	Keyword string

	// Records is the maximum number of results to return (max 50).
	Records int

	// StartingRecord is the starting index for pagination (0-based).
	StartingRecord int

	// SearchWithYourSignUpLanguage uses the language from your Mouser account.
	SearchWithYourSignUpLanguage bool

	// SearchOption filters results. Valid values: None, Rohs, InStock, RohsAndInStock
	SearchOption SearchOptionType
}

// SearchOptionType defines search filter options.
type SearchOptionType string

const (
	SearchOptionNone           SearchOptionType = "None"
	SearchOptionRohs           SearchOptionType = "Rohs"
	SearchOptionInStock        SearchOptionType = "InStock"
	SearchOptionRohsAndInStock SearchOptionType = "RohsAndInStock"
)

// PartNumberSearchOptions contains options for part number search requests.
type PartNumberSearchOptions struct {
	// PartNumber is the part number to search for.
	// Multiple part numbers can be separated by pipe (|), max 10.
	PartNumber string

	// Deprecated: Records is not supported by the V1 part number search API and is ignored.
	Records int

	// Deprecated: StartingRecord is not supported by the V1 part number search API and is ignored.
	StartingRecord int

	// Deprecated: SearchWithYourSignUpLanguage is not supported by the V1 part number search API and is ignored.
	SearchWithYourSignUpLanguage bool

	// PartSearchOption controls matching. Valid values: None, Exact
	PartSearchOption PartSearchOptionType
}

// PartSearchOptionType defines part search matching options.
type PartSearchOptionType string

const (
	PartSearchOptionNone  PartSearchOptionType = "None"
	PartSearchOptionExact PartSearchOptionType = "Exact"
)

// KeywordAndManufacturerSearchOptions contains options for keyword and manufacturer search requests.
type KeywordAndManufacturerSearchOptions struct {
	// Keyword is the search term.
	Keyword string

	// ManufacturerName is the manufacturer name to filter by.
	// Use GetManufacturerList to get valid manufacturer names.
	ManufacturerName string

	// Records is the maximum number of results to return (max 50).
	Records int

	// PageNumber is the page number for pagination (1-based).
	PageNumber int

	// SearchOption filters results. Valid values: None, Rohs, InStock, RohsAndInStock
	SearchOption SearchOptionType

	// SearchWithYourSignUpLanguage uses the language from your Mouser account.
	SearchWithYourSignUpLanguage bool
}

// PartNumberAndManufacturerSearchOptions contains options for part number and manufacturer search.
type PartNumberAndManufacturerSearchOptions struct {
	// PartNumber is the Mouser part number to search for.
	// Multiple part numbers can be separated by pipe (|), max 10.
	PartNumber string

	// ManufacturerName is the manufacturer name to filter by.
	ManufacturerName string

	// PartSearchOption controls matching. Valid values: None, Exact
	PartSearchOption PartSearchOptionType
}

// SearchResult represents the result of a search operation.
type SearchResult struct {
	// NumberOfResult is the total number of matching results.
	NumberOfResult int `json:"NumberOfResult"`

	// Parts is the list of matching parts.
	Parts []Part `json:"Parts"`
}

// Part represents a component from Mouser's catalog.
type Part struct {
	// MouserPartNumber is the Mouser-assigned part number.
	MouserPartNumber string `json:"MouserPartNumber"`

	// ManufacturerPartNumber is the manufacturer's part number.
	ManufacturerPartNumber string `json:"ManufacturerPartNumber"`

	// Manufacturer is the manufacturer name.
	Manufacturer string `json:"Manufacturer"`

	// Description is the part description.
	Description string `json:"Description"`

	// DataSheetUrl is the URL to the part's datasheet.
	DataSheetUrl string `json:"DataSheetUrl"`

	// ImagePath is the URL to the part's image.
	ImagePath string `json:"ImagePath"`

	// Category is the part category.
	Category string `json:"Category"`

	// Availability is the stock availability message.
	Availability string `json:"Availability"`

	// AvailabilityInStock is the quantity in stock.
	AvailabilityInStock string `json:"AvailabilityInStock"`

	// AvailabilityOnOrder is the quantity on order (array of scheduled availability).
	AvailabilityOnOrder []AvailabilityOnOrderObject `json:"AvailabilityOnOrder"`

	// FactoryStock is the factory stock quantity.
	FactoryStock string `json:"FactoryStock"`

	// LifecycleStatus indicates if the part is active, obsolete, etc.
	LifecycleStatus string `json:"LifecycleStatus"`

	// ROHSStatus is the RoHS compliance status.
	ROHSStatus string `json:"ROHSStatus"`

	// LeadTime is the lead time for the part.
	LeadTime string `json:"LeadTime"`

	// Min is the minimum order quantity.
	Min string `json:"Min"`

	// Mult is the order multiple.
	Mult string `json:"Mult"`

	// ProductDetailUrl is the URL to the product detail page.
	ProductDetailUrl string `json:"ProductDetailUrl"`

	// PriceBreaks is the list of price breaks.
	PriceBreaks []PriceBreak `json:"PriceBreaks"`

	// AlternatePackagings contains alternate packaging options.
	AlternatePackagings []AlternatePackaging `json:"AlternatePackagings"`

	// ProductAttributes contains additional product attributes.
	ProductAttributes []ProductAttribute `json:"ProductAttributes"`

	// UnitWeightKg contains the unit weight in kilograms.
	UnitWeightKg UnitWeight `json:"UnitWeightKg"`

	// StandardCost is the standard cost.
	StandardCost StandardCost `json:"StandardCost"`

	// Reeling indicates if the part can be reeled.
	Reeling bool `json:"Reeling"`

	// SuggestedReplacement is a suggested replacement part number.
	SuggestedReplacement string `json:"SuggestedReplacement"`

	// MultiSimBlue indicates multi-sim blue status.
	MultiSimBlue int `json:"MultiSimBlue"`

	// InfoMessages contains informational messages.
	InfoMessages []string `json:"InfoMessages"`

	// IsDiscontinued indicates if the part is discontinued.
	IsDiscontinued string `json:"IsDiscontinued"`

	// MouserProductCategory is the Mouser product category.
	MouserProductCategory string `json:"MouserProductCategory"`

	// IPCCode is the IPC code.
	IPCCode string `json:"IPCCode"`

	// ProductCompliance contains compliance information.
	ProductCompliance []ProductCompliance `json:"ProductCompliance"`

	// RestrictionMessage contains any restriction messages.
	RestrictionMessage string `json:"RestrictionMessage"`

	// ActualMfrName is the actual manufacturer name.
	ActualMfrName string `json:"ActualMfrName"`

	// AvailableOnOrder is the total quantity available on order.
	AvailableOnOrder string `json:"AvailableOnOrder"`

	// PID is the product identifier.
	PID string `json:"PID"`

	// REACH_SVHC contains REACH Substances of Very High Concern.
	REACH_SVHC []string `json:"REACH-SVHC"`

	// RTM is the RTM field.
	RTM string `json:"RTM"`

	// SField is the S field.
	SField string `json:"SField"`

	// SalesMaximumOrderQty is the maximum order quantity for sales.
	SalesMaximumOrderQty string `json:"SalesMaximumOrderQty"`

	// SurchargeMessages contains surcharge messages for the part.
	SurchargeMessages []SurchargeMessage `json:"SurchargeMessages"`

	// VNum is the V number.
	VNum string `json:"VNum"`
}

// AvailabilityOnOrderObject represents a scheduled on-order availability.
type AvailabilityOnOrderObject struct {
	// Quantity is the quantity available.
	Quantity int `json:"Quantity"`

	// Date is the date the quantity will be available.
	Date string `json:"Date"`
}

// PriceBreak represents a quantity-based price break.
type PriceBreak struct {
	// Quantity is the minimum quantity for this price break.
	Quantity int `json:"Quantity"`

	// Price is the unit price at this quantity.
	Price string `json:"Price"`

	// Currency is the currency code.
	Currency string `json:"Currency"`
}

// AlternatePackaging represents an alternate packaging option.
type AlternatePackaging struct {
	// APMfrPN is the alternate manufacturer part number.
	APMfrPN string `json:"APMfrPN"`
}

// ProductAttribute represents a product attribute.
type ProductAttribute struct {
	// AttributeName is the name of the attribute.
	AttributeName string `json:"AttributeName"`

	// AttributeValue is the value of the attribute.
	AttributeValue string `json:"AttributeValue"`

	// AttributeCost is the cost associated with this attribute.
	AttributeCost string `json:"AttributeCost"`
}

// SurchargeMessage represents a surcharge message associated with a part.
type SurchargeMessage struct {
	// Code is the surcharge code.
	Code string `json:"code"`

	// Message is the surcharge message text.
	Message string `json:"message"`
}

// UnitWeight represents the unit weight.
type UnitWeight struct {
	// UnitWeight is the weight value.
	UnitWeight float64 `json:"UnitWeight"`
}

// StandardCost represents standard cost information.
type StandardCost struct {
	// StandardCost is the cost value.
	StandardCost float64 `json:"StandardCost"`
}

// ProductCompliance represents product compliance information.
type ProductCompliance struct {
	// ComplianceName is the name of the compliance standard.
	ComplianceName string `json:"ComplianceName"`

	// ComplianceValue is the compliance value/status.
	ComplianceValue string `json:"ComplianceValue"`
}

// Manufacturer represents a manufacturer in the Mouser catalog.
type Manufacturer struct {
	// ManufacturerName is the name of the manufacturer.
	ManufacturerName string `json:"ManufacturerName"`
}

// ManufacturerListResult represents the result of a manufacturer list query.
type ManufacturerListResult struct {
	// Count is the total number of manufacturers.
	Count int `json:"Count"`

	// ManufacturerList is the list of manufacturers.
	ManufacturerList []Manufacturer `json:"ManufacturerList"`
}

// --- Request types for Mouser API ---

// keywordSearchRequest is the request format for keyword search (V1 compatibility).
type keywordSearchRequest struct {
	SearchByKeywordRequest searchByKeywordRequest `json:"SearchByKeywordRequest"`
}

type searchByKeywordRequest struct {
	Keyword                      string `json:"keyword"`
	Records                      int    `json:"records"`
	StartingRecord               int    `json:"startingRecord"`
	SearchOptions                string `json:"searchOptions,omitempty"`
	SearchWithYourSignUpLanguage bool   `json:"searchWithYourSignUpLanguage"`
}

// partNumberSearchRequest is the request format for part number search (V1 compatibility).
type partNumberSearchRequest struct {
	SearchByPartRequest searchByPartRequest `json:"SearchByPartRequest"`
}

type searchByPartRequest struct {
	MouserPartNumber  string `json:"mouserPartNumber"`
	PartSearchOptions string `json:"partSearchOptions,omitempty"`
}

// keywordAndManufacturerSearchRequest is the request format for V2 keyword and manufacturer search.
type keywordAndManufacturerSearchRequest struct {
	SearchByKeywordMfrNameRequest searchByKeywordMfrNameRequest `json:"SearchByKeywordMfrNameRequest"`
}

type searchByKeywordMfrNameRequest struct {
	Keyword                      string `json:"keyword"`
	ManufacturerName             string `json:"manufacturerName,omitempty"`
	Records                      int    `json:"records,omitempty"`
	PageNumber                   int    `json:"pageNumber,omitempty"`
	SearchOptions                string `json:"searchOptions,omitempty"`
	SearchWithYourSignUpLanguage bool   `json:"searchWithYourSignUpLanguage"`
}

// partNumberAndManufacturerSearchRequest is the request format for V2 part number and manufacturer search.
type partNumberAndManufacturerSearchRequest struct {
	SearchByPartMfrNameRequest searchByPartMfrNameRequest `json:"SearchByPartMfrNameRequest"`
}

type searchByPartMfrNameRequest struct {
	MouserPartNumber  string `json:"mouserPartNumber"`
	ManufacturerName  string `json:"manufacturerName,omitempty"`
	PartSearchOptions string `json:"partSearchOptions,omitempty"`
}

// --- Response types ---

// searchResponse is the response format for search requests.
type searchResponse struct {
	Errors        []APIError   `json:"Errors"`
	SearchResults SearchResult `json:"SearchResults"`
}

// manufacturerListResponse is the response format for manufacturer list requests.
type manufacturerListResponse struct {
	Errors                 []APIError             `json:"Errors"`
	MouserManufacturerList ManufacturerListResult `json:"MouserManufacturerList"`
}
