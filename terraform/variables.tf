variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "node_count" {
  description = "Number of worker nodes"
  type        = number
  default     = 1 # Scaled down from 100 for testing
}

variable "key_pair_name" {
  description = "Name of the SSH key pair"
  type        = string
  default     = "sovereign-fl-key"
}

variable "iam_role" {
  description = "IAM role name for EC2"
  type        = string
  default     = "SovereignFL-EC2-Role"
}
