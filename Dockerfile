# Build stage
FROM golang:1.25-alpine AS builder

# Install git and ca-certificates (needed for go modules and HTTPS)
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Extract build metadata (similar to binary-release workflow)
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME

# Build the application with optimizations and embedded metadata
ARG TARGETARCH
RUN set -e && \
    # Extract module information \
    MODULE_NAME=$(grep "^module " go.mod | cut -d' ' -f2) && \
    MAIN_PACKAGE="./cmd/${MODULE_NAME}" && \
    \
    # Set build time if not provided \
    BUILD_TIME=${BUILD_TIME:-$(date -u +%Y-%m-%dT%H:%M:%SZ)} && \
    \
    # Prepare linker flags with embedded metadata \
    LDFLAGS="-w -s -extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" && \
    \
    # Log build information \
    echo "Building ${MODULE_NAME} version ${VERSION}" && \
    echo "📦 Module: ${MODULE_NAME}" && \
    echo "📁 Package: ${MAIN_PACKAGE}" && \
    echo "🏗️  Build flags: ${LDFLAGS}" && \
    \
    # Build the application \
    CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -ldflags="${LDFLAGS}" \
    -a -installsuffix cgo \
    -o ${MODULE_NAME} \
    ${MAIN_PACKAGE}

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -s /bin/sh appuser

# Create app directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/qwiklip .

# Change ownership to non-root user
RUN chown appuser:appuser qwiklip

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Environment variables
ENV PORT=8080
ENV LOG_LEVEL=info
ENV LOG_FORMAT=text

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./qwiklip"]
