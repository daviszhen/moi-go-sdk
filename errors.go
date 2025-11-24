package sdk

import (
	"errors"
	"fmt"
)

var (
	// ErrBaseURLRequired indicates that NewRawClient was called without a base URL.
	//
	// The baseURL parameter must be a non-empty string containing a valid URL
	// with scheme and host (e.g., "https://api.example.com").
	ErrBaseURLRequired = errors.New("sdk: baseURL is required")

	// ErrAPIKeyRequired indicates that NewRawClient was called without an API key.
	//
	// The apiKey parameter must be a non-empty string used for authentication.
	ErrAPIKeyRequired = errors.New("sdk: apiKey is required")

	// ErrNilRequest indicates that a required request payload was nil.
	//
	// All API methods require a non-nil request parameter. If you need to pass
	// an empty request, use an empty struct literal (e.g., &CatalogListRequest{}).
	ErrNilRequest = errors.New("sdk: request payload cannot be nil")
)

// APIError captures an application-level error returned by the catalog service envelope.
//
// APIError represents business logic errors returned by the server, such as
// validation errors, resource not found, permission denied, etc.
//
// Example:
//
//	resp, err := client.CreateCatalog(ctx, req)
//	if err != nil {
//		if apiErr, ok := err.(*sdk.APIError); ok {
//			fmt.Printf("API Error: %s (code: %s, request_id: %s)\n",
//				apiErr.Message, apiErr.Code, apiErr.RequestID)
//		}
//	}
type APIError struct {
	// Code is the error code returned by the server (e.g., "ErrInternal").
	Code string

	// Message is the human-readable error message.
	Message string

	// RequestID is the unique request identifier for tracking purposes.
	RequestID string

	// HTTPStatus is the HTTP status code of the response.
	HTTPStatus int
}

func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("catalog service error: code=%s msg=%s request_id=%s status=%d", e.Code, e.Message, e.RequestID, e.HTTPStatus)
}

// HTTPError represents a non-2xx HTTP response that occurred before the SDK could parse the envelope.
//
// HTTPError represents network-level errors or server errors that occur before
// the response can be parsed as a valid API response envelope.
//
// Example:
//
//	resp, err := client.CreateCatalog(ctx, req)
//	if err != nil {
//		if httpErr, ok := err.(*sdk.HTTPError); ok {
//			fmt.Printf("HTTP Error: %d\n", httpErr.StatusCode)
//			fmt.Printf("Response Body: %s\n", string(httpErr.Body))
//		}
//	}
type HTTPError struct {
	// StatusCode is the HTTP status code (e.g., 401, 404, 500).
	StatusCode int

	// Body contains the raw response body, if available.
	Body []byte
}

func (e *HTTPError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if len(e.Body) == 0 {
		return fmt.Sprintf("http error: status=%d", e.StatusCode)
	}
	return fmt.Sprintf("http error: status=%d body=%s", e.StatusCode, string(e.Body))
}
