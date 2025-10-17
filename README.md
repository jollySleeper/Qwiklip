# <img src="web/static/svg/favicon.svg" width="32" height="32" style="vertical-align: middle; margin-right: 12px;" alt="Qwiklip"> Qwiklip

A privacy-focused Go-based web server that provides an alternative frontend for watching Instagram reels without tracking.

> ⚠️ **Disclaimer**
> This tool is for educational and personal use only. Please respect Instagram's Terms of Service and be mindful of rate limiting. The server does not store any content locally and streams content directly from Instagram's servers. **This is not a downloader - it's a privacy frontend for viewing content.**

**🎉 Fun Fact :** **Qwiklip** means **QuickClip**!
If you remove the 'c' from QuickClip, you get `qwiklip` - as I can't C. Just kidding, it's just a clever play on words that captures the essence of fast, efficient way of watching video clips privately.

## 📖 Table of Contents

- [✨ Features](#features)
- [🚀 Installation](#installation)
- [🛠️ Usage](#usage)
- [⚙️ Configuration](#configuration)
- [🔧 Development](#development)
- [📚 Documentation](#documentation)
- [🐛 Bugs or Requests](#bugs-or-requests)
- [🤝 Contributing](#contributing)
- [📄 License](#license)
- [🙏 Acknowledgments](#acknowledgments)

## ✨ Features

- **🔒 Privacy-Focused**: Watch Instagram content without tracking or ads
- **🏗️ Modern Architecture**: Clean, modular design with proper separation of concerns
- **📊 Structured Logging**: Comprehensive logging with slog (Go 1.21+)
- **⚡ High Performance**: Optimized for low latency and high throughput
- **🔄 Multiple Extraction Strategies**: Robust fallback mechanisms for Instagram's API changes
- **📺 Direct Video Streaming**: Efficient streaming without local storage
- **🎯 Range Request Support**: Full HTTP range request support for video seeking
- **🔮 HTML Video Player**: Coming soon - native video player with comments integration
- **💬 Comments Display**: Future feature - view comments alongside videos
- **🏥 Health Monitoring**: Built-in health checks and metrics
- **🐳 Docker Ready**: Multi-stage Docker builds with security best practices
- **🛡️ Error Handling**: Custom error types with proper HTTP status codes
- **🔧 Configuration Management**: Environment-based configuration
- **📦 Graceful Shutdown**: Proper cleanup and signal handling
- **🎨 Automatic Theme Support**: Respects your system's light/dark mode preference
- **🔒 Security**: Non-root container execution and minimal attack surface

## 🚀 Installation

### Prerequisites

- Go 1.25 or later
- Internet connection
- Docker (optional, for containerized deployment)

### Configuration

Before running the application, you can configure it using environment variables. Copy the sample configuration file and modify it as needed:

```bash
# Copy the sample environment file
cp configs/environments/sample.env .env

# Edit the configuration (optional)
# nano .env  # or your preferred editor
```

Available configuration options:
- `PORT`: Server port (default: 8080)
- `LOG_LEVEL`: Logging level - `debug`, `info`, `warn`, `error` (default: info)
- `LOG_FORMAT`: Log format - `text` or `json` (default: text)
- `DEBUG`: Enable debug mode for Instagram client (default: false)

### Quick Start

#### Using Go directly:

```bash
# Clone or download the project
cd qwiklip

# Run the server
go run ./cmd/qwiklip

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

#### Using Docker Compose:

```yaml
version: '3.8'
services:
  qwiklip:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - LOG_LEVEL=info
      - DEBUG=false
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
```

#### Using Pre-built Images from GitHub Container Registry:

```bash
# Pull and run the latest image
docker run -p 8080:8080 ghcr.io/jollySleeper/qwiklip:latest

# Pull and run a specific version
docker run -p 8080:8080 ghcr.io/jollySleeper/qwiklip:v1.0.0

# Run with Docker Compose using pre-built image
```

```yaml
version: '3.8'
services:
  qwiklip:
    image: ghcr.io/jollySleeper/qwiklip:latest
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - LOG_LEVEL=info
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
```

**Multi-Platform Support**: Images are built for both AMD64 and ARM64 architectures, making them compatible with Intel/AMD and Apple Silicon Macs, as well as various server architectures.

## 🛠️ Usage

### URL Format

The server provides an alternative interface for viewing Instagram content using the same URL structure:

**Instagram URL:**
```
https://www.instagram.com/reel/ABC123XYZ/
```

**Qwiklip URL (Privacy-Focused Viewing):**
```
http://localhost:8080/reel/ABC123XYZ/
```

### Supported Content Types

- **Reels**: `/reel/{shortcode}/` - Watch reels privately without ads

### Examples

```bash
# Watch a reel privately
curl http://localhost:8080/reel/C2Z4BcJJ0LU/

# Check server health
curl http://localhost:8080/health
```

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `LOG_FORMAT` | `text` | Log format (text, json) |
| `DEBUG` | `false` | Enable debug mode with additional logging |

### Examples

```bash
# Using .env file (recommended)
cp configs/environments/sample.env .env
# Edit .env file as needed, then run:
go run ./cmd/qwiklip

# Or override specific variables:
PORT=3000 go run ./cmd/qwiklip

# Development with debug logging
DEBUG=true LOG_LEVEL=debug LOG_FORMAT=json go run ./cmd/qwiklip

# Production configuration
PORT=8080 LOG_LEVEL=warn go run ./cmd/qwiklip
```

### Docker Configuration

```bash
# Run with custom configuration
docker run -p 8080:8080 \
  -e PORT=8080 \
  -e LOG_LEVEL=debug \
  -e LOG_FORMAT=json \
  -e DEBUG=true \
  qwiklip
```

## 🔧 Development

### Quick Setup

```bash
# Install dependencies
go mod download

# Run in development mode
DEBUG=true LOG_LEVEL=debug go run ./cmd/qwiklip

# Build for production
make build
```

### Project Structure

```
qwiklip/
├── cmd/qwiklip/              # Application entry point
│   └── main.go             # Main function and startup logic
├── internal/               # Private application code
│   ├── config/            # Configuration management
│   ├── instagram/         # Instagram client logic (3 files)
│   ├── middleware/        # HTTP middleware
│   ├── models/           # Data models and types (2 files)
│   └── server/           # HTTP server logic (2 files)
├── docs/                  # Comprehensive documentation
├── bin/                  # Build artifacts (generated)
├── Dockerfile           # Multi-stage Docker build
├── Makefile            # Build automation
└── README.md           # This file
```

### Architecture

The application follows **Clean Architecture** with clear separation of concerns:
- **Presentation Layer**: HTTP handlers and middleware
- **Application Layer**: Business logic and use cases
- **Domain Layer**: Core business entities and rules
- **Infrastructure Layer**: External dependencies and data access

### Key Technologies

- **Go 1.25**: Latest language features and optimizations
- **Structured Logging**: slog package for comprehensive observability
- **Dependency Injection**: Clean component wiring and testing
- **Context Propagation**: Proper cancellation and timeouts
- **Custom Error Types**: Structured error handling with HTTP mapping

### Build & Test

```bash
# Run tests
go test ./...

# Build binary
go build -o qwiklip ./cmd/qwiklip

# Run linter
golangci-lint run

# Docker build
docker build -t qwiklip .
```

### CI/CD Pipeline

The project includes a comprehensive GitHub Actions workflow (`.github/workflows/docker.yml`) that:

- **Triggers**: On pushes to `main` branch and when tags are created (e.g., `v1.0.0`)
- **Multi-Platform Builds**: Creates Docker images for both AMD64 and ARM64 architectures
- **Automated Publishing**: Pushes images to GitHub Container Registry (ghcr.io)
- **Smart Tagging**: Applies appropriate tags (latest, version numbers, commit SHAs)
- **Caching**: Uses GitHub Actions cache for faster builds
- **Security**: Includes provenance attestations and proper permissions

#### Workflow Features:

- ✅ **Automated builds** on every push to main
- ✅ **Release builds** when you create git tags
- ✅ **Multi-arch support** (AMD64 + ARM64)
- ✅ **GitHub Container Registry** integration
- ✅ **Build caching** for faster subsequent builds
- ✅ **Pull request validation** (builds but doesn't push)

#### Creating a Release:

```bash
# Create and push a new tag
git tag v1.0.0
git push origin v1.0.0
```

This will trigger the workflow to build and publish Docker images with the `v1.0.0` tag to GitHub Container Registry.

## 📚 Documentation

Comprehensive documentation is available in the [`docs/`](./docs/) directory:

### 🏗️ **Architecture & Design**
- [Project Structure](./docs/architecture/project-structure.md) - Complete codebase organization
- [Design Patterns](./docs/architecture/design-patterns.md) - Dependency injection, context usage, and more
- [Package Organization](./docs/architecture/package-organization.md) - How internal packages are structured
- [Configuration Management](./docs/architecture/configuration.md) - Environment-based configuration

### 🔧 **Core Components**
- [Instagram Client](./docs/components/instagram-client.md) - Video extraction and processing
- [HTTP Server](./docs/components/http-server.md) - Request handling and middleware
- [Error Handling](./docs/components/error-handling.md) - Custom error types and responses
- [Logging System](./docs/components/logging.md) - Structured logging with slog

### 📋 **API Reference**
- [HTTP Endpoints](./docs/api/endpoints.md) - Available API endpoints and usage
- [Response Formats](./docs/api/responses.md) - API response structures
- [Error Codes](./docs/api/errors.md) - Error response codes and meanings

### 🚀 **Quick Access**
```bash
# Open main documentation
open docs/README.md

# Or start with project structure
open docs/architecture/project-structure.md
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
