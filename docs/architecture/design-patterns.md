# 🎨 Design Patterns

This document explores the design patterns and architectural principles implemented in Qwiklip, demonstrating modern Go development practices.

## 📋 **Core Design Patterns**

### **1. Dependency Injection (DI)**

**What is it?** A design pattern where dependencies are "injected" into components rather than being created inside them.

**Why use it?**
- **Testability**: Easy to inject mock dependencies
- **Flexibility**: Easy to swap implementations
- **Decoupling**: Components don't need to know how dependencies are created
- **Maintainability**: Changes to dependency creation don't affect components

#### **Implementation in Qwiklip**

```go
// Constructor injection pattern
func NewServer(cfg *config.Config, instagramClient *instagram.Client, logger *slog.Logger) *Server {
    return &Server{
        config:          cfg,
        instagramClient: instagramClient,
        logger:          logger,
    }
}

// Usage in main.go
func main() {
    cfg := config.Load()           // 1. Create config
    logger := setupLogger(cfg)     // 2. Create logger
    client := instagram.NewClient(&cfg.Instagram, logger)  // 3. Create client
    server := server.New(cfg, client, logger)              // 4. Create server
    // ... start server
}
```

#### **Benefits Achieved**

1. **Testable Components**:
```go
func TestServer(t *testing.T) {
    // Inject mock dependencies
    mockClient := &mockInstagramClient{}
    mockLogger := &mockLogger{}
    server := NewServer(testConfig, mockClient, mockLogger)

    // Test server behavior
    // ...
}
```

2. **Flexible Configuration**:
```go
// Easy to swap implementations
client := &MockInstagramClient{} // For testing
// or
client := &RealInstagramClient{} // For production
```

### **2. Context Propagation**

**What is it?** Using Go's `context.Context` to pass request-scoped values and cancellation signals through the call chain.

**Why use it?**
- **Cancellation**: Cancel operations when client disconnects
- **Timeouts**: Set operation deadlines
- **Request Tracing**: Pass request IDs through the call chain
- **Resource Cleanup**: Automatic cleanup on context cancellation

#### **Implementation in Qwiklip**

```go
// Server-level context handling
func (s *Server) Start(ctx context.Context) error {
    // Wait for shutdown signal
    <-ctx.Done()

    // Graceful shutdown with timeout
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    return s.httpServer.Shutdown(shutdownCtx)
}

// HTTP request context propagation
func (s *Server) handleReel(w http.ResponseWriter, r *http.Request) {
    // Context flows from HTTP request
    mediaInfo, err := s.instagramClient.GetMediaInfo(instagramURL)

    // Context is automatically cancelled if client disconnects
    req, err := http.NewRequestWithContext(r.Context(), "GET", videoURL, nil)
    // ...
}
```

#### **Context Flow Diagram**

```
Main Context (signal handling)
    ↓
Server Context (graceful shutdown)
    ↓
HTTP Request Context (client cancellation)
    ↓
Instagram Client Context (request timeout)
    ↓
HTTP Client Context (network timeout)
```

### **3. Builder Pattern**

**What is it?** A creational pattern for constructing complex objects step by step.

**Why use it?**
- **Flexible Configuration**: Configure objects with many optional parameters
- **Readable Code**: Fluent interface for configuration
- **Immutable Objects**: Prevent modification after construction

#### **Implementation in Qwiklip**

```go
// Configuration builder pattern
type Config struct {
    Server   ServerConfig
    Instagram InstagramConfig
    Logging   LoggingConfig
}

func Load() (*Config, error) {
    return &Config{
        Server: ServerConfig{
            Port:         getEnv("PORT", "8080"),
            ReadTimeout:  30 * time.Second,
            WriteTimeout: 300 * time.Second,
            IdleTimeout:  120 * time.Second,
        },
        // ... other configs
    }, nil
}
```

### **4. Factory Pattern**

**What is it?** A creational pattern that provides an interface for creating objects without specifying their concrete classes.

**Why use it?**
- **Decoupling**: Separate object creation from usage
- **Flexibility**: Easy to change implementations
- **Testing**: Easy to inject different implementations

#### **Implementation in Qwiklip**

```go
// Factory functions for component creation
func NewInstagramClient(cfg *config.InstagramConfig, logger *slog.Logger) *Client {
    return &Client{
        httpClient: &http.Client{Timeout: cfg.Timeout},
        config:     cfg,
        logger:     logger,
    }
}

func NewServer(cfg *config.Config, client *instagram.Client, logger *slog.Logger) *Server {
    return &Server{
        config:          cfg,
        instagramClient: client,
        logger:          logger,
    }
}
```

### **5. Strategy Pattern**

**What is it?** A behavioral pattern that defines a family of algorithms and makes them interchangeable.

**Why use it?**
- **Algorithm Selection**: Choose different algorithms at runtime
- **Extensibility**: Easy to add new algorithms
- **Single Responsibility**: Each strategy has one job

#### **Implementation in Qwiklip**

```go
// Multiple extraction strategies for Instagram data
func (c *Client) GetMediaInfo(urlStr string) (*models.InstagramMediaInfo, error) {
    // Try different URL formats (strategies)
    urlFormats := []struct {
        url       string
        userAgent string
    }{
        {fmt.Sprintf("https://www.instagram.com/reel/%s/", shortcode), DefaultUserAgent},
        {fmt.Sprintf("https://www.instagram.com/reel/%s/?__a=1&__d=dis", shortcode), MobileUserAgent},
        // ... more strategies
    }

    for _, format := range urlFormats {
        // Try each strategy until one works
        if success := c.tryExtractionStrategy(format); success {
            return mediaInfo, nil
        }
    }
    // ...
}
```

## 🏛️ **Architectural Patterns**

### **1. Clean Architecture**

