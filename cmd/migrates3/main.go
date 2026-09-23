package main

import (
	"context"
	"fmt"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"

	"github.com/COMTOP1/AFC-GO/infrastructure/storage"
	"github.com/COMTOP1/AFC-GO/infrastructure/telemetry"
)

var tracer = otel.Tracer("github.com/COMTOP1/AFC-GO/cmd/migrates3")

func main() {
	_ = godotenv.Load(".env")
	_ = godotenv.Overload(".env.local")

	otelServiceName := os.Getenv("OTEL_SERVICE_NAME")
	if otelServiceName == "" {
		otelServiceName = "afc-go-migrates3"
	}

	ctx := context.Background()
	otelShutdown, err := telemetry.Setup(ctx, telemetry.Config{
		ServiceName: otelServiceName,
		Endpoint:    os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		Headers:     telemetry.ParseHeaders(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")),
	})
	if err != nil {
		slog.Error(fmt.Sprintf("failed to set up telemetry: %+v", err))
		os.Exit(1)
	}
	exitCode := run(ctx)

	if shutdownErr := otelShutdown(context.Background()); shutdownErr != nil {
		slog.Error(fmt.Sprintf("failed to shut down telemetry: %+v", shutdownErr))
	}
	os.Exit(exitCode)
}

func run(ctx context.Context) int {
	ctx, span := tracer.Start(ctx, "migrates3.Run")
	defer span.End()

	sourceDir := os.Getenv("FILESTORE_DIR")
	if sourceDir == "" {
		if stat, statErr := os.Stat("/FileStore"); statErr == nil && stat.IsDir() {
			sourceDir = "/FileStore"
		} else {
			sourceDir = "./FileStore"
		}
	}

	s3Region := os.Getenv("S3_REGION")
	if s3Region == "" {
		s3Region = "us-east-1"
	}

	store := storage.NewStore(ctx, storage.Config{
		Endpoint:  os.Getenv("S3_ENDPOINT"),
		Region:    s3Region,
		Bucket:    os.Getenv("S3_BUCKET"),
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
	})

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		span.RecordError(err)
		slog.Error("failed to read source directory " + sanitizeLogValue(sourceDir) + ": " + err.Error()) //nolint:gosec // sanitizeLogValue strips the newlines a log-injection attack relies on
		return 1
	}

	var uploaded, skipped, failed int

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		key := entry.Name()
		safeKey := sanitizeLogValue(key)

		exists, existsErr := store.Exists(ctx, key)
		if existsErr != nil {
			span.RecordError(existsErr)
			slog.Error(fmt.Sprintf("failed to check existence of %q: %+v", safeKey, existsErr)) //nolint:gosec // safeKey is sanitizeLogValue(key), newlines already stripped
			failed++
			continue
		}
		if exists {
			slog.Info(fmt.Sprintf("skipping %q: already exists in bucket", safeKey)) //nolint:gosec // safeKey is sanitizeLogValue(key), newlines already stripped
			skipped++
			continue
		}

		if uploadErr := uploadFile(ctx, store, sourceDir, key); uploadErr != nil {
			span.RecordError(uploadErr)
			slog.Error(fmt.Sprintf("failed to upload %q: %+v", safeKey, uploadErr)) //nolint:gosec // safeKey is sanitizeLogValue(key), newlines already stripped
			failed++
			continue
		}

		slog.Info(fmt.Sprintf("uploaded %q", safeKey)) //nolint:gosec // safeKey is sanitizeLogValue(key), newlines already stripped
		uploaded++
	}

	slog.Info(fmt.Sprintf("migration complete: uploaded=%d skipped=%d failed=%d total=%d",
		uploaded, skipped, failed, uploaded+skipped+failed))

	if failed > 0 {
		return 1
	}
	return 0
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
