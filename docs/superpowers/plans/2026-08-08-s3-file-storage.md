# S3 File Storage Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace local-disk file storage (`v.conf.FileDir`, `os.Create`/`os.Remove`, `/file/<name>` redirect) with an S3-compatible object store, reachable over a public CDN endpoint.

**Architecture:** A new `infrastructure/storage` package wraps an AWS SDK v2 S3 client configured for a Ceph RadosGW endpoint (custom base endpoint, path-style addressing, forced HTTP/1.1). `views.Views` gains a `storage *storage.Store` field; `fileUpload` and every delete/download call site is switched from local-disk operations to `Put`/`Delete`/`Exists`/`PublicURL` calls against that store.

**Tech Stack:** Go 1.26, `github.com/aws/aws-sdk-go-v2` (`config`, `credentials`, `service/s3`, `service/s3/types`), `github.com/aws/smithy-go`, echo v4 (existing).

Design doc: `docs/superpowers/specs/2026-08-08-s3-file-storage-design.md`

## Global Constraints

- New AFC-specific bucket and credentials, supplied later via env vars (`S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`) — never hardcode bucket name, endpoint, or credentials.
- Public bucket, direct CDN URLs — no presigned URLs, no auth check on the object URL itself (matches current behavior).
- Uploads stay server-side: the Go backend calls `PutObject` directly using static credentials from env vars; no browser-to-S3 direct upload flow.
- Object keys for new uploads are `"<category>/<uuid>.<ext>"`; the full key is stored verbatim in the DB `FileName` column, exactly as today's flat filename is, so every downstream consumer keeps treating `FileName` as an opaque key.
- The migration script uploads existing files under their existing flat filename (no category prefix) so current DB rows resolve unchanged — it does not rewrite the database.
- Local disk storage is removed entirely once all call sites are migrated — no dual-write, no fallback.
- This codebase has no existing test files anywhere. Follow that convention for view-handler wiring (mechanical, hard to test without a live DB/echo harness) — verify those changes with `go build ./...` / `go vet ./...` only. The one exception is the new `infrastructure/storage` package: its `Put`/`Delete`/`Exists`/`PublicURL` logic is pure/mockable and gets real unit tests, matching the value TDD provides for genuinely testable logic.

---

### Task 1: `infrastructure/storage` package

**Files:**
- Create: `infrastructure/storage/storage.go`
- Create: `infrastructure/storage/storage_test.go`
- Modify: `go.mod`, `go.sum` (via `go get`)

**Interfaces:**
- Produces:
  - `storage.Config struct { Endpoint, Region, Bucket, AccessKey, SecretKey string }`
  - `storage.NewStore(cfg Config) *Store`
  - `(*Store) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error`
  - `(*Store) Delete(ctx context.Context, key string) error`
  - `(*Store) Exists(ctx context.Context, key string) (bool, error)`
  - `(*Store) PublicURL(key string) string`

- [ ] **Step 1: Add the AWS SDK dependencies**

Run:
```bash
go get github.com/aws/aws-sdk-go-v2/aws
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/credentials
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/aws/smithy-go
go mod tidy
```

- [ ] **Step 2: Write the failing tests**

Create `infrastructure/storage/storage_test.go`:

```go
package storage

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `go test ./infrastructure/storage/... -v`
Expected: FAIL — build fails with `undefined: Store`, `undefined: s3API` (storage.go doesn't exist yet).

- [ ] **Step 4: Write the implementation**

Create `infrastructure/storage/storage.go`:

```go
package storage

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

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
func NewStore(cfg Config) *Store {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{},
			ForceAttemptHTTP2: false,
		},
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
		config.WithHTTPClient(httpClient),
	)
	if err != nil {
		log.Fatalf("failed to load aws config for storage: %+v", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	return &Store{
		client:   client,
		bucket:   cfg.Bucket,
		endpoint: cfg.Endpoint,
	}
}

// Put uploads body to key, setting contentType and size (Content-Length) on the object.
func (s *Store) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to put object %q: %w", key, err)
	}
	return nil
}

// Delete removes key from the bucket.
func (s *Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object %q: %w", key, err)
	}
	return nil
}

