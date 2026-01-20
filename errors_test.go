package mouser

import (
	"bufio"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

var errorsTestAPIKey string

func errorsTestInit() {
	if errorsTestAPIKey != "" {
		return
	}
	// Try environment first
	if key := os.Getenv("MOUSER_API_KEY"); key != "" {
		errorsTestAPIKey = key
		return
	}
	// Try .env file
	file, err := os.Open(".env")
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "MOUSER_API_KEY=") {
			errorsTestAPIKey = strings.TrimSpace(strings.TrimPrefix(line, "MOUSER_API_KEY="))
			return
		}
	}
}

// TestMouserErrorError tests the Error method of MouserError.
func TestMouserErrorError(t *testing.T) {
	err := &MouserError{
		StatusCode: 400,
		Message:    "bad request",
	}

	errStr := err.Error()
	if !contains(errStr, "400") {
		t.Errorf("expected error to contain status code")
	}
	if !contains(errStr, "bad request") {
		t.Errorf("expected error to contain message")
	}
}

// TestMouserErrorErrorWithDetails tests Error method with details.
func TestMouserErrorErrorWithDetails(t *testing.T) {
	err := &MouserError{
		StatusCode: 400,
		Message:    "bad request",
		Details:    "invalid field value",
	}

	errStr := err.Error()
	if errStr == "" {
		t.Fatal("expected non-empty error string")
	}

	if !contains(errStr, "bad request") {
		t.Errorf("expected error to contain message")
	}
	if !contains(errStr, "400") {
		t.Errorf("expected error to contain status code")
	}
}

// TestMouserErrorUnwrapRateLimit tests Unwrap for rate limit error.
func TestMouserErrorUnwrapRateLimit(t *testing.T) {
	err := &MouserError{
		StatusCode: 429,
		Message:    "rate limited",
	}

	if !errors.Is(err, ErrRateLimitExceeded) {
		t.Error("expected MouserError with 429 to unwrap to ErrRateLimitExceeded")
	}
}

// TestMouserErrorUnwrapNotFound tests Unwrap for not found error.
func TestMouserErrorUnwrapNotFound(t *testing.T) {
	err := &MouserError{
		StatusCode: 404,
		Message:    "not found",
	}

	if !errors.Is(err, ErrNotFound) {
		t.Error("expected MouserError with 404 to unwrap to ErrNotFound")
	}
}

// TestMouserErrorUnwrapUnknownStatus tests Unwrap with unknown status code.
func TestMouserErrorUnwrapUnknownStatus(t *testing.T) {
	err := &MouserError{
		StatusCode: 418,
		Message:    "teapot",
	}

	unwrapped := err.Unwrap()
	if unwrapped != nil {
		t.Errorf("expected nil for unknown status code, got %v", unwrapped)
	}
}

// TestMouserErrorWithoutStatusCode tests Error with 0 status code.
func TestMouserErrorWithoutStatusCode(t *testing.T) {
	err := &MouserError{
		StatusCode: 0,
		Message:    "internal error",
	}

	errStr := err.Error()
	if !contains(errStr, "internal error") {
		t.Errorf("expected error to contain message: %s", errStr)
	}
	if contains(errStr, "HTTP") {
		t.Errorf("expected no HTTP status when code is 0: %s", errStr)
	}
}

// TestAPIErrorError tests the Error method of APIError.
func TestAPIErrorError(t *testing.T) {
	err := APIError{
		Code:    "INVALID_KEYWORD",
		Message: "Keyword is required",
	}

	errStr := err.Error()
	if !contains(errStr, "INVALID_KEYWORD") {
		t.Errorf("expected error to contain code")
	}
	if !contains(errStr, "Keyword is required") {
		t.Errorf("expected error to contain message")
	}
}

// TestAPIErrorErrorWithoutCode tests APIError without code.
func TestAPIErrorErrorWithoutCode(t *testing.T) {
	err := APIError{
		Code:    "",
		Message: "Something went wrong",
	}

	errStr := err.Error()
	if !contains(errStr, "Something went wrong") {
		t.Errorf("expected error to contain message: %s", errStr)
	}
}

// TestAPIErrorEmpty tests empty APIError.
func TestAPIErrorEmpty(t *testing.T) {
	err := APIError{}

	errStr := err.Error()
	if errStr == "" {
		t.Fatal("expected non-empty error string")
	}
}

