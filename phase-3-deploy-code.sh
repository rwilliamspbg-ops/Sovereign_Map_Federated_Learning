#!/bin/bash
echo "=========================================="
echo "PHASE 3: Code Deployment (Terraform Mode)"
echo "=========================================="

# 1. Configuration - Paths are relative to the root of the repo
KEY_PATH="./terraform/sovereign-fl-key.pem"
REPO_URL="https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning.git"

# 2. Get the Aggregator IP from Terraform
# We use -chdir=terraform so we don't have to move into the folder
AGGREGATOR_IP=$(terraform -chdir=terraform output -raw aggregator_ip)

if [ -z "$AGGREGATOR_IP" ] || [ "$AGGREGATOR_IP" == "null" ]; then
    echo "❌ Error: Aggregator IP not found in Terraform. Run 'terraform apply' first."
    exit 1
fi

echo "🚀 Found Aggregator IP: $AGGREGATOR_IP"

# 3. Get Client Private IPs using AWS CLI
echo "📡 Finding running client nodes..."
CLIENT_IPS=$(aws ec2 describe-instances \
    --filters "Name=tag:aws:autoscaling:groupName,Values=sovereign-fl-clients" "Name=instance-state-name,Values=running" \
    --query "Reservations[*].Instances[*].PrivateIpAddress" \
    --output text)

if [ -z "$CLIENT_IPS" ]; then
    echo "⚠️ No running clients found. Checking for initialization..."
    sleep 10
    exit 1
fi

# 4. Deploy to Aggregator
echo "📦 Configuring Aggregator ($AGGREGATOR_IP)..."
ssh -i "$KEY_PATH" -o StrictHostKeyChecking=no ubuntu@"$AGGREGATOR_IP" << 'EOF'
    sudo apt-get update && sudo apt-get install -y docker.io git python3-pip
    if [ ! -d "Sovereign_Map_Federated_Learning" ]; then
        git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning.git
    else
        cd Sovereign_Map_Federated_Learning && git pull
    fi
    echo "✓ Aggregator setup complete"
EOF

# 5. Deploy to Clients (using Aggregator as Jump Host)
for IP in $CLIENT_IPS; do
    echo "📲 Deploying to Client: $IP"
    ssh -i "$KEY_PATH" -o StrictHostKeyChecking=no -J ubuntu@"$AGGREGATOR_IP" ubuntu@"$IP" << EOF
        sudo apt-get update && sudo apt-get install -y python3-pip git
        if [ ! -d "Sovereign_Map_Federated_Learning" ]; then
            git clone $REPO_URL
        else
            cd Sovereign_Map_Federated_Learning && git pull
        fi
        pip3 install -r Sovereign_Map_Federated_Learning/requirements.txt
        echo "✓ Client $IP configured"
EOF
done

echo ""
echo "=========================================="
echo "PHASE 3 COMPLETE: 1 Aggregator & $(echo $CLIENT_IPS | wc -w) Clients Ready"
echo "=========================================="
