variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "ami_id" {
  description = "AMI ID (ubuntu)"
  type        = string
}

variable "instance_type" {
  description = "EC2 instance type - ARM based, 2GB RAM"
  type        = string
}

variable "key_name" {
  description = "Existing EC2 key pair name"
  type        = string
}

variable "security_group_id" {
  description = "Existing security group ID"
  type        = string
}

# variable "subnet_id" {
#   description = "Existing subnet ID"
#   type        = string
# }

variable "instance_name" {
  description = "Tag name for the EC2 instance"
  type        = string
}