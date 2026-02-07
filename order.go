package mouser

import (
	"context"
	"encoding/json"
	"net/url"
)

// QueryOrderOptions queries available order options (shipping, payment, etc.) for a cart.
func (c *Client) QueryOrderOptions(ctx context.Context, req OrderOptionsRequest) (*OrderOptionsResponse, error) {
	wrapped := orderOptionsRequestWrapper{OrderOptionsRequest: req}

	var resp OrderOptionsResponse
	if err := c.doRequest(ctx, "POST", "/order/options/query", wrapped, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}

// GetCurrencies retrieves the list of available currencies.
// Results are cached for 24 hours by default.
func (c *Client) GetCurrencies(ctx context.Context, shippingCountryCode string) (*CurrenciesResponse, error) {
	cacheKey := cacheKeyForCurrencies(shippingCountryCode)
	if cached, ok := c.getCached(cacheKey); ok {
		var result CurrenciesResponse
		if err := json.Unmarshal(cached, &result); err == nil {
			return &result, nil
		}
	}

	query := url.Values{}
	if shippingCountryCode != "" {
		query.Set("shippingCountryCode", shippingCountryCode)
	}

	var resp CurrenciesResponse
	if err := c.doRequestWithQuery(ctx, "GET", "/order/currencies", query, nil, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	if data, err := json.Marshal(resp); err == nil {
		c.setCache(cacheKey, data, c.cacheConfig.CurrenciesTTL)
	}

	return &resp, nil
}

// GetCountries retrieves the list of available countries and their states/provinces.
// Results are cached for 24 hours by default.
func (c *Client) GetCountries(ctx context.Context, countryCode string) (*CountriesResponse, error) {
	cacheKey := cacheKeyForCountries(countryCode)
	if cached, ok := c.getCached(cacheKey); ok {
		var result CountriesResponse
		if err := json.Unmarshal(cached, &result); err == nil {
			return &result, nil
		}
	}

	query := url.Values{}
	if countryCode != "" {
		query.Set("countryCode", countryCode)
	}

	var resp CountriesResponse
	if err := c.doRequestWithQuery(ctx, "GET", "/order/countries", query, nil, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	if data, err := json.Marshal(resp); err == nil {
		c.setCache(cacheKey, data, c.cacheConfig.CountriesTTL)
	}

	return &resp, nil
}

// CreateOrder creates a new order from a cart.
func (c *Client) CreateOrder(ctx context.Context, req CreateOrderRequest) (*OrderResponse, error) {
	wrapped := createOrderRequestWrapper{CreateOrderRequest: req}

	var resp OrderResponse
	if err := c.doRequest(ctx, "POST", "/order", wrapped, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}

// CreateOrderFromPrevious creates a new order based on a previous order.
func (c *Client) CreateOrderFromPrevious(ctx context.Context, orderNumber, countryCode, currencyCode string, req CreateOrderRequest) (*OrderResponse, error) {
	query := url.Values{}
	query.Set("orderNumber", orderNumber)
	if countryCode != "" {
		query.Set("countryCode", countryCode)
	}
	if currencyCode != "" {
		query.Set("currencyCode", currencyCode)
	}

	wrapped := createOrderRequestWrapper{CreateOrderRequest: req}

	var resp OrderResponse
	if err := c.doRequestWithQuery(ctx, "POST", "/order/CreateFromOrder", query, wrapped, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}

// GetOrderDetails retrieves details for a specific order by order number.
func (c *Client) GetOrderDetails(ctx context.Context, orderNumber string) (*OrderResponse, error) {
	path := "/order/" + url.PathEscape(orderNumber)

	var resp OrderResponse
	if err := c.doRequest(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}

// CreateCartFromOrder creates a new cart from an existing order.
func (c *Client) CreateCartFromOrder(ctx context.Context, orderNumber, countryCode, currencyCode string) (*CartResponse, error) {
	query := url.Values{}
	query.Set("orderNumber", orderNumber)
	if countryCode != "" {
		query.Set("countryCode", countryCode)
	}
	if currencyCode != "" {
		query.Set("currencyCode", currencyCode)
	}

	var resp CartResponse
	if err := c.doRequestWithQuery(ctx, "POST", "/order/item/CreateCartFromOrder", query, nil, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}
