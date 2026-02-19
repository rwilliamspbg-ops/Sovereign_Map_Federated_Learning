#!/bin/bash
#================================================================================
# SOVEREIGN MAP: AWS CLEANUP & TERMINATION SCRIPT
# Purpose: Full teardown of the 200-node simulation environment.
#================================================================================
set -e

echo "=========================================="
echo "🧹 STARTING AWS CLEANUP"
echo "=========================================="

# 1. Load Configuration
if [ -f "aws-config.env" ]; then
    source aws-config.env
else
    echo "❌ Error: aws-config.env not found. Manual cleanup may be required."
    exit 1
fi

KEY_PATH="./terraform/sovereign-fl-key.pem"

# 2. Get Aggregator IP
AGGREGATOR_IP=$(terraform -chdir=terraform output -raw aggregator_ip)

# 3. Stop Docker Containers (Optional but recommended for clean exit)
echo "Step 1: Stopping all 200 Docker containers..."
CLIENT_IPS=$(aws ec2 describe-instances \
    --filters "Name=tag:aws:autoscaling:groupName,Values=sovereign-fl-clients" "Name=instance-state-name,Values=running" \
    --query "Reservations[*].Instances[*].PrivateIpAddress" \
    --output text)

for IP in $CLIENT_IPS; do
    (
        ssh -i "$KEY_PATH" -o StrictHostKeyChecking=no -J ubuntu@"$AGGREGATOR_IP" ubuntu@"$IP" \
        "docker rm -f \$(docker ps -aq --filter 'label=project=sovereign-map')" > /dev/null 2>&1 || true
    ) &
done
wait
echo "✓ All containers stopped."

# 4. Terraform Destroy (The Core Teardown)
echo "Step 2: Running Terraform Destroy..."
# This removes EC2 instances, Security Groups, and VPC networking
terraform -chdir=terraform destroy -auto-approve

# 5. Clean up S3 (Optional - Keeps the results bucket unless specified)
read -p "❓ Do you want to delete the S3 bucket ($S3_BUCKET) and all test results? (y/n): " DELETE_S3
if [[ $DELETE_S3 == "y" ]]; then
    echo "🗑️  Deleting S3 bucket contents..."
    aws s3 rb s3://$S3_BUCKET --force
    echo "✓ S3 bucket removed."
fi

# 6. Final Housekeeping
echo "Step 3: Cleaning up local environment files..."
rm -f aws-config.env deployment-outputs.env
echo "✓ Local config files removed."

echo "=========================================="
echo "✅ CLEANUP COMPLETE: All AWS resources terminated."
echo "=========================================="
