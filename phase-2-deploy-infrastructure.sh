#!/bin/bash
#================================================================================
# PHASE 2: INFRASTRUCTURE DEPLOYMENT (OPTIMIZED & FIXED)
# Deploy 8 hosts for the 200-node high-density simulation
#================================================================================
set -e

echo "=========================================="
echo "PHASE 2: Infrastructure Deployment"
echo "=========================================="

# 1. Load configuration from Phase 1
if [[ -f "aws-config.env" ]]; then
    source aws-config.env
else
    echo "❌ Error: aws-config.env not found. Run Phase 1 first."
    exit 1
fi

# Ensure variables have valid defaults
AWS_REGION=${AWS_REGION:-"us-east-1"}
NODE_COUNT=${NODE_COUNT:-8}
KEY_NAME=${KEY_NAME:-"sovereign-fl-key"}
IAM_ROLE=${IAM_ROLE:-"SovereignFL-EC2-Role"}
S3_BACKEND_BUCKET=${S3_BACKEND_BUCKET:-"sovereign-map-terraform-state"}

echo "Configuration loaded:"
echo "  AWS_REGION: $AWS_REGION"
echo "  NODE_COUNT: $NODE_COUNT"
echo "  KEY_NAME: $KEY_NAME"
echo "  S3_BACKEND_BUCKET: $S3_BACKEND_BUCKET"

# 2. Setup Terraform Workspace
mkdir -p terraform
cd terraform

# 3. Verify main.tf exists
if [[ ! -f "main.tf" ]]; then
    echo "❌ Error: terraform/main.tf not found!"
    echo "   The Terraform configuration must be present before running Phase 2."
    echo "   Please ensure main.tf exists in the terraform/ directory."
    exit 1
fi

# 4. Create backend configuration with dynamic region
cat > backend.tf << EOF
terraform {
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 5.0" }
  }
  backend "s3" {
    bucket = "$S3_BACKEND_BUCKET"
    key    = "200-node-test/terraform.tfstate"
    region = "$AWS_REGION"
    encrypt = true
  }
}
EOF

echo "✓ Created backend.tf with bucket=$S3_BACKEND_BUCKET, region=$AWS_REGION"

# 5. Create variables.tf with environment values
cat > variables.tf << EOF
variable "aws_region" { default = "$AWS_REGION" }
variable "node_count" { default = $NODE_COUNT }
variable "key_pair_name" { default = "$KEY_NAME" }
variable "iam_role" { default = "$IAM_ROLE" }
EOF

echo "✓ Created variables.tf with NODE_COUNT=$NODE_COUNT"

# 6. Verify S3 backend bucket exists
echo "Verifying S3 backend bucket..."
if ! aws s3 ls "s3://$S3_BACKEND_BUCKET" --region "$AWS_REGION" >/dev/null 2>&1; then
    echo "⚠ Warning: S3 bucket '$S3_BACKEND_BUCKET' not found in region '$AWS_REGION'"
    echo "   Attempting to create bucket..."
    if aws s3 mb "s3://$S3_BACKEND_BUCKET" --region "$AWS_REGION"; then
        echo "✓ Created S3 bucket: $S3_BACKEND_BUCKET"
        # Enable versioning for safety
        aws s3api put-bucket-versioning \
            --bucket "$S3_BACKEND_BUCKET" \
            --versioning-configuration Status=Enabled \
            --region "$AWS_REGION"
        echo "✓ Enabled versioning on bucket"
    else
        echo "❌ Failed to create S3 bucket. Please create it manually or check permissions."
        exit 1
    fi
fi

# 7. Verify key pair exists
echo "Verifying EC2 key pair..."
if ! aws ec2 describe-key-pairs --key-names "$KEY_NAME" --region "$AWS_REGION" >/dev/null 2>&1; then
    echo "❌ Error: EC2 key pair '$KEY_NAME' not found in region '$AWS_REGION'"
    echo "   Please run Phase 1 first to create the key pair, or create it manually."
    exit 1
fi

# 8. Verify IAM role exists
echo "Verifying IAM role..."
if ! aws iam get-role --role-name "$IAM_ROLE" >/dev/null 2>&1; then
    echo "❌ Error: IAM role '$IAM_ROLE' not found"
    echo "   Please create the IAM role or run Phase 1 first."
    exit 1
fi

# 9. Ensure PEM file is present for later phases
if [[ ! -f "sovereign-fl-key.pem" ]]; then
    echo "⚠ Warning: sovereign-fl-key.pem not found in terraform/ directory"
    echo "   Later phases (3-5) will need this file for SSH access."
    echo "   If you have the key elsewhere, please copy it to: ./terraform/sovereign-fl-key.pem"
fi

# 10. Execute Infrastructure Build
echo "Step 2.1: Initializing Terraform..."
terraform init -reconfigure

echo "Step 2.2: Planning Infrastructure..."
terraform plan -out=tfplan

echo "Step 2.3: Applying Infrastructure..."
terraform apply tfplan

# 11. Save outputs for later phases
echo "Step 2.4: Capturing deployment outputs..."
cat > ../deployment-outputs.env << EOF
AGGREGATOR_IP=$(terraform output -raw aggregator_ip 2>/dev/null || echo "")
WORKER_IPS=$(terraform output -json worker_ips 2>/dev/null || echo "[]")
DEPLOYMENT_TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
EOF

cd ..
echo "==========================================="
echo "✅ PHASE 2 COMPLETE"
echo "   Aggregator IP: $(grep AGGREGATOR_IP deployment-outputs.env | cut -d= -f2)"
echo "   Worker Count: $NODE_COUNT"
echo "==========================================="
