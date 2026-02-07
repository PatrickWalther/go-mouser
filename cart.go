package mouser

import (
	"context"
	"net/url"
)

// Get retrieves the contents of a cart.
func (s *CartService) Get(ctx context.Context, cartKey, countryCode, currencyCode string) (*CartResponse, error) {
	c := s.client

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

// Update updates an existing cart with the provided items.
func (s *CartService) Update(ctx context.Context, body CartItemRequestBody, countryCode, currencyCode string) (*CartResponse, error) {
	c := s.client

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

// InsertItems inserts new items into a cart.
func (s *CartService) InsertItems(ctx context.Context, body CartItemRequestBody, countryCode, currencyCode string) (*CartResponse, error) {
	c := s.client

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

// UpdateItems updates existing items in a cart.
func (s *CartService) UpdateItems(ctx context.Context, body CartItemRequestBody, countryCode, currencyCode string) (*CartResponse, error) {
	c := s.client

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

// RemoveItem removes an item from the cart.
func (s *CartService) RemoveItem(ctx context.Context, cartKey, mouserPartNumber, countryCode, currencyCode string) (*CartResponse, error) {
	c := s.client

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

// InsertSchedule inserts scheduled releases for cart items.
func (s *CartService) InsertSchedule(ctx context.Context, body ScheduleCartItemsRequestBody) (*CartResponse, error) {
	c := s.client

	var resp CartResponse
	if err := c.doRequest(ctx, "POST", "/cart/insert/schedule", body, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}

// UpdateSchedule updates scheduled releases for cart items.
func (s *CartService) UpdateSchedule(ctx context.Context, body ScheduleCartItemsRequestBody) (*CartResponse, error) {
	c := s.client

	var resp CartResponse
	if err := c.doRequest(ctx, "POST", "/cart/update/schedule", body, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}

// DeleteAllSchedules deletes all scheduled releases for a cart.
func (s *CartService) DeleteAllSchedules(ctx context.Context, cartKey string) (*CartResponse, error) {
	c := s.client

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
