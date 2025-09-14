# 🏗️ Project Structure

This document provides a comprehensive overview of Qwiklip's codebase organization, following the **official Go project layout** standards.

## 📁 **Complete Directory Structure**

```
qwiklip/
├── cmd/server/                      # Application entry points
│   └── main.go                     # Main application entry point
├── internal/                       # Private application code
│   ├── config/                    # Configuration management
│   │   └── config.go              # Configuration structs and loading
│   ├── instagram/                 # Instagram client logic
│   │   ├── client.go              # Main Instagram client implementation
│   │   ├── extraction.go          # JSON/HTML data extraction logic
│   │   └── parser.go              # Data parsing and validation
│   ├── middleware/                # HTTP middleware components
│   │   └── logging.go             # Logging, CORS, and other middleware
│   ├── models/                   # Data models and types
│   │   ├── media.go              # Instagram media data structures
│   │   └── errors.go             # Custom error types and handling
│   └── server/                   # HTTP server logic
│       ├── server.go             # Server setup and lifecycle management
│       └── handlers.go           # HTTP request handlers
├── docs/                         # Comprehensive documentation
│   ├── README.md                 # Documentation overview
│   ├── architecture/             # Architecture documentation
│   ├── components/               # Component documentation
│   ├── development/              # Development guides
│   ├── api/                      # API documentation
│   └── deployment/               # Deployment guides
├── bin/                          # Build artifacts (generated)
│   └── qwiklip                   # Compiled binary
├── Dockerfile                   # Multi-stage Docker build
├── .dockerignore                # Docker build exclusions
├── Makefile                    # Build automation and tasks
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
└── README.md                   # Project README
```

## 🗂️ **Directory Explanations**

### **`cmd/server/` - Application Entry Points**

**Purpose**: Contains the main entry points for different applications or commands.

**Files:**
- `main.go` - The main application entry point that initializes dependencies and starts the server

**Why this structure?**
- Allows multiple commands in the same repository (e.g., `cmd/server/`, `cmd/cli/`, `cmd/worker/`)
- Each command has its own main.go with minimal dependencies
- Clear separation of different executable binaries

### **`internal/` - Private Application Code**

**Purpose**: Contains all private application code that should not be imported by external packages.

**Why `internal/`?**
- **Go 1.4+ feature**: Packages inside `internal/` can only be imported by parent directories
- Prevents external modules from importing internal implementation details
- Enforces clean API boundaries
- Allows internal refactoring without breaking external consumers

#### **`internal/config/` - Configuration Management**
- **Purpose**: Centralized configuration management
- **Responsibilities**:
  - Environment variable loading
  - Configuration validation
  - Default value handling
  - Type-safe configuration structs

#### **`internal/instagram/` - Instagram Client Logic**
- **Purpose**: All Instagram-related functionality
- **Responsibilities**:
  - Video URL extraction from Instagram pages
  - Multiple extraction strategies and fallbacks
  - HTTP client management for Instagram requests
  - JSON parsing and data validation

#### **`internal/middleware/` - HTTP Middleware**
- **Purpose**: HTTP middleware components
- **Responsibilities**:
  - Request logging and monitoring
  - CORS handling
  - Request timeouts
  - Error recovery
  - Authentication (future)

#### **`internal/models/` - Data Models**
- **Purpose**: Data structures and types
- **Responsibilities**:
  - Instagram media information structs
  - Custom error types
  - API request/response models
  - Type definitions

#### **`internal/server/` - HTTP Server Logic**
- **Purpose**: HTTP server setup and request handling
- **Responsibilities**:
  - Server lifecycle management
  - Route configuration
  - HTTP request handlers
  - Response formatting
  - Middleware chaining

### **`docs/` - Documentation**

**Purpose**: Comprehensive project documentation following Diátaxis framework.

**Structure:**
- `architecture/` - System design and organization
- `components/` - Individual component documentation
- `development/` - Development workflows and guides
- `api/` - API reference and specifications
- `deployment/` - Deployment and operations guides

### **`bin/` - Build Artifacts**

**Purpose**: Contains generated build artifacts.

**Contents:**
- Compiled binaries for different platforms
- Temporary build files
- Distribution packages

**Note**: This directory is typically `.gitignore`d and generated during builds.

## 📊 **Package Dependencies**

```
cmd/server/main.go
    ↓
internal/config      # Configuration loading
internal/instagram   # Instagram client
internal/server      # HTTP server
    ↓
internal/middleware  # HTTP middleware
internal/models      # Data models
```

### **Dependency Flow**

1. **`main.go`** - Entry point, orchestrates all components
2. **`config`** - Loaded first, provides configuration to all components
3. **`instagram`** - Created with config and logger
4. **`server`** - Created with config, instagram client, and logger
5. **`middleware`** - Used by server for request processing
6. **`models`** - Used by all packages for data structures

## 🔒 **Import Rules**

### **Internal Package Rules**
```go
// ✅ Allowed: Internal packages can import each other
import "qwiklip/internal/config"
import "qwiklip/internal/models"

// ❌ Not Allowed: External packages cannot import internal packages
// This would fail if attempted from outside the module
import "qwiklip/internal/server"
```

### **Module Boundaries**
- **Within module**: All packages can import each other
- **Across modules**: Only public packages (not in `internal/`) can be imported
- **Standard library**: Always available
- **Third-party**: Available based on go.mod dependencies

## 🏆 **Benefits of This Structure**

### **1. Clear Separation of Concerns**
- Each package has a single, well-defined responsibility
- Easy to understand what each component does
- Simple to modify individual components

### **2. Scalability**
- Easy to add new features in appropriate packages
- New commands can be added as `cmd/newtool/main.go`
- New internal packages can be added as needed

### **3. Maintainability**
- Clear dependency graph makes changes predictable
- Internal packages prevent external coupling
- Easy to refactor within package boundaries

### **4. Testability**
- Each package can be unit tested independently
- Dependency injection enables easy mocking
- Clear interfaces make testing straightforward

### **5. Developer Experience**
- Standard Go layout is familiar to all Go developers
- Tools like `go mod` work optimally with this structure
- IDE navigation and code completion work better

## 🎯 **Best Practices Applied**

### **Package Naming**
- **Short, descriptive names**: `config`, `server`, `models`
- **Action-based for complex operations**: `extraction`, `parser`
- **Domain-driven**: `instagram` for Instagram-specific logic

### **File Organization**
- **Single responsibility per file**: Each file has one clear purpose
- **Logical grouping**: Related functionality in same package
- **Size management**: Files kept reasonably sized (100-300 lines)

### **Import Organization**
- **Standard library first**
- **Third-party packages second**
- **Internal packages last**
- **Clear import grouping with blank lines**

## 🚀 **Evolution Path**

This structure supports easy evolution:

1. **Add new commands**: `cmd/cli/main.go` for CLI tools
2. **Add new features**: New packages in `internal/`
3. **Split packages**: When packages grow too large, split into subpackages
4. **Add public APIs**: Move stable packages to `pkg/` directory
5. **Add services**: `internal/cache/`, `internal/auth/`, etc.

## 📚 **Further Reading**

- [Official Go Project Layout](https://go.dev/doc/modules/layout)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Go Modules Reference](https://go.dev/doc/modules/developing)
- [Internal Packages](https://go.dev/doc/go1.4#internalpackages)

---

**Next**: Learn about the [design patterns](./design-patterns.md) used in this architecture.
