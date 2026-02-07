package mouser

// AddressLocationTypeID defines the type of address location.
type AddressLocationTypeID string

const (
	AddressLocationResidential AddressLocationTypeID = "Residential"
	AddressLocationCommercial  AddressLocationTypeID = "Commercial"
	AddressLocationPostOffice  AddressLocationTypeID = "PostOfficeBox"
	AddressLocationAPOFPO      AddressLocationTypeID = "APOFPO"
)

// OrderType defines the type of order.
type OrderType string

const (
	OrderTypeUnspecified OrderType = "Unspecified"
	OrderTypeRush        OrderType = "Rush"
	OrderTypeComplete    OrderType = "Complete"
)

// OrderAddress represents an address for order operations.
type OrderAddress struct {
	// AddressLocationTypeID is the type of address location.
	AddressLocationTypeID AddressLocationTypeID `json:"AddressLocationTypeID,omitempty"`

	// CountryCode is the country code.
	CountryCode string `json:"CountryCode,omitempty"`

	// FirstName is the first name.
	FirstName string `json:"FirstName,omitempty"`

	// LastName is the last name.
	LastName string `json:"LastName,omitempty"`

	// AttentionLine is the attention line.
	AttentionLine string `json:"AttentionLine,omitempty"`

	// Company is the company name.
	Company string `json:"Company,omitempty"`

	// AddressOne is the first address line.
	AddressOne string `json:"AddressOne,omitempty"`

	// AddressTwo is the second address line.
	AddressTwo string `json:"AddressTwo,omitempty"`

	// City is the city.
	City string `json:"City,omitempty"`

	// StateOrProvince is the state or province.
	StateOrProvince string `json:"StateOrProvince,omitempty"`

	// PostalCode is the postal code.
	PostalCode string `json:"PostalCode,omitempty"`

	// PhoneNumber is the phone number.
	PhoneNumber string `json:"PhoneNumber,omitempty"`

	// PhoneExtension is the phone extension.
	PhoneExtension string `json:"PhoneExtension,omitempty"`

	// EmailAddress is the email address.
	EmailAddress string `json:"EmailAddress,omitempty"`
}

// OrderOptionsRequest is the request body for querying order options.
type OrderOptionsRequest struct {
	// ShippingAddress is the shipping address for the order.
	ShippingAddress *OrderAddress `json:"ShippingAddress,omitempty"`

	// CurrencyCode is the desired currency code.
	CurrencyCode string `json:"CurrencyCode,omitempty"`

	// CartKey is the cart key to query options for.
	CartKey string `json:"CartKey,omitempty"`
}

// orderOptionsRequestWrapper wraps the options request for the API.
type orderOptionsRequestWrapper struct {
	OrderOptionsRequest OrderOptionsRequest `json:"OrderOptionsRequest"`
}

// OrderOptionsResponse is the response from querying order options.
type OrderOptionsResponse struct {
	// Errors contains any API errors.
	Errors []APIError `json:"Errors"`

	// CurrencyCode is the currency for the order.
	CurrencyCode string `json:"CurrencyCode"`

	// BillingAddress is the billing address.
	BillingAddress Address `json:"BillingAddress"`

	// ShippingAddress is the shipping address.
	ShippingAddress Address `json:"ShippingAddress"`

	// Shipping contains shipping options.
	Shipping ShippingOptions `json:"Shipping"`

	// Payment contains payment options.
	Payment PaymentOptions `json:"Payment"`

	// Languages contains available languages.
	Languages []string `json:"Languages"`
}

// ShippingOptions contains available shipping methods and freight accounts.
type ShippingOptions struct {
	// Methods contains the available shipping methods.
	Methods []ShippingMethod `json:"Methods"`

	// FreightAccounts contains available freight accounts.
	FreightAccounts []FreightAccount `json:"FreightAccounts"`
}

// ShippingMethod represents a shipping method with its rate.
type ShippingMethod struct {
	// Method is the shipping method name.
	Method string `json:"Method"`

	// Rate is the shipping rate.
	Rate float64 `json:"Rate"`

	// Code is the shipping method code.
	Code int `json:"Code"`
}

