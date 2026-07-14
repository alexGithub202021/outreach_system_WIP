# main.tf

provider "aws" {
  region = var.aws_region
}

data "aws_instances" "existing" {
  filter {
    name   = "tag:Name"
    values = [var.instance_name]
  }
  filter {
    name   = "instance-type"
    values = [var.instance_type]
  }
}

resource "aws_instance" "app" {
  count = length(data.aws_instances.existing.ids) == 0 ? 1 : 0
  ami                    = var.ami_id
  instance_type          = var.instance_type
  key_name               = var.key_name
  vpc_security_group_ids = [var.security_group_id]
  # subnet_id            = var.subnet_id  # Uncomment if using

  tags = {
    Name = var.instance_name
  }
}