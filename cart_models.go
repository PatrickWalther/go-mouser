package mouser

// PackagingChoiceType defines the packaging choice for a cart item.
type PackagingChoiceType string

const (
	PackagingChoiceNone     PackagingChoiceType = "None"
	PackagingChoiceCutTape  PackagingChoiceType = "Cut_Tape"
	PackagingChoiceMouseReel PackagingChoiceType = "MouseReel"
	PackagingChoiceFullReel PackagingChoiceType = "FullReel"
)

// CartItemRequest represents a single item to add/update in a cart.
type CartItemRequest struct {
	// MouserPartNumber is the Mouser part number.
	MouserPartNumber string `json:"MouserPartNumber"`

	// Quantity is the desired quantity.
	Quantity int `json:"Quantity"`

	// CustomerPartNumber is an optional customer-assigned part number.
	CustomerPartNumber string `json:"CustomerPartNumber,omitempty"`

	// PackagingChoice is the packaging type. Valid values: None, Cut_Tape, MouseReel, FullReel.
	PackagingChoice PackagingChoiceType `json:"PackagingChoice,omitempty"`
}

// CartItemRequestBody is the request body for cart insert/update operations.
type CartItemRequestBody struct {
	// CartKey is the unique cart identifier (UUID).
	CartKey string `json:"CartKey,omitempty"`

	// MouserPaysCustomsAndDuties indicates whether Mouser pays customs and duties.
	MouserPaysCustomsAndDuties bool `json:"MouserPaysCustomsAndDuties,omitempty"`

	// CartItems is the list of items to add or update.
	CartItems []CartItemRequest `json:"CartItems"`
}

// CartResponse represents the response from cart operations.
type CartResponse struct {
	// Errors contains any API errors.
	Errors []APIError `json:"Errors"`

	// CartKey is the unique cart identifier.
	CartKey string `json:"CartKey"`

	// CurrencyCode is the currency for the cart.
	CurrencyCode string `json:"CurrencyCode"`

	// CartItems contains the items in the cart.
	CartItems []CartOrderLine `json:"CartItems"`

	// TotalItemCount is the total number of items in the cart.
	TotalItemCount int `json:"TotalItemCount"`

	// AdditionalFeesTotal is the total of additional fees.
	AdditionalFeesTotal float64 `json:"AdditionalFeesTotal"`

	// MerchandiseTotal is the total merchandise cost.
	MerchandiseTotal float64 `json:"MerchandiseTotal"`
}

// CartOrderLine represents a single line item in a cart.
type CartOrderLine struct {
	// Errors contains any errors for this line item.
	Errors []APIError `json:"Errors"`

	// MouserATS is the Mouser available-to-ship quantity.
	MouserATS string `json:"MouserATS"`

	// Quantity is the quantity in the cart.
	Quantity int `json:"Quantity"`

	// PartsPerReel is the number of parts per reel.
	PartsPerReel int `json:"PartsPerReel"`

	// PackagingChoice is the selected packaging.
	PackagingChoice string `json:"PackagingChoice"`

	// ScheduledReleases contains scheduled release information.
	ScheduledReleases []ScheduleRelease `json:"ScheduledReleases"`

	// InfoMessages contains informational messages.
	InfoMessages []string `json:"InfoMessages"`

	// MouserPartNumber is the Mouser part number.
	MouserPartNumber string `json:"MouserPartNumber"`

	// MfrPartNumber is the manufacturer part number.
	MfrPartNumber string `json:"MfrPartNumber"`

	// Description is the part description.
	Description string `json:"Description"`

	// CartItemCustPartNumber is the customer part number.
	CartItemCustPartNumber string `json:"CartItemCustPartNumber"`

	// UnitPrice is the unit price.
	UnitPrice float64 `json:"UnitPrice"`

	// ExtendedPrice is the extended price (quantity * unit price).
	ExtendedPrice float64 `json:"ExtendedPrice"`

	// LifeCycle is the lifecycle status.
	LifeCycle string `json:"LifeCycle"`

	// Manufacturer is the manufacturer name.
	Manufacturer string `json:"Manufacturer"`

	// SalesMultipleQty is the sales multiple quantity.
	SalesMultipleQty string `json:"SalesMultipleQty"`

	// SalesMinimumOrderQty is the minimum order quantity.
	SalesMinimumOrderQty string `json:"SalesMinimumOrderQty"`

	// SalesMaximumOrderQty is the maximum order quantity.
	SalesMaximumOrderQty string `json:"SalesMaximumOrderQty"`

	// AdditionalFees contains additional fees for this line item.
	AdditionalFees []CartAdditionalFee `json:"AdditionalFees"`
}

// CartAdditionalFee represents an additional fee on a cart line item.
type CartAdditionalFee struct {
	// Amount is the fee amount per unit.
	Amount float64 `json:"Amount"`

	// ExtendedAmount is the total fee amount (quantity * amount).
	ExtendedAmount float64 `json:"ExtendedAmount"`

	// Code is the fee code.
	Code string `json:"Code"`
}

// ScheduleCartItemsRequestBody is the request body for scheduling cart items.
type ScheduleCartItemsRequestBody struct {
	// CartKey is the unique cart identifier.
	CartKey string `json:"CartKey"`

	// ScheduleCartItems contains the items to schedule.
	ScheduleCartItems []ScheduleReleaseRequest `json:"ScheduleCartItems"`
}

// ScheduleReleaseRequest represents a schedule release request for a part.
type ScheduleReleaseRequest struct {
	// MouserPartNumber is the Mouser part number.
	MouserPartNumber string `json:"MouserPartNumber"`

	// ScheduledReleases is the list of scheduled releases.
	ScheduledReleases []ScheduleRelease `json:"ScheduledReleases"`
}

// ScheduleRelease represents a scheduled release with a date and quantity.
type ScheduleRelease struct {
	// Key is the scheduled release date.
	Key string `json:"Key"`

	// Value is the quantity for this release.
	Value int `json:"Value"`
}
