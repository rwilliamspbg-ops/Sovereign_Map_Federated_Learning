#!/bin/bash
#================================================================================
# PHASE 5: RESULTS CAPTURE AND DOCUMENTATION (200-NODE DENSITY OPTIMIZED & FIXED)
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
if [[ -f "aws-config.env" ]]; then
    source aws-config.env
else
    echo "⚠ Warning: aws-config.env not found, using defaults"
fi

if [[ -f "deployment-outputs.env" ]]; then
    source deployment-outputs.env
else
    echo "⚠ Warning: deployment-outputs.env not found"
fi

KEY_PATH="./terraform/sovereign-fl-key.pem"
AWS_REGION=${AWS_REGION:-"us-east-1"}
ASG_NAME=${ASG_NAME:-"sovereign-fl-clients"}
CONTAINER_LABEL=${CONTAINER_LABEL:-"project=sovereign-map"}

# 2. Verify Prerequisites
if [[ ! -f "$KEY_PATH" ]]; then
    echo "❌ Error: SSH key not found at $KEY_PATH"
    echo "   Phase 2 should have created this file. Please verify Phase 2 completed successfully."
    exit 1
fi

if [[ ! -d "terraform" ]]; then
    echo "❌ Error: terraform/ directory not found"
    echo "   Phase 2 must be run before Phase 5."
    exit 1
fi

# 3. Get Infrastructure IPs
echo "Step 5.1: Discovering infrastructure..."

# Get aggregator IP from terraform
if ! AGGREGATOR_IP=$(terraform -chdir=terraform output -raw aggregator_ip 2>/dev/null); then
    echo "❌ Error: Failed to get aggregator IP from Terraform state"
    echo "   This usually means Phase 2 did not complete successfully."
    echo "   Please run Phase 2 first to deploy infrastructure."
    exit 1
fi

if [[ -z "$AGGREGATOR_IP" ]]; then
    echo "❌ Error: Aggregator IP is empty"
    exit 1
fi

echo "✓ Aggregator IP: $AGGREGATOR_IP"

# Get worker private IPs from AWS or Terraform
echo "Discovering worker nodes..."
WORKER_HOSTS=$(aws ec2 describe-instances \
    --region "$AWS_REGION" \
    --filters "Name=tag:aws:autoscaling:groupName,Values=$ASG_NAME" "Name=instance-state-name,Values=running" \
    --query 'Reservations[].Instances[].PrivateIpAddress' \
    --output text 2>/dev/null || true)

# Fallback: try getting worker IPs from terraform output
if [[ -z "$WORKER_HOSTS" ]]; then
    echo "⚠ ASG discovery failed, trying Terraform output..."
    WORKER_HOSTS=$(terraform -chdir=terraform output -json worker_ips 2>/dev/null | jq -r '.[]' | tr '\n' ' ' || true)
fi

if [[ -z "$WORKER_HOSTS" ]]; then
    echo "❌ Error: Could not discover any worker nodes"
    echo "   Checked:"
    echo "   - AWS EC2 instances with ASG tag: $ASG_NAME"
    echo "   - Terraform output: worker_ips"
    echo "   Please verify Phase 2 deployment succeeded and instances are running."
    exit 1
fi

WORKER_COUNT=$(echo $WORKER_HOSTS | wc -w)
echo "✓ Found $WORKER_COUNT worker nodes"

#================================================================================
# PHASE 5.2: DOWNLOAD LOGS (DOCKER & DENSITY OPTIMIZED) - FIXED SCP BUG
#================================================================================
echo "Step 5.2: Downloading Docker logs from cluster..."
mkdir -p ${RESULTS_DIR}/logs

