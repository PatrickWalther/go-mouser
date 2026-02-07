package mouser

// DateFilterType defines date filter options for order history queries.
type DateFilterType string

const (
	DateFilterNone        DateFilterType = "None"
	DateFilterAll         DateFilterType = "All"
	DateFilterToday       DateFilterType = "Today"
	DateFilterYesterday   DateFilterType = "Yesterday"
	DateFilterThisWeek    DateFilterType = "ThisWeek"
	DateFilterLastWeek    DateFilterType = "LastWeek"
	DateFilterThisMonth   DateFilterType = "ThisMonth"
	DateFilterLastMonth   DateFilterType = "LastMonth"
	DateFilterThisQuarter DateFilterType = "ThisQuarter"
	DateFilterLastQuarter DateFilterType = "LastQuarter"
	DateFilterThisYear    DateFilterType = "ThisYear"
	DateFilterLastYear    DateFilterType = "LastYear"
	DateFilterYearToDate  DateFilterType = "YearToDate"
)

// OrderHistoryResponse represents the response from order history queries.
type OrderHistoryResponse struct {
	// Errors contains any API errors.
	Errors []APIError `json:"Errors"`

	// NumberOfOrders is the total number of matching orders.
	NumberOfOrders int `json:"NumberOfOrders"`

	// OrderHistoryItems contains the order history entries.
	OrderHistoryItems []OrderHistoryItem `json:"OrderHistoryItems"`
}

// OrderHistoryItem represents a single order in the order history list.
type OrderHistoryItem struct {
	// DateCreated is the date the order was created.
	DateCreated string `json:"DateCreated"`

	// SalesOrderNumber is the Mouser sales order number.
	SalesOrderNumber string `json:"SalesOrderNumber"`

	// WebOrderNumber is the web order number.
	WebOrderNumber string `json:"WebOrderNumber"`

	// PoNumber is the purchase order number.
	PoNumber string `json:"PoNumber"`

	// BuyerName is the name of the buyer.
	BuyerName string `json:"BuyerName"`

	// OrderStatusDisplay is the display text for the order status.
	OrderStatusDisplay string `json:"OrderStatusDisplay"`
}

// OrderDetailResponse represents the detailed view of a single order.
type OrderDetailResponse struct {
	// Errors contains any API errors.
	Errors []APIError `json:"Errors"`

	// OrderLines contains the line items of the order.
	OrderLines []OrderDetailLine `json:"OrderLines"`

	// SalesOrderId is the Mouser sales order ID.
	SalesOrderId string `json:"SalesOrderId"`

	// WebOrderId is the web order ID.
	WebOrderId string `json:"WebOrderId"`

	// OrderStatus is the numeric order status.
	OrderStatus int `json:"OrderStatus"`

	// OrderStatusName is the display name of the order status.
	OrderStatusName string `json:"OrderStatusName"`

	// OrderDate is the date the order was placed.
	OrderDate string `json:"OrderDate"`

	// BillingAddress is the billing address.
	BillingAddress Address `json:"BillingAddress"`

	// ShippingAddress is the shipping address.
	ShippingAddress Address `json:"ShippingAddress"`

	// PaymentDetail contains payment information.
	PaymentDetail Payment `json:"PaymentDetail"`

	// DeliveryDetail contains delivery information.
	DeliveryDetail Delivery `json:"DeliveryDetail"`

	// CurrencyCode is the currency for the order.
	CurrencyCode string `json:"CurrencyCode"`

	// IsScheduled indicates if the order has scheduled releases.
	IsScheduled bool `json:"IsScheduled"`

	// IsPendingOrder indicates if the order is pending.
	IsPendingOrder bool `json:"IsPendingOrder"`

	// BuyerName is the name of the buyer.
	BuyerName string `json:"BuyerName"`

	// SummaryDetail contains order totals.
	SummaryDetail OrderDetailSummary `json:"SummaryDetail"`
}

