# 📊 Logging System

This document explains Qwiklip's structured logging system, built on Go's `slog` package for comprehensive observability and debugging.

## 🎯 **Overview**

Qwiklip implements a modern logging system that provides:

- **Structured Logging** - Consistent, machine-readable log format
- **Configurable Levels** - Debug, Info, Warn, Error levels
- **Multiple Outputs** - Text and JSON formats
- **Context Propagation** - Request tracing and correlation
- **Performance Optimized** - Minimal overhead in production

## 🏗️ **Architecture**

### **Component Structure**

```
internal/middleware/
└── logging.go    # Logging middleware and utilities
```

### **Core Logger Interface**

```go
// slog.Logger provides structured logging
type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
    With(args ...any) *Logger
}
```

### **Logger Configuration**

```go
// Logger creation based on configuration
func setupLogger(cfg *config.LoggingConfig) *slog.Logger {
    var handler slog.Handler

    if cfg.Format == "json" {
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: getLogLevel(cfg.Level),
        })
    } else {
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: getLogLevel(cfg.Level),
        })
    }

    return slog.New(handler)
}
```

## 🔧 **Log Levels**

### **Level Hierarchy**

```go
type Level int

const (
    LevelDebug Level = -4
    LevelInfo  Level = 0
    LevelWarn  Level = 4
    LevelError Level = 8
)
```

### **Level Usage Guidelines**

```go
// DEBUG: Detailed diagnostic information
logger.Debug("extraction attempt",
    "shortcode", shortcode,
    "strategy", strategyName,
    "attempt", attemptNumber)

// INFO: General information about application operation
logger.Info("server started",
    "port", port,
    "mode", mode)

// WARN: Warning messages for potentially harmful situations
logger.Warn("fallback strategy used",
    "primary_strategy", "json_extract",
    "fallback_strategy", "html_parse")

// ERROR: Error conditions that don't stop the application
logger.Error("extraction failed",
    "shortcode", shortcode,
    "error", err,
    "strategies_tried", len(strategies))
```

## 📝 **Log Formats**

### **Text Format (Development)**

```
2025-01-14T06:48:30.894+05:30 INFO Qwiklip starting port=8080 reel_url=http://localhost:8080/reel/{shortcode}/ info_url=http://localhost:8080/
2025-01-14T06:48:30.894+05:30 INFO Server started. Press Ctrl+C to stop.
2025-01-14T06:48:30.894+05:30 ERROR Server failed to start error="listen tcp :8080: bind: address already in use"
```

### **JSON Format (Production)**

```json
{
  "time": "2025-01-14T06:48:30.894+05:30",
  "level": "INFO",
  "msg": "Qwiklip starting",
  "port": 8080,
  "reel_url": "http://localhost:8080/reel/{shortcode}/",
  "info_url": "http://localhost:8080/"
}
```

## 🎯 **Logging Middleware**

### **HTTP Request Logging**

```go
func LoggingMiddleware(logger *slog.Logger) func(http.HandlerFunc) http.HandlerFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            clientIP := getClientIP(r)

            // Log request start
            logger.Info("Request started",
                "method", r.Method,
                "path", r.URL.Path,
                "client_ip", clientIP,
                "user_agent", r.UserAgent())

            // Create response wrapper to capture status code
            wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

            // Process request
            next(wrapper, r)

            // Calculate duration
            duration := time.Since(start)

            // Log request completion
            logger.Info("Request completed",
                "method", r.Method,
                "path", r.URL.Path,
                "status", wrapper.statusCode,
                "duration_ms", duration.Milliseconds(),
                "client_ip", clientIP)
        }
    }
}
```

### **Response Writer Wrapper**

```go
type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}
```

### **Client IP Extraction**

```go
func getClientIP(r *http.Request) string {
    // Check X-Forwarded-For header (most common with proxies)
    if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
        return forwarded
    }

    // Check X-Real-IP header (used by some proxies)
    if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
        return realIP
    }

    // Fall back to RemoteAddr
    return r.RemoteAddr
}
```

## 📊 **Context-Aware Logging**

### **Logger with Context**

```go
// Add context to logger for request tracing
func (s *Server) handleReel(w http.ResponseWriter, r *http.Request) {
    // Create request-scoped logger
    requestLogger := s.logger.With(
        "request_id", generateRequestID(),
        "method", r.Method,
        "path", r.URL.Path,
        "client_ip", getClientIP(r))

    requestLogger.Info("Processing Instagram request")

    // Pass logger to downstream functions
    mediaInfo, err := s.instagramClient.GetMediaInfo(instagramURL, requestLogger)
    if err != nil {
        requestLogger.Error("Failed to get media info", "error", err)
        return
    }

    requestLogger.Info("Successfully processed request",
        "video_url", mediaInfo.VideoURL[:100]+"...")
}
```

