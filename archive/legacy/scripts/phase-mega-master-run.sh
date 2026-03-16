#!/bin/bash
#================================================================================
# SOVEREIGN MAP: MASTER ORCHESTRATOR (200-NODE TEST)
# Purpose: Full automation of Setup, Deploy, Test, Capture, and Cleanup.
#================================================================================
set -e

# Configuration
TEST_NAME="simulation-$(date +%Y%m%d-%H%M%S)"
LOG_FILE="master-run-${TEST_NAME}.log"

exec > >(tee -a "$LOG_FILE") 2>&1

echo "=========================================================="
echo "🚀 STARTING MASTER SIMULATION: $TEST_NAME"
echo "=========================================================="

# Step 1: AWS Setup & Configuration
echo "Step 1/6: Initializing AWS Environment..."
./phase-1-aws-setup.sh

# Step 2: Infrastructure Provisioning
echo "Step 2/6: Provisioning EC2 Cluster via Terraform..."
./phase-2-deploy-infrastructure.sh

# Step 3: High-Density Code Deployment
echo "Step 3/6: Deploying 200 Nodes (25 per host)..."
./phase-3-deploy-code.sh

# Step 4: Verification
echo "Step 4/6: Verifying Mesh Health..."
# Wait 30s for containers to stabilize before checking
sleep 30
python3 verify-mesh.py

# Step 5: Execution
echo "Step 5/6: Starting Federated Learning Rounds..."
./phase-4-execute-test.sh

# Step 6: Capture & Document
echo "Step 6/6: Capturing Results and Committing..."
./phase-5-capture-results.sh "$TEST_NAME"

# Step 7: Automated Cleanup
echo "----------------------------------------------------------"
read -p "🏁 Test Finished. Ready to destroy infrastructure? (y/n): " RUN_CLEANUP
if [[ $RUN_CLEANUP == "y" ]]; then
    ./phase-6-cleanup-aws.sh
else
    echo "⚠️  Resources remain active. Run ./phase-6-cleanup-aws.sh manually to stop billing."
fi

echo "=========================================================="
echo "✅ MASTER RUN COMPLETE: Log saved to $LOG_FILE"
echo "=========================================================="
