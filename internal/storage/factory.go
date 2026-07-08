package storage

import (
	"fmt"
	"time"

	"github.com/nivas/server/internal/config"
)

func New(cfg config.StorageConfig) (BlobStore, error) {
	switch cfg.Driver {
	case "s3":
		return NewS3Store(cfg)
	default:
		return NewLocalStore(cfg.LocalPath)
	}
}

// NewFromConfig is an alias for New.
func NewFromConfig(cfg config.StorageConfig) (BlobStore, error) {
	return New(cfg)
}

// IsPresigned returns true when downloads should use presigned URLs.
func IsPresigned(cfg config.StorageConfig) bool {
	return cfg.Driver == "s3"
}

// DownloadTTL is how long presigned URLs remain valid.
func DownloadTTL(cfg config.StorageConfig) time.Duration {
	if cfg.PresignTTL > 0 {
		return cfg.PresignTTL
	}
	return 15 * time.Minute
}

// ValidateContentType checks allowed MIME types for uploads.
func ValidateContentType(contentType string) error {
	switch contentType {
	case "application/pdf", "image/jpeg", "image/png", "image/webp":
		return nil
	default:
		return fmt.Errorf("unsupported file type: %s (allowed: PDF, JPEG, PNG, WebP)", contentType)
	}
}
