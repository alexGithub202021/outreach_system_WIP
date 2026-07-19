output "lambda_function_name" {
  description = "Name of the provisioned Lambda function"
  value       = aws_lambda_function.app.function_name
}

output "ecr_repository_url" {
  description = "URL of the ECR repository hosting the Lambda image"
  value       = aws_ecr_repository.app.repository_url
}

output "s3_bucket_name" {
  description = "Name of the S3 bucket holding app data"
  value       = aws_s3_bucket.app.bucket
}
