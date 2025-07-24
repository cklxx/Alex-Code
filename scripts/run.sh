#!/bin/bash

# Universal run script for Deep Coding Agent
# This script provides a unified interface to run the agent in different ways

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

# Detect best way to run the agent
detect_runtime() {
    # Check for compiled binary
    if [ -f "deep-coding-agent" ] && [ -x "deep-coding-agent" ]; then
        echo "binary"
        return
    fi
    
    # Check for Docker
    if command -v docker &> /dev/null && docker image inspect deep-coding-agent:latest &> /dev/null; then
        echo "docker"
        return
    fi
    
    # Check for Go
    if command -v go &> /dev/null && [ -f "cmd/simple-main.go" ]; then
        echo "go"
        return
    fi
    
    # Check for container in docker-compose
    if command -v docker-compose &> /dev/null && [ -f "docker-compose.yml" ]; then
        echo "compose"
        return
    fi
    
    echo "none"
}

# Run using compiled binary
run_binary() {
    print_status "Running with compiled binary"
    ./deep-coding-agent "$@"
}

# Run using Go
run_go() {
    print_status "Running with Go"
    go run cmd/simple-main.go "$@"
}

# Run using Docker
run_docker() {
    print_status "Running with Docker"
    docker run --rm -it \
        -v "$(pwd):/workspace:ro" \
        -v "alex-config:/home/appuser/.config" \
        deep-coding-agent:latest "$@"
}

# Run using docker-compose
run_compose() {
    print_status "Running with docker-compose"
    docker-compose run --rm deep-coding-agent "$@"
}

# Auto-build if needed
auto_build() {
    local runtime="$1"
    
    case "$runtime" in
        binary)
            if [ ! -f "deep-coding-agent" ] || [ "cmd/simple-main.go" -nt "deep-coding-agent" ]; then
                print_status "Building binary..."
                go build -o deep-coding-agent cmd/simple-main.go
                print_success "Binary built successfully"
            fi
            ;;
        docker)
            if ! docker image inspect deep-coding-agent:latest &> /dev/null; then
                print_status "Building Docker image..."
                docker build -t deep-coding-agent:latest .
                print_success "Docker image built successfully"
            fi
            ;;
        compose)
            print_status "Using docker-compose (will build if needed)..."
            docker-compose build deep-coding-agent
            ;;
    esac
}

# Quick development run (with hot reload)
dev_run() {
    print_status "Starting development mode with hot reload..."
    
    if command -v air &> /dev/null; then
        air -c .air.toml -- "$@"
    elif [ -f ".air.toml" ]; then
        print_warning "Air not found, installing..."
        go install github.com/air-verse/air@latest
        air -c .air.toml -- "$@"
    else
        print_warning "Hot reload not available, using go run"
        go run cmd/simple-main.go "$@"
    fi
}

# Performance monitoring run
perf_run() {
    print_status "Running with performance monitoring..."
    
    local runtime=$(detect_runtime)
    auto_build "$runtime"
    
    # Time the execution
    start_time=$(date +%s%N)
    
    case "$runtime" in
        binary) run_binary "$@" ;;
        go) run_go "$@" ;;
        docker) run_docker "$@" ;;
        compose) run_compose "$@" ;;
        *) 
            print_error "No suitable runtime found"
            exit 1
            ;;
    esac
    
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))
    
    print_success "Execution completed in ${duration}ms"
}

# Debug run with verbose output
debug_run() {
    print_status "Running in debug mode..."
    
    export DEBUG=1
    export VERBOSE=1
    
    # Add debug flags if not present
    local args=("$@")
    local has_verbose=false
    
    for arg in "${args[@]}"; do
        if [[ "$arg" == "-v" || "$arg" == "--verbose" ]]; then
            has_verbose=true
            break
        fi
    done
    
    if [ "$has_verbose" = false ]; then
        args=("-v" "${args[@]}")
    fi
    
    local runtime=$(detect_runtime)
    auto_build "$runtime"
    
    case "$runtime" in
        binary) run_binary "${args[@]}" ;;
        go) run_go "${args[@]}" ;;
        docker) run_docker "${args[@]}" ;;
        compose) run_compose "${args[@]}" ;;
        *) 
            print_error "No suitable runtime found"
            exit 1
            ;;
    esac
}