// Exists reports whether key is present in the bucket.
func (s *Store) Exists(ctx context.Context, key string) (bool, error) {
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
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./infrastructure/storage/... -v`
Expected: PASS (all 7 tests).

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum infrastructure/storage/storage.go infrastructure/storage/storage_test.go
git commit -m "$(cat <<'EOF'
Add infrastructure/storage package for S3-compatible object storage

Wraps aws-sdk-go-v2 for a Ceph RadosGW endpoint (path-style addressing,
forced HTTP/1.1 to work around a RadosGW HTTP/2 PROTOCOL_ERROR).

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: Wire S3 config into `views.Config` and `main.go`

This task is additive only — `FileDir` stays in place and in use. It is removed in Task 5, once
nothing references it any more.

**Files:**
- Modify: `views/views.go`
- Modify: `main.go`

**Interfaces:**
- Consumes: `storage.Config` and `storage.NewStore(cfg Config) *Store` from Task 1.
- Produces: `views.Config.S3 storage.Config` field; `Views.storage *storage.Store` field, populated in `New()`.

- [ ] **Step 1: Add the storage import and `S3` config field to `views/views.go`**

In the import block, add `"github.com/COMTOP1/AFC-GO/infrastructure/storage"` after `"github.com/COMTOP1/AFC-GO/infrastructure/mail"`:

```go
	"github.com/COMTOP1/AFC-GO/affiliation"
	"github.com/COMTOP1/AFC-GO/document"
	"github.com/COMTOP1/AFC-GO/image"
	"github.com/COMTOP1/AFC-GO/infrastructure/db"
	"github.com/COMTOP1/AFC-GO/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/infrastructure/storage"
	"github.com/COMTOP1/AFC-GO/news"
```

In the `Config` struct, add `S3` after `FileDir`:

```go
	Config struct {
		Address           string
		DatabaseURL       string
		DomainName        string
		SessionCookieName string
		FileDir           string
		S3                storage.Config
		Mail              SMTPConfig
		Security          SecurityConfig
	}
```

- [ ] **Step 2: Add the `storage` field to `Views` and initialise it in `New()`**

In the `Views` struct, add `storage` after `sponsor`:

```go
	Views struct {
		affiliation *affiliation.Store
		cache       *cache.Cache
		conf        *Config
		cookie      *sessions.CookieStore
		document    *document.Store
		image       *image.Store
		mailer      *mail.MailerInit
		news        *news.Store
		player      *player.Store
		programme   *programme.Store
		setting     *setting.Store
		sponsor     *sponsor.Store
		storage     *storage.Store
		team        *team.Store
		template    *templates.Templater
		user        *user.Store
		whatsOn     *whatson.Store

		// Visitor tracking
		count         int
		countMutex    sync.Mutex
		flushInterval time.Duration
		stopChan      chan struct{}
	}
```

In `New()`, add the storage init after `v.sponsor = sponsor.NewSponsorRepo(dbStore)`:

```go
	v.sponsor = sponsor.NewSponsorRepo(dbStore)
	v.storage = storage.NewStore(conf.S3)
	v.team = team.NewTeamRepo(dbStore)
```

- [ ] **Step 3: Load S3 env vars and pass them through in `main.go`**

Add the import:

```go
	"github.com/COMTOP1/AFC-GO/infrastructure/storage"
	"github.com/COMTOP1/AFC-GO/views"
```

Before the `conf := &views.Config{` line, add:

```go
	s3Region := os.Getenv("S3_REGION")
	if s3Region == "" {
		s3Region = "us-east-1"
	}
```

In the `conf := &views.Config{...}` literal, add `S3` after `FileDir`:

```go
	conf := &views.Config{
		Address:           address,
		DatabaseURL:       dbConnectionString,
		DomainName:        domainName,
		SessionCookieName: sessionCookieName,
		FileDir:           fileDir,
		S3: storage.Config{
			Endpoint:  os.Getenv("S3_ENDPOINT"),
			Region:    s3Region,
			Bucket:    os.Getenv("S3_BUCKET"),
			AccessKey: os.Getenv("S3_ACCESS_KEY"),
			SecretKey: os.Getenv("S3_SECRET_KEY"),
		},
		Mail: views.SMTPConfig{
```

- [ ] **Step 4: Verify it builds**

Run: `go build ./...`
Expected: succeeds with no errors.

- [ ] **Step 5: Commit**

```bash
git add views/views.go main.go
git commit -m "$(cat <<'EOF'
Wire S3 config and storage.Store into views and main

Additive: FileDir and local storage stay in place until every call
site has moved to storage.Store in the following tasks.

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: Migrate `fileUpload` and every upload/delete call site to `storage.Store`

`fileUpload`'s signature is changing (`ctx` and `category` are new required parameters), and every
call site lives in the same `views` package, so this has to land as one atomic change — the
package won't compile with `helpers.go` rewritten but a single call site left on the old
signature. That's why this is one task with many steps rather than one task per file.

**Files:**
- Modify: `views/helpers.go`
- Modify: `views/team.go`, `views/user.go`, `views/account.go`, `views/document.go`,
  `views/affilitaion.go`, `views/whatson.go`, `views/programme.go`, `views/player.go`,
  `views/gallery.go`, `views/sponsor.go`, `views/news.go`

**Interfaces:**
- Consumes: `Views.storage *storage.Store` (Task 2); `(*storage.Store) Put/Delete` (Task 1).
- Produces: `(v *Views) fileUpload(ctx context.Context, file *multipart.FileHeader, category string) (string, error)` — consumed by every call site in this task.

- [ ] **Step 1: Rewrite `fileUpload` in `views/helpers.go`**

Replace the import block:

```go
import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	// importing time zones in case the system doesn't have them
	_ "time/tzdata"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/document"
	"github.com/COMTOP1/AFC-GO/news"
	"github.com/COMTOP1/AFC-GO/player"
	"github.com/COMTOP1/AFC-GO/programme"
	"github.com/COMTOP1/AFC-GO/role"
	"github.com/COMTOP1/AFC-GO/sponsor"
	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/user"
	"github.com/COMTOP1/AFC-GO/whatson"
)
```

with:

```go
import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"regexp"
	"strings"
	"time"

	// importing time zones in case the system doesn't have them
	_ "time/tzdata"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/document"
	"github.com/COMTOP1/AFC-GO/news"
	"github.com/COMTOP1/AFC-GO/player"
	"github.com/COMTOP1/AFC-GO/programme"
	"github.com/COMTOP1/AFC-GO/role"
	"github.com/COMTOP1/AFC-GO/sponsor"
	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/user"
	"github.com/COMTOP1/AFC-GO/whatson"
)
```

Replace the `fileUpload` function:

```go
func (v *Views) fileUpload(file *multipart.FileHeader) (string, error) {
	var fileName, fileType string
	switch file.Header.Get("content-type") {
	case "application/pdf":
		fileType = ".pdf"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		fileType = ".docx"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		fileType = ".pptx"
	case "text/plain":
		fileType = ".txt"
	case "image/apng":
		fileType = ".apng"
	case "image/avif":
		fileType = ".avif"
	case "image/gif":
		fileType = ".gif"
	case "image/jpeg":
		fileType = ".jpg"
	case "image/png":
		fileType = ".png"
	case "image/svg+xml":
		fileType = ".svg"
	case "image/webp":
		fileType = ".webp"
	default:
		return "", fmt.Errorf("invalid file type: %s", file.Header.Get("content-type"))
	}

	fileName = uuid.NewString() + fileType

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file for fileUpload: %w", err)
	}
	defer src.Close()

	// Destination
	dst, err := os.Create(filepath.Join(v.conf.FileDir, fileName))
	if err != nil {
		return "", fmt.Errorf("failed to create file for fileUpload: %w", err)
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy contents to file for fileUpload: %w", err)
	}

	return fileName, nil
}
```

with:

```go
func (v *Views) fileUpload(ctx context.Context, file *multipart.FileHeader, category string) (string, error) {
	var fileType string
	contentType := file.Header.Get("content-type")
	switch contentType {
	case "application/pdf":
		fileType = ".pdf"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		fileType = ".docx"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		fileType = ".pptx"
	case "text/plain":
		fileType = ".txt"
	case "image/apng":
		fileType = ".apng"
	case "image/avif":
		fileType = ".avif"
	case "image/gif":
		fileType = ".gif"
	case "image/jpeg":
		fileType = ".jpg"
	case "image/png":
		fileType = ".png"
	case "image/svg+xml":
		fileType = ".svg"
	case "image/webp":
		fileType = ".webp"
	default:
		return "", fmt.Errorf("invalid file type: %s", contentType)
	}

	key := category + "/" + uuid.NewString() + fileType

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file for fileUpload: %w", err)
	}
	defer src.Close()

	if err = v.storage.Put(ctx, key, src, file.Size, contentType); err != nil {
		return "", fmt.Errorf("failed to upload file for fileUpload: %w", err)
	}

	return key, nil
}
```

- [ ] **Step 2: Update `views/team.go` (category `"team"`)**

Replace the import block:

```go
import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)
```

with:

```go
import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)
```

In `TeamAddFunc`, replace:

```go
		if hasUpload {
			fileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for team add: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for team add: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}
```

with:

```go
		if hasUpload {
			fileName, err = v.fileUpload(c.Request().Context(), file, "team")
			if err != nil {
				log.Printf("failed to upload file for team add: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for team add: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}
```

In `TeamEditFunc`, replace:

```go
		if hasUpload {
			var fileName string
			fileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for team edit, team id: %d, error: %+v", teamID, err)
				data.Error = fmt.Sprintf("failed to upload file for team edit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if teamDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, teamDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete old image for team edit, team id: %d, error: %+v", teamID, err)
				}
			}
			teamDB.FileName = null.StringFrom(fileName)
		}

		tempRemoveTeamImage := c.FormValue("removeTeamImage")
		if tempRemoveTeamImage == "Y" {
			if teamDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, teamDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete image for team edit, team id: %d, error: %+v", teamID, err)
				}
			}
			teamDB.FileName = null.NewString("", false)
```

with:

```go
		if hasUpload {
			var fileName string
			fileName, err = v.fileUpload(c.Request().Context(), file, "team")
			if err != nil {
				log.Printf("failed to upload file for team edit, team id: %d, error: %+v", teamID, err)
				data.Error = fmt.Sprintf("failed to upload file for team edit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if teamDB.FileName.Valid {
				err = v.storage.Delete(c.Request().Context(), teamDB.FileName.String)
				if err != nil {
					log.Printf("failed to delete old image for team edit, team id: %d, error: %+v", teamID, err)
				}
			}
			teamDB.FileName = null.StringFrom(fileName)
		}

		tempRemoveTeamImage := c.FormValue("removeTeamImage")
		if tempRemoveTeamImage == "Y" {
			if teamDB.FileName.Valid {
				err = v.storage.Delete(c.Request().Context(), teamDB.FileName.String)
				if err != nil {
					log.Printf("failed to delete image for team edit, team id: %d, error: %+v", teamID, err)
				}
			}
			teamDB.FileName = null.NewString("", false)
```

In `TeamDeleteFunc`, replace:

```go
		if teamDB.FileName.Valid {
			err = os.Remove(filepath.Join(v.conf.FileDir, teamDB.FileName.String))
			if err != nil {
				log.Printf("failed to delete team file for team delete, team id: %d, error: %+v", id, err)
			}
		}
```

with:

```go
		if teamDB.FileName.Valid {
			err = v.storage.Delete(c.Request().Context(), teamDB.FileName.String)
			if err != nil {
				log.Printf("failed to delete team file for team delete, team id: %d, error: %+v", id, err)
			}
		}
```

- [ ] **Step 3: Update `views/user.go` (category `"user"`)**

Replace the import block:

```go
import (
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/role"
	"github.com/COMTOP1/AFC-GO/setting"
	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
	"github.com/COMTOP1/AFC-GO/utils"
)
```

with:

```go
import (
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/role"
	"github.com/COMTOP1/AFC-GO/setting"
	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
	"github.com/COMTOP1/AFC-GO/utils"
)
```

In `UserAddFunc`, replace:

```go
		if hasUpload {
			fileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for user add, error: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for user add: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}
```

with:

```go
		if hasUpload {
			fileName, err = v.fileUpload(c.Request().Context(), file, "user")
			if err != nil {
				log.Printf("failed to upload file for user add, error: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for user add: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}
```

In `UserEditFunc`, replace:

```go
		if hasUpload {
			var tempFileName string
			tempFileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for user edit, user id: %d, error: %+v", userID, err)
				data.Error = fmt.Sprintf("failed to upload file for user edit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if userDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, userDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete old image for user edit, user id: %d, error: %+v", userID, err)
				}
			}
			userDB.FileName = null.NewString(tempFileName, len(tempFileName) > 0)
		}

		tempRemoveUserImage := c.FormValue("removeUserImage")
		if tempRemoveUserImage == "Y" {
			if userDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, userDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete image for user edit, user id: %d, error: %+v", userID, err)
				}
			}
```

with:

```go
		if hasUpload {
			var tempFileName string
			tempFileName, err = v.fileUpload(c.Request().Context(), file, "user")
			if err != nil {
				log.Printf("failed to upload file for user edit, user id: %d, error: %+v", userID, err)
				data.Error = fmt.Sprintf("failed to upload file for user edit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if userDB.FileName.Valid {
				err = v.storage.Delete(c.Request().Context(), userDB.FileName.String)
				if err != nil {
					log.Printf("failed to delete old image for user edit, user id: %d, error: %+v", userID, err)
				}
			}
			userDB.FileName = null.NewString(tempFileName, len(tempFileName) > 0)
		}

		tempRemoveUserImage := c.FormValue("removeUserImage")
		if tempRemoveUserImage == "Y" {
			if userDB.FileName.Valid {
				err = v.storage.Delete(c.Request().Context(), userDB.FileName.String)
				if err != nil {
					log.Printf("failed to delete image for user edit, user id: %d, error: %+v", userID, err)
				}
			}
```

In `UserDeleteFunc`, replace:

```go
		if userDB.FileName.Valid {
			err = os.Remove(filepath.Join(v.conf.FileDir, userDB.FileName.String))
			if err != nil {
				log.Printf("failed to delete user image for user delete, user id: %d, error: %+v", id, err)
			}
		}
```

with:

```go
		if userDB.FileName.Valid {
			err = v.storage.Delete(c.Request().Context(), userDB.FileName.String)
			if err != nil {
				log.Printf("failed to delete user image for user delete, user id: %d, error: %+v", id, err)
			}
		}
```

- [ ] **Step 4: Update `views/account.go` (category `"user"`)**

Replace the import block:

```go
import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/templates"
)
```

with:

```go
import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/templates"
)
```

In `UploadImageFunc`, replace:

```go
		if c1.User.FileName.Valid {
			err := os.Remove(filepath.Join(v.conf.FileDir, c1.User.FileName.String))
			if err != nil {
				log.Printf("failed to delete image for uploadImage, user id: %d, error: %+v", c1.User.ID, err)
			}
		}

		file, err := c.FormFile("upload")
		if err != nil {
			log.Printf("failed to get file for uploadImage, user id: %d, error: %+v", c1.User.ID, err)
			data.Error = fmt.Sprintf("failed to get file for uploadImage: %+v", err)
			return c.JSON(http.StatusOK, data)
		}
		var fileName string
		fileName, err = v.fileUpload(file)
```

with:

```go
		if c1.User.FileName.Valid {
			err := v.storage.Delete(c.Request().Context(), c1.User.FileName.String)
			if err != nil {
				log.Printf("failed to delete image for uploadImage, user id: %d, error: %+v", c1.User.ID, err)
			}
		}

		file, err := c.FormFile("upload")
		if err != nil {
			log.Printf("failed to get file for uploadImage, user id: %d, error: %+v", c1.User.ID, err)
			data.Error = fmt.Sprintf("failed to get file for uploadImage: %+v", err)
			return c.JSON(http.StatusOK, data)
		}
		var fileName string
		fileName, err = v.fileUpload(c.Request().Context(), file, "user")
```

In `RemoveImageFunc`, replace:

```go
		if c1.User.FileName.Valid {
			err := os.Remove(filepath.Join(v.conf.FileDir, c1.User.FileName.String))
			if err != nil {
				log.Printf("failed to delete image for removeImage, user id: %d, error: %+v", c1.User.ID, err)
			}
		}
```

with:

```go
		if c1.User.FileName.Valid {
			err := v.storage.Delete(c.Request().Context(), c1.User.FileName.String)
			if err != nil {
				log.Printf("failed to delete image for removeImage, user id: %d, error: %+v", c1.User.ID, err)
			}
		}
```

- [ ] **Step 5: Update `views/document.go` (category `"document"`)**

Replace the import block:

```go
import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/document"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)
```

with:

```go
import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/document"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)
```

In `DocumentAddFunc`, replace:

```go
		fileName, err := v.fileUpload(file)
```

with:

```go
		fileName, err := v.fileUpload(c.Request().Context(), file, "document")
```

In `DocumentDeleteFunc`, replace:

```go
		err = os.Remove(filepath.Join(v.conf.FileDir, documentDB.FileName))
		if err != nil {
			log.Printf("failed to delete document file for document delete, document id: %d, error: %+v", id, err)
		}
```

with:

```go
		err = v.storage.Delete(c.Request().Context(), documentDB.FileName)
		if err != nil {
			log.Printf("failed to delete document file for document delete, document id: %d, error: %+v", id, err)
		}
```

- [ ] **Step 6: Update `views/affilitaion.go` (category `"affiliation"`)**

Replace the import block:

```go
import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/affiliation"
)
```

with:

```go
import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/affiliation"
)
```

In `AffiliationAddFunc`, replace:

```go
		fileName, err := v.fileUpload(file)
```

with:

```go
		fileName, err := v.fileUpload(c.Request().Context(), file, "affiliation")
```

In `AffiliationDeleteFunc`, replace:

```go
		if affiliationDB.FileName.Valid {
			err = os.Remove(filepath.Join(v.conf.FileDir, affiliationDB.FileName.String))
			if err != nil {
				log.Printf("failed to delete affiliation image for affiliation delete, affiliation id: %d, error: %+v", id, err)
			}
		}
```

with:

```go
		if affiliationDB.FileName.Valid {
			err = v.storage.Delete(c.Request().Context(), affiliationDB.FileName.String)
			if err != nil {
				log.Printf("failed to delete affiliation image for affiliation delete, affiliation id: %d, error: %+v", id, err)
			}
		}
```

- [ ] **Step 7: Update `views/whatson.go` (category `"whatson"`)**

Replace the import block:

```go
import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
	"github.com/COMTOP1/AFC-GO/whatson"
)
```

with:

```go
import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
	"github.com/COMTOP1/AFC-GO/whatson"
)
```

In `WhatsOnAddFunc`, replace:

```go
		if hasUpload {
			fileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for whats on add, error: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for whats on add: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}
```

with:

```go
		if hasUpload {
			fileName, err = v.fileUpload(c.Request().Context(), file, "whatson")
			if err != nil {
				log.Printf("failed to upload file for whats on add, error: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for whats on add: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}
```

In `WhatsOnEditFunc`, replace:

```go
		if hasUpload {
			var tempFileName string
			tempFileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for whats on edit, whats on id: %d, error: %+v", whatsOnID, err)
				data.Error = fmt.Sprintf("failed to upload file for whats on edit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if whatsOnDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, whatsOnDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete old image for whats on edit, whats on id: %d, error: %+v", whatsOnID, err)
				}
			}
			whatsOnDB.FileName = null.NewString(tempFileName, len(tempFileName) > 0)
		}
```

with:

```go
		if hasUpload {
			var tempFileName string
			tempFileName, err = v.fileUpload(c.Request().Context(), file, "whatson")
			if err != nil {
				log.Printf("failed to upload file for whats on edit, whats on id: %d, error: %+v", whatsOnID, err)
				data.Error = fmt.Sprintf("failed to upload file for whats on edit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if whatsOnDB.FileName.Valid {
				err = v.storage.Delete(c.Request().Context(), whatsOnDB.FileName.String)
				if err != nil {
					log.Printf("failed to delete old image for whats on edit, whats on id: %d, error: %+v", whatsOnID, err)
				}
			}
			whatsOnDB.FileName = null.NewString(tempFileName, len(tempFileName) > 0)
		}
```

In `WhatsOnDeleteFunc`, replace:

```go
		if whatsOnDB.FileName.Valid {
			err = os.Remove(filepath.Join(v.conf.FileDir, whatsOnDB.FileName.String))
			if err != nil {
				log.Printf("failed to delete whatsOn image for whats on delete, whats on id: %d, error: %+v", id, err)
			}
		}
```

with:

```go
		if whatsOnDB.FileName.Valid {
			err = v.storage.Delete(c.Request().Context(), whatsOnDB.FileName.String)
			if err != nil {
				log.Printf("failed to delete whatsOn image for whats on delete, whats on id: %d, error: %+v", id, err)
			}
		}
```

(`WhatsOnEditFunc`'s `removeWhatsOnImage == "Y"` branch never called `os.Remove` even before this
change — it only clears `FileName`. That's pre-existing behavior, out of scope here; leave it as
is.)

- [ ] **Step 8: Update `views/programme.go` (category `"programme"`)**

Replace the import block:

```go
import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/programme"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)
```

with:

```go
import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/programme"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)
```

In `ProgrammeAddFunc`, replace:

```go
		fileName, err := v.fileUpload(file)
```

with:

```go
		fileName, err := v.fileUpload(c.Request().Context(), file, "programme")
```

In `ProgrammeDeleteFunc`, replace:

```go
		err = os.Remove(filepath.Join(v.conf.FileDir, programmeDB.FileName))
		if err != nil {
			log.Printf("failed to delete programme image for programme delete, programme id: %d, error: %+v", id, err)
		}
```

with:

```go
		err = v.storage.Delete(c.Request().Context(), programmeDB.FileName)
		if err != nil {
			log.Printf("failed to delete programme image for programme delete, programme id: %d, error: %+v", id, err)
		}
```

- [ ] **Step 9: Update `views/player.go` (category `"player"`)**

Replace the import block:

```go
import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/player"
	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)
```

with:

```go
import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/player"
	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)
```

In `PlayerAddFunc`, replace:

```go
		if hasUpload {
			fileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for player add, error: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for player add: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}
```

with:

```go
		if hasUpload {
			fileName, err = v.fileUpload(c.Request().Context(), file, "player")
			if err != nil {
				log.Printf("failed to upload file for player add, error: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for player add: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}
```

In `PlayerEditFunc`, replace:

```go
		if hasUpload {
			var tempFileName string
			tempFileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for player edit, player id: %d, error: %+v", playerID, err)
				data.Error = fmt.Sprintf("failed to upload file for player edit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if playerDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, playerDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete old image for player edit, player id: %d, error: %+v", playerID, err)
				}
			}
			playerDB.FileName = null.NewString(tempFileName, len(tempFileName) > 0)
		}

		tempRemovePlayerImage := c.FormValue("removePlayerImage")
		if tempRemovePlayerImage == "Y" {
			if playerDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, playerDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete image for player edit, player id: %d, error: %+v", playerID, err)
				}
			}
```

with:

```go
		if hasUpload {
			var tempFileName string
			tempFileName, err = v.fileUpload(c.Request().Context(), file, "player")
			if err != nil {
				log.Printf("failed to upload file for player edit, player id: %d, error: %+v", playerID, err)
				data.Error = fmt.Sprintf("failed to upload file for player edit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if playerDB.FileName.Valid {
				err = v.storage.Delete(c.Request().Context(), playerDB.FileName.String)
				if err != nil {
					log.Printf("failed to delete old image for player edit, player id: %d, error: %+v", playerID, err)
				}
			}
			playerDB.FileName = null.NewString(tempFileName, len(tempFileName) > 0)
		}

		tempRemovePlayerImage := c.FormValue("removePlayerImage")
		if tempRemovePlayerImage == "Y" {
			if playerDB.FileName.Valid {
				err = v.storage.Delete(c.Request().Context(), playerDB.FileName.String)
				if err != nil {
					log.Printf("failed to delete image for player edit, player id: %d, error: %+v", playerID, err)
				}
			}
```

In `PlayerDeleteFunc`, replace:

```go
		if playerDB.FileName.Valid {
			err = os.Remove(filepath.Join(v.conf.FileDir, playerDB.FileName.String))
			if err != nil {
				log.Printf("failed to delete player image for player delete, player id: %d, error: %+v", id, err)
			}
		}
```

with:

```go
		if playerDB.FileName.Valid {
			err = v.storage.Delete(c.Request().Context(), playerDB.FileName.String)
			if err != nil {
				log.Printf("failed to delete player image for player delete, player id: %d, error: %+v", id, err)
			}
		}
```

- [ ] **Step 10: Update `views/gallery.go` (category `"gallery"`)**

Replace the import block:

```go
import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/image"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)
```

with:

```go
import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/image"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)
```

In `ImageAddFunc`, replace:

```go
		fileName, err := v.fileUpload(file)
```

with:

```go
		fileName, err := v.fileUpload(c.Request().Context(), file, "gallery")
```

In `ImageDeleteFunc`, replace:

```go
		err = os.Remove(filepath.Join(v.conf.FileDir, imageDB.FileName))
		if err != nil {
			log.Printf("failed to delete image file for image delete, image id: %d, error: %+v", id, err)
		}
```

with:

```go
		err = v.storage.Delete(c.Request().Context(), imageDB.FileName)
		if err != nil {
			log.Printf("failed to delete image file for image delete, image id: %d, error: %+v", id, err)
		}
```

- [ ] **Step 11: Update `views/sponsor.go` (category `"sponsor"`)**

Replace the import block:

```go
import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/sponsor"
	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)
```

with:

```go
import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/sponsor"
	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)
```

In `SponsorAddFunc`, replace:

```go
		fileName, err := v.fileUpload(file)
```

with:

```go
		fileName, err := v.fileUpload(c.Request().Context(), file, "sponsor")
```

In `SponsorDeleteFunc`, replace:

```go
		if sponsorDB.FileName.Valid {
			err = os.Remove(filepath.Join(v.conf.FileDir, sponsorDB.FileName.String))
			if err != nil {
				log.Printf("failed to delete sponsor image for sponsor delete, sponsor id: %d, error: %+v", id, err)
			}
		}
```

with:

```go
		if sponsorDB.FileName.Valid {
			err = v.storage.Delete(c.Request().Context(), sponsorDB.FileName.String)
			if err != nil {
				log.Printf("failed to delete sponsor image for sponsor delete, sponsor id: %d, error: %+v", id, err)
			}
		}
```

- [ ] **Step 12: Update `views/news.go` (category `"news"`)**

Replace the import block:

```go
import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/news"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)
```

with:

```go
import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/news"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)
```

In `NewsAddFunc`, replace:

```go
		if hasUpload {
			fileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for news add, error: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for news add: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}
```

with:

```go
		if hasUpload {
			fileName, err = v.fileUpload(c.Request().Context(), file, "news")
			if err != nil {
				log.Printf("failed to upload file for news add, error: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for news add: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}
```

In `NewsEditFunc`, replace:

```go
		if hasUpload {
			var tempFileName string
			tempFileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for news edit, news id: %d, error: %+v", newsID, err)
				data.Error = fmt.Sprintf("failed to upload file for news edit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if newsDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, path.Clean(newsDB.FileName.String)))
				if err != nil {
					log.Printf("failed to delete old image for news edit, news id: %d, error: %+v", newsID, err)
				}
			}
			newsDB.FileName = null.NewString(tempFileName, len(tempFileName) > 0)
		}

		tempRemoveNewsImage := c.FormValue("removeNewsImage")
		if tempRemoveNewsImage == "Y" {
			if newsDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, path.Clean(newsDB.FileName.String)))
				if err != nil {
					log.Printf("failed to delete image for news edit, news id: %d, error: %+v", newsID, err)
				}
```

with:

```go
		if hasUpload {
			var tempFileName string
			tempFileName, err = v.fileUpload(c.Request().Context(), file, "news")
			if err != nil {
				log.Printf("failed to upload file for news edit, news id: %d, error: %+v", newsID, err)
				data.Error = fmt.Sprintf("failed to upload file for news edit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if newsDB.FileName.Valid {
				err = v.storage.Delete(c.Request().Context(), newsDB.FileName.String)
				if err != nil {
					log.Printf("failed to delete old image for news edit, news id: %d, error: %+v", newsID, err)
				}
			}
			newsDB.FileName = null.NewString(tempFileName, len(tempFileName) > 0)
		}

		tempRemoveNewsImage := c.FormValue("removeNewsImage")
		if tempRemoveNewsImage == "Y" {
			if newsDB.FileName.Valid {
				err = v.storage.Delete(c.Request().Context(), newsDB.FileName.String)
				if err != nil {
					log.Printf("failed to delete image for news edit, news id: %d, error: %+v", newsID, err)
				}
```

In `NewsDeleteFunc`, replace:

```go
		if newsDB.FileName.Valid {
			err = os.Remove(filepath.Join(v.conf.FileDir, newsDB.FileName.String))
			if err != nil {
				log.Printf("failed to delete news image for news delete, news id: %d, error: %+v", id, err)
			}
		}
```

with:

```go
		if newsDB.FileName.Valid {
			err = v.storage.Delete(c.Request().Context(), newsDB.FileName.String)
			if err != nil {
				log.Printf("failed to delete news image for news delete, news id: %d, error: %+v", id, err)
			}
		}
```

- [ ] **Step 13: Verify it builds and vets clean**

Run: `go build ./... && go vet ./...`
Expected: both succeed with no errors (no unused imports, `fileUpload` signature matches every call site).

- [ ] **Step 14: Commit**

```bash
git add views/helpers.go views/team.go views/user.go views/account.go views/document.go \
  views/affilitaion.go views/whatson.go views/programme.go views/player.go views/gallery.go \
  views/sponsor.go views/news.go
git commit -m "$(cat <<'EOF'
Move fileUpload and every delete call site to storage.Store

fileUpload now uploads to S3 under "<category>/<uuid>.<ext>" instead of
writing to local disk; every os.Remove(FileDir, ...) call site now
calls storage.Store.Delete with the same DB-stored filename as key.

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: Serve downloads from S3 instead of local disk

**Files:**
- Modify: `views/download.go`

**Interfaces:**
- Consumes: `(*storage.Store) Exists(ctx, key) (bool, error)`, `(*storage.Store) PublicURL(key) string` (Task 1); `Views.storage` (Task 2).

- [ ] **Step 1: Update imports**

Replace:

```go
import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	affiliation1 "github.com/COMTOP1/AFC-GO/affiliation"
	document1 "github.com/COMTOP1/AFC-GO/document"
	image1 "github.com/COMTOP1/AFC-GO/image"
	news1 "github.com/COMTOP1/AFC-GO/news"
	player1 "github.com/COMTOP1/AFC-GO/player"
	programme1 "github.com/COMTOP1/AFC-GO/programme"
	sponsor1 "github.com/COMTOP1/AFC-GO/sponsor"
	team1 "github.com/COMTOP1/AFC-GO/team"
	user1 "github.com/COMTOP1/AFC-GO/user"
	whatson1 "github.com/COMTOP1/AFC-GO/whatson"
)
```

with:

```go
import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	affiliation1 "github.com/COMTOP1/AFC-GO/affiliation"
	document1 "github.com/COMTOP1/AFC-GO/document"
	image1 "github.com/COMTOP1/AFC-GO/image"
	news1 "github.com/COMTOP1/AFC-GO/news"
	player1 "github.com/COMTOP1/AFC-GO/player"
	programme1 "github.com/COMTOP1/AFC-GO/programme"
	sponsor1 "github.com/COMTOP1/AFC-GO/sponsor"
	team1 "github.com/COMTOP1/AFC-GO/team"
	user1 "github.com/COMTOP1/AFC-GO/user"
	whatson1 "github.com/COMTOP1/AFC-GO/whatson"
)
```

- [ ] **Step 2: Rewrite `_downloadFunc`**

Replace:

```go
func (v *Views) _downloadFunc(c echo.Context, fileName, page string, id int) error {
	path := filepath.Join(v.conf.FileDir, fileName)
	_, err := os.Stat(path)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			log.Printf("failed to get file for %s download: no such file, id: %d", page, id)
			return c.String(http.StatusNotFound,
				fmt.Sprintf("failed to get file for %s download: no such file, id: %d", page, id))
		}
		return fmt.Errorf("failed to get file for %s download: %w, id: %d", page, err, id)
	}
	return c.Redirect(http.StatusFound, "/file/"+url.PathEscape(fileName))
}
```

with:

```go
func (v *Views) _downloadFunc(c echo.Context, fileName, page string, id int) error {
	exists, err := v.storage.Exists(c.Request().Context(), fileName)
	if err != nil {
		return fmt.Errorf("failed to check file exists for %s download: %w, id: %d", page, err, id)
	}
	if !exists {
		log.Printf("failed to get file for %s download: no such file, id: %d", page, id)
		return c.String(http.StatusNotFound,
			fmt.Sprintf("failed to get file for %s download: no such file, id: %d", page, id))
	}
	return c.Redirect(http.StatusFound, v.storage.PublicURL(fileName))
}
```

- [ ] **Step 3: Verify it builds and vets clean**

Run: `go build ./... && go vet ./...`
Expected: both succeed with no errors.

- [ ] **Step 4: Commit**

```bash
git add views/download.go
git commit -m "$(cat <<'EOF'
Redirect downloads straight to the S3 CDN URL

