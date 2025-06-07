#!/bin/bash

# shellcheck disable=SC2155

# Profiling helper script for Greenlight API
# Usage: ./scripts/profile.sh [cpu|heap|goroutine|allocs|block|mutex|trace]

set -e

PROFILING_HOST="localhost:5000" # port can be changed to port of choice
PROFILE_TYPE="${1:-cpu}"
DURATION="${2:-30}" # value can be changed if you prefer to use make
OUTPUT_DIR="./profiles"

# Create profiles directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if profiling server is running
check_server() {
    if ! curl -s "http://$PROFILING_HOST/health" > /dev/null; then
        print_error "Profiling server is not running on $PROFILING_HOST"
        print_info "Start it with: make run/api/profiling"
        exit 1
    fi
    print_info "Profiling server is running on $PROFILING_HOST"
}

# Generate filename with timestamp
generate_filename() {
    local type="$1"
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    echo "${OUTPUT_DIR}/${type}_${timestamp}.prof"
}

# Capture CPU profile
profile_cpu() {
    local filename=$(generate_filename "cpu")
    print_info "Capturing CPU profile for ${DURATION} seconds..."
    curl -s "http://$PROFILING_HOST/debug/pprof/profile?seconds=$DURATION" -o "$filename"
    print_info "CPU profile saved to: $filename"
    print_info "Analyze with: go tool pprof $filename"
}

# Capture heap profile
profile_heap() {
    local filename=$(generate_filename "heap")
    print_info "Capturing heap profile..."
    curl -s "http://$PROFILING_HOST/debug/pprof/heap" -o "$filename"
    print_info "Heap profile saved to: $filename"
    print_info "Analyze with: go tool pprof $filename"
}

# Capture goroutine profile
profile_goroutine() {
    local filename=$(generate_filename "goroutine")
    print_info "Capturing goroutine profile..."
    curl -s "http://$PROFILING_HOST/debug/pprof/goroutine" -o "$filename"
    print_info "Goroutine profile saved to: $filename"
    print_info "Analyze with: go tool pprof $filename"
}

# Capture allocation profile
profile_allocs() {
    local filename=$(generate_filename "allocs")
    print_info "Capturing allocation profile..."
    curl -s "http://$PROFILING_HOST/debug/pprof/allocs" -o "$filename"
    print_info "Allocation profile saved to: $filename"
    print_info "Analyze with: go tool pprof $filename"
}

# Capture block profile
profile_block() {
    local filename=$(generate_filename "block")
    print_info "Capturing block profile..."
    curl -s "http://$PROFILING_HOST/debug/pprof/block" -o "$filename"
    print_info "Block profile saved to: $filename"
    print_info "Analyze with: go tool pprof $filename"
}

# Capture mutex profile
profile_mutex() {
    local filename=$(generate_filename "mutex")
    print_info "Capturing mutex profile..."
    curl -s "http://$PROFILING_HOST/debug/pprof/mutex" -o "$filename"
    print_info "Mutex profile saved to: $filename"
    print_info "Analyze with: go tool pprof $filename"
}

# Capture trace
profile_trace() {
    local filename=$(generate_filename "trace")
    print_info "Capturing trace for ${DURATION} seconds..."
    curl -s "http://$PROFILING_HOST/debug/pprof/trace?seconds=$DURATION" -o "$filename"
    print_info "Trace saved to: $filename"
    print_info "Analyze with: go tool trace $filename"
}

# Show current metrics
show_metrics() {
    print_info "Current application metrics:"
    curl -s "http://$PROFILING_HOST/debug/vars" | jq . 2>/dev/null || curl -s "http://$PROFILING_HOST/debug/vars"
}

# Show help
show_help() {
    echo "Greenlight API Profiling Helper"
    echo ""
    echo "Usage: $0 [OPTION] [DURATION]"
    echo ""
    echo "Options:"
    echo "  cpu        Capture CPU profile (default: 30s)"
    echo "  heap       Capture heap memory profile"
    echo "  goroutine  Capture goroutine profile"
    echo "  allocs     Capture allocation profile"
    echo "  block      Capture blocking operations profile"
    echo "  mutex      Capture mutex contention profile"
    echo "  trace      Capture execution trace (default: 30s)"
    echo "  metrics    Show current application metrics"
    echo "  all        Capture all profiles"
    echo "  help       Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 cpu 60          # Capture CPU profile for 60 seconds"
    echo "  $0 heap            # Capture heap profile"
    echo "  $0 all             # Capture all available profiles"
    echo ""
    echo "Prerequisites:"
    echo "  - Start the API with profiling: make run/api/profiling"
    echo "  - Ensure curl and jq are installed"
}

# Main execution
case "$PROFILE_TYPE" in
    "cpu")
        check_server
        profile_cpu
        ;;
    "heap")
        check_server
        profile_heap
        ;;
    "goroutine")
        check_server
        profile_goroutine
        ;;
    "allocs")
        check_server
        profile_allocs
        ;;
    "block")
        check_server
        profile_block
        ;;
    "mutex")
        check_server
        profile_mutex
        ;;
    "trace")
        check_server
        profile_trace
        ;;
    "metrics")
        check_server
        show_metrics
        ;;
    "all")
        check_server
        print_info "Capturing all profiles..."
        profile_heap
        profile_goroutine
        profile_allocs
        profile_block
        profile_mutex
        print_info "Starting CPU profile (this will take ${DURATION} seconds)..."
        profile_cpu
        print_info "All profiles captured successfully!"
        ;;
    "help"|"-h"|"--help")
        show_help
        ;;
    *)
        print_error "Unknown profile type: $PROFILE_TYPE"
        show_help
        exit 1
        ;;
esac
