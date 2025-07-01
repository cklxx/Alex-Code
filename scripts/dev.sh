#!/bin/bash

# Development workflow script for Deep Coding Agent
# This script provides common development tasks

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.18 or later."
        exit 1
    fi
    
    GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
    print_status "Using Go version: $GO_VERSION"
}

# Format code
format() {
    print_status "Formatting Go code..."
    go fmt ./...
    print_success "Code formatting completed"
}

# Vet code
vet() {
    print_status "Vetting Go code..."
    go vet ./...
    print_success "Code vetting completed"
}

# Run tests
test() {
    print_status "Running tests..."
    go test -v ./...
    print_success "All tests passed"
}

# Run tests with coverage
test_coverage() {
    print_status "Running tests with coverage..."
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    print_success "Coverage report generated: coverage.html"
}

# Build the binary
build() {
    print_status "Building binary..."
    go build -o deep-coding-agent cmd/simple-main.go
    print_success "Binary built: deep-coding-agent"
}

# Build for multiple platforms
build_all() {
    print_status "Building for multiple platforms..."
    mkdir -p build
    
    platforms=("linux/amd64" "darwin/amd64" "darwin/arm64" "windows/amd64")
    
    for platform in "${platforms[@]}"; do
        platform_split=(${platform//\// })
        GOOS=${platform_split[0]}
        GOARCH=${platform_split[1]}
        
        output_name="deep-coding-agent-$GOOS-$GOARCH"
        if [ $GOOS = "windows" ]; then
            output_name+='.exe'
        fi
        
        print_status "Building for $GOOS/$GOARCH..."
        env GOOS=$GOOS GOARCH=$GOARCH go build -o build/$output_name cmd/simple-main.go
    done
    
    print_success "Multi-platform build completed in build/ directory"
}

# Clean build artifacts
clean() {
    print_status "Cleaning build artifacts..."
    rm -f deep-coding-agent
    rm -rf build/
    rm -f coverage.out coverage.html
    print_success "Clean completed"
}

# Install binary
install() {
    build
    print_status "Installing binary to GOPATH/bin..."
    
    if [ -z "$GOPATH" ]; then
        GOBIN=$(go env GOPATH)/bin
    else
        GOBIN=$GOPATH/bin
    fi
    
    cp deep-coding-agent "$GOBIN/"
    print_success "Binary installed to $GOBIN/deep-coding-agent"
}

# Run linting (if golangci-lint is available)
lint() {
    if command -v golangci-lint &> /dev/null; then
        print_status "Running golangci-lint..."
        golangci-lint run
        print_success "Linting completed"
    else
        print_warning "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
        print_status "Running basic go vet instead..."
        vet
    fi
}

# Quick functionality test
test_functionality() {
    if [ ! -f "deep-coding-agent" ]; then
        build
    fi
    
    print_status "Testing core functionality..."
    
    # Test version command
    ./deep-coding-agent version
    
    # Test config command
    ./deep-coding-agent config --list > /dev/null
    
    # Test analyze command on self
    ./deep-coding-agent analyze cmd/simple-main.go > /dev/null
    
    # Test AI status
    ./deep-coding-agent ai-status > /dev/null
    
    print_success "Functionality test completed"
}

# Development workflow
dev() {
    check_go
    format
    vet
    test
    build
    test_functionality
    print_success "Development workflow completed successfully!"
}

# Setup development environment
setup() {
    check_go
    
    print_status "Setting up development environment..."
    
    # Install golangci-lint if not present
    if ! command -v golangci-lint &> /dev/null; then
        print_status "Installing golangci-lint..."
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    fi
    
    # Create .vscode settings if directory exists
    if [ -d ".vscode" ]; then
        cat > .vscode/settings.json << EOF
{
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.testFlags": ["-v"],
    "go.coverOnSave": true,
    "go.coverageDecorator": {
        "type": "gutter"
    }
}
EOF
        print_status "Created VS Code settings"
    fi
    
    # Run initial build
    build
    
    print_success "Development environment setup completed!"
}

# Watch for file changes and run tests (requires fswatch)
watch() {
    if ! command -v fswatch &> /dev/null; then
        print_error "fswatch not found. Install it with: brew install fswatch (macOS) or apt-get install fswatch (Linux)"
        exit 1
    fi
    
    print_status "Watching for file changes..."
    print_warning "Press Ctrl+C to stop watching"
    
    fswatch -r --exclude='\.git' --exclude='build' --exclude='coverage' . | while read change; do
        if [[ $change == *.go ]]; then
            print_status "File changed: $change"
            if test; then
                print_success "Tests passed"
            else
                print_error "Tests failed"
            fi
        fi
    done
}

# Show help
help() {
    echo "Deep Coding Agent Development Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  dev             Run complete development workflow (format, vet, test, build)"
    echo "  setup           Setup development environment"
    echo "  format          Format Go code"
    echo "  vet             Vet Go code"
    echo "  lint            Run linter (golangci-lint)"
    echo "  test            Run tests"
    echo "  test-coverage   Run tests with coverage report"
    echo "  build           Build binary"
    echo "  build-all       Build for multiple platforms"
    echo "  install         Install binary to GOPATH/bin"
    echo "  clean           Clean build artifacts"
    echo "  test-func       Quick functionality test"
    echo "  watch           Watch for changes and run tests"
    echo "  help            Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 dev          # Full development workflow"
    echo "  $0 test         # Run tests only"
    echo "  $0 build-all    # Build for all platforms"
}

# Main command dispatcher
case "${1:-help}" in
    dev)
        dev
        ;;
    setup)
        setup
        ;;
    format)
        format
        ;;
    vet)
        vet
        ;;
    lint)
        lint
        ;;
    test)
        test
        ;;
    test-coverage)
        test_coverage
        ;;
    build)
        build
        ;;
    build-all)
        build_all
        ;;
    install)
        install
        ;;
    clean)
        clean
        ;;
    test-func)
        test_functionality
        ;;
    watch)
        watch
        ;;
    help|--help|-h)
        help
        ;;
    *)
        print_error "Unknown command: $1"
        echo ""
        help
        exit 1
        ;;
esac