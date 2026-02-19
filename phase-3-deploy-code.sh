#!/bin/bash
echo "=========================================="
echo "PHASE 3: Optimized Code Deployment (200-Node Scale)"
echo "=========================================="

# 1. Load Configuration
if [ -f "aws-config.env" ]; then
    source aws-config.env
else
    echo "❌ Error: aws-config.env not found. Run Phase 1 first."
    exit 1
fi

KEY_PATH="./terraform/sovereign-fl-key.pem"

# 2. Get Aggregator IP
AGGREGATOR_IP=$(terraform -chdir=terraform output -raw aggregator_ip)
if [ -z "$AGGREGATOR_IP" ] || [ "$AGGREGATOR_IP" == "null" ]; then
    echo "❌ Error: Aggregator IP not found in Terraform."
    exit 1
fi

# 3. Get Client IPs
CLIENT_IPS=$(aws ec2 describe-instances \
    --filters "Name=tag:aws:autoscaling:groupName,Values=sovereign-fl-clients" "Name=instance-state-name,Values=running" \
    --query "Reservations[*].Instances[*].PrivateIpAddress" \
    --output text)

# 4. Deploy to Aggregator
echo "📦 Setting up Aggregator ($AGGREGATOR_IP)..."
ssh -i "$KEY_PATH" -o StrictHostKeyChecking=no ubuntu@"$AGGREGATOR_IP" << EOF
    sudo apt-get update && sudo apt-get install -y docker.io docker-compose
    if [ ! -d "Sovereign_Map_Federated_Learning" ]; then
        git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning.git
    fi
    cd Sovereign_Map_Federated_Learning && git pull
    # Start the aggregator service
    docker-compose up -d aggregator
EOF

# 5. Deploy to Client Cluster (High Density)
for IP in $CLIENT_IPS; do
    echo "📲 Deploying 25 nodes to Client Host: $IP"
    ssh -i "$KEY_PATH" -o StrictHostKeyChecking=no -J ubuntu@"$AGGREGATOR_IP" ubuntu@"$IP" << EOF
        sudo apt-get update && sudo apt-get install -y docker.io docker-compose
        
        # Clone repo if missing
        [ ! -d "Sovereign_Map_Federated_Learning" ] && git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning.git
        cd Sovereign_Map_Federated_Learning && git pull

        # Clean up old runs
        docker rm -f \$(docker ps -aq --filter "label=project=sovereign-map") 2>/dev/null

        # Launch 25 containers using the optimized config
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
                $DOCKER_IMAGE_NAME
        done
        echo "✓ Client $IP: 25 nodes launched successfully."
EOF
done

echo "=========================================="
echo "PHASE 3 COMPLETE: 200 Nodes Deployed Across 8 Hosts"
echo "=========================================="
