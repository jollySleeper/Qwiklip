# 🚨 Error Handling

This document explains Qwiklip's error handling system, featuring custom error types, proper HTTP status mapping, and structured error responses.

## 🎯 **Overview**

Qwiklip implements a comprehensive error handling system that provides:

- **Custom Error Types** - Domain-specific error classification
- **HTTP Status Mapping** - Proper HTTP status codes for different error types
- **Structured Error Responses** - Consistent JSON error responses
- **Error Context** - Detailed error information for debugging
- **Error Wrapping** - Preserving error chains with additional context

## 🏗️ **Architecture**

### **Component Structure**

```
internal/models/
└── errors.go    # Custom error types and handling
```

### **Error Type Hierarchy**

```go
// Base error interface
type error interface {
    Error() string
}

// Custom error type
type AppError struct {
    Type    ErrorType       `json:"type"`
    Message string         `json:"message"`
    Cause   error          `json:"-"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

## 📋 **Error Types**

### **Error Classification**

```go
type ErrorType string

const (
    ErrorTypeInvalidURL      ErrorType = "invalid_url"
    ErrorTypeNetwork         ErrorType = "network"
    ErrorTypeExtraction      ErrorType = "extraction"
    ErrorTypeParsing         ErrorType = "parsing"
    ErrorTypeNotFound        ErrorType = "not_found"
    ErrorTypeUnsupported     ErrorType = "unsupported"
    ErrorTypeAuthentication  ErrorType = "authentication"
    ErrorTypeRateLimited     ErrorType = "rate_limited"
)
```

### **HTTP Status Mapping**

```go
func (e *AppError) HTTPStatusCode() int {
    switch e.Type {
    case ErrorTypeInvalidURL:
        return 400  // Bad Request
    case ErrorTypeNotFound:
        return 404  // Not Found
    case ErrorTypeUnsupported:
        return 415  // Unsupported Media Type
    case ErrorTypeAuthentication:
        return 401  // Unauthorized
    case ErrorTypeRateLimited:
        return 429  // Too Many Requests
    case ErrorTypeNetwork, ErrorTypeExtraction, ErrorTypeParsing:
        return 502  // Bad Gateway
    default:
        return 500  // Internal Server Error
    }
}
```

## 🔧 **Error Creation Functions**

### **Constructor Functions**

```go
// URL validation errors
func NewInvalidURLError(url string, cause error) *AppError {
    return &AppError{
        Type:    ErrorTypeInvalidURL,
        Message: fmt.Sprintf("invalid Instagram URL: %s", url),
        Cause:   cause,
        Details: map[string]interface{}{"url": url},
    }
}

// Network operation errors
func NewNetworkError(operation string, cause error) *AppError {
    return &AppError{
        Type:    ErrorTypeNetwork,
        Message: fmt.Sprintf("network error during %s", operation),
        Cause:   cause,
        Details: map[string]interface{}{"operation": operation},
    }
}

// Content extraction errors
func NewExtractionError(shortcode string, cause error) *AppError {
    return &AppError{
        Type:    ErrorTypeExtraction,
        Message: fmt.Sprintf("failed to extract media info for shortcode: %s", shortcode),
        Cause:   cause,
        Details: map[string]interface{}{"shortcode": shortcode},
    }
}

// Data parsing errors
func NewParsingError(dataType string, cause error) *AppError {
    return &AppError{
        Type:    ErrorTypeParsing,
        Message: fmt.Sprintf("failed to parse %s", dataType),
        Cause:   cause,
        Details: map[string]interface{}{"data_type": dataType},
    }
}
```

## 🎯 **Error Usage Examples**

### **Instagram Client Errors**

```go
// URL extraction failure
func (c *Client) ExtractShortcode(urlStr string) (string, error) {
    if !strings.Contains(urlStr, "instagram.com") {
        return "", NewInvalidURLError(urlStr, nil)
    }

    parsedURL, err := url.Parse(urlStr)
    if err != nil {
        return "", NewInvalidURLError(urlStr, err)
    }

    // ... extraction logic ...

    return "", NewInvalidURLError(urlStr, fmt.Errorf("could not extract shortcode"))
}
```

### **Network Operation Errors**

```go
func (c *Client) fetchContent(url string) ([]byte, error) {
    resp, err := c.httpClient.Get(url)
    if err != nil {
        return nil, NewNetworkError("content fetch", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return nil, NewNetworkError(fmt.Sprintf("HTTP %d", resp.StatusCode),
            fmt.Errorf("server returned status %d", resp.StatusCode))
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, NewNetworkError("response reading", err)
    }

    return body, nil
}
```

### **Extraction Strategy Errors**