Replaces the os.Stat + /file/<name> local-proxy redirect with an S3
existence check and a direct redirect to the object's public CDN URL.

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: Remove local disk storage and finish config

**Files:**
- Modify: `main.go`
- Modify: `views/views.go`
- Modify: `example.env`

**Interfaces:**
- Consumes: nothing new — this only removes the now-dead `FileDir` path, since Tasks 3 and 4
  moved every reader of it onto `storage.Store`.

- [ ] **Step 1: Remove the `FileDir` field from `views.Config`**

In `views/views.go`, replace:

```go
	Config struct {
		Address           string
		DatabaseURL       string
		DomainName        string
		SessionCookieName string
		FileDir           string
		S3                storage.Config
		Mail              SMTPConfig
		Security          SecurityConfig
	}
```

with:

```go
	Config struct {
		Address           string
		DatabaseURL       string
		DomainName        string
		SessionCookieName string
		S3                storage.Config
		Mail              SMTPConfig
		Security          SecurityConfig
	}
```

- [ ] **Step 2: Remove the `/FileStore` resolution logic and `FileDir` field from `main.go`**

Replace:

```go
	var fileDir string

	stat, err := os.Stat("/FileStore")
	if err == nil && stat.IsDir() {
		log.Println("using root /FileStore")
		fileDir = "/FileStore"
	} else {
		stat, err = os.Stat("./FileStore")
		if err == nil && stat.IsDir() {
			log.Println("using local ./FileStore")
			fileDir = "./FileStore"
		} else {
			log.Fatalf("failed to get fileStore - stat: %+v, error: %+v", stat, err)
		}
	}

	mailPort, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
```

