variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "image_tag" {
  description = "ECR image tag to deploy (e.g. git SHA)"
  type        = string
}

variable "s3_bucket_name" {
  description = "Name of the S3 bucket holding the SQLite DB and prospects CSV"
  type        = string
}

variable "ecr_repo_name" {
  description = "Name of the ECR repository (and Lambda function name)"
  type        = string
}

variable "lambda_timeout" {
  description = "Lambda timeout in seconds"
  type        = number
  default     = 120
}

variable "lambda_memory" {
  description = "Lambda memory size in MB"
  type        = number
  default     = 256
}

variable "db_s3_key" {
  description = "S3 object key for the SQLite DB"
  type        = string
  default     = "mydata.db"
}

variable "csv_s3_key" {
  description = "S3 object key for the prospects CSV"
  type        = string
  default     = "prospects.csv"
}

variable "sender" {
  description = "SMTP sender address (non-secret; default in dev.tfvars)"
  type        = string
  default     = "modernization@steadypartner.online"
}

variable "zoho_pwd" {
  description = "Zoho SMTP password (SECRET — supplied via GitHub secret / -var)"
  type        = string
  sensitive   = true
}

variable "smtp_host" {
  description = "SMTP host (non-secret; default in dev.tfvars)"
  type        = string
  default     = "smtp.zoho.com"
}

variable "smtp_port" {
  description = "SMTP port (non-secret; default in dev.tfvars)"
  type        = string
  default     = "465"
}

variable "resend_api_url" {
  description = "resend POST api url (non-secret; default in dev.tfvars)"
  type        = string
  default     = "https://api.resend.com/emails"
}

variable "resend_api_key" {
  description = "resend API KEY (SECRET — supplied via GitHub secret / -var)"
  type        = string
  sensitive   = true
}

variable "new_sender" {
  description = "SMTP new sender address (non-secret; default in dev.tfvars)"
  type        = string
  default     = "modernization@steadypartner.co"
}