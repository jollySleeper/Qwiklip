# Qwiklip Tasks

This Maskfile.md defines all the development, build, and deployment tasks for the Qwiklip application.

## Environment Variables

```bash
export BINARY_NAME=qwiklip
export VERSION=${VERSION:-dev}
export BUILD_DIR=bin
export MAIN_PACKAGE=./cmd/qwiklip
export GOCMD=go
export GOBUILD="$GOCMD build"
export GOCLEAN="$GOCMD clean"
export GOTEST="$GOCMD test"
export GOGET="$GOCMD get"
export GOMOD="$GOCMD mod"
export LDFLAGS="-ldflags \"-X main.version=$VERSION -s -w\""
export GCFLAGS="-gcflags=\"all=-l -B\""
```

## all

> Clean, lint, test, and build the project (default target)

```bash
export BINARY_NAME=qwiklip
export VERSION=${VERSION:-dev}
export BUILD_DIR=bin
export MAIN_PACKAGE=./cmd/qwiklip
export GOCMD=go
export GOBUILD="$GOCMD build"
export GOCLEAN="$GOCMD clean"
export GOTEST="$GOCMD test"
export GOGET="$GOCMD get"
export GOMOD="$GOCMD mod"
export LDFLAGS="-ldflags \"-X main.version=$VERSION -s -w\""
export GCFLAGS="-gcflags=\"all=-l -B\""

mask clean
mask lint
mask test
mask build
```

## build

> Build the binary

```bash
export BINARY_NAME=qwiklip
export VERSION=${VERSION:-dev}
export BUILD_DIR=bin
export MAIN_PACKAGE=./cmd/qwiklip
export GOCMD=go
export GOBUILD="$GOCMD build"

echo "Building $BINARY_NAME..."
mkdir -p $BUILD_DIR
$GOBUILD -o $BUILD_DIR/$BINARY_NAME $MAIN_PACKAGE
echo "Binary built: $BUILD_DIR/$BINARY_NAME"
```

## build-all

> Build for multiple platforms

```bash
export BINARY_NAME=qwiklip
export VERSION=${VERSION:-dev}
export BUILD_DIR=bin
export MAIN_PACKAGE=./cmd/qwiklip
export GOCMD=go
export GOBUILD="$GOCMD build"
export LDFLAGS="-ldflags \"-X main.version=$VERSION -s -w\""

echo "Building for multiple platforms..."
mkdir -p $BUILD_DIR
GOOS=linux GOARCH=amd64 $GOBUILD $LDFLAGS -o $BUILD_DIR/$BINARY_NAME-linux-amd64 $MAIN_PACKAGE
GOOS=linux GOARCH=arm64 $GOBUILD $LDFLAGS -o $BUILD_DIR/$BINARY_NAME-linux-arm64 $MAIN_PACKAGE
GOOS=darwin GOARCH=amd64 $GOBUILD $LDFLAGS -o $BUILD_DIR/$BINARY_NAME-darwin-amd64 $MAIN_PACKAGE
GOOS=darwin GOARCH=arm64 $GOBUILD $LDFLAGS -o $BUILD_DIR/$BINARY_NAME-darwin-arm64 $MAIN_PACKAGE
GOOS=windows GOARCH=amd64 $GOBUILD $LDFLAGS -o $BUILD_DIR/$BINARY_NAME-windows-amd64.exe $MAIN_PACKAGE
echo "All binaries built in $BUILD_DIR/"
```

## test

> Run tests

```bash
export GOCMD=go
export GOTEST="$GOCMD test"

echo "Running tests..."
$GOTEST -v -race -coverprofile=coverage.out ./...
```

## coverage

> Run tests with coverage report

```bash
export GOCMD=go
export GOTEST="$GOCMD test"

mask test
echo "Generating coverage report..."
$GOCMD tool cover -html=coverage.out -o coverage.html
echo "Coverage report generated: coverage.html"
```

## run

> Run the application

```bash
export BINARY_NAME=qwiklip
export MAIN_PACKAGE=./cmd/qwiklip
export GOCMD=go

echo "Running $BINARY_NAME..."
$GOCMD run $MAIN_PACKAGE
```

## run-debug

> Run with debug logging

```bash
export BINARY_NAME=qwiklip
export MAIN_PACKAGE=./cmd/qwiklip
export GOCMD=go

echo "Running $BINARY_NAME with debug logging..."
LOG_LEVEL=debug $GOCMD run $MAIN_PACKAGE
```

## clean

> Clean build artifacts

```bash
export BUILD_DIR=bin
export GOCMD=go
export GOCLEAN="$GOCMD clean"

echo "Cleaning..."
$GOCLEAN
rm -rf $BUILD_DIR
rm -f coverage.out coverage.html
find . -name "*.log" -delete
find . -name "debug-*.html" -delete
```

## fmt

> Format code

```bash
export GOCMD=go

echo "Formatting code..."
$GOCMD fmt ./...
```

## lint

> Run linter

```bash
echo "Running linter..."
if command -v golangci-lint >/dev/null 2>&1; then
    golangci-lint run
else
    echo "golangci-lint not found. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin v1.55.2"
fi
```