# Profile run for memory/CPU analysis
profile_run() {
    print_status "Running with profiling enabled..."
    
    if [ ! -d "profiles" ]; then
        mkdir profiles
    fi
    
    # Only works with Go runtime
    if ! command -v go &> /dev/null; then
        print_error "Profiling requires Go runtime"
        exit 1
    fi
    
    local profile_type="${1:-cpu}"
    shift
    
    case "$profile_type" in
        cpu)
            print_status "CPU profiling enabled"
            go run -ldflags="-X main.EnableCPUProfile=true" cmd/simple-main.go "$@"
            ;;
        mem|memory)
            print_status "Memory profiling enabled"
            go run -ldflags="-X main.EnableMemProfile=true" cmd/simple-main.go "$@"
            ;;
        *)
            print_error "Unknown profile type: $profile_type (use cpu or memory)"
            exit 1
            ;;
    esac
    
    print_success "Profile saved to profiles/ directory"
}

# Benchmark run
benchmark_run() {
    print_status "Running benchmark tests..."
    
    # Create test data
    if [ ! -d "benchmark_data" ]; then
        mkdir benchmark_data
        for i in {1..100}; do
            cat > "benchmark_data/test$i.go" << EOF
package main

import (
    "fmt"
    "time"
)

func main() {
    start := time.Now()
    result := fibonacci($((i % 40 + 1)))
    duration := time.Since(start)
    fmt.Printf("fibonacci($((i % 40 + 1))) = %d (took %v)\n", result, duration)
}

func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}
EOF
        done
        print_status "Created benchmark data with 100 files"
    fi
    
    local runtime=$(detect_runtime)
    auto_build "$runtime"
    
    print_status "Benchmarking analysis performance..."
    
    # Warm up
    case "$runtime" in
        binary) ./deep-coding-agent analyze benchmark_data/test1.go > /dev/null ;;
        go) go run cmd/simple-main.go analyze benchmark_data/test1.go > /dev/null ;;
        docker) docker run --rm -v "$(pwd):/workspace:ro" deep-coding-agent:latest analyze /workspace/benchmark_data/test1.go > /dev/null ;;
    esac
    
    # Benchmark different scenarios
    local scenarios=(
        "Single file analysis"
        "Directory analysis (sequential)"
        "Directory analysis (concurrent)"
        "Code generation"
    )
    
    local commands=(
        "analyze benchmark_data/test1.go"
        "analyze benchmark_data/"
        "analyze benchmark_data/ --concurrent"
        "generate 'simple calculator' go"
    )
    
    for i in "${!scenarios[@]}"; do
        scenario="${scenarios[$i]}"
        command="${commands[$i]}"
        
        print_status "Benchmarking: $scenario"
        
        local total_time=0
        local runs=5
        
        for run in $(seq 1 $runs); do
            start_time=$(date +%s%N)
            
            case "$runtime" in
                binary) ./deep-coding-agent $command > /dev/null ;;
                go) go run cmd/simple-main.go $command > /dev/null ;;
                docker) docker run --rm -v "$(pwd):/workspace:ro" deep-coding-agent:latest $command > /dev/null ;;
            esac
            
            end_time=$(date +%s%N)
            duration=$(( (end_time - start_time) / 1000000 ))
            total_time=$((total_time + duration))
            
            echo "  Run $run: ${duration}ms"
        done
        
        avg_time=$((total_time / runs))
        print_success "$scenario average: ${avg_time}ms"
        echo ""
    done
    
    # Cleanup
    rm -rf benchmark_data/
}

