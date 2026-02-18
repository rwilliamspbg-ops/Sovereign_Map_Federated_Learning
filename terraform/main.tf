# Configure the AWS Provider
provider "aws" {
  region = "us-east-1"
}

# 1. Generate Key Pair for SSH Access
resource "aws_key_pair" "sovereign_key" {
  key_name   = "sovereign-fl-key"
  public_key = file("~/.ssh/id_rsa.pub") # Ensure this path is correct for your local machine
}

# 2. Define Security Group with Port Fixes
resource "aws_security_group" "sovereign_fl_sg" {
  name        = "sovereign-fl-sg"
  description = "Security group for Sovereign Federated Learning 200-Node Test"

  # SSH Access
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Federated Learning Aggregator API (Step 4.1 Fix)
  ingress {
    from_port   = 5000
    to_port     = 5000
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Frontend HUD / Dashboard
  ingress {
    from_port   = 3000
    to_port     = 3000
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Prometheus Metrics (Fix for 98.84.34.57:9090)
  ingress {
    from_port   = 9090
    to_port     = 9090
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # MongoDB (Optional: internal access only usually, but open for debugging)
  ingress {
    from_port   = 27017
    to_port     = 27017
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Outbound Traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# 3. Define the EC2 Instance (Master Node)
resource "aws_instance" "master_node" {
  ami           = "ami-0e2c8ccd4e0269736" # Ubuntu 22.04 LTS in us-east-1
  instance_type = "t3.xlarge"             # Recommended for 200-node simulation
  key_name      = aws_key_pair.sovereign_key.key_name
  vpc_security_group_ids = [aws_security_group.sovereign_fl_sg.id]

  root_block_device {
    volume_size = 40
    volume_type = "gp3"
  }

  tags = {
    Name = "Sovereign-FL-Master"
  }
}

# 4. Output the IP Address for easy access
output "instance_public_ip" {
  value = aws_instance.master_node.public_ip
}