terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

variable "aws_region" {
  default = "us-east-1"
}

# SET TO 10 TO STAY UNDER 32 VCPU LIMIT
variable "node_count" {
  default = 10 
}

variable "key_pair_name" {
  default = "sovereign-fl-key"
}

data "aws_ami" "ubuntu" {
  most_recent = true
  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }
  owners = ["099720109477"]
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"
  name    = "sovereign-fl-vpc"
  cidr    = "10.0.0.0/16"

  azs             = ["us-east-1a", "us-east-1b", "us-east-1c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

  enable_nat_gateway = true
  single_nat_gateway = true 
}

resource "aws_security_group" "aggregator" {
  name_prefix = "sovereign-aggregator-"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "client" {
  name_prefix = "sovereign-client-"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "aggregator" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = "t3.medium" # Changed from c5.2xlarge to stay under 32 vCPU limit
  key_name               = var.key_pair_name
  vpc_security_group_ids = [aws_security_group.aggregator.id]
  subnet_id              = module.vpc.public_subnets[0]
  tags = {
    Name = "sovereign-aggregator"
  }
}

resource "aws_launch_template" "client" {
  name_prefix   = "sovereign-client-lt-"
  image_id      = data.aws_ami.ubuntu.id
  key_name      = var.key_pair_name
  vpc_security_group_ids = [aws_security_group.client.id]
}

resource "aws_autoscaling_group" "clients" {
  name                = "sovereign-fl-clients"
  vpc_zone_identifier = module.vpc.private_subnets
  min_size            = 0
  max_size            = 20
  desired_capacity    = var.node_count

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.client.id
        version            = "$Latest"
      }
      override { instance_type = "t3.medium" }
    }
  }
}

output "aggregator_ip" {
  value = aws_instance.aggregator.public_ip
}
