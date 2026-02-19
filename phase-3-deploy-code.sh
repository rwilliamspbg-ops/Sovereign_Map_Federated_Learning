#!/bin/bash
echo "=========================================================="
echo "PHASE 3: High-Density Parallel Deployment (200-Node Scale)"
echo "=========================================================="

# 1. Load Configuration
if [ -f "aws-config.env" ]; then
    source aws-config.env
else
    echo "❌ Error: aws-config.env not found. Run Phase 1 first."
    exit 1
fi

KEY_PATH="./terraform/sovereign-fl-key.pem"
chmod 400 "$KEY_PATH"

# 2. Get Aggregator IP
AGGREGATOR_IP=$(terraform -chdir=terraform output -raw aggregator_ip)
if [ -z "$AGGREGATOR_IP" ] || [ "$AGGREGATOR_IP" == "null" ]; then
    echo "❌ Error: Aggregator IP not found in Terraform output."
    exit 1
fi

# 3. Get Client IPs
CLIENT_IPS=$(aws ec2 describe-instances \
    --filters "Name=tag:aws:autoscaling:groupName,Values=sovereign-fl-clients" "Name=instance-state-name,Values=running" \
    --query "Reservations[*].Instances[*].PrivateIpAddress" \
    --output text)

# 4. Step 1: Initialize Aggregator (Must be ready first)
echo "📦 Setting up Genesis Aggregator ($AGGREGATOR_IP)..."
ssh -i "$KEY_PATH" -o StrictHostKeyChecking=no ubuntu@"$AGGREGATOR_IP" << EOF
    sudo apt-get update && sudo apt-get install -y docker.io docker-compose
    [ ! -d "Sovereign_Map_Federated_Learning" ] && git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning.git
    cd Sovereign_Map_Federated_Learning && git pull
    docker-compose up -d aggregator
EOF

echo "✓ Aggregator is online. Starting parallel node launch..."

# 5. Step 2: Parallel Deployment to Clients
for IP in $CLIENT_IPS; do
    (
        echo "📲 [Host $IP] Launching 25 nodes..."
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
            echo "✓ Host $IP: All 25 nodes are UP."
EOF
    ) & 
done

# Wait for all background SSH processes to finish
wait

echo ""
echo "=========================================================="
echo "SUCCESS: 200 Nodes deployed across 8 instances."
echo "Aggregator: http://$AGGREGATOR_IP:8081"
echo "Run 'phase-4-execute-test.sh' to begin FL training."
echo "=========================================================="