```go
func (c *Client) GetMediaInfo(urlStr string) (*models.InstagramMediaInfo, error) {
    shortcode, err := c.ExtractShortcode(urlStr)
    if err != nil {
        return nil, fmt.Errorf("failed to extract shortcode: %w", err)
    }

    // Try multiple strategies
    for _, strategy := range c.getExtractionStrategies() {
        mediaInfo, err := c.tryStrategy(strategy, shortcode)
        if err == nil {
            return mediaInfo, nil
        }
        // Log failed strategy but continue
        c.logger.Debug("Strategy failed", "strategy", strategy.name, "error", err)
    }

    // All strategies failed
    return nil, NewExtractionError(shortcode, fmt.Errorf("all extraction strategies failed"))
}
```

## 🛡️ **Error Handling Patterns**

### **Error Wrapping**

```go
// Preserve error context with additional information
func (c *Client) processContent(content string) error {
    data, err := c.parseJSON(content)
    if err != nil {
        return fmt.Errorf("failed to parse Instagram content: %w", err)
    }

    mediaInfo, err := c.extractMediaInfo(data)
    if err != nil {
        return fmt.Errorf("failed to extract media info from parsed data: %w", err)
    }

    return nil
}
```

### **Error Type Checking**

```go
func handleInstagramError(err error) {
    var appErr *models.AppError
    if errors.As(err, &appErr) {
        // Handle custom application error
        switch appErr.Type {
        case models.ErrorTypeInvalidURL:
            // Handle invalid URL
            log.Printf("Invalid URL provided: %s", appErr.Details["url"])
        case models.ErrorTypeRateLimited:
            // Handle rate limiting
            retryAfter := appErr.Details["retry_after"]
            log.Printf("Rate limited, retry after: %v", retryAfter)
        case models.ErrorTypeNotFound:
            // Handle not found
            log.Printf("Content not found: %s", appErr.Message)
        default:
            // Handle other application errors
            log.Printf("Application error: %s", appErr.Message)
        }
    } else {
        // Handle standard errors
        log.Printf("Unexpected error: %v", err)
    }
}
```

### **HTTP Error Responses**

```go
func (s *Server) handleError(w http.ResponseWriter, err error) {
    var appErr *models.AppError
    if errors.As(err, &appErr) {
        // Send structured error response
        s.sendErrorResponse(w, appErr.HTTPStatusCode(), appErr)
        return
    }

    // Generic error for unexpected errors
    s.sendErrorResponse(w, http.StatusInternalServerError,
        &models.AppError{
            Type:    "internal_error",
            Message: "An unexpected error occurred",
            Details: map[string]interface{}{"original_error": err.Error()},
        })
}

func (s *Server) sendErrorResponse(w http.ResponseWriter, statusCode int, appErr *models.AppError) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)

    response := map[string]interface{}{
        "error": map[string]interface{}{
            "type":    appErr.Type,
            "message": appErr.Message,
            "details": appErr.Details,
        },
        "timestamp": time.Now().Format(time.RFC3339),
    }

    json.NewEncoder(w).Encode(response)
}
```

## 📊 **Error Response Formats**

### **Client Error Response (400)**

```json
{
  "error": {
    "type": "invalid_url",
    "message": "invalid Instagram URL: https://example.com",
    "details": {
      "url": "https://example.com"
    }
  },
  "timestamp": "2025-01-14T06:48:30Z"
}
```

### **Server Error Response (502)**

```json
{
  "error": {
    "type": "extraction",
    "message": "failed to extract media info for shortcode: ABC123",
    "details": {
      "shortcode": "ABC123"
    }
  },
  "timestamp": "2025-01-14T06:48:30Z"
}
```

### **Rate Limit Response (429)**

```json
{
  "error": {
    "type": "rate_limited",
    "message": "rate limited by Instagram",
    "details": {
      "retry_after": "60"
    }
  },
  "timestamp": "2025-01-14T06:48:30Z"
}
```

## 🧪 **Testing Error Handling**

### **Unit Tests for Error Types**

```go
func TestNewInvalidURLError(t *testing.T) {
    url := "https://invalid.com"
    cause := fmt.Errorf("test error")
    err := NewInvalidURLError(url, cause)

    assert.Equal(t, ErrorTypeInvalidURL, err.Type)
    assert.Contains(t, err.Message, url)
    assert.Equal(t, cause, err.Cause)
    assert.Equal(t, 400, err.HTTPStatusCode())
    assert.Equal(t, url, err.Details["url"])
}

func TestHTTPStatusCodeMapping(t *testing.T) {
    testCases := []struct {
        errorType   ErrorType
        expectedCode int
    }{
        {ErrorTypeInvalidURL, 400},
        {ErrorTypeNotFound, 404},
        {ErrorTypeRateLimited, 429},
        {ErrorTypeNetwork, 502},
        {"unknown", 500},
    }

    for _, tc := range testCases {
        err := &AppError{Type: tc.errorType}
        assert.Equal(t, tc.expectedCode, err.HTTPStatusCode())
    }
}
```

### **Integration Tests**

