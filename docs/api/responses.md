# 📤 Response Formats

This document details the response formats for Qwiklip's HTTP API, including success responses, error responses, and data structures.

## 📋 **Response Overview**

Qwiklip returns responses in different formats depending on the endpoint:

- **JSON**: For health checks and error responses
- **HTML**: For informational pages
- **Binary**: For video content streaming

## 🎯 **Success Responses**

### **1. Health Check Response**

**Endpoint:** `GET /health`

**Content-Type:** `application/json`

**Status Code:** `200 OK`

**Response Body:**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-14T06:48:30.894+05:30",
  "uptime": "1h30m45s",
  "active_connections": 5,
  "version": "1.0.0"
}
```

**Field Descriptions:**

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `status` | string | Service health status | `"healthy"` |
| `timestamp` | string | ISO 8601 timestamp | `"2025-01-14T06:48:30.894+05:30"` |
| `uptime` | string | Service uptime duration | `"1h30m45s"` |
| `active_connections` | number | Current active connections | `5` |
| `version` | string | Service version | `"1.0.0"` |

### **2. Video Streaming Response**

**Endpoints:** `GET /reel/{shortcode}/`

**Content-Type:** `video/mp4`

**Status Codes:** `200 OK` or `206 Partial Content`

**Response Headers:**
```
Content-Type: video/mp4
Content-Length: 5242880
Accept-Ranges: bytes
Content-Range: bytes 0-5242879/5242880
```

**Response Body:** Binary video data

**Header Descriptions:**

| Header | Description | Example |
|--------|-------------|---------|
| `Content-Type` | MIME type of the content | `video/mp4` |
| `Content-Length` | Total size in bytes | `5242880` |
| `Accept-Ranges` | Range request support | `bytes` |
| `Content-Range` | Range of bytes returned | `bytes 0-5242879/5242880` |

### **3. Information Page Response**

**Endpoint:** `GET /`

**Content-Type:** `text/html`

**Status Code:** `200 OK`

**Response Body:** HTML page with usage information

## 🚨 **Error Responses**

### **Error Response Structure**

All error responses follow this JSON structure:

```json
{
  "error": {
    "type": "error_type",
    "message": "Human-readable error message",
    "details": {
      "field1": "value1",
      "field2": "value2"
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

### **1. Invalid URL Error**

**Status Code:** `400 Bad Request`

**Response:**
```json
{
  "error": {
    "type": "invalid_url",
    "message": "invalid Instagram URL: http://localhost:8080/reel/invalid/",
    "details": {
      "url": "http://localhost:8080/reel/invalid/"
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

### **2. Content Not Found Error**

**Status Code:** `404 Not Found`

**Response:**
```json
{
  "error": {
    "type": "not_found",
    "message": "Content not found",
    "details": {
      "shortcode": "ABC123",
      "reason": "private_or_deleted"
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

### **3. Rate Limited Error**

**Status Code:** `429 Too Many Requests`

**Response:**
```json
{
  "error": {
    "type": "rate_limited",
    "message": "rate limited by Instagram",
    "details": {
      "retry_after": "60",
      "limit": "100",
      "remaining": "0"
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

### **4. Extraction Failed Error**

**Status Code:** `502 Bad Gateway`

**Response:**
```json
{
  "error": {
    "type": "extraction",
    "message": "failed to extract media info for shortcode: ABC123",
    "details": {
      "shortcode": "ABC123",
      "strategies_tried": 4,
      "last_error": "JSON parsing failed"
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

### **5. Network Error**

**Status Code:** `502 Bad Gateway`

**Response:**
```json
{
  "error": {
    "type": "network",
    "message": "network error during content fetch",
    "details": {
      "operation": "content_fetch",
      "timeout": "30s",
      "retries": 3
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

### **6. Internal Server Error**

**Status Code:** `500 Internal Server Error`

**Response:**
```json
{
  "error": {
    "type": "internal_error",
    "message": "An unexpected error occurred",
    "details": {
      "request_id": "req_123456",
      "component": "instagram_client"
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

## 📊 **Error Types Reference**

| Error Type | HTTP Status | Description | Common Causes |
|------------|-------------|-------------|---------------|
| `invalid_url` | 400 | Invalid Instagram URL format | Malformed URL, wrong domain |
| `not_found` | 404 | Content not available | Private content, deleted post |
| `unsupported` | 415 | Unsupported content type | Non-video content, stories |
| `rate_limited` | 429 | Rate limit exceeded | Too many requests |
| `extraction` | 502 | Failed to extract video | Instagram layout changed |
| `parsing` | 502 | Failed to parse response | Invalid JSON, unexpected format |
| `network` | 502 | Network communication error | Timeout, DNS failure |
| `authentication` | 401 | Authentication required | Private content access |
| `internal_error` | 500 | Unexpected server error | Bugs, configuration issues |

## 🔄 **Response Format Standards**

### **JSON Response Standards**

- **Content-Type:** Always `application/json`
- **Character Encoding:** UTF-8
- **Date Format:** ISO 8601 with timezone
- **Field Naming:** snake_case for consistency
- **Error Structure:** Consistent error object format

### **HTTP Header Standards**

- **Content-Type:** Accurate MIME types
- **Content-Length:** Present when known
- **Accept-Ranges:** `bytes` for video content
- **Cache-Control:** Appropriate caching directives
- **CORS Headers:** Cross-origin support

## 📋 **Data Structures**

### **InstagramMediaInfo**

```go
type InstagramMediaInfo struct {
    VideoURL     string `json:"videoUrl"`
    FileName     string `json:"fileName"`
    ThumbnailURL string `json:"thumbnailUrl,omitempty"`
    Caption      string `json:"caption,omitempty"`
    Username     string `json:"username,omitempty"`
}
```

**Field Descriptions:**

| Field | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| `videoUrl` | string | ✅ | Direct video URL | `"https://cdn.instagram.com/..."` |
| `fileName` | string | ✅ | Suggested filename | `"reel_ABC123.mp4"` |
| `thumbnailUrl` | string | ❌ | Thumbnail image URL | `"https://cdn.instagram.com/..."` |
| `caption` | string | ❌ | Post caption text | `"Amazing sunset! 🌅"` |
| `username` | string | ❌ | Author's username | `"instagram_user"` |

### **Error Details Structure**

```json
{
  "error": {
    "type": "string",
    "message": "string",
    "details": {
      "additional_field": "value"
    }
  },
  "timestamp": "string"
}
```

**Error Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `error.type` | string | Error classification |
| `error.message` | string | Human-readable message |
| `error.details` | object | Additional error context |
| `timestamp` | string | Error occurrence time |

## 🎯 **Response Examples**

### **Complete Video Request Flow**

```bash
# Request video
curl -v "http://localhost:8080/reel/ABC123/"

# Response headers
< HTTP/1.1 200 OK
< Content-Type: video/mp4
< Content-Length: 5242880
< Accept-Ranges: bytes
< Date: Mon, 14 Jan 2025 06:48:30 GMT

# Binary video data follows...
```

### **Range Request Example**

```bash
# Request partial content
curl -H "Range: bytes=0-1023" \
     -v "http://localhost:8080/reel/ABC123/"

# Response
< HTTP/1.1 206 Partial Content
< Content-Type: video/mp4
< Content-Length: 1024
< Content-Range: bytes 0-1023/5242880
< Accept-Ranges: bytes

# First 1024 bytes of video...
```

### **Error Handling Example**

```bash
# Request invalid content
curl "http://localhost:8080/reel/invalid-shortcode/"

# Response
{
  "error": {
    "type": "invalid_url",
    "message": "invalid Instagram URL: http://localhost:8080/reel/invalid-shortcode/",
    "details": {
      "url": "http://localhost:8080/reel/invalid-shortcode/"
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

## 📊 **Response Metrics**

### **Success Response Metrics**

- **Response Time:** < 500ms for health checks
- **Throughput:** Variable based on video size and network
- **Availability:** 99.9% uptime target
- **Error Rate:** < 1% for valid requests

### **Error Response Metrics**

- **Error Response Time:** < 100ms
- **Detailed Error Rate:** Debug mode only
- **Error Classification:** 90% of errors properly categorized

## 🧪 **Testing Response Formats**

### **Unit Tests for Responses**

```go
func TestHealthResponse(t *testing.T) {
    server := setupTestServer()

    req := httptest.NewRequest("GET", "/health", nil)
    w := httptest.NewRecorder()

    server.handleHealthCheck(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    assert.Equal(t, "healthy", response["status"])
    assert.NotEmpty(t, response["timestamp"])
}
```

### **Integration Tests**

```go
func TestVideoStreamingResponse(t *testing.T) {
    server := setupTestServer()

    req := httptest.NewRequest("GET", "/reel/test123/", nil)
    w := httptest.NewRecorder()

    server.handleReel(w, req)

    // Check headers
    assert.Equal(t, "video/mp4", w.Header().Get("Content-Type"))
    assert.Equal(t, "bytes", w.Header().Get("Accept-Ranges"))

    // Check status
    assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusPartialContent)
}
```

### **Error Response Tests**

```go
func TestErrorResponseFormat(t *testing.T) {
    testCases := []struct {
        errorType   string
        statusCode  int
        expectError bool
    }{
        {"invalid_url", 400, true},
        {"not_found", 404, true},
        {"rate_limited", 429, true},
        {"extraction", 502, true},
    }

    for _, tc := range testCases {
        // Test error response format
        // Verify JSON structure, required fields, etc.
    }
}
```

## 🔄 **Response Evolution**

### **Backwards Compatibility**

- **Additive Changes:** New optional fields can be added
- **Format Preservation:** Existing field types and names maintained
- **Deprecation Notices:** Deprecated fields marked in documentation
- **Version Headers:** API version in response headers (future)

### **Future Enhancements**

```json
{
  "data": {
    "video_url": "https://...",
    "metadata": {
      "duration": 30,
      "resolution": "1080p",
      "bitrate": "2500kbps"
    }
  },
  "version": "1.1.0",
  "processing_time_ms": 150
}
```

## 📚 **Further Reading**

- [JSON API Specification](https://jsonapi.org/)
- [HTTP Status Codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
- [MIME Types](https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types)

---

**Next**: Learn about [error codes](./errors.md) and their detailed meanings.