### **Component-Specific Loggers**

```go
func (c *Client) GetMediaInfo(urlStr string) (*models.InstagramMediaInfo, error) {
    // Component logger with context
    componentLogger := c.logger.With(
        "component", "instagram_client",
        "url", urlStr)

    componentLogger.Info("Starting media extraction")

    shortcode, err := c.ExtractShortcode(urlStr)
    if err != nil {
        componentLogger.Error("Shortcode extraction failed", "error", err)
        return nil, err
    }

    componentLogger.Info("Shortcode extracted", "shortcode", shortcode)

    // Continue with detailed logging...
}
```

## 🔍 **Debug Logging**

### **Conditional Debug Logging**

```go
func (c *Client) debugLogExtraction(shortcode, htmlSnippet string) {
    if !c.config.Debug {
        return
    }

    c.logger.Debug("Extraction attempt details",
        "shortcode", shortcode,
        "html_length", len(htmlSnippet),
        "html_preview", htmlSnippet[:min(200, len(htmlSnippet))])
}
```

### **Performance Debug Logging**

```go
func (c *Client) logPerformance(operation string, start time.Time, success bool) {
    duration := time.Since(start)

    if !c.config.Debug && duration < 100*time.Millisecond {
        return // Don't log fast operations in production
    }

    c.logger.Info("Operation completed",
        "operation", operation,
        "duration_ms", duration.Milliseconds(),
        "success", success)
}
```

## 📈 **Structured Data Logging**

### **Error Logging with Context**

```go
func (s *Server) logError(err error, r *http.Request) {
    var appErr *models.AppError
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

### **Business Metrics Logging**

```go
func (c *Client) logExtractionMetrics(shortcode string, strategiesTried int, success bool, duration time.Duration) {
    c.logger.Info("Extraction completed",
        "shortcode", shortcode,
        "strategies_tried", strategiesTried,
        "success", success,
        "duration_ms", duration.Milliseconds(),
        "success_rate", calculateSuccessRate(success))
}
```

## 🎯 **Configuration Integration**

### **Dynamic Log Level Changes**

```go
type LogLevelManager struct {
    currentLevel slog.Level
    mu          sync.RWMutex
}

func (lm *LogLevelManager) SetLevel(level slog.Level) {
    lm.mu.Lock()
    defer lm.mu.Unlock()
    lm.currentLevel = level
}

func (lm *LogLevelManager) GetLevel() slog.Level {
    lm.mu.RLock()
    defer lm.mu.RUnlock()
    return lm.currentLevel
}
```

### **Environment-Based Configuration**

```go
func getLogLevel(levelStr string) slog.Level {
    switch strings.ToLower(levelStr) {
    case "debug":
        return slog.LevelDebug
    case "info":
        return slog.LevelInfo
    case "warn":
        return slog.LevelWarn
    case "error":
        return slog.LevelError
    default:
        return slog.LevelInfo
    }
}
```

## 📊 **Log Aggregation and Analysis**

### **Log Parsing for Metrics**

```go
type LogParser struct {
    pattern *regexp.Regexp
}

func (lp *LogParser) ParseLogLine(line string) (*LogEntry, error) {
    matches := lp.pattern.FindStringSubmatch(line)
    if len(matches) == 0 {
        return nil, fmt.Errorf("log line doesn't match pattern")
    }

    return &LogEntry{
        Timestamp:   parseTimestamp(matches[1]),
        Level:       matches[2],
        Message:     matches[3],
        Fields:      parseFields(matches[4:]),
    }, nil
}
```

### **Metrics Extraction**

```go
func extractMetrics(logs []*LogEntry) *LogMetrics {
    metrics := &LogMetrics{}

    for _, entry := range logs {
        switch entry.Level {
        case "ERROR":
            metrics.ErrorCount++
            if strings.Contains(entry.Message, "rate_limited") {
                metrics.RateLimitErrors++
            }
        case "WARN":
            metrics.WarnCount++
        case "INFO":
            metrics.InfoCount++
            if strings.Contains(entry.Message, "extraction") {
                metrics.ExtractionOperations++
            }
        }
    }

    return metrics
}
```

## 🧪 **Testing Logging**

### **Logger Mocking**

```go
type mockLogger struct {
    logs []LogEntry
    mu   sync.Mutex
}

func (m *mockLogger) Info(msg string, args ...any) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.logs = append(m.logs, LogEntry{
        Level:   "INFO",
        Message: msg,
        Args:    args,
    })
}