// FreightAccount represents a freight account.
type FreightAccount struct {
	// Number is the freight account number.
	Number string `json:"Number"`

	// Type is the freight account type.
	Type string `json:"Type"`
}

// PaymentOptions contains available payment methods.
type PaymentOptions struct {
	// PaymentTypes contains available payment types.
	PaymentTypes []string `json:"PaymentTypes"`

	// TaxCertificates contains available tax certificates.
	TaxCertificates []string `json:"TaxCertificates"`

	// VatAccounts contains available VAT accounts.
	VatAccounts []string `json:"VatAccounts"`
}

// CurrenciesResponse is the response from the currencies endpoint.
type CurrenciesResponse struct {
	// Errors contains any API errors.
	Errors []APIError `json:"Errors"`

	// Currencies contains the list of available currencies.
	Currencies []Currency `json:"Currencies"`
}

// Currency represents a currency.
type Currency struct {
	// CurrencyCode is the ISO currency code.
	CurrencyCode string `json:"CurrencyCode"`

	// CurrencyName is the display name of the currency.
	CurrencyName string `json:"CurrencyName"`
}

// CountriesResponse is the response from the countries endpoint.
type CountriesResponse struct {
	// Errors contains any API errors.
	Errors []APIError `json:"Errors"`

	// Countries contains the list of countries.
	Countries []Country `json:"Countries"`
}

// Country represents a country with its states/provinces.
type Country struct {
	// CountryName is the display name of the country.
	CountryName string `json:"CountryName"`

	// CountryCode is the ISO country code.
	CountryCode string `json:"CountryCode"`

	// States contains the states/provinces for this country.
	States []State `json:"States"`
}

// State represents a state or province within a country.
type State struct {
	// StateName is the display name of the state.
	StateName string `json:"StateName"`

	// StateCode is the state code.
	StateCode string `json:"StateCode"`
}

// CreateOrderRequest is the request body for creating an order.
type CreateOrderRequest struct {
	// ShippingAddress is the shipping address.
	ShippingAddress *OrderAddress `json:"ShippingAddress,omitempty"`

	// OrderType is the type of order.
	OrderType OrderType `json:"OrderType,omitempty"`

	// PrimaryShipping is the primary shipping method code.
	PrimaryShipping int `json:"PrimaryShipping,omitempty"`

	// SecondaryShipping is the secondary (backorder) shipping method code.
	SecondaryShipping int `json:"SecondaryShipping,omitempty"`

	// Payment is the payment method.
	Payment string `json:"Payment,omitempty"`

	// CurrencyCode is the currency code.
	CurrencyCode string `json:"CurrencyCode,omitempty"`

	// CartKey is the cart to order from.
	CartKey string `json:"CartKey,omitempty"`

	// LanguageCode is the language code.
	LanguageCode string `json:"LanguageCode,omitempty"`

	// SubmitOrder indicates whether to submit the order (true) or just validate (false).
	SubmitOrder bool `json:"SubmitOrder"`
}

// createOrderRequestWrapper wraps the order request for the API.
type createOrderRequestWrapper struct {
	CreateOrderRequest CreateOrderRequest `json:"CreateOrderRequest"`
}

// OrderResponse is the response from order creation/retrieval.
type OrderResponse struct {
	// Errors contains any API errors.
	Errors []APIError `json:"Errors"`

	// OrderNumber is the created order number.
	OrderNumber string `json:"OrderNumber"`

	// CartKey is the cart key associated with the order.
	CartKey string `json:"CartKey"`

	// CurrencyCode is the currency code.
	CurrencyCode string `json:"CurrencyCode"`

	// OrderLines contains the order line items.
	OrderLines []OrderDetailLine `json:"OrderLines"`

	// BillingAddress is the billing address.
	BillingAddress Address `json:"BillingAddress"`

	// ShippingAddress is the shipping address.
	ShippingAddress Address `json:"ShippingAddress"`

	// SummaryDetail contains order totals.
	SummaryDetail OrderDetailSummary `json:"SummaryDetail"`
}
