package service

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/yunuskargi/configbox/internal/crypto"
	"github.com/yunuskargi/configbox/internal/database"
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

func isPrivateIP(host string) bool {
	// Strip port using net.SplitHostPort (handles IPv6 brackets)
	h, _, err := net.SplitHostPort(host)
	if err != nil {
		h = host // no port
	}
	// Remove IPv6 brackets if still present
	h = strings.TrimPrefix(h, "[")
	h = strings.TrimSuffix(h, "]")

	// Resolve hostname
	ips, err := net.LookupIP(h)
	if err != nil {
		// If DNS fails, check raw IP
		ip := net.ParseIP(h)
		if ip == nil {
			return false
		}
		ips = []net.IP{ip}
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return true
		}
		// Block cloud metadata endpoints (169.254.169.254)
		if ip.Equal(net.ParseIP("169.254.169.254")) {
			return true
		}
	}
	return false
}

func validateS3Endpoint(endpoint string) error {
	if endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	host := endpoint
	if idx := strings.Index(host, "://"); idx != -1 {
		host = host[idx+3:]
	}
	if isPrivateIP(host) {
		return fmt.Errorf("S3 endpoint cannot point to private or internal addresses")
	}
	return nil
}

func newS3Client(s s3Settings) (*minio.Client, error) {
	if err := validateS3Endpoint(s.Endpoint); err != nil {
		return nil, err
	}
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
