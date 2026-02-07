package mouser

import (
	"context"
	"net/url"
)

// GetCart retrieves the contents of a cart.
func (c *Client) GetCart(ctx context.Context, cartKey, countryCode, currencyCode string) (*CartResponse, error) {
	query := url.Values{}
	query.Set("cartKey", cartKey)
	if countryCode != "" {
		query.Set("countryCode", countryCode)
	}
	if currencyCode != "" {
		query.Set("currencyCode", currencyCode)
	}

	var resp CartResponse
	if err := c.doRequestWithQuery(ctx, "GET", "/cart", query, nil, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}

// UpdateCart updates an existing cart with the provided items.
func (c *Client) UpdateCart(ctx context.Context, body CartItemRequestBody, countryCode, currencyCode string) (*CartResponse, error) {
	query := url.Values{}
	if countryCode != "" {
		query.Set("countryCode", countryCode)
	}
	if currencyCode != "" {
		query.Set("currencyCode", currencyCode)
	}

	var resp CartResponse
	if err := c.doRequestWithQuery(ctx, "POST", "/cart", query, body, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}

// InsertCartItems inserts new items into a cart.
func (c *Client) InsertCartItems(ctx context.Context, body CartItemRequestBody, countryCode, currencyCode string) (*CartResponse, error) {
	query := url.Values{}
	if countryCode != "" {
		query.Set("countryCode", countryCode)
	}
	if currencyCode != "" {
		query.Set("currencyCode", currencyCode)
	}

	var resp CartResponse
	if err := c.doRequestWithQuery(ctx, "POST", "/cart/items/insert", query, body, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}

// UpdateCartItems updates existing items in a cart.
func (c *Client) UpdateCartItems(ctx context.Context, body CartItemRequestBody, countryCode, currencyCode string) (*CartResponse, error) {
	query := url.Values{}
	if countryCode != "" {
		query.Set("countryCode", countryCode)
	}
	if currencyCode != "" {
		query.Set("currencyCode", currencyCode)
	}

	var resp CartResponse
	if err := c.doRequestWithQuery(ctx, "POST", "/cart/items/update", query, body, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}

// RemoveCartItem removes an item from the cart.
func (c *Client) RemoveCartItem(ctx context.Context, cartKey, mouserPartNumber, countryCode, currencyCode string) (*CartResponse, error) {
	query := url.Values{}
	query.Set("cartKey", cartKey)
	query.Set("mouserPartNumber", mouserPartNumber)
	if countryCode != "" {
		query.Set("countryCode", countryCode)
	}
	if currencyCode != "" {
		query.Set("currencyCode", currencyCode)
	}

	var resp CartResponse
	if err := c.doRequestWithQuery(ctx, "POST", "/cart/item/remove", query, nil, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}

// InsertCartSchedule inserts scheduled releases for cart items.
func (c *Client) InsertCartSchedule(ctx context.Context, body ScheduleCartItemsRequestBody) (*CartResponse, error) {
	var resp CartResponse
	if err := c.doRequest(ctx, "POST", "/cart/insert/schedule", body, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}

// UpdateCartSchedule updates scheduled releases for cart items.
func (c *Client) UpdateCartSchedule(ctx context.Context, body ScheduleCartItemsRequestBody) (*CartResponse, error) {
	var resp CartResponse
	if err := c.doRequest(ctx, "POST", "/cart/update/schedule", body, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}

// DeleteAllCartSchedules deletes all scheduled releases for a cart.
func (c *Client) DeleteAllCartSchedules(ctx context.Context, cartKey string) (*CartResponse, error) {
	query := url.Values{}
	query.Set("cartKey", cartKey)

	var resp CartResponse
	if err := c.doRequestWithQuery(ctx, "POST", "/cart/deleteall/schedule", query, nil, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}
