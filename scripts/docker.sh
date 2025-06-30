#!/bin/bash

# Docker management script for Deep Coding Agent
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

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

# Build Docker images
build() {
    local target="${1:-production}"
    
    print_status "Building Docker image for target: $target"
    
    case "$target" in
        production|prod)
            docker build -t deep-coding-agent:latest .
            print_success "Production image built: deep-coding-agent:latest"
            ;;
        development|dev)
            docker build -f Dockerfile.dev --target development -t deep-coding-agent:dev .
            print_success "Development image built: deep-coding-agent:dev"
            ;;
        test)
            docker build -f Dockerfile.dev --target testing -t deep-coding-agent:test .
            print_success "Test image built: deep-coding-agent:test"
            ;;
        all)
            build production
            build development
            build test
            ;;
        *)
            print_error "Unknown target: $target"
            exit 1
            ;;
    esac
}

# Run container
run() {
    local service="${1:-deep-coding-agent}"
    shift
    
    print_status "Running service: $service"
    
    case "$service" in
        prod|production)
            docker run --rm -it \
                -v "$(pwd):/workspace:ro" \
                -v "deep-coding-config:/home/appuser/.config" \
                deep-coding-agent:latest "$@"
            ;;
        dev|development)
            docker run --rm -it \
                -v "$(pwd):/app" \
                -v "go-modules:/go/pkg/mod" \
                -v "go-build-cache:/root/.cache/go-build" \
                -p 8080:8080 \
                deep-coding-agent:dev "$@"
            ;;
        test)
            docker run --rm \
                -v "$(pwd):/app" \
                -v "test-results:/app/test-results" \
                deep-coding-agent:test "$@"
            ;;
        *)
            print_error "Unknown service: $service"
            exit 1
            ;;
    esac
}

# Use docker-compose
compose() {
    local action="$1"
    shift
    
    case "$action" in
        up)
            print_status "Starting services with docker-compose"
            docker-compose up "$@"
            ;;
        down)
            print_status "Stopping services"
            docker-compose down "$@"
            ;;
        build)
            print_status "Building services with docker-compose"
            docker-compose build "$@"
            ;;
        logs)
            docker-compose logs "$@"
            ;;
        exec)
            docker-compose exec "$@"
            ;;
        run)
            docker-compose run --rm "$@"
            ;;
        *)
            print_status "Running docker-compose $action"
            docker-compose "$action" "$@"
            ;;
    esac
}

# Development workflow
dev() {
    print_status "Starting development environment"
    
    # Build development image if it doesn't exist
    if ! docker image inspect deep-coding-agent:dev &> /dev/null; then
        build dev
    fi
    
    # Start development container with hot reload
    docker-compose up dev
}

# Run tests in container
test() {
    local test_type="${1:-all}"
    
    print_status "Running tests in container: $test_type"
    
    # Build test image if it doesn't exist
    if ! docker image inspect deep-coding-agent:test &> /dev/null; then
        build test
    fi
    
    case "$test_type" in
        unit|integration|performance|coverage|memory|stress|all)
            docker-compose run --rm test ./scripts/test.sh "$test_type"
            ;;
        *)
            print_error "Unknown test type: $test_type"
            exit 1
            ;;
    esac
}

# Performance benchmark
benchmark() {
    print_status "Running performance benchmark"
    
    # Create test data if it doesn't exist
    if [ ! -d "test/performance" ]; then
        mkdir -p test/performance
        for i in {1..50}; do
            cat > "test/performance/test$i.go" << EOF
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}

