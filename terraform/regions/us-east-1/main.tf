# Scheduled prospect-mailer: AWS Lambda (container) + SQLite-on-S3 + EventBridge Scheduler.

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# ---------------------------------------------------------------------------
# ECR repository for the Lambda container image
# ---------------------------------------------------------------------------
# Import block adopts the repo if it already exists (e.g. created manually or
# by a previous run), so apply is idempotent instead of failing on create.
import {
  to = aws_ecr_repository.app
  id = var.ecr_repo_name
}

resource "aws_ecr_repository" "app" {
  name                 = var.ecr_repo_name
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_lifecycle_policy" "app" {
  repository = aws_ecr_repository.app.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep only the last 10 images"
        action = {
          type = "expire"
        }
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 10
        }
      }
    ]
  })
}

# ---------------------------------------------------------------------------
# S3 bucket holding the SQLite DB and prospects CSV
# ---------------------------------------------------------------------------
# Import block adopts the bucket if it already exists (created by a prior
# run), so apply is idempotent instead of failing with BucketAlreadyExists.
import {
  to = aws_s3_bucket.app
  id = var.s3_bucket_name
}

resource "aws_s3_bucket" "app" {
  bucket = var.s3_bucket_name
}

resource "aws_s3_bucket_public_access_block" "app" {
  bucket                  = aws_s3_bucket.app.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_versioning" "app" {
  bucket = aws_s3_bucket.app.id
  versioning_configuration {
    status = "Enabled"
  }
}

# ---------------------------------------------------------------------------
# IAM role for the Lambda function
# ---------------------------------------------------------------------------
data "aws_iam_policy_document" "lambda_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

import {
  to = aws_iam_role.lambda
  id = "${var.ecr_repo_name}-lambda-role"
}

resource "aws_iam_role" "lambda" {
  name               = "${var.ecr_repo_name}-lambda-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
}

data "aws_iam_policy_document" "lambda_inline" {
  statement {
    sid       = "S3ReadWriteAppData"
    actions   = ["s3:GetObject", "s3:PutObject"]
    resources = ["${aws_s3_bucket.app.arn}/*"]
  }
}

resource "aws_iam_role_policy" "lambda_inline" {
  name   = "${var.ecr_repo_name}-lambda-s3"
  role   = aws_iam_role.lambda.id
  policy = data.aws_iam_policy_document.lambda_inline.json
}

resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# ---------------------------------------------------------------------------
# Lambda function (container image)
# ---------------------------------------------------------------------------
# Import block adopts the function if it already exists (created by a prior
# run), so apply is idempotent instead of failing with AlreadyExists.
import {
  to = aws_lambda_function.app
  id = var.ecr_repo_name
}

resource "aws_lambda_function" "app" {
  function_name = var.ecr_repo_name
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.app.repository_url}:${var.image_tag}"

  architectures                  = ["arm64"]
  reserved_concurrent_executions = 1
  timeout                        = var.lambda_timeout
  memory_size                    = var.lambda_memory
  role                           = aws_iam_role.lambda.arn

  environment {
    variables = {
      S3_BUCKET  = aws_s3_bucket.app.bucket
      DB_S3_KEY  = var.db_s3_key
      CSV_S3_KEY = var.csv_s3_key
      TZ         = "Asia/Bangkok"
      SENDER     = var.sender
      ZOHO_PWD   = var.zoho_pwd
      SMTP_HOST  = var.smtp_host
      SMTP_PORT  = var.smtp_port
    }
  }
}

# ---------------------------------------------------------------------------
# EventBridge Scheduler execution role (invokes the Lambda)
# ---------------------------------------------------------------------------
data "aws_iam_policy_document" "scheduler_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["scheduler.amazonaws.com"]
    }
  }
}

import {
  to = aws_iam_role.scheduler
  id = "${var.ecr_repo_name}-scheduler-role"
}

resource "aws_iam_role" "scheduler" {
  name               = "${var.ecr_repo_name}-scheduler-role"
  assume_role_policy = data.aws_iam_policy_document.scheduler_assume.json
}

data "aws_iam_policy_document" "scheduler_invoke" {
  statement {
    sid     = "InvokeLambda"
    actions = ["lambda:InvokeFunction"]
    resources = [
      aws_lambda_function.app.arn,
      "${aws_lambda_function.app.arn}:*",
    ]
  }
}

resource "aws_iam_role_policy" "scheduler_invoke" {
  name   = "${var.ecr_repo_name}-scheduler-invoke"
  role   = aws_iam_role.scheduler.id
  policy = data.aws_iam_policy_document.scheduler_invoke.json
}

# ---------------------------------------------------------------------------
# Three EventBridge schedules (Tue 06:00→midnight, Wed, Thu)
# ---------------------------------------------------------------------------
resource "aws_scheduler_schedule" "tue" {
  name       = "${var.ecr_repo_name}-tue"
  group_name = "default"

  schedule_expression          = "cron(0/15 6-23 ? * TUE *)"
  schedule_expression_timezone = "Asia/Bangkok"

  flexible_time_window {
    mode = "OFF"
  }

  target {
    arn      = aws_lambda_function.app.arn
    role_arn = aws_iam_role.scheduler.arn

    retry_policy {
      maximum_retry_attempts = 0
    }
  }
}

resource "aws_scheduler_schedule" "wed" {
  name       = "${var.ecr_repo_name}-wed"
  group_name = "default"

  schedule_expression          = "cron(0/15 * ? * WED *)"
  schedule_expression_timezone = "Asia/Bangkok"

  flexible_time_window {
    mode = "OFF"
  }

  target {
    arn      = aws_lambda_function.app.arn
    role_arn = aws_iam_role.scheduler.arn

    retry_policy {
      maximum_retry_attempts = 0
    }
  }
}

resource "aws_scheduler_schedule" "thu" {
  name       = "${var.ecr_repo_name}-thu"
  group_name = "default"

  schedule_expression          = "cron(0/15 * ? * THU *)"
  schedule_expression_timezone = "Asia/Bangkok"

  flexible_time_window {
    mode = "OFF"
  }

  target {
    arn      = aws_lambda_function.app.arn
    role_arn = aws_iam_role.scheduler.arn

    retry_policy {
      maximum_retry_attempts = 0
    }
  }
}