// OrderDetailLine represents a line item in an order.
type OrderDetailLine struct {
	// Quantity is the ordered quantity.
	Quantity int `json:"Quantity"`

	// UnitPrice is the unit price.
	UnitPrice float64 `json:"UnitPrice"`

	// ExtPrice is the extended price.
	ExtPrice float64 `json:"ExtPrice"`

	// FormattedUnitPrice is the formatted unit price string.
	FormattedUnitPrice string `json:"FormattedUnitPrice"`

	// FormattedExtendedPrice is the formatted extended price string.
	FormattedExtendedPrice string `json:"FormattedExtendedPrice"`

	// AdditionalFees contains additional fees for this line.
	AdditionalFees []CartAdditionalFee `json:"AdditionalFees"`

	// ProductInfo contains product information for the line item.
	ProductInfo OrderLineProduct `json:"ProductInfo"`

	// Activities contains activity/shipment information.
	Activities []OrderLineActivity `json:"Activities"`
}

// Address represents a postal address.
type Address struct {
	// CountryCode is the country code.
	CountryCode string `json:"CountryCode"`

	// AttentionLine is the attention line.
	AttentionLine string `json:"AttentionLine"`

	// CompanyName is the company name.
	CompanyName string `json:"CompanyName"`

	// AddressOne is the first address line.
	AddressOne string `json:"AddressOne"`

	// AddressTwo is the second address line.
	AddressTwo string `json:"AddressTwo"`

	// City is the city.
	City string `json:"City"`

	// StateOrProvince is the state or province.
	StateOrProvince string `json:"StateOrProvince"`

	// PostalCode is the postal code.
	PostalCode string `json:"PostalCode"`
}

// Payment represents payment information.
type Payment struct {
	// PoNumber is the purchase order number.
	PoNumber string `json:"PoNumber"`

	// PaymentMethodName is the payment method name.
	PaymentMethodName string `json:"PaymentMethodName"`
}

// Delivery represents delivery information.
type Delivery struct {
	// ShippingMethodName is the shipping method name.
	ShippingMethodName string `json:"ShippingMethodName"`

	// BackOrderShippingMethodName is the back order shipping method name.
	BackOrderShippingMethodName string `json:"BackOrderShippingMethodName"`

	// TrackingDetails contains tracking information.
	TrackingDetails []Tracking `json:"TrackingDetails"`
}

// Tracking represents shipment tracking information.
type Tracking struct {
	// Number is the tracking number.
	Number string `json:"Number"`

	// Link is the tracking URL.
	Link string `json:"Link"`
}

// OrderDetailSummary contains order total information.
type OrderDetailSummary struct {
	// MerchandiseTotal is the merchandise total.
	MerchandiseTotal float64 `json:"MerchandiseTotal"`

	// OrderTotal is the total order amount.
	OrderTotal float64 `json:"OrderTotal"`

	// AdditionalFeesTotal is the total additional fees.
	AdditionalFeesTotal float64 `json:"AdditionalFeesTotal"`
}

// OrderLineProduct represents product information for an order line.
type OrderLineProduct struct {
	// MouserPartNumber is the Mouser part number.
	MouserPartNumber string `json:"MouserPartNumber"`

	// CustomerPartNumber is the customer part number.
	CustomerPartNumber string `json:"CustomerPartNumber"`

	// ManufacturerName is the manufacturer name.
	ManufacturerName string `json:"ManufacturerName"`

	// ManufacturerPartNumber is the manufacturer part number.
	ManufacturerPartNumber string `json:"ManufacturerPartNumber"`

	// PartDescription is the part description.
	PartDescription string `json:"PartDescription"`
}

// OrderLineActivity represents an activity (shipment/invoice) on an order line.
type OrderLineActivity struct {
	// InvoiceNumber is the invoice number.
	InvoiceNumber string `json:"InvoiceNumber"`

	// Date is the activity date.
	Date string `json:"Date"`
}
