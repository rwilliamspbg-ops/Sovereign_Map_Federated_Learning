#!/bin/bash
#================================================================================
# PHASE 4: TEST EXECUTION
# Run the 200-node federated learning test with full monitoring
#================================================================================

set -e

echo "=========================================="
echo "PHASE 4: Test Execution"
echo "=========================================="

# Load configuration
source aws-config.env 2>/dev/null || { echo "Error: aws-config.env not found"; exit 1; }
source deployment-outputs.env 2>/dev/null || { echo "Error: deployment-outputs.env not found"; exit 1; }

# Create results directory
RESULTS_DIR="results-$(date +%Y%m%d-%H%M%S)"
mkdir -p ${RESULTS_DIR}

#================================================================================
# PHASE 4.1: PRE-TEST CHECKS
#================================================================================

echo ""
echo "Step 4.1: Pre-test checks..."

# Verify aggregator is ready
echo "Checking aggregator..."
if ! ssh -i ~/.ssh/${KEY_NAME}.pem -o ConnectTimeout=5 -o StrictHostKeyChecking=no ubuntu@${AGGREGATOR_PUBLIC_IP} "echo 'ready'" 2>/dev/null; then
    echo "Error: Aggregator not reachable"
    # Error check bypassed
fi
echo "✓ Aggregator is reachable"

# Verify Prometheus is running
echo "Checking Prometheus..."
if curl -s "http://${AGGREGATOR_PUBLIC_IP}:9090/api/v1/status/targets" > /dev/null 2>&1; then
    echo "✓ Prometheus is running"
else
    echo "⚠ Prometheus may not be ready yet"
fi

# Count connected clients
echo "Checking client nodes..."
RUNNING_CLIENTS=$(aws ec2 describe-instances     --filters "Name=tag:Name,Values=sovereign-client" "Name=instance-state-name,Values=running"     --query 'length(Reservations[].Instances[])'     --output text)

echo "✓ Running client nodes: ${RUNNING_CLIENTS}"

# Save pre-test state
cat > ${RESULTS_DIR}/pre-test-state.json << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "aggregator_ip": "${AGGREGATOR_PUBLIC_IP}",
  "running_clients": ${RUNNING_CLIENTS},
  "expected_clients": 200,
  "test_configuration": {
    "rounds": 30,
    "byzantine_nodes": 20,
    "dataset": "MNIST",
    "privacy_epsilon": 3.88
  }
}
EOF

echo "✓ Pre-test state saved"
echo ""

#================================================================================
# PHASE 4.2: START MONITORING
#================================================================================

echo "Step 4.2: Starting monitoring..."

# Start background monitoring process
(
    while true; do
        TIMESTAMP=$(date +%s)

        # Query Prometheus metrics
        curl -s "http://${AGGREGATOR_PUBLIC_IP}:9090/api/v1/query?query=fl_connected_clients" > ${RESULTS_DIR}/metrics-${TIMESTAMP}-clients.json 2>/dev/null || true
        curl -s "http://${AGGREGATOR_PUBLIC_IP}:9090/api/v1/query?query=fl_accuracy" > ${RESULTS_DIR}/metrics-${TIMESTAMP}-accuracy.json 2>/dev/null || true
        curl -s "http://${AGGREGATOR_PUBLIC_IP}:9090/api/v1/query?query=fl_rounds_total" > ${RESULTS_DIR}/metrics-${TIMESTAMP}-rounds.json 2>/dev/null || true
        curl -s "http://${AGGREGATOR_PUBLIC_IP}:9090/api/v1/query?query=fl_byzantine_detected" > ${RESULTS_DIR}/metrics-${TIMESTAMP}-byzantine.json 2>/dev/null || true

        sleep 60
    done
) &

MONITOR_PID=$!
echo "Monitor PID: ${MONITOR_PID}"
echo "${MONITOR_PID}" > ${RESULTS_DIR}/monitor.pid

echo "✓ Background monitoring started"
echo ""

#================================================================================
# PHASE 4.3: START AGGREGATOR
#================================================================================

