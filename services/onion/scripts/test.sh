#!/bin/bash

# Test script for the Onion project
# Usage: ./test.sh [option]
# Options:
#   unit      - Run unit tests only (default)
#   coverage  - Run tests with coverage report
#   race      - Run tests with race detection
#   bench     - Run benchmarks
#   all       - Run all test types
#   clean     - Clean test cache and coverage files

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Function to run unit tests
run_unit_tests() {
    print_status "Running unit tests..."
    go test ./... -v
    print_success "Unit tests completed"
}

# Function to run tests with coverage
run_coverage_tests() {
    print_status "Running tests with coverage..."
    
    # Create coverage directory if it doesn't exist
    mkdir -p coverage
    
    # Run tests with coverage
    go test ./... -v -coverprofile=coverage/coverage.out
    
    # Generate HTML coverage report
    go tool cover -html=coverage/coverage.out -o coverage/coverage.html
    
    # Show coverage summary
    go tool cover -func=coverage/coverage.out
    
    print_success "Coverage report generated in coverage/coverage.html"
}

# Function to run tests with race detection
run_race_tests() {
    print_status "Running tests with race detection..."
    go test ./... -v -race
    print_success "Race detection tests completed"
}

# Function to run benchmarks
run_benchmarks() {
    print_status "Running benchmarks..."
    go test ./... -v -bench=. -benchmem
    print_success "Benchmarks completed"
}

# Function to run all tests
run_all_tests() {
    print_status "Running all test types..."
    
    echo
    run_unit_tests
    echo
    
    run_coverage_tests
    echo
    
    run_race_tests
    echo
    
    run_benchmarks
    echo
    
    print_success "All tests completed successfully!"
}

# Function to clean test artifacts
clean_tests() {
    print_status "Cleaning test cache and coverage files..."
    
    # Clean test cache
    go clean -testcache
    
    # Remove coverage directory
    if [ -d "coverage" ]; then
        rm -rf coverage
        print_success "Coverage directory removed"
    fi
    
    # Remove any .out files
    find . -name "*.out" -type f -delete 2>/dev/null || true
    
    print_success "Test cleanup completed"
}

# Function to show help
show_help() {
    echo "Test script for the Onion project"
    echo
    echo "Usage: ./test.sh [option]"
    echo
    echo "Options:"
    echo "  unit      - Run unit tests only (default)"
    echo "  coverage  - Run tests with coverage report"
    echo "  race      - Run tests with race detection"
    echo "  bench     - Run benchmarks"
    echo "  all       - Run all test types"
    echo "  clean     - Clean test cache and coverage files"
    echo "  help      - Show this help message"
    echo
    echo "Examples:"
    echo "  ./test.sh           # Run unit tests"
    echo "  ./test.sh coverage  # Run tests with coverage"
    echo "  ./test.sh all       # Run all test types"
}

# Main script logic
main() {
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check if we're in a Go module
    if [ ! -f "go.mod" ]; then
        print_error "go.mod not found. Please run this script from the project root."
        exit 1
    fi

    # Parse command line arguments
    case "${1:-unit}" in
        "unit")
            run_unit_tests
            ;;
        "coverage")
            run_coverage_tests
            ;;
        "race")
            run_race_tests
            ;;
        "bench")
            run_benchmarks
            ;;
        "all")
            run_all_tests
            ;;
        "clean")
            clean_tests
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_warning "Unknown option: $1"
            echo
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
