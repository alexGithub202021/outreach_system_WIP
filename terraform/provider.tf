terraform {
  required_version = ">= 1.9" # optional

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.70" # optional
    }
  }
}

provider "aws" {
  region = regions.var.aws_region
}