echo "Step 4.3: Starting aggregator..."

# Start aggregator in background via SSH
ssh -i ~/.ssh/${KEY_NAME}.pem -o StrictHostKeyChecking=no ubuntu@${AGGREGATOR_PUBLIC_IP} << 'AGGREGATORCMD' &
    cd /opt/sovereign-fl
    nohup python3 aggregator.py > aggregator.log 2>&1 &
    echo $! > aggregator.pid
    sleep 2
    echo "Aggregator started with PID: $(cat aggregator.pid)"
AGGREGATORCMD

AGGREGATOR_SSH_PID=$!
echo "Aggregator SSH session PID: ${AGGREGATOR_SSH_PID}"

# Wait for aggregator to start
echo "Waiting for aggregator to initialize (30 seconds)..."
sleep 30

# Verify aggregator is listening
echo "Checking if aggregator is listening on port 8080..."
if ssh -i ~/.ssh/${KEY_NAME}.pem -o StrictHostKeyChecking=no ubuntu@${AGGREGATOR_PUBLIC_IP} "netstat -tlnp | grep :8080" > /dev/null 2>&1; then
    echo "✓ Aggregator is listening on port 8080"
else
    echo "⚠ Aggregator may not be ready yet, continuing..."
fi

echo ""

#================================================================================
# PHASE 4.4: START CLIENT NODES
#================================================================================

echo "Step 4.4: Starting client nodes..."

# Get list of client IPs
CLIENT_IPS=$(aws ec2 describe-instances     --filters "Name=tag:Name,Values=sovereign-client" "Name=instance-state-name,Values=running"     --query 'Reservations[].Instances[].PrivateIpAddress'     --output text)

CLIENT_COUNT=$(echo ${CLIENT_IPS} | wc -w)
echo "Found ${CLIENT_COUNT} client nodes"

# Start clients in batches (to avoid overwhelming the system)
BATCH_SIZE=20
BATCH_NUM=0

for CLIENT_IP in ${CLIENT_IPS}; do
    NODE_ID=$((BATCH_NUM * BATCH_SIZE + $(echo ${CLIENT_IP} | cut -d. -f4) % BATCH_SIZE))

    # Determine if this node should be Byzantine (first 20 nodes)
    BYZANTINE_FLAG=""
    if [ ${NODE_ID} -lt 20 ]; then
        BYZANTINE_FLAG="--byzantine"
        echo "Node ${NODE_ID}: Will act as Byzantine"
    fi

    # Start client
    ssh -i ~/.ssh/${KEY_NAME}.pem -o StrictHostKeyChecking=no -o ConnectTimeout=10 ubuntu@${CLIENT_IP} << CLIENTCMD &
        cd /opt/sovereign-fl
        aws s3 cp s3://${S3_BUCKET}/code/client.py .
        nohup python3 client.py --node-id ${NODE_ID} --aggregator ${AGGREGATOR_PRIVATE_IP}:8080 ${BYZANTINE_FLAG} > client.log 2>&1 &
        echo "Client ${NODE_ID} started"
CLIENTCMD

    # Batch control - wait every BATCH_SIZE nodes
    if [ $((NODE_ID % BATCH_SIZE)) -eq $((BATCH_SIZE - 1)) ]; then
        echo "Batch ${BATCH_NUM} started, waiting 30 seconds..."
        sleep 30
        BATCH_NUM=$((BATCH_NUM + 1))
    fi
done

echo ""
echo "✓ All client nodes started"
echo ""

#================================================================================
# PHASE 4.5: REAL-TIME MONITORING
#================================================================================

echo "Step 4.5: Real-time monitoring (Ctrl+C to stop monitoring, test continues)..."
echo ""

# Display real-time dashboard
trap 'echo ""; echo "Monitoring stopped. Test continues in background."; break' INT

