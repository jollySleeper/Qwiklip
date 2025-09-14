# 🖼️ Qwiklip

A Go-based web server that allows you to access reels through a simple proxy interface.

> ⚠️ **Disclaimer**
> This tool is for educational and personal use only. Please respect Instagram's Terms of Service and be mindful of rate limiting. The server does not store any content locally and only proxies requests to Instagram's servers.

**🎉 Fun Fact :** **Qwiklip** means **QuickClip**!
If you remove the 'c' from QuickClip, you get `qwiklip` - as I can't C. Just kidding, it's just a clever play on words that captures the essence of fast, efficient way of watching video clips.

## 📖 Table of Contents

- [✨ Features](#features)
- [🚀 Installation](#installation)
- [🛠️ Usage](#usage)
- [⚙️ Configuration](#configuration)
- [🔧 Development](#development)
- [🐛 Bugs or Requests](#bugs-or-requests)
- [🤝 Contributing](#contributing)
- [📄 License](#license)
- [🙏 Acknowledgments](#acknowledgments)

## ✨ Features

- **One-to-One URL Mirroring**: Access videos using the same path structure as Instagram
- **Multiple Extraction Strategies**: Robust fallback mechanisms to handle Instagram's frequent API changes
- **Direct Video Streaming**: Efficiently streams video content without downloading to disk
- **Range Request Support**: Supports partial content requests for better performance
- **Health Check Endpoint**: Built-in monitoring capabilities
- **Cross-Platform**: Runs on any platform that supports Go
- **Docker Support**: Easy containerized deployment

## 🚀 Installation

> Please note that you should have [Go](https://golang.org) 1.25 or later installed on your system.

### Prerequisites

- Go 1.25 or later
- Internet connection

### Quick Start

#### Using Go directly:

```bash
# Clone or download the project
cd qwiklip

# Run the server
go run .

# Or build and run
go build -o qwiklip
./qwiklip
```

The server will start on port 8080 by default.

#### Using Docker:

```bash
# Build the Docker image
docker build -t qwiklip .

# Run the container
docker run -p 8080:8080 qwiklip

# Run with custom port
docker run -p 3000:8080 -e PORT=3000 qwiklip
```

##### Docker Compose (Advanced)

Create a `docker-compose.yml`:

```yaml
version: '3.8'
services:
  qwiklip:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
    restart: unless-stopped
```

## 🛠️ Usage

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

## 🔧 Development

### Project Structure

```
qwiklip/
├── main.go          # Web server and routing
├── instagram.go     # Instagram API client and extraction logic
├── go.mod           # Go module definition
├── Dockerfile       # Docker build configuration
└── README.md        # This file
```

### Architecture

The server implements multiple layers of extraction strategies:

#### 1. **URL Format Attempts**
- Desktop user agent with `/p/` format
- Desktop user agent with `/reel/` format
- Mobile user agent with `/p/` format
- Mobile user agent with API parameters

#### 2. **JSON Data Extraction**
- `window.__additionalDataLoaded` pattern
- `window._sharedData` pattern
- `window.__APOLLO_STATE__` pattern
- Direct JSON response parsing

#### 3. **Video URL Discovery**
- GraphQL structure parsing
- Apollo State structure parsing
- Direct video URL extraction from HTML
- Multiple fallback mechanisms

#### 4. **Video Streaming**
- Efficient byte-range streaming
- Proper HTTP headers for video content
- Progress tracking and error handling

### Building

```bash
# Build for current platform
go build -o qwiklip

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o qwiklip-linux
GOOS=darwin GOARCH=amd64 go build -o qwiklip-mac
GOOS=windows GOARCH=amd64 go build -o qwiklip-windows.exe
```

### Health Monitoring

The server includes a health check endpoint:

```bash
curl http://localhost:8080/health
```

Response:
```json
{"status":"healthy","timestamp":"2025-01-13T10:30:00Z"}
```

## 🐛 Bugs or Requests

### Troubleshooting

#### Common Issues

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

### Reporting Bugs

If you encounter any problem(s) feel free to open an [issue](https://github.com/rmali/jollySleeper/issues/new).

### Feature Requests

If you feel the project is missing a feature, please raise an [issue](https://github.com/jollySleeper/qwiklip/issues/new) with `FeatureRequest` as heading.

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature/YourFeature`).
3. Make your changes and commit them (`git commit -m 'Add some feature'`).
4. Push to the branch (`git push origin feature/YourFeature`).
5. Open a pull request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

This is free and open-source software. You are free to use, modify, and distribute it for personal use only.

## 🙏 Acknowledgments

This project was built with the help of:

- [sayedmahmoud266/instagram-reel-downloader](https://github.com/sayedmahmoud266/instagram-reel-downloader) for the original project inspiration
- [Cursor](https://cursor.sh) for providing an excellent AI-powered development environment
- [grok-code](https://x.ai/grok) LLM model for assisting in the development and implementation
