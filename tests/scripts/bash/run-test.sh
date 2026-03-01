#!/bin/bash
#================================================================================
# MASTER ORCHESTRATION SCRIPT
# Sovereign FL 200-Node Test
#================================================================================

set -e

VERSION="1.0.0"
TEST_NAME="Sovereign FL 200-Node Test"

echo "=========================================="
echo "${TEST_NAME}"
echo "Version: ${VERSION}"
echo "=========================================="
echo ""

# Check if running in correct directory
if [[ ! -f "phase-1-aws-setup.sh" ]]; then
    echo "Error: Not in test directory"
    echo "Please run from the directory containing phase-*.sh scripts"
    exit 1
fi

# Parse command line arguments
COMMAND=${1:-"help"}

show_help() {
    cat << EOF
Usage: ./run-test.sh <command>

Commands:
  setup       Run Phase 1: AWS Setup
  deploy      Run Phase 2: Infrastructure Deployment  
  code        Run Phase 3: Code Deployment
  test        Run Phase 4: Test Execution
  results     Run Phase 5: Results Capture
  cleanup     Run Phase 6: Cleanup
  all         Run all phases (setup → test)
  status      Check test status
  help        Show this help

Examples:
  ./run-test.sh setup      # Setup AWS account
  ./run-test.sh deploy     # Deploy infrastructure
  ./run-test.sh all        # Run complete pipeline
  ./run-test.sh cleanup    # Cleanup everything

EOF
}

run_phase_1() {
    echo "Running Phase 1: AWS Setup"
    ./phase-1-aws-setup.sh
}

run_phase_2() {
    echo "Running Phase 2: Infrastructure Deployment"
    ./phase-2-deploy-infrastructure.sh
}

run_phase_3() {
    echo "Running Phase 3: Code Deployment"
    ./phase-3-deploy-code.sh
}

run_phase_4() {
    echo "Running Phase 4: Test Execution"
    ./phase-4-execute-test.sh
}

run_phase_5() {
    echo "Running Phase 5: Results Capture"
    RESULTS_DIR=$(ls -td results-* 2>/dev/null | head -1)
    if [[ -z "$RESULTS_DIR" ]]; then
        echo "Error: No results directory found"
        exit 1
    fi
    ./phase-5-capture-results.sh "$RESULTS_DIR"
}

run_phase_6() {
    echo "Running Phase 6: Cleanup"
    ./phase-6-cleanup.sh
}

run_all() {
    echo "Running complete test pipeline"
    run_phase_1
    run_phase_2
    run_phase_3
    run_phase_4
    run_phase_5
    echo "Pipeline complete. Run cleanup when done."
}

check_status() {
    echo "Checking test status..."
    [[ -f "aws-config.env" ]] && echo "✓ Phase 1 (AWS Setup) complete" || echo "✗ Phase 1 not run"
    [[ -f "deployment-outputs.env" ]] && echo "✓ Phase 2 (Deploy) complete" || echo "✗ Phase 2 not run"
    ls -d results-* 1>/dev/null 2>&1 && echo "✓ Results available" || echo "✗ No results yet"
}

# Main command dispatch
case $COMMAND in
    setup|phase1) run_phase_1 ;;
    deploy|phase2) run_phase_2 ;;
    code|phase3) run_phase_3 ;;
    test|phase4) run_phase_4 ;;
    results|phase5) run_phase_5 ;;
    cleanup|phase6) run_phase_6 ;;
    all|full) run_all ;;
    status) check_status ;;
    help|--help|-h) show_help ;;
    *) echo "Unknown command: $COMMAND"; show_help; exit 1 ;;
esac