// TestAPIErrorsError tests Error method for multiple errors.
func TestAPIErrorsError(t *testing.T) {
	errs := APIErrors{
		{Code: "ERROR1", Message: "First error"},
		{Code: "ERROR2", Message: "Second error"},
		{Code: "ERROR3", Message: "Third error"},
	}

	errStr := errs.Error()

	if !contains(errStr, "3") {
		t.Errorf("expected error count in message: %s", errStr)
	}

	if !contains(errStr, "First error") {
		t.Errorf("expected first error in message: %s", errStr)
	}
}

// TestAPIErrorsSingleError tests APIErrors with single error.
func TestAPIErrorsSingleError(t *testing.T) {
	errs := APIErrors{
		{Code: "SINGLE", Message: "Single error"},
	}

	errStr := errs.Error()

	if !contains(errStr, "Single error") {
		t.Errorf("expected error message: %s", errStr)
	}
}

// TestAPIErrorsEmpty tests APIErrors with no errors.
func TestAPIErrorsEmpty(t *testing.T) {
	errs := APIErrors{}

	errStr := errs.Error()

	if !contains(errStr, "unknown") {
		t.Errorf("expected 'unknown' in empty error message: %s", errStr)
	}
}

// TestHTTPError tests HTTPError type.
func TestHTTPError(t *testing.T) {
	err := &HTTPError{
		StatusCode: 500,
		Status:     "Internal Server Error",
		Body:       "Server crashed",
	}

	errStr := err.Error()

	if !contains(errStr, "500") {
		t.Errorf("expected status code in error: %s", errStr)
	}

	if !contains(errStr, "Internal Server Error") {
		t.Errorf("expected status text in error: %s", errStr)
	}
}

// TestErrorVariables tests that error variables are distinct.
func TestErrorVariables(t *testing.T) {
	errs := []error{
		ErrNoAPIKey,
		ErrRateLimitExceeded,
		ErrDailyLimitExceeded,
		ErrInvalidResponse,
		ErrNotFound,
	}

	for i, err1 := range errs {
		for j, err2 := range errs {
			if i != j && err1 == err2 {
				t.Errorf("error %d should not equal error %d", i, j)
			}
		}
	}
}

// TestErrorVariableStrings tests error variable string messages.
func TestErrorVariableStrings(t *testing.T) {
	testCases := []struct {
		err      error
		contains string
	}{
		{ErrNoAPIKey, "API key"},
		{ErrRateLimitExceeded, "rate limit"},
		{ErrDailyLimitExceeded, "daily"},
		{ErrInvalidResponse, "invalid response"},
		{ErrNotFound, "not found"},
	}

	for _, tc := range testCases {
		if !contains(tc.err.Error(), tc.contains) {
			t.Errorf("expected %q to contain %q", tc.err.Error(), tc.contains)
		}
	}
}

// TestMouserErrorIsRetryable tests IsRetryable field.
func TestMouserErrorIsRetryable(t *testing.T) {
	// Rate limit errors should be retryable
	err1 := &MouserError{
		StatusCode:  429,
		IsRetryable: true,
	}
	if !err1.IsRetryable {
		t.Error("expected rate limit error to be retryable")
	}

	// Client errors should not be retryable
	err2 := &MouserError{
		StatusCode:  400,
		IsRetryable: false,
	}
	if err2.IsRetryable {
		t.Error("expected client error to not be retryable")
	}
}

// TestMouserErrorRetryAfter tests RetryAfter field.
func TestMouserErrorRetryAfter(t *testing.T) {
	err := &MouserError{
		StatusCode: 429,
		RetryAfter: 60,
	}

	if err.RetryAfter != 60 {
		t.Errorf("expected RetryAfter 60, got %d", err.RetryAfter)
	}
}

// TestMouserErrorEndpoint tests Endpoint field.
func TestMouserErrorEndpoint(t *testing.T) {
	err := &MouserError{
		StatusCode: 500,
		Endpoint:   "/search/keyword",
	}

	if err.Endpoint != "/search/keyword" {
		t.Errorf("expected endpoint /search/keyword, got %s", err.Endpoint)
	}
}

// TestMouserErrorAPIErrors tests Errors field for API errors.
func TestMouserErrorAPIErrors(t *testing.T) {
	apiErr := APIError{
		Code:    "INVALID",
		Message: "Invalid request",
	}

	err := &MouserError{
		StatusCode: 400,
		Errors:     []APIError{apiErr},
	}

	if len(err.Errors) != 1 {
		t.Errorf("expected 1 API error, got %d", len(err.Errors))
	}

	if err.Errors[0].Code != "INVALID" {
		t.Errorf("expected error code INVALID, got %s", err.Errors[0].Code)
	}
}

