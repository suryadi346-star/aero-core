#!/bin/bash
#
# AeroCore AI - Benchmark & Load Testing Script
# Tests: RAM usage, latency, concurrent requests, SSE streaming stability
#
# Usage: ./scripts/bench.sh [iterations] [concurrency]
# Example: ./scripts/bench.sh 100 10
#

set -euo pipefail

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
ITERATIONS="${1:-50}"
CONCURRENCY="${2:-5}"
SESSION_ID="bench_$(date +%s)"
MODEL="${MODEL:-qwen2.5:1.5b-instruct-q4_k_m}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test message
TEST_MESSAGE="Jelaskan secara singkat apa itu AI yang efisien di perangkat terbatas?"

echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  AeroCore AI - Benchmark Suite v0.1.0     ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""

# Function to get current RAM usage (in MB)
get_ram_usage() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        ps aux | grep "[a]ero-core" | awk '{sum+=$6} END {print sum/1024}'
    else
        # Linux
        ps -eo rss,comm | grep "[a]ero-core" | awk '{sum+=$1} END {print sum/1024}'
    fi
}

# Function to measure single request latency
measure_latency() {
    local start_time=$(date +%s%N)

    curl -s -N -X POST "${BASE_URL}/chat/stream" \
        -H "Content-Type: application/json" \
        -d "{
            \"session_id\": \"${SESSION_ID}_latency\",
            \"message\": \"${TEST_MESSAGE}\",
            \"model\": \"${MODEL}\"
        }" > /dev/null

    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
    echo $duration
}

# Function to test SSE streaming
test_sse_stream() {
    echo -e "${YELLOW}Testing SSE Stream...${NC}"

    local response=$(curl -s -N -X POST "${BASE_URL}/chat/stream" \
        -H "Content-Type: application/json" \
        -d "{
            \"session_id\": \"${SESSION_ID}_sse_test\",
            \"message\": \"Halo, apa kabar?\",
            \"model\": \"${MODEL}\"
        }" | head -n 5)

    if echo "$response" | grep -q "event: chunk"; then
        echo -e "${GREEN}✓ SSE streaming works correctly${NC}"
        return 0
    else
        echo -e "${RED}✗ SSE streaming failed${NC}"
        return 1
    fi
}

# Function to test health endpoints
test_health() {
    echo -e "${YELLOW}Testing health endpoints...${NC}"

    # Basic health
    local health=$(curl -s "${BASE_URL}/health")
    if echo "$health" | grep -q "ok"; then
        echo -e "${GREEN}✓ /health endpoint OK${NC}"
    else
        echo -e "${RED}✗ /health endpoint failed${NC}"
        return 1
    fi

    # Deep health
    local deep_health=$(curl -s "${BASE_URL}/health/deep")
    if echo "$deep_health" | grep -q "ok"; then
        echo -e "${GREEN}✓ /health/deep endpoint OK${NC}"
    else
        echo -e "${RED}✗ /health/deep endpoint failed${NC}"
        return 1
    fi
}

# Function to get metrics
get_metrics() {
    echo -e "${YELLOW}Fetching runtime metrics...${NC}"
    curl -s "${BASE_URL}/metrics" | grep -E "go_memstats|go_goroutines"
    echo ""
}

# Function to run concurrent load test
run_load_test() {
    echo -e "${YELLOW}Running load test: ${ITERATIONS} requests, ${CONCURRENCY} concurrent${NC}"

    local total_time=0
    local success_count=0
    local fail_count=0
    local latencies=()

    # Create temp file for parallel execution
    local temp_file=$(mktemp)

    for ((i=1; i<=ITERATIONS; i++)); do
        {
            local start=$(date +%s%N)
            local http_code=$(curl -s -o /dev/null -w "%{http_code}" -N -X POST "${BASE_URL}/chat/stream" \
                -H "Content-Type: application/json" \
                -d "{
                    \"session_id\": \"${SESSION_ID}_${i}\",
                    \"message\": \"${TEST_MESSAGE}\",
                    \"model\": \"${MODEL}\"
                }")
            local end=$(date +%s%N)
            local duration=$(( (end - start) / 1000000 ))

            if [[ "$http_code" == "200" ]]; then
                echo "1 $duration" >> "$temp_file"
            else
                echo "0 $duration" >> "$temp_file"
            fi
        } &

        # Limit concurrency
        if (( i % CONCURRENCY == 0 )); then
            wait
        fi
    done

    wait

    # Calculate statistics
    while read -r success latency; do
        if [[ "$success" == "1" ]]; then
            ((success_count++))
        else
            ((fail_count++))
        fi
        latencies+=("$latency")
        total_time=$((total_time + latency))
    done < "$temp_file"

    rm -f "$temp_file"

    # Calculate averages
    local avg_latency=$((total_time / ITERATIONS))

    # Sort latencies for percentile calculation
    IFS=$'\n' sorted=($(sort -n <<<"${latencies[*]}")); unset IFS
    local median_idx=$((ITERATIONS / 2))
    local p95_idx=$((ITERATIONS * 95 / 100))
    local median_latency=${sorted[$median_idx]}
    local p95_latency=${sorted[$p95_idx]}

    # Print results
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║           Load Test Results                ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "Total Requests:    ${ITERATIONS}"
    echo -e "Successful:        ${GREEN}${success_count}${NC}"
    echo -e "Failed:            ${RED}${fail_count}${NC}"
    echo -e "Success Rate:      $(echo "scale=2; $success_count * 100 / $ITERATIONS" | bc)%"
    echo ""
    echo -e "Average Latency:   ${avg_latency}ms"
    echo -e "Median Latency:    ${median_latency}ms"
    echo -e "P95 Latency:       ${p95_latency}ms"
    echo -e "Total Time:        ${total_time}ms"
    echo ""
}

# Main execution
main() {
    echo -e "${YELLOW}Target: ${BASE_URL}${NC}"
    echo -e "${YELLOW}Model: ${MODEL}${NC}"
    echo ""

    # Initial RAM usage
    local initial_ram=$(get_ram_usage)
    echo -e "${YELLOW}Initial RAM Usage: ${initial_ram} MB${NC}"
    echo ""

    # Run tests
    test_health || exit 1
    echo ""

    test_sse_stream || exit 1
    echo ""

    get_metrics
    echo ""

    run_load_test

    # Final RAM usage
    sleep 2
    local final_ram=$(get_ram_usage)
    echo -e "${YELLOW}Final RAM Usage: ${final_ram} MB${NC}"
    echo -e "${YELLOW}RAM Delta: $((final_ram - initial_ram)) MB${NC}"
    echo ""

    # Memory stats from server
    echo -e "${BLUE}Server Runtime Metrics:${NC}"
    curl -s "${BASE_URL}/metrics" | grep -E "go_memstats_alloc_bytes|go_memstats_sys_bytes|go_goroutines"
    echo ""

    echo -e "${GREEN}✓ Benchmark completed successfully${NC}"
}

# Run main
main
