#!/bin/bash
#================================================================================
# PHASE 2: INFRASTRUCTURE DEPLOYMENT (OPTIMIZED)
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

# 2. Setup Terraform Workspace
mkdir -p terraform
cd terraform

# 3. Create backend configuration
cat > backend.tf << 'EOF'
terraform {
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 5.0" }
  }
  backend "s3" {
    bucket = "sovereign-fl-terraform-state"
    key    = "200-node-test/terraform.tfstate"
    region = "us-east-1"
  }
}
EOF

# 4. Create variables.tf (FIXED: Using double quotes to resolve shell variables)
cat > variables.tf << EOF
variable "aws_region" { default = "$AWS_REGION" }
variable "node_count" { default = $NODE_COUNT }
variable "key_pair_name" { default = "$KEY_NAME" }
variable "iam_role" { default = "$IAM_ROLE" }
EOF

# 5. Execute Infrastructure Build
echo "Step 2.1: Initializing Terraform..."
terraform init -reconfigure

echo "Step 2.2: Applying Infrastructure..."
# Note: This assumes your main.tf is already in the /terraform folder.
# If main.tf is missing, the script will use the version currently in the repo.
terraform apply -auto-approve

cd ..
echo "=========================================="
echo "✅ PHASE 2 COMPLETE"
echo "=========================================="
