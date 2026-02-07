package mouser

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrNoAPIKey is returned when no API key is provided.
	ErrNoAPIKey = errors.New("mouser: API key is required")

	// ErrRateLimitExceeded is returned when the rate limit is exceeded.
	ErrRateLimitExceeded = errors.New("mouser: rate limit exceeded")

	// ErrDailyLimitExceeded is returned when the daily request limit is exceeded.
	ErrDailyLimitExceeded = errors.New("mouser: daily request limit exceeded")

	// ErrInvalidResponse is returned when the API returns an invalid response.
	ErrInvalidResponse = errors.New("mouser: invalid response from API")

	// ErrNotFound is returned when a part is not found.
	ErrNotFound = errors.New("mouser: part not found")

	// ErrUnauthorized is returned when the API key is invalid or missing.
	ErrUnauthorized = errors.New("mouser: unauthorized")

	// ErrForbidden is returned when access is denied.
	ErrForbidden = errors.New("mouser: forbidden")

	// ErrInvalidRequest is returned when the request is malformed.
	ErrInvalidRequest = errors.New("mouser: invalid request")

	// ErrServerError is returned when the server returns a 5xx error.
	ErrServerError = errors.New("mouser: server error")
)

// MouserError represents a structured error from the Mouser API.
type MouserError struct {
	StatusCode  int        // HTTP status code
	Message     string     // Error message
	Details     string     // Additional details
	Errors      []APIError // Mouser domain errors
	Endpoint    string     // API endpoint that failed
	RetryAfter  int        // Seconds to wait before retrying (from Retry-After header)
	IsRetryable bool       // Whether this error is retryable
}

// Error implements the error interface.
func (e *MouserError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("mouser: HTTP %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("mouser: %s", e.Message)
}

// Unwrap returns the underlying error for errors.Is compatibility.
func (e *MouserError) Unwrap() error {
	switch e.StatusCode {
	case 400:
		return ErrInvalidRequest
	case 401:
		return ErrUnauthorized
	case 403:
		return ErrForbidden
	case 404:
		return ErrNotFound
	case 429:
		return ErrRateLimitExceeded
	default:
		if e.StatusCode >= 500 {
			return ErrServerError
		}
		return nil
	}
}

// APIError represents an error returned by the Mouser API in the response body.
type APIError struct {
	ID                    int    `json:"Id"`
	Code                  string `json:"Code"`
	Message               string `json:"Message"`
	ResourceKey           string `json:"ResourceKey"`
	ResourceFormatString  string `json:"ResourceFormatString"`
	ResourceFormatString2 string `json:"ResourceFormatString2"`
	PropertyName          string `json:"PropertyName"`
}

// Error implements the error interface for APIError.
func (e APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("mouser API error [%s]: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("mouser API error: %s", e.Message)
}

// RateLimitError represents a rate limit error with details about the limit.
type RateLimitError struct {
	Limit     int       // The rate limit that was exceeded
	Remaining int       // Remaining requests (typically 0)
	ResetAt   time.Time // When the limit resets
	Type      string    // "minute" or "day"
}

// Error implements the error interface.
func (e *RateLimitError) Error() string {
	return fmt.Sprintf("mouser: %s rate limit exceeded (limit: %d, resets at: %s)",
		e.Type, e.Limit, e.ResetAt.Format(time.RFC3339))
}

// Unwrap returns the underlying sentinel error.
func (e *RateLimitError) Unwrap() error {
	if e.Type == "day" {
		return ErrDailyLimitExceeded
	}
	return ErrRateLimitExceeded
}

// APIErrors represents a collection of API errors.
type APIErrors []APIError

// Error implements the error interface for APIErrors.
func (e APIErrors) Error() string {
	if len(e) == 0 {
		return "mouser: unknown API error"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return fmt.Sprintf("mouser: %d API errors: %s (and %d more)", len(e), e[0].Message, len(e)-1)
}
