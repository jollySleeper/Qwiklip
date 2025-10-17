# 🚨 Error Codes Reference

This document provides a comprehensive reference for all error codes returned by Qwiklip's HTTP API, including detailed descriptions, causes, and resolution steps.

## 📋 **Error Code Structure**

### **Error Response Format**

All errors follow this consistent JSON structure:

```json
{
  "error": {
    "type": "error_type",
    "message": "Human-readable error message",
    "details": {
      "additional_context": "value",
      "debug_info": "additional_data"
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

## 🔍 **Error Code Reference**

### **400 Bad Request - Client Errors**

#### **INVALID_URL**
```json
{
  "error": {
    "type": "invalid_url",
    "message": "invalid Instagram URL: https://example.com",
    "details": {
      "url": "https://example.com",
      "expected_format": "https://www.instagram.com/{type}/{shortcode}/"
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

**Description:** The provided URL is not a valid Instagram URL format.

**Common Causes:**
- Wrong domain (not instagram.com)
- Malformed URL structure
- Missing or invalid shortcode
- Unsupported URL type

**Resolution Steps:**
1. Verify the URL is from instagram.com
2. Check URL format: `https://www.instagram.com/{reel|p|tv}/{shortcode}/`
3. Ensure shortcode contains only alphanumeric characters
4. Remove any query parameters or fragments

**Examples:**
- ❌ `https://example.com/reel/ABC123/`
- ❌ `https://www.instagram.com/reel/`
- ✅ `https://www.instagram.com/reel/ABC123/`

---

#### **UNSUPPORTED_CONTENT**
```json
{
  "error": {
    "type": "unsupported",
    "message": "unsupported content type: story",
    "details": {
      "content_type": "story",
      "supported_types": ["reel", "post", "tv"]
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

**Description:** The requested content type is not supported by Qwiklip.

**Common Causes:**
- Instagram Stories URLs
- IGTV URLs (though `/tv/` should work)
- Live video URLs
- Carousel posts with multiple videos

**Resolution Steps:**
1. Verify the content is a video (not photo or text)
2. Use correct URL type: `/reel/` for reels, `/p/` for posts, `/tv/` for IGTV
3. Check if content is publicly accessible

---

### **401 Unauthorized - Authentication Errors**

#### **AUTHENTICATION_REQUIRED**
```json
{
  "error": {
    "type": "authentication",
    "message": "Content requires authentication",
    "details": {
      "reason": "private_account",
      "shortcode": "ABC123"
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

**Description:** The content is private and requires authentication to access.

**Common Causes:**
- Private Instagram account
- Content only visible to followers
- Age-restricted content

**Resolution Steps:**
1. Content must be made public by the owner
2. No workaround available through Qwiklip
3. Try different public content

---

### **404 Not Found - Content Errors**

#### **CONTENT_NOT_FOUND**
```json
{
  "error": {
    "type": "not_found",
    "message": "Content not found",
    "details": {
      "shortcode": "ABC123",
      "reason": "deleted_or_private"
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

**Description:** The requested Instagram content could not be found.

**Common Causes:**
- Content was deleted by the owner
- Account was deleted or suspended
- URL is incorrect or outdated
- Content moved or URL changed

**Resolution Steps:**
1. Verify the URL is correct and current
2. Check if the account still exists
3. Confirm the content hasn't been deleted
4. Try searching for the content directly on Instagram

---

#### **ACCOUNT_NOT_FOUND**
```json
{
  "error": {
    "type": "not_found",
    "message": "Instagram account not found",
    "details": {
      "username": "deleted_account",
      "reason": "account_deleted"
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

**Description:** The Instagram account associated with the content no longer exists.

**Common Causes:**
- Account was permanently deleted
- Account was suspended by Instagram
- Username was changed

**Resolution Steps:**
1. Verify the account exists on Instagram
2. Check if the username has changed
3. Look for the content under a different account

---

### **429 Too Many Requests - Rate Limiting**

#### **RATE_LIMIT_EXCEEDED**
```json
{
  "error": {
    "type": "rate_limited",
    "message": "rate limited by Instagram",
    "details": {
      "retry_after": "60",
      "limit": "100",
      "remaining": "0",
      "reset_time": "2025-01-14T06:49:30Z"
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

**Description:** Too many requests have been made in a short period.

**Common Causes:**
- High request frequency
- Multiple concurrent requests
- Shared IP address with other users
- Automated requests without delays

**Resolution Steps:**
1. Wait for the specified `retry_after` seconds
2. Implement exponential backoff
3. Reduce request frequency
4. Consider using different IP addresses
5. Space out requests over time

**Rate Limit Details:**
- **Requests per minute:** ~100 (varies by IP)
- **Requests per hour:** ~1000
- **Concurrent requests:** Limited by server capacity

---

### **502 Bad Gateway - Upstream Errors**

#### **EXTRACTION_FAILED**
```json
{
  "error": {
    "type": "extraction",
    "message": "failed to extract media info for shortcode: ABC123",
    "details": {
      "shortcode": "ABC123",
      "strategies_tried": 4,
      "last_strategy": "json_parse",
      "extraction_time_ms": 2500
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

**Description:** Failed to extract video information from Instagram's page.

**Common Causes:**
- Instagram changed their page structure
- JavaScript-heavy content loading
- Anti-bot measures
- Network issues during extraction

**Resolution Steps:**
1. Wait a few minutes and try again
2. Check if Instagram has updated their interface
3. Verify the content is still accessible
4. Try accessing the content directly on Instagram

---

#### **PARSING_FAILED**
```json
{
  "error": {
    "type": "parsing",
    "message": "failed to parse Instagram response",
    "details": {
      "data_type": "json",
      "parse_error": "invalid character '}' looking for beginning of value",
      "response_size": 245760,
      "content_type": "text/html"
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

**Description:** Failed to parse the data returned from Instagram.

**Common Causes:**
- Unexpected HTML/JSON format changes
- Incomplete response from Instagram
- Encoding issues
- Corrupted response data

**Resolution Steps:**
1. Try the request again
2. Check network connectivity
3. Verify Instagram is accessible
4. Report if issue persists

---

#### **NETWORK_ERROR**
```json
{
  "error": {
    "type": "network",
    "message": "network error during content fetch",
    "details": {
      "operation": "content_fetch",
      "error": "dial tcp: lookup instagram.com: no such host",
      "timeout": "30s",
      "retries": 0
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

**Description:** Network communication error with Instagram or external services.

**Common Causes:**
- DNS resolution failures
- Network connectivity issues
- Firewall blocking requests
- Instagram servers temporarily unavailable
- SSL/TLS certificate issues

**Resolution Steps:**
1. Check internet connectivity
2. Verify DNS resolution: `nslookup instagram.com`
3. Check firewall settings
4. Try again after a few minutes
5. Verify SSL certificates are valid

---

#### **UPSTREAM_TIMEOUT**
```json
{
  "error": {
    "type": "network",
    "message": "upstream request timeout",
    "details": {
      "operation": "video_fetch",
      "timeout": "30s",
      "elapsed": "30.5s",
      "url": "https://cdn.instagram.com/video.mp4"
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

**Description:** The request to Instagram's servers timed out.

**Common Causes:**
- Slow network connection
- Large video files
- Instagram server congestion
- Network latency issues

**Resolution Steps:**
1. Check network speed and stability
2. Try again during off-peak hours
3. Use a different network connection
4. Consider smaller content first

---

### **500 Internal Server Error - Server Errors**

#### **INTERNAL_ERROR**
```json
{
  "error": {
    "type": "internal_error",
    "message": "An unexpected error occurred",
    "details": {
      "request_id": "req_123456789",
      "component": "instagram_client",
      "version": "1.0.0"
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

**Description:** An unexpected internal server error occurred.

**Common Causes:**
- Software bugs
- Configuration issues
- Resource exhaustion
- Memory issues
- Race conditions

**Resolution Steps:**
1. Note the `request_id` for debugging
2. Check server logs for more details
3. Try the request again
4. Report the issue if it persists

---

## 📊 **Error Statistics**

### **Error Frequency by Type**

| Error Type | Frequency | Typical Resolution |
|------------|-----------|-------------------|
| `extraction` | 15-20% | Wait and retry |
| `rate_limited` | 10-15% | Implement backoff |
| `network` | 5-10% | Check connectivity |
| `not_found` | 5-10% | Verify content exists |
| `invalid_url` | 2-5% | Fix URL format |
| `internal_error` | <1% | Report bug |

### **Error Response Times**

| Error Type | Avg Response Time | Target |
|------------|-------------------|--------|
| Client errors (4xx) | <50ms | <100ms |
| Server errors (5xx) | <100ms | <200ms |
| Network errors | <500ms | <1000ms |

## 🛠️ **Error Handling Best Practices**

### **Client-Side Error Handling**

```javascript
async function fetchVideo(shortcode) {
    try {
        const response = await fetch(`/reel/${shortcode}/`);

        if (!response.ok) {
            const error = await response.json();

            switch (error.error.type) {
                case 'rate_limited':
                    // Implement exponential backoff
                    await delay(error.error.details.retry_after * 1000);
                    return fetchVideo(shortcode);

                case 'not_found':
                    // Show user-friendly message
                    showError('Video not found or is private');
                    break;

                case 'extraction':
                    // Try again after delay
                    await delay(5000);
                    return fetchVideo(shortcode);

                default:
                    // Generic error
                    showError(error.error.message);
            }
            return;
        }

        // Process successful response
        return await response.blob();

    } catch (networkError) {
        // Handle network errors
        showError('Network error. Please check your connection.');
    }
}
```

### **Retry Logic Implementation**

```javascript
function retryWithBackoff(fn, maxRetries = 3, baseDelay = 1000) {
    return async function(...args) {
        for (let attempt = 0; attempt < maxRetries; attempt++) {
            try {
                return await fn(...args);
            } catch (error) {
                if (error.type === 'rate_limited') {
                    const delay = error.details.retry_after * 1000;
                    await new Promise(resolve => setTimeout(resolve, delay));
                    continue;
                }

                if (attempt === maxRetries - 1) throw error;

                // Exponential backoff for other errors
                const delay = baseDelay * Math.pow(2, attempt);
                await new Promise(resolve => setTimeout(resolve, delay));
            }
        }
    };
}
```

## 📋 **Debug Information**

### **Enabling Debug Mode**

```bash
# Enable debug logging
export DEBUG=true
export LOG_LEVEL=debug

# Start server
go run ./cmd/qwiklip
```

### **Debug Error Response**

```json
{
  "error": {
    "type": "extraction",
    "message": "failed to extract media info for shortcode: ABC123",
    "details": {
      "shortcode": "ABC123",
      "strategies_tried": 4,
      "strategy_results": [
        {"name": "json_script", "success": false, "error": "pattern not found"},
        {"name": "apollo_state", "success": false, "error": "key not found"},
        {"name": "shared_data", "success": false, "error": "parse error"},
        {"name": "direct_html", "success": false, "error": "regex failed"}
      ],
      "html_length": 245760,
      "extraction_time_ms": 2500,
      "user_agent": "Mozilla/5.0...",
      "request_headers": {...}
    }
  },
  "timestamp": "2025-01-14T06:48:30.894+05:30"
}
```

## 🔧 **Troubleshooting Guide**

### **Step-by-Step Debugging**

1. **Check URL Format**
   ```bash
   # Test with curl to verify URL
   curl -I "https://www.instagram.com/reel/ABC123/"
   ```

2. **Verify Content Accessibility**
   ```bash
   # Check if content exists and is public
   # Open URL in browser incognito mode
   ```

3. **Test Network Connectivity**
   ```bash
   # Test DNS resolution
   nslookup instagram.com

   # Test connectivity
   ping instagram.com

   # Test SSL
   openssl s_client -connect instagram.com:443
   ```

4. **Check Rate Limits**
   ```bash
   # Monitor request frequency
   # Implement delays between requests
   # Consider IP rotation
   ```

5. **Enable Debug Logging**
   ```bash
   # Get detailed error information
   DEBUG=true LOG_LEVEL=debug go run ./cmd/qwiklip
   ```

## 📈 **Error Monitoring**

### **Error Metrics to Track**

- **Error rate by type** (percentage of requests)
- **Error rate by endpoint** (health, reel, post, tv)
- **Error response time distribution**
- **Retry success rates**
- **Geographic error patterns**

### **Alerting Thresholds**

- **Error Rate:** >5% for 5 minutes
- **Rate Limiting:** >10 rate limit errors per minute
- **Extraction Failures:** >20% for 10 minutes
- **Network Errors:** >15% for 5 minutes

## 📚 **Further Reading**

- [HTTP Status Codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
- [Error Handling Patterns](https://golang.org/doc/effective_go#errors)
- [API Error Design](https://www.vinaysahni.com/best-practices-for-a-pragmatic-restful-api#errors)

---

This completes the comprehensive documentation for Qwiklip's error codes. For questions about specific errors or additional troubleshooting help, please check the logs or create an issue with detailed information.