with:

```go
	mailPort, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
```

Replace:

```go
	conf := &views.Config{
		Address:           address,
		DatabaseURL:       dbConnectionString,
		DomainName:        domainName,
		SessionCookieName: sessionCookieName,
		FileDir:           fileDir,
		S3: storage.Config{
```

with:

```go
	conf := &views.Config{
		Address:           address,
		DatabaseURL:       dbConnectionString,
		DomainName:        domainName,
		SessionCookieName: sessionCookieName,
		S3: storage.Config{
```

- [ ] **Step 3: Add the S3 env vars to `example.env`**

Replace:

```
ADDRESS=
DB_HOSTNAME=
DB_PORT=
DB_NAME=
DB_USERNAME=
DB_PASSWORD=
JWT_SECRET=
ITERATIONS=
KEY_LENGTH_BYTES=
SESSION_COOKIE_NAME=
ENCRYPTION_KEY=
AUTHENTICATION_KEY=
```

with:

```
ADDRESS=
DB_HOSTNAME=
DB_PORT=
DB_NAME=
DB_USERNAME=
DB_PASSWORD=
JWT_SECRET=
ITERATIONS=
KEY_LENGTH_BYTES=
SESSION_COOKIE_NAME=
ENCRYPTION_KEY=
AUTHENTICATION_KEY=
S3_ENDPOINT=
S3_REGION=
S3_BUCKET=
S3_ACCESS_KEY=
S3_SECRET_KEY=
```

