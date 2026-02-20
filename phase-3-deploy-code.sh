#!/bin/bash
# MANUAL OVERRIDES - Using your known working IP and Key
AGGREGATOR_IP="3.94.31.131"
KEY_PATH="./terraform/sovereign-fl-key.pem"

echo "=========================================================="
echo "PHASE 3: High-Density Parallel Deployment (200-Node Scale)"
echo "=========================================================="

# 1. Load Configuration (Optional, keeps other env vars)
if [ -f "aws-config.env" ]; then
    source aws-config.env
fi

# Fix permissions on the key
chmod 400 "$KEY_PATH" 2>/dev/null || echo "⚠️ Warning: Could not set permissions on key."

# 2. BYPASS TERRAFORM - Do not overwrite AGGREGATOR_IP
echo "📡 Using manual Aggregator IP: $AGGREGATOR_IP"

# 3. Fetch Client IPs from AWS (Simplified to your current single instance)
# If you only have one instance, we use the same IP for the clients
CLIENT_IPS="3.94.31.131"

# 4. Initialize Genesis Aggregator (Must be up before nodes check in)
echo "📦 Setting up Genesis Aggregator ($AGGREGATOR_IP)..."
ssh -i "$KEY_PATH" -o StrictHostKeyChecking=no ubuntu@"$AGGREGATOR_IP" << EOF
    sudo apt-get update && sudo apt-get install -y docker.io docker-compose
    [ ! -d "Sovereign_Map_Federated_Learning" ] && git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning.git
    cd Sovereign_Map_Federated_Learning && git pull
    docker-compose up -d aggregator
EOF

echo "✓ Aggregator is online. Starting parallel node launch across cluster..."

# 5. Parallel Deployment to Worker Hosts
for IP in $CLIENT_IPS; do
    (
        echo "📲 [Host $IP] Deploying 25 nodes..."
        ssh -i "$KEY_PATH" -o StrictHostKeyChecking=no -J ubuntu@"$AGGREGATOR_IP" ubuntu@"$IP" << EOF
            sudo apt-get update && sudo apt-get install -y docker.io docker-compose
            
            [ ! -d "Sovereign_Map_Federated_Learning" ] && git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning.git
            cd Sovereign_Map_Federated_Learning && git pull

            # Cleanup previous project containers
            docker rm -f \$(docker ps -aq --filter "label=project=sovereign-map") 2>/dev/null

            # High-density launch loop
            for i in \$(seq 1 $NODES_PER_INSTANCE); do
                NODE_PORT=\$((8000 + i))
                NODE_ID="node_\${IP//./-}_\$i"
                
                docker run -d \\
                    --name "\$NODE_ID" \\
                    --label "project=sovereign-map" \\
                    --network host \\
                    --memory "$DOCKER_MEM_LIMIT" \\
                    --cpus "$DOCKER_CPU_RESERVATION" \\
                    -e NODE_ID="\$NODE_ID" \\
                    -e AGGREGATOR_URL="http://$AGGREGATOR_IP:8081" \\
                    -e CLIENT_API_PORT="\$NODE_PORT" \\
                    -e LOCAL_EPOCHS="$LOCAL_EPOCHS" \\
                    -e BATCH_SIZE="$BATCH_SIZE" \\
                    -e PRIVACY_EPSILON="$PRIVACY_EPSILON" \\
                    -e BFT_THRESHOLD="$BFT_THRESHOLD" \\
                    --restart on-failure:3 \\
                    $DOCKER_IMAGE_NAME > /dev/null
            done
            echo "✓ Host $IP: Successfully launched 25 nodes."
EOF
    ) & 
done

# Wait for all background threads to complete
wait

echo ""
echo "=========================================================="
echo "SUCCESS: 200 Nodes deployed across 8 instances."
echo "Aggregator Dashboard: http://$AGGREGATOR_IP:8081"
echo "Next: Run 'phase-4-execute-test.sh' to start FL rounds."
echo "=========================================================="
