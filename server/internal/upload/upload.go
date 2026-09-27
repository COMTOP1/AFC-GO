// Package upload validates and stores user-uploaded files in object storage.
package upload

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"

	"github.com/google/uuid"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
)

// File is an uploaded file, decoupled from multipart so services and tests
// don't depend on HTTP types.
type File struct {
	Name        string
	ContentType string
	Size        int64
	Open        func() (io.ReadCloser, error)
}

// FromHeader adapts a multipart upload; it returns nil for a nil header.
func FromHeader(fh *multipart.FileHeader) *File {
	if fh == nil {
		return nil
	}
	return &File{
		Name:        fh.Filename,
		ContentType: fh.Header.Get("Content-Type"),
		Size:        fh.Size,
		Open:        func() (io.ReadCloser, error) { return fh.Open() },
	}
}

// Storage is the object store (satisfied by *storage.Store).
type Storage interface {
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	PublicURL(key string) string
}

// Files saves and removes uploads.
type Files struct {
	store Storage
}

func New(store Storage) *Files {
	return &Files{store: store}
}

// extensions is the allow-list of upload content types (unchanged from the
// legacy fileUpload helper).
var extensions = map[string]string{
	"application/pdf": ".pdf",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   ".docx",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": ".pptx",
	"text/plain":    ".txt",
	"image/apng":    ".apng",
	"image/avif":    ".avif",
	"image/gif":     ".gif",
	"image/jpeg":    ".jpg",
	"image/png":     ".png",
	"image/svg+xml": ".svg",
	"image/webp":    ".webp",
}

// Save checks the content type against the allow-list, uploads the file as
// <category>/<uuid><ext>, and returns the object key.
func (f *Files) Save(ctx context.Context, file *File, category string) (string, error) {
	ext, ok := extensions[file.ContentType]
	if !ok {
		return "", svcerr.InvalidField("file", "unsupported file type: "+file.ContentType)
	}
	key := category + "/" + uuid.NewString() + ext

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open upload: %w", err)
	}
	defer func() { _ = src.Close() }()

	if err = f.store.Put(ctx, key, src, file.Size, file.ContentType); err != nil {
		return "", fmt.Errorf("failed to store upload: %w", err)
	}
	return key, nil
}

// Remove deletes an object. Failures are logged, not returned: as in the
// legacy views, a failed S3 delete never blocks the database change.
func (f *Files) Remove(ctx context.Context, key string) {
	if key == "" {
		return
	}
	if err := f.store.Delete(ctx, key); err != nil {
		slog.InfoContext(ctx, fmt.Sprintf("failed to delete object %q: %+v", key, err))
	}
}

// URL returns the public URL for key, or "" when key is empty.
func (f *Files) URL(key string) string {
	if key == "" {
		return ""
	}
	return f.store.PublicURL(key)
}

// Exists reports whether key is present in the store.
func (f *Files) Exists(ctx context.Context, key string) (bool, error) {
	return f.store.Exists(ctx, key)
}
