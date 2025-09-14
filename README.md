# Instagram Proxy Server

A Go-based web server that mirrors Instagram video URLs, allowing you to access Instagram reels and videos through a simple proxy interface. This server uses the same sophisticated extraction strategies as the original TypeScript version to reliably fetch video URLs from Instagram's changing page structure.

## 🚀 Features

- **One-to-One URL Mirroring**: Access videos using the same path structure as Instagram
- **Multiple Extraction Strategies**: Robust fallback mechanisms to handle Instagram's frequent API changes
- **Direct Video Streaming**: Efficiently streams video content without downloading to disk
- **Range Request Support**: Supports partial content requests for better performance
- **Health Check Endpoint**: Built-in monitoring capabilities
- **Cross-Platform**: Runs on any platform that supports Go
- **Docker Support**: Easy containerized deployment

## 📋 Requirements

- Go 1.19 or later
- Internet connection

## 🏃‍♂️ Quick Start

### Using Go directly:

```bash
# Clone or download the project
cd instagram-proxy-go

# Run the server
go run .

# Or build and run
go build -o instagram-proxy
./instagram-proxy
```

The server will start on port 8080 by default.

### Using Docker:

```bash
# Build the Docker image
docker build -t instagram-proxy .

# Run the container
docker run -p 8080:8080 instagram-proxy
```

## 📖 Usage

### URL Format

The server mirrors Instagram's URL structure exactly:

**Instagram URL:**
```
https://www.instagram.com/reel/ABC123XYZ/
```

**Proxy URL:**
```
http://localhost:8080/reel/ABC123XYZ/
```

### Supported URL Types

- **Reels**: `/reel/{shortcode}/`
- **Posts**: `/p/{shortcode}/`
- **TV**: `/tv/{shortcode}/`

### Examples

```bash
# Access a reel
curl http://localhost:8080/reel/C2Z4BcJJ0LU/

# Access a post
curl http://localhost:8080/p/C2Z4BcJJ0LU/

# Health check
curl http://localhost:8080/health
```

## ⚙️ Configuration

### Environment Variables

- `PORT`: Server port (default: `8080`)

```bash
# Run on a different port
PORT=3000 go run .
```

## 🏗️ Architecture

The server implements multiple layers of extraction strategies:

### 1. **URL Format Attempts**
- Desktop user agent with `/p/` format
- Desktop user agent with `/reel/` format
- Mobile user agent with `/p/` format
- Mobile user agent with API parameters

### 2. **JSON Data Extraction**
- `window.__additionalDataLoaded` pattern
- `window._sharedData` pattern
- `window.__APOLLO_STATE__` pattern
- Direct JSON response parsing

### 3. **Video URL Discovery**
- GraphQL structure parsing
- Apollo State structure parsing
- Direct video URL extraction from HTML
- Multiple fallback mechanisms

### 4. **Video Streaming**
- Efficient byte-range streaming
- Proper HTTP headers for video content
- Progress tracking and error handling

## 🔧 Development

### Project Structure

```
instagram-proxy-go/
├── main.go           # Web server and routing
├── instagram.go      # Instagram API client and extraction logic
├── go.mod           # Go module definition
├── Dockerfile       # Docker build configuration
└── README.md        # This file
```

### Building

```bash
# Build for current platform
go build -o instagram-proxy

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o instagram-proxy-linux
GOOS=darwin GOARCH=amd64 go build -o instagram-proxy-mac
GOOS=windows GOARCH=amd64 go build -o instagram-proxy-windows.exe
```

## 🐳 Docker Deployment

### Build and Run

```bash
# Build image
docker build -t instagram-proxy .

# Run container
docker run -p 8080:8080 instagram-proxy

# Run with custom port
docker run -p 3000:8080 -e PORT=3000 instagram-proxy
```

### Docker Compose

Create a `docker-compose.yml`:

```yaml
version: '3.8'
services:
  instagram-proxy:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
    restart: unless-stopped
```

## 🔍 Health Monitoring

The server includes a health check endpoint:

```bash
curl http://localhost:8080/health
```

Response:
```json
{"status":"healthy","timestamp":"2025-01-13T10:30:00Z"}
```

## 🛠️ Troubleshooting

### Common Issues

1. **"Could not extract video URL"**
   - Instagram may have changed their page structure
   - The content might be private or deleted
   - Try a different URL format

2. **Connection timeouts**
   - Check your internet connection
   - Instagram may be rate-limiting requests
   - Try again after a few minutes

3. **Server not starting**
   - Ensure port 8080 is not in use
   - Check if another process is using the port
   - Try a different port with `PORT=3000`

### Debug Logging

The server provides detailed logging for troubleshooting:

```
🚀 Instagram Proxy Server starting on port 8080
📺 Access videos at: http://localhost:8080/reel/{shortcode}/
Processing request: /reel/ABC123XYZ/ -> https://www.instagram.com/reel/ABC123XYZ/
Extracted shortcode: ABC123XYZ
Trying URL format: https://www.instagram.com/p/ABC123XYZ/
Successfully fetched content from: https://www.instagram.com/reel/ABC123XYZ/
Successfully extracted JSON data using pattern
Found video URL in graphql structure
Streamed 1 MB for ABC123XYZ.mp4
Successfully streamed video: ABC123XYZ.mp4 (2097152 bytes)
```

## 📄 License

This project is open source. Feel free to use, modify, and distribute.

## 🤝 Contributing

Contributions are welcome! Please feel free to:

- Report bugs
- Suggest features
- Submit pull requests
- Improve documentation

## ⚠️ Disclaimer

This tool is for educational and personal use only. Please respect Instagram's Terms of Service and be mindful of rate limiting. The server does not store any content locally and only proxies requests to Instagram's servers.