func (m *mockLogger) AssertLog(t *testing.T, expectedLevel, expectedMessage string) {
    m.mu.Lock()
    defer m.mu.Unlock()

    for _, log := range m.logs {
        if log.Level == expectedLevel && strings.Contains(log.Message, expectedMessage) {
            return // Found expected log
        }
    }

    t.Errorf("Expected log not found: %s %s", expectedLevel, expectedMessage)
}
```

### **Test Logger Setup**

```go
func setupTestLogger() *slog.Logger {
    return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    }))
}

func TestWithLogging(t *testing.T) {
    logger := setupTestLogger()
    client := NewInstagramClient(testConfig, logger)

    // Test that logs appropriate messages
    _, err := client.GetMediaInfo("invalid-url")
    assert.Error(t, err)

    // Verify error was logged
    // (In a real test, you'd capture logs and verify)
}
```

## 🚀 **Advanced Logging Features**

### **Log Rotation**

```go
func setupLogRotation(filename string, maxSize int64, maxBackups int) *lumberjack.Logger {
    return &lumberjack.Logger{
        Filename:   filename,
        MaxSize:    int(maxSize / 1024 / 1024), // MB
        MaxBackups: maxBackups,
        Compress:   true,
    }
}
```

### **Remote Logging**

```go
func setupRemoteLogger(endpoint string) *slog.Logger {
    // Send logs to remote service
    return slog.New(slog.NewJSONHandler(&remoteWriter{endpoint: endpoint}, nil))
}

type remoteWriter struct {
    endpoint string
}

func (w *remoteWriter) Write(p []byte) (n int, err error) {
    // Send log to remote service
    resp, err := http.Post(w.endpoint, "application/json", bytes.NewReader(p))
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()
    return len(p), nil
}
```

### **Sampling**

```go
type SamplingLogger struct {
    logger   *slog.Logger
    rate     float64 // Sample rate (0.1 = 10% of logs)
    counter  int64
}

func (sl *SamplingLogger) Info(msg string, args ...any) {
    if atomic.AddInt64(&sl.counter, 1) % int64(1/sl.rate) == 0 {
        sl.logger.Info(msg, args...)
    }
}
```

## 📊 **Performance Considerations**

### **Logging Overhead**

```go
// Expensive operation - only compute if logging enabled
func (c *Client) logExpensiveOperation(data []byte) {
    if !c.logger.Enabled(slog.LevelDebug) {
        return // Skip expensive computation
    }

    expensiveResult := computeExpensiveSummary(data)
    c.logger.Debug("expensive operation result", "result", expensiveResult)
}
```

### **Async Logging**

```go
type AsyncLogger struct {
    logger *slog.Logger
    queue  chan LogEntry
}

func (al *AsyncLogger) Start() {
    go func() {
        for entry := range al.queue {
            al.logger.Log(entry.Level, entry.Message, entry.Args...)
        }
    }()
}

func (al *AsyncLogger) Info(msg string, args ...any) {
    select {
    case al.queue <- LogEntry{Level: slog.LevelInfo, Message: msg, Args: args}:
    default:
        // Queue full, drop log to prevent blocking
    }
}
```

## 📚 **Best Practices**

### **1. Consistent Field Names**

```go
// ✅ Good: Consistent field naming
logger.Info("request completed",
    "method", r.Method,
    "path", r.URL.Path,
    "status_code", statusCode,
    "duration_ms", duration.Milliseconds())

// ❌ Bad: Inconsistent field naming
logger.Info("request completed",
    "method", r.Method,
    "url", r.URL.Path,        // Different name for same thing
    "status", statusCode,     // Different name
    "time", duration)         // Different format
```

### **2. Appropriate Log Levels**

```go
// ✅ Good: Appropriate levels
logger.Debug("detailed diagnostic info")  // Development only
logger.Info("user action completed")      // Normal operations
logger.Warn("degraded performance")       // Needs attention
logger.Error("operation failed")          // Requires action
```

### **3. Structured Data**

```go
// ✅ Good: Structured data
logger.Info("extraction completed",
    "shortcode", shortcode,
    "strategies_tried", len(strategies),
    "success", success,
    "duration_ms", duration.Milliseconds())

// ❌ Bad: String concatenation
logger.Info(fmt.Sprintf("extraction completed for %s, tried %d strategies, success: %v",
    shortcode, len(strategies), success))
```

## 📚 **Further Reading**

- [Go slog Package](https://pkg.go.dev/log/slog)
- [Structured Logging](https://www.ardanlabs.com/blog/2023/07/structured-logging-in-go.html)
- [Logging Best Practices](https://dave.cheney.net/2015/11/05/lets-talk-about-logging)

---

**Next**: Learn about the [API endpoints](./../api/endpoints.md) and their usage.
