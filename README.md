# About

Automated outreach system built on Go, SQLite, Docker, deployed on AWS using github actions and terraform, leveraging AI-augmented workfow (agent mode and chatbots)


## Architecture

- Dev tech stack:
  - Go
  - SQLite

- DevOps stack
  - Docker
  - Github actions
  - Terraform
  - AWS

- AWS architecture:
  - Image registry: ECR
  - Compute: Lambda 
  - Data persistence: S3 bucket (versioned DB + CSV)
  - Scheduler: EventBridge Scheduler
  - Observability: Cloudwatch

