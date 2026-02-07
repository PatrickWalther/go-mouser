package mouser

import (
	"context"
	"net/url"
)

// GetOrderHistoryByDateFilter retrieves order history filtered by a predefined date filter.
func (c *Client) GetOrderHistoryByDateFilter(ctx context.Context, dateFilter DateFilterType) (*OrderHistoryResponse, error) {
	query := url.Values{}
	query.Set("dateFilter", string(dateFilter))

	var resp OrderHistoryResponse
	if err := c.doRequestWithQuery(ctx, "GET", "/orderhistory/ByDateFilter", query, nil, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}

// GetOrderHistoryByDateRange retrieves order history within a specific date range.
// Dates should be in the format expected by the Mouser API (e.g. "2025-01-01").
func (c *Client) GetOrderHistoryByDateRange(ctx context.Context, startDate, endDate string) (*OrderHistoryResponse, error) {
	query := url.Values{}
	query.Set("startDate", startDate)
	query.Set("endDate", endDate)

	var resp OrderHistoryResponse
	if err := c.doRequestWithQuery(ctx, "GET", "/orderhistory/ByDateRange", query, nil, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}

// GetOrderBySalesOrderNumber retrieves order details by sales order number.
func (c *Client) GetOrderBySalesOrderNumber(ctx context.Context, salesOrderNumber string) (*OrderDetailResponse, error) {
	query := url.Values{}
	query.Set("salesOrderNumber", salesOrderNumber)

	var resp OrderDetailResponse
	if err := c.doRequestWithQuery(ctx, "GET", "/orderhistory/salesOrderNumber", query, nil, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}

// GetOrderByWebOrderNumber retrieves order details by web order number.
func (c *Client) GetOrderByWebOrderNumber(ctx context.Context, webOrderNumber string) (*OrderDetailResponse, error) {
	query := url.Values{}
	query.Set("webOrderNumber", webOrderNumber)

	var resp OrderDetailResponse
	if err := c.doRequestWithQuery(ctx, "GET", "/orderhistory/webOrderNumber", query, nil, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, APIErrors(resp.Errors)
	}

	return &resp, nil
}
