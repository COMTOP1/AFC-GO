# S3 File Storage Design

## Context

AFC-GO currently stores uploaded files (player/user photos, documents, programmes, sponsor logos,
gallery images, etc.) on local disk under `v.conf.FileDir` (`/FileStore` or `./FileStore`,
resolved in `main.go`). `views/helpers.go`'s `fileUpload` writes new files with `os.Create`;
~10 view files delete old files with `os.Remove(filepath.Join(v.conf.FileDir, ...))`;
`views/download.go`'s `_downloadFunc` checks existence with `os.Stat` and 302-redirects the
browser to `/file/<filename>`, a path served outside this app (reverse proxy over the mounted
volume — there is no `/file` route in `router.go`).

This design replaces local disk storage with an S3-compatible object store (Ceph RadosGW,
validated in `/Users/liam/Code/Go/test/s3`), reachable over a public CDN endpoint.

## Decisions

- **New AFC-specific bucket and credentials** — not shared with other projects (e.g. `daisy-public`).
  Bucket name/endpoint/credentials are supplied later via env vars; not hardcoded.
- **Public bucket, direct CDN URLs** — matches current behavior (files already sit at an
  unauthenticated, UUID-named path with no access check beyond obscurity). No presigned URLs.
- **Uploads stay server-side** — the browser POSTs multipart form data to the existing Go
  endpoints (unchanged), which already validate content-type and enforce a 15MB body limit. The
  Go server then calls `PutObject` directly using static credentials from env vars. Presigned
  upload URLs are unnecessary: they exist to let an untrusted browser upload directly to S3
  without exposing server credentials, but here the credentials never leave the backend — there's
  no security benefit, and switching to direct-from-browser uploads would require a frontend
  rewrite that's out of scope.
- **One-off migration script** for the ~72 existing files, run manually, not part of the request
  path.
- **Local disk storage removed entirely** — no dual-write, no fallback.

## Architecture

New package `infrastructure/storage`, following the existing `infrastructure/db` /
`infrastructure/mail` convention:

```go
package storage

type Config struct {
    Endpoint  string
    Region    string
    Bucket    string
    AccessKey string
    SecretKey string
}

type Store struct {
    // s3 client, bucket, endpoint
}

func NewStore(cfg Config) *Store
func (s *Store) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
func (s *Store) Delete(ctx context.Context, key string) error
func (s *Store) Exists(ctx context.Context, key string) (bool, error)
func (s *Store) PublicURL(key string) string
```

`NewStore` configures the `s3.Client` with a custom `BaseEndpoint`, `UsePathStyle: true`, static
credentials, and an HTTP client that forces HTTP/1.1 — mirroring the proven test script, which
works around a Ceph RadosGW HTTP/2 `PROTOCOL_ERROR`.

`PublicURL(key)` builds `<endpoint>/<bucket>/<key>`, percent-escaping each `/`-delimited path
segment individually (not the whole key, since keys contain a literal `/` separator that must
survive escaping).

## Config wiring

- Add `S3 storage.Config` (or equivalent fields) to `views.Config`.
- `main.go` loads `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY` from
  env, alongside the existing DB/mail/security env vars. Add blank keys to `example.env`.
- Remove `FileDir` from `views.Config` and the `/FileStore` / `./FileStore` `os.Stat` fallback
  logic in `main.go`.
- `views/views.go`: add a `storage *storage.Store` field to `Views`, constructed in `New()`.

## Key naming

`fileUpload` gains a `category` parameter, one per call site, matching `download.go`'s existing
source-letter switch:

| Call site | Category |
|---|---|
| `views/team.go` | `team` |
| `views/user.go`, `views/account.go` | `user` |
| `views/document.go` | `document` |
| `views/affilitaion.go` | `affiliation` |
| `views/whatson.go` | `whatson` |
| `views/programme.go` | `programme` |
| `views/player.go` | `player` |
| `views/gallery.go` | `gallery` |
| `views/sponsor.go` | `sponsor` |
| `views/news.go` | `news` |

Generated key: `"<category>/<uuid>.<ext>"`. This full key is stored verbatim in the DB `FileName`
column, exactly as the flat filename is today — no DB schema change. Every downstream consumer
(delete, download) continues to treat `FileName` as an opaque key, unchanged from current
behavior.

## Call-site changes

- `views/helpers.go`: `fileUpload(ctx context.Context, file *multipart.FileHeader, category string) (string, error)`
  — opens the multipart file, determines extension from content-type (unchanged switch), builds
  the key, calls `v.storage.Put(ctx, key, src, file.Size, contentType)`. Content-Type is passed
  through so the CDN serves correct types (needed for `<img>` tags etc. that hit the object URL
  directly).
- All call sites: `v.fileUpload(file)` → `v.fileUpload(c.Request().Context(), file, "<category>")`.
- All delete sites: `os.Remove(filepath.Join(v.conf.FileDir, x))` → `v.storage.Delete(c.Request().Context(), x)`,
  preserving today's log-only, non-fatal error handling.
- `views/download.go`: `_downloadFunc` replaces the `os.Stat` check with
  `v.storage.Exists(ctx, key)`, and redirects to `v.storage.PublicURL(fileName)` instead of
  `/file/<name>`. All existing business logic (youth-team/under-18 gating, source-type lookups)
  is unchanged — only the final redirect target changes.

## Migration script

Standalone command, `cmd/migrates3/main.go`, reusing the `storage` package:

- Walks `./FileStore`.
- Uploads each file under its existing flat filename as the S3 key (no category prefix) — so
  existing DB rows, which store flat filenames, keep resolving without a DB migration. New and
  legacy keys coexist in the same bucket.
- Infers Content-Type from file extension via `mime.TypeByExtension`.
- Skips objects that already exist in the bucket (idempotent, safe to re-run/resume).
- Logs a summary (uploaded / skipped / failed counts).
- Does not delete local files. Removing `./FileStore` after verifying the migration is a manual
  step.

## Error handling

Unchanged from today's behavior:

- Upload failures are returned to the caller and surfaced as a request error (same as current
  `os.Create`/`io.Copy` failure handling).
- Delete failures are logged but non-fatal (best-effort cleanup, matching every existing
  `os.Remove` call site).
- Download existence checks return 404 the same way a missing local file does today.

## Out of scope

- Presigned/private URLs (public bucket chosen; see Decisions).
- Automatically deleting the 72 local files after migration.
- Nomad job files / volume-mount changes — these live in a separate ops repo, not in this
  repository.
