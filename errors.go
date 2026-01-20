package mouser

import (
	"errors"
	"fmt"
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
)

// MouserError represents a structured error from the Mouser API.
type MouserError struct {
	StatusCode   int        // HTTP status code
	Message      string     // Error message
	Details      string     // Additional details
	Errors       []APIError // Mouser domain errors
	Endpoint     string     // API endpoint that failed
	RetryAfter   int        // Seconds to wait before retrying (from Retry-After header)
	IsRetryable  bool       // Whether this error is retryable
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
	if e.StatusCode == 429 {
		return ErrRateLimitExceeded
	}
	if e.StatusCode == 404 {
		return ErrNotFound
	}
	return nil
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

// HTTPError represents an HTTP-level error from the API.
// Deprecated: Use MouserError instead for richer error information.
type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
}

// Error implements the error interface for HTTPError.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("mouser: HTTP error %d: %s", e.StatusCode, e.Status)
}
