# 🌐 HTTP Endpoints

This document provides a comprehensive reference for Qwiklip's HTTP API endpoints, including request/response formats and usage examples.

## 🎯 **Base URL**

```
http://localhost:8080
```

All endpoints are relative to the base URL. The default port is `8080` but can be configured via the `PORT` environment variable.

## 📋 **Available Endpoints**

### **1. Health Check**

**Endpoint:** `GET /health`

**Purpose:** Check if the server is running and healthy.

**Request:**
```http
GET /health HTTP/1.1
Host: localhost:8080
```

**Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-14T06:48:30Z"
}
```

**Response (500 Internal Server Error):**
```json
{
  "error": {
    "type": "internal_error",
    "message": "Service unavailable"
  },
  "timestamp": "2025-01-14T06:48:30Z"
}
```

**Usage:**
```bash
curl http://localhost:8080/health
```

### **2. Server Information**

**Endpoint:** `GET /`

**Purpose:** Get information about the server and available endpoints.

**Request:**
```http
GET / HTTP/1.1
Host: localhost:8080
```

**Response (200 OK):**
```html
<!DOCTYPE html>
<html>
<head>
    <title>Qwiklip</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .example { background: #f5f5f5; padding: 15px; border-radius: 5px; margin: 10px 0; }
        code { background: #e0e0e0; padding: 2px 4px; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Qwiklip</h1>
        <p>Access Instagram videos through Qwiklip.</p>

        <div class="example">
            <strong>Instagram URL:</strong><br>
            <code>https://www.instagram.com/reel/ABC123/</code><br><br>
            <strong>Qwiklip URL:</strong><br>
            <code>http://localhost:8080/reel/ABC123/</code>
        </div>
    </div>
</body>
</html>
```

### **3. Instagram Reel Streaming**

**Endpoint:** `GET /reel/{shortcode}/`

**Purpose:** Stream an Instagram reel video.

**Parameters:**
- `shortcode`: The Instagram reel shortcode (e.g., `ABC123`)

**Request:**
```http
GET /reel/ABC123/ HTTP/1.1
Host: localhost:8080
Accept: */*
User-Agent: Mozilla/5.0 (compatible)
Range: bytes=0-1023
```

**Response (200 OK):**
```http
HTTP/1.1 200 OK
Content-Type: video/mp4
Content-Length: 5242880
Accept-Ranges: bytes
Content-Range: bytes 0-5242879/5242880

[Binary video data]
```

**Response (206 Partial Content):**
```http
HTTP/1.1 206 Partial Content
Content-Type: video/mp4
Content-Length: 1024
Accept-Ranges: bytes
Content-Range: bytes 0-1023/5242880

[Partial binary video data]
```



## 🔍 **Request/Response Details**

### **HTTP Methods**

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `GET` | `/health` | Health check |
| `GET` | `/` | Server information |
| `GET` | `/reel/{shortcode}/` | Stream reel video |

### **Content Types**

**Request Content Types:**
- All endpoints accept: `*/*`

**Response Content Types:**
- `/health`: `application/json`
- `/`: `text/html`
- Video endpoints: `video/mp4`

### **HTTP Status Codes**

| Code | Meaning | When Returned |
|------|---------|---------------|
| `200` | OK | Successful video streaming |
| `206` | Partial Content | Range request fulfilled |
| `400` | Bad Request | Invalid URL or shortcode |
| `404` | Not Found | Content not found or private |
| `415` | Unsupported Media Type | Non-video content |
| `429` | Too Many Requests | Rate limited |
| `500` | Internal Server Error | Server error |
| `502` | Bad Gateway | Instagram API error |

### **HTTP Headers**

#### **Request Headers (Supported)**

| Header | Purpose | Example |
|--------|---------|---------|
| `User-Agent` | Client identification | `Mozilla/5.0 ...` |
| `Accept` | Accepted content types | `*/*` |
| `Range` | Partial content request | `bytes=0-1023` |
| `Accept-Language` | Language preference | `en-US,en;q=0.9` |

#### **Response Headers**

| Header | Purpose | Example |
|--------|---------|---------|
| `Content-Type` | Response content type | `video/mp4` |
| `Content-Length` | Response size in bytes | `5242880` |
| `Accept-Ranges` | Range request support | `bytes` |
| `Content-Range` | Partial content info | `bytes 0-1023/5242880` |

## 📝 **Usage Examples**

### **Basic Video Streaming**

```bash
# Stream a reel
curl -O "http://localhost:8080/reel/ABC123/"
```

### **Partial Content (Resuming Downloads)**

```bash
# Request first 1MB of video
curl -H "Range: bytes=0-1048575" \
     "http://localhost:8080/reel/ABC123/" \
     -o video_part.mp4

# Request next 1MB
curl -H "Range: bytes=1048576-2097151" \
     "http://localhost:8080/reel/ABC123/" \
     -o video_part2.mp4
```

### **With Custom Headers**

```bash
# Simulate browser request
curl -H "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" \
     -H "Accept: */*" \
     "http://localhost:8080/reel/ABC123/"
```

### **Error Handling**

```bash
# Test invalid shortcode
curl "http://localhost:8080/reel/invalid/"

# Response:
{
  "error": {
    "type": "invalid_url",
    "message": "invalid Instagram URL: http://localhost:8080/reel/invalid/",
    "details": {
      "url": "http://localhost:8080/reel/invalid/"
    }
  },
  "timestamp": "2025-01-14T06:48:30Z"
}
```

## 🔄 **URL Patterns**

### **Supported Instagram URL Formats**

Qwiklip supports the same URL patterns as Instagram:

| Instagram URL | Qwiklip URL |
|---------------|-----------|
| `https://www.instagram.com/reel/ABC123/` | `http://localhost:8080/reel/ABC123/` |

### **Shortcode Requirements**

- **Format**: Alphanumeric characters (A-Z, a-z, 0-9)
- **Length**: Typically 10-12 characters
- **Case**: Case-insensitive (converted to uppercase internally)
- **Validation**: Must be valid Instagram shortcode format

## ⚡ **Performance Considerations**

### **Streaming Optimization**

- Videos are streamed directly from Instagram's CDN
- No local storage or caching
- Supports HTTP range requests for seeking
- Connection pooling for optimal performance

### **Rate Limiting**

- Subject to Instagram's rate limits
- Implements backoff strategies
- Returns `429 Too Many Requests` when rate limited

### **Timeout Handling**

- Default request timeout: 30 seconds
- Configurable via `INSTAGRAM_TIMEOUT` environment variable
- Automatic cleanup of timed-out connections

## 🛡️ **Security Features**

### **Input Validation**

- URL format validation
- Shortcode format validation
- Header sanitization

### **Error Information Disclosure**

- Sensitive information not exposed in error messages
- Generic error messages for security
- Detailed errors only in debug mode

### **CORS Support**

- Cross-origin requests supported
- Appropriate CORS headers set
- Configurable origin policies

## 🧪 **Testing Endpoints**

### **Health Check Testing**

```bash
# Test health endpoint
curl -w "@curl-format.txt" -o /dev/null -s "http://localhost:8080/health"

# curl-format.txt
     time_namelookup:  %{time_namelookup}\n
        time_connect:  %{time_connect}\n
     time_appconnect:  %{time_appconnect}\n
    time_pretransfer:  %{time_pretransfer}\n
       time_redirect:  %{time_redirect}\n
  time_starttransfer:  %{time_starttransfer}\n
                     ----------\n
          time_total:  %{time_total}\n
```

### **Video Endpoint Testing**

```bash
# Test video endpoint (get headers only)
curl -I "http://localhost:8080/reel/ABC123/"

# Test with range request
curl -H "Range: bytes=0-99" \
     -o /dev/null \
     "http://localhost:8080/reel/ABC123/"
```

### **Load Testing**

```bash
# Simple load test
for i in {1..10}; do
  curl -s "http://localhost:8080/health" &
done
wait
```

## 📊 **Monitoring and Metrics**

### **Health Metrics**

```json
{
  "status": "healthy",
  "timestamp": "2025-01-14T06:48:30Z",
  "uptime": "1h30m45s",
  "active_connections": 5,
  "version": "1.0.0"
}
```

### **Performance Metrics**

- Response time distribution
- Success/failure rates
- Bandwidth usage
- Error rates by type

## 🚀 **Advanced Usage**

### **Programmatic Access**

```python
import requests

# Python example - Stream video for viewing
def stream_instagram_video(shortcode):
    url = f"http://localhost:8080/reel/{shortcode}/"
    response = requests.get(url, stream=True)

    # Process stream for viewing (not saving to disk)
    for chunk in response.iter_content(chunk_size=8192):
        # Process chunk for video player or display
        process_video_chunk(chunk)
```

```javascript
// JavaScript example - Embed in web player
async function loadVideoInPlayer(shortcode) {
    const videoElement = document.getElementById('video-player');
    const response = await fetch(`http://localhost:8080/reel/${shortcode}/`);

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error.message);
    }

    // Create blob URL for video element
    const blob = await response.blob();
    const videoUrl = URL.createObjectURL(blob);
    videoElement.src = videoUrl;
}
```

### **Browser Integration**

```bash
# Direct browser access (recommended)
open "http://localhost:8080/reel/ABC123/"

