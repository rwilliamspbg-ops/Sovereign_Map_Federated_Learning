#!/bin/bash

# WEEK 2 MASTER TEST RUNNER
# Execute all Week 2 tests in optimal order
# Usage: bash run_week2_tests.sh [fast|full|<test_number>]

set -e

TESTS_DIR="."
PYTHON_CMD="python3"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo ""
echo "╔════════════════════════════════════════════════════════════════════════════════╗"
echo "║                     WEEK 2: PRODUCTION READINESS SUITE                        ║"
echo "║                     Real Data | Failures | Partitions | Scale                 ║"
echo "╚════════════════════════════════════════════════════════════════════════════════╝"
echo ""

# Function to run a test
run_test() {
    local test_num=$1
    local test_name=$2
    local test_file=$3
    local est_time=$4
    
    echo -e "${YELLOW}[TEST $test_num] $test_name${NC}"
    echo "  File: $test_file"
    echo "  Est. Time: $est_time"
    echo ""
    
    if [ -f "$test_file" ]; then
        start_time=$(date +%s)
        
        if $PYTHON_CMD "$test_file"; then
            end_time=$(date +%s)
            elapsed=$((end_time - start_time))
            echo -e "${GREEN}✓ PASSED ($elapsed seconds)${NC}"
        else
            echo -e "${RED}✗ FAILED${NC}"
            return 1
        fi
    else
        echo -e "${RED}✗ File not found: $test_file${NC}"
        return 1
    fi
    
    echo ""
    return 0
}

# Determine which tests to run
TEST_MODE="${1:-full}"

if [ "$TEST_MODE" = "fast" ]; then
    echo "  MODE: FAST (Essential tests only)"
    echo "  Estimated time: ~80 seconds"
    echo ""
    
    run_test 1 "MNIST Real Dataset Validation" \
        "bft_week2_mnist_validation.py" "40s"
    
    run_test 2 "Failure Mode Testing" \
        "bft_week2_failure_modes.py" "35s"
    
    run_test 7 "Production Readiness Report" \
        "bft_week2_production_readiness.py" "5s"

elif [ "$TEST_MODE" = "full" ]; then
    echo "  MODE: FULL (Complete validation)"
    echo "  Estimated time: ~180 seconds"
    echo ""
    
    run_test 1 "MNIST Real Dataset Validation" \
        "bft_week2_mnist_validation.py" "40s"
    
    run_test 2 "Failure Mode Testing" \
        "bft_week2_failure_modes.py" "35s"
    
    run_test 3 "Network Partition Testing" \
        "bft_week2_network_partitions.py" "30s"
    
    run_test 4 "Cascading Failure Analysis" \
        "bft_week2_cascading_failures.py" "25s"
    
    run_test 5 "GPU Profiling" \
        "bft_week2_gpu_profiling.py" "20s"
    
    run_test 6 "Ultra-Scale Testing (5000 Nodes)" \
        "bft_week2_5000_node_scaling.py" "25s"
    
    run_test 7 "Production Readiness Report" \
        "bft_week2_production_readiness.py" "5s"

elif [[ "$TEST_MODE" =~ ^[0-9]$ ]]; then
    echo "  MODE: SINGLE TEST ($TEST_MODE)"
    echo ""
    
    case $TEST_MODE in
        1)
            run_test 1 "MNIST Real Dataset Validation" \
                "bft_week2_mnist_validation.py" "40s"
            ;;
        2)
            run_test 2 "Failure Mode Testing" \
                "bft_week2_failure_modes.py" "35s"
            ;;
        3)
            run_test 3 "Network Partition Testing" \
                "bft_week2_network_partitions.py" "30s"
            ;;
        4)
            run_test 4 "Cascading Failure Analysis" \
                "bft_week2_cascading_failures.py" "25s"
            ;;
        5)
            run_test 5 "GPU Profiling" \
                "bft_week2_gpu_profiling.py" "20s"
            ;;
        6)
            run_test 6 "Ultra-Scale Testing (5000 Nodes)" \
                "bft_week2_5000_node_scaling.py" "25s"
            ;;
        7)
            run_test 7 "Production Readiness Report" \
                "bft_week2_production_readiness.py" "5s"
            ;;
        *)
            echo -e "${RED}Unknown test number: $TEST_MODE${NC}"
            exit 1
            ;;
    esac
else
    echo -e "${RED}Unknown mode: $TEST_MODE${NC}"
    echo ""
    echo "Usage: bash run_week2_tests.sh [fast|full|<test_number>]"
    echo ""
    echo "  fast   - Essential tests only (~80 seconds)"
    echo "  full   - All tests (~180 seconds)"
    echo "  1-7    - Run specific test"
    echo ""
    exit 1
fi

echo ""
echo "╔════════════════════════════════════════════════════════════════════════════════╗"
echo "║                          ALL TESTS COMPLETED                                  ║"
echo "║                                                                                ║"
echo "║  Review results in: PRODUCTION_READINESS_REPORT.md                            ║"
echo "║  Raw test data in:  results/ directory                                        ║"
echo "╚════════════════════════════════════════════════════════════════════════════════╝"
echo ""
