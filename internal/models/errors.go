package models

import "fmt"

// Error types for better error handling
type ErrorType string

const (
	ErrorTypeInvalidURL     ErrorType = "invalid_url"
	ErrorTypeNetwork        ErrorType = "network"
	ErrorTypeExtraction     ErrorType = "extraction"
	ErrorTypeParsing        ErrorType = "parsing"
	ErrorTypeNotFound       ErrorType = "not_found"
	ErrorTypeUnsupported    ErrorType = "unsupported"
	ErrorTypeAuthentication ErrorType = "authentication"
	ErrorTypeRateLimited    ErrorType = "rate_limited"
)

// AppError represents a custom application error
type AppError struct {
	Type    ErrorType              `json:"type"`
	Message string                 `json:"message"`
	Cause   error                  `json:"-"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// HTTPStatusCode returns the appropriate HTTP status code for the error
func (e *AppError) HTTPStatusCode() int {
	switch e.Type {
	case ErrorTypeInvalidURL, ErrorTypeNetwork, ErrorTypeExtraction, ErrorTypeParsing:
		return 400
	case ErrorTypeNotFound:
		return 404
	case ErrorTypeUnsupported:
		return 415
	case ErrorTypeAuthentication:
		return 401
	case ErrorTypeRateLimited:
		return 429
	default:
		return 500
	}
}

// NewInvalidURLError creates a new invalid URL error
func NewInvalidURLError(url string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeInvalidURL,
		Message: fmt.Sprintf("invalid Instagram URL: %s", url),
		Cause:   cause,
		Details: map[string]interface{}{"url": url},
	}
}

// NewNetworkError creates a new network error
func NewNetworkError(operation string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeNetwork,
		Message: fmt.Sprintf("network error during %s", operation),
		Cause:   cause,
		Details: map[string]interface{}{"operation": operation},
	}
}

// NewExtractionError creates a new extraction error
func NewExtractionError(shortcode string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeExtraction,
		Message: fmt.Sprintf("failed to extract media info for shortcode: %s", shortcode),
		Cause:   cause,
		Details: map[string]interface{}{"shortcode": shortcode},
	}
}

// NewParsingError creates a new parsing error
func NewParsingError(dataType string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeParsing,
		Message: fmt.Sprintf("failed to parse %s", dataType),
		Cause:   cause,
		Details: map[string]interface{}{"data_type": dataType},
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Details: map[string]interface{}{"resource": resource},
	}
}

// NewUnsupportedError creates a new unsupported content error
func NewUnsupportedError(contentType string) *AppError {
	return &AppError{
		Type:    ErrorTypeUnsupported,
		Message: fmt.Sprintf("unsupported content type: %s", contentType),
		Details: map[string]interface{}{"content_type": contentType},
	}
}

// NewRateLimitedError creates a new rate limited error
func NewRateLimitedError(retryAfter string) *AppError {
	return &AppError{
		Type:    ErrorTypeRateLimited,
		Message: "rate limited by Instagram",
		Details: map[string]interface{}{"retry_after": retryAfter},
	}
}