**What is it?** An architectural pattern that separates concerns into layers with clear boundaries.

#### **Qwiklip's Clean Architecture**

```
┌─────────────────────────────────────────┐
│           Presentation Layer           │
│        (HTTP Handlers & Routes)        │
├─────────────────────────────────────────┤
│          Application Layer             │
│     (Business Logic & Use Cases)       │
├─────────────────────────────────────────┤
│            Domain Layer                │
│      (Entities & Core Business)        │
├─────────────────────────────────────────┤
│         Infrastructure Layer           │
│  (External APIs & Data Access)         │
└─────────────────────────────────────────┘
```

#### **Layer Mapping in Code**

- **Presentation**: `internal/server/handlers.go`, `internal/middleware/`
- **Application**: `internal/server/server.go`, business logic
- **Domain**: `internal/models/`, core business entities
- **Infrastructure**: `internal/instagram/`, external API calls

### **2. Hexagonal Architecture (Ports & Adapters)**

**What is it?** An architectural pattern that isolates the core business logic from external concerns.

#### **Qwiklip's Hexagonal Architecture**

```
           ┌─────────────────────────────────┐
           │         Core Business          │
           │      (Instagram Processing)    │
           └─────────────────────────────────┘
                    │            │
            ┌───────┼────────────┼───────┐
            │       │            │       │
        ┌───▼──┐ ┌──▼──┐    ┌───▼──┐ ┌──▼──┐
        │ HTTP │ │  WS  │    │Instagram│ │Cache│
        │ API  │ │ API  │    │  API   │ │ API │
        └──────┘ └─────┘    └────────┘ └─────┘
```

- **Core**: Instagram video processing logic
- **Ports**: Interfaces defining how core interacts with outside world
- **Adapters**: Concrete implementations (HTTP handlers, Instagram client)

## 🔧 **Behavioral Patterns**

### **1. Chain of Responsibility**

**What is it?** A behavioral pattern where a request is passed along a chain of handlers.

#### **Implementation in Middleware**

```go
// Middleware chain
mux.HandleFunc("/reel/", recovery(logging(cors(handler))))

func LoggingMiddleware(logger *slog.Logger) func(http.HandlerFunc) http.HandlerFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            // Do logging work
            logger.Info("Request started", "path", r.URL.Path)

            // Pass to next handler
            next(w, r)

            // Do more logging work
            logger.Info("Request completed", "status", getStatusCode(w))
        }
    }
}
```

### **2. Template Method Pattern**

**What is it?** A behavioral pattern that defines the skeleton of an algorithm in a base class.

#### **Implementation in Extraction Logic**

```go
// Template method for media extraction
func (c *Client) GetMediaInfo(urlStr string) (*models.InstagramMediaInfo, error) {
    // 1. Extract shortcode (common step)
    shortcode, err := c.ExtractShortcode(urlStr)
    if err != nil {
        return nil, err
    }

    // 2. Try different extraction strategies (variable step)
    for _, strategy := range c.getExtractionStrategies() {
        if mediaInfo := c.tryStrategy(strategy); mediaInfo != nil {
            return mediaInfo, nil
        }
    }

    // 3. Handle failure (common step)
    return nil, fmt.Errorf("all extraction strategies failed")
}
```

## 🎯 **Error Handling Patterns**

### **1. Custom Error Types**

```go
// Custom error with context and HTTP status mapping
type AppError struct {
    Type    ErrorType
    Message string
    Cause   error
    Details map[string]interface{}
}

func (e *AppError) HTTPStatusCode() int {
    switch e.Type {
    case ErrorTypeInvalidURL:
        return 400
    case ErrorTypeNotFound:
        return 404
    case ErrorTypeRateLimited:
        return 429
    default:
        return 500
    }
}
```

### **2. Error Wrapping**

```go
// Wrap errors with context
func (c *Client) GetMediaInfo(urlStr string) (*models.InstagramMediaInfo, error) {
    shortcode, err := c.ExtractShortcode(urlStr)
    if err != nil {
        return nil, fmt.Errorf("failed to extract shortcode from %s: %w", urlStr, err)
    }
    // ...
}
```

## 📊 **Concurrency Patterns**

### **1. Context with Timeout**

```go
func (c *Client) fetchWithTimeout(ctx context.Context, url string) (*http.Response, error) {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // Use context in HTTP request
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    return c.httpClient.Do(req)
}
```

### **2. Graceful Shutdown**

```go
func (s *Server) Start(ctx context.Context) error {
    // Start server in goroutine
    go func() {
        if err := s.httpServer.ListenAndServe(); err != nil {
            // Handle server errors
        }
    }()

    // Wait for shutdown signal
    <-ctx.Done()

    // Graceful shutdown with timeout
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    return s.httpServer.Shutdown(shutdownCtx)
}
```

## 🏆 **Benefits of These Patterns**

### **1. Testability**
- Dependency injection enables easy mocking
- Clear interfaces make testing straightforward
- Isolated components can be tested independently

### **2. Maintainability**
- Clear separation of concerns
- Easy to modify individual components
- Consistent patterns across codebase

### **3. Scalability**
- Patterns support easy addition of new features
- Clean boundaries prevent tight coupling
- Easy to add new strategies and handlers

### **4. Reliability**
- Context propagation ensures proper cleanup
- Error handling patterns provide consistent error responses
- Graceful shutdown prevents resource leaks

### **5. Developer Experience**
- Familiar patterns are easy to understand
- Consistent code structure
- Clear architectural boundaries

## 📚 **Further Reading**

- [Design Patterns in Go](https://refactoring.guru/design-patterns/go)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Go Context Package](https://golang.org/pkg/context/)

---

**Next**: Learn about [package organization](./package-organization.md) and how internal packages are structured.
