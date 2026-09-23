package main

import (
	"context"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"

	"github.com/COMTOP1/AFC-GO/infrastructure/storage"
)

func main() {
	_ = godotenv.Load(".env")
	_ = godotenv.Overload(".env.local")

	sourceDir := os.Getenv("FILESTORE_DIR")
	if sourceDir == "" {
		if stat, err := os.Stat("/FileStore"); err == nil && stat.IsDir() {
			sourceDir = "/FileStore"
		} else {
			sourceDir = "./FileStore"
		}
	}

	s3Region := os.Getenv("S3_REGION")
	if s3Region == "" {
		s3Region = "us-east-1"
	}

	store := storage.NewStore(storage.Config{
		Endpoint:  os.Getenv("S3_ENDPOINT"),
		Region:    s3Region,
		Bucket:    os.Getenv("S3_BUCKET"),
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
	})

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		log.Fatalf("failed to read source directory %s: %+v", sanitizeLogValue(sourceDir), err) //nolint:gosec // sanitizeLogValue strips the newlines a log-injection attack relies on
	}

	ctx := context.Background()

	var uploaded, skipped, failed int

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		key := entry.Name()
		safeKey := sanitizeLogValue(key)

		exists, err := store.Exists(ctx, key)
		if err != nil {
			log.Printf("failed to check existence of %q: %+v", safeKey, err) //nolint:gosec // safeKey is sanitizeLogValue(key), newlines already stripped
			failed++
			continue
		}
		if exists {
			log.Printf("skipping %q: already exists in bucket", safeKey) //nolint:gosec // safeKey is sanitizeLogValue(key), newlines already stripped
			skipped++
			continue
		}

		if err = uploadFile(ctx, store, sourceDir, key); err != nil {
			log.Printf("failed to upload %q: %+v", safeKey, err) //nolint:gosec // safeKey is sanitizeLogValue(key), newlines already stripped
			failed++
			continue
		}

		log.Printf("uploaded %q", safeKey) //nolint:gosec // safeKey is sanitizeLogValue(key), newlines already stripped
		uploaded++
	}

	log.Printf("migration complete: uploaded=%d skipped=%d failed=%d total=%d",
		uploaded, skipped, failed, uploaded+skipped+failed)

	if failed > 0 {
		os.Exit(1)
	}
}

func uploadFile(ctx context.Context, store *storage.Store, sourceDir, key string) error {
	path, err := safeJoin(sourceDir, key)
	if err != nil {
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat %q: %w", path, err)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open %q: %w", path, err)
	}
	defer f.Close()

	contentType := mime.TypeByExtension(filepath.Ext(key))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if err = store.Put(ctx, key, f, info.Size(), contentType); err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}

	return nil
}

// safeJoin joins dir and name, rejecting any name that could escape dir via path traversal.
// Directory entries from os.ReadDir are always bare file names, but this guards against a
// crafted or unexpected entry regardless.
func safeJoin(dir, name string) (string, error) {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("invalid file name %q", sanitizeLogValue(name))
	}
	return filepath.Join(dir, name), nil
}

// sanitizeLogValue strips characters that would let a filesystem or environment value forge
// extra log lines (log injection) when written to the log.
func sanitizeLogValue(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, s)
}
