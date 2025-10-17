# 📦 Package Organization

This document explains how Qwiklip's internal packages are organized, their responsibilities, and how they interact with each other.

## 🏗️ **Internal Package Structure**

```
internal/
├── config/          # Configuration management
├── instagram/       # Instagram client and extraction logic
├── middleware/      # HTTP middleware components
├── models/         # Data models and types
└── server/         # HTTP server and request handling
```

## 📋 **Package Responsibilities**

### **`internal/config/` - Configuration Management**

**Purpose**: Centralized configuration loading and validation.

**Key Components:**
- `Config` struct - Main configuration container
- `Load()` function - Environment variable loading
- Configuration validation and defaults

**Responsibilities:**
- ✅ Load environment variables with defaults
- ✅ Validate configuration values
- ✅ Provide type-safe configuration access
- ✅ Support different environments (dev/prod)

**Example Usage:**
```go
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}
// Use cfg.Server.Port, cfg.Instagram.Timeout, etc.
```

### **`internal/instagram/` - Instagram Client Logic**

**Purpose**: All Instagram-related functionality including video extraction and API interactions.

**Key Components:**
- `Client` struct - Main Instagram client
- Extraction strategies and fallbacks
- JSON parsing and video URL discovery
- HTTP client management for Instagram requests

**Files:**
- `client.go` - Main client implementation and public API
- `extraction.go` - HTML/JSON data extraction logic
- `parser.go` - Data parsing and validation

**Responsibilities:**
- ✅ Extract video URLs from Instagram pages
- ✅ Handle multiple extraction strategies
- ✅ Parse Instagram's JSON responses
- ✅ Manage HTTP requests to Instagram
- ✅ Handle rate limiting and errors

**Example Usage:**
```go
client := instagram.NewClient(&cfg.Instagram, logger)
mediaInfo, err := client.GetMediaInfo("https://instagram.com/reel/ABC123/")
if err != nil {
    // handle error
}
// Use mediaInfo.VideoURL, mediaInfo.FileName, etc.
```

### **`internal/middleware/` - HTTP Middleware**

**Purpose**: HTTP middleware components for request processing, logging, and cross-cutting concerns.

**Key Components:**
- Logging middleware with structured logging
- CORS middleware for cross-origin requests
- Recovery middleware for panic handling
- Timeout middleware for request timeouts

**Responsibilities:**
- ✅ Request/response logging with structured data
- ✅ CORS header management
- ✅ Panic recovery and error logging
- ✅ Request timeout enforcement
- ✅ Security headers (future)

**Example Usage:**
```go
// Chain middleware
handler := recovery(logging(cors(actualHandler)))
mux.HandleFunc("/api/", handler)
```

### **`internal/models/` - Data Models**

**Purpose**: Data structures, types, and domain models used throughout the application.

**Key Components:**
- `InstagramMediaInfo` - Instagram media data structure
- Custom error types with HTTP status mapping
- API request/response models
- Type definitions and constants

**Files:**
- `media.go` - Instagram media structures
- `errors.go` - Custom error types and handling

**Responsibilities:**
- ✅ Define data structures for Instagram content
- ✅ Provide custom error types with context
- ✅ Map errors to appropriate HTTP status codes
- ✅ Ensure type safety across packages

**Example Usage:**
```go
// Media info structure
media := &models.InstagramMediaInfo{
    VideoURL:    "https://cdn.instagram.com/...",
    FileName:    "reel_ABC123.mp4",
    Username:    "instagram_user",
    Caption:     "Amazing content!",
}

// Custom error
if err := processMedia(media); err != nil {
    var appErr *models.AppError
    if errors.As(err, &appErr) {
        httpStatus := appErr.HTTPStatusCode()
        // Handle based on HTTP status
    }
}
```

### **`internal/server/` - HTTP Server**

**Purpose**: HTTP server setup, routing, and request handling.

**Key Components:**
- `Server` struct - Main server implementation
- HTTP route configuration
- Request handlers for different endpoints
- Server lifecycle management

**Files:**
- `server.go` - Server setup, lifecycle, and routing
- `handlers.go` - Individual HTTP request handlers

**Responsibilities:**
- ✅ HTTP server lifecycle management
- ✅ Route configuration and middleware chaining
- ✅ Request handling and response formatting
- ✅ Graceful shutdown handling
- ✅ Health check endpoints

**Example Usage:**
```go
server := server.New(cfg, instagramClient, logger)
ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer cancel()

if err := server.Start(ctx); err != nil {
    log.Fatal("Server failed:", err)
}
```

## 🔄 **Package Dependencies**

### **Dependency Graph**

```
cmd/qwiklip/main.go
        ↓
    internal/config     (configuration)
        ↓
    internal/instagram  (business logic)
    internal/server     (HTTP layer)
        ↓
    internal/middleware (cross-cutting)
    internal/models     (data types)
```

### **Import Rules**

1. **Main package** imports all internal packages
2. **Internal packages** can import other internal packages
3. **External packages** cannot import internal packages (Go restriction)
4. **Circular imports** are prevented by clean layering

### **Clean Architecture Layers**

