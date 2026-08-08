package main

import (
	"context"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

	"github.com/COMTOP1/AFC-GO/infrastructure/storage"
)

func main() {
	_ = godotenv.Load(".env")
	_ = godotenv.Overload(".env.local")

	sourceDir := os.Getenv("FILESTORE_DIR")
	if sourceDir == "" {
		sourceDir = "./FileStore"
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
		log.Fatalf("failed to read source directory %q: %+v", sourceDir, err)
	}

	ctx := context.Background()

	var uploaded, skipped, failed int

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		key := entry.Name()

		exists, err := store.Exists(ctx, key)
		if err != nil {
			log.Printf("failed to check existence of %q: %+v", key, err)
			failed++
			continue
		}
		if exists {
			log.Printf("skipping %q: already exists in bucket", key)
			skipped++
			continue
		}

		if err = uploadFile(ctx, store, sourceDir, key); err != nil {
			log.Printf("failed to upload %q: %+v", key, err)
			failed++
			continue
		}

		log.Printf("uploaded %q", key)
		uploaded++
	}

	log.Printf("migration complete: uploaded=%d skipped=%d failed=%d total=%d",
		uploaded, skipped, failed, uploaded+skipped+failed)

	if failed > 0 {
		os.Exit(1)
	}
}

func uploadFile(ctx context.Context, store *storage.Store, sourceDir, key string) error {
	path := filepath.Join(sourceDir, key)

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
