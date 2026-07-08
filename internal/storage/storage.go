package storage

import (
	"context"
	"fmt"
	"io"
	"time"
)

// BlobStore persists and retrieves document blobs.
type BlobStore interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, key string) error
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
}

// Key builds a storage path for a document.
func Key(orgID, category, ownerID, docID, filename string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", orgID, category, ownerID, docID, filename)
}