## vet

> Run go vet

```bash
export GOCMD=go

echo "Running go vet..."
$GOCMD vet ./...
```

## deps

> Download dependencies

```bash
export GOCMD=go
export GOMOD="$GOCMD mod"

echo "Downloading dependencies..."
$GOMOD download
$GOMOD tidy
```

## deps-update

> Update dependencies

```bash
export GOCMD=go
export GOMOD="$GOCMD mod"

echo "Updating dependencies..."
$GOMOD tidy
$GOCMD get -u ./...
```

## docker-build

> Build Docker image

```bash
export BINARY_NAME=qwiklip
export VERSION=${VERSION:-dev}

echo "Building Docker image..."
docker build -t $BINARY_NAME:$VERSION .
```

## docker-run

> Run Docker container

```bash
export BINARY_NAME=qwiklip
export VERSION=${VERSION:-dev}

echo "Running Docker container..."
docker run --rm -p 8080:8080 $BINARY_NAME:$VERSION
```

## docker-run-debug

> Run Docker container with debug

```bash
export BINARY_NAME=qwiklip
export VERSION=${VERSION:-dev}

echo "Running Docker container with debug..."
docker run --rm -p 8080:8080 -e LOG_LEVEL=debug -e DEBUG=true $BINARY_NAME:$VERSION
```

## dev-setup

> Setup development environment

```bash
mask deps
echo "Setting up development environment..."
if ! command -v golangci-lint >/dev/null 2>&1; then
    echo "Installing golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
fi
echo "Development environment ready!"
```

## mocks

> Generate mocks

```bash
echo "Generating mocks..."
if command -v mockery >/dev/null 2>&1; then
    mockery --all --output ./mocks
else
    echo "mockery not found. Install with: go install github.com/vektra/mockery/v2@latest"
fi
```

## bench

> Run benchmarks

```bash
export GOCMD=go
export GOTEST="$GOCMD test"

echo "Running benchmarks..."
$GOTEST -bench=. -benchmem ./...
```

## integration-test

> Run integration tests by starting server and making HTTP calls

```bash
export BINARY_NAME=qwiklip
export BUILD_DIR=bin
export PORT=8080

echo "Running integration tests..."

# Build the server
mask build

# Start server in background
echo "Starting server on port $PORT..."
$BUILD_DIR/$BINARY_NAME &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Test health endpoint
echo "Testing health endpoint..."
HEALTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$PORT/health)
if [ "$HEALTH_STATUS" -eq 200 ]; then
    echo "✅ Health check passed (status: $HEALTH_STATUS)"
else
    echo "❌ Health check failed (status: $HEALTH_STATUS)"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# Test root endpoint (should return HTML)
echo "Testing root endpoint..."
ROOT_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$PORT/)
if [ "$ROOT_STATUS" -eq 200 ]; then
    echo "✅ Root endpoint passed (status: $ROOT_STATUS)"
else
    echo "❌ Root endpoint failed (status: $ROOT_STATUS)"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# Test 404 endpoint
echo "Testing 404 endpoint..."
NOTFOUND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$PORT/nonexistent)
if [ "$NOTFOUND_STATUS" -eq 404 ]; then
    echo "✅ 404 endpoint passed (status: $NOTFOUND_STATUS)"
else
    echo "❌ 404 endpoint failed (status: $NOTFOUND_STATUS)"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# Test invalid reel endpoint (should return error)
echo "Testing invalid reel endpoint..."
INVALID_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$PORT/reel/invalid)
if [ "$INVALID_STATUS" -eq 500 ] || [ "$INVALID_STATUS" -eq 400 ]; then
    echo "✅ Invalid reel endpoint handled correctly (status: $INVALID_STATUS)"
else
    echo "❌ Invalid reel endpoint failed (status: $INVALID_STATUS)"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# Stop server
echo "Stopping server..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo "🎉 All integration tests passed!"
```

## security

> Run security checks

```bash
echo "Running security checks..."
if command -v gosec >/dev/null 2>&1; then
    gosec ./...
else
    echo "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
fi
```

## help

> Show available commands

```bash
echo "Available commands:"
echo "  all             - Clean, lint, test, and build"
echo "  build           - Build the binary"
echo "  build-all       - Build for multiple platforms"
echo "  test            - Run unit tests"
echo "  integration-test - Run integration tests with HTTP calls"
echo "  coverage        - Run tests with coverage report"
echo "  run             - Run the application"
echo "  run-debug       - Run with debug logging"
echo "  clean           - Clean build artifacts"
echo "  fmt             - Format code"
echo "  lint            - Run linter"
echo "  vet             - Run go vet"
echo "  deps            - Download dependencies"
echo "  deps-update     - Update dependencies"
echo "  docker-build    - Build Docker image"
echo "  docker-run      - Run Docker container"
echo "  docker-run-debug - Run Docker container with debug"
echo "  dev-setup       - Setup development environment"
echo "  mocks           - Generate mocks"
echo "  bench           - Run benchmarks"
echo "  security        - Run security checks"
echo "  help            - Show this help message"
```
