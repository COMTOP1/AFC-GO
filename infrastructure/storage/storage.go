package storage

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("github.com/COMTOP1/AFC-GO/infrastructure/storage")

type (
	// Config holds the connection details for an S3-compatible object store.
	Config struct {
		Endpoint  string
		Region    string
		Bucket    string
		AccessKey string
		SecretKey string
	}

	// s3API is the subset of *s3.Client this package depends on, so tests can supply a fake.
	s3API interface {
		PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
		DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
		HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	}

	// Store puts, deletes, checks and links objects in an S3-compatible bucket.
	Store struct {
		client   s3API
		bucket   string
		endpoint string
	}
)

// NewStore builds a Store connected to the S3-compatible endpoint described by cfg.
//
// The HTTP client forces HTTP/1.1: the Ceph RadosGW endpoint this app targets returns a
// PROTOCOL_ERROR over HTTP/2.
func NewStore(ctx context.Context, cfg Config) *Store {
	if cfg.Endpoint == "" || cfg.Bucket == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		slog.Error("storage: missing required S3 config (endpoint, bucket, access key, and secret key must all be set)")
		os.Exit(1)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{},
			ForceAttemptHTTP2: false,
		},
	}

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
		config.WithHTTPClient(httpClient),
	)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to load aws config for storage: %+v", err))
		os.Exit(1)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	slog.Info("connected to storage: " + cfg.Endpoint)

	return &Store{
		client:   client,
		bucket:   cfg.Bucket,
		endpoint: cfg.Endpoint,
	}
}

// Put uploads body to key, setting contentType and size (Content-Length) on the object.
func (s *Store) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	ctx, span := tracer.Start(ctx, "storage.Put")
	defer span.End()

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to put object %q: %w", key, err)
	}
	return nil
}

// Delete removes key from the bucket.
func (s *Store) Delete(ctx context.Context, key string) error {
	ctx, span := tracer.Start(ctx, "storage.Delete")
	defer span.End()

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete object %q: %w", key, err)
	}
	return nil
}

// Exists reports whether key is present in the bucket.
func (s *Store) Exists(ctx context.Context, key string) (bool, error) {
	ctx, span := tracer.Start(ctx, "storage.Exists")
	defer span.End()

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		return true, nil
	}

	var notFound *types.NotFound
	if errors.As(err, &notFound) {
		return false, nil
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "404") {
		return false, nil
	}

	span.RecordError(err)
	return false, fmt.Errorf("failed to head object %q: %w", key, err)
}

// PublicURL builds the public CDN URL for key, e.g. "https://cdn.example.com/bucket/category/uuid.png".
func (s *Store) PublicURL(key string) string {
	segments := strings.Split(key, "/")
	escaped := make([]string, len(segments))
	for i, segment := range segments {
		escaped[i] = url.PathEscape(segment)
	}
	return strings.TrimRight(s.endpoint, "/") + "/" + url.PathEscape(s.bucket) + "/" + strings.Join(escaped, "/")
}
