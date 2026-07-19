package lib

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// localDBPath is the on-disk location the SQLite DB is read from / written to
// inside the Lambda (only /tmp is writable). Shared with the CSV loader via
// the DB_PATH env var set by the handler.
const localDBPath = "/tmp/mydata.db"

// DownloadFile downloads an object from S3 to a local path. A missing object
// (NoSuchKey) surfaces as an error to the caller so retry logic stays explicit.
func DownloadFile(ctx context.Context, client *s3.Client, bucket, key, dest string) error {
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("downloadFile: create %q: %w", dest, err)
	}
	defer f.Close()

	downloader := manager.NewDownloader(client)
	if _, err := downloader.Download(ctx, f, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}); err != nil {
		return fmt.Errorf("downloadFile: get %q/%q: %w", bucket, key, err)
	}
	return nil
}

// uploadFile uploads a local file to S3, overwriting the object.
func UploadFile(ctx context.Context, client *s3.Client, bucket, key, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("uploadFile: open %q: %w", src, err)
	}
	defer f.Close()

	uploader := manager.NewUploader(client)
	if _, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   f,
	}); err != nil {
		return fmt.Errorf("uploadFile: put %q/%q: %w", bucket, key, err)
	}
	return nil
}

// objectExists reports whether the object is present in S3.
func objectExists(ctx context.Context, client *s3.Client, bucket, key string) (bool, error) {
	_, err := client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return false, nil
		}
		var nf *types.NotFound
		if errors.As(err, &nf) {
			return false, nil
		}
		return false, fmt.Errorf("objectExists: head %q/%q: %w", bucket, key, err)
	}
	return true, nil
}

// ensureSQLite makes sure a usable SQLite DB exists in S3. If the object is
// missing, a fresh local DB with the prospects table is created and uploaded,
// so the first run after a new empty bucket self-initializes.
func EnsureSQLite(ctx context.Context, client *s3.Client, bucket, key string) error {
	exists, err := objectExists(ctx, client, bucket, key)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	slog.Info("ensureSQLite: object not found, creating fresh DB", "bucket", bucket, "key", key)

	if err := initSQLiteSchema(localDBPath); err != nil {
		return fmt.Errorf("ensureSQLite: init schema: %w", err)
	}
	if err := UploadFile(ctx, client, bucket, key, localDBPath); err != nil {
		return fmt.Errorf("ensureSQLite: upload fresh db: %w", err)
	}
	return nil
}

// initSQLiteSchema opens (creating if needed) a SQLite file at path and ensures
// the prospects table exists. It uses the same columns/import as the app.
func initSQLiteSchema(path string) error {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("initSQLiteSchema: open %q: %w", path, err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("initSQLiteSchema: ping %q: %w", path, err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS prospects (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			name_p      TEXT,
			email       TEXT,
			gender      TEXT,
			tmz         TEXT,
			sending_date TEXT,
			sending_time TEXT,
			status_p     INTEGER DEFAULT 0
		)`)
	if err != nil {
		return fmt.Errorf("initSQLiteSchema: create table: %w", err)
	}
	return nil
}
