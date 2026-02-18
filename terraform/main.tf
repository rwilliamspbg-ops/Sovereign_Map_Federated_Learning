#!/bin/bash

# Configuration
KEY_PATH="$HOME/.ssh/sovereign-fl-key.pem"
REPO_URL="https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning.git"

# 1. Get the Aggregator IP from Terraform
AGGREGATOR_IP=$(terraform output -raw aggregator_ip)
echo "🚀 Target Aggregator IP: $AGGREGATOR_IP"

# 2. Setup the Aggregator
echo "Setting up Aggregator..."
ssh -i $KEY_PATH -o StrictHostKeyChecking=no ubuntu@$AGGREGATOR_IP << 'EOF'
    sudo apt-get update
    sudo apt-get install -y docker.io git
    git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning.git
    cd Sovereign_Map_Federated_Learning
    # Start the aggregator (Assuming a start script exists)
    # ./start-aggregator.sh & 
EOF

# 3. Get all Client Private IPs
CLIENT_IPS=$(aws ec2 describe-instances \
    --filters "Name=tag:aws:autoscaling:groupName,Values=sovereign-fl-clients" "Name=instance-state-name,Values=running" \
    --query "Reservations[*].Instances[*].PrivateIpAddress" \
    --output text)

echo "📡 Found $(echo $CLIENT_IPS | wc -w) clients. Starting deployment..."

# 4. Deploy to Clients (Parallel)
for IP in $CLIENT_IPS; do
    echo "Configuring Client: $IP"
    # We use the Aggregator as a jump host to reach private clients
    ssh -i $KEY_PATH -o StrictHostKeyChecking=no -J ubuntu@$AGGREGATOR_IP ubuntu@$IP << EOF
        sudo apt-get update
        sudo apt-get install -y python3-pip git
        git clone $REPO_URL
        cd Sovereign_Map_Federated_Learning
        pip3 install -r requirements.txt
        # Start the learning process pointing to the Aggregator
        # python3 client.py --aggregator $AGGREGATOR_IP &
EOF
done

echo "✅ Deployment complete across all active nodes."