LOG_COUNT=0
for HOST_IP in $WORKER_HOSTS; do
    echo "  → Collecting logs from Host: $HOST_IP"
    
    # Extract logs from ALL containers on the host
    # FIXED: Use a static tarball name to avoid hostname evaluation issues
    if ssh -i "$KEY_PATH" -o StrictHostKeyChecking=no -o ConnectTimeout=10 \
        -J ubuntu@"$AGGREGATOR_IP" ubuntu@"$HOST_IP" << EOF
        mkdir -p ~/node_logs
        # Target containers by project label
        CONTAINER_IDS=\$(docker ps -q --filter "label=$CONTAINER_LABEL")
        
        if [[ -z "\$CONTAINER_IDS" ]]; then
            echo "  ⚠ Warning: No containers found with label=$CONTAINER_LABEL on $HOST_IP"
            exit 0
        fi
        
        for ID in \$CONTAINER_IDS; do
            NAME=\$(docker inspect --format '{{.Name}}' \$ID | sed 's/\\///')
            docker logs --tail 1000 \$ID > ~/node_logs/\${NAME}.log 2>/dev/null
        done
        
        # Use a static tarball name to avoid hostname mismatch
        tar -czf node_logs.tar.gz -C ~/node_logs . 2>/dev/null
        rm -rf ~/node_logs
        echo "  ✓ Packaged logs from \$(echo \$CONTAINER_IDS | wc -w) containers"
EOF
    then
        # Pull the tarball through the JumpHost
        # FIXED: Use the same static name that was created on the remote host
        if scp -i "$KEY_PATH" -o StrictHostKeyChecking=no -o ConnectTimeout=10 \
            -o ProxyCommand="ssh -i $KEY_PATH -W %h:%p ubuntu@$AGGREGATOR_IP" \
            ubuntu@$HOST_IP:~/node_logs.tar.gz ${RESULTS_DIR}/logs/node_logs_${HOST_IP}.tar.gz 2>/dev/null; then
            echo "  ✓ Downloaded logs from $HOST_IP"
            ((LOG_COUNT++))
        else
            echo "  ⚠ Warning: Failed to download logs from $HOST_IP"
        fi
    else
        echo "  ⚠ Warning: Failed to connect to $HOST_IP or extract logs"
    fi
done

echo "✓ Successfully collected logs from $LOG_COUNT/$WORKER_COUNT hosts"

if [[ $LOG_COUNT -eq 0 ]]; then
    echo "⚠ Warning: No logs were collected from any worker nodes"
    echo "   This may indicate:"
    echo "   - SSH connectivity issues"
    echo "   - No containers running with label=$CONTAINER_LABEL"
    echo "   - Docker not installed/running on worker nodes"
fi

#================================================================================
# PHASE 5.7: GIT COMMIT RESULTS
#================================================================================
echo "Step 5.7: Committing results to repository..."
TEST_ID=$(basename ${RESULTS_DIR})

if git rev-parse --git-dir > /dev/null 2>&1; then
    # Check if there are any files to commit
    if [[ -f "${RESULTS_DIR}/REPORT.md" ]] || [[ $(ls ${RESULTS_DIR}/logs/*.tar.gz 2>/dev/null | wc -l) -gt 0 ]]; then
        git checkout -b "test-results-${TEST_ID}" 2>/dev/null || git checkout "test-results-${TEST_ID}"
        
        [[ -f "${RESULTS_DIR}/REPORT.md" ]] && git add ${RESULTS_DIR}/REPORT.md
        
        if [[ $(ls ${RESULTS_DIR}/logs/*.tar.gz 2>/dev/null | wc -l) -gt 0 ]]; then
            git add ${RESULTS_DIR}/logs/*.tar.gz
        fi
        
        git commit -m "docs: capture 200-node test results ${TEST_ID}" || echo "⚠ Nothing new to commit"
        echo "✓ Results committed to branch test-results-${TEST_ID}"
    else
        echo "⚠ No results files found to commit (no REPORT.md or log tarballs)"
    fi
else
    echo "⚠ Not a git repo, skipping commit."
fi

echo "=========================================="
echo "PHASE 5 COMPLETE: Data Secured"
echo "  Logs collected: $LOG_COUNT/$WORKER_COUNT hosts"
echo "  Output directory: $RESULTS_DIR"
echo "=========================================="