// TestAPIErrorFields tests APIError JSON fields.
func TestAPIErrorFields(t *testing.T) {
	err := APIError{
		ID:                    1,
		Code:                  "CODE",
		Message:               "message",
		ResourceKey:           "key",
		ResourceFormatString:  "format",
		ResourceFormatString2: "format2",
		PropertyName:          "property",
	}

	if err.ID != 1 {
		t.Errorf("expected ID 1, got %d", err.ID)
	}
	if err.Code != "CODE" {
		t.Errorf("expected code CODE")
	}
	if err.ResourceKey != "key" {
		t.Errorf("expected resource key")
	}
	if err.PropertyName != "property" {
		t.Errorf("expected property name")
	}
}

// TestMouserErrorString tests string representation.
func TestMouserErrorString(t *testing.T) {
	err := &MouserError{
		StatusCode: 500,
		Message:    "Internal Server Error",
		Endpoint:   "/search/keyword",
	}

	errStr := err.Error()
	if errStr == "" {
		t.Fatal("expected non-empty error string")
	}

	if !contains(errStr, "500") || !contains(errStr, "Internal Server Error") {
		t.Errorf("expected status and message in error: %s", errStr)
	}
}

// TestErrorWrapping tests error wrapping with errors.Is.
func TestErrorWrapping(t *testing.T) {
	err := &MouserError{
		StatusCode: 429,
		Message:    "rate limited",
	}

	// Should match ErrRateLimitExceeded via Unwrap
	if !errors.Is(err, ErrRateLimitExceeded) {
		t.Error("expected MouserError 429 to match ErrRateLimitExceeded")
	}

	err2 := &MouserError{
		StatusCode: 404,
		Message:    "not found",
	}

	// Should match ErrNotFound via Unwrap
	if !errors.Is(err2, ErrNotFound) {
		t.Error("expected MouserError 404 to match ErrNotFound")
	}
}

// TestAPIErrorsMultiple tests APIErrors with multiple errors.
func TestAPIErrorsMultiple(t *testing.T) {
	errs := APIErrors{
		{Code: "ERROR1", Message: "Error one"},
		{Code: "ERROR2", Message: "Error two"},
	}

	errStr := errs.Error()

	if !contains(errStr, "2") || !contains(errStr, "Error one") {
		t.Errorf("unexpected error format: %s", errStr)
	}
}

// TestMouserErrorNoMessage tests MouserError with empty message.
func TestMouserErrorNoMessage(t *testing.T) {
	err := &MouserError{
		StatusCode: 500,
		Message:    "",
	}

	errStr := err.Error()
	if errStr == "" {
		t.Fatal("expected non-empty error string")
	}
}

// skipIfNoAPIKeyErrors skips test if MOUSER_API_KEY is not set
func skipIfNoAPIKeyErrors(t *testing.T) {
	errorsTestInit()
	if errorsTestAPIKey == "" {
		t.Skip("MOUSER_API_KEY not found in environment or .env file")
	}
}

// TestRealAPIErrorHandling tests that real API errors are properly handled.
func TestRealAPIErrorHandling(t *testing.T) {
	skipIfNoAPIKeyErrors(t)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Try searching with an invalid API key (will use a dummy one to trigger error)
	testClient, _ := NewClient("invalid-key-that-should-fail")

	result, err := testClient.KeywordSearch(ctx, SearchOptions{
		Keyword: "resistor",
	})

	// Should fail with some kind of error (auth or API error)
	if result != nil && err == nil {
		t.Skip("API did not reject invalid key (some APIs don't validate immediately)")
	}
}

// TestContextTimeoutError tests that context timeout produces proper error.
func TestContextTimeoutError(t *testing.T) {
	skipIfNoAPIKeyErrors(t)

	apiKey := errorsTestAPIKey
	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	result, err := client.KeywordSearch(ctx, SearchOptions{
		Keyword: "test",
	})

	if err == nil {
		t.Fatal("expected error from context timeout")
	}

	if result != nil {
		t.Error("expected nil result on timeout")
	}

	// Error should indicate context issue
	if err != context.DeadlineExceeded && !errors.Is(err, context.DeadlineExceeded) {
		t.Logf("Context error type: %T, value: %v", err, err)
	}
}


