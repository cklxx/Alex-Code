# Multi-stage build for Deep Coding Agent
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go modules files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o deep-coding-agent \
    cmd/simple-main.go

# Final stage - minimal image
FROM scratch

# Copy timezone data and certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /app/deep-coding-agent /usr/local/bin/deep-coding-agent

# Create a non-root user (scratch doesn't have useradd, so we copy from alpine)
FROM alpine:latest AS user-builder
RUN adduser -D -s /bin/sh appuser
FROM scratch
COPY --from=user-builder /etc/passwd /etc/passwd
COPY --from=user-builder /etc/group /etc/group

# Copy binary and set up environment
COPY --from=builder /app/deep-coding-agent /usr/local/bin/deep-coding-agent

# Set user
USER appuser

# Set working directory
WORKDIR /workspace

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/deep-coding-agent", "version"]

# Default command
ENTRYPOINT ["/usr/local/bin/deep-coding-agent"]
CMD ["--help"]

# Labels for metadata
LABEL maintainer="Deep Coding Agent Team" \
      description="High-performance CLI tool for intelligent code analysis, generation, and refactoring" \
      version="latest" \
      org.opencontainers.image.title="Deep Coding Agent" \
      org.opencontainers.image.description="AI-powered code analysis and generation tool" \
      org.opencontainers.image.vendor="Deep Coding Agent" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/your-org/deep-coding" \
      org.opencontainers.image.documentation="https://github.com/your-org/deep-coding/blob/main/README.md"