func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}
EOF
        done
    fi
    
    # Run performance test
    docker-compose run --rm performance
    
    # Show results
    if [ -d "performance-results" ]; then
        print_success "Performance results:"
        cat performance-results/* 2>/dev/null || echo "No results found"
    fi
}

# Clean up Docker resources
clean() {
    local level="${1:-basic}"
    
    case "$level" in
        basic)
            print_status "Cleaning up containers and networks"
            docker-compose down -v --remove-orphans
            ;;
        images)
            print_status "Cleaning up images"
            docker rmi deep-coding-agent:latest deep-coding-agent:dev deep-coding-agent:test 2>/dev/null || true
            ;;
        all)
            print_status "Cleaning up all Docker resources"
            docker-compose down -v --remove-orphans
            docker rmi deep-coding-agent:latest deep-coding-agent:dev deep-coding-agent:test 2>/dev/null || true
            docker volume prune -f
            docker network prune -f
            ;;
        *)
            print_error "Unknown clean level: $level"
            exit 1
            ;;
    esac
    
    print_success "Cleanup completed"
}

# Show container status
status() {
    print_status "Container status:"
    docker-compose ps
    
    echo ""
    print_status "Images:"
    docker images | grep deep-coding-agent || echo "No deep-coding-agent images found"
    
    echo ""
    print_status "Volumes:"
    docker volume ls | grep deep-coding || echo "No deep-coding volumes found"
}

# Shell into container
shell() {
    local service="${1:-dev}"
    
    print_status "Opening shell in $service container"
    
    case "$service" in
        dev|development)
            docker-compose run --rm dev /bin/bash
            ;;
        prod|production)
            # Create a temporary container with shell
            docker run --rm -it \
                --entrypoint /bin/sh \
                -v "$(pwd):/workspace:ro" \
                deep-coding-agent:latest
            ;;
        *)
            print_error "Unknown service: $service"
            exit 1
            ;;
    esac
}

# Analyze current directory
analyze() {
    print_status "Analyzing current directory with containerized agent"
    
    run prod analyze /workspace "$@"
}

# Generate code
generate() {
    local spec="$1"
    local lang="${2:-go}"
    shift 2
    
    if [ -z "$spec" ]; then
        print_error "Please provide a code specification"
        exit 1
    fi
    
    print_status "Generating $lang code: $spec"
    
    run prod generate "$spec" "$lang" "$@"
}

# Show help
help() {
    cat << EOF
Deep Coding Agent Docker Management Script

Usage: $0 [command] [options]

Commands:
    build [target]          Build Docker image (production|dev|test|all)
    run [service] [args]    Run container (prod|dev|test)
    compose [action]        Docker-compose operations (up|down|build|logs)
    
    dev                     Start development environment
    test [type]             Run tests (unit|integration|performance|all)
    benchmark               Run performance benchmark
    
    clean [level]           Clean up (basic|images|all)
    status                  Show container status
    shell [service]         Open shell in container (dev|prod)
    
    analyze [args]          Analyze code using container
    generate <spec> [lang]  Generate code using container
    
    help                    Show this help

Examples:
    $0 build all                    # Build all images
    $0 dev                          # Start development environment
    $0 test unit                    # Run unit tests
    $0 analyze --concurrent         # Analyze current directory
    $0 generate "REST API" go       # Generate Go REST API
    $0 shell dev                    # Open shell in dev container

Environment Variables:
    DOCKER_BUILDKIT=1              # Enable BuildKit for faster builds
    COMPOSE_DOCKER_CLI_BUILD=1     # Use Docker CLI for compose builds

EOF
}

# Main command dispatcher
case "${1:-help}" in
    build)
        build "${2:-production}"
        ;;
    run)
        run "${2:-production}" "${@:3}"
        ;;
    compose)
        compose "${2:-up}" "${@:3}"
        ;;
    dev)
        dev
        ;;
    test)
        test "${2:-all}"
        ;;
    benchmark)
        benchmark
        ;;
    clean)
        clean "${2:-basic}"
        ;;
    status)
        status
        ;;
    shell)
        shell "${2:-dev}"
        ;;
    analyze)
        analyze "${@:2}"
        ;;
    generate)
        generate "${2}" "${3:-go}" "${@:4}"
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