package storage

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

type fakeS3 struct {
	putInput    *s3.PutObjectInput
	deleteInput *s3.DeleteObjectInput
	headErr     error
}

func (f *fakeS3) PutObject(_ context.Context, params *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	f.putInput = params
	return &s3.PutObjectOutput{}, nil
}

func (f *fakeS3) DeleteObject(_ context.Context, params *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	f.deleteInput = params
	return &s3.DeleteObjectOutput{}, nil
}

func (f *fakeS3) HeadObject(_ context.Context, _ *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	if f.headErr != nil {
		return nil, f.headErr
	}
	return &s3.HeadObjectOutput{}, nil
}

type fakeAPIError struct {
	code string
}

func (e *fakeAPIError) Error() string                 { return e.code }
func (e *fakeAPIError) ErrorCode() string             { return e.code }
func (e *fakeAPIError) ErrorMessage() string          { return e.code }
func (e *fakeAPIError) ErrorFault() smithy.ErrorFault { return smithy.FaultUnknown }

func newTestStore(client s3API) *Store {
	return &Store{client: client, bucket: "test-bucket", endpoint: "https://cdn.example.com"}
}

func TestPublicURL(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{name: "category-prefixed key", key: "player/uuid-1.png", want: "https://cdn.example.com/test-bucket/player/uuid-1.png"},
		{name: "legacy flat key with no category", key: "uuid-1.png", want: "https://cdn.example.com/test-bucket/uuid-1.png"},
		{name: "key with spaces gets escaped", key: "player/my file.png", want: "https://cdn.example.com/test-bucket/player/my%20file.png"},
	}

	s := newTestStore(&fakeS3{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.PublicURL(tt.key)
			if got != tt.want {
				t.Errorf("PublicURL(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestExistsReturnsTrueWhenHeadObjectSucceeds(t *testing.T) {
	s := newTestStore(&fakeS3{})

	exists, err := s.Exists(context.Background(), "player/uuid-1.png")
	if err != nil {
		t.Fatalf("Exists returned unexpected error: %v", err)
	}
	if !exists {
		t.Fatal("Exists = false, want true")
	}
}

func TestExistsReturnsFalseOnNotFound(t *testing.T) {
	s := newTestStore(&fakeS3{headErr: &types.NotFound{}})

	exists, err := s.Exists(context.Background(), "player/missing.png")
	if err != nil {
		t.Fatalf("Exists returned unexpected error: %v", err)
	}
	if exists {
		t.Fatal("Exists = true, want false")
	}
}

func TestExistsReturnsFalseOnAPIErrorNotFoundCode(t *testing.T) {
	s := newTestStore(&fakeS3{headErr: &fakeAPIError{code: "NotFound"}})

	exists, err := s.Exists(context.Background(), "player/missing.png")
	if err != nil {
		t.Fatalf("Exists returned unexpected error: %v", err)
	}
	if exists {
		t.Fatal("Exists = true, want false")
	}
}

func TestExistsReturnsErrorOnOtherFailures(t *testing.T) {
	s := newTestStore(&fakeS3{headErr: errors.New("connection refused")})

	_, err := s.Exists(context.Background(), "player/uuid-1.png")
	if err == nil {
		t.Fatal("Exists returned nil error, want non-nil")
	}
}

func TestPutSendsBucketKeyContentTypeAndSize(t *testing.T) {
	fake := &fakeS3{}
	s := newTestStore(fake)

	body := strings.NewReader("file contents")
	err := s.Put(context.Background(), "player/uuid-1.png", body, int64(body.Len()), "image/png")
	if err != nil {
		t.Fatalf("Put returned unexpected error: %v", err)
	}

	if aws.ToString(fake.putInput.Bucket) != "test-bucket" {
		t.Errorf("Bucket = %q, want %q", aws.ToString(fake.putInput.Bucket), "test-bucket")
	}
	if aws.ToString(fake.putInput.Key) != "player/uuid-1.png" {
		t.Errorf("Key = %q, want %q", aws.ToString(fake.putInput.Key), "player/uuid-1.png")
	}
	if aws.ToString(fake.putInput.ContentType) != "image/png" {
		t.Errorf("ContentType = %q, want %q", aws.ToString(fake.putInput.ContentType), "image/png")
	}
	if aws.ToInt64(fake.putInput.ContentLength) != int64(len("file contents")) {
		t.Errorf("ContentLength = %d, want %d", aws.ToInt64(fake.putInput.ContentLength), len("file contents"))
	}
}

func TestDeleteSendsBucketAndKey(t *testing.T) {
	fake := &fakeS3{}
	s := newTestStore(fake)

	err := s.Delete(context.Background(), "player/uuid-1.png")
	if err != nil {
		t.Fatalf("Delete returned unexpected error: %v", err)
	}

	if aws.ToString(fake.deleteInput.Bucket) != "test-bucket" {
		t.Errorf("Bucket = %q, want %q", aws.ToString(fake.deleteInput.Bucket), "test-bucket")
	}
	if aws.ToString(fake.deleteInput.Key) != "player/uuid-1.png" {
		t.Errorf("Key = %q, want %q", aws.ToString(fake.deleteInput.Key), "player/uuid-1.png")
	}
}
