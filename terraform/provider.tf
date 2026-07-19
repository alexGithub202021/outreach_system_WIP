terraform {
  required_version = ">= 1.9"

  # Remote backend (S3 only — no DynamoDB lock; the pipeline runs a single
  # region per push, so concurrent applies are not a risk).
  # The state bucket must be created once before the first `terraform init`
  # (see terraform/bootstrap/).
  backend "s3" {
    bucket = "prospect-mailer-terraform-state"
    key    = "us-east-1/prospect-mailer.tfstate"
    region = "us-east-1"
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.70, < 7.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}