- [ ] **Step 4: Verify the whole module builds, vets, and is gofmt-clean**

Run: `go build ./... && go vet ./... && gofmt -l .`
Expected: `go build`/`go vet` succeed with no errors; `gofmt -l .` prints nothing (no unformatted files).

Run: `grep -rn "FileDir\|conf.FileDir" --include="*.go" .`
Expected: no output — confirms every reference to `FileDir` is gone.

- [ ] **Step 5: Commit**

```bash
git add main.go views/views.go example.env
git commit -m "$(cat <<'EOF'
Remove local disk file storage

FileDir and the /FileStore resolution logic are dead now that uploads,
deletes, and downloads all go through storage.Store. S3 is the only
storage backend from here on.

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 6: One-off migration script for the existing 72 files

**Files:**
- Create: `cmd/migrates3/main.go`

**Interfaces:**
- Consumes: `storage.Config`, `storage.NewStore`, `(*storage.Store) Exists`, `(*storage.Store) Put` (Task 1).

- [ ] **Step 1: Write the migration command**

Create `cmd/migrates3/main.go`:

```go
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
```

- [ ] **Step 2: Verify it builds**

Run: `go build ./...`
Expected: succeeds with no errors, produces a `cmd/migrates3` binary target.

- [ ] **Step 3: Commit**

```bash
git add cmd/migrates3/main.go
git commit -m "$(cat <<'EOF'
Add one-off migration script for existing FileStore files

Uploads every file under ./FileStore to S3 under its existing flat
filename, so current DB rows resolve unchanged. Idempotent: skips
objects that already exist in the bucket. Run manually once S3
credentials are available; does not touch local files or the DB.

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
EOF
)"
```

**Manual verification (not automatable without real credentials):** once `S3_ENDPOINT`,
`S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY` are set in the environment, run
`go run ./cmd/migrates3` against the real `./FileStore` directory and confirm the summary line
reports `failed=0` and `uploaded` + `skipped` equals the file count in `./FileStore`. Spot-check a
few uploaded objects' Content-Type in the bucket (e.g. a `.jpg` should read `image/jpeg`, not
`application/octet-stream`).

---

## Post-plan manual steps (not part of any task above)

- Provision the actual S3 bucket/credentials and set `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`,
  `S3_ACCESS_KEY`, `S3_SECRET_KEY` in the real `.env` / deployment secrets.
- Run `cmd/migrates3` once against production data (see Task 6's manual verification note).
- After confirming the CDN serves migrated files correctly, delete the local `./FileStore` /
  `/FileStore` directory and its volume mount from the Nomad job files (in the separate ops repo —
  out of scope for this repository).
