package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/yunuskargi/confbox/internal/crypto"
	"github.com/yunuskargi/confbox/internal/database"
)

type s3Settings struct {
	Enabled   bool
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Prefix    string
}

func getS3Settings() s3Settings {
	get := func(key, def string) string {
		var val *string
		err := database.DB.Get(&val, "SELECT value FROM settings WHERE key = ?", key)
		if err != nil || val == nil {
			return def
		}
		return *val
	}

	return s3Settings{
		Enabled:   get("s3_enabled", "false") == "true",
		Endpoint:  get("s3_endpoint", ""),
		Region:    get("s3_region", "us-east-1"),
		Bucket:    get("s3_bucket", ""),
		AccessKey: crypto.Decrypt(get("s3_access_key", "")),
		SecretKey: crypto.Decrypt(get("s3_secret_key", "")),
		UseSSL:    get("s3_use_ssl", "true") == "true",
		Prefix:    get("s3_prefix", ""),
	}
}

func newS3Client(s s3Settings) (*minio.Client, error) {
	return minio.New(s.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.AccessKey, s.SecretKey, ""),
		Secure: s.UseSSL,
		Region: s.Region,
	})
}

func uploadToS3(filePath, vendor, deviceName string) error {
	s := getS3Settings()
	if !s.Enabled || s.Endpoint == "" || s.Bucket == "" {
		return nil
	}

	client, err := newS3Client(s)
	if err != nil {
		return fmt.Errorf("S3 client error: %v", err)
	}

	fileName := filepath.Base(filePath)
	objectKey := fmt.Sprintf("%s%s/%s/%s", s.Prefix, vendor, deviceName, fileName)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	fi, _ := file.Stat()
	_, err = client.PutObject(ctx, s.Bucket, objectKey, file, fi.Size(), minio.PutObjectOptions{
		ContentType: "text/plain",
	})
	if err != nil {
		return fmt.Errorf("S3 upload failed: %v", err)
	}

	slog.Info("backup uploaded to S3", "bucket", s.Bucket, "key", objectKey)
	return nil
}

func TestS3Connection() error {
	s := getS3Settings()
	if s.Endpoint == "" || s.Bucket == "" {
		return fmt.Errorf("S3 endpoint and bucket are required")
	}

	client, err := newS3Client(s)
	if err != nil {
		return fmt.Errorf("S3 client error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, s.Bucket)
	if err != nil {
		return fmt.Errorf("S3 connection failed: %v", err)
	}
	if !exists {
		return fmt.Errorf("bucket '%s' does not exist", s.Bucket)
	}

	return nil
}
