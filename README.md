# Go project

Prospeccting batch job deployed as a scheduled AWS Lambda (container image on ECR)
with SQLite and `prospects.csv` stored in S3, triggered by EventBridge Scheduler.

## About

The Go binary is a **run-once batch job** (`app/main.go`) wired into the AWS Lambda
runtime via `github.com/aws/lambda-go`. On each invocation it:

1. Downloads the SQLite DB and `prospects.csv` from S3 to `/tmp`.
2. Loads the CSV into the `prospects` table and processes pending prospects
   (sending emails, then marking them `status_p = 1`).
3. Uploads the modified DB back to S3.

SQLite uses `modernc.org/sqlite` (pure Go, **no CGO**), so it runs on the
Lambda Graviton (`arm64`) runtime without native extensions.

### Architecture

| Concern        | Service                    |
| -------------- | -------------------------- |
| Compute        | AWS Lambda (`arm64`, Image) |
| Image registry | Amazon ECR (`prospect-mailer`) |
| Data store     | S3 bucket (versioned DB + CSV) |
| Scheduler      | EventBridge Scheduler (Tue 06:00→midnight, Wed, Thu; every 15 min, `Asia/Bangkok`) |
| IaC            | Terraform (`terraform/`)    |
| CI/CD          | GitHub Actions (`.github/workflows/main.yml`) |

### Secrets

Runtime secrets are supplied as Lambda environment variables from GitHub
secrets: `SENDER`, `ZOHO_PWD`, `SMTP_HOST`, `SMTP_PORT` (plus `AWS_ACCESS_KEY_ID`
/ `AWS_SECRET_ACCESS_KEY` for CI auth). The app no longer reads a local `.env`.

## One-time manual prerequisites

1. Apply Terraform once (creates the S3 bucket, ECR repo, Lambda, IAM, schedules).
2. Seed S3 with the initial data (the DB self-initializes if missing, but seeding
   avoids the first empty run):
   ```sh
   aws s3 cp SQLite/data/mydata.db s3://<bucket>/mydata.db
   aws s3 cp prospects.csv          s3://<bucket>/prospects.csv
   ```
 3. Add the GitHub repository secrets listed above.

### Terraform remote state (S3)

State is stored in a dedicated S3 bucket (`prospect-mailer-terraform-state`,
configured in `terraform/provider.tf`'s `backend "s3"` block). The CI pipeline
creates this bucket automatically on the first run if it does not exist, so no
manual step is required. For a one-off local run, provision it with:

```sh
terraform -chdir=terraform/bootstrap init
terraform -chdir=terraform/bootstrap apply
```

## Build / local validation

```sh
# Compile and vet the app
cd app && go build ./... && go vet ./...

# Build the Lambda image for arm64
docker build --platform linux/arm64 ./app

# Validate Terraform
terraform -chdir=terraform/regions/us-east-1 init
terraform -chdir=terraform/regions/us-east-1 validate
```

After deploy, invoke the Lambda and confirm CloudWatch shows the S3 download,
CSV load, processing, and the DB upload (a new S3 version with `status_p` flipped).