# Show available runtimes
show_runtimes() {
    print_status "Available runtimes:"
    echo ""
    
    # Binary
    if [ -f "deep-coding-agent" ] && [ -x "deep-coding-agent" ]; then
        print_success "✓ Compiled binary (deep-coding-agent)"
    else
        print_warning "✗ Compiled binary (not found)"
    fi
    
    # Go
    if command -v go &> /dev/null && [ -f "cmd/simple-main.go" ]; then
        print_success "✓ Go runtime ($(go version | cut -d' ' -f3))"
    else
        print_warning "✗ Go runtime (not available)"
    fi
    
    # Docker
    if command -v docker &> /dev/null; then
        if docker image inspect deep-coding-agent:latest &> /dev/null; then
            print_success "✓ Docker image (deep-coding-agent:latest)"
        else
            print_warning "✗ Docker image (not built)"
        fi
    else
        print_warning "✗ Docker (not installed)"
    fi
    
    # Docker Compose
    if command -v docker-compose &> /dev/null && [ -f "docker-compose.yml" ]; then
        print_success "✓ Docker Compose"
    else
        print_warning "✗ Docker Compose (not available)"
    fi
    
    echo ""
    local current_runtime=$(detect_runtime)
    if [ "$current_runtime" != "none" ]; then
        print_status "Current default runtime: $current_runtime"
    else
        print_error "No suitable runtime detected"
    fi
}

# Show help
show_help() {
    cat << EOF
Deep Coding Agent Universal Run Script

Usage: $0 [mode] [arguments]

Modes:
    run [args]              Auto-detect and run with best available runtime
    dev [args]              Development mode with hot reload
    perf [args]             Performance monitoring mode
    debug [args]            Debug mode with verbose output
    profile [type] [args]   Profiling mode (cpu|memory)
    benchmark               Run performance benchmarks
    
    binary [args]           Force run with compiled binary
    go [args]               Force run with Go
    docker [args]           Force run with Docker
    compose [args]          Force run with docker-compose
    
    runtimes                Show available runtimes
    help                    Show this help

Examples:
    $0 run analyze .                    # Auto-run analysis
    $0 dev analyze . --ai              # Development mode with hot reload
    $0 perf generate "API" go           # Performance monitoring
    $0 debug refactor main.go           # Debug mode
    $0 profile cpu analyze large/       # CPU profiling
    $0 benchmark                        # Run benchmarks
    
    $0 docker analyze /workspace        # Force Docker runtime
    $0 go --help                        # Force Go runtime

The script automatically detects the best available runtime in this order:
1. Compiled binary (fastest)
2. Docker image (containerized)
3. Go runtime (development)
4. Docker Compose (fallback)

EOF
}

# Main command dispatcher
case "${1:-run}" in
    run)
        runtime=$(detect_runtime)
        if [ "$runtime" = "none" ]; then
            print_error "No suitable runtime found. Please build the binary or install Go/Docker."
            exit 1
        fi
        auto_build "$runtime"
        case "$runtime" in
            binary) run_binary "${@:2}" ;;
            go) run_go "${@:2}" ;;
            docker) run_docker "${@:2}" ;;
            compose) run_compose "${@:2}" ;;
        esac
        ;;
    dev)
        dev_run "${@:2}"
        ;;
    perf)
        perf_run "${@:2}"
        ;;
    debug)
        debug_run "${@:2}"
        ;;
    profile)
        profile_run "${2:-cpu}" "${@:3}"
        ;;
    benchmark)
        benchmark_run
        ;;
    binary)
        auto_build binary
        run_binary "${@:2}"
        ;;
    go)
        run_go "${@:2}"
        ;;
    docker)
        auto_build docker
        run_docker "${@:2}"
        ;;
    compose)
        auto_build compose
        run_compose "${@:2}"
        ;;
    runtimes)
        show_runtimes
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        # If first argument looks like a deep-coding-agent command, run it
        if [[ "$1" =~ ^(analyze|generate|refactor|config|explain|ai-status|version|help)$ ]]; then
            runtime=$(detect_runtime)
            if [ "$runtime" = "none" ]; then
                print_error "No suitable runtime found"
                exit 1
            fi
            auto_build "$runtime"
            case "$runtime" in
                binary) run_binary "$@" ;;
                go) run_go "$@" ;;
                docker) run_docker "$@" ;;
                compose) run_compose "$@" ;;
            esac
        else
            print_error "Unknown mode: $1"
            echo ""
            show_help
            exit 1
        fi
        ;;
esac