terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.33.0"
    }
  }

  backend "s3" {
    bucket  = "sovereign-fl-results-722762405531"
    key     = "terraform/terraform.tfstate"
    region  = "us-east-1"
    encrypt = true
  }
}

provider "aws" {
  region = var.aws_region
}

# 1. Network Infrastructure
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  tags = { Name = "sovereign-fl-vpc" }
}

resource "aws_subnet" "main" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true
  availability_zone       = "${var.aws_region}a"
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
}

resource "aws_route_table" "main" {
  vpc_id = aws_vpc.main.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
}

resource "aws_route_table_association" "main" {
  subnet_id      = aws_subnet.main.id
  route_table_id = aws_route_table.main.id
}

# 2. Security Group
resource "aws_security_group" "main" {
  name   = "sovereign-fl-sg"
  vpc_id = aws_vpc.main.id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["10.0.0.0/16"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# 3. CloudWatch Log Groups
resource "aws_cloudwatch_log_group" "aggregator" {
  name              = "sovereign-fl/aggregator"
  retention_in_days = 7
}

resource "aws_cloudwatch_log_group" "clients" {
  name              = "sovereign-fl/clients"
  retention_in_days = 7
}

# 4. Compute Resources
resource "aws_instance" "aggregator" {
  ami                    = "ami-0c7217cdde317cfec"
  instance_type          = "c5.2xlarge"
  subnet_id              = aws_subnet.main.id
  vpc_security_group_ids = [aws_security_group.main.id]
  key_name               = var.key_pair_name
  iam_instance_profile   = var.iam_role

  tags = { Name = "sovereign-aggregator" }
}

resource "aws_launch_template" "client" {
  name          = "sovereign-client-template"
  image_id      = "ami-0c7217cdde317cfec"
  instance_type = "m7g.4xlarge"
  key_name      = var.key_pair_name

  iam_instance_profile {
    name = var.iam_role
  }

  network_interfaces {
    associate_public_ip_address = true
    security_groups             = [aws_security_group.main.id]
    subnet_id                   = aws_subnet.main.id
  }
}

resource "aws_instance" "worker" {
  count = var.node_count

  launch_template {
    id      = aws_launch_template.client.id
    version = "$Latest"
  }

  tags = { Name = "sovereign-worker-${count.index}" }
}

# 5. Outputs
output "aggregator_ip" {
  value = aws_instance.aggregator.public_ip
}

output "worker_ips" {
  value = aws_instance.worker[*].public_ip
}
