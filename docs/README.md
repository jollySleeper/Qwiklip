# 📚 Qwiklip Documentation

Welcome to the comprehensive documentation for **Qwiklip**, a privacy-focused Instagram frontend built with Go 1.25. Qwiklip provides an alternative way to watch Instagram reels and posts without tracking, ads, or the official Instagram interface.

## 📖 Table of Contents

### 🏗️ **Architecture & Design**
- [Project Structure](./architecture/project-structure.md) - Complete codebase organization
- [Design Patterns](./architecture/design-patterns.md) - Dependency injection, context usage, and more
- [Package Organization](./architecture/package-organization.md) - How internal packages are structured
- [Configuration Management](./architecture/configuration.md) - Environment-based configuration

### 🔧 **Core Components**
- [Instagram Client](./components/instagram-client.md) - Video extraction and processing
- [HTTP Server](./components/http-server.md) - Request handling and middleware
- [Error Handling](./components/error-handling.md) - Custom error types and responses
- [Logging System](./components/logging.md) - Structured logging with slog

### 📋 **API Reference**
- [HTTP Endpoints](./api/endpoints.md) - Available API endpoints and usage
- [Response Formats](./api/responses.md) - API response structures
- [Error Codes](./api/errors.md) - Error response codes and meanings

### 📈 **Documentation Coverage**

This documentation provides **comprehensive coverage** of:
- **🏗️ 4 Architecture & Design guides** - Project structure, patterns, packages, config
- **🔧 4 Core Component guides** - Instagram client, HTTP server, errors, logging
- **📋 3 API Reference guides** - Endpoints, responses, error codes
- **📚 2,500+ lines** of technical documentation with examples
- **🎯 Complete implementation details** for production-ready code

## 🎯 **Quick Start**

```bash
# Clone and setup
git clone <repository-url>
cd qwiklip

# Run the server
go run ./cmd/server

# Or build and run
make build
./bin/qwiklip
```

## 🏛️ **Architecture Overview**

Qwiklip follows a **Clean Architecture** pattern with clear separation of concerns, designed specifically for privacy-focused content viewing:

```
┌─────────────────────────────────────────┐
│             Presentation Layer         │
│         (HTTP Handlers & Middleware)   │
├─────────────────────────────────────────┤
│           Application Layer            │
│      (Privacy & Content Logic)         │
├─────────────────────────────────────────┤
│             Domain Layer               │
│       (Media & Privacy Entities)       │
├─────────────────────────────────────────┤
│          Infrastructure Layer          │
│   (Instagram API & Streaming)          │
└─────────────────────────────────────────┘
```

### **Key Principles**

- ✅ **Privacy-First Design** - No tracking, no ads, no data collection
- ✅ **Dependency Injection** - Clean dependency management
- ✅ **Context Usage** - Proper cancellation and timeouts
- ✅ **Error Handling** - Custom error types with HTTP mapping
- ✅ **Structured Logging** - Comprehensive observability
- ✅ **Configuration Management** - Environment-based config
- ✅ **Graceful Shutdown** - Proper resource cleanup

## 🔄 **Request Flow**

```
Privacy Request → HTTP Handler → Instagram Client → Stream Response
                    ↓              ↓
            Middleware      Privacy Extraction
               ↓              ↓
            Logging        JSON Parsing
               ↓              ↓
            CORS           Video URL
               ↓              ↓
            Timeout        Privacy Streaming
```

## 📊 **Features**

- **🔒 Privacy-Focused** - Watch Instagram content without tracking or ads
- **🏗️ Modern Architecture** - Clean, modular design with proper separation of concerns
- **📊 Structured Logging** - Comprehensive logging with slog (Go 1.21+)
- **⚡ High Performance** - Optimized for low latency and high throughput
- **🔄 Multiple Extraction Strategies** - Robust fallback mechanisms for Instagram's API changes
- **📺 Direct Video Streaming** - Efficient streaming without local storage
- **🎯 Range Request Support** - Full HTTP range request support for video seeking
- **🔮 HTML Video Player** - Coming soon - native video player with comments integration
- **💬 Comments Display** - Future feature - view comments alongside videos
- **🏥 Health Monitoring** - Built-in health checks and metrics
- **🐳 Docker Ready** - Multi-stage Docker builds with security best practices
- **🎨 Automatic Theme Support** - Respects your system's light/dark mode preference
- **🛡️ Error Handling** - Custom error types with proper HTTP status codes
- **🔧 Configuration Management** - Environment-based configuration
- **📦 Graceful Shutdown** - Proper cleanup and signal handling
- **🔒 Security** - Non-root container execution and minimal attack surface

## 🎉 **What Makes This Special**

This project demonstrates **modern Go development practices** for 2025 with a focus on **privacy and user freedom**:

1. **Privacy-First Architecture** - Built specifically to protect user privacy and avoid tracking
2. **Go 1.25 Features** - Latest language features and optimizations
3. **Production Ready** - Designed for real-world deployment
4. **Industry Standards** - Follows official Go project layout
5. **Best Practices** - Clean architecture, dependency injection, proper error handling
6. **Developer Experience** - Comprehensive tooling and documentation
7. **Scalability** - Designed to grow and evolve over time
8. **User Empowerment** - Giving users control over their viewing experience

---

## 📞 **Getting Help**

- 📖 **Documentation** - You're reading it! Check individual sections for detailed guides
- 📖 **Project README** - [Main project README](../README.md) for quick start and usage
- 🐛 **Issues** - Report bugs or request features in the main repository
- 💬 **Discussions** - Ask questions about the codebase architecture

---

**🎯 Ready to dive deeper?** Start with [Project Structure](./architecture/project-structure.md) to understand how everything fits together!
