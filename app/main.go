package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	lib "myapp/lib"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	defaultDBS3Key  = "mydata.db"
	defaultCSVS3Key = "prospects.csv"
	localDBPath     = "/tmp/mydata.db"
	localCSVPath    = "/tmp/prospects.csv"
)

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context) error {

	// --- S3 client --------------------------------------------------------
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		return fmt.Errorf("S3_BUCKET env var is required")
	}

	dbKey := os.Getenv("DB_S3_KEY")
	if dbKey == "" {
		dbKey = defaultDBS3Key
	}
	csvKey := os.Getenv("CSV_S3_KEY")
	if csvKey == "" {
		csvKey = defaultCSVS3Key
	}

	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("handleRequest: load aws config: %w", err)
	}
	s3Client := s3.NewFromConfig(awsCfg)

	// --- Logger -----------------------------------------------------------
	// CloudWatch captures stdout; do not write to a file (Lambda FS is read-only
	// except /tmp).
	cleanup, err := lib.InitGlobalLogger("")
	if err != nil {
		return fmt.Errorf("handleRequest: init logger: %w", err)
	}
	defer cleanup()

	slog.Info("lambda invocation started",
		"bucket", bucket,
		"db_key", dbKey,
		"csv_key", csvKey,
	)

	// --- Download SQLite DB (create a fresh one if missing) ---------------
	if err := lib.EnsureSQLite(ctx, s3Client, bucket, dbKey); err != nil {
		return fmt.Errorf("handleRequest: ensure SQLite: %w", err)
	}
	if err := lib.DownloadFile(ctx, s3Client, bucket, dbKey, localDBPath); err != nil {
		return fmt.Errorf("handleRequest: download db: %w", err)
	}
	slog.Info("SQLite downloaded", "path", localDBPath)

	// --- Download prospects CSV ------------------------------------------
	if err := lib.DownloadFile(ctx, s3Client, bucket, csvKey, localCSVPath); err != nil {
		return fmt.Errorf("handleRequest: download csv: %w", err)
	}
	slog.Info("CSV downloaded", "path", localCSVPath)

	// --- Point the business logic at the local copies --------------------
	os.Setenv("DB_PATH", localDBPath)
	os.Setenv("CSV_PATH", localCSVPath)

	// --- Run the business logic ------------------------------------------
	if err := lib.LoadProspectsFromCSV(); err != nil {
		slog.Error("LoadProspectsFromCSV failed", "err", err)
	}

	if err := lib.ProcessProspects(); err != nil {
		slog.Error("ProcessProspects failed", "err", err)
	}

	// --- Upload the modified DB back to S3 -------------------------------
	if err := lib.UploadFile(ctx, s3Client, bucket, dbKey, localDBPath); err != nil {
		return fmt.Errorf("handleRequest: upload db: %w", err)
	}
	slog.Info("SQLite uploaded back to S3", "bucket", bucket, "key", dbKey)

	slog.Info("lambda invocation finished")
	return nil
}