```
┌─────────────────────────────────────┐
│        Presentation Layer          │
│     internal/server/handlers.go    │
├─────────────────────────────────────┤
│        Application Layer           │
│     internal/server/server.go      │
│     internal/instagram/client.go   │
├─────────────────────────────────────┤
│          Domain Layer              │
│     internal/models/media.go       │
│     internal/models/errors.go      │
├─────────────────────────────────────┤
│      Infrastructure Layer          │
│   internal/instagram/extraction.go │
│   internal/instagram/parser.go     │
└─────────────────────────────────────┘
```

## 📊 **Package Metrics**

### **Size and Complexity**

| Package | Files | Lines | Complexity |
|---------|-------|-------|------------|
| config | 1 | ~80 | Low |
| instagram | 3 | ~450 | Medium |
| middleware | 1 | ~120 | Low |
| models | 2 | ~150 | Low |
| server | 2 | ~250 | Medium |

### **Test Coverage Goals**

- **config**: 90%+ (simple configuration loading)
- **models**: 95%+ (data structures and error handling)
- **middleware**: 85%+ (HTTP middleware logic)
- **instagram**: 80%+ (complex extraction logic)
- **server**: 75%+ (HTTP handlers and server logic)

## 🎯 **Package Design Principles**

### **1. Single Responsibility Principle**

Each package has one clear, well-defined responsibility:

- **`config`**: Configuration only
- **`instagram`**: Instagram operations only
- **`middleware`**: HTTP middleware only
- **`models`**: Data types only
- **`server`**: HTTP server only

### **2. Dependency Inversion**

High-level modules don't depend on low-level modules:

```go
// ✅ Good: Server depends on interfaces/abstractions
type Server struct {
    config          *config.Config      // Configuration interface
    instagramClient *instagram.Client   // Instagram interface
    logger          *slog.Logger        // Logging interface
}

// ❌ Bad: Direct dependencies on implementations
type Server struct {
    config *config.Config
    httpClient *http.Client  // Direct implementation
}
```

### **3. Interface Segregation**

Keep interfaces small and focused:

```go
// ✅ Good: Focused interfaces
type InstagramClient interface {
    GetMediaInfo(url string) (*models.InstagramMediaInfo, error)
}

type Logger interface {
    Info(msg string, args ...any)
    Error(msg string, args ...any)
}

// ❌ Bad: Large, unfocused interfaces
type Service interface {
    GetMediaInfo(url string) (*models.InstagramMediaInfo, error)
    LogInfo(msg string, args ...any)
    SaveToDatabase(data interface{}) error
    SendEmail(to, subject, body string) error
}
```

## 🔧 **Package Development Guidelines**

### **1. Package Naming**

- **Short and descriptive**: `config`, `server`, `models`
- **Domain-driven**: `instagram` for Instagram-specific logic
- **Action-based for complex**: `extraction`, `parser`

### **2. File Organization**

- **Single responsibility**: Each file has one clear purpose
- **Logical grouping**: Related functionality in same package
- **Size limits**: Keep files under 300 lines when possible

### **3. Import Organization**

```go
import (
    // Standard library
    "context"
    "fmt"
    "net/http"

    // Third-party packages
    "log/slog"

    // Internal packages
    "qwiklip/internal/config"
    "qwiklip/internal/models"
)
```

### **4. Error Handling**

- Use custom error types from `internal/models`
- Wrap errors with context using `fmt.Errorf`
- Return appropriate HTTP status codes

### **5. Testing**

- Unit tests for each package
- Mock dependencies using dependency injection
- Integration tests for package interactions

## 🚀 **Package Evolution**

### **Adding New Packages**

1. **Identify the need**: What functionality is missing?
2. **Create package**: `internal/newpackage/`
3. **Define interfaces**: What does this package need to do?
4. **Implement**: Write the core functionality
5. **Integrate**: Update main.go and other packages to use it

### **Package Splitting**

When a package grows too large:

1. **Identify boundaries**: What can be separated?
2. **Create subpackage**: `internal/package/subpackage/`
3. **Move code**: Migrate related functionality
4. **Update imports**: Fix import statements
5. **Update tests**: Ensure all tests still pass

### **Package Consolidation**

When packages are too small or closely related:

1. **Assess coupling**: How tightly are they related?
2. **Merge packages**: Combine into single package
3. **Update imports**: Fix all import statements
4. **Refactor**: Clean up any duplicated code

## 🏆 **Benefits of This Organization**

### **1. Clear Boundaries**
- Easy to understand what each package does
- Simple to modify individual components
- Clear ownership and responsibilities

### **2. Scalability**
- Easy to add new features in appropriate packages
- New packages can be added as needed
- Clear growth path for the codebase

### **3. Testability**
- Each package can be tested independently
- Dependency injection enables easy mocking
- Clear interfaces make testing straightforward

### **4. Maintainability**
- Changes are localized to specific packages
- Clear dependency graph makes changes predictable
- Easy to refactor within package boundaries

### **5. Developer Experience**
- Standard Go layout is familiar
- Easy to navigate and understand
- Clear where to add new code

## 📚 **Further Reading**

- [Effective Go - Package Organization](https://golang.org/doc/effective_go#package-names)
- [Go Package Guidelines](https://rakyll.org/style-packages/)
- [Package Design Principles](https://www.ardanlabs.com/blog/2017/02/package-oriented-design.html)

---

**Next**: Learn about the [configuration management](./configuration.md) system.