# curl for inspection/testing
curl -I "http://localhost:8080/reel/ABC123/"

# Stream to video player
curl "http://localhost:8080/reel/ABC123/" | mpv -
```

## 📚 **Troubleshooting**

### **Common Issues**

#### **1. 400 Bad Request**

```json
{
  "error": {
    "type": "invalid_url",
    "message": "invalid Instagram URL: http://localhost:8080/reel/invalid/",
    "details": {
      "url": "http://localhost:8080/reel/invalid/"
    }
  }
}
```

**Solution:** Check the shortcode format and ensure it's a valid Instagram URL.

#### **2. 404 Not Found**

```json
{
  "error": {
    "type": "not_found",
    "message": "Content not found",
    "details": {
      "shortcode": "ABC123"
    }
  }
}
```

**Solution:** The content may be private, deleted, or the shortcode is incorrect.

#### **3. 429 Too Many Requests**

```json
{
  "error": {
    "type": "rate_limited",
    "message": "rate limited by Instagram",
    "details": {
      "retry_after": "60"
    }
  }
}
```

**Solution:** Wait before making another request. Consider using different IP addresses.

#### **4. 502 Bad Gateway**

```json
{
  "error": {
    "type": "extraction",
    "message": "failed to extract media info for shortcode: ABC123",
    "details": {
      "shortcode": "ABC123"
    }
  }
}
```

**Solution:** Instagram may have changed their page structure. Check for updates or try again later.

## 📖 **API Evolution**

### **Versioning**

- Current API version: `v1` (implicit)
- Breaking changes will be versioned as `v2`, `v3`, etc.
- Deprecated endpoints will be marked in documentation

### **Backwards Compatibility**

- Existing endpoints will remain functional
- New optional parameters may be added
- Response formats will maintain structure
- New error types may be introduced

## 📚 **Further Reading**

- [HTTP Status Codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
- [HTTP Range Requests](https://developer.mozilla.org/en-US/docs/Web/HTTP/Range_requests)
- [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)

---

**Next**: Learn about [response formats](./responses.md) and data structures.