```go
func TestErrorHandlingIntegration(t *testing.T) {
    // Setup test server
    server := setupTestServer()

    // Test invalid URL
    req := httptest.NewRequest("GET", "/invalid", nil)
    w := httptest.NewRecorder()

    server.handleReel(w, req)

    assert.Equal(t, http.StatusBadRequest, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    errorInfo := response["error"].(map[string]interface{})
    assert.Equal(t, "invalid_url", errorInfo["type"])
}
```

## 📈 **Error Monitoring**

### **Error Metrics**

```go
type ErrorMetrics struct {
    ErrorsByType      map[ErrorType]int64
    ErrorsByHTTPCode  map[int]int64
    TotalErrors       int64
    RecentErrors      []*AppError
}

func (em *ErrorMetrics) RecordError(err *AppError) {
    em.ErrorsByType[err.Type]++
    em.ErrorsByHTTPCode[err.HTTPStatusCode()]++
    em.TotalErrors++

    // Keep recent errors for debugging
    em.RecentErrors = append(em.RecentErrors, err)
    if len(em.RecentErrors) > 100 {
        em.RecentErrors = em.RecentErrors[1:]
    }
}
```

### **Error Logging**

```go
func (s *Server) logError(err error, r *http.Request) {
    var appErr *AppError
    if errors.As(err, &appErr) {
        s.logger.Error("Request failed",
            "error_type", appErr.Type,
            "error_message", appErr.Message,
            "http_method", r.Method,
            "request_path", r.URL.Path,
            "client_ip", getClientIP(r),
            "user_agent", r.UserAgent(),
            "details", appErr.Details)
    } else {
        s.logger.Error("Unexpected error",
            "error", err.Error(),
            "http_method", r.Method,
            "request_path", r.URL.Path,
            "client_ip", getClientIP(r))
    }
}
```

## 🚀 **Advanced Error Features**

### **Error Recovery**

```go
func (s *Server) recoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                s.logger.Error("Panic recovered",
                    "panic", err,
                    "stack", string(debug.Stack()),
                    "request_path", r.URL.Path,
                    "client_ip", getClientIP(r))

                s.sendErrorResponse(w, http.StatusInternalServerError,
                    NewInternalError("unexpected server error", nil))
            }
        }()
        next(w, r)
    }
}
```

### **Error Aggregation**

```go
type ErrorAggregator struct {
    errors chan *AppError
    mu     sync.RWMutex
    counts map[string]int64
}

func (ea *ErrorAggregator) Start() {
    go func() {
        ticker := time.NewTicker(1 * time.Minute)
        defer ticker.Stop()

        for {
            select {
            case err := <-ea.errors:
                ea.mu.Lock()
                ea.counts[string(err.Type)]++
                ea.mu.Unlock()

            case <-ticker.C:
                ea.reportAggregatedErrors()
            }
        }
    }()
}

func (ea *ErrorAggregator) reportAggregatedErrors() {
    ea.mu.RLock()
    defer ea.mu.RUnlock()

    for errorType, count := range ea.counts {
        if count > 0 {
            log.Printf("Error type %s occurred %d times in last minute", errorType, count)
        }
    }
}
```

## 📚 **Best Practices**

### **1. Error Type Consistency**

```go
// ✅ Good: Consistent error types
func validateURL(url string) error {
    if !isValidInstagramURL(url) {
        return NewInvalidURLError(url, nil)
    }
    return nil
}

// ❌ Bad: Inconsistent error messages
func validateURL(url string) error {
    if !isValidInstagramURL(url) {
        return fmt.Errorf("bad url: %s", url) // Inconsistent format
    }
    return nil
}
```

### **2. Error Context Preservation**

```go
// ✅ Good: Preserve error chain
func processRequest(r *http.Request) error {
    data, err := parseRequest(r)
    if err != nil {
        return fmt.Errorf("failed to process request: %w", err)
    }
    return nil
}

// ❌ Bad: Lose original error context
func processRequest(r *http.Request) error {
    data, err := parseRequest(r)
    if err != nil {
        return fmt.Errorf("request processing failed") // Lost original error
    }
    return nil
}
```

### **3. Appropriate HTTP Status Codes**

```go
// ✅ Good: Appropriate status codes
switch err := err.(type) {
case *InvalidURLError:
    return http.StatusBadRequest
case *NotFoundError:
    return http.StatusNotFound
case *RateLimitError:
    return http.StatusTooManyRequests
default:
    return http.StatusInternalServerError
}

// ❌ Bad: Wrong status codes
if err != nil {
    return http.StatusInternalServerError // Always 500
}
```

## 📚 **Further Reading**

- [Go Error Handling Best Practices](https://golang.org/doc/effective_go#errors)
- [Error Wrapping in Go 1.13+](https://golang.org/pkg/errors/)
- [HTTP Status Codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)

---

**Next**: Learn about the [logging system](./logging.md) implementation.