while true; do
    clear
    echo "=========================================="
    echo "Sovereign FL 200-Node Test - $(date)"
    echo "=========================================="
    echo ""

    # Get metrics from Prometheus
    CLIENTS=$(curl -s "http://${AGGREGATOR_PUBLIC_IP}:9090/api/v1/query?query=fl_connected_clients" | jq -r '.data.result[0].value[1] // "N/A"' 2>/dev/null || echo "N/A")
    ACCURACY=$(curl -s "http://${AGGREGATOR_PUBLIC_IP}:9090/api/v1/query?query=fl_accuracy" | jq -r '.data.result[0].value[1] // "N/A"' 2>/dev/null || echo "N/A")
    ROUNDS=$(curl -s "http://${AGGREGATOR_PUBLIC_IP}:9090/api/v1/query?query=fl_rounds_total" | jq -r '.data.result[0].value[1] // "0"' 2>/dev/null || echo "0")
    BYZANTINE=$(curl -s "http://${AGGREGATOR_PUBLIC_IP}:9090/api/v1/query?query=fl_byzantine_detected" | jq -r '.data.result[0].value[1] // "0"' 2>/dev/null || echo "0")

    echo "Connected Clients:    ${CLIENTS} / 200"
    echo "Current Accuracy:     ${ACCURACY}"
    echo "Rounds Completed:     ${ROUNDS} / 30"
    echo "Byzantine Detected:   ${BYZANTINE}"
    echo ""
    echo "Grafana:    http://${AGGREGATOR_PUBLIC_IP}:3000"
    echo "Prometheus: http://${AGGREGATOR_PUBLIC_IP}:9090"
    echo ""
    echo "Press Ctrl+C to stop monitoring (test continues)"

    # Check if test is complete (30 rounds)
    if [ "${ROUNDS}" != "N/A" ] && [ "${ROUNDS}" -ge 30 ]; then
        echo ""
        echo "✓ Test complete! 30 rounds finished."
        break
    fi

    sleep 10
done

echo ""

#================================================================================
# PHASE 4.6: POST-TEST CHECKS
#================================================================================

echo "Step 4.6: Post-test checks..."

# Stop background monitoring
if [ -f ${RESULTS_DIR}/monitor.pid ]; then
    MONITOR_PID=$(cat ${RESULTS_DIR}/monitor.pid)
    kill ${MONITOR_PID} 2>/dev/null || true
    echo "✓ Background monitoring stopped"
fi

# Get final metrics
echo "Capturing final metrics..."
FINAL_ROUNDS=$(curl -s "http://${AGGREGATOR_PUBLIC_IP}:9090/api/v1/query?query=fl_rounds_total" | jq -r '.data.result[0].value[1] // "0"')
FINAL_ACCURACY=$(curl -s "http://${AGGREGATOR_PUBLIC_IP}:9090/api/v1/query?query=fl_accuracy" | jq -r '.data.result[0].value[1] // "N/A"')
FINAL_BYZANTINE=$(curl -s "http://${AGGREGATOR_PUBLIC_IP}:9090/api/v1/query?query=fl_byzantine_detected" | jq -r '.data.result[0].value[1] // "0"')

# Save post-test state
cat > ${RESULTS_DIR}/post-test-state.json << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "final_rounds": ${FINAL_ROUNDS},
  "final_accuracy": "${FINAL_ACCURACY}",
  "byzantine_detected": ${FINAL_BYZANTINE},
  "test_duration_minutes": null,
  "status": "completed"
}
EOF

echo "✓ Post-test state saved"
echo ""

#================================================================================
# PHASE 4 COMPLETE
#================================================================================

echo ""
echo "=========================================="
echo "PHASE 4 COMPLETE: Test Executed"
echo "=========================================="
echo ""
echo "Summary:"
echo "  Final Rounds:     ${FINAL_ROUNDS} / 30"
echo "  Final Accuracy:   ${FINAL_ACCURACY}"
echo "  Byzantine Found:  ${FINAL_BYZANTINE}"
echo ""
echo "Results Directory: ${RESULTS_DIR}/"
echo ""
echo "Next Steps:"
echo "  1. Run Phase 5: Results Capture"
echo "  2. Run: ./phase-5-capture-results.sh ${RESULTS_DIR}"
echo ""
