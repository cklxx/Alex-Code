#!/bin/bash

# Comprehensive testing script for Deep Coding Agent
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

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Unit tests
run_unit_tests() {
    print_status "Running unit tests..."
    go test -v -count=1 ./internal/... ./pkg/...
    print_success "Unit tests completed"
}

# Integration tests
run_integration_tests() {
    print_status "Running integration tests..."
    
    # Build the binary first
    go build -o deep-coding-agent cmd/simple-main.go
    
    # Test CLI commands
    print_status "Testing CLI commands..."
    
    # Test version
    ./deep-coding-agent version > /dev/null
    print_status "✓ Version command works"
    
    # Test config commands
    ./deep-coding-agent config --list > /dev/null
    print_status "✓ Config list works"
    
    ./deep-coding-agent config --set testKey=testValue > /dev/null
    ./deep-coding-agent config --get testKey > /dev/null
    print_status "✓ Config set/get works"
    
    # Test analyze command
    ./deep-coding-agent analyze cmd/simple-main.go > /dev/null
    print_status "✓ Analyze file works"
    
    ./deep-coding-agent analyze internal/ --depth=1 > /dev/null
    print_status "✓ Analyze directory works"
    
    # Test generate command
    ./deep-coding-agent generate "test function" go > /dev/null
    print_status "✓ Generate command works"
    
    # Test AI status
    ./deep-coding-agent ai-status > /dev/null
    print_status "✓ AI status works"
    
    # Test refactor (dry run)
    echo "var test = 'hello';" > test_temp.js
    ./deep-coding-agent refactor test_temp.js --pattern=modernize --dry-run > /dev/null
    rm -f test_temp.js test_temp.js.backup
    print_status "✓ Refactor dry run works"
    
    print_success "Integration tests completed"
}

# Performance tests
run_performance_tests() {
    print_status "Running performance tests..."
    
    if [ ! -f "deep-coding-agent" ]; then
        go build -o deep-coding-agent cmd/simple-main.go
    fi
    
    # Create test files for performance testing
    mkdir -p test_perf
    for i in {1..10}; do
        cat > test_perf/test$i.go << EOF
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

func processData(data []string) []string {
    var result []string
    for _, item := range data {
        if len(item) > 0 {
            result = append(result, item)
        }
    }
    return result
}
EOF
    done
    
    # Test analysis performance
    print_status "Testing analysis performance..."
    start_time=$(date +%s%N)
    ./deep-coding-agent analyze test_perf/ --concurrent > /dev/null
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))
    print_status "Analysis of 10 files took ${duration}ms"
    
    if [ $duration -gt 1000 ]; then
        print_error "Performance test failed: Analysis took too long (${duration}ms > 1000ms)"
        exit 1
    fi
    
    # Test generation performance
    start_time=$(date +%s%N)
    ./deep-coding-agent generate "simple calculator" go > /dev/null
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))
    print_status "Code generation took ${duration}ms"
    
    if [ $duration -gt 500 ]; then
        print_error "Performance test failed: Generation took too long (${duration}ms > 500ms)"
        exit 1
    fi
    
    # Cleanup
    rm -rf test_perf/
    
    print_success "Performance tests completed"
}

# Test with coverage
run_coverage_tests() {
    print_status "Running tests with coverage..."
    
    go test -coverprofile=coverage.out ./...
    coverage=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
    
    print_status "Test coverage: ${coverage}%"
    
    if (( $(echo "$coverage < 70" | bc -l) )); then
        print_error "Coverage too low: ${coverage}% < 70%"
        exit 1
    fi
    
    go tool cover -html=coverage.out -o coverage.html
    print_success "Coverage report generated: coverage.html"
}

# Memory leak tests
run_memory_tests() {
    print_status "Running memory tests..."
    
    if [ ! -f "deep-coding-agent" ]; then
        go build -o deep-coding-agent cmd/simple-main.go
    fi
    
    # Create larger test directory
    mkdir -p test_memory
    for i in {1..50}; do
        cat > test_memory/large$i.go << EOF
package main

import (
    "fmt"
    "os"
    "time"
    "strings"
)

func main() {
    data := make([]string, 1000)
    for i := range data {
        data[i] = fmt.Sprintf("item_%d_%s", i, strings.Repeat("x", 100))
    }
    
    for j := 0; j < 100; j++ {
        processLargeData(data)
        time.Sleep(1 * time.Millisecond)
    }
}

func processLargeData(data []string) {
    result := make(map[string]int)
    for _, item := range data {
        result[item] = len(item)
    }
}

$(for k in {1..20}; do
echo "func function$k() {"
echo "    fmt.Println(\"Function $k\")"
echo "}"
done)
EOF
    done
    
    # Test memory usage during analysis
    if command -v valgrind &> /dev/null; then
        print_status "Running valgrind memory check..."
        valgrind --leak-check=full --error-exitcode=1 ./deep-coding-agent analyze test_memory/ --concurrent > /dev/null 2>&1
        print_status "✓ No memory leaks detected"
    else
        print_status "Valgrind not available, skipping detailed memory leak detection"
        # Just run the command and check it doesn't crash
        ./deep-coding-agent analyze test_memory/ --concurrent > /dev/null
        print_status "✓ Large directory analysis completed without crashes"
    fi
    
    # Cleanup
    rm -rf test_memory/
    
    print_success "Memory tests completed"
}

# Stress tests
run_stress_tests() {
    print_status "Running stress tests..."
    
    if [ ! -f "deep-coding-agent" ]; then
        go build -o deep-coding-agent cmd/simple-main.go
    fi
    
    # Test rapid command execution
    print_status "Testing rapid command execution..."
    for i in {1..20}; do
        ./deep-coding-agent version > /dev/null &
    done
    wait
    print_status "✓ Rapid execution test passed"
    
    # Test with very long input
    print_status "Testing with long input..."
    long_spec=$(printf 'a%.0s' {1..1000})
    ./deep-coding-agent generate "$long_spec" go > /dev/null
    print_status "✓ Long input test passed"
    
    # Test concurrent analysis
    print_status "Testing concurrent operations..."
    mkdir -p stress_test
    for i in {1..5}; do
        echo "package main; func test$i() {}" > stress_test/test$i.go
    done
    
    # Run multiple analyses concurrently
    for i in {1..5}; do
        ./deep-coding-agent analyze stress_test/ > /dev/null &
    done
    wait
    
    rm -rf stress_test/
    print_status "✓ Concurrent operations test passed"
    
    print_success "Stress tests completed"
}

# Run specific test type
case "${1:-all}" in
    unit)
        run_unit_tests
        ;;
    integration)
        run_integration_tests
        ;;
    performance)
        run_performance_tests
        ;;
    coverage)
        run_coverage_tests
        ;;
    memory)
        run_memory_tests
        ;;
    stress)
        run_stress_tests
        ;;
    all)
        print_status "Running comprehensive test suite..."
        run_unit_tests
        run_integration_tests
        run_performance_tests
        run_coverage_tests
        run_memory_tests
        run_stress_tests
        print_success "All tests completed successfully!"
        ;;
    *)
        echo "Usage: $0 [unit|integration|performance|coverage|memory|stress|all]"
        exit 1
        ;;
esac