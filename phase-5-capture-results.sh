#!/bin/bash
#================================================================================
# PHASE 5: RESULTS CAPTURE AND DOCUMENTATION (200-NODE DENSITY OPTIMIZED)
# Capture results from Docker containers across 8 high-density instances.
#================================================================================
set -e

RESULTS_DIR=$1
if [[ -z "$RESULTS_DIR" ]]; then
    echo "Usage: ./phase-5-capture-results.sh <RESULTS_DIRECTORY>"
    exit 1
fi

mkdir -p "$RESULTS_DIR"
echo "=========================================="
echo "PHASE 5: Results Capture and Documentation"
echo "=========================================="

# 1. Load Configuration
source aws-config.env 2>/dev/null || true
source deployment-outputs.env 2>/dev/null || true
KEY_PATH="./terraform/sovereign-fl-key.pem"

# 2. Get Infrastructure IPs
AGGREGATOR_IP=$(terraform -chdir=terraform output -raw aggregator_ip)
WORKER_HOSTS=$(aws ec2 describe-instances \
    --filters "Name=tag:aws:autoscaling:groupName,Values=sovereign-fl-clients" "Name=instance-state-name,Values=running" \
    --query 'Reservations[].Instances[].PrivateIpAddress' \
    --output text)

#================================================================================
# PHASE 5.2: DOWNLOAD LOGS (DOCKER & DENSITY OPTIMIZED)
#================================================================================
echo "Step 5.2: Downloading Docker logs from cluster..."
mkdir -p ${RESULTS_DIR}/logs

for HOST_IP in $WORKER_HOSTS; do
    echo "  → Collecting logs from Host: $HOST_IP"
    
    # Extract logs from ALL 25 containers on the host
    ssh -i "$KEY_PATH" -o StrictHostKeyChecking=no -J ubuntu@"$AGGREGATOR_IP" ubuntu@"$HOST_IP" << 'EOF'
        mkdir -p ~/node_logs
        # Target containers by project label
        CONTAINER_IDS=$(docker ps -q --filter "label=project=sovereign-map")
        for ID in $CONTAINER_IDS; do
            NAME=$(docker inspect --format '{{.Name}}' $ID | sed 's/\///')
            docker logs --tail 1000 $ID > ~/node_logs/${NAME}.log 2>/dev/null
        done
        tar -czf node_logs_$(hostname).tar.gz -C ~/node_logs .
        rm -rf ~/node_logs
EOF

    # Pull the tarball through the JumpHost
    scp -i "$KEY_PATH" -o StrictHostKeyChecking=no -o ProxyCommand="ssh -i $KEY_PATH -W %h:%p ubuntu@$AGGREGATOR_IP" \
        ubuntu@$HOST_IP:~/node_logs_$(hostname).tar.gz ${RESULTS_DIR}/logs/
done

#================================================================================
# PHASE 5.7: GIT COMMIT RESULTS
#================================================================================
echo "Step 5.7: Committing results to repository..."
TEST_ID=$(basename ${RESULTS_DIR})

if git rev-parse --git-dir > /dev/null 2>&1; then
    git checkout -b "test-results-${TEST_ID}"
    git add ${RESULTS_DIR}/REPORT.md
    git add ${RESULTS_DIR}/logs/*.tar.gz
    git commit -m "docs: capture 200-node test results ${TEST_ID}"
    echo "✓ Results committed to branch test-results-${TEST_ID}"
else
    echo "⚠ Not a git repo, skipping commit."
fi

echo "=========================================="
echo "PHASE 5 COMPLETE: Data Secured"
echo "=========================================="
