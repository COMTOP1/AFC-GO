# Server Restructure & JSON API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move AFC-GO into the MV-Controller layout (`server/cmd`, `server/internal/<domain>`), and add a JSON API under `/api/v1` backed by per-domain services that the existing template site also uses.

**Architecture:** Each domain package (`news`, `team`, …) holds its store (the existing SQL), a `Service` (validation, uploads, sanitising, side effects) and Echo `Handlers`. The legacy template views in `server/internal/legacy` call the same services, so the site stays deployable after every task. Shared plumbing is split across these packages:
- `web`: error envelope, JSON/form helpers, CSRF, the API group.
- `auth`: cookie sessions, guards, reset tokens, login/password.
- `upload`: S3 uploads.
- `svcerr`: typed service errors.
- `app`: wiring.

**Tech Stack:** Go 1.26, Echo v4.15, sqlx + lib/pq, gorilla/sessions, OTel (otelecho), swaggo (`swag` + `echo-swagger`), testify.

**Spec:** `docs/superpowers/specs/2026-09-27-server-api-split-design.md`. Read it before starting any task.

**Working directory:** `/Users/liam/Code/Go/AFC-server-api-split` (git worktree, branch `server-api-split`). Every command below runs from there.

## Global Constraints

- Module path stays `github.com/COMTOP1/AFC-GO`; `go.mod` stays at the repo root; Go `1.26.7`.
- Keep Echo (`github.com/labstack/echo/v4`) and `otelecho`. Do not switch to chi.
- API base is `/api/v1`. Every error under `/api/` is `{"error":{"code":<int>,"message":<string>,"fields":{…}}}`; `fields` only on 422.
- Status codes: validation `422`, not found `404`, not logged in `401`, wrong role `403`, conflict `409`, anything unexpected `500` with message `internal server error` (never leak internal errors).
- Session cookie: same name (`WAUTH_SESSION_COOKIE_NAME`, default `session`) and same keys as today, value key `"user"` holding `user.User` (gob). Add `SameSite=Lax`, and `Secure` when `DOMAIN_NAME` is not `localhost`.
- CSRF on `/api/v1`: header `X-CSRF-Token`, cookie `_csrf` (`SameSite=Strict`, not HttpOnly); every CSRF failure → `403`.
- Dates in API JSON are RFC 3339 (`time.Time`). Files are exposed as absolute URLs (`imageUrl`, `fileUrl`), omitted when absent. `user.User` is never serialised directly.
- Player photos are never exposed for a player on a youth team or aged under 18.
- API checkbox fields are real booleans; legacy form adapters map `"Y"` → `true`, anything else → `false`.
- No DB schema changes. The only store SQL changes allowed are the ones this plan specifies (`RETURNING id` on inserts, write-through edit wrappers).
- Real database dumps (`postgres_*.sql`) are never read by tooling or committed. Test fixtures are synthetic.
- After every task: `go build ./...`, `go vet ./...`, `go test ./...` and `go tool golangci-lint run ./...` pass, and the legacy site still serves every page.
- Commit messages end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Plan Decisions (beyond the spec)

These are implementation choices this plan makes. Flag them in review if they look wrong.

1. **Handler tests use real services on fake stores** rather than fake services. The spec says "fake services", but this exercises more real code for the same setup cost, and still needs no Postgres.
2. **PATCH is a true partial update.** Multipart and JSON fields that are *absent* are left unchanged, while fields sent *empty* clear optional values. Legacy form adapters always send every field.
3. **`title`/`name` are required** on every create/update (422 when blank). Legacy forms already mark these `required` in HTML.
4. **Store fixes:**
   - Inserts use `RETURNING id`, so the API can return the created resource's ID.
   - `news.EditNews` and `player.EditPlayer` become write-through, so removing an image actually clears `file_name`. Today the S3 object is deleted, but the row keeps the dead key.
   - `user.EditUser` can clear `phone`.
5. **Old S3 objects are deleted after the DB write succeeds**, not before (legacy deleted first, which could orphan the row on a DB error).
6. **API login `remember`:** `true` → 31-day cookie, `false` → the default 1-day cookie. Legacy's inverted check (`!= "on"` → 31 days) is left untouched in the legacy view.
7. **Legacy read pages keep reading the stores directly.** Only legacy *writes* go through services, because the read templates depend on legacy formats (`null.String` file names, preformatted dates) and are deleted in sub-project 5. The exception is the player-photo rule: legacy templates call the shared `player.PhotoVisible` (Task 16).
8. **Import cycles:** `player`, `sponsor` and `user` import `team`. So team deletion unlinks through a `team.Detacher` interface (one `UPDATE` per store), and the team detail page `GET /teams/{id}` is served by the `site` package (Task 18).
9. **Security fix (Task 21, not in the spec):** forced-reset login currently issues a reset link without checking the password (an account takeover for new and admin-reset users). Fixed in `user.Store.VerifyUser`, which also fixes the legacy site. Recommended as a separate hotfix to `main` as well.
10. **DB-backed tests** use `AFC_TEST_DB` (a Postgres URL). Each test gets its own schema via `search_path`, so packages can run in parallel. Tests skip when it is unset. Local setup:
   ```bash
   docker run -d --name afc-test-db -e POSTGRES_PASSWORD=test -p 55432:5432 postgres:17-alpine
   export AFC_TEST_DB='postgres://postgres:test@localhost:55432/postgres?sslmode=disable'
   ```

## Review Focus

The five failure modes most likely to bite a real user that the spec implies but does not spell out. Each has a pinned test in the task that owns the code.

1. **Child safety: player photos leak through a new path.** Youth-team or under-18 players must have no `imageUrl` in `/players` and `/teams/{id}`, and `/files/player/{id}` must 404 for them. Pinned in Task 16 (`TestPhotoVisible`, `TestListHidesMinorPhotos`), Task 18 (`TestTeamDetailHidesMinorPhotos`) and Task 19 (`TestPlayerFileHiddenForMinor`).
2. **Existing logged-in users get logged out, or aren't recognised by the API.** A cookie written the way legacy `LoginFunc` writes it must authenticate `/api/v1/auth/me`. Pinned in Task 5 (`TestLegacySessionCookieIsRecognised`).
3. **PATCH blanks fields the client didn't send, or "remove image" leaves a dead key.** Omitted fields must be untouched. `removeImage=true` must clear `file_name` in the DB and delete the S3 object only after the DB write. Pinned in Task 8 (`TestUpdateOmittedFieldsUnchanged`, `TestUpdateRemoveImage`, `TestEditNewsClearsFileName`).
4. **Dangerous uploads are accepted.** A `text/html` or executable content type must give 422 `fields.file` with nothing written to S3. Pinned in Task 3 (`TestSaveRejectsUnsupportedType`) and Task 8 (`TestCreateRejectsHTMLUpload`).
5. **An API client gets an HTML page instead of JSON.** Unknown `/api/v1/*` paths, wrong methods and CSRF failures must all return the JSON envelope, not the legacy 404/error template. Pinned in Task 4 (`TestUnknownAPIPathIsJSON404`, `TestWrongMethodIsJSON405`, `TestCrossSitePostRejected`).

## File Structure (end state of this plan)

```
server/
├── cmd/
│   ├── afc/main.go                     env → app.Config → app.New → Start (swag general info lives here)
│   ├── migrates3/main.go               moved from cmd/migrates3
│   └── offline/                        moved from offline/ (unchanged standalone binary)
└── internal/
    ├── app/app.go, app_test.go          builds Echo: middleware, legacy routes, API handlers
    ├── app/smoke_test.go                legacy page smoke test (DB-backed)
    ├── web/{errors,json,csrf,api,guards,health}.go (+ tests)
    ├── web/apitest/apitest.go           test client helpers
    ├── auth/{sessions,guards,tokens,service,handlers,types}.go (+ tests)
    ├── auth/authtest/authtest.go        test sessions + cookie helper
    ├── svcerr/svcerr.go (+ test)
    ├── upload/upload.go (+ test), upload/uploadtest/fake.go
    ├── sanitize/sanitize.go (+ test)
    ├── emails/{emails.go,resetEmail.tmpl,signupEmail.tmpl,mjml/}
    ├── visitors/visitors.go (+ test)
    ├── testdb/{testdb.go,schema.sql,seed.sql}
    ├── affiliation/ document/ image/ news/ player/ programme/ setting/ sponsor/ team/ user/ whatson/
    │     <domain>.go (model + store wrappers), store.go (SQL; was db.go),
    │     service.go, types.go, handlers.go, service_test.go, handlers_test.go
    ├── site/{service,handlers,types}.go (+ test)      /site, /home, /contact
    ├── files/{service,handlers}.go (+ test)           /files/{kind}/{id}
    ├── account/handlers.go (+ test)                   /account, /account/image
    ├── role/ utils/ infrastructure/{db,mail,storage,telemetry}/
    ├── docs/{generate.go, docs.go, swagger.json, swagger.yaml}
    └── legacy/
        ├── assets.go (embeds public/), routes.go (old router.go loadRoutes)
        ├── public/  templates/  views/
```

---
## Phase A: Safety net and move

### Task 1: Test database fixture + legacy smoke test

Written **before** anything moves, so every later task can prove the legacy site still works.

**Files:**
- Create: `testdb/testdb.go`, `testdb/schema.sql`, `testdb/seed.sql`
- Create: `router_smoke_test.go` (package `main`, repo root)
- Modify: `go.mod`, `go.sum` (add testify)

**Interfaces:**
- Produces: `testdb.Open(t *testing.T) (*sqlx.DB, string)`. It returns a connection plus a DSN whose `search_path` points at a fresh schema loaded with `schema.sql` + `seed.sql`, and skips the test when `AFC_TEST_DB` is unset.

- [ ] **Step 1: Add testify**

```bash
go get github.com/stretchr/testify@latest
```

- [ ] **Step 2: Write the schema fixture**

This is derived from the columns the stores read and write. No real dump is involved.

`testdb/schema.sql`:
```sql
CREATE TABLE teams (
    id           SERIAL PRIMARY KEY,
    name         TEXT    NOT NULL,
    description  TEXT,
    league       TEXT,
    division     TEXT,
    league_table TEXT,
    fixtures     TEXT,
    coach        TEXT,
    physio       TEXT,
    file_name    TEXT,
    active       BOOLEAN NOT NULL DEFAULT FALSE,
    youth        BOOLEAN NOT NULL DEFAULT FALSE,
    ages         INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE users (
    id             SERIAL PRIMARY KEY,
    name           TEXT    NOT NULL,
    email          TEXT    NOT NULL UNIQUE,
    phone          TEXT,
    team_id        INTEGER NOT NULL DEFAULT 0,
    role           TEXT    NOT NULL,
    file_name      TEXT,
    reset_password BOOLEAN NOT NULL DEFAULT FALSE,
    password       TEXT,
    hash           TEXT,
    salt           TEXT
);

CREATE TABLE players (
    id            SERIAL PRIMARY KEY,
    name          TEXT    NOT NULL,
    file_name     TEXT,
    date_of_birth DATE,
    position      TEXT,
    captain       BOOLEAN NOT NULL DEFAULT FALSE,
    team_id       INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE sponsors (
    id        SERIAL PRIMARY KEY,
    name      TEXT NOT NULL,
    website   TEXT,
    file_name TEXT,
    purpose   TEXT,
    team_id   TEXT NOT NULL DEFAULT ''
);

CREATE TABLE affiliations (
    id        SERIAL PRIMARY KEY,
    name      TEXT NOT NULL,
    website   TEXT,
    file_name TEXT
);

CREATE TABLE documents (
    id        SERIAL PRIMARY KEY,
    name      TEXT NOT NULL,
    file_name TEXT NOT NULL
);

CREATE TABLE images (
    id        SERIAL PRIMARY KEY,
    file_name TEXT NOT NULL,
    caption   TEXT
);

CREATE TABLE news (
    id        SERIAL PRIMARY KEY,
    title     TEXT        NOT NULL,
    file_name TEXT,
    content   TEXT,
    date      TIMESTAMPTZ NOT NULL
);

CREATE TABLE whatson (
    id            SERIAL PRIMARY KEY,
    title         TEXT        NOT NULL,
    file_name     TEXT,
    content       TEXT,
    date          TIMESTAMPTZ NOT NULL,
    date_of_event DATE        NOT NULL
);

CREATE TABLE programme_seasons (
    id     SERIAL PRIMARY KEY,
    season TEXT NOT NULL
);

CREATE TABLE programmes (
    id                  SERIAL PRIMARY KEY,
    name                TEXT    NOT NULL,
    file_name           TEXT    NOT NULL,
    date_of_programme   DATE    NOT NULL,
    programme_season_id INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE settings (
    id           TEXT PRIMARY KEY,
    setting_text TEXT NOT NULL
);
```

- [ ] **Step 3: Write the synthetic seed**

`testdb/seed.sql`:
```sql
INSERT INTO teams (name, description, league, division, active, youth, ages) VALUES
    ('First Team', 'Senior men', 'Hellenic League', 'Division One East', TRUE, FALSE, 99),
    ('Under 12s', 'Youth squad', 'Youth League', 'U12', TRUE, TRUE, 12),
    ('Old Team', NULL, NULL, NULL, FALSE, FALSE, 99);

INSERT INTO users (name, email, phone, team_id, role, file_name) VALUES
    ('Web Master', 'webmaster@example.test', '01234 567890', 0, 'WEBMASTER', NULL),
    ('Team Manager', 'manager@example.test', NULL, 1, 'MANAGER', NULL),
    ('Club Secretary', 'secretary@example.test', NULL, 0, 'CLUB_SECRETARY', 'user/secretary.png');

INSERT INTO players (name, file_name, date_of_birth, position, captain, team_id) VALUES
    ('Adult Player', 'player/adult.png', '1990-05-01', 'Striker', TRUE, 1),
    ('Youth Player', 'player/youth.png', '2014-03-02', 'Keeper', FALSE, 2),
    ('Young Senior', 'player/young.png', CURRENT_DATE - INTERVAL '16 years', 'Defender', FALSE, 1);

INSERT INTO sponsors (name, website, file_name, purpose, team_id) VALUES
    ('Club Sponsor', 'https://sponsor.example.test', 'sponsor/club.png', 'Kit', 'A'),
    ('Team Sponsor', NULL, 'sponsor/team.png', NULL, '1');

INSERT INTO affiliations (name, website, file_name) VALUES
    ('County FA', 'https://fa.example.test', 'affiliation/fa.png');

INSERT INTO documents (name, file_name) VALUES ('Club Rules', 'document/rules.pdf');

INSERT INTO images (file_name, caption) VALUES ('gallery/one.jpg', 'Match day');

INSERT INTO news (title, file_name, content, date) VALUES
    ('Season opener', 'news/opener.jpg', '<p>We won!</p>', NOW() - INTERVAL '2 days'),
    ('Training update', NULL, '<p>Tuesday 7pm</p>', NOW() - INTERVAL '1 day');

INSERT INTO whatson (title, file_name, content, date, date_of_event) VALUES
    ('Presentation night', NULL, '<p>Clubhouse</p>', NOW(), CURRENT_DATE + 30),
    ('Summer BBQ', NULL, '<p>Done</p>', NOW(), CURRENT_DATE - 30);

INSERT INTO programme_seasons (season) VALUES ('2025-26');

INSERT INTO programmes (name, file_name, date_of_programme, programme_season_id) VALUES
    ('Opening day programme', 'programme/opening.pdf', CURRENT_DATE - 10, 1);

INSERT INTO settings (id, setting_text) VALUES
    ('displayEmail', 'hello@example.test'),
    ('infoContent', '<div>Club history</div>'),
    ('visitorCount', '42');
```

- [ ] **Step 4: Write the testdb helper**

`testdb/testdb.go`:
```go
// Package testdb gives DB-backed tests a throwaway Postgres schema loaded
// with a synthetic fixture. Tests are skipped when AFC_TEST_DB is unset.
package testdb

import (
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"net/url"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
	"github.com/stretchr/testify/require"
)

//go:embed schema.sql
var schemaSQL string

//go:embed seed.sql
var seedSQL string

// Open creates a uniquely named schema in the AFC_TEST_DB database, loads
// schema.sql and seed.sql into it, and returns a connection plus a DSN whose
// search_path points at that schema. The schema is dropped when the test ends.
func Open(t *testing.T) (*sqlx.DB, string) {
	t.Helper()
	base := os.Getenv("AFC_TEST_DB")
	if base == "" {
		t.Skip("AFC_TEST_DB not set; skipping DB-backed test")
	}

	suffix := make([]byte, 6)
	_, err := rand.Read(suffix)
	require.NoError(t, err)
	schema := "t_" + hex.EncodeToString(suffix)

	admin, err := sqlx.Connect("postgres", base)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = admin.Exec("DROP SCHEMA IF EXISTS " + schema + " CASCADE")
		_ = admin.Close()
	})
	_, err = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)

	u, err := url.Parse(base)
	require.NoError(t, err)
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	dsn := u.String()

	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(schemaSQL)
	require.NoError(t, err, "loading schema.sql")
	_, err = db.Exec(seedSQL)
	require.NoError(t, err, "loading seed.sql")
	return db, dsn
}
```

- [ ] **Step 5: Write the legacy smoke test**

`router_smoke_test.go`:
```go
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/infrastructure/storage"
	"github.com/COMTOP1/AFC-GO/testdb"
	"github.com/COMTOP1/AFC-GO/views"
)

// TestLegacyPagesSmoke guards the template site through the restructure:
// every public page renders and every guarded page redirects when logged out.
func TestLegacyPagesSmoke(t *testing.T) {
	_, dsn := testdb.Open(t)

	conf := &views.Config{
		DatabaseURL:       dsn,
		DomainName:        "localhost",
		SessionCookieName: "session",
		S3: storage.Config{
			Endpoint: "http://s3.invalid", Region: "us-east-1", Bucket: "test",
			AccessKey: "test", SecretKey: "test",
		},
		Security: views.SecurityConfig{
			Iterations: 1, ScryptWorkFactor: 2, ScryptBlockSize: 1, ScryptParallelismFactor: 1, KeyLength: 32,
		},
	}
	v := views.New(conf, "test", time.Hour)
	t.Cleanup(v.Stop)
	r := NewRouter(&RouterConf{Config: conf, Views: v})

	cases := []struct {
		path string
		want int
	}{
		{"/", http.StatusOK},
		{"/news", http.StatusOK},
		{"/news/1", http.StatusOK},
		{"/whatson", http.StatusOK},
		{"/whatson/1", http.StatusOK},
		{"/teams", http.StatusOK},
		{"/team/1", http.StatusOK},
		{"/sponsors", http.StatusOK},
		{"/gallery", http.StatusOK},
		{"/documents", http.StatusOK},
		{"/programmes", http.StatusOK},
		{"/info", http.StatusOK},
		{"/contact", http.StatusOK},
		{"/public/stylesheet.css", http.StatusOK},
		{"/does-not-exist", http.StatusNotFound},
		{"/players", http.StatusFound},
		{"/account", http.StatusFound},
		{"/users", http.StatusFound},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			r.router.ServeHTTP(rec, req)
			assert.Equal(t, tc.want, rec.Code, "body: %.300s", rec.Body.String())
		})
	}
}
```

- [ ] **Step 6: Run it with a database**

```bash
docker run -d --name afc-test-db -e POSTGRES_PASSWORD=test -p 55432:5432 postgres:17-alpine
export AFC_TEST_DB='postgres://postgres:test@localhost:55432/postgres?sslmode=disable'
go test -run TestLegacyPagesSmoke -v .
```
Expected: PASS for every sub-test. If a page 500s, the fixture is missing a column the page needs. Fix `schema.sql`/`seed.sql` (not the app). If you have prod access, cross-check column types against `pg_dump --schema-only` output **without committing that output**.

- [ ] **Step 7: Confirm it skips without a database**

```bash
env -u AFC_TEST_DB go test -run TestLegacyPagesSmoke -v .
```
Expected: `--- SKIP: TestLegacyPagesSmoke`.

- [ ] **Step 8: Commit**

```bash
git add testdb router_smoke_test.go go.mod go.sum
git commit -m "Add synthetic test DB fixture and legacy page smoke test

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 2: Move into the `server/` layout

A pure move: no behaviour changes. The smoke test from Task 1 proves it.

**Files:**
- Move: everything listed in Step 1
- Create: `server/internal/legacy/assets.go`
- Modify: `server/cmd/afc/router.go` (embed), all Go import paths, `Dockerfile`, `.github/workflows/go-template-lint.yaml`, `.github/workflows/nilaway.yaml`
- Delete: `middleware/` (unused; verified by `grep -rn '"github.com/COMTOP1/AFC-GO/middleware"'` returning nothing)

**Interfaces:**
- Produces: import paths `github.com/COMTOP1/AFC-GO/server/internal/<pkg>` for every package, and `legacy.Public embed.FS` (contains `public/…`).

- [ ] **Step 1: Move files with git**

```bash
mkdir -p server/cmd server/internal/legacy
mkdir -p server/cmd/afc
git mv main.go router.go router_smoke_test.go server/cmd/afc/
git mv cmd/migrates3 server/cmd/migrates3
git mv offline server/cmd/offline
for p in affiliation document image news player programme role setting sponsor team user whatson utils infrastructure testdb; do
  git mv "$p" "server/internal/$p"
done
git mv views server/internal/legacy/views
git mv templates server/internal/legacy/templates
git mv public server/internal/legacy/public
git rm -r -q middleware
rmdir cmd 2>/dev/null || true
for d in affiliation document image news player programme setting sponsor team user whatson; do
  git mv "server/internal/$d/db.go" "server/internal/$d/store.go"
done
```

- [ ] **Step 2: Rewrite import paths**

```bash
grep -rl --include='*.go' 'github.com/COMTOP1/AFC-GO/' server | xargs sed -i '' \
  -e 's#"github.com/COMTOP1/AFC-GO/views"#"github.com/COMTOP1/AFC-GO/server/internal/legacy/views"#' \
  -e 's#"github.com/COMTOP1/AFC-GO/templates"#"github.com/COMTOP1/AFC-GO/server/internal/legacy/templates"#' \
  -E -e 's#"github.com/COMTOP1/AFC-GO/(affiliation|document|image|news|player|programme|role|setting|sponsor|team|user|whatson|utils|infrastructure|testdb)#"github.com/COMTOP1/AFC-GO/server/internal/\1#'
```
(`sed -i ''` is BSD/macOS syntax; on Linux use `sed -i`.) Also update the tracer names so spans keep matching their packages:
```bash
grep -rl --include='*.go' 'otel.Tracer("github.com/COMTOP1/AFC-GO/' server | xargs sed -i '' -E \
  -e 's#otel.Tracer\("github.com/COMTOP1/AFC-GO/(views|templates)"#otel.Tracer("github.com/COMTOP1/AFC-GO/server/internal/legacy/\1"#' \
  -e 's#otel.Tracer\("github.com/COMTOP1/AFC-GO/(affiliation|document|image|news|player|programme|setting|sponsor|team|user|whatson|infrastructure/[a-z]+)"#otel.Tracer("github.com/COMTOP1/AFC-GO/server/internal/\1"#' \
  -e 's#otel.Tracer\("github.com/COMTOP1/AFC-GO/cmd/migrates3"#otel.Tracer("github.com/COMTOP1/AFC-GO/server/cmd/migrates3"#'
```

- [ ] **Step 3: Re-home the embedded public assets**

`router.go` embedded `public/*` from its own directory; `public/` now lives in `legacy`. Create `server/internal/legacy/assets.go`:
```go
// Package legacy holds the server-rendered template site that is being
// replaced by the React client. It is deleted once the client ships.
package legacy

import "embed"

// Public holds the legacy static assets served under /public/.
//
//go:embed public
var Public embed.FS
```
In `server/cmd/afc/router.go`:
- delete the `//go:embed public/*` line and the `var embeddedFiles embed.FS` declaration
- drop the `"embed"` import
- add `"github.com/COMTOP1/AFC-GO/server/internal/legacy"`
- change the asset handler line to:
```go
	assetHandler := http.FileServer(http.FS(echo.MustSubFS(legacy.Public, "public")))
```

- [ ] **Step 4: Update build and CI paths**

`Dockerfile`: replace the two build lines with:
```dockerfile
RUN GOOS=linux GOARCH=amd64 go build -ldflags="$(cat ./ldflags)" -o /bin/afc ./server/cmd/afc
RUN GOOS=linux GOARCH=amd64 go build -o /bin/migrates3 ./server/cmd/migrates3
```
`.github/workflows/go-template-lint.yaml`, last line:
```yaml
        run: go-template-lint -f=server/internal/legacy/templates/template.go -td=server/internal/legacy/templates/ -t=server/internal/legacy/templates/template.go
```
`.github/workflows/nilaway.yaml`, last line (same package set as the old `nilaway .`):
```yaml
        run: nilaway ./server/cmd/afc
```

- [ ] **Step 5: Build, vet, lint, test**

```bash
gofmt -l server; go build ./... && go vet ./...
go tool golangci-lint run ./... || go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run ./...
go test ./...
```
Expected: no gofmt output, build OK, lint clean, and all tests pass (the smoke test runs if `AFC_TEST_DB` is exported). If lint isn't available as a tool yet, add it the same way MV-Controller does: `go get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`, then commit `go.mod`/`go.sum` with this task.

- [ ] **Step 6: Run the server once by hand**

```bash
cp /Users/liam/Code/Go/AFC/.env .env 2>/dev/null || true   # local dev env, gitignored
go run ./server/cmd/afc &   sleep 3; curl -s -o /dev/null -w '%{http_code}\n' localhost:${ADDRESS##*:}/; kill %1
```
Expected: `200`. (Skip this step if you have no local `.env`; the smoke test already covers it.)

- [ ] **Step 7: Commit**

```bash
git add -A server Dockerfile .github go.mod go.sum
git status --short | grep -v '^R ' | head   # expect only the new assets.go and edits
git commit -m "Move Go code into server/cmd and server/internal (no behaviour change)

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---
## Phase B: Foundations

### Task 3: Service building blocks (`svcerr`, `upload`, `sanitize`, role predicates)

**Files:**
- Create: `server/internal/svcerr/svcerr.go`, `server/internal/svcerr/svcerr_test.go`
- Create: `server/internal/upload/upload.go`, `server/internal/upload/upload_test.go`, `server/internal/upload/uploadtest/fake.go`
- Create: `server/internal/sanitize/sanitize.go`, `server/internal/sanitize/sanitize_test.go`
- Modify: `server/internal/role/role.go`; Create: `server/internal/role/role_test.go`

**Interfaces:**
- Produces (`svcerr`):
  - Kinds: `KindNotFound`, `KindForbidden`, `KindInvalid`, `KindConflict`.
  - `type Error struct{Kind; Message string; Fields map[string]string; Err error}`.
  - Constructors: `NotFound(msg string, err error) *Error`, `Forbidden(msg string) *Error`, `Conflict(msg string, err error) *Error`, `Invalid(fields map[string]string) *Error`, `InvalidField(field, msg string) *Error`.
  - `type Fields map[string]string` with `Add(field, msg)` and `Err() error`.
  - `FromStore(err error, what string) error`, which maps `sql.ErrNoRows` → NotFound.
  - `As(err error) (*Error, bool)`.
- Produces (`upload`):
  - `type File struct{Name, ContentType string; Size int64; Open func() (io.ReadCloser, error)}` and `FromHeader(*multipart.FileHeader) *File`.
  - `type Storage interface{Put; Delete; Exists; PublicURL}`.
  - `type Files`, with `New(Storage) *Files` and methods:
    - `(*Files).Save(ctx, *File, category string) (key string, err error)`
    - `(*Files).Remove(ctx, key string)` (logs errors, never returns them)
    - `(*Files).URL(key string) string`
    - `(*Files).Exists(ctx, key) (bool, error)`
- Produces (`uploadtest`): `New() *Storage` (in-memory fake with `Objects map[string]string`, `Deleted []string`, `PutErr error`) and `File(name, contentType, body string) *upload.File`. `PublicURL(k)` returns `"https://cdn.test/" + k`.
- Produces (`sanitize`): `HTML(string) string`, the exact bluemonday policy used by news, what's on and info today.
- Produces (`role`): `(Role).CanEdit() bool` (not Manager, not Photographer), `(Role).CanManageGallery() bool` (not Manager), `(Role).IsClubSecretaryHigher() bool` (SafeguardingOfficer, ClubSecretary, Chairperson, Webmaster).

- [ ] **Step 1: Write the failing tests**

`server/internal/svcerr/svcerr_test.go`:
```go
package svcerr_test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
)

func TestFromStoreMapsNoRowsToNotFound(t *testing.T) {
	err := svcerr.FromStore(fmt.Errorf("failed to get news article: %w", sql.ErrNoRows), "news article")
	se, ok := svcerr.As(err)
	require.True(t, ok)
	assert.Equal(t, svcerr.KindNotFound, se.Kind)
	assert.Equal(t, "news article not found", se.Message)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestFromStoreWrapsOtherErrors(t *testing.T) {
	boom := errors.New("connection refused")
	err := svcerr.FromStore(boom, "news article")
	_, ok := svcerr.As(err)
	assert.False(t, ok)
	assert.ErrorIs(t, err, boom)
	assert.NoError(t, svcerr.FromStore(nil, "x"))
}

func TestInvalidErrorListsFieldsSorted(t *testing.T) {
	err := svcerr.Invalid(map[string]string{"title": "title is required", "file": "unsupported file type: text/html"})
	assert.Equal(t, "validation failed: file: unsupported file type: text/html; title: title is required", err.Error())
}

func TestFieldsErr(t *testing.T) {
	f := svcerr.Fields{}
	require.NoError(t, f.Err())
	f.Add("name", "name is required")
	f.Add("name", "second message is ignored")
	se, ok := svcerr.As(f.Err())
	require.True(t, ok)
	assert.Equal(t, map[string]string{"name": "name is required"}, se.Fields)
}
```

`server/internal/upload/upload_test.go`:
```go
package upload_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

func TestSaveStoresUnderCategoryWithExtension(t *testing.T) {
	store := uploadtest.New()
	files := upload.New(store)

	key, err := files.Save(context.Background(), uploadtest.File("photo.png", "image/png", "PNGDATA"), "news")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(key, "news/"), key)
	assert.True(t, strings.HasSuffix(key, ".png"), key)
	assert.Equal(t, "PNGDATA", store.Objects[key])
	assert.Equal(t, "https://cdn.test/"+key, files.URL(key))
}

func TestSaveRejectsUnsupportedType(t *testing.T) {
	for _, ct := range []string{"text/html", "application/x-msdownload", "application/octet-stream", ""} {
		t.Run(ct, func(t *testing.T) {
			store := uploadtest.New()
			_, err := upload.New(store).Save(context.Background(), uploadtest.File("x", ct, "<script>"), "news")
			se, ok := svcerr.As(err)
			require.True(t, ok, "want svcerr, got %v", err)
			assert.Equal(t, svcerr.KindInvalid, se.Kind)
			assert.Contains(t, se.Fields, "file")
			assert.Empty(t, store.Objects)
		})
	}
}

func TestRemoveIgnoresEmptyKeyAndLogsFailures(t *testing.T) {
	store := uploadtest.New()
	files := upload.New(store)
	files.Remove(context.Background(), "")
	assert.Empty(t, store.Deleted)
	files.Remove(context.Background(), "news/a.png")
	assert.Equal(t, []string{"news/a.png"}, store.Deleted)
	assert.Empty(t, files.URL(""))
}
```

`server/internal/sanitize/sanitize_test.go`:
```go
package sanitize_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/sanitize"
)

func TestHTML(t *testing.T) {
	cases := map[string]struct{ in, want string }{
		"script removed":        {`<p>hi</p><script>alert(1)</script>`, `<p>hi</p>`},
		"heading class kept":    {`<h2 class="title">T</h2>`, `<h2 class="title">T</h2>`},
		"mailto kept":           {`<a href="mailto:a@b.c">x</a>`, `<a href="mailto:a@b.c">x</a>`},
		"javascript href gone":  {`<a href="javascript:alert(1)">x</a>`, `x`},
		"alignment style kept":  {`<div style="text-align: center">c</div>`, `<div style="text-align: center">c</div>`},
		"img not allowed":       {`<img src="x" onerror="alert(1)">`, ``},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, sanitize.HTML(tc.in))
		})
	}
}
```

`server/internal/role/role_test.go`:
```go
package role_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/role"
)

func TestPredicates(t *testing.T) {
	all := []role.Role{role.Photographer, role.Manager, role.ProgrammeEditor, role.LeagueSecretary,
		role.Treasurer, role.SafeguardingOfficer, role.ClubSecretary, role.Chairperson, role.Webmaster}
	for _, r := range all {
		assert.Equal(t, r != role.Manager && r != role.Photographer, r.CanEdit(), "CanEdit %s", r)
		assert.Equal(t, r != role.Manager, r.CanManageGallery(), "CanManageGallery %s", r)
		want := r == role.SafeguardingOfficer || r == role.ClubSecretary || r == role.Chairperson || r == role.Webmaster
		assert.Equal(t, want, r.IsClubSecretaryHigher(), "IsClubSecretaryHigher %s", r)
	}
}
```

- [ ] **Step 2: Run to confirm they fail**

Run: `go test ./server/internal/svcerr/... ./server/internal/upload/... ./server/internal/sanitize/... ./server/internal/role/...`
Expected: compile failures (`undefined: svcerr.FromStore`, `undefined: upload.New`, …).

- [ ] **Step 3: Implement `svcerr`**

`server/internal/svcerr/svcerr.go`:
```go
// Package svcerr defines the typed errors services return, so the JSON API
// and the legacy template views can map them to responses consistently.
package svcerr

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Kind classifies a service error.
type Kind int

const (
	KindNotFound Kind = iota + 1
	KindForbidden
	KindInvalid
	KindConflict
)

// Error is a service error that transports translate into a response.
type Error struct {
	Kind    Kind
	Message string
	Fields  map[string]string
	Err     error
}

func (e *Error) Error() string {
	msg := e.Message
	if len(e.Fields) > 0 {
		keys := make([]string, 0, len(e.Fields))
		for k := range e.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, k+": "+e.Fields[k])
		}
		msg += ": " + strings.Join(parts, "; ")
	}
	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}
	return msg
}

func (e *Error) Unwrap() error { return e.Err }

func NotFound(message string, err error) *Error {
	return &Error{Kind: KindNotFound, Message: message, Err: err}
}

func Forbidden(message string) *Error {
	return &Error{Kind: KindForbidden, Message: message}
}

func Conflict(message string, err error) *Error {
	return &Error{Kind: KindConflict, Message: message, Err: err}
}

func Invalid(fields map[string]string) *Error {
	return &Error{Kind: KindInvalid, Message: "validation failed", Fields: fields}
}

func InvalidField(field, message string) *Error {
	return Invalid(map[string]string{field: message})
}

// Fields accumulates validation messages; the first message per field wins.
type Fields map[string]string

func (f Fields) Add(field, message string) {
	if _, ok := f[field]; !ok {
		f[field] = message
	}
}

// Err returns an Invalid error, or nil when no field failed.
func (f Fields) Err() error {
	if len(f) == 0 {
		return nil
	}
	return Invalid(f)
}

// FromStore turns a store error into a service error: sql.ErrNoRows
// becomes NotFound, anything else is wrapped unchanged.
func FromStore(err error, what string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return NotFound(what+" not found", err)
	}
	return fmt.Errorf("%s: %w", what, err)
}

// As reports whether err is (or wraps) a service error.
func As(err error) (*Error, bool) {
	var e *Error
	ok := errors.As(err, &e)
	return e, ok
}
```

- [ ] **Step 4: Implement `upload` and its fake**

`server/internal/upload/upload.go`:
```go
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
	defer src.Close()

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
```

`server/internal/upload/uploadtest/fake.go`:
```go
// Package uploadtest provides an in-memory upload.Storage for tests.
package uploadtest

import (
	"context"
	"io"
	"strings"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

// Storage is an in-memory object store.
type Storage struct {
	mu      sync.Mutex
	Objects map[string]string
	Deleted []string
	PutErr  error
}

func New() *Storage {
	return &Storage{Objects: map[string]string{}}
}

func (s *Storage) Put(_ context.Context, key string, body io.Reader, _ int64, _ string) error {
	if s.PutErr != nil {
		return s.PutErr
	}
	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Objects[key] = string(b)
	return nil
}

func (s *Storage) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Deleted = append(s.Deleted, key)
	delete(s.Objects, key)
	return nil
}

func (s *Storage) Exists(_ context.Context, key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.Objects[key]
	return ok, nil
}

func (s *Storage) PublicURL(key string) string {
	return "https://cdn.test/" + key
}

// File builds an upload.File with the given contents.
func File(name, contentType, body string) *upload.File {
	return &upload.File{
		Name:        name,
		ContentType: contentType,
		Size:        int64(len(body)),
		Open:        func() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader(body)), nil },
	}
}
```

- [ ] **Step 5: Implement `sanitize`**

`server/internal/sanitize/sanitize.go`:
```go
// Package sanitize cleans rich-text HTML submitted from the editors.
package sanitize

import "github.com/microcosm-cc/bluemonday"

// policy is the legacy editor policy (views/news.go, whatson.go, info.go).
// A bluemonday policy is safe for concurrent use once built.
var policy = func() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements("a", "ul", "ol", "li", "h2", "b", "i", "u", "strike", "div", "br", "p",
		"blockquote", "pre", "hr")
	p.AllowAttrs("class").OnElements("h2")
	p.AllowAttrs("href", "style").OnElements("a")
	p.AllowURLSchemes("mailto", "http", "https")
	p.RequireNoFollowOnLinks(false)
	// Justification - via inline style
	p.AllowAttrs("style").OnElements("div", "p", "h2", "span")
	return p
}()

// HTML returns s with everything outside the editor policy removed.
func HTML(s string) string {
	return policy.Sanitize(s)
}
```
If bluemonday's exact output formatting differs from an expected string (for example, it may keep an empty `<a>` or reformat `style`), adjust the **expected string in the test**, not the policy, because the policy must stay identical to legacy. The adjusted expectations must still contain no `<script`, `javascript:` or `<img`.

- [ ] **Step 6: Add role predicates**

Append to `server/internal/role/role.go`:
```go
// CanEdit reports whether the role may manage site content (everyone except
// Managers and Photographers), matching RequireNotManagerNotPhotographer.
func (r Role) CanEdit() bool {
	return r != Manager && r != Photographer
}

// CanManageGallery reports whether the role may add or remove gallery
// images, matching RequireNotManager.
func (r Role) CanManageGallery() bool {
	return r != Manager
}

// IsClubSecretaryHigher matches RequireClubSecretaryHigher.
func (r Role) IsClubSecretaryHigher() bool {
	switch r {
	case SafeguardingOfficer, ClubSecretary, Chairperson, Webmaster:
		return true
	case Photographer, Manager, ProgrammeEditor, LeagueSecretary, Treasurer:
		return false
	}
	return false
}
```
These only look at the role. Callers must check the user is logged in first, because the empty role `""` passes `CanEdit`.

- [ ] **Step 7: Run tests**

Run: `go test ./server/internal/svcerr/... ./server/internal/upload/... ./server/internal/sanitize/... ./server/internal/role/...`
Expected: PASS.

- [ ] **Step 8: Lint and commit**

```bash
go tool golangci-lint run ./server/internal/svcerr/... ./server/internal/upload/... ./server/internal/sanitize/... ./server/internal/role/...
git add server/internal/svcerr server/internal/upload server/internal/sanitize server/internal/role
git commit -m "Add service error, upload, sanitize and role predicate building blocks

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---
### Task 4: `web`: error envelope, request helpers, CSRF, API group

**Files:**
- Create: `server/internal/web/errors.go`, `json.go`, `csrf.go`, `guards.go`, `api.go`
- Create: `server/internal/web/apitest/apitest.go`
- Create: `server/internal/web/errors_test.go`, `server/internal/web/api_test.go`, `server/internal/web/json_test.go`
- Modify: `go.mod`/`go.sum` (add `github.com/swaggo/echo-swagger`)

**Interfaces:**
- Consumes: `svcerr` (Task 3), `upload.File`/`upload.FromHeader` (Task 3).
- Produces (`web`):
  - Response types: `type APIError struct{Code int; Message string; Fields map[string]string}`, `type ErrorResponse struct{Error APIError}`.
  - `ErrorHandler(legacy echo.HTTPErrorHandler) echo.HTTPErrorHandler`. It renders JSON for paths under `/api/` and delegates everything else to `legacy`.
  - `NewAPI(e *echo.Echo, secureCookies bool) *echo.Group`. It returns the `/api/v1` group with CSRF, swagger UI at `/api/v1/swagger/*`, `/api` and `/api/v1` → swagger redirects, JSON 404s, and `GET /api/v1/health` plus the legacy alias `GET /api/health`.
  - `type Guards struct{Login, Editor, NotManager, ClubSecretaryHigher echo.MiddlewareFunc}`.
  - Request helpers:
    - `BindJSON(c, dst any) error`
    - `ParamID(c, name string) (int, error)`
    - `FormFile(c, field) (*upload.File, error)` (nil when absent)
    - `FormString(c, field) *string` (nil when absent)
    - `FormBool(c, field) (*bool, error)`
    - `FormDate(c, field) (*time.Time, error)` (`YYYY-MM-DD`)
  - `const DateLayout = "2006-01-02"`
- Produces (`apitest`):
  - `NewEcho() *echo.Echo` (RemoveTrailingSlash + `web.ErrorHandler` whose legacy side writes `text/html` "legacy error")
  - `New(e) *Client` and `(*Client).As(cookie *http.Cookie) *Client`
  - Request methods: `(*Client).Get(t, path)`, `(*Client).JSON(t, method, path, body any)` and `(*Client).Multipart(t, method, path, fields map[string]string, files ...FilePart)`, all returning `*httptest.ResponseRecorder` and all sending `Sec-Fetch-Site: same-origin`. Also `(*Client).Do(t, *http.Request)`.
  - `type FilePart struct{Field, Name, ContentType, Body string}`
  - Decoders: `Decode[T any](t, rec) T` and `ErrorOf(t, rec) web.APIError`

- [ ] **Step 1: Add echo-swagger**

```bash
go get github.com/swaggo/echo-swagger@latest
```

- [ ] **Step 2: Write the failing tests**

`server/internal/web/errors_test.go`:
```go
package web_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestErrorMapping(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		code    int
		message string
		fields  map[string]string
	}{
		{"not found", svcerr.NotFound("news article not found", nil), 404, "news article not found", nil},
		{"forbidden", svcerr.Forbidden("nope"), 403, "nope", nil},
		{"invalid", svcerr.InvalidField("title", "title is required"), 422, "validation failed", map[string]string{"title": "title is required"}},
		{"conflict", svcerr.Conflict("email already used", nil), 409, "email already used", nil},
		{"echo error keeps message", echo.NewHTTPError(http.StatusUnauthorized, "login required"), 401, "login required", nil},
		{"echo 5xx hides message", echo.NewHTTPError(http.StatusBadGateway, "upstream exploded at 10.0.0.1"), 502, "Bad Gateway", nil},
		{"plain error is 500", errors.New("pq: connection refused"), 500, "internal server error", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := apitest.NewEcho()
			api := web.NewAPI(e, false)
			api.GET("/boom", func(echo.Context) error { return tc.err })

			rec := apitest.New(e).Get(t, "/api/v1/boom")
			assert.Equal(t, tc.code, rec.Code)
			got := apitest.ErrorOf(t, rec)
			assert.Equal(t, tc.code, got.Code)
			assert.Equal(t, tc.message, got.Message)
			assert.Equal(t, tc.fields, got.Fields)
		})
	}
}

func TestNonAPIErrorsGoToLegacyHandler(t *testing.T) {
	e := apitest.NewEcho()
	web.NewAPI(e, false)
	e.GET("/news", func(echo.Context) error { return errors.New("boom") })

	rec := apitest.New(e).Get(t, "/news")
	assert.Equal(t, "legacy error", rec.Body.String())
}
```

`server/internal/web/api_test.go`:
```go
package web_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func newAPI() (*echo.Echo, *echo.Group) {
	e := apitest.NewEcho()
	e.RouteNotFound("/*", func(c echo.Context) error { return c.HTML(http.StatusNotFound, "<h1>legacy 404</h1>") })
	api := web.NewAPI(e, false)
	api.GET("/thing", func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	api.POST("/thing", func(c echo.Context) error { return c.NoContent(http.StatusCreated) })
	return e, api
}

func TestUnknownAPIPathIsJSON404(t *testing.T) {
	e, _ := newAPI()
	for _, path := range []string{"/api/v1/nope", "/api/v1/nope/deeper", "/api/nope"} {
		rec := apitest.New(e).Get(t, path)
		assert.Equal(t, http.StatusNotFound, rec.Code, path)
		assert.Equal(t, 404, apitest.ErrorOf(t, rec).Code, path)
	}
}

func TestWrongMethodIsJSON405(t *testing.T) {
	e, _ := newAPI()
	rec := apitest.New(e).JSON(t, http.MethodDelete, "/api/v1/thing", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Equal(t, 405, apitest.ErrorOf(t, rec).Code)
}

func TestHealth(t *testing.T) {
	e, _ := newAPI()
	for _, path := range []string{"/api/v1/health", "/api/health"} {
		rec := apitest.New(e).Get(t, path)
		assert.Equal(t, http.StatusOK, rec.Code, path)
		assert.JSONEq(t, `{"status":"ok"}`, rec.Body.String(), path)
	}
}

func TestSwaggerRedirects(t *testing.T) {
	e, _ := newAPI()
	for _, path := range []string{"/api", "/api/", "/api/v1", "/api/v1/"} {
		rec := apitest.New(e).Get(t, path)
		assert.Equal(t, http.StatusFound, rec.Code, path)
		assert.Equal(t, "/api/v1/swagger/index.html", rec.Header().Get("Location"), path)
	}
}

func TestSameOriginPostPasses(t *testing.T) {
	e, _ := newAPI()
	rec := apitest.New(e).JSON(t, http.MethodPost, "/api/v1/thing", map[string]string{})
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCrossSitePostRejected(t *testing.T) {
	e, _ := newAPI()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/thing", strings.NewReader("{}"))
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, 403, apitest.ErrorOf(t, rec).Code)
}

func TestTokenlessNonBrowserPostRejected(t *testing.T) {
	e, _ := newAPI()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/thing", strings.NewReader("{}"))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code, "missing token must be 403, not echo's default 400")
}

func TestDoubleSubmitTokenAccepted(t *testing.T) {
	e, _ := newAPI()
	// A GET without Sec-Fetch-Site issues the _csrf cookie.
	get := httptest.NewRecorder()
	e.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/api/v1/thing", nil))
	var token *http.Cookie
	for _, c := range get.Result().Cookies() {
		if c.Name == "_csrf" {
			token = c
		}
	}
	require.NotNil(t, token, "expected _csrf cookie")
	assert.False(t, token.HttpOnly, "client JS must be able to read _csrf")
	assert.Equal(t, http.SameSiteStrictMode, token.SameSite)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/thing", strings.NewReader("{}"))
	req.AddCookie(token)
	req.Header.Set("X-CSRF-Token", token.Value)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}
```

`server/internal/web/json_test.go`:
```go
package web_test

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestFormHelpers(t *testing.T) {
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	type seen struct {
		Title   *string `json:"title"`
		Missing *string `json:"missing"`
		Flag    *bool   `json:"flag"`
		Date    string  `json:"date"`
		File    string  `json:"file"`
		NoFile  bool    `json:"noFile"`
	}
	api.POST("/form", func(c echo.Context) error {
		var s seen
		s.Title = web.FormString(c, "title")
		s.Missing = web.FormString(c, "missing")
		var err error
		if s.Flag, err = web.FormBool(c, "flag"); err != nil {
			return err
		}
		d, err := web.FormDate(c, "date")
		if err != nil {
			return err
		}
		s.Date = d.Format(web.DateLayout)
		f, err := web.FormFile(c, "image")
		if err != nil {
			return err
		}
		s.File = f.Name
		none, err := web.FormFile(c, "other")
		if err != nil {
			return err
		}
		s.NoFile = none == nil
		return c.JSON(http.StatusOK, s)
	})

	rec := apitest.New(e).Multipart(t, http.MethodPost, "/api/v1/form",
		map[string]string{"title": "", "flag": "true", "date": "2026-09-01"},
		apitest.FilePart{Field: "image", Name: "a.png", ContentType: "image/png", Body: "x"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	got := apitest.Decode[seen](t, rec)
	require.NotNil(t, got.Title)
	assert.Empty(t, *got.Title, "present-but-empty must be a non-nil empty string")
	assert.Nil(t, got.Missing, "absent must be nil")
	require.NotNil(t, got.Flag)
	assert.True(t, *got.Flag)
	assert.Equal(t, "2026-09-01", got.Date)
	assert.Equal(t, "a.png", got.File)
	assert.True(t, got.NoFile)
}

func TestFormBoolAndDateValidation(t *testing.T) {
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	api.POST("/bool", func(c echo.Context) error { _, err := web.FormBool(c, "flag"); return err })
	api.POST("/date", func(c echo.Context) error { _, err := web.FormDate(c, "date"); return err })

	rec := apitest.New(e).Multipart(t, http.MethodPost, "/api/v1/bool", map[string]string{"flag": "maybe"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, apitest.ErrorOf(t, rec).Fields, "flag")

	rec = apitest.New(e).Multipart(t, http.MethodPost, "/api/v1/date", map[string]string{"date": "01/09/2026"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, apitest.ErrorOf(t, rec).Fields, "date")
}

func TestBindJSONRejectsUnknownFieldsAndBadIDs(t *testing.T) {
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	api.POST("/bind", func(c echo.Context) error {
		var in struct {
			Name string `json:"name"`
		}
		return web.BindJSON(c, &in)
	})
	api.GET("/items/:id", func(c echo.Context) error { _, err := web.ParamID(c, "id"); return err })

	rec := apitest.New(e).JSON(t, http.MethodPost, "/api/v1/bind", map[string]string{"nmae": "typo"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	for _, id := range []string{"abc", "0", "-3"} {
		rec = apitest.New(e).Get(t, "/api/v1/items/"+id)
		assert.Equal(t, http.StatusBadRequest, rec.Code, id)
	}
}
```

- [ ] **Step 3: Run to confirm they fail**

Run: `go test ./server/internal/web/...`
Expected: compile errors (`undefined: web.NewAPI`, package `apitest` not found).

- [ ] **Step 4: Implement `errors.go`**

`server/internal/web/errors.go`:
```go
// Package web holds the HTTP plumbing shared by every API handler: the JSON
// error envelope, request helpers, CSRF and the /api/v1 group.
package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
)

// APIPrefix is the path prefix whose errors are rendered as JSON.
const APIPrefix = "/api/"

// APIError is the body of every API error response.
type APIError struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// ErrorResponse wraps APIError as {"error": {...}}.
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// ErrorHandler renders errors on /api/* as the JSON envelope and hands every
// other path to the legacy (HTML) handler.
func ErrorHandler(legacy echo.HTTPErrorHandler) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		path := c.Request().URL.Path
		if path != "/api" && !strings.HasPrefix(path, APIPrefix) {
			legacy(err, c)
			return
		}
		if c.Response().Committed {
			return
		}
		ctx := c.Request().Context()
		trace.SpanFromContext(ctx).RecordError(err)

		apiErr := toAPIError(err)
		if apiErr.Code >= http.StatusInternalServerError {
			slog.ErrorContext(ctx, fmt.Sprintf("api error on %s %s: %+v", c.Request().Method, path, err))
		}

		var writeErr error
		if c.Request().Method == http.MethodHead {
			writeErr = c.NoContent(apiErr.Code)
		} else {
			writeErr = c.JSON(apiErr.Code, ErrorResponse{Error: apiErr})
		}
		if writeErr != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("failed to write api error: %+v", writeErr))
		}
	}
}

func toAPIError(err error) APIError {
	if se, ok := svcerr.As(err); ok {
		switch se.Kind {
		case svcerr.KindNotFound:
			return APIError{Code: http.StatusNotFound, Message: se.Message}
		case svcerr.KindForbidden:
			return APIError{Code: http.StatusForbidden, Message: se.Message}
		case svcerr.KindInvalid:
			return APIError{Code: http.StatusUnprocessableEntity, Message: se.Message, Fields: se.Fields}
		case svcerr.KindConflict:
			return APIError{Code: http.StatusConflict, Message: se.Message}
		}
	}

	var he *echo.HTTPError
	if errors.As(err, &he) {
		msg := http.StatusText(he.Code)
		if s, ok := he.Message.(string); ok && s != "" && he.Code < http.StatusInternalServerError {
			msg = s
		}
		return APIError{Code: he.Code, Message: msg}
	}

	return APIError{Code: http.StatusInternalServerError, Message: "internal server error"}
}
```

- [ ] **Step 5: Implement `json.go`**

`server/internal/web/json.go`:
```go
package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

// DateLayout is the date-only format API inputs use (dates without times,
// e.g. date of birth or date of event).
const DateLayout = "2006-01-02"

// BindJSON decodes the request body into dst, rejecting unknown fields.
func BindJSON(c echo.Context, dst any) error {
	dec := json.NewDecoder(c.Request().Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return svcerr.InvalidField("body", "invalid JSON: "+err.Error())
	}
	return nil
}

// ParamID parses a positive integer path parameter.
func ParamID(c echo.Context, name string) (int, error) {
	id, err := strconv.Atoi(c.Param(name))
	if err != nil || id < 1 {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+name)
	}
	return id, nil
}

// FormFile returns the uploaded file for field, or nil when none was sent.
func FormFile(c echo.Context, field string) (*upload.File, error) {
	fh, err := c.FormFile(field)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) || errors.Is(err, http.ErrNotMultipart) {
			return nil, nil //nolint:nilerr // uploads are optional; absence is not an error
		}
		return nil, svcerr.InvalidField(field, "invalid upload: "+err.Error())
	}
	return upload.FromHeader(fh), nil
}

// FormString returns the form value for field, or nil when the field was not
// sent at all (as opposed to sent empty).
func FormString(c echo.Context, field string) *string {
	form, err := c.FormParams()
	if err != nil {
		return nil
	}
	values, ok := form[field]
	if !ok || len(values) == 0 {
		return nil
	}
	v := values[0]
	return &v
}

// FormBool parses an optional boolean form field ("true"/"false"/"1"/"0").
func FormBool(c echo.Context, field string) (*bool, error) {
	raw := FormString(c, field)
	if raw == nil || *raw == "" {
		return nil, nil
	}
	b, err := strconv.ParseBool(*raw)
	if err != nil {
		return nil, svcerr.InvalidField(field, "must be true or false")
	}
	return &b, nil
}

// FormDate parses an optional YYYY-MM-DD form field.
func FormDate(c echo.Context, field string) (*time.Time, error) {
	raw := FormString(c, field)
	if raw == nil || *raw == "" {
		return nil, nil
	}
	d, err := time.Parse(DateLayout, *raw)
	if err != nil {
		return nil, svcerr.InvalidField(field, "must be a date in YYYY-MM-DD format")
	}
	return &d, nil
}
```

- [ ] **Step 6: Implement `csrf.go`, `guards.go` and `api.go`**

`server/internal/web/csrf.go`:
```go
package web

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CSRF protects unsafe methods. Browsers that send Sec-Fetch-Site:
// same-origin pass without a token; other clients must echo the _csrf
// cookie in X-CSRF-Token. Every failure is a 403.
func CSRF(secureCookie bool) echo.MiddlewareFunc {
	return middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:X-CSRF-Token",
		CookieName:     "_csrf",
		CookiePath:     "/",
		CookieSameSite: http.SameSiteStrictMode,
		CookieSecure:   secureCookie,
		CookieHTTPOnly: false,
		ErrorHandler: func(_ error, _ echo.Context) error {
			return echo.NewHTTPError(http.StatusForbidden, "invalid or missing CSRF token")
		},
	})
}
```

`server/internal/web/guards.go`:
```go
package web

import "github.com/labstack/echo/v4"

// Guards are the authorisation middlewares handlers attach to routes. They
// are built by the auth package; web only defines the shape so domain
// packages don't import auth.
type Guards struct {
	Login               echo.MiddlewareFunc // any logged-in user
	Editor              echo.MiddlewareFunc // logged in, not Manager or Photographer
	NotManager          echo.MiddlewareFunc // logged in, not Manager (gallery)
	ClubSecretaryHigher echo.MiddlewareFunc // Safeguarding Officer, Club Secretary, Chairperson, Webmaster
}
```

`server/internal/web/api.go`:
```go
package web

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

const swaggerIndex = "/api/v1/swagger/index.html"

// NewAPI mounts /api and returns the /api/v1 group, protected by CSRF.
// Unknown /api paths are JSON 404s rather than the legacy HTML page.
func NewAPI(e *echo.Echo, secureCookies bool) *echo.Group {
	api := e.Group("/api")
	api.GET("", swaggerRedirect)
	api.GET("/health", Health) // legacy alias, kept for existing monitors
	api.RouteNotFound("/*", func(echo.Context) error { return echo.ErrNotFound })

	v1 := api.Group("/v1", CSRF(secureCookies))
	v1.GET("", swaggerRedirect)
	v1.GET("/health", Health)
	v1.GET("/swagger/*", echoSwagger.WrapHandler)
	return v1
}

func swaggerRedirect(c echo.Context) error {
	return c.Redirect(http.StatusFound, swaggerIndex)
}

// HealthResponse is the body of GET /health.
type HealthResponse struct {
	Status string `json:"status"`
}

// Health reports that the server is up.
//
//	@Summary	Health check
//	@Tags		system
//	@Produce	json
//	@Success	200	{object}	web.HealthResponse
//	@Router		/health [get]
func Health(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}
```
`Group.Use` (called by `Group("/v1", CSRF(...))`) also registers group `RouteNotFound` handlers, so `/api/v1/*` misses come back as `echo.ErrNotFound` → JSON.

- [ ] **Step 7: Implement `apitest`**

`server/internal/web/apitest/apitest.go`:
```go
// Package apitest drives the API in handler tests the way a same-origin
// browser would.
package apitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// NewEcho returns an Echo set up like production for API routes. Non-API
// errors write "legacy error" so tests can tell the two paths apart.
func NewEcho() *echo.Echo {
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	e.HTTPErrorHandler = web.ErrorHandler(func(_ error, c echo.Context) {
		_ = c.HTML(http.StatusInternalServerError, "legacy error")
	})
	return e
}

// Client sends requests to an Echo instance, optionally with a session cookie.
type Client struct {
	e      *echo.Echo
	cookie *http.Cookie
}

func New(e *echo.Echo) *Client { return &Client{e: e} }

// As returns a client that sends cookie with every request.
func (c *Client) As(cookie *http.Cookie) *Client { return &Client{e: c.e, cookie: cookie} }

// Do serves req as a same-origin browser request.
func (c *Client) Do(t *testing.T, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	if c.cookie != nil {
		req.AddCookie(c.cookie)
	}
	rec := httptest.NewRecorder()
	c.e.ServeHTTP(rec, req)
	return rec
}

func (c *Client) Get(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	return c.Do(t, httptest.NewRequest(http.MethodGet, path, nil))
}

// JSON sends body encoded as JSON; a nil body sends no body.
func (c *Client) JSON(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, r)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.Do(t, req)
}

// FilePart is one file in a multipart request.
type FilePart struct {
	Field, Name, ContentType, Body string
}

// Multipart sends fields and files as multipart/form-data.
func (c *Client) Multipart(t *testing.T, method, path string, fields map[string]string, files ...FilePart) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		require.NoError(t, w.WriteField(k, v))
	}
	for _, f := range files {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, f.Field, f.Name))
		h.Set("Content-Type", f.ContentType)
		part, err := w.CreatePart(h)
		require.NoError(t, err)
		_, err = part.Write([]byte(f.Body))
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set(echo.HeaderContentType, w.FormDataContentType())
	return c.Do(t, req)
}

// Decode unmarshals the response body into T.
func Decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &v), "body: %s", rec.Body.String())
	return v
}

// ErrorOf decodes the JSON error envelope.
func ErrorOf(t *testing.T, rec *httptest.ResponseRecorder) web.APIError {
	t.Helper()
	return Decode[web.ErrorResponse](t, rec).Error
}
```

- [ ] **Step 8: Run tests**

Run: `go test ./server/internal/web/...`
Expected: PASS. If `TestWrongMethodIsJSON405` gets 404, Echo matched the group's `RouteNotFound("/*")` before the method check. In that case assert `404` there and note it in the commit message. It is still a JSON envelope, which is the behaviour that matters.

- [ ] **Step 9: Lint and commit**

```bash
go tool golangci-lint run ./server/internal/web/...
git add server/internal/web go.mod go.sum
git commit -m "Add web package: JSON error envelope, request helpers, CSRF, /api/v1 group

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---
### Task 5: `auth`: cookie sessions, guards, `GET /auth/me`

**Files:**
- Create: `server/internal/auth/sessions.go`, `guards.go`, `types.go`, `handlers.go`
- Create: `server/internal/auth/authtest/authtest.go`
- Create: `server/internal/auth/sessions_test.go`, `server/internal/auth/guards_test.go`

**Interfaces:**
- Consumes:
  - `web.Guards`, `web.NewAPI`, `apitest.*` (Task 4)
  - `upload.Files` (Task 3)
  - the role predicates (Task 3)
  - `user.User` (existing)
- Produces (`auth`):
  - `type Config struct{CookieName, AuthenticationKey, EncryptionKey string; Secure bool}` (keys are hex, as in `.env` today).
  - `type UserGetter interface{GetUser(ctx, user.User) (user.User, error)}`, satisfied by `*user.Store`.
  - `NewSessions(Config, UserGetter) *Sessions`, with methods:
    - `CookieStore() *sessions.CookieStore`
    - `Name() string`
    - `User(r *http.Request) (user.User, bool)`
    - `Login(w, r, u user.User, remember bool) error`
    - `Logout(w, r) error`
    - `RequireLogin`
    - `RequireRole(func(role.Role) bool) echo.MiddlewareFunc`
    - `Guards() web.Guards`
  - `Current(c echo.Context) (user.User, bool)`, which returns the user a guard loaded.
  - `type CurrentUser` and `NewCurrentUser(user.User, *upload.Files) CurrentUser`.
  - `NewHandlers(*Sessions, *upload.Files) *Handlers` and `(*Handlers).Register(g *echo.Group, guards web.Guards)`, which registers `GET /auth/me`.
  - `gob.Register(user.User{})` now lives in `auth` (in `init`).
- Produces (`authtest`):
  - Personas: `Webmaster` (ID 1), `Manager` (2, team 1), `Photographer` (3), `Treasurer` (4).
  - `New(users ...user.User) *auth.Sessions`
  - `Everyone() *auth.Sessions`
  - `Cookie(t, s, u) *http.Cookie`
  - `Files() *upload.Files` (an `uploadtest`-backed files instance)

- [ ] **Step 1: Write the failing tests**

`server/internal/auth/sessions_test.go`:
```go
package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func newMeAPI(s *auth.Sessions) *apitest.Client {
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	auth.NewHandlers(s, authtest.Files()).Register(api, s.Guards())
	return apitest.New(e)
}

func TestMeRequiresLogin(t *testing.T) {
	rec := newMeAPI(authtest.Everyone()).Get(t, "/api/v1/auth/me")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, 401, apitest.ErrorOf(t, rec).Code)
}

func TestMeReturnsUserAndPermissions(t *testing.T) {
	s := authtest.Everyone()
	rec := newMeAPI(s).As(authtest.Cookie(t, s, authtest.Treasurer)).Get(t, "/api/v1/auth/me")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	me := apitest.Decode[auth.CurrentUser](t, rec)
	assert.Equal(t, authtest.Treasurer.ID, me.ID)
	assert.Equal(t, "Treasurer", me.Role)
	assert.True(t, me.Permissions.CanEdit)
	assert.False(t, me.Permissions.CanManageUsers)
}

func TestMeNeverLeaksSecrets(t *testing.T) {
	u := authtest.Webmaster
	u.Hash = null.StringFrom("deadbeef")
	u.Salt = null.StringFrom("cafe")
	u.Password = null.StringFrom("hunter2")
	s := authtest.New(u)
	rec := newMeAPI(s).As(authtest.Cookie(t, s, u)).Get(t, "/api/v1/auth/me")
	require.Equal(t, http.StatusOK, rec.Code)
	for _, secret := range []string{"deadbeef", "cafe", "hunter2", "hash", "salt", "password"} {
		assert.NotContains(t, rec.Body.String(), secret)
	}
}

// TestLegacySessionCookieIsRecognised pins Review Focus #2: a cookie written
// exactly as legacy LoginFunc writes it must authenticate the API, so nobody
// is logged out by the deploy.
func TestLegacySessionCookieIsRecognised(t *testing.T) {
	s := authtest.Everyone()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	sess, _ := s.CookieStore().Get(req, s.Name())
	legacyUser := authtest.Webmaster
	legacyUser.Authenticated = true
	sess.Values["user"] = legacyUser
	require.NoError(t, sess.Save(req, rec))
	cookie := rec.Result().Cookies()[0]

	got := newMeAPI(s).As(cookie).Get(t, "/api/v1/auth/me")
	require.Equal(t, http.StatusOK, got.Code, got.Body.String())
	assert.Equal(t, authtest.Webmaster.ID, apitest.Decode[auth.CurrentUser](t, got).ID)
}

func TestLogoutClearsSession(t *testing.T) {
	s := authtest.Everyone()
	cookie := authtest.Cookie(t, s, authtest.Webmaster)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	require.NoError(t, s.Logout(rec, req))
	cleared := rec.Result().Cookies()[0]
	assert.Negative(t, cleared.MaxAge)

	check := httptest.NewRequest(http.MethodGet, "/", nil)
	check.AddCookie(cleared)
	_, ok := s.User(check)
	assert.False(t, ok)
}

func TestSessionCookieAttributes(t *testing.T) {
	s := auth.NewSessions(auth.Config{CookieName: "session", Secure: true}, authtest.NewUsers())
	rec := httptest.NewRecorder()
	require.NoError(t, s.Login(rec, httptest.NewRequest(http.MethodPost, "/", nil), user.User{ID: 1}, true))
	c := rec.Result().Cookies()[0]
	assert.Equal(t, "session", c.Name)
	assert.True(t, c.HttpOnly)
	assert.True(t, c.Secure)
	assert.Equal(t, http.SameSiteLaxMode, c.SameSite)
	assert.Equal(t, 86400*31, c.MaxAge, "remember=true keeps the session for 31 days")
}
```

`server/internal/auth/guards_test.go`:
```go
package auth_test

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestGuards(t *testing.T) {
	s := authtest.Everyone()
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	g := s.Guards()
	ok := func(c echo.Context) error { return c.NoContent(http.StatusNoContent) }
	api.GET("/login", ok, g.Login)
	api.GET("/editor", ok, g.Editor)
	api.GET("/gallery", ok, g.NotManager)
	api.GET("/secretary", ok, g.ClubSecretaryHigher)

	ghost := user.User{ID: 99, Email: "gone@example.test"} // has a cookie but no longer exists
	type want map[string]int
	cases := []struct {
		name string
		who  *user.User
		want want
	}{
		{"anonymous", nil, want{"/login": 401, "/editor": 401, "/gallery": 401, "/secretary": 401}},
		{"deleted user", &ghost, want{"/login": 401, "/editor": 401, "/gallery": 401, "/secretary": 401}},
		{"manager", &authtest.Manager, want{"/login": 204, "/editor": 403, "/gallery": 403, "/secretary": 403}},
		{"photographer", &authtest.Photographer, want{"/login": 204, "/editor": 403, "/gallery": 204, "/secretary": 403}},
		{"treasurer", &authtest.Treasurer, want{"/login": 204, "/editor": 204, "/gallery": 204, "/secretary": 403}},
		{"webmaster", &authtest.Webmaster, want{"/login": 204, "/editor": 204, "/gallery": 204, "/secretary": 204}},
	}
	for _, tc := range cases {
		client := apitest.New(e)
		if tc.who != nil {
			client = client.As(authtest.Cookie(t, s, *tc.who))
		}
		for path, code := range tc.want {
			rec := client.Get(t, "/api/v1"+path)
			assert.Equal(t, code, rec.Code, "%s GET %s", tc.name, path)
		}
	}
}

func TestGuardUsesFreshRole(t *testing.T) {
	// The cookie says Webmaster, but the DB now says Manager: the DB wins.
	demoted := authtest.Webmaster
	s := authtest.New(user.User{ID: demoted.ID, Name: demoted.Name, Email: demoted.Email, Role: authtest.Manager.Role})
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	api.GET("/editor", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) }, s.Guards().Editor)

	rec := apitest.New(e).As(authtest.Cookie(t, s, demoted)).Get(t, "/api/v1/editor")
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
```

- [ ] **Step 2: Run to confirm they fail**

Run: `go test ./server/internal/auth/...`
Expected: compile errors (`undefined: auth.NewSessions`, missing `authtest`).

- [ ] **Step 3: Implement `sessions.go`**

`server/internal/auth/sessions.go`:
```go
// Package auth owns login state: the encrypted session cookie shared with the
// legacy views, the route guards, password reset tokens and the login and
// password endpoints.
package auth

import (
	"context"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

// userKey is the session value holding the logged-in user.User. The legacy
// views read and write the same key, so its name must not change.
const userKey = "user"

func init() {
	gob.Register(user.User{})
}

// Config configures the session cookie. Keys are hex encoded; an empty or
// invalid key falls back to a random one (sessions then reset on restart,
// exactly as today).
type Config struct {
	CookieName        string
	AuthenticationKey string
	EncryptionKey     string
	Secure            bool
}

// UserGetter loads a user from the database (satisfied by *user.Store).
type UserGetter interface {
	GetUser(ctx context.Context, u user.User) (user.User, error)
}

// Sessions reads and writes the session cookie.
type Sessions struct {
	store *sessions.CookieStore
	name  string
	users UserGetter
}

func NewSessions(conf Config, users UserGetter) *Sessions {
	store := sessions.NewCookieStore(
		decodeKey(conf.AuthenticationKey, 64, "authentication"),
		decodeKey(conf.EncryptionKey, 32, "encryption"),
	)
	store.Options = &sessions.Options{
		MaxAge:   60 * 60 * 24,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   conf.Secure,
	}
	name := conf.CookieName
	if name == "" {
		name = "session"
	}
	return &Sessions{store: store, name: name, users: users}
}

func decodeKey(hexKey string, size int, what string) []byte {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to decode %s key: %+v", what, err))
	}
	if len(key) == 0 {
		key = securecookie.GenerateRandomKey(size)
	}
	return key
}

// CookieStore exposes the underlying store for the legacy views.
func (s *Sessions) CookieStore() *sessions.CookieStore { return s.store }

// Name is the session cookie name.
func (s *Sessions) Name() string { return s.name }

// User returns the user recorded in the session cookie, without touching the
// database. ok is false when nobody is logged in.
func (s *Sessions) User(r *http.Request) (user.User, bool) {
	sess, err := s.store.Get(r, s.name)
	if err != nil {
		return user.User{}, false
	}
	u, ok := sess.Values[userKey].(user.User)
	if !ok || !u.Authenticated || u.ID <= 0 {
		return user.User{}, false
	}
	return u, true
}

// Login records u in the session. remember extends the cookie to 31 days.
func (s *Sessions) Login(w http.ResponseWriter, r *http.Request, u user.User, remember bool) error {
	// A decode error (e.g. rotated keys) still yields a fresh session to overwrite.
	sess, _ := s.store.Get(r, s.name) //nolint:errcheck // see comment above
	u.Authenticated = true
	u.Password, u.Hash, u.Salt = null.String{}, null.String{}, null.String{}
	sess.Values[userKey] = u
	if remember {
		sess.Options.MaxAge = 86400 * 31
	}
	return sess.Save(r, w)
}

// Logout clears the session and expires the cookie.
func (s *Sessions) Logout(w http.ResponseWriter, r *http.Request) error {
	sess, _ := s.store.Get(r, s.name) //nolint:errcheck // a bad cookie is overwritten anyway
	sess.Values[userKey] = user.User{}
	sess.Options.MaxAge = -1
	return sess.Save(r, w)
}
```

- [ ] **Step 4: Implement `guards.go`**

`server/internal/auth/guards.go`:
```go
package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

const contextKey = "auth.user"

// Current returns the user loaded by a guard earlier in the chain.
func Current(c echo.Context) (user.User, bool) {
	u, ok := c.Get(contextKey).(user.User)
	return u, ok
}

// RequireLogin rejects anonymous requests with 401. It reloads the user from
// the database so role changes and deletions take effect immediately.
func (s *Sessions) RequireLogin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if _, err := s.load(c); err != nil {
			return err
		}
		return next(c)
	}
}

// RequireRole rejects anonymous requests with 401 and users whose role fails
// allowed with 403.
func (s *Sessions) RequireRole(allowed func(role.Role) bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			u, err := s.load(c)
			if err != nil {
				return err
			}
			if !allowed(u.Role) {
				return echo.NewHTTPError(http.StatusForbidden, "you are not authorised for accessing this")
			}
			return next(c)
		}
	}
}

// Guards returns the guard set handlers attach to routes.
func (s *Sessions) Guards() web.Guards {
	return web.Guards{
		Login:               s.RequireLogin,
		Editor:              s.RequireRole(role.Role.CanEdit),
		NotManager:          s.RequireRole(role.Role.CanManageGallery),
		ClubSecretaryHigher: s.RequireRole(role.Role.IsClubSecretaryHigher),
	}
}

func (s *Sessions) load(c echo.Context) (user.User, error) {
	if u, ok := Current(c); ok {
		return u, nil
	}
	u, ok := s.User(c.Request())
	if !ok {
		return user.User{}, echo.NewHTTPError(http.StatusUnauthorized, "login required")
	}
	fresh, err := s.users.GetUser(c.Request().Context(), u)
	if err != nil || fresh.ID != u.ID {
		return user.User{}, echo.NewHTTPError(http.StatusUnauthorized, "login required").SetInternal(err)
	}
	fresh.Authenticated = true
	c.Set(contextKey, fresh)
	return fresh, nil
}
```
The `fresh.ID != u.ID` check matters because `user.GetUser` matches on `email OR id`. A user whose email was reassigned must not inherit another account.

- [ ] **Step 5: Implement `types.go` and `handlers.go`**

`server/internal/auth/types.go`:
```go
package auth

import (
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

// CurrentUser is the logged-in user as the API exposes it.
type CurrentUser struct {
	ID          int         `json:"id"`
	Name        string      `json:"name"`
	Email       string      `json:"email"`
	Phone       string      `json:"phone,omitempty"`
	Role        string      `json:"role"`
	TeamID      int         `json:"teamId,omitempty"`
	ImageURL    string      `json:"imageUrl,omitempty"`
	Permissions Permissions `json:"permissions"`
}

// Permissions tells the client which admin controls to show. The server
// still enforces every rule with guards.
type Permissions struct {
	CanEdit          bool `json:"canEdit"`
	CanManageGallery bool `json:"canManageGallery"`
	CanManageUsers   bool `json:"canManageUsers"`
}

// NewCurrentUser projects u for the API; it never includes password data.
func NewCurrentUser(u user.User, files *upload.Files) CurrentUser {
	return CurrentUser{
		ID:       u.ID,
		Name:     u.Name,
		Email:    u.Email,
		Phone:    u.Phone.String,
		Role:     u.Role.String(),
		TeamID:   u.TeamID,
		ImageURL: files.URL(u.FileName.String),
		Permissions: Permissions{
			CanEdit:          u.Role.CanEdit(),
			CanManageGallery: u.Role.CanManageGallery(),
			CanManageUsers:   u.Role.IsClubSecretaryHigher(),
		},
	}
}
```

`server/internal/auth/handlers.go`:
```go
package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves the /auth endpoints.
type Handlers struct {
	sessions *Sessions
	files    *upload.Files
}

func NewHandlers(sessions *Sessions, files *upload.Files) *Handlers {
	return &Handlers{sessions: sessions, files: files}
}

// Register mounts the /auth routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/auth/me", h.me, guards.Login)
}

// me returns the logged-in user.
//
//	@Summary	Current user
//	@Tags		auth
//	@Produce	json
//	@Success	200	{object}	auth.CurrentUser
//	@Failure	401	{object}	web.ErrorResponse
//	@Router		/auth/me [get]
func (h *Handlers) me(c echo.Context) error {
	u, _ := Current(c)
	return c.JSON(http.StatusOK, NewCurrentUser(u, h.files))
}
```

- [ ] **Step 6: Implement `authtest`**

`server/internal/auth/authtest/authtest.go`:
```go
// Package authtest builds sessions and cookies for handler tests.
package authtest

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

// Personas used across handler tests.
var (
	Webmaster    = user.User{ID: 1, Name: "Web Master", Email: "webmaster@example.test", Role: role.Webmaster}
	Manager      = user.User{ID: 2, Name: "Team Manager", Email: "manager@example.test", Role: role.Manager, TeamID: 1}
	Photographer = user.User{ID: 3, Name: "Photo Grapher", Email: "photo@example.test", Role: role.Photographer}
	Treasurer    = user.User{ID: 4, Name: "Trea Surer", Email: "treasurer@example.test", Role: role.Treasurer}
)

// Users is an in-memory auth.UserGetter.
type Users struct {
	mu   sync.Mutex
	byID map[int]user.User
}

func NewUsers(users ...user.User) *Users {
	u := &Users{byID: map[int]user.User{}}
	for _, x := range users {
		u.byID[x.ID] = x
	}
	return u
}

func (u *Users) GetUser(_ context.Context, p user.User) (user.User, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if x, ok := u.byID[p.ID]; ok {
		return x, nil
	}
	return user.User{}, fmt.Errorf("failed to get user: %w", sql.ErrNoRows)
}

// New returns sessions with fixed keys backed by the given users.
func New(users ...user.User) *auth.Sessions {
	return auth.NewSessions(auth.Config{
		CookieName:        "session",
		AuthenticationKey: strings.Repeat("ab", 64),
		EncryptionKey:     strings.Repeat("cd", 32),
	}, NewUsers(users...))
}

// Everyone returns sessions that know all four personas.
func Everyone() *auth.Sessions {
	return New(Webmaster, Manager, Photographer, Treasurer)
}

// Cookie logs u in and returns the resulting session cookie.
func Cookie(t *testing.T, s *auth.Sessions, u user.User) *http.Cookie {
	t.Helper()
	rec := httptest.NewRecorder()
	require.NoError(t, s.Login(rec, httptest.NewRequest(http.MethodPost, "/", nil), u, false))
	for _, c := range rec.Result().Cookies() {
		if c.Name == s.Name() {
			return c
		}
	}
	t.Fatal("no session cookie written")
	return nil
}

// Files returns an upload.Files over an in-memory store.
func Files() *upload.Files {
	return upload.New(uploadtest.New())
}
```

- [ ] **Step 7: Run tests**

Run: `go test ./server/internal/auth/...`
Expected: PASS.

- [ ] **Step 8: Lint and commit**

```bash
go tool golangci-lint run ./server/internal/auth/...
git add server/internal/auth
git commit -m "Add auth sessions, route guards and GET /auth/me

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---
### Task 6: Extract visitor counting and reset tokens out of `views`

These are standalone packages here. Task 7 switches the legacy views over to them.

**Files:**
- Create: `server/internal/visitors/visitors.go`, `server/internal/visitors/visitors_test.go`
- Create: `server/internal/auth/tokens.go`, `server/internal/auth/tokens_test.go`

**Interfaces:**
- Produces (`visitors`):
  - `New(store SettingStore, interval time.Duration, cookieDomain string) *Counter`
  - Methods: `Start(ctx)` (seeds the count from the DB and starts the flusher), `Stop()`, `Record(visitorID string)`, `Count() int`, `Flush(ctx)`, `Middleware(next echo.HandlerFunc) echo.HandlerFunc`
  - `type SettingStore interface{GetSetting(ctx, string) (setting.Setting, error); IncrementSetting(ctx, string, int) (setting.Setting, error)}`
- Produces (`auth`):
  - `type RedisConfig struct{Addresses []string; MasterName, Username, Password string; DB int; TLS bool; KeyPrefix string}`
  - `NewTokens(RedisConfig) *Tokens`
  - Methods: `Set(ctx, token string, userID int, ttl time.Duration) error`, `Get(ctx, token) (int, bool)`, `Delete(ctx, token)`, `Close()`

- [ ] **Step 1: Write the failing tests**

`server/internal/visitors/visitors_test.go`:
```go
package visitors_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/setting"
	"github.com/COMTOP1/AFC-GO/server/internal/visitors"
)

type fakeSettings struct {
	mu    sync.Mutex
	value int
	incs  []int
}

func (f *fakeSettings) GetSetting(_ context.Context, id string) (setting.Setting, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if id != "visitorCount" {
		return setting.Setting{}, errors.New("unexpected id " + id)
	}
	return setting.Setting{ID: id, SettingText: strconv.Itoa(f.value)}, nil
}

func (f *fakeSettings) IncrementSetting(_ context.Context, id string, delta int) (setting.Setting, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.incs = append(f.incs, delta)
	f.value += delta
	return setting.Setting{ID: id, SettingText: strconv.Itoa(f.value)}, nil
}

func TestRecordCountsEachVisitorOnceAndFlushes(t *testing.T) {
	store := &fakeSettings{value: 40}
	c := visitors.New(store, time.Hour, "afcaldermaston.co.uk")
	c.Flush(context.Background()) // seeds from the DB: nothing pending
	assert.Equal(t, 40, c.Count())

	c.Record("1.1.1.1")
	c.Record("1.1.1.1")
	c.Record("2.2.2.2")
	c.Flush(context.Background())

	assert.Equal(t, []int{2}, store.incs)
	assert.Equal(t, 42, c.Count())
}

func TestFlushPicksUpOtherInstances(t *testing.T) {
	store := &fakeSettings{value: 10}
	c := visitors.New(store, time.Hour, "afcaldermaston.co.uk")
	c.Flush(context.Background())
	store.value = 15 // another instance flushed
	c.Flush(context.Background())
	assert.Equal(t, 15, c.Count())
	assert.Empty(t, store.incs)
}

func TestMiddlewareSetsCookieAndRecordsOnce(t *testing.T) {
	store := &fakeSettings{}
	c := visitors.New(store, time.Hour, "afcaldermaston.co.uk")
	e := echo.New()
	e.Use(c.Middleware)
	e.GET("/", func(ctx echo.Context) error { return ctx.NoContent(http.StatusOK) })

	first := httptest.NewRecorder()
	e.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/", nil))
	cookies := first.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "afc_aldermaston_visited", cookies[0].Name)

	again := httptest.NewRequest(http.MethodGet, "/", nil)
	again.AddCookie(cookies[0])
	e.ServeHTTP(httptest.NewRecorder(), again)

	c.Flush(context.Background())
	assert.Equal(t, []int{1}, store.incs)
}
```

`server/internal/auth/tokens_test.go`:
```go
package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth"
)

func TestTokensInProcess(t *testing.T) {
	ctx := context.Background()
	tokens := auth.NewTokens(auth.RedisConfig{})
	t.Cleanup(tokens.Close)

	require.NoError(t, tokens.Set(ctx, "abc", 7, time.Hour))
	id, ok := tokens.Get(ctx, "abc")
	assert.True(t, ok)
	assert.Equal(t, 7, id)

	tokens.Delete(ctx, "abc")
	_, ok = tokens.Get(ctx, "abc")
	assert.False(t, ok)

	_, ok = tokens.Get(ctx, "never-issued")
	assert.False(t, ok)
}

func TestTokensExpire(t *testing.T) {
	ctx := context.Background()
	tokens := auth.NewTokens(auth.RedisConfig{})
	require.NoError(t, tokens.Set(ctx, "short", 1, 10*time.Millisecond))
	time.Sleep(30 * time.Millisecond)
	_, ok := tokens.Get(ctx, "short")
	assert.False(t, ok)
}
```

- [ ] **Step 2: Run to confirm they fail**

Run: `go test ./server/internal/visitors/... ./server/internal/auth/...`
Expected: compile errors (`undefined: visitors.New`, `undefined: auth.NewTokens`).

- [ ] **Step 3: Implement `visitors`**

`server/internal/visitors/visitors.go` (logic moved from `views/views.go` and `views/middleware.go`):
```go
// Package visitors counts unique daily visitors and persists the running
// total in the settings table, shared across instances.
package visitors

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"go.opentelemetry.io/otel"

	"github.com/COMTOP1/AFC-GO/server/internal/setting"
)

var tracer = otel.Tracer("github.com/COMTOP1/AFC-GO/server/internal/visitors")

const (
	settingID  = "visitorCount"
	cookieName = "afc_aldermaston_visited"
)

// SettingStore persists the running total (satisfied by *setting.Store).
type SettingStore interface {
	GetSetting(ctx context.Context, settingID string) (setting.Setting, error)
	IncrementSetting(ctx context.Context, settingID string, delta int) (setting.Setting, error)
}

// Counter records visits locally and flushes them to the database.
type Counter struct {
	store        SettingStore
	seen         *cache.Cache
	interval     time.Duration
	cookieDomain string

	mu      sync.Mutex
	pending int
	total   int

	stop     chan struct{}
	stopOnce sync.Once
}

func New(store SettingStore, interval time.Duration, cookieDomain string) *Counter {
	return &Counter{
		store:        store,
		seen:         cache.New(time.Hour, time.Hour),
		interval:     interval,
		cookieDomain: cookieDomain,
		stop:         make(chan struct{}),
	}
}

// Start seeds the total from the database and flushes every interval until Stop.
func (c *Counter) Start(ctx context.Context) {
	c.Flush(ctx)
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.Flush(context.Background())
			case <-c.stop:
				return
			}
		}
	}()
}

// Stop ends the flusher. It is safe to call more than once.
func (c *Counter) Stop() {
	c.stopOnce.Do(func() { close(c.stop) })
}

// Record counts visitorID once per hour-long cache window.
func (c *Counter) Record(visitorID string) {
	if _, found := c.seen.Get(visitorID); found {
		return
	}
	c.seen.Set(visitorID, true, cache.DefaultExpiration)
	c.mu.Lock()
	c.pending++
	c.mu.Unlock()
}

// Count is the last known total across all instances.
func (c *Counter) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.total
}

// Flush adds pending visits to the database total, or, with nothing pending,
// re-reads the total so this instance sees other instances' visits.
func (c *Counter) Flush(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "visitors.Flush")
	defer span.End()

	c.mu.Lock()
	delta := c.pending
	c.pending = 0
	c.mu.Unlock()

	var (
		s   setting.Setting
		err error
	)
	if delta == 0 {
		s, err = c.store.GetSetting(ctx, settingID)
		if err != nil {
			return // not created yet; it will be on the first increment
		}
	} else {
		s, err = c.store.IncrementSetting(ctx, settingID, delta)
		if err != nil {
			span.RecordError(err)
			slog.InfoContext(ctx, fmt.Sprintf("failed to increment visitor count: %+v", err))
			c.mu.Lock()
			c.pending += delta // retry on the next flush
			c.mu.Unlock()
			return
		}
	}

	total, err := strconv.Atoi(s.SettingText)
	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, fmt.Sprintf("failed to parse visitor count: %+v", err))
		return
	}
	c.mu.Lock()
	c.total = total
	c.mu.Unlock()
}

// Middleware counts a visit the first time a browser is seen in 24 hours.
func (c *Counter) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if _, err := ctx.Cookie(cookieName); errors.Is(err, http.ErrNoCookie) {
			ctx.SetCookie(&http.Cookie{
				Name:     cookieName,
				Value:    "visited",
				Expires:  time.Now().Add(24 * time.Hour),
				Domain:   c.cookieDomain,
				Path:     "/",
				SameSite: http.SameSiteStrictMode,
				Secure:   true,
				HttpOnly: true,
			})
			c.Record(ctx.RealIP())
		}
		return next(ctx)
	}
}
```
Behaviour change: legacy dropped the pending delta when an increment failed. It is now retried on the next flush.

- [ ] **Step 4: Implement `auth/tokens.go`**

`server/internal/auth/tokens.go` (moved from `views/views.go`: `SetResetToken`, `GetResetToken`, `DeleteResetToken` and the Redis client setup):
```go
package auth

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("github.com/COMTOP1/AFC-GO/server/internal/auth")

// RedisConfig configures the optional Redis/Valkey used to share reset
// tokens across instances. With no Addresses an in-process cache is used,
// which only works for a single instance.
type RedisConfig struct {
	Addresses  []string
	MasterName string
	Username   string
	Password   string
	DB         int
	TLS        bool
	// KeyPrefix namespaces keys per environment, e.g. "afc:prod:". Defaults to "afc:".
	KeyPrefix string
}

// Tokens stores one-time password reset tokens.
type Tokens struct {
	redis  redis.UniversalClient
	cache  *cache.Cache
	prefix string
}

func NewTokens(conf RedisConfig) *Tokens {
	t := &Tokens{cache: cache.New(time.Hour, time.Hour), prefix: conf.KeyPrefix}
	if t.prefix == "" {
		t.prefix = "afc:"
	}
	if len(conf.Addresses) > 0 {
		opts := &redis.UniversalOptions{
			Addrs:      conf.Addresses,
			MasterName: conf.MasterName,
			Username:   conf.Username,
			Password:   conf.Password,
			DB:         conf.DB,
		}
		if conf.TLS {
			opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}
		t.redis = redis.NewUniversalClient(opts)
		slog.Info("using redis/valkey for shared cache")
	} else {
		slog.Info("no redis addresses configured, using in-process cache")
	}
	return t
}

func (t *Tokens) key(token string) string { return t.prefix + "reset-token:" + token }

// Set maps token to userID for ttl.
func (t *Tokens) Set(ctx context.Context, token string, userID int, ttl time.Duration) error {
	ctx, span := tracer.Start(ctx, "auth.Tokens.Set")
	defer span.End()
	if t.redis != nil {
		if err := t.redis.Set(ctx, t.key(token), userID, ttl).Err(); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to store reset token: %w", err)
		}
		return nil
	}
	t.cache.Set(token, userID, ttl)
	return nil
}

// Get returns the user ID a token was issued for.
func (t *Tokens) Get(ctx context.Context, token string) (int, bool) {
	ctx, span := tracer.Start(ctx, "auth.Tokens.Get")
	defer span.End()
	if t.redis != nil {
		id, err := t.redis.Get(ctx, t.key(token)).Int()
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				span.RecordError(err)
			}
			return 0, false
		}
		return id, true
	}
	v, found := t.cache.Get(token)
	if !found {
		return 0, false
	}
	id, ok := v.(int)
	return id, ok
}

// Delete invalidates a token after use.
func (t *Tokens) Delete(ctx context.Context, token string) {
	ctx, span := tracer.Start(ctx, "auth.Tokens.Delete")
	defer span.End()
	if t.redis != nil {
		if err := t.redis.Del(ctx, t.key(token)).Err(); err != nil {
			span.RecordError(err)
			slog.ErrorContext(ctx, fmt.Sprintf("failed to delete reset token from redis: %+v", err))
		}
		return
	}
	t.cache.Delete(token)
}

// Close releases the Redis client, if any.
func (t *Tokens) Close() {
	if t.redis == nil {
		return
	}
	if err := t.redis.Close(); err != nil {
		slog.Error(fmt.Sprintf("failed to close redis client: %+v", err))
	}
}
```
The Redis key format (`<prefix>reset-token:<token>`) is unchanged, so tokens issued before the deploy stay valid.

- [ ] **Step 5: Run tests**

Run: `go test ./server/internal/visitors/... ./server/internal/auth/...`
Expected: PASS.

- [ ] **Step 6: Lint and commit**

```bash
go tool golangci-lint run ./server/internal/visitors/... ./server/internal/auth/...
git add server/internal/visitors server/internal/auth
git commit -m "Extract visitor counter and reset token store into their own packages

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---
### Task 7: `app` wiring, legacy views on shared deps, swagger

After this task `main.go` only reads the environment. `app.Build` constructs everything. The legacy views get their session store, visitor counter and reset tokens injected instead of building them. Swagger docs are generated.

**Files:**
- Create: `server/internal/app/app.go`, `server/internal/app/app_test.go`, `server/internal/app/helpers_test.go`
- Move: `server/cmd/afc/router_smoke_test.go` → `server/internal/app/smoke_test.go` (rewritten below)
- Create: `server/internal/legacy/routes.go` (body of the old `loadRoutes`)
- Delete: `server/cmd/afc/router.go`
- Create: `server/internal/auth/passwords.go`
- Modify: `server/cmd/afc/main.go`, `server/internal/legacy/views/views.go`, `server/internal/legacy/views/middleware.go`, `server/internal/legacy/templates/template.go`
- Create: `server/internal/docs/generate.go`; generated `server/internal/docs/{docs.go,swagger.json,swagger.yaml}`
- Modify: `go.mod`/`go.sum` (swag tool)

**Interfaces:**
- Consumes:
  - `auth.NewSessions`, `auth.Config`, `auth.NewTokens`, `auth.RedisConfig`, `auth.NewHandlers` (Tasks 5–6)
  - `visitors.New` (Task 6)
  - `web.NewAPI`, `web.ErrorHandler` (Task 4)
  - `upload.New`, `upload.Storage` (Task 3)
- Produces (`app`):
  - `type Config` (fields below)
  - `type Stores struct{Affiliation *affiliation.Store; Document; Image; News; Player; Programme; Setting; Sponsor; Team; User; WhatsOn}` and `NewStores(*sqlx.DB) Stores`
  - `New(Config) *App`, which connects to the DB and S3
  - `Build(Config, Stores, upload.Storage, *mail.MailerInit) *App`, which connects nothing
  - `(*App).Start() error`, `(*App).Stop()` and the `(*App).Echo` field
  - **Every later task adds its handlers inside `Build`, after the `auth.NewHandlers` line.**
- Produces (`auth`): `type PasswordConfig struct{Iterations, ScryptWorkFactor, ScryptBlockSize, ScryptParallelismFactor, KeyLength int}`.
- Produces (`legacy`): `Mount(e *echo.Echo, v *views.Views)`.
- Produces (`views`):
  - `type Deps struct{Conf *Config; Sessions *sessions.CookieStore; Storage upload.Storage; Mailer *mail.MailerInit; Visitors *visitors.Counter; Tokens *auth.Tokens; <every *Store>}`
  - `New(Deps) *Views`
  - `type SecurityConfig = auth.PasswordConfig`
- Test helpers (`app_test`):
  - `testConfig() app.Config`
  - `bareApp(t) *app.App` (zero stores, in-memory storage)
  - `guardedGets []string`, the GET routes that must 401 when logged out. Later tasks append to it.

- [ ] **Step 1: Add the swag tool**

```bash
go get -tool github.com/swaggo/swag/cmd/swag@v1.16.6
```

- [ ] **Step 2: Write the failing tests**

`server/internal/app/helpers_test.go`:
```go
package app_test

import (
	"testing"

	"github.com/COMTOP1/AFC-GO/server/internal/app"
	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

func testConfig() app.Config {
	return app.Config{
		DomainName: "localhost",
		Session:    auth.Config{CookieName: "session"},
		Passwords: auth.PasswordConfig{
			Iterations: 1, ScryptWorkFactor: 2, ScryptBlockSize: 1, ScryptParallelismFactor: 1, KeyLength: 32,
		},
	}
}

// bareApp wires every route with no database behind it. Only requests that
// guards reject before reaching a store are safe to send.
func bareApp(t *testing.T) *app.App {
	t.Helper()
	a := app.Build(testConfig(), app.Stores{}, uploadtest.New(), mail.NewMailer(mail.Config{}))
	t.Cleanup(a.Stop)
	return a
}

// guardedGets are GET routes that must reject anonymous users. Append to this
// list whenever a task adds a guarded GET route.
var guardedGets = []string{
	"/api/v1/auth/me",
}
```

`server/internal/app/app_test.go`:
```go
package app_test

import (
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

// publicWrites are the only unsafe API routes anonymous users may call.
var publicWrites = map[string]bool{
	"POST /api/v1/auth/login":        true,
	"POST /api/v1/auth/reset/:token": true,
}

var pathParam = regexp.MustCompile(`:[A-Za-z]+`)

// TestUnsafeAPIRoutesRequireLogin walks every registered route, so each new
// write endpoint is covered automatically.
func TestUnsafeAPIRoutesRequireLogin(t *testing.T) {
	a := bareApp(t)
	checked := 0
	for _, r := range a.Echo.Routes() {
		if !strings.HasPrefix(r.Path, "/api/v1/") {
			continue
		}
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		default:
			continue
		}
		key := r.Method + " " + r.Path
		if publicWrites[key] {
			continue
		}
		path := pathParam.ReplaceAllString(r.Path, "1")
		rec := apitest.New(a.Echo).JSON(t, r.Method, path, nil)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, key)
		checked++
	}
	t.Logf("checked %d unsafe API routes", checked)
}

func TestGuardedGetRoutesRequireLogin(t *testing.T) {
	a := bareApp(t)
	for _, path := range guardedGets {
		rec := apitest.New(a.Echo).Get(t, path)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, path)
	}
}

func TestSwaggerDocServed(t *testing.T) {
	a := bareApp(t)
	rec := apitest.New(a.Echo).Get(t, "/api/v1/swagger/doc.json")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"/auth/me"`)
}

func TestLegacyAndAPINotFoundDiffer(t *testing.T) {
	a := bareApp(t)
	api := apitest.New(a.Echo).Get(t, "/api/v1/nope")
	assert.Equal(t, http.StatusNotFound, api.Code)
	assert.Equal(t, 404, apitest.ErrorOf(t, api).Code)
}
```

`server/internal/app/smoke_test.go` (replaces `server/cmd/afc/router_smoke_test.go`; `git rm` the old file):
```go
package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/app"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

// TestLegacyPagesSmoke guards the template site through the restructure:
// every public page renders and every guarded page redirects when logged out.
func TestLegacyPagesSmoke(t *testing.T) {
	db, _ := testdb.Open(t)
	a := app.Build(testConfig(), app.NewStores(db), uploadtest.New(), mail.NewMailer(mail.Config{}))
	t.Cleanup(a.Stop)

	cases := []struct {
		path string
		want int
	}{
		{"/", http.StatusOK},
		{"/news", http.StatusOK},
		{"/news/1", http.StatusOK},
		{"/whatson", http.StatusOK},
		{"/whatson/1", http.StatusOK},
		{"/teams", http.StatusOK},
		{"/team/1", http.StatusOK},
		{"/sponsors", http.StatusOK},
		{"/gallery", http.StatusOK},
		{"/documents", http.StatusOK},
		{"/programmes", http.StatusOK},
		{"/info", http.StatusOK},
		{"/contact", http.StatusOK},
		{"/public/stylesheet.css", http.StatusOK},
		{"/does-not-exist", http.StatusNotFound},
		{"/players", http.StatusFound},
		{"/account", http.StatusFound},
		{"/users", http.StatusFound},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			a.Echo.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
			assert.Equal(t, tc.want, rec.Code, "body: %.300s", rec.Body.String())
		})
	}
}
```

- [ ] **Step 3: Run to confirm they fail**

Run: `go test ./server/internal/app/...`
Expected: compile errors (package `app` does not exist).

- [ ] **Step 4: Add `auth.PasswordConfig`**

`server/internal/auth/passwords.go`:
```go
package auth

// PasswordConfig holds the password hashing parameters (legacy PBKDF2
// iterations plus scrypt parameters), loaded from the environment.
type PasswordConfig struct {
	Iterations              int
	ScryptWorkFactor        int
	ScryptBlockSize         int
	ScryptParallelismFactor int
	KeyLength               int
}
```

- [ ] **Step 5: Point the legacy views at injected dependencies**

In `server/internal/legacy/views/views.go`:

1. Replace the `Config`, `RedisConfig`, `SMTPConfig` and `SecurityConfig` type declarations with:
```go
	// Config is what the legacy views still need from the environment.
	Config struct {
		DomainName        string
		SessionCookieName string
		Security          SecurityConfig
	}

	// SecurityConfig is kept as an alias so existing v.conf.Security.* call
	// sites don't change.
	SecurityConfig = auth.PasswordConfig
```
2. In the `Views` struct, delete the `cache`, `redis`, `redisPrefix` fields and the whole `// Visitor tracking` block (`count`, `countMutex`, `flushInterval`, `stopChan`). Change `storage *storage.Store` to `storage upload.Storage`, and add:
```go
		tokens   *auth.Tokens
		visitors *visitors.Counter
```
3. Replace `func New(conf *Config, host string, interval time.Duration) *Views { … }` and everything after it in the file (`RecordVisit` … `GetVisitorCount`) with:
```go
	// Deps are the shared services the legacy views use.
	Deps struct {
		Conf     *Config
		Sessions *sessions.CookieStore
		Storage  upload.Storage
		Mailer   *mail.MailerInit
		Visitors *visitors.Counter
		Tokens   *auth.Tokens

		Affiliation *affiliation.Store
		Document    *document.Store
		Image       *image.Store
		News        *news.Store
		Player      *player.Store
		Programme   *programme.Store
		Setting     *setting.Store
		Sponsor     *sponsor.Store
		Team        *team.Store
		User        *user.Store
		WhatsOn     *whatson.Store
	}
)

// New builds the legacy views from shared dependencies. It connects to nothing.
func New(d Deps) *Views {
	// InternalContext carries flash messages in the session; user.User is
	// registered by the auth package.
	gob.Register(InternalContext{})
	return &Views{
		affiliation: d.Affiliation,
		conf:        d.Conf,
		cookie:      d.Sessions,
		document:    d.Document,
		image:       d.Image,
		mailer:      d.Mailer,
		news:        d.News,
		player:      d.Player,
		programme:   d.Programme,
		setting:     d.Setting,
		sponsor:     d.Sponsor,
		storage:     d.Storage,
		team:        d.Team,
		template:    templates.NewTemplate(d.Team, d.Storage),
		tokens:      d.Tokens,
		user:        d.User,
		visitors:    d.Visitors,
		whatsOn:     d.WhatsOn,
	}
}

// GetVisitorCount is the site-wide visitor total shown in the footer.
func (v *Views) GetVisitorCount() int { return v.visitors.Count() }

// SetResetToken, GetResetToken and DeleteResetToken keep the existing view
// call sites unchanged while the tokens live in auth.
func (v *Views) SetResetToken(ctx context.Context, token string, userID int, ttl time.Duration) error {
	return v.tokens.Set(ctx, token, userID, ttl)
}

func (v *Views) GetResetToken(ctx context.Context, token string) (int, bool) {
	return v.tokens.Get(ctx, token)
}

func (v *Views) DeleteResetToken(ctx context.Context, token string) {
	v.tokens.Delete(ctx, token)
}
```
   The closing `)` of the `type (…)` block moves to just after `Deps`; `TemplateHelper` stays inside the block.
4. Fix the imports: remove `crypto/tls`, `encoding/hex`, `errors`, `fmt`, `log/slog`, `strconv`, `sync`, `securecookie`, `go-cache`, `go-redis`, `infrastructure/db`, `infrastructure/storage` and `otel`; add `auth`, `upload` and `visitors`. Keep the `var tracer` line only if other files in the package use `tracer`. They do, so keep `go.opentelemetry.io/otel` and the `tracer` var. Also delete `const visitorCount`. Let `goimports -w server/internal/legacy/views` settle the rest.

In `server/internal/legacy/views/middleware.go`, delete `VisitorTrackingMiddleware` (and the now-unused `time` import); the app uses `visitors.Counter.Middleware`.

In `server/internal/legacy/templates/template.go`, replace the `Storage *storage.Store` field type and the `NewTemplate` signature:
```go
	// URLer builds public URLs for stored files (satisfied by upload.Storage).
	URLer interface {
		PublicURL(key string) string
	}
```
```go
		Storage URLer
```
```go
func NewTemplate(team *team.Store, storage URLer) *Templater {
```
Then drop the now-unused `infrastructure/storage` import.

- [ ] **Step 6: Move the legacy routes**

Create `server/internal/legacy/routes.go` holding the body of the old `(*Router).loadRoutes`, with `r.router` → `e` and `r.views` → `v`. Leave out three things: `r.router.HTTPErrorHandler = …`, `r.router.Use(r.views.VisitorTrackingMiddleware)` and the `base.GET("api/health", …)` block (`web.NewAPI` serves `/api/health` now):
```go
package legacy

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/legacy/views"
)

// Mount registers the server-rendered template routes and /public assets.
func Mount(e *echo.Echo, v *views.Views) {
	e.RouteNotFound("/*", v.Error404)

	assetHandler := http.FileServer(http.FS(echo.MustSubFS(Public, "public")))
	e.GET("/public/*", echo.WrapHandler(http.StripPrefix("/public/", assetHandler)))

	validMethods := []string{http.MethodGet, http.MethodPost}

	base := e.Group("/")

	// base is the functions that don't require being logged in
	base.GET("", v.HomeFunc)

	// … every remaining line of the old loadRoutes, from
	//   affiliation := base.Group("affiliation", …)
	// through
	//   base.Match(validMethods, "logout", v.LogoutFunc, v.RequiresLogin)
	// copied verbatim with r.views → v …
}
```
Copy the lines verbatim: no route may be added, removed or re-ordered. After copying, `diff <(git show HEAD:server/cmd/afc/router.go | sed -n '/^func (r \*Router) loadRoutes/,/^}/p' | sed 's/r\.views/v/g; s/r\.router/e/g') <(sed -n '/^func Mount/,/^}/p' server/internal/legacy/routes.go)` must show only the three removed blocks and the signature line.

Then `git rm server/cmd/afc/router.go`.

- [ ] **Step 7: Write `app.go`**

`server/internal/app/app.go`:
```go
// Package app wires the stores, services, legacy views and API handlers into
// one Echo server.
package app

import (
	"context"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho" //nolint:staticcheck // still functional; see main branch note

	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	_ "github.com/COMTOP1/AFC-GO/server/internal/docs" // registers the swagger spec
	"github.com/COMTOP1/AFC-GO/server/internal/document"
	"github.com/COMTOP1/AFC-GO/server/internal/image"
	infradb "github.com/COMTOP1/AFC-GO/server/internal/infrastructure/db"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/storage"
	"github.com/COMTOP1/AFC-GO/server/internal/legacy"
	"github.com/COMTOP1/AFC-GO/server/internal/legacy/views"
	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/programme"
	"github.com/COMTOP1/AFC-GO/server/internal/setting"
	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/visitors"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

// Config is everything the server needs, loaded from the environment by main.
type Config struct {
	Address      string
	DomainName   string
	DatabaseURL  string
	DatabaseHost string
	Version      string
	Session      auth.Config
	S3           storage.Config
	Mail         mail.Config
	Passwords    auth.PasswordConfig
	Redis        auth.RedisConfig
}

// Stores are the database repositories.
type Stores struct {
	Affiliation *affiliation.Store
	Document    *document.Store
	Image       *image.Store
	News        *news.Store
	Player      *player.Store
	Programme   *programme.Store
	Setting     *setting.Store
	Sponsor     *sponsor.Store
	Team        *team.Store
	User        *user.Store
	WhatsOn     *whatson.Store
}

// NewStores builds every repository over one connection pool.
func NewStores(db *sqlx.DB) Stores {
	return Stores{
		Affiliation: affiliation.NewAffiliationRepo(db),
		Document:    document.NewDocumentRepo(db),
		Image:       image.NewImageRepo(db),
		News:        news.NewNewsRepo(db),
		Player:      player.NewPlayerRepo(db),
		Programme:   programme.NewProgrammeRepo(db),
		Setting:     setting.NewSettingRepo(db),
		Sponsor:     sponsor.NewSponsorRepo(db),
		Team:        team.NewTeamRepo(db),
		User:        user.NewUserRepo(db),
		WhatsOn:     whatson.NewWhatsOnRepo(db),
	}
}

// App is a wired server.
type App struct {
	Echo     *echo.Echo
	address  string
	visitors *visitors.Counter
	tokens   *auth.Tokens
}

// New connects to Postgres and S3 and builds the server.
func New(conf Config) *App {
	db := infradb.NewStore(conf.DatabaseURL, conf.DatabaseHost)
	objects := storage.NewStore(context.Background(), conf.S3)
	return Build(conf, NewStores(db), objects, mail.NewMailer(conf.Mail))
}

// Build wires everything without connecting to anything, so tests can pass
// fakes. Background work starts in Start.
func Build(conf Config, s Stores, objects upload.Storage, mailer *mail.MailerInit) *App {
	uploads := upload.New(objects)
	sessions := auth.NewSessions(conf.Session, s.User)
	tokens := auth.NewTokens(conf.Redis)
	counter := visitors.New(s.Setting, 30*time.Second, "afcaldermaston.co.uk")

	legacyViews := views.New(views.Deps{
		Conf: &views.Config{
			DomainName:        conf.DomainName,
			SessionCookieName: sessions.Name(),
			Security:          conf.Passwords,
		},
		Sessions:    sessions.CookieStore(),
		Storage:     objects,
		Mailer:      mailer,
		Visitors:    counter,
		Tokens:      tokens,
		Affiliation: s.Affiliation,
		Document:    s.Document,
		Image:       s.Image,
		News:        s.News,
		Player:      s.Player,
		Programme:   s.Programme,
		Setting:     s.Setting,
		Sponsor:     s.Sponsor,
		Team:        s.Team,
		User:        s.User,
		WhatsOn:     s.WhatsOn,
	})

	e := echo.New()
	e.HideBanner = true
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(otelecho.Middleware("afc-go", otelecho.WithSkipper(func(c echo.Context) bool {
		return c.Path() == "/api/health" || c.Path() == "/api/v1/health"
	})))
	e.Use(middleware.BodyLimit("15M"))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
		Skipper: func(c echo.Context) bool {
			// File downloads are redirects to the CDN; don't gzip them.
			return strings.HasPrefix(c.Path(), "/download") || strings.HasPrefix(c.Path(), "/api/v1/files")
		},
	}))
	e.Use(counter.Middleware)
	e.HTTPErrorHandler = web.ErrorHandler(legacyViews.CustomHTTPErrorHandler)

	legacy.Mount(e, legacyViews)

	api := web.NewAPI(e, conf.Session.Secure)
	guards := sessions.Guards()
	auth.NewHandlers(sessions, uploads).Register(api, guards)

	return &App{Echo: e, address: conf.Address, visitors: counter, tokens: tokens}
}

// Start begins background work and serves until the server stops.
func (a *App) Start() error {
	a.visitors.Start(context.Background())
	return a.Echo.Start(a.address)
}

// Stop ends background work.
func (a *App) Stop() {
	a.visitors.Stop()
	a.tokens.Close()
}
```
Copy the exact `//nolint` comment on the `otelecho` import from the old `router.go`.

- [ ] **Step 8: Slim down `main.go`**

In `server/cmd/afc/main.go`:
- Add the swagger general info as the package doc comment:
```go
// Command afc serves the AFC Aldermaston website and its JSON API.
//
//	@title			AFC Aldermaston API
//	@version		1
//	@description	JSON API behind the AFC Aldermaston website.
//	@BasePath		/api/v1
package main
```
- Leave the environment parsing unchanged, from `godotenv.Load` down to `redisTLS`.
- Replace everything from `// Generate config` to the end of `main` with:
```go
	a := app.New(app.Config{
		Address:      address,
		DomainName:   domainName,
		DatabaseURL:  dbConnectionString,
		DatabaseHost: dbHost,
		Version:      Version,
		Session: auth.Config{
			CookieName:        sessionCookieName,
			AuthenticationKey: os.Getenv("AUTHENTICATION_KEY"),
			EncryptionKey:     os.Getenv("ENCRYPTION_KEY"),
			Secure:            !strings.HasPrefix(domainName, "localhost"),
		},
		S3: storage.Config{
			Endpoint:  os.Getenv("S3_ENDPOINT"),
			Region:    s3Region,
			Bucket:    os.Getenv("S3_BUCKET"),
			AccessKey: os.Getenv("S3_ACCESS_KEY"),
			SecretKey: os.Getenv("S3_SECRET_KEY"),
		},
		Mail: mail.Config{
			Host:     os.Getenv("MAIL_HOST"),
			Username: os.Getenv("MAIL_USER"),
			Password: os.Getenv("MAIL_PASS"),
			Port:     mailPort,
		},
		Passwords: auth.PasswordConfig{
			Iterations:              iter,
			ScryptWorkFactor:        sWorkFactor,
			ScryptBlockSize:         sBlockSize,
			ScryptParallelismFactor: sParallelismFactor,
			KeyLength:               keyLen,
		},
		Redis: auth.RedisConfig{
			Addresses:  redisAddresses,
			MasterName: os.Getenv("REDIS_MASTER_NAME"),
			Username:   os.Getenv("REDIS_USERNAME"),
			Password:   os.Getenv("REDIS_PASSWORD"),
			DB:         redisDB,
			TLS:        redisTLS,
			KeyPrefix:  os.Getenv("REDIS_KEY_PREFIX"),
		},
	})

	err = a.Start()
	a.Stop()
	fatal(fmt.Sprintf("The web server couldn't be started!\n\n%s\n\nExiting!", err))
}
```
- Fix the imports: drop `time` (if unused) and `legacy/views`; add `app`, `auth` and `infrastructure/mail`. Keep `storage`, which is still used for `storage.Config`.

- [ ] **Step 9: Generate swagger**

`server/internal/docs/generate.go`:
```go
// Package docs holds the generated OpenAPI description of /api/v1.
// Regenerate with: go generate ./server/internal/docs
package docs

//go:generate go tool swag init --quiet --parseInternal -g main.go -d ../../cmd/afc,.. -o .
```
Run:
```bash
go generate ./server/internal/docs
ls server/internal/docs   # docs.go generate.go swagger.json swagger.yaml
```

- [ ] **Step 10: Run everything**

```bash
gofmt -l server; go build ./... && go vet ./...
go test ./...                      # with AFC_TEST_DB exported, the smoke test runs too
go tool golangci-lint run ./...
```
Expected: all green. `TestUnsafeAPIRoutesRequireLogin` logs `checked 0 unsafe API routes` for now.

- [ ] **Step 11: Run the server once**

```bash
go run ./server/cmd/afc &   sleep 3
curl -s localhost:${ADDRESS##*:}/api/v1/health; echo
curl -s -o /dev/null -w '%{http_code}\n' localhost:${ADDRESS##*:}/
kill %1
```
Expected: `{"status":"ok"}` then `200`. (Needs a local `.env`; skip if unavailable.)

- [ ] **Step 12: Commit**

```bash
git add -A server go.mod go.sum
git commit -m "Wire server through app package; legacy views use injected sessions, visitors, tokens; add swagger

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---
## Phase C: Domains

Every domain task follows the same shape, and Task 8 (news) is the fully worked reference. Later tasks still spell out all their code, because an implementer may read them in isolation. Each domain task:
1. Fixes the store where the plan says to (DB-backed test).
2. Writes service tests on a fake store and implements `service.go` + `types.go`.
3. Writes handler tests and implements `handlers.go`.
4. Points the legacy **write** views at the service. Legacy **read** views keep their store calls, because their templates depend on the legacy formats and are deleted in sub-project 5.
5. Wires the service into `app.Build` and `views.Deps`, regenerates swagger, runs everything, and commits.

### Task 8: News

**Files:**
- Modify: `server/internal/news/store.go` (`addNews` returns ID), `server/internal/news/news.go` (`EditNews` write-through)
- Create: `server/internal/news/store_test.go`, `types.go`, `service.go`, `handlers.go`, `fake_test.go`, `service_test.go`, `handlers_test.go`
- Modify: `server/internal/legacy/views/news.go`, `server/internal/legacy/views/helpers.go`, `server/internal/legacy/views/views.go`, `server/internal/app/app.go`
- Regenerate: `server/internal/docs/*`

**Interfaces:**
- Consumes:
  - `upload.Files`, `upload.File`, `uploadtest` (Task 3)
  - `sanitize.HTML` (Task 3)
  - `svcerr` (Task 3)
  - `web.*`, `apitest` (Task 4)
  - `authtest` (Task 5)
  - `testdb.Open` (Task 1)
- Produces (`news`):
  - Types: `Article{ID int; Title, Content string; Date time.Time; ImageURL string}`, `CreateInput{Title, Content string}` and `UpdateInput{Title, Content *string; RemoveImage bool}`.
  - `NewService(store, *upload.Files) *Service`, with methods:
    - `List(ctx) ([]Article, error)`
    - `Latest(ctx) (Article, bool, error)`
    - `Get(ctx, id) (Article, error)`
    - `Create(ctx, CreateInput, *upload.File) (Article, error)`
    - `Update(ctx, id, UpdateInput, *upload.File) (Article, error)`
    - `Delete(ctx, id) (Article, error)`
  - `NewHandlers(*Service) *Handlers` and `(*Handlers).Register(*echo.Group, web.Guards)`.
- Produces (legacy `views`): helpers `legacyUpload(c, field) (*upload.File, error)`, `(*Views).flash(c, *Context, msg)` and `formYes(c, field) bool`; `Deps.NewsService *news.Service`.

- [ ] **Step 1: Write the failing store tests**

`server/internal/news/store_test.go`:
```go
package news_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestAddNewsReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	store := news.NewNewsRepo(db)
	ctx := context.Background()

	added, err := store.AddNews(ctx, news.News{Title: "Fresh", Content: null.StringFrom("<p>x</p>")})
	require.NoError(t, err)
	require.Positive(t, added.ID)

	got, err := store.GetNewsArticle(ctx, news.News{ID: added.ID})
	require.NoError(t, err)
	assert.Equal(t, "Fresh", got.Title)
}

// TestEditNewsClearsFileName pins Review Focus #3 at the SQL level: before
// this change EditNews ignored an invalid FileName, so "remove image" left
// the dead key in the row.
func TestEditNewsClearsFileName(t *testing.T) {
	db, _ := testdb.Open(t)
	store := news.NewNewsRepo(db)
	ctx := context.Background()

	n, err := store.GetNewsArticle(ctx, news.News{ID: 1})
	require.NoError(t, err)
	require.True(t, n.FileName.Valid, "fixture row 1 has an image")

	n.FileName = null.String{}
	_, err = store.EditNews(ctx, n)
	require.NoError(t, err)

	got, err := store.GetNewsArticle(ctx, news.News{ID: 1})
	require.NoError(t, err)
	assert.False(t, got.FileName.Valid)
}
```

- [ ] **Step 2: Run to confirm they fail (with `AFC_TEST_DB` set)**

Run: `go test -run 'TestAddNewsReturnsID|TestEditNewsClearsFileName' ./server/internal/news/`
Expected: FAIL. `added.ID` is 0, and `got.FileName.Valid` is still true.

- [ ] **Step 3: Fix the store**

In `server/internal/news/store.go`, `addNews`: add `.Suffix("RETURNING id")` to the builder and replace the `ExecContext` / `RowsAffected` block with:
```go
	err = s.db.GetContext(ctx, &newsParam.ID, sql, args...)
	if err != nil {
		span.RecordError(err)
		return News{}, fmt.Errorf("failed to add news: %w", err)
	}
	return newsParam, nil
```
In `server/internal/news/news.go`, make `EditNews` write-through. Every caller loads the full row first, so merging only stopped fields from being cleared:
```go
// EditNews writes n as given. Callers load the row first and change only the
// fields they mean to, so empty values are deliberate (e.g. removing an image).
func (s *Store) EditNews(ctx context.Context, newsParam News) (News, error) {
	return s.editNews(ctx, newsParam)
}
```
Run the two store tests again. Expected: PASS.

- [ ] **Step 4: Write the fake store and failing service tests**

`server/internal/news/fake_test.go`:
```go
package news_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/COMTOP1/AFC-GO/server/internal/news"
)

type fakeStore struct {
	mu      sync.Mutex
	rows    map[int]news.News
	nextID  int
	editErr error
}

func newFakeStore(rows ...news.News) *fakeStore {
	f := &fakeStore{rows: map[int]news.News{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) GetNews(context.Context) ([]news.News, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]news.News, 0, len(f.rows))
	for _, r := range f.rows {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date.After(out[j].Date) })
	return out, nil
}

func (f *fakeStore) GetNewsLatest(ctx context.Context) (news.News, error) {
	all, _ := f.GetNews(ctx)
	if len(all) == 0 {
		return news.News{}, fmt.Errorf("failed to get news latest: %w", sql.ErrNoRows)
	}
	return all[0], nil
}

func (f *fakeStore) GetNewsArticle(_ context.Context, n news.News) (news.News, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[n.ID]
	if !ok {
		return news.News{}, fmt.Errorf("failed to get news article: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddNews(_ context.Context, n news.News) (news.News, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	n.ID = f.nextID
	n.Date = time.Now()
	f.rows[n.ID] = n
	return n, nil
}

func (f *fakeStore) EditNews(_ context.Context, n news.News) (news.News, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.editErr != nil {
		return news.News{}, f.editErr
	}
	if _, ok := f.rows[n.ID]; !ok {
		return news.News{}, errors.New("no such row")
	}
	f.rows[n.ID] = n
	return n, nil
}

func (f *fakeStore) DeleteNews(_ context.Context, n news.News) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, n.ID)
	return nil
}

func (f *fakeStore) row(id int) news.News {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.rows[id]
}
```

`server/internal/news/service_test.go`:
```go
package news_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

var ctx = context.Background()

func seeded() news.News {
	return news.News{
		ID: 1, Title: "Opener", Content: null.StringFrom("<p>hi</p>"),
		FileName: null.StringFrom("news/old.jpg"), Date: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
	}
}

func newService(rows ...news.News) (*news.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore(rows...)
	objects := uploadtest.New()
	objects.Objects["news/old.jpg"] = "OLD"
	return news.NewService(store, upload.New(objects)), store, objects
}

func kind(t *testing.T, err error) svcerr.Kind {
	t.Helper()
	se, ok := svcerr.As(err)
	require.True(t, ok, "want svcerr, got %v", err)
	return se.Kind
}

func TestCreateSanitisesAndStoresImage(t *testing.T) {
	svc, store, objects := newService()
	a, err := svc.Create(ctx, news.CreateInput{Title: "  Win  ", Content: `<p>ok</p><script>x()</script>`},
		uploadtest.File("a.png", "image/png", "IMG"))
	require.NoError(t, err)

	assert.Positive(t, a.ID)
	assert.Equal(t, "Win", a.Title)
	assert.Equal(t, "<p>ok</p>", a.Content)
	row := store.row(a.ID)
	require.True(t, row.FileName.Valid)
	assert.Equal(t, "IMG", objects.Objects[row.FileName.String])
	assert.Equal(t, "https://cdn.test/"+row.FileName.String, a.ImageURL)
}

func TestCreateRequiresTitle(t *testing.T) {
	svc, _, _ := newService()
	_, err := svc.Create(ctx, news.CreateInput{Title: "   "}, nil)
	assert.Equal(t, svcerr.KindInvalid, kind(t, err))
}

func TestCreateRejectsHTMLUpload(t *testing.T) {
	svc, store, objects := newService()
	_, err := svc.Create(ctx, news.CreateInput{Title: "x"}, uploadtest.File("x.html", "text/html", "<script>"))
	se, _ := svcerr.As(err)
	require.NotNil(t, se)
	assert.Contains(t, se.Fields, "file")
	assert.Len(t, objects.Objects, 1, "only the pre-existing object")
	assert.Empty(t, store.rows)
}

func TestUpdateOmittedFieldsUnchanged(t *testing.T) {
	svc, store, objects := newService(seeded())
	title := "Renamed"
	_, err := svc.Update(ctx, 1, news.UpdateInput{Title: &title}, nil)
	require.NoError(t, err)

	row := store.row(1)
	assert.Equal(t, "Renamed", row.Title)
	assert.Equal(t, "<p>hi</p>", row.Content.String, "content was not sent, so it must be untouched")
	assert.Equal(t, "news/old.jpg", row.FileName.String)
	assert.Empty(t, objects.Deleted)
}

func TestUpdateEmptyContentClearsIt(t *testing.T) {
	svc, store, _ := newService(seeded())
	empty := ""
	_, err := svc.Update(ctx, 1, news.UpdateInput{Content: &empty}, nil)
	require.NoError(t, err)
	assert.False(t, store.row(1).Content.Valid)
}

func TestUpdateRemoveImage(t *testing.T) {
	svc, store, objects := newService(seeded())
	a, err := svc.Update(ctx, 1, news.UpdateInput{RemoveImage: true}, nil)
	require.NoError(t, err)
	assert.False(t, store.row(1).FileName.Valid)
	assert.Empty(t, a.ImageURL)
	assert.Equal(t, []string{"news/old.jpg"}, objects.Deleted)
}

func TestUpdateReplaceImage(t *testing.T) {
	svc, store, objects := newService(seeded())
	_, err := svc.Update(ctx, 1, news.UpdateInput{}, uploadtest.File("n.webp", "image/webp", "NEW"))
	require.NoError(t, err)
	key := store.row(1).FileName.String
	assert.NotEqual(t, "news/old.jpg", key)
	assert.Equal(t, "NEW", objects.Objects[key])
	assert.Equal(t, []string{"news/old.jpg"}, objects.Deleted)
}

func TestUpdateDBFailureKeepsOldImage(t *testing.T) {
	svc, store, objects := newService(seeded())
	store.editErr = errors.New("db down")
	_, err := svc.Update(ctx, 1, news.UpdateInput{}, uploadtest.File("n.png", "image/png", "NEW"))
	require.Error(t, err)
	assert.Equal(t, "OLD", objects.Objects["news/old.jpg"], "old image must survive a failed edit")
	assert.Len(t, objects.Objects, 1, "the new upload is cleaned up")
}

func TestGetMissingIsNotFound(t *testing.T) {
	svc, _, _ := newService()
	_, err := svc.Get(ctx, 42)
	assert.Equal(t, svcerr.KindNotFound, kind(t, err))
}

func TestDeleteRemovesImage(t *testing.T) {
	svc, store, objects := newService(seeded())
	a, err := svc.Delete(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "Opener", a.Title)
	assert.Empty(t, store.rows)
	assert.Equal(t, []string{"news/old.jpg"}, objects.Deleted)
}

func TestLatest(t *testing.T) {
	svc, _, _ := newService()
	_, ok, err := svc.Latest(ctx)
	require.NoError(t, err)
	assert.False(t, ok)

	svc, _, _ = newService(seeded())
	a, ok, err := svc.Latest(ctx)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 1, a.ID)
}
```

- [ ] **Step 5: Run to confirm they fail**

Run: `go test ./server/internal/news/`
Expected: compile errors (`undefined: news.NewService`, `news.CreateInput`, …).

- [ ] **Step 6: Implement `types.go` and `service.go`**

`server/internal/news/types.go`:
```go
package news

import "time"

// Article is a news article as the API returns it.
type Article struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	Date     time.Time `json:"date"`
	ImageURL string    `json:"imageUrl,omitempty"`
}

// CreateInput is a new article. Content is HTML from the editor and is sanitised.
type CreateInput struct {
	Title   string
	Content string
}

// UpdateInput changes an article. Nil fields are left unchanged; an empty
// Content clears it. RemoveImage is ignored when a new image is uploaded.
type UpdateInput struct {
	Title       *string
	Content     *string
	RemoveImage bool
}
```

`server/internal/news/service.go`:
```go
package news

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/sanitize"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

const uploadCategory = "news"

type store interface {
	GetNews(ctx context.Context) ([]News, error)
	GetNewsLatest(ctx context.Context) (News, error)
	GetNewsArticle(ctx context.Context, newsParam News) (News, error)
	AddNews(ctx context.Context, newsParam News) (News, error)
	EditNews(ctx context.Context, newsParam News) (News, error)
	DeleteNews(ctx context.Context, newsParam News) error
}

// Service is the news business logic shared by the API and legacy views.
type Service struct {
	store store
	files *upload.Files
}

func NewService(store store, files *upload.Files) *Service {
	return &Service{store: store, files: files}
}

func (s *Service) article(n News) Article {
	return Article{
		ID:       n.ID,
		Title:    n.Title,
		Content:  n.Content.String,
		Date:     n.Date,
		ImageURL: s.files.URL(n.FileName.String),
	}
}

func (s *Service) List(ctx context.Context) ([]Article, error) {
	ctx, span := tracer.Start(ctx, "news.Service.List")
	defer span.End()
	rows, err := s.store.GetNews(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list news: %w", err)
	}
	out := make([]Article, 0, len(rows))
	for _, n := range rows {
		out = append(out, s.article(n))
	}
	return out, nil
}

// Latest returns the newest article; ok is false when there are none.
func (s *Service) Latest(ctx context.Context) (Article, bool, error) {
	ctx, span := tracer.Start(ctx, "news.Service.Latest")
	defer span.End()
	n, err := s.store.GetNewsLatest(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return Article{}, false, nil
	}
	if err != nil {
		return Article{}, false, fmt.Errorf("failed to get latest news: %w", err)
	}
	return s.article(n), true, nil
}

func (s *Service) Get(ctx context.Context, id int) (Article, error) {
	ctx, span := tracer.Start(ctx, "news.Service.Get")
	defer span.End()
	n, err := s.get(ctx, id)
	if err != nil {
		return Article{}, err
	}
	return s.article(n), nil
}

func (s *Service) get(ctx context.Context, id int) (News, error) {
	n, err := s.store.GetNewsArticle(ctx, News{ID: id})
	if err != nil {
		return News{}, svcerr.FromStore(err, "news article")
	}
	return n, nil
}

func (s *Service) Create(ctx context.Context, in CreateInput, image *upload.File) (Article, error) {
	ctx, span := tracer.Start(ctx, "news.Service.Create")
	defer span.End()

	title := strings.TrimSpace(in.Title)
	if title == "" {
		return Article{}, svcerr.InvalidField("title", "title is required")
	}
	content := sanitize.HTML(in.Content)

	var key string
	if image != nil {
		var err error
		if key, err = s.files.Save(ctx, image, uploadCategory); err != nil {
			return Article{}, err
		}
	}

	n, err := s.store.AddNews(ctx, News{
		Title:    title,
		Content:  null.NewString(content, content != ""),
		FileName: null.NewString(key, key != ""),
	})
	if err != nil {
		s.files.Remove(ctx, key)
		return Article{}, fmt.Errorf("failed to add news: %w", err)
	}
	return s.article(n), nil
}

func (s *Service) Update(ctx context.Context, id int, in UpdateInput, image *upload.File) (Article, error) {
	ctx, span := tracer.Start(ctx, "news.Service.Update")
	defer span.End()

	n, err := s.get(ctx, id)
	if err != nil {
		return Article{}, err
	}
	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return Article{}, svcerr.InvalidField("title", "title is required")
		}
		n.Title = title
	}
	if in.Content != nil {
		content := sanitize.HTML(*in.Content)
		n.Content = null.NewString(content, content != "")
	}

	oldKey, newKey := n.FileName.String, ""
	switch {
	case image != nil:
		if newKey, err = s.files.Save(ctx, image, uploadCategory); err != nil {
			return Article{}, err
		}
		n.FileName = null.StringFrom(newKey)
	case in.RemoveImage:
		n.FileName = null.String{}
	}

	if _, err = s.store.EditNews(ctx, n); err != nil {
		s.files.Remove(ctx, newKey)
		return Article{}, fmt.Errorf("failed to edit news: %w", err)
	}
	if oldKey != "" && oldKey != n.FileName.String {
		s.files.Remove(ctx, oldKey)
	}
	return s.article(n), nil
}

// Delete removes the article and its image, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Article, error) {
	ctx, span := tracer.Start(ctx, "news.Service.Delete")
	defer span.End()

	n, err := s.get(ctx, id)
	if err != nil {
		return Article{}, err
	}
	deleted := s.article(n)
	if err = s.store.DeleteNews(ctx, n); err != nil {
		return Article{}, fmt.Errorf("failed to delete news: %w", err)
	}
	s.files.Remove(ctx, n.FileName.String)
	return deleted, nil
}
```
`tracer` is the package-level tracer already declared in `news.go`.

- [ ] **Step 7: Run service tests**

Run: `go test ./server/internal/news/`
Expected: PASS (store tests skip without `AFC_TEST_DB`).

- [ ] **Step 8: Write the failing handler tests**

`server/internal/news/handlers_test.go`:
```go
package news_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

type harness struct {
	anon, editor, manager *apitest.Client
	store                 *fakeStore
}

func newHarness(t *testing.T) harness {
	t.Helper()
	svc, store, _ := newService(seeded())
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	news.NewHandlers(svc).Register(api, sessions.Guards())
	c := apitest.New(e)
	return harness{
		anon:    c,
		editor:  c.As(authtest.Cookie(t, sessions, authtest.Webmaster)),
		manager: c.As(authtest.Cookie(t, sessions, authtest.Manager)),
		store:   store,
	}
}

func TestListAndGet(t *testing.T) {
	h := newHarness(t)
	rec := h.anon.Get(t, "/api/v1/news")
	require.Equal(t, http.StatusOK, rec.Code)
	list := apitest.Decode[[]news.Article](t, rec)
	require.Len(t, list, 1)
	assert.Equal(t, "https://cdn.test/news/old.jpg", list[0].ImageURL)

	rec = h.anon.Get(t, "/api/v1/news/1")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Opener", apitest.Decode[news.Article](t, rec).Title)

	rec = h.anon.Get(t, "/api/v1/news/999")
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestWritesRequireEditor(t *testing.T) {
	h := newHarness(t)
	cases := []struct {
		method, path string
	}{
		{http.MethodPost, "/api/v1/news"},
		{http.MethodPatch, "/api/v1/news/1"},
		{http.MethodDelete, "/api/v1/news/1"},
	}
	for _, tc := range cases {
		assert.Equal(t, http.StatusUnauthorized, h.anon.Multipart(t, tc.method, tc.path, map[string]string{"title": "x"}).Code, "anon %s %s", tc.method, tc.path)
		assert.Equal(t, http.StatusForbidden, h.manager.Multipart(t, tc.method, tc.path, map[string]string{"title": "x"}).Code, "manager %s %s", tc.method, tc.path)
	}
	assert.Equal(t, "Opener", h.store.row(1).Title, "nothing changed")
}

func TestCreate(t *testing.T) {
	h := newHarness(t)
	rec := h.editor.Multipart(t, http.MethodPost, "/api/v1/news",
		map[string]string{"title": "Cup run", "content": "<p>on</p>"},
		apitest.FilePart{Field: "image", Name: "c.jpg", ContentType: "image/jpeg", Body: "JPG"})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	a := apitest.Decode[news.Article](t, rec)
	assert.Positive(t, a.ID)
	assert.NotEmpty(t, a.ImageURL)

	rec = h.anon.Get(t, fmt.Sprintf("/api/v1/news/%d", a.ID))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCreateRejectsHTMLUploadOverHTTP(t *testing.T) {
	h := newHarness(t)
	rec := h.editor.Multipart(t, http.MethodPost, "/api/v1/news",
		map[string]string{"title": "x"},
		apitest.FilePart{Field: "image", Name: "x.html", ContentType: "text/html", Body: "<script>"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, apitest.ErrorOf(t, rec).Fields, "file")
}

func TestPatchIsPartial(t *testing.T) {
	h := newHarness(t)
	rec := h.editor.Multipart(t, http.MethodPatch, "/api/v1/news/1", map[string]string{"title": "Only title"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	row := h.store.row(1)
	assert.Equal(t, "Only title", row.Title)
	assert.Equal(t, "<p>hi</p>", row.Content.String)
	assert.Equal(t, "news/old.jpg", row.FileName.String)

	rec = h.editor.Multipart(t, http.MethodPatch, "/api/v1/news/1", map[string]string{"removeImage": "true"})
	require.Equal(t, http.StatusOK, rec.Code)
	assert.False(t, h.store.row(1).FileName.Valid)
}

func TestDelete(t *testing.T) {
	h := newHarness(t)
	rec := h.editor.JSON(t, http.MethodDelete, "/api/v1/news/1", nil)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, http.StatusNotFound, h.anon.Get(t, "/api/v1/news/1").Code)
	assert.Equal(t, http.StatusNotFound, h.editor.JSON(t, http.MethodDelete, "/api/v1/news/1", nil).Code)
}
```

- [ ] **Step 9: Run to confirm they fail**

Run: `go test ./server/internal/news/`
Expected: compile error (`undefined: news.NewHandlers`).

- [ ] **Step 10: Implement `handlers.go`**

`server/internal/news/handlers.go`:
```go
package news

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /news.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the news routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/news", h.list)
	g.GET("/news/:id", h.get)
	g.POST("/news", h.create, guards.Editor)
	g.PATCH("/news/:id", h.update, guards.Editor)
	g.DELETE("/news/:id", h.remove, guards.Editor)
}

// list returns every article, newest first.
//
//	@Summary	List news
//	@Tags		news
//	@Produce	json
//	@Success	200	{array}	news.Article
//	@Router		/news [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// get returns one article.
//
//	@Summary	Get a news article
//	@Tags		news
//	@Produce	json
//	@Param		id	path		int	true	"Article ID"
//	@Success	200	{object}	news.Article
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/news/{id} [get]
func (h *Handlers) get(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	a, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, a)
}

// create adds an article.
//
//	@Summary	Create a news article
//	@Tags		news
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		title	formData	string	true	"Title"
//	@Param		content	formData	string	false	"HTML content (sanitised)"
//	@Param		image	formData	file	false	"Image"
//	@Success	201		{object}	news.Article
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/news [post]
func (h *Handlers) create(c echo.Context) error {
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	a, err := h.svc.Create(c.Request().Context(), CreateInput{
		Title:   c.FormValue("title"),
		Content: c.FormValue("content"),
	}, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, a)
}

// update changes an article; omitted fields are left as they are.
//
//	@Summary	Update a news article
//	@Tags		news
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		id			path		int		true	"Article ID"
//	@Param		title		formData	string	false	"Title"
//	@Param		content		formData	string	false	"HTML content (sanitised); empty clears it"
//	@Param		image		formData	file	false	"Replacement image"
//	@Param		removeImage	formData	bool	false	"Remove the current image"
//	@Success	200			{object}	news.Article
//	@Failure	404			{object}	web.ErrorResponse
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/news/{id} [patch]
func (h *Handlers) update(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	remove, err := web.FormBool(c, "removeImage")
	if err != nil {
		return err
	}
	a, err := h.svc.Update(c.Request().Context(), id, UpdateInput{
		Title:       web.FormString(c, "title"),
		Content:     web.FormString(c, "content"),
		RemoveImage: remove != nil && *remove,
	}, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, a)
}

// remove deletes an article.
//
//	@Summary	Delete a news article
//	@Tags		news
//	@Param		id	path	int	true	"Article ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/news/{id} [delete]
func (h *Handlers) remove(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if _, err = h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
```

- [ ] **Step 11: Run handler tests**

Run: `go test ./server/internal/news/`
Expected: PASS.

- [ ] **Step 12: Add legacy helpers and point legacy news writes at the service**

Append to `server/internal/legacy/views/helpers.go`:
```go
// legacyUpload returns the optional file in field, or nil when none was sent.
func legacyUpload(c echo.Context, field string) (*upload.File, error) {
	fh, err := c.FormFile(field)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return nil, nil //nolint:nilerr // uploads are optional
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	return upload.FromHeader(fh), nil
}

// flash queues a success message for the next page view.
func (v *Views) flash(c echo.Context, c1 *Context, message string) {
	c1.Message = message
	c1.MsgType = "is-success"
	if err := v.setMessagesInSession(c, c1); err != nil {
		slog.Info(fmt.Sprintf("failed to set flash message: %+v", err))
	}
}

// formYes maps a legacy "Y" checkbox to a bool; anything else is false.
func formYes(c echo.Context, field string) bool {
	return c.FormValue(field) == "Y"
}
```
(Add `errors`, `net/http` and `upload` to the imports.)

In `server/internal/legacy/views/views.go`, add a field `newsSvc *news.Service` to `Views`, `NewsService *news.Service` to `Deps`, and `newsSvc: d.NewsService,` in `New`.

In `server/internal/legacy/views/news.go`, leave `NewsFunc` and `NewsArticleFunc` alone and replace `NewsAddFunc`, `NewsEditFunc` and `NewsDeleteFunc` with:
```go
func (v *Views) NewsAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.NewsAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}

	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for news add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	article, err := v.newsSvc.Create(c.Request().Context(), news.CreateInput{
		Title:   c.FormValue("title"),
		Content: c.FormValue("htmlContent"),
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add news for news add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add news for news add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}

	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", article.Title))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) NewsEditFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.NewsEditFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	newsID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse id for news edit, error: %w", err))
	}
	data := struct {
		Error string `json:"error"`
	}{}

	remove := c.FormValue("removeNewsImage")
	if remove != "" && remove != "Y" {
		data.Error = "failed to parse removeNewsImage for news edit: " + remove
		return c.JSON(http.StatusOK, data)
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for news edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	title, content := c.FormValue("title"), c.FormValue("htmlContent")
	article, err := v.newsSvc.Update(c.Request().Context(), newsID, news.UpdateInput{
		Title:       &title,
		Content:     &content,
		RemoveImage: remove == "Y",
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to edit news for news edit, news id: %d, error: %+v", newsID, err))
		data.Error = fmt.Sprintf("failed to edit news for news edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}

	v.flash(c, c1, fmt.Sprintf("successfully edited \"%s\"", article.Title))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) NewsDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.NewsDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for news delete, error: %w", err)
	}
	article, err := v.newsSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete news for news delete, news id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", article.Title))
	return c.Redirect(http.StatusFound, "/news")
}
```
Run `goimports -w server/internal/legacy/views` to drop the now-unused `bluemonday`, `null` and `strings` imports from `news.go`.

- [ ] **Step 13: Wire into `app.Build`**

In `server/internal/app/app.go` `Build`:
- right after `uploads := upload.New(objects)`, add:
```go
	newsSvc := news.NewService(s.News, uploads)
```
- in the `views.Deps{…}` literal, add `NewsService: newsSvc,`
- after `auth.NewHandlers(sessions, uploads).Register(api, guards)`, add:
```go
	news.NewHandlers(newsSvc).Register(api, guards)
```

- [ ] **Step 14: Regenerate docs, run everything**

```bash
go generate ./server/internal/docs
gofmt -l server; go build ./... && go vet ./...
go test ./...
go tool golangci-lint run ./...
```
Expected: all green. `TestUnsafeAPIRoutesRequireLogin` now logs `checked 3 unsafe API routes`.

- [ ] **Step 15: Commit**

```bash
git add -A server
git commit -m "Add news service and /api/v1/news; legacy news writes use the service

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---
### Task 9: What's On

The same shape as Task 8, plus a required `dateOfEvent` and a `period` filter.

**Files:**
- Modify: `server/internal/whatson/store.go` (`addWhatsOn` returns ID)
- Create: `server/internal/whatson/{store_test,types,service,handlers,fake_test,service_test,handlers_test}.go`
- Modify: `server/internal/legacy/views/whatson.go`, `server/internal/legacy/views/views.go`, `server/internal/app/app.go`

**Interfaces:**
- Consumes: the same as Task 8, plus the legacy helpers `legacyUpload`, `flash` (Task 8).
- Produces (`whatson`):
  - `type Event{ID int; Title, Content string; Date, DateOfEvent time.Time; ImageURL string}` (JSON `dateOfEvent`, `imageUrl`)
  - `type Period string` with `PeriodAll`, `PeriodFuture`, `PeriodPast`
  - `CreateInput{Title, Content string; DateOfEvent time.Time}`
  - `UpdateInput{Title, Content *string; DateOfEvent *time.Time; RemoveImage bool}`
  - `NewService(store, *upload.Files) *Service`, with methods:
    - `List(ctx, Period) ([]Event, error)`
    - `Next(ctx) (Event, bool, error)` (the soonest upcoming event)
    - `Get`, `Create`, `Update`, `Delete(ctx, id) (Event, error)`
  - `NewHandlers(*Service) *Handlers` and `Register`.
- Produces (legacy): `Deps.WhatsOnService *whatson.Service`.

- [ ] **Step 1: Write the failing store test**

`server/internal/whatson/store_test.go`:
```go
package whatson_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

func TestAddWhatsOnReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	store := whatson.NewWhatsOnRepo(db)
	ctx := context.Background()

	added, err := store.AddWhatsOn(ctx, whatson.WhatsOn{Title: "Quiz", DateOfEvent: time.Now().AddDate(0, 0, 7)})
	require.NoError(t, err)
	require.Positive(t, added.ID)
	got, err := store.GetWhatsOnArticle(ctx, whatson.WhatsOn{ID: added.ID})
	require.NoError(t, err)
	assert.Equal(t, "Quiz", got.Title)
}
```
Run: `go test -run TestAddWhatsOnReturnsID ./server/internal/whatson/`. Expected: FAIL (`added.ID` is 0).

- [ ] **Step 2: Fix the store**

In `server/internal/whatson/store.go` `addWhatsOn`: add `.Suffix("RETURNING id")` to the builder and replace the `ExecContext`/`RowsAffected` block with:
```go
	err = s.db.GetContext(ctx, &whatsOnParam.ID, sql, args...)
	if err != nil {
		span.RecordError(err)
		return WhatsOn{}, fmt.Errorf("failed to add whatsOn: %w", err)
	}
	return whatsOnParam, nil
```
Re-run: PASS. (`EditWhatsOn` is already write-through.)

- [ ] **Step 3: Write the fake store and failing service tests**

`server/internal/whatson/fake_test.go`:
```go
package whatson_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

var today = time.Now().Truncate(24 * time.Hour)

type fakeStore struct {
	mu      sync.Mutex
	rows    map[int]whatson.WhatsOn
	nextID  int
	editErr error
}

func newFakeStore(rows ...whatson.WhatsOn) *fakeStore {
	f := &fakeStore{rows: map[int]whatson.WhatsOn{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) filter(keep func(whatson.WhatsOn) bool, asc bool) []whatson.WhatsOn {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []whatson.WhatsOn{}
	for _, r := range f.rows {
		if keep(r) {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if asc {
			return out[i].DateOfEvent.Before(out[j].DateOfEvent)
		}
		return out[i].DateOfEvent.After(out[j].DateOfEvent)
	})
	return out
}

func (f *fakeStore) GetWhatsOn(context.Context) ([]whatson.WhatsOn, error) {
	return f.filter(func(whatson.WhatsOn) bool { return true }, true), nil
}

func (f *fakeStore) GetWhatsOnFuture(context.Context) ([]whatson.WhatsOn, error) {
	return f.filter(func(w whatson.WhatsOn) bool { return !w.DateOfEvent.Before(today) }, true), nil
}

func (f *fakeStore) GetWhatsOnPast(context.Context) ([]whatson.WhatsOn, error) {
	return f.filter(func(w whatson.WhatsOn) bool { return w.DateOfEvent.Before(today) }, false), nil
}

func (f *fakeStore) GetWhatsOnLatest(ctx context.Context) (whatson.WhatsOn, error) {
	future, _ := f.GetWhatsOnFuture(ctx)
	if len(future) == 0 {
		return whatson.WhatsOn{}, fmt.Errorf("failed to get whats on latest: %w", sql.ErrNoRows)
	}
	return future[0], nil
}

func (f *fakeStore) GetWhatsOnArticle(_ context.Context, w whatson.WhatsOn) (whatson.WhatsOn, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[w.ID]
	if !ok {
		return whatson.WhatsOn{}, fmt.Errorf("failed to get whats on: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddWhatsOn(_ context.Context, w whatson.WhatsOn) (whatson.WhatsOn, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	w.ID = f.nextID
	w.Date = time.Now()
	f.rows[w.ID] = w
	return w, nil
}

func (f *fakeStore) EditWhatsOn(_ context.Context, w whatson.WhatsOn) (whatson.WhatsOn, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.editErr != nil {
		return whatson.WhatsOn{}, f.editErr
	}
	if _, ok := f.rows[w.ID]; !ok {
		return whatson.WhatsOn{}, errors.New("no such row")
	}
	f.rows[w.ID] = w
	return w, nil
}

func (f *fakeStore) DeleteWhatsOn(_ context.Context, w whatson.WhatsOn) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, w.ID)
	return nil
}

func (f *fakeStore) row(id int) whatson.WhatsOn {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.rows[id]
}
```

`server/internal/whatson/service_test.go`:
```go
package whatson_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

var ctx = context.Background()

func upcoming() whatson.WhatsOn {
	return whatson.WhatsOn{ID: 1, Title: "Presentation night", Content: null.StringFrom("<p>Clubhouse</p>"),
		FileName: null.StringFrom("whatson/old.jpg"), DateOfEvent: today.AddDate(0, 0, 30)}
}

func finished() whatson.WhatsOn {
	return whatson.WhatsOn{ID: 2, Title: "Summer BBQ", DateOfEvent: today.AddDate(0, 0, -30)}
}

func newService(rows ...whatson.WhatsOn) (*whatson.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore(rows...)
	objects := uploadtest.New()
	objects.Objects["whatson/old.jpg"] = "OLD"
	return whatson.NewService(store, upload.New(objects)), store, objects
}

func TestListByPeriod(t *testing.T) {
	svc, _, _ := newService(upcoming(), finished())
	for period, want := range map[whatson.Period][]int{
		whatson.PeriodAll:    {2, 1},
		whatson.PeriodFuture: {1},
		whatson.PeriodPast:   {2},
		"":                   {2, 1},
	} {
		got, err := svc.List(ctx, period)
		require.NoError(t, err, period)
		ids := []int{}
		for _, e := range got {
			ids = append(ids, e.ID)
		}
		assert.Equal(t, want, ids, "period %q", period)
	}
}

func TestListRejectsUnknownPeriod(t *testing.T) {
	svc, _, _ := newService()
	_, err := svc.List(ctx, "someday")
	se, ok := svcerr.As(err)
	require.True(t, ok)
	assert.Contains(t, se.Fields, "period")
}

func TestCreateValidates(t *testing.T) {
	svc, _, _ := newService()
	_, err := svc.Create(ctx, whatson.CreateInput{Title: "", DateOfEvent: today}, nil)
	se, _ := svcerr.As(err)
	require.NotNil(t, se)
	assert.Contains(t, se.Fields, "title")

	_, err = svc.Create(ctx, whatson.CreateInput{Title: "x"}, nil)
	se, _ = svcerr.As(err)
	require.NotNil(t, se)
	assert.Contains(t, se.Fields, "dateOfEvent")
}

func TestCreateSanitises(t *testing.T) {
	svc, store, _ := newService()
	e, err := svc.Create(ctx, whatson.CreateInput{Title: "Quiz", Content: `<p>a</p><script>b</script>`, DateOfEvent: today}, nil)
	require.NoError(t, err)
	assert.Equal(t, "<p>a</p>", store.row(e.ID).Content.String)
}

func TestUpdateRemoveImageDeletesObject(t *testing.T) {
	svc, store, objects := newService(upcoming())
	_, err := svc.Update(ctx, 1, whatson.UpdateInput{RemoveImage: true}, nil)
	require.NoError(t, err)
	assert.False(t, store.row(1).FileName.Valid)
	assert.Equal(t, []string{"whatson/old.jpg"}, objects.Deleted, "legacy left this object behind; now it is removed")
}

func TestUpdatePartial(t *testing.T) {
	svc, store, _ := newService(upcoming())
	moved := today.AddDate(0, 0, 40)
	_, err := svc.Update(ctx, 1, whatson.UpdateInput{DateOfEvent: &moved}, nil)
	require.NoError(t, err)
	row := store.row(1)
	assert.True(t, row.DateOfEvent.Equal(moved))
	assert.Equal(t, "Presentation night", row.Title)
	assert.Equal(t, "whatson/old.jpg", row.FileName.String)
}

func TestNext(t *testing.T) {
	svc, _, _ := newService(finished())
	_, ok, err := svc.Next(ctx)
	require.NoError(t, err)
	assert.False(t, ok)

	svc, _, _ = newService(upcoming(), finished())
	e, ok, err := svc.Next(ctx)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, 1, e.ID)
}

func TestDeleteRemovesImage(t *testing.T) {
	svc, _, objects := newService(upcoming())
	e, err := svc.Delete(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "Presentation night", e.Title)
	assert.Equal(t, []string{"whatson/old.jpg"}, objects.Deleted)
}
```
Run: `go test ./server/internal/whatson/`. Expected: compile errors (`undefined: whatson.NewService`).

- [ ] **Step 4: Implement `types.go` and `service.go`**

`server/internal/whatson/types.go`:
```go
package whatson

import "time"

// Event is a what's-on entry as the API returns it.
type Event struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Date        time.Time `json:"date"`
	DateOfEvent time.Time `json:"dateOfEvent"`
	ImageURL    string    `json:"imageUrl,omitempty"`
}

// Period filters events by date of event relative to today.
type Period string

const (
	PeriodAll    Period = "all"
	PeriodFuture Period = "future"
	PeriodPast   Period = "past"
)

// CreateInput is a new event. Content is sanitised HTML.
type CreateInput struct {
	Title       string
	Content     string
	DateOfEvent time.Time
}

// UpdateInput changes an event; nil fields are left unchanged.
type UpdateInput struct {
	Title       *string
	Content     *string
	DateOfEvent *time.Time
	RemoveImage bool
}
```

`server/internal/whatson/service.go`:
```go
package whatson

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/sanitize"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

const uploadCategory = "whatson"

type store interface {
	GetWhatsOn(ctx context.Context) ([]WhatsOn, error)
	GetWhatsOnFuture(ctx context.Context) ([]WhatsOn, error)
	GetWhatsOnPast(ctx context.Context) ([]WhatsOn, error)
	GetWhatsOnLatest(ctx context.Context) (WhatsOn, error)
	GetWhatsOnArticle(ctx context.Context, whatsOnParam WhatsOn) (WhatsOn, error)
	AddWhatsOn(ctx context.Context, whatsOnParam WhatsOn) (WhatsOn, error)
	EditWhatsOn(ctx context.Context, whatsOnParam WhatsOn) (WhatsOn, error)
	DeleteWhatsOn(ctx context.Context, whatsOnParam WhatsOn) error
}

// Service is the what's-on business logic shared by the API and legacy views.
type Service struct {
	store store
	files *upload.Files
}

func NewService(store store, files *upload.Files) *Service {
	return &Service{store: store, files: files}
}

func (s *Service) event(w WhatsOn) Event {
	return Event{
		ID:          w.ID,
		Title:       w.Title,
		Content:     w.Content.String,
		Date:        w.Date,
		DateOfEvent: w.DateOfEvent,
		ImageURL:    s.files.URL(w.FileName.String),
	}
}

// List returns events for period; an empty period means all.
func (s *Service) List(ctx context.Context, period Period) ([]Event, error) {
	ctx, span := tracer.Start(ctx, "whatson.Service.List")
	defer span.End()

	var (
		rows []WhatsOn
		err  error
	)
	switch period {
	case PeriodAll, "":
		rows, err = s.store.GetWhatsOn(ctx)
	case PeriodFuture:
		rows, err = s.store.GetWhatsOnFuture(ctx)
	case PeriodPast:
		rows, err = s.store.GetWhatsOnPast(ctx)
	default:
		return nil, svcerr.InvalidField("period", "period must be all, future or past")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list whats on: %w", err)
	}
	out := make([]Event, 0, len(rows))
	for _, w := range rows {
		out = append(out, s.event(w))
	}
	return out, nil
}

// Next returns the soonest upcoming event; ok is false when there is none.
func (s *Service) Next(ctx context.Context) (Event, bool, error) {
	ctx, span := tracer.Start(ctx, "whatson.Service.Next")
	defer span.End()
	w, err := s.store.GetWhatsOnLatest(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return Event{}, false, nil
	}
	if err != nil {
		return Event{}, false, fmt.Errorf("failed to get next whats on: %w", err)
	}
	return s.event(w), true, nil
}

func (s *Service) Get(ctx context.Context, id int) (Event, error) {
	ctx, span := tracer.Start(ctx, "whatson.Service.Get")
	defer span.End()
	w, err := s.get(ctx, id)
	if err != nil {
		return Event{}, err
	}
	return s.event(w), nil
}

func (s *Service) get(ctx context.Context, id int) (WhatsOn, error) {
	w, err := s.store.GetWhatsOnArticle(ctx, WhatsOn{ID: id})
	if err != nil {
		return WhatsOn{}, svcerr.FromStore(err, "whats on event")
	}
	return w, nil
}

func (s *Service) Create(ctx context.Context, in CreateInput, image *upload.File) (Event, error) {
	ctx, span := tracer.Start(ctx, "whatson.Service.Create")
	defer span.End()

	title := strings.TrimSpace(in.Title)
	fields := svcerr.Fields{}
	if title == "" {
		fields.Add("title", "title is required")
	}
	if in.DateOfEvent.IsZero() {
		fields.Add("dateOfEvent", "date of event is required")
	}
	if err := fields.Err(); err != nil {
		return Event{}, err
	}
	content := sanitize.HTML(in.Content)

	var key string
	if image != nil {
		var err error
		if key, err = s.files.Save(ctx, image, uploadCategory); err != nil {
			return Event{}, err
		}
	}
	w, err := s.store.AddWhatsOn(ctx, WhatsOn{
		Title:       title,
		Content:     null.NewString(content, content != ""),
		FileName:    null.NewString(key, key != ""),
		DateOfEvent: in.DateOfEvent,
	})
	if err != nil {
		s.files.Remove(ctx, key)
		return Event{}, fmt.Errorf("failed to add whats on: %w", err)
	}
	return s.event(w), nil
}

func (s *Service) Update(ctx context.Context, id int, in UpdateInput, image *upload.File) (Event, error) {
	ctx, span := tracer.Start(ctx, "whatson.Service.Update")
	defer span.End()

	w, err := s.get(ctx, id)
	if err != nil {
		return Event{}, err
	}
	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return Event{}, svcerr.InvalidField("title", "title is required")
		}
		w.Title = title
	}
	if in.Content != nil {
		content := sanitize.HTML(*in.Content)
		w.Content = null.NewString(content, content != "")
	}
	if in.DateOfEvent != nil {
		if in.DateOfEvent.IsZero() {
			return Event{}, svcerr.InvalidField("dateOfEvent", "date of event is required")
		}
		w.DateOfEvent = *in.DateOfEvent
	}

	oldKey, newKey := w.FileName.String, ""
	switch {
	case image != nil:
		if newKey, err = s.files.Save(ctx, image, uploadCategory); err != nil {
			return Event{}, err
		}
		w.FileName = null.StringFrom(newKey)
	case in.RemoveImage:
		w.FileName = null.String{}
	}

	if _, err = s.store.EditWhatsOn(ctx, w); err != nil {
		s.files.Remove(ctx, newKey)
		return Event{}, fmt.Errorf("failed to edit whats on: %w", err)
	}
	if oldKey != "" && oldKey != w.FileName.String {
		s.files.Remove(ctx, oldKey)
	}
	return s.event(w), nil
}

// Delete removes the event and its image, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Event, error) {
	ctx, span := tracer.Start(ctx, "whatson.Service.Delete")
	defer span.End()
	w, err := s.get(ctx, id)
	if err != nil {
		return Event{}, err
	}
	deleted := s.event(w)
	if err = s.store.DeleteWhatsOn(ctx, w); err != nil {
		return Event{}, fmt.Errorf("failed to delete whats on: %w", err)
	}
	s.files.Remove(ctx, w.FileName.String)
	return deleted, nil
}
```
Run: `go test ./server/internal/whatson/`. Expected: PASS.

- [ ] **Step 5: Write the failing handler tests**

`server/internal/whatson/handlers_test.go`:
```go
package whatson_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

func newHarness(t *testing.T) (anon, editor, manager *apitest.Client, store *fakeStore) {
	t.Helper()
	svc, store, _ := newService(upcoming(), finished())
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	whatson.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	c := apitest.New(e)
	return c, c.As(authtest.Cookie(t, sessions, authtest.Webmaster)), c.As(authtest.Cookie(t, sessions, authtest.Manager)), store
}

func TestListPeriodQuery(t *testing.T) {
	anon, _, _, _ := newHarness(t)
	rec := anon.Get(t, "/api/v1/whatson?period=future")
	require.Equal(t, http.StatusOK, rec.Code)
	got := apitest.Decode[[]whatson.Event](t, rec)
	require.Len(t, got, 1)
	assert.Equal(t, 1, got[0].ID)

	assert.Equal(t, http.StatusUnprocessableEntity, anon.Get(t, "/api/v1/whatson?period=soon").Code)
	assert.Equal(t, http.StatusOK, anon.Get(t, "/api/v1/whatson/2").Code)
	assert.Equal(t, http.StatusNotFound, anon.Get(t, "/api/v1/whatson/99").Code)
}

func TestWhatsOnWritesRequireEditor(t *testing.T) {
	anon, _, manager, _ := newHarness(t)
	fields := map[string]string{"title": "x", "dateOfEvent": "2030-01-01"}
	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/api/v1/whatson"},
		{http.MethodPatch, "/api/v1/whatson/1"},
		{http.MethodDelete, "/api/v1/whatson/1"},
	} {
		assert.Equal(t, http.StatusUnauthorized, anon.Multipart(t, tc.method, tc.path, fields).Code)
		assert.Equal(t, http.StatusForbidden, manager.Multipart(t, tc.method, tc.path, fields).Code)
	}
}

func TestCreateAndPatch(t *testing.T) {
	_, editor, _, store := newHarness(t)
	rec := editor.Multipart(t, http.MethodPost, "/api/v1/whatson",
		map[string]string{"title": "Quiz", "content": "<p>7pm</p>", "dateOfEvent": "2030-02-03"})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	created := apitest.Decode[whatson.Event](t, rec)
	assert.Equal(t, "2030-02-03", created.DateOfEvent.Format(web.DateLayout))

	rec = editor.Multipart(t, http.MethodPost, "/api/v1/whatson", map[string]string{"title": "No date"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, apitest.ErrorOf(t, rec).Fields, "dateOfEvent")

	rec = editor.Multipart(t, http.MethodPatch, "/api/v1/whatson/1", map[string]string{"dateOfEvent": "03/02/2030"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "dates must be YYYY-MM-DD")
	assert.Equal(t, "Presentation night", store.row(1).Title)
}
```
Run: `go test ./server/internal/whatson/`. Expected: compile error (`undefined: whatson.NewHandlers`).

- [ ] **Step 6: Implement `handlers.go`**

`server/internal/whatson/handlers.go`:
```go
package whatson

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /whatson.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the what's-on routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/whatson", h.list)
	g.GET("/whatson/:id", h.get)
	g.POST("/whatson", h.create, guards.Editor)
	g.PATCH("/whatson/:id", h.update, guards.Editor)
	g.DELETE("/whatson/:id", h.remove, guards.Editor)
}

// list returns events, optionally filtered by period.
//
//	@Summary	List what's on
//	@Tags		whatson
//	@Produce	json
//	@Param		period	query		string	false	"all (default), future or past"
//	@Success	200		{array}		whatson.Event
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/whatson [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context(), Period(c.QueryParam("period")))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// get returns one event.
//
//	@Summary	Get a what's-on event
//	@Tags		whatson
//	@Produce	json
//	@Param		id	path		int	true	"Event ID"
//	@Success	200	{object}	whatson.Event
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/whatson/{id} [get]
func (h *Handlers) get(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	e, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, e)
}

// create adds an event.
//
//	@Summary	Create a what's-on event
//	@Tags		whatson
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		title		formData	string	true	"Title"
//	@Param		content		formData	string	false	"HTML content (sanitised)"
//	@Param		dateOfEvent	formData	string	true	"YYYY-MM-DD"
//	@Param		image		formData	file	false	"Image"
//	@Success	201			{object}	whatson.Event
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/whatson [post]
func (h *Handlers) create(c echo.Context) error {
	date, err := web.FormDate(c, "dateOfEvent")
	if err != nil {
		return err
	}
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	in := CreateInput{Title: c.FormValue("title"), Content: c.FormValue("content")}
	if date != nil {
		in.DateOfEvent = *date
	}
	e, err := h.svc.Create(c.Request().Context(), in, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, e)
}

// update changes an event; omitted fields are left as they are.
//
//	@Summary	Update a what's-on event
//	@Tags		whatson
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		id			path		int		true	"Event ID"
//	@Param		title		formData	string	false	"Title"
//	@Param		content		formData	string	false	"HTML content; empty clears it"
//	@Param		dateOfEvent	formData	string	false	"YYYY-MM-DD"
//	@Param		image		formData	file	false	"Replacement image"
//	@Param		removeImage	formData	bool	false	"Remove the current image"
//	@Success	200			{object}	whatson.Event
//	@Failure	404			{object}	web.ErrorResponse
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/whatson/{id} [patch]
func (h *Handlers) update(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	var date *time.Time
	if date, err = web.FormDate(c, "dateOfEvent"); err != nil {
		return err
	}
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	remove, err := web.FormBool(c, "removeImage")
	if err != nil {
		return err
	}
	e, err := h.svc.Update(c.Request().Context(), id, UpdateInput{
		Title:       web.FormString(c, "title"),
		Content:     web.FormString(c, "content"),
		DateOfEvent: date,
		RemoveImage: remove != nil && *remove,
	}, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, e)
}

// remove deletes an event.
//
//	@Summary	Delete a what's-on event
//	@Tags		whatson
//	@Param		id	path	int	true	"Event ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/whatson/{id} [delete]
func (h *Handlers) remove(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if _, err = h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
```
Run: `go test ./server/internal/whatson/`. Expected: PASS.

- [ ] **Step 7: Point legacy what's-on writes at the service**

In `views.go`, add `whatsOnSvc *whatson.Service` to `Views`, `WhatsOnService *whatson.Service` to `Deps`, and `whatsOnSvc: d.WhatsOnService,` in `New`.

In `server/internal/legacy/views/whatson.go`, keep `WhatsOnFunc`, `WhatsOnTomePeriodFunc`, `WhatsOnSelectFunc` and `WhatsOnArticleFunc` as they are. Replace the three write handlers:
```go
func (v *Views) WhatsOnAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.WhatsOnAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}

	dateOfEvent, err := time.Parse("02/01/2006", c.FormValue("dateOfEvent"))
	if err != nil {
		data.Error = fmt.Sprintf("failed to parse dateOfEvent for whats on add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for whats on add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	event, err := v.whatsOnSvc.Create(c.Request().Context(), whatson.CreateInput{
		Title:       c.FormValue("title"),
		Content:     c.FormValue("htmlContent"),
		DateOfEvent: dateOfEvent,
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add whatsOn for whats on add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add whatsOn for whats on add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", event.Title))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) WhatsOnEditFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.WhatsOnEditFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	whatsOnID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse id for whats on edit, error: %w", err))
	}
	data := struct {
		Error string `json:"error"`
	}{}

	dateOfEvent, err := time.Parse("02/01/2006", c.FormValue("dateOfEvent"))
	if err != nil {
		data.Error = fmt.Sprintf("failed to parse dateOfEvent for whats on edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	remove := c.FormValue("removeWhatsOnImage")
	if remove != "" && remove != "Y" {
		data.Error = "failed to parse removeWhatsOnImage for whats on edit, value: " + remove
		return c.JSON(http.StatusOK, data)
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for whats on edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	title, content := c.FormValue("title"), c.FormValue("htmlContent")
	event, err := v.whatsOnSvc.Update(c.Request().Context(), whatsOnID, whatson.UpdateInput{
		Title:       &title,
		Content:     &content,
		DateOfEvent: &dateOfEvent,
		RemoveImage: remove == "Y",
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to edit whatsOn for whats on edit, whats on id: %d, error: %+v", whatsOnID, err))
		data.Error = fmt.Sprintf("failed to edit whatsOn for whats on edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully edited \"%s\"", event.Title))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) WhatsOnDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.WhatsOnDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for whats on delete, error: %w", err)
	}
	event, err := v.whatsOnSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete whatsOn for whats on delete, whats on id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", event.Title))
	return c.Redirect(http.StatusFound, "/whatson")
}
```
Run `goimports -w server/internal/legacy/views`.

- [ ] **Step 8: Wire into `app.Build`**

- After `newsSvc := …` add `whatsOnSvc := whatson.NewService(s.WhatsOn, uploads)`.
- Add `WhatsOnService: whatsOnSvc,` to `views.Deps`.
- After the news handlers add `whatson.NewHandlers(whatsOnSvc).Register(api, guards)`.

- [ ] **Step 9: Regenerate docs, run everything, commit**

```bash
go generate ./server/internal/docs
gofmt -l server; go build ./... && go vet ./... && go test ./... && go tool golangci-lint run ./...
git add -A server
git commit -m "Add what's on service and /api/v1/whatson; legacy writes use the service

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---
### Task 10: Sponsors

Sponsors support create, list and delete (there is no edit today). A sponsor's `team` is `"A"`, `"O"` or `"Y"` (club-wide audience codes kept verbatim from legacy) or a team ID as a string.

**Files:**
- Modify: `server/internal/sponsor/store.go` (`addSponsor` returns ID)
- Create: `server/internal/sponsor/{store_test,types,service,handlers,fake_test,service_test,handlers_test}.go`
- Modify: `server/internal/legacy/views/sponsor.go`, `views.go`, `server/internal/app/app.go`

**Interfaces:**
- Produces (`sponsor`):
  - Types: `Public{ID int; Name, Website, Purpose, Team, ImageURL string}` (the API type; `Sponsor` stays the DB model) and `CreateInput{Name, Website, Purpose, Team string}`.
  - `type TeamGetter interface{GetTeam(ctx, team.Team) (team.Team, error)}`
  - `NewService(store, TeamGetter, *upload.Files) *Service`, with methods:
    - `List(ctx) ([]Public, error)`
    - `ListMinimal(ctx) ([]Public, error)` (home page)
    - `ForTeam(ctx, teamID int) ([]Public, error)`
    - `Create(ctx, CreateInput, *upload.File) (Public, error)` (image required)
    - `Delete(ctx, id) (Public, error)`
  - `NewHandlers`, `Register`.
- Produces (legacy): `Deps.SponsorService`.

- [ ] **Step 1: Store test, then fix `addSponsor`**

`server/internal/sponsor/store_test.go`:
```go
package sponsor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestAddSponsorReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	added, err := sponsor.NewSponsorRepo(db).AddSponsor(context.Background(), sponsor.Sponsor{Name: "New", TeamID: "A"})
	require.NoError(t, err)
	require.Positive(t, added.ID)
}
```
Run it (FAIL: ID is 0). Then in `store.go` `addSponsor`, add `.Suffix("RETURNING id")` and replace the `ExecContext`/`RowsAffected` block with:
```go
	err = s.db.GetContext(ctx, &sponsorParam.ID, sql, args...)
	if err != nil {
		span.RecordError(err)
		return Sponsor{}, fmt.Errorf("failed to add sponsor: %w", err)
	}
	return sponsorParam, nil
```
Re-run: PASS.

- [ ] **Step 2: Fake store and failing service tests**

`server/internal/sponsor/fake_test.go`:
```go
package sponsor_test

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
)

type fakeStore struct {
	mu     sync.Mutex
	rows   map[int]sponsor.Sponsor
	nextID int
}

func newFakeStore(rows ...sponsor.Sponsor) *fakeStore {
	f := &fakeStore{rows: map[int]sponsor.Sponsor{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) all() []sponsor.Sponsor {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]sponsor.Sponsor, 0, len(f.rows))
	for _, r := range f.rows {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (f *fakeStore) GetSponsors(context.Context) ([]sponsor.Sponsor, error)        { return f.all(), nil }
func (f *fakeStore) GetSponsorsMinimal(context.Context) ([]sponsor.Sponsor, error) { return f.all(), nil }

func (f *fakeStore) GetSponsorsTeam(_ context.Context, t team.Team) ([]sponsor.Sponsor, error) {
	out := []sponsor.Sponsor{}
	for _, r := range f.all() {
		if r.TeamID == strconv.Itoa(t.ID) {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f *fakeStore) GetSponsor(_ context.Context, s sponsor.Sponsor) (sponsor.Sponsor, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[s.ID]
	if !ok {
		return sponsor.Sponsor{}, fmt.Errorf("failed to get sponsor: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddSponsor(_ context.Context, s sponsor.Sponsor) (sponsor.Sponsor, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	s.ID = f.nextID
	f.rows[s.ID] = s
	return s, nil
}

func (f *fakeStore) DeleteSponsor(_ context.Context, s sponsor.Sponsor) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, s.ID)
	return nil
}

type fakeTeams map[int]team.Team

func (f fakeTeams) GetTeam(_ context.Context, t team.Team) (team.Team, error) {
	if x, ok := f[t.ID]; ok {
		return x, nil
	}
	return team.Team{}, fmt.Errorf("failed to get team: %w", sql.ErrNoRows)
}
```

`server/internal/sponsor/service_test.go`:
```go
package sponsor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

var ctx = context.Background()

func clubSponsor() sponsor.Sponsor {
	return sponsor.Sponsor{ID: 1, Name: "Club Sponsor", Website: null.StringFrom("https://s.example.test"),
		FileName: null.StringFrom("sponsor/club.png"), TeamID: "A"}
}

func newService(rows ...sponsor.Sponsor) (*sponsor.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore(rows...)
	objects := uploadtest.New()
	teams := fakeTeams{1: {ID: 1, Name: "First Team"}}
	return sponsor.NewService(store, teams, upload.New(objects)), store, objects
}

func png() *upload.File { return uploadtest.File("s.png", "image/png", "PNG") }

func fieldsOf(t *testing.T, err error) map[string]string {
	t.Helper()
	se, ok := svcerr.As(err)
	require.True(t, ok, "want svcerr, got %v", err)
	return se.Fields
}

func TestCreateValidates(t *testing.T) {
	svc, _, objects := newService()
	cases := map[string]struct {
		in    sponsor.CreateInput
		image *upload.File
		field string
	}{
		"name required":     {sponsor.CreateInput{Team: "A"}, png(), "name"},
		"image required":    {sponsor.CreateInput{Name: "x", Team: "A"}, nil, "image"},
		"bad website":       {sponsor.CreateInput{Name: "x", Website: "not a url", Team: "A"}, png(), "website"},
		"bad team code":     {sponsor.CreateInput{Name: "x", Team: "Z"}, png(), "team"},
		"unknown team id":   {sponsor.CreateInput{Name: "x", Team: "42"}, png(), "team"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := svc.Create(ctx, tc.in, tc.image)
			assert.Contains(t, fieldsOf(t, err), tc.field)
		})
	}
	assert.Empty(t, objects.Objects, "nothing uploaded when validation fails")
}

func TestCreate(t *testing.T) {
	svc, store, _ := newService()
	for _, teamValue := range []string{"A", "O", "Y", "1", ""} {
		s, err := svc.Create(ctx, sponsor.CreateInput{Name: "Kit Co", Website: "https://kit.example.test", Team: teamValue}, png())
		require.NoError(t, err, teamValue)
		assert.Equal(t, teamValue, store.rows[s.ID].TeamID)
		assert.NotEmpty(t, s.ImageURL)
	}
}

func TestForTeam(t *testing.T) {
	team1 := clubSponsor()
	team1.ID, team1.TeamID = 2, "1"
	svc, _, _ := newService(clubSponsor(), team1)
	got, err := svc.ForTeam(ctx, 1)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 2, got[0].ID)
}

func TestDelete(t *testing.T) {
	svc, store, objects := newService(clubSponsor())
	s, err := svc.Delete(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "Club Sponsor", s.Name)
	assert.Empty(t, store.rows)
	assert.Equal(t, []string{"sponsor/club.png"}, objects.Deleted)

	_, err = svc.Delete(ctx, 1)
	se, _ := svcerr.As(err)
	require.NotNil(t, se)
	assert.Equal(t, svcerr.KindNotFound, se.Kind)
}
```
Run: `go test ./server/internal/sponsor/`. Expected: compile errors.

- [ ] **Step 3: Implement `types.go` and `service.go`**

`server/internal/sponsor/types.go`:
```go
package sponsor

// Public is a sponsor as the API returns it. Team is "A", "O" or "Y"
// (legacy audience codes) or a team ID as a string; empty means unassigned.
type Public struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Website  string `json:"website,omitempty"`
	Purpose  string `json:"purpose,omitempty"`
	Team     string `json:"team"`
	ImageURL string `json:"imageUrl,omitempty"`
}

// CreateInput is a new sponsor; an image is required.
type CreateInput struct {
	Name    string
	Website string
	Purpose string
	Team    string
}
```

`server/internal/sponsor/service.go`:
```go
package sponsor

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

const uploadCategory = "sponsor"

type store interface {
	GetSponsors(ctx context.Context) ([]Sponsor, error)
	GetSponsorsMinimal(ctx context.Context) ([]Sponsor, error)
	GetSponsorsTeam(ctx context.Context, teamParam team.Team) ([]Sponsor, error)
	GetSponsor(ctx context.Context, sponsorParam Sponsor) (Sponsor, error)
	AddSponsor(ctx context.Context, sponsorParam Sponsor) (Sponsor, error)
	DeleteSponsor(ctx context.Context, sponsorParam Sponsor) error
}

// TeamGetter checks team IDs (satisfied by *team.Store).
type TeamGetter interface {
	GetTeam(ctx context.Context, teamParam team.Team) (team.Team, error)
}

// Service is the sponsor business logic shared by the API and legacy views.
type Service struct {
	store store
	teams TeamGetter
	files *upload.Files
}

func NewService(store store, teams TeamGetter, files *upload.Files) *Service {
	return &Service{store: store, teams: teams, files: files}
}

func (s *Service) sponsor(x Sponsor) Public {
	return Public{
		ID:       x.ID,
		Name:     x.Name,
		Website:  x.Website.String,
		Purpose:  x.Purpose.String,
		Team:     x.TeamID,
		ImageURL: s.files.URL(x.FileName.String),
	}
}

func (s *Service) many(rows []Sponsor) []Public {
	out := make([]Public, 0, len(rows))
	for _, r := range rows {
		out = append(out, s.sponsor(r))
	}
	return out
}

func (s *Service) List(ctx context.Context) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "sponsor.Service.List")
	defer span.End()
	rows, err := s.store.GetSponsors(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sponsors: %w", err)
	}
	return s.many(rows), nil
}

// ListMinimal is the lighter list shown on the home page.
func (s *Service) ListMinimal(ctx context.Context) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "sponsor.Service.ListMinimal")
	defer span.End()
	rows, err := s.store.GetSponsorsMinimal(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sponsors: %w", err)
	}
	return s.many(rows), nil
}

// ForTeam lists the sponsors attached to one team.
func (s *Service) ForTeam(ctx context.Context, teamID int) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "sponsor.Service.ForTeam")
	defer span.End()
	rows, err := s.store.GetSponsorsTeam(ctx, team.Team{ID: teamID})
	if err != nil {
		return nil, fmt.Errorf("failed to list team sponsors: %w", err)
	}
	return s.many(rows), nil
}

func (s *Service) Create(ctx context.Context, in CreateInput, image *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "sponsor.Service.Create")
	defer span.End()

	name := strings.TrimSpace(in.Name)
	fields := svcerr.Fields{}
	if name == "" {
		fields.Add("name", "name is required")
	}
	if image == nil {
		fields.Add("image", "image is required")
	}
	if in.Website != "" {
		if _, err := url.ParseRequestURI(in.Website); err != nil {
			fields.Add("website", "website must be a full URL, e.g. https://example.com")
		}
	}
	if msg := s.checkTeam(ctx, in.Team); msg != "" {
		fields.Add("team", msg)
	}
	if err := fields.Err(); err != nil {
		return Public{}, err
	}

	key, err := s.files.Save(ctx, image, uploadCategory)
	if err != nil {
		return Public{}, err
	}
	x, err := s.store.AddSponsor(ctx, Sponsor{
		Name:     name,
		Website:  null.NewString(in.Website, in.Website != ""),
		Purpose:  null.NewString(in.Purpose, in.Purpose != ""),
		FileName: null.StringFrom(key),
		TeamID:   in.Team,
	})
	if err != nil {
		s.files.Remove(ctx, key)
		return Public{}, fmt.Errorf("failed to add sponsor: %w", err)
	}
	return s.sponsor(x), nil
}

// checkTeam returns a validation message, or "" when value is acceptable.
func (s *Service) checkTeam(ctx context.Context, value string) string {
	switch value {
	case "", "A", "O", "Y":
		return ""
	}
	id, err := strconv.Atoi(value)
	if err != nil {
		return `team must be "A", "O", "Y" or a team ID`
	}
	if _, err = s.teams.GetTeam(ctx, team.Team{ID: id}); err != nil {
		return "team does not exist"
	}
	return ""
}

// Delete removes the sponsor and its logo, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Public, error) {
	ctx, span := tracer.Start(ctx, "sponsor.Service.Delete")
	defer span.End()
	x, err := s.store.GetSponsor(ctx, Sponsor{ID: id})
	if err != nil {
		return Public{}, svcerr.FromStore(err, "sponsor")
	}
	deleted := s.sponsor(x)
	if err = s.store.DeleteSponsor(ctx, x); err != nil {
		return Public{}, fmt.Errorf("failed to delete sponsor: %w", err)
	}
	s.files.Remove(ctx, x.FileName.String)
	return deleted, nil
}
```
Run: `go test ./server/internal/sponsor/`. Expected: PASS.

- [ ] **Step 4: Failing handler tests, then `handlers.go`**

`server/internal/sponsor/handlers_test.go`:
```go
package sponsor_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestSponsorRoutes(t *testing.T) {
	svc, store, _ := newService(clubSponsor())
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	sponsor.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	editor := anon.As(authtest.Cookie(t, sessions, authtest.Treasurer))
	manager := anon.As(authtest.Cookie(t, sessions, authtest.Manager))

	rec := anon.Get(t, "/api/v1/sponsors")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, apitest.Decode[[]sponsor.Public](t, rec), 1)

	logo := apitest.FilePart{Field: "image", Name: "l.png", ContentType: "image/png", Body: "PNG"}
	fields := map[string]string{"name": "Kit Co", "team": "O"}
	assert.Equal(t, http.StatusUnauthorized, anon.Multipart(t, http.MethodPost, "/api/v1/sponsors", fields, logo).Code)
	assert.Equal(t, http.StatusForbidden, manager.Multipart(t, http.MethodPost, "/api/v1/sponsors", fields, logo).Code)

	rec = editor.Multipart(t, http.MethodPost, "/api/v1/sponsors", fields, logo)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Equal(t, "O", apitest.Decode[sponsor.Public](t, rec).Team)

	rec = editor.Multipart(t, http.MethodPost, "/api/v1/sponsors", fields) // no image
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, apitest.ErrorOf(t, rec).Fields, "image")

	assert.Equal(t, http.StatusForbidden, manager.JSON(t, http.MethodDelete, "/api/v1/sponsors/1", nil).Code)
	assert.Equal(t, http.StatusNoContent, editor.JSON(t, http.MethodDelete, "/api/v1/sponsors/1", nil).Code)
	_, stillThere := store.rows[1]
	assert.False(t, stillThere)
}
```

`server/internal/sponsor/handlers.go`:
```go
package sponsor

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /sponsors.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the sponsor routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/sponsors", h.list)
	g.POST("/sponsors", h.create, guards.Editor)
	g.DELETE("/sponsors/:id", h.remove, guards.Editor)
}

// list returns every sponsor.
//
//	@Summary	List sponsors
//	@Tags		sponsors
//	@Produce	json
//	@Success	200	{array}	sponsor.Public
//	@Router		/sponsors [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// create adds a sponsor.
//
//	@Summary	Create a sponsor
//	@Tags		sponsors
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		name	formData	string	true	"Name"
//	@Param		website	formData	string	false	"Full URL"
//	@Param		purpose	formData	string	false	"What they sponsor"
//	@Param		team	formData	string	false	"A, O, Y or a team ID"
//	@Param		image	formData	file	true	"Logo"
//	@Success	201		{object}	sponsor.Public
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/sponsors [post]
func (h *Handlers) create(c echo.Context) error {
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	out, err := h.svc.Create(c.Request().Context(), CreateInput{
		Name:    c.FormValue("name"),
		Website: c.FormValue("website"),
		Purpose: c.FormValue("purpose"),
		Team:    c.FormValue("team"),
	}, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// remove deletes a sponsor.
//
//	@Summary	Delete a sponsor
//	@Tags		sponsors
//	@Param		id	path	int	true	"Sponsor ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/sponsors/{id} [delete]
func (h *Handlers) remove(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if _, err = h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
```
Run: `go test ./server/internal/sponsor/`. Expected: PASS.

- [ ] **Step 5: Point legacy sponsor writes at the service**

Add `sponsorSvc *sponsor.Service` to `Views`, `SponsorService *sponsor.Service` to `Deps`, and `sponsorSvc: d.SponsorService,` in `New`. In `server/internal/legacy/views/sponsor.go` keep `SponsorsFunc` and replace the two write handlers:
```go
func (v *Views) SponsorAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.SponsorAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for sponsor add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	created, err := v.sponsorSvc.Create(c.Request().Context(), sponsor.CreateInput{
		Name:    c.FormValue("name"),
		Website: c.FormValue("website"),
		Purpose: c.FormValue("purpose"),
		Team:    c.FormValue("team"),
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add sponsor for sponsor add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add sponsor for sponsor add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", created.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) SponsorDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.SponsorDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for sponsor delete, error: %w", err)
	}
	deleted, err := v.sponsorSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete sponsor for sponsor delete, sponsor id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", deleted.Name))
	return c.Redirect(http.StatusFound, "/sponsors")
}
```
Run `goimports -w server/internal/legacy/views`.

- [ ] **Step 6: Wire into `app.Build`, regenerate docs, run, commit**

- Add `sponsorSvc := sponsor.NewService(s.Sponsor, s.Team, uploads)` next to the other services.
- Add `SponsorService: sponsorSvc,` to `views.Deps`.
- Add `sponsor.NewHandlers(sponsorSvc).Register(api, guards)` after the other handlers.
```bash
go generate ./server/internal/docs
gofmt -l server; go build ./... && go vet ./... && go test ./... && go tool golangci-lint run ./...
git add -A server
git commit -m "Add sponsor service and /api/v1/sponsors; legacy writes use the service

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---
### Task 11: Affiliations

The same shape as Task 10 (create, list, delete; image required), with no team field. The API type is `affiliation.Public`.

**Files:**
- Modify: `server/internal/affiliation/store.go` (`addAffiliation` returns ID)
- Create: `server/internal/affiliation/{store_test,types,service,handlers,fake_test,service_test,handlers_test}.go`
- Modify: `server/internal/legacy/views/affilitaion.go` (sic: the file name is misspelled in the repo), `views.go`, `server/internal/app/app.go`

**Interfaces:**
- Produces (`affiliation`):
  - Types: `Public{ID int; Name, Website, ImageURL string}` and `CreateInput{Name, Website string}`.
  - `NewService(store, *upload.Files) *Service`, with methods:
    - `List(ctx) ([]Public, error)`
    - `ListMinimal(ctx) ([]Public, error)`
    - `Create(ctx, CreateInput, *upload.File) (Public, error)`
    - `Delete(ctx, id) (Public, error)`
  - `NewHandlers`, `Register` (`GET /affiliations`, `POST /affiliations` ✏️, `DELETE /affiliations/:id` ✏️).
- Produces (legacy): `Deps.AffiliationService`.

- [ ] **Step 1: Store test, then fix `addAffiliation`**

`server/internal/affiliation/store_test.go`:
```go
package affiliation_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestAddAffiliationReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	added, err := affiliation.NewAffiliationRepo(db).AddAffiliation(context.Background(), affiliation.Affiliation{Name: "League"})
	require.NoError(t, err)
	require.Positive(t, added.ID)
}
```
Run (FAIL), then in `addAffiliation` add `.Suffix("RETURNING id")` and replace the `ExecContext`/`RowsAffected` block with:
```go
	err = s.db.GetContext(ctx, &affiliationParam.ID, sql, args...)
	if err != nil {
		span.RecordError(err)
		return Affiliation{}, fmt.Errorf("failed to add affiliation: %w", err)
	}
	return affiliationParam, nil
```
Re-run: PASS.

- [ ] **Step 2: Fake store and failing tests**

`server/internal/affiliation/fake_test.go`:
```go
package affiliation_test

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
)

type fakeStore struct {
	mu     sync.Mutex
	rows   map[int]affiliation.Affiliation
	nextID int
}

func newFakeStore(rows ...affiliation.Affiliation) *fakeStore {
	f := &fakeStore{rows: map[int]affiliation.Affiliation{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) GetAffiliations(context.Context) ([]affiliation.Affiliation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]affiliation.Affiliation, 0, len(f.rows))
	for _, r := range f.rows {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (f *fakeStore) GetAffiliationsMinimal(ctx context.Context) ([]affiliation.Affiliation, error) {
	return f.GetAffiliations(ctx)
}

func (f *fakeStore) GetAffiliation(_ context.Context, a affiliation.Affiliation) (affiliation.Affiliation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[a.ID]
	if !ok {
		return affiliation.Affiliation{}, fmt.Errorf("failed to get affiliation: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddAffiliation(_ context.Context, a affiliation.Affiliation) (affiliation.Affiliation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	a.ID = f.nextID
	f.rows[a.ID] = a
	return a, nil
}

func (f *fakeStore) DeleteAffiliation(_ context.Context, a affiliation.Affiliation) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, a.ID)
	return nil
}
```

`server/internal/affiliation/service_test.go`:
```go
package affiliation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

var ctx = context.Background()

func countyFA() affiliation.Affiliation {
	return affiliation.Affiliation{ID: 1, Name: "County FA", Website: null.StringFrom("https://fa.example.test"),
		FileName: null.StringFrom("affiliation/fa.png")}
}

func newService(rows ...affiliation.Affiliation) (*affiliation.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore(rows...)
	objects := uploadtest.New()
	return affiliation.NewService(store, upload.New(objects)), store, objects
}

func TestCreateValidates(t *testing.T) {
	svc, _, objects := newService()
	img := uploadtest.File("a.png", "image/png", "PNG")
	for name, tc := range map[string]struct {
		in    affiliation.CreateInput
		image *upload.File
		field string
	}{
		"name":    {affiliation.CreateInput{}, img, "name"},
		"image":   {affiliation.CreateInput{Name: "x"}, nil, "image"},
		"website": {affiliation.CreateInput{Name: "x", Website: "fa dot com"}, img, "website"},
	} {
		_, err := svc.Create(ctx, tc.in, tc.image)
		se, ok := svcerr.As(err)
		require.True(t, ok, name)
		assert.Contains(t, se.Fields, tc.field, name)
	}
	assert.Empty(t, objects.Objects)
}

func TestCreateAndDelete(t *testing.T) {
	svc, store, objects := newService(countyFA())
	a, err := svc.Create(ctx, affiliation.CreateInput{Name: "League", Website: "https://l.example.test"},
		uploadtest.File("l.png", "image/png", "PNG"))
	require.NoError(t, err)
	assert.NotEmpty(t, a.ImageURL)

	deleted, err := svc.Delete(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "County FA", deleted.Name)
	assert.Equal(t, []string{"affiliation/fa.png"}, objects.Deleted)
	assert.Len(t, store.rows, 1)
}

func TestAffiliationRoutes(t *testing.T) {
	svc, _, _ := newService(countyFA())
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	affiliation.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	editor := anon.As(authtest.Cookie(t, sessions, authtest.Webmaster))
	manager := anon.As(authtest.Cookie(t, sessions, authtest.Manager))

	rec := anon.Get(t, "/api/v1/affiliations")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, apitest.Decode[[]affiliation.Public](t, rec), 1)

	logo := apitest.FilePart{Field: "image", Name: "l.png", ContentType: "image/png", Body: "PNG"}
	assert.Equal(t, http.StatusForbidden, manager.Multipart(t, http.MethodPost, "/api/v1/affiliations", map[string]string{"name": "x"}, logo).Code)
	assert.Equal(t, http.StatusCreated, editor.Multipart(t, http.MethodPost, "/api/v1/affiliations", map[string]string{"name": "x"}, logo).Code)
	assert.Equal(t, http.StatusNoContent, editor.JSON(t, http.MethodDelete, "/api/v1/affiliations/1", nil).Code)
	assert.Equal(t, http.StatusNotFound, editor.JSON(t, http.MethodDelete, "/api/v1/affiliations/1", nil).Code)
}
```
Run: `go test ./server/internal/affiliation/`. Expected: compile errors.

- [ ] **Step 3: Implement**

`server/internal/affiliation/types.go`:
```go
package affiliation

// Public is an affiliation as the API returns it.
type Public struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Website  string `json:"website,omitempty"`
	ImageURL string `json:"imageUrl,omitempty"`
}

// CreateInput is a new affiliation; an image is required.
type CreateInput struct {
	Name    string
	Website string
}
```

`server/internal/affiliation/service.go`:
```go
package affiliation

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

const uploadCategory = "affiliation"

type store interface {
	GetAffiliations(ctx context.Context) ([]Affiliation, error)
	GetAffiliationsMinimal(ctx context.Context) ([]Affiliation, error)
	GetAffiliation(ctx context.Context, affiliationParam Affiliation) (Affiliation, error)
	AddAffiliation(ctx context.Context, affiliationParam Affiliation) (Affiliation, error)
	DeleteAffiliation(ctx context.Context, affiliationParam Affiliation) error
}

// Service is the affiliation business logic shared by the API and legacy views.
type Service struct {
	store store
	files *upload.Files
}

func NewService(store store, files *upload.Files) *Service {
	return &Service{store: store, files: files}
}

func (s *Service) public(a Affiliation) Public {
	return Public{ID: a.ID, Name: a.Name, Website: a.Website.String, ImageURL: s.files.URL(a.FileName.String)}
}

func (s *Service) many(rows []Affiliation) []Public {
	out := make([]Public, 0, len(rows))
	for _, r := range rows {
		out = append(out, s.public(r))
	}
	return out
}

func (s *Service) List(ctx context.Context) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "affiliation.Service.List")
	defer span.End()
	rows, err := s.store.GetAffiliations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list affiliations: %w", err)
	}
	return s.many(rows), nil
}

// ListMinimal is the lighter list shown on the home page.
func (s *Service) ListMinimal(ctx context.Context) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "affiliation.Service.ListMinimal")
	defer span.End()
	rows, err := s.store.GetAffiliationsMinimal(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list affiliations: %w", err)
	}
	return s.many(rows), nil
}

func (s *Service) Create(ctx context.Context, in CreateInput, image *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "affiliation.Service.Create")
	defer span.End()

	name := strings.TrimSpace(in.Name)
	fields := svcerr.Fields{}
	if name == "" {
		fields.Add("name", "name is required")
	}
	if image == nil {
		fields.Add("image", "image is required")
	}
	if in.Website != "" {
		if _, err := url.ParseRequestURI(in.Website); err != nil {
			fields.Add("website", "website must be a full URL, e.g. https://example.com")
		}
	}
	if err := fields.Err(); err != nil {
		return Public{}, err
	}

	key, err := s.files.Save(ctx, image, uploadCategory)
	if err != nil {
		return Public{}, err
	}
	a, err := s.store.AddAffiliation(ctx, Affiliation{
		Name:     name,
		Website:  null.NewString(in.Website, in.Website != ""),
		FileName: null.StringFrom(key),
	})
	if err != nil {
		s.files.Remove(ctx, key)
		return Public{}, fmt.Errorf("failed to add affiliation: %w", err)
	}
	return s.public(a), nil
}

// Delete removes the affiliation and its logo, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Public, error) {
	ctx, span := tracer.Start(ctx, "affiliation.Service.Delete")
	defer span.End()
	a, err := s.store.GetAffiliation(ctx, Affiliation{ID: id})
	if err != nil {
		return Public{}, svcerr.FromStore(err, "affiliation")
	}
	deleted := s.public(a)
	if err = s.store.DeleteAffiliation(ctx, a); err != nil {
		return Public{}, fmt.Errorf("failed to delete affiliation: %w", err)
	}
	s.files.Remove(ctx, a.FileName.String)
	return deleted, nil
}
```
`tracer` is the package-level tracer already declared in `affiliation.go`.

`server/internal/affiliation/handlers.go`:
```go
package affiliation

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /affiliations.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the affiliation routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/affiliations", h.list)
	g.POST("/affiliations", h.create, guards.Editor)
	g.DELETE("/affiliations/:id", h.remove, guards.Editor)
}

// list returns every affiliation.
//
//	@Summary	List affiliations
//	@Tags		affiliations
//	@Produce	json
//	@Success	200	{array}	affiliation.Public
//	@Router		/affiliations [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// create adds an affiliation.
//
//	@Summary	Create an affiliation
//	@Tags		affiliations
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		name	formData	string	true	"Name"
//	@Param		website	formData	string	false	"Full URL"
//	@Param		image	formData	file	true	"Logo"
//	@Success	201		{object}	affiliation.Public
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/affiliations [post]
func (h *Handlers) create(c echo.Context) error {
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	out, err := h.svc.Create(c.Request().Context(), CreateInput{
		Name:    c.FormValue("name"),
		Website: c.FormValue("website"),
	}, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// remove deletes an affiliation.
//
//	@Summary	Delete an affiliation
//	@Tags		affiliations
//	@Param		id	path	int	true	"Affiliation ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/affiliations/{id} [delete]
func (h *Handlers) remove(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if _, err = h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
```
Run: `go test ./server/internal/affiliation/`. Expected: PASS.

- [ ] **Step 4: Legacy writes → service**

Add `affiliationSvc *affiliation.Service` to `Views`, `AffiliationService *affiliation.Service` to `Deps`, and `affiliationSvc: d.AffiliationService,` in `New`. Replace both handlers in `server/internal/legacy/views/affilitaion.go`:
```go
func (v *Views) AffiliationAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.AffiliationAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for affiliation add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	created, err := v.affiliationSvc.Create(c.Request().Context(), affiliation.CreateInput{
		Name:    c.FormValue("name"),
		Website: c.FormValue("website"),
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add affiliation for affiliation add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add affiliation for affiliation add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", created.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) AffiliationDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.AffiliationDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for affiliation delete, error: %w", err)
	}
	deleted, err := v.affiliationSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete affiliation for affiliation delete, affiliation id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", deleted.Name))
	return c.Redirect(http.StatusFound, "/")
}
```

- [ ] **Step 5: Wire, regenerate, run, commit**

In `app.Build`:
- Add `affiliationSvc := affiliation.NewService(s.Affiliation, uploads)`.
- Add `AffiliationService: affiliationSvc,` to `views.Deps`.
- Add `affiliation.NewHandlers(affiliationSvc).Register(api, guards)`.
```bash
go generate ./server/internal/docs
goimports -w server/internal/legacy/views
gofmt -l server; go build ./... && go vet ./... && go test ./... && go tool golangci-lint run ./...
git add -A server
git commit -m "Add affiliation service and /api/v1/affiliations; legacy writes use the service

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 12: Documents

Create, list and delete a named file (PDF, DOCX, etc.). The API type is `document.Public` with `fileUrl`. The upload field is `file`, and it is required.

**Files:**
- Modify: `server/internal/document/store.go` (`addDocument` returns ID)
- Create: `server/internal/document/{store_test,types,service,handlers,fake_test,service_test}.go`
- Modify: `server/internal/legacy/views/document.go`, `views.go`, `server/internal/app/app.go`

**Interfaces:**
- Produces (`document`):
  - Types: `Public{ID int; Name, FileURL string}` and `CreateInput{Name string}`.
  - `NewService(store, *upload.Files) *Service`, with methods `List(ctx) ([]Public, error)`, `Create(ctx, CreateInput, *upload.File) (Public, error)` and `Delete(ctx, id) (Public, error)`.
  - `NewHandlers`, `Register` (`GET /documents`, `POST /documents` ✏️, `DELETE /documents/:id` ✏️).
- Produces (legacy): `Deps.DocumentService`.

- [ ] **Step 1: Store test, then fix `addDocument`**

`server/internal/document/store_test.go`:
```go
package document_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/document"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestAddDocumentReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	added, err := document.NewDocumentRepo(db).AddDocument(context.Background(), document.Document{Name: "Policy", FileName: "document/p.pdf"})
	require.NoError(t, err)
	require.Positive(t, added.ID)
}
```
Run (FAIL). In `addDocument` add `.Suffix("RETURNING id")` and replace the `ExecContext`/`RowsAffected` block with:
```go
	err = s.db.GetContext(ctx, &documentParam.ID, sql, args...)
	if err != nil {
		span.RecordError(err)
		return Document{}, fmt.Errorf("failed to add document: %w", err)
	}
	return documentParam, nil
```
Re-run: PASS.

- [ ] **Step 2: Fake store and failing tests**

`server/internal/document/fake_test.go`:
```go
package document_test

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/document"
)

type fakeStore struct {
	mu     sync.Mutex
	rows   map[int]document.Document
	nextID int
}

func newFakeStore(rows ...document.Document) *fakeStore {
	f := &fakeStore{rows: map[int]document.Document{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) GetDocuments(context.Context) ([]document.Document, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]document.Document, 0, len(f.rows))
	for _, r := range f.rows {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (f *fakeStore) GetDocument(_ context.Context, d document.Document) (document.Document, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[d.ID]
	if !ok {
		return document.Document{}, fmt.Errorf("failed to get document: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddDocument(_ context.Context, d document.Document) (document.Document, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	d.ID = f.nextID
	f.rows[d.ID] = d
	return d, nil
}

func (f *fakeStore) DeleteDocument(_ context.Context, d document.Document) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, d.ID)
	return nil
}
```

`server/internal/document/service_test.go`:
```go
package document_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/document"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

var ctx = context.Background()

func rules() document.Document {
	return document.Document{ID: 1, Name: "Club Rules", FileName: "document/rules.pdf"}
}

func newService(rows ...document.Document) (*document.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore(rows...)
	objects := uploadtest.New()
	return document.NewService(store, upload.New(objects)), store, objects
}

func TestCreateValidates(t *testing.T) {
	svc, _, _ := newService()
	_, err := svc.Create(ctx, document.CreateInput{}, uploadtest.File("r.pdf", "application/pdf", "PDF"))
	se, ok := svcerr.As(err)
	require.True(t, ok)
	assert.Contains(t, se.Fields, "name")

	_, err = svc.Create(ctx, document.CreateInput{Name: "x"}, nil)
	se, ok = svcerr.As(err)
	require.True(t, ok)
	assert.Contains(t, se.Fields, "file")
}

func TestCreateAndDelete(t *testing.T) {
	svc, store, objects := newService(rules())
	d, err := svc.Create(ctx, document.CreateInput{Name: "Minutes"}, uploadtest.File("m.pdf", "application/pdf", "PDF"))
	require.NoError(t, err)
	assert.Contains(t, d.FileURL, "https://cdn.test/document/")

	deleted, err := svc.Delete(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "Club Rules", deleted.Name)
	assert.Equal(t, []string{"document/rules.pdf"}, objects.Deleted)
	assert.Len(t, store.rows, 1)
}

func TestDocumentRoutes(t *testing.T) {
	svc, _, _ := newService(rules())
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	document.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	editor := anon.As(authtest.Cookie(t, sessions, authtest.Webmaster))
	photographer := anon.As(authtest.Cookie(t, sessions, authtest.Photographer))

	rec := anon.Get(t, "/api/v1/documents")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, apitest.Decode[[]document.Public](t, rec), 1)

	pdf := apitest.FilePart{Field: "file", Name: "m.pdf", ContentType: "application/pdf", Body: "PDF"}
	assert.Equal(t, http.StatusForbidden, photographer.Multipart(t, http.MethodPost, "/api/v1/documents", map[string]string{"name": "x"}, pdf).Code)
	assert.Equal(t, http.StatusCreated, editor.Multipart(t, http.MethodPost, "/api/v1/documents", map[string]string{"name": "x"}, pdf).Code)
	assert.Equal(t, http.StatusNoContent, editor.JSON(t, http.MethodDelete, "/api/v1/documents/1", nil).Code)
}
```
Run: compile errors expected.

- [ ] **Step 3: Implement**

`server/internal/document/types.go`:
```go
package document

// Public is a downloadable club document as the API returns it.
type Public struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	FileURL string `json:"fileUrl"`
}

// CreateInput is a new document; a file is required.
type CreateInput struct {
	Name string
}
```

`server/internal/document/service.go`:
```go
package document

import (
	"context"
	"fmt"
	"strings"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

const uploadCategory = "document"

type store interface {
	GetDocuments(ctx context.Context) ([]Document, error)
	GetDocument(ctx context.Context, documentParam Document) (Document, error)
	AddDocument(ctx context.Context, documentParam Document) (Document, error)
	DeleteDocument(ctx context.Context, documentParam Document) error
}

// Service is the document business logic shared by the API and legacy views.
type Service struct {
	store store
	files *upload.Files
}

func NewService(store store, files *upload.Files) *Service {
	return &Service{store: store, files: files}
}

func (s *Service) public(d Document) Public {
	return Public{ID: d.ID, Name: d.Name, FileURL: s.files.URL(d.FileName)}
}

func (s *Service) List(ctx context.Context) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "document.Service.List")
	defer span.End()
	rows, err := s.store.GetDocuments(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	out := make([]Public, 0, len(rows))
	for _, d := range rows {
		out = append(out, s.public(d))
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, in CreateInput, file *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "document.Service.Create")
	defer span.End()

	name := strings.TrimSpace(in.Name)
	fields := svcerr.Fields{}
	if name == "" {
		fields.Add("name", "name is required")
	}
	if file == nil {
		fields.Add("file", "file is required")
	}
	if err := fields.Err(); err != nil {
		return Public{}, err
	}
	key, err := s.files.Save(ctx, file, uploadCategory)
	if err != nil {
		return Public{}, err
	}
	d, err := s.store.AddDocument(ctx, Document{Name: name, FileName: key})
	if err != nil {
		s.files.Remove(ctx, key)
		return Public{}, fmt.Errorf("failed to add document: %w", err)
	}
	return s.public(d), nil
}

// Delete removes the document and its file, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Public, error) {
	ctx, span := tracer.Start(ctx, "document.Service.Delete")
	defer span.End()
	d, err := s.store.GetDocument(ctx, Document{ID: id})
	if err != nil {
		return Public{}, svcerr.FromStore(err, "document")
	}
	deleted := s.public(d)
	if err = s.store.DeleteDocument(ctx, d); err != nil {
		return Public{}, fmt.Errorf("failed to delete document: %w", err)
	}
	s.files.Remove(ctx, d.FileName)
	return deleted, nil
}
```
(`tracer` is already declared in `document.go`.)

`server/internal/document/handlers.go`:
```go
package document

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /documents.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the document routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/documents", h.list)
	g.POST("/documents", h.create, guards.Editor)
	g.DELETE("/documents/:id", h.remove, guards.Editor)
}

// list returns every document, by name.
//
//	@Summary	List documents
//	@Tags		documents
//	@Produce	json
//	@Success	200	{array}	document.Public
//	@Router		/documents [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// create uploads a document.
//
//	@Summary	Upload a document
//	@Tags		documents
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		name	formData	string	true	"Display name"
//	@Param		file	formData	file	true	"PDF, DOCX, PPTX, TXT or image"
//	@Success	201		{object}	document.Public
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/documents [post]
func (h *Handlers) create(c echo.Context) error {
	file, err := web.FormFile(c, "file")
	if err != nil {
		return err
	}
	out, err := h.svc.Create(c.Request().Context(), CreateInput{Name: c.FormValue("name")}, file)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// remove deletes a document.
//
//	@Summary	Delete a document
//	@Tags		documents
//	@Param		id	path	int	true	"Document ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/documents/{id} [delete]
func (h *Handlers) remove(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if _, err = h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
```
Run: `go test ./server/internal/document/`. Expected: PASS.

- [ ] **Step 4: Legacy writes → service**

Add `documentSvc` / `DocumentService` to `Views`/`Deps`/`New`. In `server/internal/legacy/views/document.go` keep `DocumentsFunc` and replace:
```go
func (v *Views) DocumentAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.DocumentAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}
	file, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for document add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	created, err := v.documentSvc.Create(c.Request().Context(), document.CreateInput{Name: c.FormValue("name")}, file)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add document for document add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add document for document add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", created.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) DocumentDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.DocumentDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for document delete, error: %w", err)
	}
	deleted, err := v.documentSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete document for document delete, document id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", deleted.Name))
	return c.Redirect(http.StatusFound, "/documents")
}
```

- [ ] **Step 5: Wire, regenerate, run, commit**

In `app.Build`:
- Add `documentSvc := document.NewService(s.Document, uploads)`.
- Add `DocumentService: documentSvc,` to `views.Deps`.
- Add `document.NewHandlers(documentSvc).Register(api, guards)`.
```bash
go generate ./server/internal/docs
goimports -w server/internal/legacy/views
gofmt -l server; go build ./... && go vet ./... && go test ./... && go tool golangci-lint run ./...
git add -A server
git commit -m "Add document service and /api/v1/documents; legacy writes use the service

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 13: Gallery

Photos with an optional caption. The difference from the others: writes need **`NotManager`**, so Photographers may upload. The Go package is `image`, and the API path is `/gallery`.

**Files:**
- Modify: `server/internal/image/store.go` (`addImage` returns ID)
- Create: `server/internal/image/{store_test,types,service,handlers,fake_test,service_test}.go`
- Modify: `server/internal/legacy/views/gallery.go`, `views.go`, `server/internal/app/app.go`

**Interfaces:**
- Produces (`image`):
  - Types: `Public{ID int; Caption, ImageURL string}` and `CreateInput{Caption string}`.
  - `NewService(store, *upload.Files) *Service`, with methods `List(ctx) ([]Public, error)`, `Create(ctx, CreateInput, *upload.File) (Public, error)` and `Delete(ctx, id) (Public, error)`.
  - `NewHandlers`, `Register` (`GET /gallery`, `POST /gallery` 📷, `DELETE /gallery/:id` 📷).
- Produces (legacy): `Deps.GalleryService *image.Service`.

- [ ] **Step 1: Store test, then fix `addImage`**

`server/internal/image/store_test.go`:
```go
package image_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/image"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestAddImageReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	added, err := image.NewImageRepo(db).AddImage(context.Background(), image.Image{FileName: "gallery/x.jpg"})
	require.NoError(t, err)
	require.Positive(t, added.ID)
}
```
Run (FAIL). In `addImage` add `.Suffix("RETURNING id")` and replace the `ExecContext`/`RowsAffected` block with:
```go
	err = s.db.GetContext(ctx, &imageParam.ID, sql, args...)
	if err != nil {
		span.RecordError(err)
		return Image{}, fmt.Errorf("failed to add image: %w", err)
	}
	return imageParam, nil
```
Re-run: PASS.

- [ ] **Step 2: Fake store and failing tests**

`server/internal/image/fake_test.go`:
```go
package image_test

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/image"
)

type fakeStore struct {
	mu     sync.Mutex
	rows   map[int]image.Image
	nextID int
}

func newFakeStore(rows ...image.Image) *fakeStore {
	f := &fakeStore{rows: map[int]image.Image{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) GetImages(context.Context) ([]image.Image, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]image.Image, 0, len(f.rows))
	for _, r := range f.rows {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (f *fakeStore) GetImage(_ context.Context, i image.Image) (image.Image, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[i.ID]
	if !ok {
		return image.Image{}, fmt.Errorf("failed to get image: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddImage(_ context.Context, i image.Image) (image.Image, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	i.ID = f.nextID
	f.rows[i.ID] = i
	return i, nil
}

func (f *fakeStore) DeleteImage(_ context.Context, i image.Image) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, i.ID)
	return nil
}
```

`server/internal/image/service_test.go`:
```go
package image_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/image"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

var ctx = context.Background()

func matchDay() image.Image {
	return image.Image{ID: 1, FileName: "gallery/one.jpg", Caption: null.StringFrom("Match day")}
}

func newService(rows ...image.Image) (*image.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore(rows...)
	objects := uploadtest.New()
	return image.NewService(store, upload.New(objects)), store, objects
}

func TestCreateRequiresImage(t *testing.T) {
	svc, _, _ := newService()
	_, err := svc.Create(ctx, image.CreateInput{Caption: "x"}, nil)
	se, ok := svcerr.As(err)
	require.True(t, ok)
	assert.Contains(t, se.Fields, "image")
}

func TestCreateAndDelete(t *testing.T) {
	svc, store, objects := newService(matchDay())
	p, err := svc.Create(ctx, image.CreateInput{}, uploadtest.File("p.jpg", "image/jpeg", "JPG"))
	require.NoError(t, err)
	assert.Empty(t, p.Caption, "caption is optional")
	assert.Contains(t, p.ImageURL, "https://cdn.test/gallery/")

	_, err = svc.Delete(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, []string{"gallery/one.jpg"}, objects.Deleted)
	assert.Len(t, store.rows, 1)
}

func TestGalleryRoutesAllowPhotographers(t *testing.T) {
	svc, _, _ := newService(matchDay())
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	image.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	photographer := anon.As(authtest.Cookie(t, sessions, authtest.Photographer))
	manager := anon.As(authtest.Cookie(t, sessions, authtest.Manager))

	rec := anon.Get(t, "/api/v1/gallery")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Match day", apitest.Decode[[]image.Public](t, rec)[0].Caption)

	jpg := apitest.FilePart{Field: "image", Name: "p.jpg", ContentType: "image/jpeg", Body: "JPG"}
	assert.Equal(t, http.StatusForbidden, manager.Multipart(t, http.MethodPost, "/api/v1/gallery", nil, jpg).Code)
	assert.Equal(t, http.StatusCreated, photographer.Multipart(t, http.MethodPost, "/api/v1/gallery", map[string]string{"caption": "Goal"}, jpg).Code)
	assert.Equal(t, http.StatusForbidden, manager.JSON(t, http.MethodDelete, "/api/v1/gallery/1", nil).Code)
	assert.Equal(t, http.StatusNoContent, photographer.JSON(t, http.MethodDelete, "/api/v1/gallery/1", nil).Code)
}
```
Run: compile errors expected.

- [ ] **Step 3: Implement**

`server/internal/image/types.go`:
```go
package image

// Public is a gallery photo as the API returns it.
type Public struct {
	ID       int    `json:"id"`
	Caption  string `json:"caption,omitempty"`
	ImageURL string `json:"imageUrl"`
}

// CreateInput is a new gallery photo; the image is required.
type CreateInput struct {
	Caption string
}
```

`server/internal/image/service.go`:
```go
package image

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

const uploadCategory = "gallery"

type store interface {
	GetImages(ctx context.Context) ([]Image, error)
	GetImage(ctx context.Context, imageParam Image) (Image, error)
	AddImage(ctx context.Context, imageParam Image) (Image, error)
	DeleteImage(ctx context.Context, imageParam Image) error
}

// Service is the gallery business logic shared by the API and legacy views.
type Service struct {
	store store
	files *upload.Files
}

func NewService(store store, files *upload.Files) *Service {
	return &Service{store: store, files: files}
}

func (s *Service) public(i Image) Public {
	return Public{ID: i.ID, Caption: i.Caption.String, ImageURL: s.files.URL(i.FileName)}
}

func (s *Service) List(ctx context.Context) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "image.Service.List")
	defer span.End()
	rows, err := s.store.GetImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list gallery: %w", err)
	}
	out := make([]Public, 0, len(rows))
	for _, i := range rows {
		out = append(out, s.public(i))
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, in CreateInput, photo *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "image.Service.Create")
	defer span.End()
	if photo == nil {
		return Public{}, svcerr.InvalidField("image", "image is required")
	}
	key, err := s.files.Save(ctx, photo, uploadCategory)
	if err != nil {
		return Public{}, err
	}
	caption := strings.TrimSpace(in.Caption)
	i, err := s.store.AddImage(ctx, Image{FileName: key, Caption: null.NewString(caption, caption != "")})
	if err != nil {
		s.files.Remove(ctx, key)
		return Public{}, fmt.Errorf("failed to add gallery image: %w", err)
	}
	return s.public(i), nil
}

// Delete removes the photo and its file, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Public, error) {
	ctx, span := tracer.Start(ctx, "image.Service.Delete")
	defer span.End()
	i, err := s.store.GetImage(ctx, Image{ID: id})
	if err != nil {
		return Public{}, svcerr.FromStore(err, "gallery image")
	}
	deleted := s.public(i)
	if err = s.store.DeleteImage(ctx, i); err != nil {
		return Public{}, fmt.Errorf("failed to delete gallery image: %w", err)
	}
	s.files.Remove(ctx, i.FileName)
	return deleted, nil
}
```
(`tracer` is already declared in `image.go`.)

`server/internal/image/handlers.go`:
```go
package image

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /gallery.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the gallery routes on the /api/v1 group. Photographers may
// write; Managers may not.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/gallery", h.list)
	g.POST("/gallery", h.create, guards.NotManager)
	g.DELETE("/gallery/:id", h.remove, guards.NotManager)
}

// list returns every gallery photo.
//
//	@Summary	List gallery photos
//	@Tags		gallery
//	@Produce	json
//	@Success	200	{array}	image.Public
//	@Router		/gallery [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// create uploads a photo.
//
//	@Summary	Upload a gallery photo
//	@Tags		gallery
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		caption	formData	string	false	"Caption"
//	@Param		image	formData	file	true	"Photo"
//	@Success	201		{object}	image.Public
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/gallery [post]
func (h *Handlers) create(c echo.Context) error {
	photo, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	out, err := h.svc.Create(c.Request().Context(), CreateInput{Caption: c.FormValue("caption")}, photo)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// remove deletes a photo.
//
//	@Summary	Delete a gallery photo
//	@Tags		gallery
//	@Param		id	path	int	true	"Photo ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/gallery/{id} [delete]
func (h *Handlers) remove(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if _, err = h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
```
Run: `go test ./server/internal/image/`. Expected: PASS.

- [ ] **Step 4: Legacy writes → service**

Add `gallerySvc *image.Service` / `GalleryService *image.Service` to `Views`/`Deps`/`New`. In `server/internal/legacy/views/gallery.go` keep `GalleryFunc` and replace:
```go
func (v *Views) ImageAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ImageAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}
	photo, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for image add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	if _, err = v.gallerySvc.Create(c.Request().Context(), image.CreateInput{Caption: c.FormValue("caption")}, photo); err != nil {
		slog.Info(fmt.Sprintf("failed to add image for image add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add image for image add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, "successfully added image")
	return c.JSON(http.StatusOK, data)
}

func (v *Views) ImageDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ImageDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for image delete, error: %w", err)
	}
	if _, err = v.gallerySvc.Delete(c.Request().Context(), id); err != nil {
		return fmt.Errorf("failed to delete image for image delete, image id: %d, error: %w", id, err)
	}
	v.flash(c, c1, "successfully deleted image")
	return c.Redirect(http.StatusFound, "/gallery")
}
```

- [ ] **Step 5: Wire, regenerate, run, commit**

In `app.Build`:
- Add `gallerySvc := image.NewService(s.Image, uploads)`.
- Add `GalleryService: gallerySvc,` to `views.Deps`.
- Add `image.NewHandlers(gallerySvc).Register(api, guards)`.
```bash
go generate ./server/internal/docs
goimports -w server/internal/legacy/views
gofmt -l server; go build ./... && go vet ./... && go test ./... && go tool golangci-lint run ./...
git add -A server
git commit -m "Add gallery service and /api/v1/gallery; legacy writes use the service

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---
### Task 14: Programmes and seasons

**Files:**
- Modify: `server/internal/programme/store.go` (`addProgramme` and `addSeason` return IDs)
- Create: `server/internal/programme/{store_test,types,service,handlers,fake_test,service_test,handlers_test}.go`
- Modify: `server/internal/legacy/views/programme.go`, `views.go`, `server/internal/app/app.go`

**Interfaces:**
- Produces (`programme`):
  - Types:
    - `Public{ID int; Name string; Date time.Time; FileURL string; Season *PublicSeason}` (JSON `date`, `fileUrl`, `season`)
    - `PublicSeason{ID int; Name string}`
    - `CreateInput{Name string; Date time.Time; SeasonID int}`
  - `NewService(store, *upload.Files) *Service`, with methods:
    - `List(ctx, seasonID int) ([]Public, error)` (0 means all; an unknown season is NotFound)
    - `Create(ctx, CreateInput, *upload.File) (Public, error)` (file required; `SeasonID` 0 or an existing season)
    - `Delete(ctx, id) (Public, error)`
    - `Seasons(ctx) ([]PublicSeason, error)`
    - `CreateSeason(ctx, name string) (PublicSeason, error)`
    - `RenameSeason(ctx, id int, name string) (PublicSeason, error)`
    - `DeleteSeason(ctx, id int) (PublicSeason, error)` (unlinks its programmes first, as legacy does)
  - `NewHandlers`, `Register`.
- Produces (legacy): `Deps.ProgrammeService`.

- [ ] **Step 1: Store tests, then fix both inserts**

`server/internal/programme/store_test.go`:
```go
package programme_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/programme"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestInsertsReturnIDs(t *testing.T) {
	db, _ := testdb.Open(t)
	store := programme.NewProgrammeRepo(db)
	ctx := context.Background()

	season, err := store.AddSeason(ctx, programme.Season{Season: "2026-27"})
	require.NoError(t, err)
	require.Positive(t, season.ID)

	p, err := store.AddProgramme(ctx, programme.Programme{Name: "Home v Away", FileName: "programme/p.pdf",
		DateOfProgramme: time.Now(), SeasonID: season.ID})
	require.NoError(t, err)
	require.Positive(t, p.ID)
}
```
Run (FAIL). In `addProgramme` and `addSeason`, add `.Suffix("RETURNING id")` and replace each `ExecContext`/`RowsAffected` block with, respectively:
```go
	err = s.db.GetContext(ctx, &programmeParam.ID, sql, args...)
	if err != nil {
		span.RecordError(err)
		return Programme{}, fmt.Errorf("failed to add programme: %w", err)
	}
	return programmeParam, nil
```
```go
	err = s.db.GetContext(ctx, &seasonParam.ID, sql, args...)
	if err != nil {
		span.RecordError(err)
		return Season{}, fmt.Errorf("failed to add season: %w", err)
	}
	return seasonParam, nil
```
Re-run: PASS.

- [ ] **Step 2: Fake store and failing service tests**

`server/internal/programme/fake_test.go`:
```go
package programme_test

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/programme"
)

type fakeStore struct {
	mu         sync.Mutex
	programmes map[int]programme.Programme
	seasons    map[int]programme.Season
	nextID     int
}

func newFakeStore() *fakeStore {
	return &fakeStore{programmes: map[int]programme.Programme{}, seasons: map[int]programme.Season{}, nextID: 100}
}

func (f *fakeStore) list(keep func(programme.Programme) bool) []programme.Programme {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []programme.Programme{}
	for _, p := range f.programmes {
		if keep(p) {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DateOfProgramme.After(out[j].DateOfProgramme) })
	return out
}

func (f *fakeStore) GetProgrammes(context.Context) ([]programme.Programme, error) {
	return f.list(func(programme.Programme) bool { return true }), nil
}

func (f *fakeStore) GetProgrammesSeason(_ context.Context, s programme.Season) ([]programme.Programme, error) {
	return f.list(func(p programme.Programme) bool { return p.SeasonID == s.ID }), nil
}

func (f *fakeStore) GetProgramme(_ context.Context, p programme.Programme) (programme.Programme, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.programmes[p.ID]
	if !ok {
		return programme.Programme{}, fmt.Errorf("failed to get programme: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddProgramme(_ context.Context, p programme.Programme) (programme.Programme, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	p.ID = f.nextID
	f.programmes[p.ID] = p
	return p, nil
}

func (f *fakeStore) EditProgramme(_ context.Context, p programme.Programme) (programme.Programme, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.programmes[p.ID] = p
	return p, nil
}

func (f *fakeStore) DeleteProgramme(_ context.Context, p programme.Programme) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.programmes, p.ID)
	return nil
}

func (f *fakeStore) GetSeasons(context.Context) ([]programme.Season, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]programme.Season, 0, len(f.seasons))
	for _, s := range f.seasons {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (f *fakeStore) GetSeason(_ context.Context, s programme.Season) (programme.Season, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.seasons[s.ID]
	if !ok {
		return programme.Season{}, fmt.Errorf("failed to get season: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddSeason(_ context.Context, s programme.Season) (programme.Season, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	s.ID = f.nextID
	f.seasons[s.ID] = s
	return s, nil
}

func (f *fakeStore) EditSeason(_ context.Context, s programme.Season) (programme.Season, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seasons[s.ID] = s
	return s, nil
}

func (f *fakeStore) DeleteSeason(_ context.Context, s programme.Season) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.seasons, s.ID)
	return nil
}
```

`server/internal/programme/service_test.go`:
```go
package programme_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/programme"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

var ctx = context.Background()

func newService() (*programme.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore()
	store.seasons[1] = programme.Season{ID: 1, Season: "2025-26"}
	store.programmes[1] = programme.Programme{ID: 1, Name: "Opening day", FileName: "programme/opening.pdf",
		DateOfProgramme: time.Date(2025, 8, 9, 0, 0, 0, 0, time.UTC), SeasonID: 1}
	store.programmes[2] = programme.Programme{ID: 2, Name: "Friendly", FileName: "programme/friendly.pdf",
		DateOfProgramme: time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)}
	objects := uploadtest.New()
	return programme.NewService(store, upload.New(objects)), store, objects
}

func pdf() *upload.File { return uploadtest.File("p.pdf", "application/pdf", "PDF") }

func kindOf(t *testing.T, err error) svcerr.Kind {
	t.Helper()
	se, ok := svcerr.As(err)
	require.True(t, ok, "want svcerr, got %v", err)
	return se.Kind
}

func TestListAttachesSeasons(t *testing.T) {
	svc, _, _ := newService()
	all, err := svc.List(ctx, 0)
	require.NoError(t, err)
	require.Len(t, all, 2)
	require.NotNil(t, all[0].Season)
	assert.Equal(t, "2025-26", all[0].Season.Name)
	assert.Nil(t, all[1].Season, "programme with no season")

	one, err := svc.List(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, one, 1)

	_, err = svc.List(ctx, 99)
	assert.Equal(t, svcerr.KindNotFound, kindOf(t, err))
}

func TestCreateValidates(t *testing.T) {
	svc, _, _ := newService()
	date := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for name, tc := range map[string]struct {
		in    programme.CreateInput
		file  *upload.File
		field string
	}{
		"name":    {programme.CreateInput{Date: date}, pdf(), "name"},
		"date":    {programme.CreateInput{Name: "x"}, pdf(), "date"},
		"file":    {programme.CreateInput{Name: "x", Date: date}, nil, "file"},
		"season":  {programme.CreateInput{Name: "x", Date: date, SeasonID: 99}, pdf(), "seasonId"},
	} {
		_, err := svc.Create(ctx, tc.in, tc.file)
		se, ok := svcerr.As(err)
		require.True(t, ok, name)
		assert.Contains(t, se.Fields, tc.field, name)
	}
}

func TestCreateAndDelete(t *testing.T) {
	svc, store, objects := newService()
	p, err := svc.Create(ctx, programme.CreateInput{Name: "Cup tie", Date: time.Now(), SeasonID: 1}, pdf())
	require.NoError(t, err)
	assert.Equal(t, 1, store.programmes[p.ID].SeasonID)

	deleted, err := svc.Delete(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "Opening day", deleted.Name)
	assert.Equal(t, []string{"programme/opening.pdf"}, objects.Deleted)
}

func TestSeasonLifecycle(t *testing.T) {
	svc, store, _ := newService()
	_, err := svc.CreateSeason(ctx, "  ")
	assert.Equal(t, svcerr.KindInvalid, kindOf(t, err))

	s, err := svc.CreateSeason(ctx, "2026-27")
	require.NoError(t, err)
	renamed, err := svc.RenameSeason(ctx, s.ID, "2026/27")
	require.NoError(t, err)
	assert.Equal(t, "2026/27", renamed.Name)

	_, err = svc.DeleteSeason(ctx, 1)
	require.NoError(t, err)
	assert.Zero(t, store.programmes[1].SeasonID, "programmes are unlinked, not deleted")
	_, stillThere := store.seasons[1]
	assert.False(t, stillThere)

	_, err = svc.RenameSeason(ctx, 1, "gone")
	assert.Equal(t, svcerr.KindNotFound, kindOf(t, err))
}
```
Run: compile errors expected.

- [ ] **Step 3: Implement `types.go` and `service.go`**

`server/internal/programme/types.go`:
```go
package programme

import "time"

// Public is a match programme as the API returns it.
type Public struct {
	ID      int           `json:"id"`
	Name    string        `json:"name"`
	Date    time.Time     `json:"date"`
	FileURL string        `json:"fileUrl"`
	Season  *PublicSeason `json:"season,omitempty"`
}

// PublicSeason is a programme season.
type PublicSeason struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// CreateInput is a new programme; a file is required. SeasonID 0 means none.
type CreateInput struct {
	Name     string
	Date     time.Time
	SeasonID int
}
```

`server/internal/programme/service.go`:
```go
package programme

import (
	"context"
	"fmt"
	"strings"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

const uploadCategory = "programme"

type store interface {
	GetProgrammes(ctx context.Context) ([]Programme, error)
	GetProgrammesSeason(ctx context.Context, seasonParam Season) ([]Programme, error)
	GetProgramme(ctx context.Context, programmeParam Programme) (Programme, error)
	AddProgramme(ctx context.Context, programmeParam Programme) (Programme, error)
	EditProgramme(ctx context.Context, programmeParam Programme) (Programme, error)
	DeleteProgramme(ctx context.Context, programmeParam Programme) error
	GetSeasons(ctx context.Context) ([]Season, error)
	GetSeason(ctx context.Context, seasonParam Season) (Season, error)
	AddSeason(ctx context.Context, seasonParam Season) (Season, error)
	EditSeason(ctx context.Context, seasonParam Season) (Season, error)
	DeleteSeason(ctx context.Context, seasonParam Season) error
}

// Service is the programme and season logic shared by the API and legacy views.
type Service struct {
	store store
	files *upload.Files
}

func NewService(store store, files *upload.Files) *Service {
	return &Service{store: store, files: files}
}

func (s *Service) public(p Programme, seasons map[int]Season) Public {
	out := Public{ID: p.ID, Name: p.Name, Date: p.DateOfProgramme, FileURL: s.files.URL(p.FileName)}
	if season, ok := seasons[p.SeasonID]; ok {
		out.Season = &PublicSeason{ID: season.ID, Name: season.Season}
	}
	return out
}

func (s *Service) seasonMap(ctx context.Context) (map[int]Season, error) {
	rows, err := s.store.GetSeasons(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list seasons: %w", err)
	}
	m := make(map[int]Season, len(rows))
	for _, r := range rows {
		m[r.ID] = r
	}
	return m, nil
}

// List returns programmes, newest first; seasonID 0 lists every season.
func (s *Service) List(ctx context.Context, seasonID int) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "programme.Service.List")
	defer span.End()

	var (
		rows []Programme
		err  error
	)
	if seasonID == 0 {
		rows, err = s.store.GetProgrammes(ctx)
	} else {
		var season Season
		if season, err = s.store.GetSeason(ctx, Season{ID: seasonID}); err != nil {
			return nil, svcerr.FromStore(err, "season")
		}
		rows, err = s.store.GetProgrammesSeason(ctx, season)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list programmes: %w", err)
	}
	seasons, err := s.seasonMap(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Public, 0, len(rows))
	for _, p := range rows {
		out = append(out, s.public(p, seasons))
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, in CreateInput, file *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "programme.Service.Create")
	defer span.End()

	name := strings.TrimSpace(in.Name)
	fields := svcerr.Fields{}
	if name == "" {
		fields.Add("name", "name is required")
	}
	if in.Date.IsZero() {
		fields.Add("date", "date is required")
	}
	if file == nil {
		fields.Add("file", "file is required")
	}
	if in.SeasonID != 0 {
		if _, err := s.store.GetSeason(ctx, Season{ID: in.SeasonID}); err != nil {
			fields.Add("seasonId", "season does not exist")
		}
	}
	if err := fields.Err(); err != nil {
		return Public{}, err
	}

	key, err := s.files.Save(ctx, file, uploadCategory)
	if err != nil {
		return Public{}, err
	}
	p, err := s.store.AddProgramme(ctx, Programme{Name: name, FileName: key, DateOfProgramme: in.Date, SeasonID: in.SeasonID})
	if err != nil {
		s.files.Remove(ctx, key)
		return Public{}, fmt.Errorf("failed to add programme: %w", err)
	}
	seasons, err := s.seasonMap(ctx)
	if err != nil {
		return Public{}, err
	}
	return s.public(p, seasons), nil
}

// Delete removes the programme and its file, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Public, error) {
	ctx, span := tracer.Start(ctx, "programme.Service.Delete")
	defer span.End()
	p, err := s.store.GetProgramme(ctx, Programme{ID: id})
	if err != nil {
		return Public{}, svcerr.FromStore(err, "programme")
	}
	if err = s.store.DeleteProgramme(ctx, p); err != nil {
		return Public{}, fmt.Errorf("failed to delete programme: %w", err)
	}
	s.files.Remove(ctx, p.FileName)
	return s.public(p, nil), nil
}

func (s *Service) Seasons(ctx context.Context) ([]PublicSeason, error) {
	ctx, span := tracer.Start(ctx, "programme.Service.Seasons")
	defer span.End()
	rows, err := s.store.GetSeasons(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list seasons: %w", err)
	}
	out := make([]PublicSeason, 0, len(rows))
	for _, r := range rows {
		out = append(out, PublicSeason{ID: r.ID, Name: r.Season})
	}
	return out, nil
}

func (s *Service) CreateSeason(ctx context.Context, name string) (PublicSeason, error) {
	ctx, span := tracer.Start(ctx, "programme.Service.CreateSeason")
	defer span.End()
	name = strings.TrimSpace(name)
	if name == "" {
		return PublicSeason{}, svcerr.InvalidField("season", "season is required")
	}
	r, err := s.store.AddSeason(ctx, Season{Season: name})
	if err != nil {
		return PublicSeason{}, fmt.Errorf("failed to add season: %w", err)
	}
	return PublicSeason{ID: r.ID, Name: r.Season}, nil
}

func (s *Service) RenameSeason(ctx context.Context, id int, name string) (PublicSeason, error) {
	ctx, span := tracer.Start(ctx, "programme.Service.RenameSeason")
	defer span.End()
	name = strings.TrimSpace(name)
	if name == "" {
		return PublicSeason{}, svcerr.InvalidField("season", "season is required")
	}
	r, err := s.store.GetSeason(ctx, Season{ID: id})
	if err != nil {
		return PublicSeason{}, svcerr.FromStore(err, "season")
	}
	r.Season = name
	if _, err = s.store.EditSeason(ctx, r); err != nil {
		return PublicSeason{}, fmt.Errorf("failed to edit season: %w", err)
	}
	return PublicSeason{ID: r.ID, Name: r.Season}, nil
}

// DeleteSeason unlinks the season's programmes (they are kept) and deletes it.
func (s *Service) DeleteSeason(ctx context.Context, id int) (PublicSeason, error) {
	ctx, span := tracer.Start(ctx, "programme.Service.DeleteSeason")
	defer span.End()
	r, err := s.store.GetSeason(ctx, Season{ID: id})
	if err != nil {
		return PublicSeason{}, svcerr.FromStore(err, "season")
	}
	linked, err := s.store.GetProgrammesSeason(ctx, r)
	if err != nil {
		return PublicSeason{}, fmt.Errorf("failed to list season programmes: %w", err)
	}
	for _, p := range linked {
		p.SeasonID = 0
		if _, err = s.store.EditProgramme(ctx, p); err != nil {
			return PublicSeason{}, fmt.Errorf("failed to unlink programme %d: %w", p.ID, err)
		}
	}
	if err = s.store.DeleteSeason(ctx, r); err != nil {
		return PublicSeason{}, fmt.Errorf("failed to delete season: %w", err)
	}
	return PublicSeason{ID: r.ID, Name: r.Season}, nil
}
```
Run: `go test ./server/internal/programme/`. Expected: PASS.

- [ ] **Step 4: Failing handler tests, then `handlers.go`**

`server/internal/programme/handlers_test.go`:
```go
package programme_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/programme"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestProgrammeRoutes(t *testing.T) {
	svc, _, _ := newService()
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	programme.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	editor := anon.As(authtest.Cookie(t, sessions, authtest.Webmaster))
	manager := anon.As(authtest.Cookie(t, sessions, authtest.Manager))

	rec := anon.Get(t, "/api/v1/programmes?season=1")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, apitest.Decode[[]programme.Public](t, rec), 1)
	assert.Equal(t, http.StatusBadRequest, anon.Get(t, "/api/v1/programmes?season=abc").Code)
	assert.Equal(t, http.StatusNotFound, anon.Get(t, "/api/v1/programmes?season=99").Code)

	pdfPart := apitest.FilePart{Field: "file", Name: "p.pdf", ContentType: "application/pdf", Body: "PDF"}
	fields := map[string]string{"name": "Cup", "date": "2026-01-10", "seasonId": "1"}
	assert.Equal(t, http.StatusForbidden, manager.Multipart(t, http.MethodPost, "/api/v1/programmes", fields, pdfPart).Code)
	assert.Equal(t, http.StatusCreated, editor.Multipart(t, http.MethodPost, "/api/v1/programmes", fields, pdfPart).Code)
	assert.Equal(t, http.StatusNoContent, editor.JSON(t, http.MethodDelete, "/api/v1/programmes/2", nil).Code)

	rec = anon.Get(t, "/api/v1/seasons")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, apitest.Decode[[]programme.PublicSeason](t, rec), 1)

	rec = editor.JSON(t, http.MethodPost, "/api/v1/seasons", map[string]string{"season": "2026-27"})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	created := apitest.Decode[programme.PublicSeason](t, rec)
	assert.Equal(t, http.StatusOK, editor.JSON(t, http.MethodPatch, "/api/v1/seasons/1", map[string]string{"season": "25/26"}).Code)
	assert.Equal(t, http.StatusUnprocessableEntity, editor.JSON(t, http.MethodPatch, "/api/v1/seasons/1", map[string]string{"season": ""}).Code)
	assert.Equal(t, http.StatusForbidden, manager.JSON(t, http.MethodDelete, "/api/v1/seasons/1", nil).Code)
	assert.Equal(t, http.StatusNoContent, editor.JSON(t, http.MethodDelete, "/api/v1/seasons/1", nil).Code)
	assert.Positive(t, created.ID)
}
```

`server/internal/programme/handlers.go`:
```go
package programme

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /programmes and /seasons.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the programme and season routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/programmes", h.list)
	g.POST("/programmes", h.create, guards.Editor)
	g.DELETE("/programmes/:id", h.remove, guards.Editor)
	g.GET("/seasons", h.seasons)
	g.POST("/seasons", h.createSeason, guards.Editor)
	g.PATCH("/seasons/:id", h.renameSeason, guards.Editor)
	g.DELETE("/seasons/:id", h.deleteSeason, guards.Editor)
}

// SeasonInput is the JSON body for creating or renaming a season.
type SeasonInput struct {
	Season string `json:"season"`
}

// list returns programmes, optionally for one season.
//
//	@Summary	List programmes
//	@Tags		programmes
//	@Produce	json
//	@Param		season	query		int	false	"Season ID; omit for all"
//	@Success	200		{array}		programme.Public
//	@Failure	404		{object}	web.ErrorResponse
//	@Router		/programmes [get]
func (h *Handlers) list(c echo.Context) error {
	seasonID := 0
	if raw := c.QueryParam("season"); raw != "" {
		id, err := strconv.Atoi(raw)
		if err != nil || id < 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid season")
		}
		seasonID = id
	}
	out, err := h.svc.List(c.Request().Context(), seasonID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// create uploads a programme.
//
//	@Summary	Upload a programme
//	@Tags		programmes
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		name		formData	string	true	"Name"
//	@Param		date		formData	string	true	"YYYY-MM-DD"
//	@Param		seasonId	formData	int		false	"Season ID (0 or omitted for none)"
//	@Param		file		formData	file	true	"Programme PDF"
//	@Success	201			{object}	programme.Public
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/programmes [post]
func (h *Handlers) create(c echo.Context) error {
	date, err := web.FormDate(c, "date")
	if err != nil {
		return err
	}
	file, err := web.FormFile(c, "file")
	if err != nil {
		return err
	}
	in := CreateInput{Name: c.FormValue("name")}
	if date != nil {
		in.Date = *date
	}
	if raw := c.FormValue("seasonId"); raw != "" {
		if in.SeasonID, err = strconv.Atoi(raw); err != nil {
			return svcerr.InvalidField("seasonId", "season must be a number")
		}
	}
	out, err := h.svc.Create(c.Request().Context(), in, file)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// remove deletes a programme.
//
//	@Summary	Delete a programme
//	@Tags		programmes
//	@Param		id	path	int	true	"Programme ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/programmes/{id} [delete]
func (h *Handlers) remove(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if _, err = h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// seasons lists programme seasons.
//
//	@Summary	List seasons
//	@Tags		programmes
//	@Produce	json
//	@Success	200	{array}	programme.PublicSeason
//	@Router		/seasons [get]
func (h *Handlers) seasons(c echo.Context) error {
	out, err := h.svc.Seasons(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// createSeason adds a season.
//
//	@Summary	Create a season
//	@Tags		programmes
//	@Accept		json
//	@Produce	json
//	@Param		body	body		programme.SeasonInput	true	"Season name"
//	@Success	201		{object}	programme.PublicSeason
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/seasons [post]
func (h *Handlers) createSeason(c echo.Context) error {
	var in SeasonInput
	if err := web.BindJSON(c, &in); err != nil {
		return err
	}
	out, err := h.svc.CreateSeason(c.Request().Context(), in.Season)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// renameSeason renames a season.
//
//	@Summary	Rename a season
//	@Tags		programmes
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int						true	"Season ID"
//	@Param		body	body		programme.SeasonInput	true	"New name"
//	@Success	200		{object}	programme.PublicSeason
//	@Failure	404		{object}	web.ErrorResponse
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/seasons/{id} [patch]
func (h *Handlers) renameSeason(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	var in SeasonInput
	if err = web.BindJSON(c, &in); err != nil {
		return err
	}
	out, err := h.svc.RenameSeason(c.Request().Context(), id, in.Season)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// deleteSeason deletes a season; its programmes are kept but unlinked.
//
//	@Summary	Delete a season
//	@Tags		programmes
//	@Param		id	path	int	true	"Season ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/seasons/{id} [delete]
func (h *Handlers) deleteSeason(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if _, err = h.svc.DeleteSeason(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
```
Run: `go test ./server/internal/programme/`. Expected: PASS.

- [ ] **Step 5: Legacy writes → service**

Add `programmeSvc` / `ProgrammeService` to `Views`/`Deps`/`New`. In `server/internal/legacy/views/programme.go` keep `ProgrammesFunc`, `ProgrammesSeasonsFunc` and `ProgrammeSeasonSelectFunc`; replace the five write handlers:
```go
func (v *Views) ProgrammeAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ProgrammeAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}
	seasonID, err := strconv.Atoi(c.FormValue("programmeSeason"))
	if err != nil {
		data.Error = fmt.Sprintf("failed to parse programmeSeason for programme add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	date, err := time.Parse("02/01/2006", c.FormValue("dateOfProgramme"))
	if err != nil {
		data.Error = fmt.Sprintf("failed to parse dateOfProgramme for programme add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	file, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for programme add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	created, err := v.programmeSvc.Create(c.Request().Context(), programme.CreateInput{
		Name: c.FormValue("name"), Date: date, SeasonID: seasonID,
	}, file)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add programme for programme add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add programme for programme add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", created.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) ProgrammeDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ProgrammeDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for programme delete, error: %w", err)
	}
	deleted, err := v.programmeSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete programme for programme delete, programme id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", deleted.Name))
	return c.Redirect(http.StatusFound, "/programmes")
}

func (v *Views) ProgrammeSeasonAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ProgrammeSeasonAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}
	created, err := v.programmeSvc.CreateSeason(c.Request().Context(), c.FormValue("season"))
	if err != nil {
		data.Error = fmt.Sprintf("failed to add season for season add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", created.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) ProgrammeSeasonEditFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ProgrammeSeasonEditFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for programme season edit, error: %w", err)
	}
	data := struct {
		Error string `json:"error"`
	}{}
	renamed, err := v.programmeSvc.RenameSeason(c.Request().Context(), id, c.FormValue("season"))
	if err != nil {
		data.Error = fmt.Sprintf("failed to edit season for season edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully edited \"%s\"", renamed.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) ProgrammeSeasonDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ProgrammeSeasonDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for programme season delete, error: %w", err)
	}
	deleted, err := v.programmeSvc.DeleteSeason(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete season for programme season delete, season id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", deleted.Name))
	return c.Redirect(http.StatusFound, "/programmes")
}
```
Behaviour note: legacy accepted an empty season name; the service now rejects it (Plan Decision 3).

- [ ] **Step 6: Wire, regenerate, run, commit**

In `app.Build`:
- Add `programmeSvc := programme.NewService(s.Programme, uploads)`.
- Add `ProgrammeService: programmeSvc,` to `views.Deps`.
- Add `programme.NewHandlers(programmeSvc).Register(api, guards)`.
```bash
go generate ./server/internal/docs
goimports -w server/internal/legacy/views
gofmt -l server; go build ./... && go vet ./... && go test ./... && go tool golangci-lint run ./...
git add -A server
git commit -m "Add programme/season service and API; legacy writes use the service

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---
### Task 15: Teams (and the `Identify` guard)

**Import-cycle note:** `player`, `sponsor` and `user` all import `team`, so `team` cannot import them. Two consequences follow:
- **Deleting a team** unlinks its players, sponsors and managers through a `team.Detacher` interface. The three stores satisfy it with one `UPDATE` each.
- **The team detail page** (team, managers, sponsors and squad) is an aggregate, served by the `site` package in Task 18 as `GET /teams/{id}`. This task serves `GET /teams`, `POST /teams`, `PATCH /teams/{id}` and `DELETE /teams/{id}`.

`GET /teams` shows inactive teams only to logged-in users, as legacy `TeamsFunc` does. That needs a guard that identifies the user without rejecting anonymous requests, so this task adds `Guards.Identify` and `web.LoggedIn`.

**Files:**
- Modify: `server/internal/web/guards.go`, `server/internal/web/json.go`, `server/internal/auth/guards.go`, `server/internal/auth/guards_test.go`
- Modify: `server/internal/team/store.go` (`addTeam` returns ID)
- Modify: `server/internal/player/store.go`, `server/internal/sponsor/store.go`, `server/internal/user/store.go` (add `DetachTeam`)
- Create: `server/internal/team/{store_test,types,service,handlers,fake_test,service_test,handlers_test}.go`
- Create: `server/internal/player/detach_test.go`, `server/internal/sponsor/detach_test.go`, `server/internal/user/detach_test.go`
- Modify: `server/internal/legacy/views/team.go`, `views.go`, `server/internal/app/app.go`

**Interfaces:**
- Produces (`web`): `Guards.Identify` (loads the user when logged in, never rejects); `const LoggedInKey`; `LoggedIn(c echo.Context) bool`; `FormInt(c, field) (*int, error)` (used by teams and players).
- Produces (store methods): `(*player.Store).DetachTeam(ctx, teamID int) error` (players → team 0), `(*sponsor.Store).DetachTeam` (sponsors → `"A"`), `(*user.Store).DetachTeam` (users → team 0).
- Produces (`team`):
  - `type Detacher interface{DetachTeam(ctx, teamID int) error}`
  - Types:
    - `Public{ID int; Name, Description, League, Division, LeagueTableURL, FixturesURL, Coach, Physio, ImageURL string; IsActive, IsYouth bool; Ages int}`
    - `CreateInput{Name, Description, League, Division, LeagueTable, Fixtures, Coach, Physio string; IsActive, IsYouth bool; Ages int}`
    - `UpdateInput{Name, Description, League, Division, LeagueTable, Fixtures, Coach, Physio *string; IsActive, IsYouth *bool; Ages *int; RemoveImage bool}`
  - `NewService(store, *upload.Files, ...Detacher) *Service`, with methods:
    - `List(ctx, includeInactive bool) ([]Public, error)`
    - `Get(ctx, id) (Public, error)`
    - `Create`, `Update`, `Delete(ctx, id) (Public, error)`
  - `NewHandlers`, `Register`.
- Produces (legacy): `Deps.TeamService`.

- [ ] **Step 1: Add `Identify`: failing test first**

Append to `server/internal/auth/guards_test.go`:
```go
func TestIdentifyNeverRejects(t *testing.T) {
	s := authtest.Everyone()
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	api.GET("/who", func(c echo.Context) error {
		if web.LoggedIn(c) {
			return c.String(http.StatusOK, "member")
		}
		return c.String(http.StatusOK, "guest")
	}, s.Guards().Identify)

	assert.Equal(t, "guest", apitest.New(e).Get(t, "/api/v1/who").Body.String())
	ghost := user.User{ID: 99, Email: "gone@example.test"}
	assert.Equal(t, "guest", apitest.New(e).As(authtest.Cookie(t, s, ghost)).Get(t, "/api/v1/who").Body.String())
	assert.Equal(t, "member", apitest.New(e).As(authtest.Cookie(t, s, authtest.Manager)).Get(t, "/api/v1/who").Body.String())
}
```
Run: `go test ./server/internal/auth/`. Expected: compile error (`Identify` undefined).

In `server/internal/web/guards.go`, add the field and the helper:
```go
	Identify            echo.MiddlewareFunc // loads the user if logged in; never rejects
```
```go
// LoggedInKey is set on the request context once a guard has loaded a
// logged-in user.
const LoggedInKey = "web.loggedIn"

// LoggedIn reports whether a guard found a logged-in user for this request.
func LoggedIn(c echo.Context) bool {
	ok, _ := c.Get(LoggedInKey).(bool)
	return ok
}
```
Append to `server/internal/web/json.go` (add `strconv` if it is not already imported):
```go
// FormInt parses an optional integer form field.
func FormInt(c echo.Context, field string) (*int, error) {
	raw := FormString(c, field)
	if raw == nil || *raw == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(*raw)
	if err != nil {
		return nil, svcerr.InvalidField(field, field+" must be a whole number")
	}
	return &n, nil
}
```
In `server/internal/auth/guards.go`:
- in `load`, just before `return fresh, nil`, add `c.Set(web.LoggedInKey, true)`
- in `Guards()` add `Identify: s.Identify,`
- add:
```go
// Identify loads the user when a valid session is present, but lets
// anonymous (or stale-session) requests through.
func (s *Sessions) Identify(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if _, ok := s.User(c.Request()); ok {
			_, _ = s.load(c) //nolint:errcheck // a stale session is simply anonymous here
		}
		return next(c)
	}
}
```
Run: PASS.

- [ ] **Step 2: Store tests for `addTeam` and the three `DetachTeam`s**

`server/internal/team/store_test.go`:
```go
package team_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestAddTeamReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	added, err := team.NewTeamRepo(db).AddTeam(context.Background(), team.Team{Name: "Reserves", IsActive: true, Ages: 99})
	require.NoError(t, err)
	require.Positive(t, added.ID)
}
```

`server/internal/player/detach_test.go`:
```go
package player_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestDetachTeam(t *testing.T) {
	db, _ := testdb.Open(t)
	store := player.NewPlayerRepo(db)
	ctx := context.Background()
	require.NoError(t, store.DetachTeam(ctx, 1))
	left, err := store.GetPlayersTeam(ctx, team.Team{ID: 1})
	require.NoError(t, err)
	assert.Empty(t, left)
	youth, err := store.GetPlayersTeam(ctx, team.Team{ID: 2})
	require.NoError(t, err)
	assert.Len(t, youth, 1, "other teams untouched")
}
```

`server/internal/sponsor/detach_test.go`:
```go
package sponsor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestDetachTeam(t *testing.T) {
	db, _ := testdb.Open(t)
	store := sponsor.NewSponsorRepo(db)
	ctx := context.Background()
	require.NoError(t, store.DetachTeam(ctx, 1))
	left, err := store.GetSponsorsTeam(ctx, team.Team{ID: 1})
	require.NoError(t, err)
	assert.Empty(t, left)
	s, err := store.GetSponsor(ctx, sponsor.Sponsor{ID: 2})
	require.NoError(t, err)
	assert.Equal(t, "A", s.TeamID, "team sponsors become club-wide, as legacy TeamDeleteFunc did")
}
```

`server/internal/user/detach_test.go`:
```go
package user_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

func TestDetachTeam(t *testing.T) {
	db, _ := testdb.Open(t)
	store := user.NewUserRepo(db)
	ctx := context.Background()
	require.NoError(t, store.DetachTeam(ctx, 1))
	managers, err := store.GetUsersManagersTeam(ctx, team.Team{ID: 1})
	require.NoError(t, err)
	assert.Empty(t, managers)
}
```
Run with `AFC_TEST_DB`: `go test -run 'TestAddTeamReturnsID|TestDetachTeam' ./server/internal/team/ ./server/internal/player/ ./server/internal/sponsor/ ./server/internal/user/`. Expected: FAIL (`DetachTeam` undefined, ID 0).

- [ ] **Step 3: Implement the store changes**

`team/store.go` `addTeam`: add `.Suffix("RETURNING id")` and replace the `ExecContext`/`RowsAffected` block with:
```go
	err = s.db.GetContext(ctx, &teamParam.ID, sql, args...)
	if err != nil {
		span.RecordError(err)
		return Team{}, fmt.Errorf("failed to add team: %w", err)
	}
	return teamParam, nil
```
Append to `player/store.go`:
```go
// DetachTeam moves every player on teamID to no team (team_id 0).
func (s *Store) DetachTeam(ctx context.Context, teamID int) error {
	ctx, span := tracer.Start(ctx, "player.DetachTeam")
	defer span.End()
	sql, args, err := utils.PSQL().Update("players").Set("team_id", 0).Where(sq.Eq{"team_id": teamID}).ToSql()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to build sql for detach team: %w", err)
	}
	if _, err = s.db.ExecContext(ctx, sql, args...); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to detach players from team: %w", err)
	}
	return nil
}
```
Append to `sponsor/store.go`:
```go
// DetachTeam makes every sponsor of teamID club-wide ("A").
func (s *Store) DetachTeam(ctx context.Context, teamID int) error {
	ctx, span := tracer.Start(ctx, "sponsor.DetachTeam")
	defer span.End()
	sql, args, err := utils.PSQL().Update("sponsors").Set("team_id", "A").
		Where(sq.Eq{"team_id": strconv.Itoa(teamID)}).ToSql()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to build sql for detach team: %w", err)
	}
	if _, err = s.db.ExecContext(ctx, sql, args...); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to detach sponsors from team: %w", err)
	}
	return nil
}
```
Append to `user/store.go`:
```go
// DetachTeam removes teamID from every user (managers of that team).
func (s *Store) DetachTeam(ctx context.Context, teamID int) error {
	ctx, span := tracer.Start(ctx, "user.DetachTeam")
	defer span.End()
	sql, args, err := utils.PSQL().Update("users").Set("team_id", 0).Where(sq.Eq{"team_id": teamID}).ToSql()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to build sql for detach team: %w", err)
	}
	if _, err = s.db.ExecContext(ctx, sql, args...); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to detach users from team: %w", err)
	}
	return nil
}
```
Each file's existing imports provide `sq` and `utils`. Add `strconv` in `sponsor/store.go`. Re-run the store tests: PASS.

- [ ] **Step 4: Fake store and failing service tests**

`server/internal/team/fake_test.go`:
```go
package team_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/team"
)

type fakeStore struct {
	mu     sync.Mutex
	rows   map[int]team.Team
	nextID int
}

func newFakeStore(rows ...team.Team) *fakeStore {
	f := &fakeStore{rows: map[int]team.Team{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) list(keep func(team.Team) bool) []team.Team {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []team.Team{}
	for _, r := range f.rows {
		if keep(r) {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (f *fakeStore) GetTeams(context.Context) ([]team.Team, error) {
	return f.list(func(team.Team) bool { return true }), nil
}

func (f *fakeStore) GetTeamsActive(context.Context) ([]team.Team, error) {
	return f.list(func(t team.Team) bool { return t.IsActive }), nil
}

func (f *fakeStore) GetTeam(_ context.Context, t team.Team) (team.Team, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[t.ID]
	if !ok {
		return team.Team{}, fmt.Errorf("failed to get team: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddTeam(_ context.Context, t team.Team) (team.Team, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	t.ID = f.nextID
	f.rows[t.ID] = t
	return t, nil
}

func (f *fakeStore) EditTeam(_ context.Context, t team.Team) (team.Team, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.rows[t.ID]; !ok {
		return team.Team{}, errors.New("no such row")
	}
	f.rows[t.ID] = t
	return t, nil
}

func (f *fakeStore) DeleteTeam(_ context.Context, t team.Team) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, t.ID)
	return nil
}

func (f *fakeStore) row(id int) team.Team {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.rows[id]
}

type recordingDetacher struct {
	name   string
	calls  *[]string
	failOn int
}

func (d recordingDetacher) DetachTeam(_ context.Context, teamID int) error {
	*d.calls = append(*d.calls, fmt.Sprintf("%s:%d", d.name, teamID))
	if teamID == d.failOn {
		return errors.New(d.name + " failed")
	}
	return nil
}
```

`server/internal/team/service_test.go`:
```go
package team_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

var ctx = context.Background()

func firstTeam() team.Team {
	return team.Team{ID: 1, Name: "First Team", League: null.StringFrom("Hellenic"),
		FileName: null.StringFrom("team/first.jpg"), IsActive: true, Ages: 99}
}

func oldTeam() team.Team {
	return team.Team{ID: 3, Name: "Old Team", IsActive: false, Ages: 99}
}

func newService(detachers ...team.Detacher) (*team.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore(firstTeam(), oldTeam())
	objects := uploadtest.New()
	objects.Objects["team/first.jpg"] = "IMG"
	return team.NewService(store, upload.New(objects), detachers...), store, objects
}

func fields(t *testing.T, err error) map[string]string {
	t.Helper()
	se, ok := svcerr.As(err)
	require.True(t, ok, "want svcerr, got %v", err)
	return se.Fields
}

func TestListHidesInactiveFromPublic(t *testing.T) {
	svc, _, _ := newService()
	public, err := svc.List(ctx, false)
	require.NoError(t, err)
	assert.Len(t, public, 1)
	all, err := svc.List(ctx, true)
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestCreateValidates(t *testing.T) {
	svc, _, _ := newService()
	assert.Contains(t, fields(t, mustErr(svc.Create(ctx, team.CreateInput{Ages: 99}, nil))), "name")
	assert.Contains(t, fields(t, mustErr(svc.Create(ctx, team.CreateInput{Name: "x", LeagueTable: "table"}, nil))), "leagueTable")
	assert.Contains(t, fields(t, mustErr(svc.Create(ctx, team.CreateInput{Name: "x", Fixtures: "fixtures"}, nil))), "fixtures")
	assert.Contains(t, fields(t, mustErr(svc.Create(ctx, team.CreateInput{Name: "x", Ages: -1}, nil))), "ages")
}

func mustErr(_ team.Public, err error) error { return err }

func TestCreateYouthRuleAndFlags(t *testing.T) {
	svc, store, _ := newService()
	u12, err := svc.Create(ctx, team.CreateInput{Name: "U12", Ages: 12}, nil)
	require.NoError(t, err)
	assert.True(t, u12.IsYouth, "ages < 19 is always youth")
	assert.False(t, u12.IsActive)

	// Legacy quirk fixed: ticking youth used to set isActive instead.
	senior, err := svc.Create(ctx, team.CreateInput{Name: "Vets", Ages: 99, IsYouth: true}, nil)
	require.NoError(t, err)
	assert.True(t, store.row(senior.ID).IsYouth)
	assert.False(t, store.row(senior.ID).IsActive)
}

func TestUpdateCanClearFlags(t *testing.T) {
	svc, store, _ := newService()
	no := false
	_, err := svc.Update(ctx, 1, team.UpdateInput{IsActive: &no}, nil)
	require.NoError(t, err)
	assert.False(t, store.row(1).IsActive, "legacy edit could never untick isActive")
	assert.Equal(t, "Hellenic", store.row(1).League.String, "omitted fields unchanged")
}

func TestUpdateAgesForcesYouth(t *testing.T) {
	svc, store, _ := newService()
	ages := 16
	_, err := svc.Update(ctx, 1, team.UpdateInput{Ages: &ages}, nil)
	require.NoError(t, err)
	assert.True(t, store.row(1).IsYouth)
}

func TestUpdateRemoveImage(t *testing.T) {
	svc, store, objects := newService()
	_, err := svc.Update(ctx, 1, team.UpdateInput{RemoveImage: true}, nil)
	require.NoError(t, err)
	assert.False(t, store.row(1).FileName.Valid)
	assert.Equal(t, []string{"team/first.jpg"}, objects.Deleted)
}

func TestDeleteDetachesThenDeletes(t *testing.T) {
	var calls []string
	svc, store, objects := newService(
		recordingDetacher{name: "players", calls: &calls},
		recordingDetacher{name: "sponsors", calls: &calls},
		recordingDetacher{name: "users", calls: &calls},
	)
	deleted, err := svc.Delete(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "First Team", deleted.Name)
	assert.Equal(t, []string{"players:1", "sponsors:1", "users:1"}, calls)
	_, still := store.rows[1]
	assert.False(t, still)
	assert.Equal(t, []string{"team/first.jpg"}, objects.Deleted)
}

func TestDeleteStopsWhenDetachFails(t *testing.T) {
	var calls []string
	svc, store, _ := newService(recordingDetacher{name: "players", calls: &calls, failOn: 1})
	_, err := svc.Delete(ctx, 1)
	require.Error(t, err)
	_, still := store.rows[1]
	assert.True(t, still, "team must not be deleted if its players could not be detached")
}
```
Run: compile errors expected.

- [ ] **Step 5: Implement `types.go` and `service.go`**

`server/internal/team/types.go`:
```go
package team

// Public is a team as the API returns it.
type Public struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	League         string `json:"league,omitempty"`
	Division       string `json:"division,omitempty"`
	LeagueTableURL string `json:"leagueTableUrl,omitempty"`
	FixturesURL    string `json:"fixturesUrl,omitempty"`
	Coach          string `json:"coach,omitempty"`
	Physio         string `json:"physio,omitempty"`
	ImageURL       string `json:"imageUrl,omitempty"`
	IsActive       bool   `json:"isActive"`
	IsYouth        bool   `json:"isYouth"`
	Ages           int    `json:"ages"`
}

// CreateInput is a new team. Ages below 19 always make a youth team.
type CreateInput struct {
	Name, Description, League, Division, LeagueTable, Fixtures, Coach, Physio string

	IsActive bool
	IsYouth  bool
	Ages     int
}

// UpdateInput changes a team; nil fields are left unchanged and empty
// strings clear optional fields.
type UpdateInput struct {
	Name, Description, League, Division, LeagueTable, Fixtures, Coach, Physio *string

	IsActive    *bool
	IsYouth     *bool
	Ages        *int
	RemoveImage bool
}
```

`server/internal/team/service.go`:
```go
package team

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

const (
	uploadCategory = "team"
	// youthAgeLimit: teams for players under this age are youth teams.
	youthAgeLimit = 19
)

type store interface {
	GetTeams(ctx context.Context) ([]Team, error)
	GetTeamsActive(ctx context.Context) ([]Team, error)
	GetTeam(ctx context.Context, teamParam Team) (Team, error)
	AddTeam(ctx context.Context, teamParam Team) (Team, error)
	EditTeam(ctx context.Context, teamParam Team) (Team, error)
	DeleteTeam(ctx context.Context, teamParam Team) error
}

// Detacher unlinks records that reference a team before it is deleted
// (satisfied by the player, sponsor and user stores).
type Detacher interface {
	DetachTeam(ctx context.Context, teamID int) error
}

// Service is the team business logic shared by the API and legacy views.
type Service struct {
	store     store
	files     *upload.Files
	detachers []Detacher
}

func NewService(store store, files *upload.Files, detachers ...Detacher) *Service {
	return &Service{store: store, files: files, detachers: detachers}
}

func optional(s string) null.String {
	s = strings.TrimSpace(s)
	return null.NewString(s, s != "")
}

// setOptional applies an optional text update: nil leaves dst alone, empty clears it.
func setOptional(dst *null.String, value *string) {
	if value != nil {
		*dst = optional(*value)
	}
}

func (s *Service) public(t Team) Public {
	return Public{
		ID:             t.ID,
		Name:           t.Name,
		Description:    t.Description.String,
		League:         t.League.String,
		Division:       t.Division.String,
		LeagueTableURL: t.LeagueTable.String,
		FixturesURL:    t.Fixtures.String,
		Coach:          t.Coach.String,
		Physio:         t.Physio.String,
		ImageURL:       s.files.URL(t.FileName.String),
		IsActive:       t.IsActive,
		IsYouth:        t.IsYouth,
		Ages:           t.Ages,
	}
}

func validate(t Team) error {
	f := svcerr.Fields{}
	if strings.TrimSpace(t.Name) == "" {
		f.Add("name", "name is required")
	}
	if t.LeagueTable.Valid {
		if _, err := url.ParseRequestURI(t.LeagueTable.String); err != nil {
			f.Add("leagueTable", "league table must be a full URL")
		}
	}
	if t.Fixtures.Valid {
		if _, err := url.ParseRequestURI(t.Fixtures.String); err != nil {
			f.Add("fixtures", "fixtures must be a full URL")
		}
	}
	if t.Ages < 0 {
		f.Add("ages", "ages must be zero or more")
	}
	return f.Err()
}

// List returns active teams, or every team when includeInactive is set
// (logged-in users see inactive teams too).
func (s *Service) List(ctx context.Context, includeInactive bool) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "team.Service.List")
	defer span.End()
	var (
		rows []Team
		err  error
	)
	if includeInactive {
		rows, err = s.store.GetTeams(ctx)
	} else {
		rows, err = s.store.GetTeamsActive(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	out := make([]Public, 0, len(rows))
	for _, t := range rows {
		out = append(out, s.public(t))
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, id int) (Public, error) {
	ctx, span := tracer.Start(ctx, "team.Service.Get")
	defer span.End()
	t, err := s.get(ctx, id)
	if err != nil {
		return Public{}, err
	}
	return s.public(t), nil
}

func (s *Service) get(ctx context.Context, id int) (Team, error) {
	t, err := s.store.GetTeam(ctx, Team{ID: id})
	if err != nil {
		return Team{}, svcerr.FromStore(err, "team")
	}
	return t, nil
}

func (s *Service) Create(ctx context.Context, in CreateInput, image *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "team.Service.Create")
	defer span.End()

	t := Team{
		Name:        strings.TrimSpace(in.Name),
		Description: optional(in.Description),
		League:      optional(in.League),
		Division:    optional(in.Division),
		LeagueTable: optional(in.LeagueTable),
		Fixtures:    optional(in.Fixtures),
		Coach:       optional(in.Coach),
		Physio:      optional(in.Physio),
		IsActive:    in.IsActive,
		IsYouth:     in.IsYouth || in.Ages < youthAgeLimit,
		Ages:        in.Ages,
	}
	if err := validate(t); err != nil {
		return Public{}, err
	}
	if image != nil {
		key, err := s.files.Save(ctx, image, uploadCategory)
		if err != nil {
			return Public{}, err
		}
		t.FileName = null.StringFrom(key)
	}
	added, err := s.store.AddTeam(ctx, t)
	if err != nil {
		s.files.Remove(ctx, t.FileName.String)
		return Public{}, fmt.Errorf("failed to add team: %w", err)
	}
	return s.public(added), nil
}

func (s *Service) Update(ctx context.Context, id int, in UpdateInput, image *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "team.Service.Update")
	defer span.End()

	t, err := s.get(ctx, id)
	if err != nil {
		return Public{}, err
	}
	if in.Name != nil {
		t.Name = strings.TrimSpace(*in.Name)
	}
	setOptional(&t.Description, in.Description)
	setOptional(&t.League, in.League)
	setOptional(&t.Division, in.Division)
	setOptional(&t.LeagueTable, in.LeagueTable)
	setOptional(&t.Fixtures, in.Fixtures)
	setOptional(&t.Coach, in.Coach)
	setOptional(&t.Physio, in.Physio)
	if in.IsActive != nil {
		t.IsActive = *in.IsActive
	}
	if in.IsYouth != nil {
		t.IsYouth = *in.IsYouth
	}
	if in.Ages != nil {
		t.Ages = *in.Ages
	}
	if t.Ages < youthAgeLimit {
		t.IsYouth = true
	}
	if err = validate(t); err != nil {
		return Public{}, err
	}

	oldKey, newKey := t.FileName.String, ""
	switch {
	case image != nil:
		if newKey, err = s.files.Save(ctx, image, uploadCategory); err != nil {
			return Public{}, err
		}
		t.FileName = null.StringFrom(newKey)
	case in.RemoveImage:
		t.FileName = null.String{}
	}
	if _, err = s.store.EditTeam(ctx, t); err != nil {
		s.files.Remove(ctx, newKey)
		return Public{}, fmt.Errorf("failed to edit team: %w", err)
	}
	if oldKey != "" && oldKey != t.FileName.String {
		s.files.Remove(ctx, oldKey)
	}
	return s.public(t), nil
}

// Delete unlinks the team's players, sponsors and managers, then deletes the
// team and its image. If any unlink fails, the team is kept.
func (s *Service) Delete(ctx context.Context, id int) (Public, error) {
	ctx, span := tracer.Start(ctx, "team.Service.Delete")
	defer span.End()

	t, err := s.get(ctx, id)
	if err != nil {
		return Public{}, err
	}
	for _, d := range s.detachers {
		if err = d.DetachTeam(ctx, t.ID); err != nil {
			return Public{}, fmt.Errorf("failed to detach team %d: %w", t.ID, err)
		}
	}
	if err = s.store.DeleteTeam(ctx, t); err != nil {
		return Public{}, fmt.Errorf("failed to delete team: %w", err)
	}
	s.files.Remove(ctx, t.FileName.String)
	return s.public(t), nil
}
```
Run: `go test ./server/internal/team/`. Expected: PASS.

- [ ] **Step 6: Failing handler tests, then `handlers.go`**

`server/internal/team/handlers_test.go`:
```go
package team_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestTeamRoutes(t *testing.T) {
	svc, store, _ := newService()
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	team.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	editor := anon.As(authtest.Cookie(t, sessions, authtest.Webmaster))
	manager := anon.As(authtest.Cookie(t, sessions, authtest.Manager))

	assert.Len(t, apitest.Decode[[]team.Public](t, anon.Get(t, "/api/v1/teams")), 1, "anonymous: active only")
	assert.Len(t, apitest.Decode[[]team.Public](t, manager.Get(t, "/api/v1/teams")), 2, "logged in: all teams")

	create := map[string]string{"name": "Reserves", "ages": "99", "isActive": "true", "leagueTable": "https://x.example.test"}
	assert.Equal(t, http.StatusForbidden, manager.Multipart(t, http.MethodPost, "/api/v1/teams", create).Code)
	rec := editor.Multipart(t, http.MethodPost, "/api/v1/teams", create)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.True(t, apitest.Decode[team.Public](t, rec).IsActive)

	rec = editor.Multipart(t, http.MethodPost, "/api/v1/teams", map[string]string{"name": "No ages"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, apitest.ErrorOf(t, rec).Fields, "ages")

	rec = editor.Multipart(t, http.MethodPatch, "/api/v1/teams/1", map[string]string{"isActive": "false"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.False(t, store.row(1).IsActive)
	assert.Equal(t, "First Team", store.row(1).Name)

	assert.Equal(t, http.StatusNoContent, editor.JSON(t, http.MethodDelete, "/api/v1/teams/3", nil).Code)
}
```

`server/internal/team/handlers.go`:
```go
package team

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /teams (the detail view GET /teams/{id} is served by the
// site package, because it aggregates players, managers and sponsors).
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the team routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/teams", h.list, guards.Identify)
	g.POST("/teams", h.create, guards.Editor)
	g.PATCH("/teams/:id", h.update, guards.Editor)
	g.DELETE("/teams/:id", h.remove, guards.Editor)
}

// list returns active teams, or all teams for logged-in users.
//
//	@Summary	List teams
//	@Tags		teams
//	@Produce	json
//	@Success	200	{array}	team.Public
//	@Router		/teams [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context(), web.LoggedIn(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// create adds a team.
//
//	@Summary	Create a team
//	@Tags		teams
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		name		formData	string	true	"Name"
//	@Param		ages		formData	int		true	"Upper age (under 19 makes a youth team)"
//	@Param		description	formData	string	false	"Description"
//	@Param		league		formData	string	false	"League"
//	@Param		division	formData	string	false	"Division"
//	@Param		leagueTable	formData	string	false	"League table URL"
//	@Param		fixtures	formData	string	false	"Fixtures URL"
//	@Param		coach		formData	string	false	"Coach"
//	@Param		physio		formData	string	false	"Physio"
//	@Param		isActive	formData	bool	false	"Active"
//	@Param		isYouth		formData	bool	false	"Youth"
//	@Param		image		formData	file	false	"Team photo"
//	@Success	201			{object}	team.Public
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/teams [post]
func (h *Handlers) create(c echo.Context) error {
	ages, err := web.FormInt(c, "ages")
	if err != nil {
		return err
	}
	if ages == nil {
		return svcerr.InvalidField("ages", "ages is required")
	}
	active, err := web.FormBool(c, "isActive")
	if err != nil {
		return err
	}
	youth, err := web.FormBool(c, "isYouth")
	if err != nil {
		return err
	}
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	out, err := h.svc.Create(c.Request().Context(), CreateInput{
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
		League:      c.FormValue("league"),
		Division:    c.FormValue("division"),
		LeagueTable: c.FormValue("leagueTable"),
		Fixtures:    c.FormValue("fixtures"),
		Coach:       c.FormValue("coach"),
		Physio:      c.FormValue("physio"),
		IsActive:    active != nil && *active,
		IsYouth:     youth != nil && *youth,
		Ages:        *ages,
	}, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// update changes a team; omitted fields are left as they are.
//
//	@Summary	Update a team
//	@Tags		teams
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		id			path		int		true	"Team ID"
//	@Param		name		formData	string	false	"Name"
//	@Param		ages		formData	int		false	"Upper age"
//	@Param		description	formData	string	false	"Description; empty clears"
//	@Param		league		formData	string	false	"League; empty clears"
//	@Param		division	formData	string	false	"Division; empty clears"
//	@Param		leagueTable	formData	string	false	"League table URL; empty clears"
//	@Param		fixtures	formData	string	false	"Fixtures URL; empty clears"
//	@Param		coach		formData	string	false	"Coach; empty clears"
//	@Param		physio		formData	string	false	"Physio; empty clears"
//	@Param		isActive	formData	bool	false	"Active"
//	@Param		isYouth		formData	bool	false	"Youth"
//	@Param		image		formData	file	false	"Replacement photo"
//	@Param		removeImage	formData	bool	false	"Remove the current photo"
//	@Success	200			{object}	team.Public
//	@Failure	404			{object}	web.ErrorResponse
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/teams/{id} [patch]
func (h *Handlers) update(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	in := UpdateInput{
		Name:        web.FormString(c, "name"),
		Description: web.FormString(c, "description"),
		League:      web.FormString(c, "league"),
		Division:    web.FormString(c, "division"),
		LeagueTable: web.FormString(c, "leagueTable"),
		Fixtures:    web.FormString(c, "fixtures"),
		Coach:       web.FormString(c, "coach"),
		Physio:      web.FormString(c, "physio"),
	}
	if in.Ages, err = web.FormInt(c, "ages"); err != nil {
		return err
	}
	if in.IsActive, err = web.FormBool(c, "isActive"); err != nil {
		return err
	}
	if in.IsYouth, err = web.FormBool(c, "isYouth"); err != nil {
		return err
	}
	remove, err := web.FormBool(c, "removeImage")
	if err != nil {
		return err
	}
	in.RemoveImage = remove != nil && *remove
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	out, err := h.svc.Update(c.Request().Context(), id, in, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// remove deletes a team, unlinking its players, sponsors and managers.
//
//	@Summary	Delete a team
//	@Tags		teams
//	@Param		id	path	int	true	"Team ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/teams/{id} [delete]
func (h *Handlers) remove(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if _, err = h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

```
Run: `go test ./server/internal/team/ ./server/internal/web/`. Expected: PASS.

- [ ] **Step 7: Legacy writes → service**

Add `teamSvc` / `TeamService` to `Views`/`Deps`/`New`. In `server/internal/legacy/views/team.go` keep `TeamsFunc` and `TeamFunc`; replace the three write handlers:
```go
// legacyTeamInput reads the legacy team form. Checkbox fields are "Y" when
// ticked; unticked now means false (see Plan Decision on checkbox fields).
func legacyTeamInput(c echo.Context) (team.CreateInput, error) {
	ages, err := strconv.Atoi(c.FormValue("ages"))
	if err != nil {
		return team.CreateInput{}, fmt.Errorf("failed to parse ages: %w", err)
	}
	return team.CreateInput{
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
		League:      c.FormValue("league"),
		Division:    c.FormValue("division"),
		LeagueTable: c.FormValue("leagueTable"),
		Fixtures:    c.FormValue("fixtures"),
		Coach:       c.FormValue("coach"),
		Physio:      c.FormValue("physio"),
		IsActive:    formYes(c, "isActive"),
		IsYouth:     formYes(c, "isYouth"),
		Ages:        ages,
	}, nil
}

func (v *Views) TeamAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.TeamAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}
	in, err := legacyTeamInput(c)
	if err != nil {
		data.Error = fmt.Sprintf("failed to parse team add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for team add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	created, err := v.teamSvc.Create(c.Request().Context(), in, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add team for team add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add team for team add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", created.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) TeamEditFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.TeamEditFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to parse id for team edit: %w", err)
	}
	data := struct {
		Error string `json:"error"`
	}{}
	in, err := legacyTeamInput(c)
	if err != nil {
		data.Error = fmt.Sprintf("failed to parse team edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	remove := c.FormValue("removeTeamImage")
	if remove != "" && remove != "Y" {
		data.Error = "failed to parse removeTeamImage for team edit: " + remove
		return c.JSON(http.StatusOK, data)
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for team edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	updated, err := v.teamSvc.Update(c.Request().Context(), teamID, team.UpdateInput{
		Name: &in.Name, Description: &in.Description, League: &in.League, Division: &in.Division,
		LeagueTable: &in.LeagueTable, Fixtures: &in.Fixtures, Coach: &in.Coach, Physio: &in.Physio,
		IsActive: &in.IsActive, IsYouth: &in.IsYouth, Ages: &in.Ages, RemoveImage: remove == "Y",
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to edit team for team edit, team id: %d, error: %+v", teamID, err))
		data.Error = fmt.Sprintf("failed to edit team for team edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully edited \"%s\"", updated.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) TeamDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.TeamDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for team delete, error: %w", err)
	}
	deleted, err := v.teamSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete team for team delete, team id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", deleted.Name))
	return c.Redirect(http.StatusFound, "/teams")
}
```

- [ ] **Step 8: Wire, regenerate, run, commit**

In `app.Build`:
- Add `teamSvc := team.NewService(s.Team, uploads, s.Player, s.Sponsor, s.User)`.
- Add `TeamService: teamSvc,` to `views.Deps`.
- Add `team.NewHandlers(teamSvc).Register(api, guards)`.
```bash
go generate ./server/internal/docs
goimports -w server/internal/legacy/views
gofmt -l server; go build ./... && go vet ./... && go test ./... && go tool golangci-lint run ./...
git add -A server
git commit -m "Add team service, /api/v1/teams and Identify guard; team delete detaches via stores

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---
### Task 16: Players and the photo-privacy rule

This task owns **Review Focus #1**. `player.PhotoVisible` becomes the one place the "no photos of youth-team or under-18 players" rule lives. The API, the legacy templates (via `views/helpers.go`) and `/files/player` (Task 19) all call it.

**Files:**
- Modify: `server/internal/player/store.go` (`addPlayer` returns ID), `server/internal/player/player.go` (`EditPlayer` write-through)
- Create: `server/internal/player/privacy.go`, `privacy_test.go`
- Create: `server/internal/player/{store_test,types,service,handlers,fake_test,service_test,handlers_test}.go`
- Modify: `server/internal/legacy/views/player.go`, `server/internal/legacy/views/helpers.go`, `views.go`, `server/internal/app/app.go`, `server/internal/app/helpers_test.go`

**Interfaces:**
- Consumes: `web.FormInt` (Task 15), `team.Team` (existing).
- Produces (`player`):
  - `Age(dob, now time.Time) (int, bool)` (false when dob is in the future) and `PhotoVisible(p Player, teamIsYouth bool, now time.Time) bool`.
  - Types:
    - `Public{ID int; Name, Position string; IsCaptain bool; DateOfBirth *time.Time; Age *int; Team *TeamRef; ImageURL string}`
    - `TeamRef{ID int; Name string; IsYouth bool}`
    - `Member{ID int; Name, Position string; IsCaptain bool; ImageURL string}` (public squad entry, no DOB)
    - `CreateInput{Name, Position string; TeamID int; DateOfBirth time.Time; IsCaptain bool}`
    - `UpdateInput{Name, Position *string; TeamID *int; DateOfBirth *time.Time; IsCaptain *bool; RemoveImage bool}`
  - `type TeamGetter interface{GetTeam; GetTeams}`
  - `NewService(store, TeamGetter, *upload.Files) *Service`, with methods:
    - `List(ctx) ([]Public, error)`
    - `Squad(ctx, t team.Team) ([]Member, error)`
    - `PhotoKey(ctx, id) (string, error)` (NotFound when hidden or absent)
    - `Create`, `Update`, `Delete(ctx, id) (Public, error)`
  - `NewHandlers`, `Register`.
- Produces (legacy): `Deps.PlayerService`.

- [ ] **Step 1: The privacy rule: failing tests first**

`server/internal/player/privacy_test.go`:
```go
package player_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/player"
)

var now = time.Date(2026, 9, 27, 15, 0, 0, 0, time.UTC)

func born(y, m, d int) null.Time { return null.TimeFrom(time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)) }

func TestAge(t *testing.T) {
	cases := []struct {
		dob  null.Time
		want int
		ok   bool
	}{
		{born(2008, 9, 27), 18, true},  // 18th birthday today
		{born(2008, 9, 28), 17, true},  // tomorrow
		{born(1990, 1, 1), 36, true},
		{born(2030, 1, 1), 0, false},   // future
	}
	for _, tc := range cases {
		got, ok := player.Age(tc.dob.Time, now)
		assert.Equal(t, tc.ok, ok, tc.dob.Time)
		if tc.ok {
			assert.Equal(t, tc.want, got, tc.dob.Time)
		}
	}
}

// TestPhotoVisible pins Review Focus #1.
func TestPhotoVisible(t *testing.T) {
	photo := null.StringFrom("player/p.png")
	cases := map[string]struct {
		p     player.Player
		youth bool
		want  bool
	}{
		"adult, senior team":           {player.Player{FileName: photo, DateOfBirth: born(1990, 1, 1)}, false, true},
		"adult, youth team":            {player.Player{FileName: photo, DateOfBirth: born(1990, 1, 1)}, true, false},
		"17, senior team":              {player.Player{FileName: photo, DateOfBirth: born(2008, 9, 28)}, false, false},
		"18 today, senior team":        {player.Player{FileName: photo, DateOfBirth: born(2008, 9, 27)}, false, true},
		"no DOB, senior team":          {player.Player{FileName: photo}, false, true},
		"no DOB, youth team":           {player.Player{FileName: photo}, true, false},
		"future DOB (bad data)":        {player.Player{FileName: photo, DateOfBirth: born(2030, 1, 1)}, false, false},
		"no photo":                     {player.Player{DateOfBirth: born(1990, 1, 1)}, false, false},
		"empty-but-valid photo string": {player.Player{FileName: null.StringFrom(""), DateOfBirth: born(1990, 1, 1)}, false, false},
	}
	for name, tc := range cases {
		assert.Equal(t, tc.want, player.PhotoVisible(tc.p, tc.youth, now), name)
	}
}
```
Run: `go test ./server/internal/player/`. Expected: compile errors.

`server/internal/player/privacy.go`:
```go
package player

import "time"

// adultAge is the age from which a player's photo may be shown.
const adultAge = 18

// Age returns the age in whole years on now's date; ok is false when dob is
// after today (bad data).
func Age(dob, now time.Time) (int, bool) {
	now = now.In(dob.Location())
	ty, tm, td := now.Date()
	today := time.Date(ty, tm, td, 0, 0, 0, 0, time.UTC)
	by, bm, bd := dob.Date()
	birth := time.Date(by, bm, bd, 0, 0, 0, 0, time.UTC)
	if today.Before(birth) {
		return 0, false
	}
	age := ty - by
	if birth.AddDate(age, 0, 0).After(today) {
		age--
	}
	return age, true
}

// PhotoVisible is the single safeguarding rule for player photos: never for
// players on a youth team, never for players under 18, and never when the
// date of birth is nonsensical. Players with no date of birth on a senior
// team are shown, matching the legacy site.
func PhotoVisible(p Player, teamIsYouth bool, now time.Time) bool {
	if !p.FileName.Valid || p.FileName.String == "" || teamIsYouth {
		return false
	}
	if !p.DateOfBirth.Valid {
		return true
	}
	age, ok := Age(p.DateOfBirth.Time, now)
	return ok && age >= adultAge
}
```
Run: PASS.

- [ ] **Step 2: Store tests, then fix `addPlayer` and `EditPlayer`**

`server/internal/player/store_test.go`:
```go
package player_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestAddPlayerReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	added, err := player.NewPlayerRepo(db).AddPlayer(context.Background(), player.Player{
		Name: "New", DateOfBirth: null.TimeFrom(time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC)), TeamID: 1})
	require.NoError(t, err)
	require.Positive(t, added.ID)
}

func TestEditPlayerClearsOptionalFields(t *testing.T) {
	db, _ := testdb.Open(t)
	store := player.NewPlayerRepo(db)
	ctx := context.Background()
	p, err := store.GetPlayer(ctx, player.Player{ID: 1})
	require.NoError(t, err)
	require.True(t, p.FileName.Valid)
	require.True(t, p.Position.Valid)
	require.True(t, p.IsCaptain)

	p.FileName, p.Position, p.IsCaptain = null.String{}, null.String{}, false
	_, err = store.EditPlayer(ctx, p)
	require.NoError(t, err)

	got, err := store.GetPlayer(ctx, player.Player{ID: 1})
	require.NoError(t, err)
	assert.False(t, got.FileName.Valid)
	assert.False(t, got.Position.Valid)
	assert.False(t, got.IsCaptain)
}
```
Run (FAIL). Then:
- In `store.go` `addPlayer`, add `.Suffix("RETURNING id")` and replace the `ExecContext`/`RowsAffected` block with:
```go
	err = s.db.GetContext(ctx, &playerParam.ID, sql, args...)
	if err != nil {
		span.RecordError(err)
		return Player{}, fmt.Errorf("failed to add player: %w", err)
	}
	return playerParam, nil
```
- In `player.go`, replace `EditPlayer`:
```go
// EditPlayer writes p as given. Callers load the row first and change only
// the fields they mean to, so empty values are deliberate.
func (s *Store) EditPlayer(ctx context.Context, playerParam Player) (Player, error) {
	return s.editPlayer(ctx, playerParam)
}
```
Re-run: PASS. (Legacy `TeamDeleteFunc` used to call `EditPlayer` with full rows; since Task 15 it uses `DetachTeam`.)

- [ ] **Step 3: Fake store and failing service tests**

`server/internal/player/fake_test.go`:
```go
package player_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
)

type fakeStore struct {
	mu     sync.Mutex
	rows   map[int]player.Player
	nextID int
}

func newFakeStore(rows ...player.Player) *fakeStore {
	f := &fakeStore{rows: map[int]player.Player{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) sorted(keep func(player.Player) bool) []player.Player {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []player.Player{}
	for _, r := range f.rows {
		if keep(r) {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (f *fakeStore) GetPlayers(context.Context) ([]player.Player, error) {
	return f.sorted(func(player.Player) bool { return true }), nil
}

func (f *fakeStore) GetPlayersTeam(_ context.Context, t team.Team) ([]player.Player, error) {
	return f.sorted(func(p player.Player) bool { return p.TeamID == t.ID }), nil
}

func (f *fakeStore) GetPlayer(_ context.Context, p player.Player) (player.Player, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[p.ID]
	if !ok {
		return player.Player{}, fmt.Errorf("failed to get player: %w", sql.ErrNoRows)
	}
	return r, nil
}

func (f *fakeStore) AddPlayer(_ context.Context, p player.Player) (player.Player, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	p.ID = f.nextID
	f.rows[p.ID] = p
	return p, nil
}

func (f *fakeStore) EditPlayer(_ context.Context, p player.Player) (player.Player, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.rows[p.ID]; !ok {
		return player.Player{}, errors.New("no such row")
	}
	f.rows[p.ID] = p
	return p, nil
}

func (f *fakeStore) DeletePlayer(_ context.Context, p player.Player) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, p.ID)
	return nil
}

func (f *fakeStore) row(id int) player.Player {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.rows[id]
}

type fakeTeams map[int]team.Team

func (f fakeTeams) GetTeam(_ context.Context, t team.Team) (team.Team, error) {
	if x, ok := f[t.ID]; ok {
		return x, nil
	}
	return team.Team{}, fmt.Errorf("failed to get team: %w", sql.ErrNoRows)
}

func (f fakeTeams) GetTeams(context.Context) ([]team.Team, error) {
	out := make([]team.Team, 0, len(f))
	for _, t := range f {
		out = append(out, t)
	}
	return out, nil
}
```

`server/internal/player/service_test.go`:
```go
package player_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

var ctx = context.Background()

func yearsAgo(n int) null.Time { return null.TimeFrom(time.Now().AddDate(-n, 0, -1)) }

var teams = fakeTeams{
	1: {ID: 1, Name: "First Team", IsActive: true, Ages: 99},
	2: {ID: 2, Name: "Under 12s", IsActive: true, IsYouth: true, Ages: 12},
}

func squad() []player.Player {
	return []player.Player{
		{ID: 1, Name: "Adult", FileName: null.StringFrom("player/adult.png"), DateOfBirth: yearsAgo(30),
			Position: null.StringFrom("Striker"), IsCaptain: true, TeamID: 1},
		{ID: 2, Name: "Youth", FileName: null.StringFrom("player/youth.png"), DateOfBirth: yearsAgo(11), TeamID: 2},
		{ID: 3, Name: "Young Senior", FileName: null.StringFrom("player/young.png"), DateOfBirth: yearsAgo(16), TeamID: 1},
	}
}

func newService() (*player.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore(squad()...)
	objects := uploadtest.New()
	for _, p := range squad() {
		objects.Objects[p.FileName.String] = "IMG"
	}
	return player.NewService(store, teams, upload.New(objects)), store, objects
}

func fieldsOf(t *testing.T, err error) map[string]string {
	t.Helper()
	se, ok := svcerr.As(err)
	require.True(t, ok, "want svcerr, got %v", err)
	return se.Fields
}

// TestListHidesMinorPhotos pins Review Focus #1 for /players.
func TestListHidesMinorPhotos(t *testing.T) {
	svc, _, _ := newService()
	list, err := svc.List(ctx)
	require.NoError(t, err)
	byID := map[int]player.Public{}
	for _, p := range list {
		byID[p.ID] = p
	}
	assert.NotEmpty(t, byID[1].ImageURL)
	assert.Empty(t, byID[2].ImageURL, "youth team")
	assert.Empty(t, byID[3].ImageURL, "under 18")
	require.NotNil(t, byID[1].Team)
	assert.Equal(t, "First Team", byID[1].Team.Name)
	require.NotNil(t, byID[1].Age)
	assert.Equal(t, 30, *byID[1].Age)
}

func TestSquadHidesMinorPhotos(t *testing.T) {
	svc, _, _ := newService()
	members, err := svc.Squad(ctx, team.Team{ID: 1, IsYouth: false})
	require.NoError(t, err)
	require.Len(t, members, 2)
	for _, m := range members {
		if m.ID == 3 {
			assert.Empty(t, m.ImageURL)
		} else {
			assert.NotEmpty(t, m.ImageURL)
		}
	}
}

func TestPhotoKey(t *testing.T) {
	svc, _, _ := newService()
	key, err := svc.PhotoKey(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "player/adult.png", key)
	for _, id := range []int{2, 3, 99} {
		_, err = svc.PhotoKey(ctx, id)
		se, ok := svcerr.As(err)
		require.True(t, ok, id)
		assert.Equal(t, svcerr.KindNotFound, se.Kind, id)
	}
}

func TestCreateValidates(t *testing.T) {
	svc, _, _ := newService()
	dob := time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Contains(t, fieldsOf(t, errOf(svc.Create(ctx, player.CreateInput{TeamID: 1, DateOfBirth: dob}, nil))), "name")
	assert.Contains(t, fieldsOf(t, errOf(svc.Create(ctx, player.CreateInput{Name: "x", TeamID: 9, DateOfBirth: dob}, nil))), "teamId")
	assert.Contains(t, fieldsOf(t, errOf(svc.Create(ctx, player.CreateInput{Name: "x", TeamID: 1}, nil))), "dateOfBirth")
	assert.Contains(t, fieldsOf(t, errOf(svc.Create(ctx, player.CreateInput{Name: "x", TeamID: 1, DateOfBirth: time.Now().AddDate(0, 0, 1)}, nil))), "dateOfBirth")
}

func errOf(_ player.Public, err error) error { return err }

func TestUpdateCanUncaptainAndRemoveImage(t *testing.T) {
	svc, store, objects := newService()
	no := false
	_, err := svc.Update(ctx, 1, player.UpdateInput{IsCaptain: &no, RemoveImage: true}, nil)
	require.NoError(t, err)
	row := store.row(1)
	assert.False(t, row.IsCaptain, "legacy edit could never un-captain")
	assert.False(t, row.FileName.Valid)
	assert.Equal(t, "Striker", row.Position.String, "omitted fields unchanged")
	assert.Equal(t, []string{"player/adult.png"}, objects.Deleted)
}

func TestDelete(t *testing.T) {
	svc, store, objects := newService()
	deleted, err := svc.Delete(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, "Youth", deleted.Name)
	assert.Empty(t, deleted.ImageURL, "the deleted record never exposes a minor's photo either")
	_, still := store.rows[2]
	assert.False(t, still)
	assert.Equal(t, []string{"player/youth.png"}, objects.Deleted)
}
```
Run: compile errors expected.

- [ ] **Step 4: Implement `types.go` and `service.go`**

`server/internal/player/types.go`:
```go
package player

import "time"

// Public is a player as the logged-in players list returns it. ImageURL is
// empty whenever PhotoVisible says so.
type Public struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Position    string     `json:"position,omitempty"`
	IsCaptain   bool       `json:"isCaptain"`
	DateOfBirth *time.Time `json:"dateOfBirth,omitempty"`
	Age         *int       `json:"age,omitempty"`
	Team        *TeamRef   `json:"team,omitempty"`
	ImageURL    string     `json:"imageUrl,omitempty"`
}

// TeamRef is the team a player belongs to.
type TeamRef struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	IsYouth bool   `json:"isYouth"`
}

// Member is a player on a public team page: no date of birth.
type Member struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Position  string `json:"position,omitempty"`
	IsCaptain bool   `json:"isCaptain"`
	ImageURL  string `json:"imageUrl,omitempty"`
}

// CreateInput is a new player.
type CreateInput struct {
	Name        string
	Position    string
	TeamID      int
	DateOfBirth time.Time
	IsCaptain   bool
}

// UpdateInput changes a player; nil fields are left unchanged.
type UpdateInput struct {
	Name        *string
	Position    *string
	TeamID      *int
	DateOfBirth *time.Time
	IsCaptain   *bool
	RemoveImage bool
}
```

`server/internal/player/service.go`:
```go
package player

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

const uploadCategory = "player"

type store interface {
	GetPlayers(ctx context.Context) ([]Player, error)
	GetPlayersTeam(ctx context.Context, teamParam team.Team) ([]Player, error)
	GetPlayer(ctx context.Context, playerParam Player) (Player, error)
	AddPlayer(ctx context.Context, playerParam Player) (Player, error)
	EditPlayer(ctx context.Context, playerParam Player) (Player, error)
	DeletePlayer(ctx context.Context, playerParam Player) error
}

// TeamGetter reads teams (satisfied by *team.Store).
type TeamGetter interface {
	GetTeam(ctx context.Context, teamParam team.Team) (team.Team, error)
	GetTeams(ctx context.Context) ([]team.Team, error)
}

// Service is the player business logic shared by the API and legacy views.
type Service struct {
	store store
	teams TeamGetter
	files *upload.Files
}

func NewService(store store, teams TeamGetter, files *upload.Files) *Service {
	return &Service{store: store, teams: teams, files: files}
}

func (s *Service) public(p Player, t *team.Team, now time.Time) Public {
	out := Public{ID: p.ID, Name: p.Name, Position: p.Position.String, IsCaptain: p.IsCaptain}
	if p.DateOfBirth.Valid {
		dob := p.DateOfBirth.Time
		out.DateOfBirth = &dob
		if age, ok := Age(dob, now); ok {
			out.Age = &age
		}
	}
	youth := false
	if t != nil {
		out.Team = &TeamRef{ID: t.ID, Name: t.Name, IsYouth: t.IsYouth}
		youth = t.IsYouth
	}
	if PhotoVisible(p, youth, now) {
		out.ImageURL = s.files.URL(p.FileName.String)
	}
	return out
}

func (s *Service) List(ctx context.Context) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "player.Service.List")
	defer span.End()
	rows, err := s.store.GetPlayers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list players: %w", err)
	}
	teamRows, err := s.teams.GetTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	byID := make(map[int]team.Team, len(teamRows))
	for _, t := range teamRows {
		byID[t.ID] = t
	}
	now := time.Now()
	out := make([]Public, 0, len(rows))
	for _, p := range rows {
		var t *team.Team
		if found, ok := byID[p.TeamID]; ok {
			t = &found
		}
		out = append(out, s.public(p, t, now))
	}
	return out, nil
}

// Squad lists a team's players for its public page.
func (s *Service) Squad(ctx context.Context, t team.Team) ([]Member, error) {
	ctx, span := tracer.Start(ctx, "player.Service.Squad")
	defer span.End()
	rows, err := s.store.GetPlayersTeam(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("failed to list squad: %w", err)
	}
	now := time.Now()
	out := make([]Member, 0, len(rows))
	for _, p := range rows {
		m := Member{ID: p.ID, Name: p.Name, Position: p.Position.String, IsCaptain: p.IsCaptain}
		if PhotoVisible(p, t.IsYouth, now) {
			m.ImageURL = s.files.URL(p.FileName.String)
		}
		out = append(out, m)
	}
	return out, nil
}

// PhotoKey returns the storage key of a player's photo, or NotFound when
// there is none or it must not be shown.
func (s *Service) PhotoKey(ctx context.Context, id int) (string, error) {
	ctx, span := tracer.Start(ctx, "player.Service.PhotoKey")
	defer span.End()
	notFound := svcerr.NotFound("player photo not found", nil)
	p, err := s.store.GetPlayer(ctx, Player{ID: id})
	if err != nil {
		return "", svcerr.FromStore(err, "player")
	}
	t, err := s.teams.GetTeam(ctx, team.Team{ID: p.TeamID})
	if err != nil {
		// Unknown team: we can't prove it isn't a youth team, so hide.
		return "", notFound
	}
	if !PhotoVisible(p, t.IsYouth, time.Now()) {
		return "", notFound
	}
	return p.FileName.String, nil
}

func (s *Service) get(ctx context.Context, id int) (Player, error) {
	p, err := s.store.GetPlayer(ctx, Player{ID: id})
	if err != nil {
		return Player{}, svcerr.FromStore(err, "player")
	}
	return p, nil
}

func (s *Service) teamFor(ctx context.Context, id int) (*team.Team, error) {
	t, err := s.teams.GetTeam(ctx, team.Team{ID: id})
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Service) validate(ctx context.Context, p Player) error {
	f := svcerr.Fields{}
	if strings.TrimSpace(p.Name) == "" {
		f.Add("name", "name is required")
	}
	if _, err := s.teamFor(ctx, p.TeamID); err != nil {
		f.Add("teamId", "team does not exist")
	}
	if !p.DateOfBirth.Valid || p.DateOfBirth.Time.IsZero() {
		f.Add("dateOfBirth", "date of birth is required")
	} else if !p.DateOfBirth.Time.Before(time.Now()) {
		f.Add("dateOfBirth", "date of birth must be in the past")
	}
	return f.Err()
}

func (s *Service) Create(ctx context.Context, in CreateInput, image *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "player.Service.Create")
	defer span.End()
	position := strings.TrimSpace(in.Position)
	p := Player{
		Name:        strings.TrimSpace(in.Name),
		Position:    null.NewString(position, position != ""),
		TeamID:      in.TeamID,
		IsCaptain:   in.IsCaptain,
		DateOfBirth: null.NewTime(in.DateOfBirth, !in.DateOfBirth.IsZero()),
	}
	if err := s.validate(ctx, p); err != nil {
		return Public{}, err
	}
	if image != nil {
		key, err := s.files.Save(ctx, image, uploadCategory)
		if err != nil {
			return Public{}, err
		}
		p.FileName = null.StringFrom(key)
	}
	added, err := s.store.AddPlayer(ctx, p)
	if err != nil {
		s.files.Remove(ctx, p.FileName.String)
		return Public{}, fmt.Errorf("failed to add player: %w", err)
	}
	t, _ := s.teamFor(ctx, added.TeamID) //nolint:errcheck // validated above
	return s.public(added, t, time.Now()), nil
}

func (s *Service) Update(ctx context.Context, id int, in UpdateInput, image *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "player.Service.Update")
	defer span.End()
	p, err := s.get(ctx, id)
	if err != nil {
		return Public{}, err
	}
	if in.Name != nil {
		p.Name = strings.TrimSpace(*in.Name)
	}
	if in.Position != nil {
		position := strings.TrimSpace(*in.Position)
		p.Position = null.NewString(position, position != "")
	}
	if in.TeamID != nil {
		p.TeamID = *in.TeamID
	}
	if in.DateOfBirth != nil {
		p.DateOfBirth = null.NewTime(*in.DateOfBirth, !in.DateOfBirth.IsZero())
	}
	if in.IsCaptain != nil {
		p.IsCaptain = *in.IsCaptain
	}
	if err = s.validate(ctx, p); err != nil {
		return Public{}, err
	}

	oldKey, newKey := p.FileName.String, ""
	switch {
	case image != nil:
		if newKey, err = s.files.Save(ctx, image, uploadCategory); err != nil {
			return Public{}, err
		}
		p.FileName = null.StringFrom(newKey)
	case in.RemoveImage:
		p.FileName = null.String{}
	}
	if _, err = s.store.EditPlayer(ctx, p); err != nil {
		s.files.Remove(ctx, newKey)
		return Public{}, fmt.Errorf("failed to edit player: %w", err)
	}
	if oldKey != "" && oldKey != p.FileName.String {
		s.files.Remove(ctx, oldKey)
	}
	t, _ := s.teamFor(ctx, p.TeamID) //nolint:errcheck // validated above
	return s.public(p, t, time.Now()), nil
}

// Delete removes the player and their photo, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Public, error) {
	ctx, span := tracer.Start(ctx, "player.Service.Delete")
	defer span.End()
	p, err := s.get(ctx, id)
	if err != nil {
		return Public{}, err
	}
	t, _ := s.teamFor(ctx, p.TeamID) //nolint:errcheck // a missing team only affects display
	deleted := s.public(p, t, time.Now())
	if err = s.store.DeletePlayer(ctx, p); err != nil {
		return Public{}, fmt.Errorf("failed to delete player: %w", err)
	}
	s.files.Remove(ctx, p.FileName.String)
	return deleted, nil
}
```
Run: `go test ./server/internal/player/`. Expected: PASS.

- [ ] **Step 5: Failing handler tests, then `handlers.go`**

`server/internal/player/handlers_test.go`:
```go
package player_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestPlayerRoutes(t *testing.T) {
	svc, store, _ := newService()
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	player.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	editor := anon.As(authtest.Cookie(t, sessions, authtest.Webmaster))
	manager := anon.As(authtest.Cookie(t, sessions, authtest.Manager))

	assert.Equal(t, http.StatusUnauthorized, anon.Get(t, "/api/v1/players").Code, "players list is login-only")
	rec := manager.Get(t, "/api/v1/players")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, apitest.Decode[[]player.Public](t, rec), 3)

	create := map[string]string{"name": "Signing", "teamId": "1", "dateOfBirth": "1999-04-01", "isCaptain": "false"}
	assert.Equal(t, http.StatusForbidden, manager.Multipart(t, http.MethodPost, "/api/v1/players", create).Code)
	rec = editor.Multipart(t, http.MethodPost, "/api/v1/players", create)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	rec = editor.Multipart(t, http.MethodPatch, "/api/v1/players/1", map[string]string{"isCaptain": "false"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.False(t, store.row(1).IsCaptain)

	assert.Equal(t, http.StatusNoContent, editor.JSON(t, http.MethodDelete, "/api/v1/players/3", nil).Code)
}
```

`server/internal/player/handlers.go`:
```go
package player

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /players.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the player routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/players", h.list, guards.Login)
	g.POST("/players", h.create, guards.Editor)
	g.PATCH("/players/:id", h.update, guards.Editor)
	g.DELETE("/players/:id", h.remove, guards.Editor)
}

// list returns every player (logged-in users only).
//
//	@Summary	List players
//	@Tags		players
//	@Produce	json
//	@Success	200	{array}		player.Public
//	@Failure	401	{object}	web.ErrorResponse
//	@Router		/players [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// create adds a player.
//
//	@Summary	Create a player
//	@Tags		players
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		name		formData	string	true	"Name"
//	@Param		teamId		formData	int		true	"Team ID"
//	@Param		dateOfBirth	formData	string	true	"YYYY-MM-DD"
//	@Param		position	formData	string	false	"Position"
//	@Param		isCaptain	formData	bool	false	"Captain"
//	@Param		image		formData	file	false	"Photo (never shown for youth or under-18 players)"
//	@Success	201			{object}	player.Public
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/players [post]
func (h *Handlers) create(c echo.Context) error {
	teamID, err := web.FormInt(c, "teamId")
	if err != nil {
		return err
	}
	dob, err := web.FormDate(c, "dateOfBirth")
	if err != nil {
		return err
	}
	captain, err := web.FormBool(c, "isCaptain")
	if err != nil {
		return err
	}
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	in := CreateInput{Name: c.FormValue("name"), Position: c.FormValue("position"), IsCaptain: captain != nil && *captain}
	if teamID != nil {
		in.TeamID = *teamID
	}
	if dob != nil {
		in.DateOfBirth = *dob
	}
	out, err := h.svc.Create(c.Request().Context(), in, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// update changes a player; omitted fields are left as they are.
//
//	@Summary	Update a player
//	@Tags		players
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		id			path		int		true	"Player ID"
//	@Param		name		formData	string	false	"Name"
//	@Param		teamId		formData	int		false	"Team ID"
//	@Param		dateOfBirth	formData	string	false	"YYYY-MM-DD"
//	@Param		position	formData	string	false	"Position; empty clears"
//	@Param		isCaptain	formData	bool	false	"Captain"
//	@Param		image		formData	file	false	"Replacement photo"
//	@Param		removeImage	formData	bool	false	"Remove the current photo"
//	@Success	200			{object}	player.Public
//	@Failure	404			{object}	web.ErrorResponse
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/players/{id} [patch]
func (h *Handlers) update(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	in := UpdateInput{Name: web.FormString(c, "name"), Position: web.FormString(c, "position")}
	if in.TeamID, err = web.FormInt(c, "teamId"); err != nil {
		return err
	}
	if in.DateOfBirth, err = web.FormDate(c, "dateOfBirth"); err != nil {
		return err
	}
	if in.IsCaptain, err = web.FormBool(c, "isCaptain"); err != nil {
		return err
	}
	remove, err := web.FormBool(c, "removeImage")
	if err != nil {
		return err
	}
	in.RemoveImage = remove != nil && *remove
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	out, err := h.svc.Update(c.Request().Context(), id, in, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// remove deletes a player.
//
//	@Summary	Delete a player
//	@Tags		players
//	@Param		id	path	int	true	"Player ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/players/{id} [delete]
func (h *Handlers) remove(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if _, err = h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
```
Run: `go test ./server/internal/player/`. Expected: PASS.

- [ ] **Step 6: Legacy writes → service; legacy templates use `PhotoVisible`**

Add `playerSvc` / `PlayerService` to `Views`/`Deps`/`New`. In `server/internal/legacy/views/player.go` keep `PlayersFunc`; replace:
```go
// legacyPlayerInput reads the legacy player form.
func legacyPlayerInput(c echo.Context) (player.CreateInput, error) {
	teamID, err := strconv.Atoi(c.FormValue("playerTeam"))
	if err != nil {
		return player.CreateInput{}, fmt.Errorf("failed to parse playerTeam: %w", err)
	}
	dob, err := time.Parse("02/01/2006", c.FormValue("dateOfBirth"))
	if err != nil {
		return player.CreateInput{}, fmt.Errorf("failed to parse dateOfBirth: %w", err)
	}
	return player.CreateInput{
		Name:        c.FormValue("name"),
		Position:    c.FormValue("position"),
		TeamID:      teamID,
		DateOfBirth: dob,
		IsCaptain:   formYes(c, "isCaptain"),
	}, nil
}

func (v *Views) PlayerAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.PlayerAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}
	in, err := legacyPlayerInput(c)
	if err != nil {
		data.Error = fmt.Sprintf("failed to parse player add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for player add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	created, err := v.playerSvc.Create(c.Request().Context(), in, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add player for player add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add player for player add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", created.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) PlayerEditFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.PlayerEditFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	playerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to parse id for player edit, error: %w", err)
	}
	data := struct {
		Error string `json:"error"`
	}{}
	in, err := legacyPlayerInput(c)
	if err != nil {
		data.Error = fmt.Sprintf("failed to parse player edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	remove := c.FormValue("removePlayerImage")
	if remove != "" && remove != "Y" {
		data.Error = "failed to parse removePlayerImage for player edit: " + remove
		return c.JSON(http.StatusOK, data)
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for player edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	updated, err := v.playerSvc.Update(c.Request().Context(), playerID, player.UpdateInput{
		Name: &in.Name, Position: &in.Position, TeamID: &in.TeamID, DateOfBirth: &in.DateOfBirth,
		IsCaptain: &in.IsCaptain, RemoveImage: remove == "Y",
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to edit player for player edit, player id: %d, error: %+v", playerID, err))
		data.Error = fmt.Sprintf("failed to edit player for player edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully edited \"%s\"", updated.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) PlayerDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.PlayerDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for player delete, error: %w", err)
	}
	deleted, err := v.playerSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete player for player delete, player id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", deleted.Name))
	return c.Redirect(http.StatusFound, "/players")
}
```
In `server/internal/legacy/views/helpers.go`, route both template formatters through the shared rule:
- In `DBPlayersToTemplateFormat`, replace the block starting `// Photos of youth-team or under-18 players are never exposed` (the `if playerTemplate.IsFileValid && !playerTemplate.Team.IsYouth && …` statement) with:
```go
		if player.PhotoVisible(playerDB, playerTemplate.Team.IsYouth, time.Now()) {
			playerTemplate.FileName = playerDB.FileName
		}
```
- In `DBPlayersTeamToTemplateFormat`, replace the whole `age := -1` … `if playerTemplate.IsFileValid && !isYouthTeam …{…}` section with:
```go
		if player.PhotoVisible(playerDB, isYouthTeam, time.Now()) {
			playerTemplate.FileName = playerDB.FileName
		}
```
The display `Age` computation in `DBPlayersToTemplateFormat` stays as it is (it only formats text).

- [ ] **Step 7: Wire, add the guarded GET, regenerate, run, commit**

In `app.Build`:
- Add `playerSvc := player.NewService(s.Player, s.Team, uploads)`.
- Add `PlayerService: playerSvc,` to `views.Deps`.
- Add `player.NewHandlers(playerSvc).Register(api, guards)`.

In `server/internal/app/helpers_test.go` append `"/api/v1/players",` to `guardedGets`.
```bash
go generate ./server/internal/docs
goimports -w server/internal/legacy/views
gofmt -l server; go build ./... && go vet ./... && go test ./... && go tool golangci-lint run ./...
git add -A server
git commit -m "Add player service and /api/v1/players; single photo-privacy rule shared with legacy

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```
With `AFC_TEST_DB` set, the legacy smoke test still covers `/team/1`, now rendered through `PhotoVisible`.

---
### Task 17: Info page content and display email (`setting`)

Both values live in the `settings` table (`infoContent`, `displayEmail`). Email syntax checks use `emailverifier.IsAddressValid`. It gives the same syntax verdict as legacy's `verifier.Verify(...).Syntax.Valid` but does no DNS lookups, so tests stay offline.

**Files:**
- Create: `server/internal/setting/{types,service,handlers,fake_test,service_test}.go`
- Modify: `server/internal/legacy/views/info.go`, `server/internal/legacy/views/user.go` (`UsersSetDisplayEmailFunc` only), `views.go`, `server/internal/app/app.go`

**Interfaces:**
- Produces (`setting`):
  - `NewService(store) *Service`, with methods:
    - `Info(ctx) (string, error)` ("" when unset)
    - `SetInfo(ctx, html string) (string, error)` (sanitised and upserted)
    - `DisplayEmail(ctx) (string, error)` ("" when unset)
    - `SetDisplayEmail(ctx, email string) (string, error)` (empty deletes the setting)
  - Types: `InfoContent{Content string}` and `DisplayEmail{Email string}` (JSON bodies).
  - `NewHandlers`, `Register` (`GET /info`, `PUT /info` ✏️, `PUT /settings/display-email` 🛡).
- Produces (legacy): `Deps.SettingService`.

- [ ] **Step 1: Fake store and failing tests**

`server/internal/setting/fake_test.go`:
```go
package setting_test

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/setting"
)

type fakeStore struct {
	mu   sync.Mutex
	rows map[string]string
}

func newFakeStore(kv map[string]string) *fakeStore {
	if kv == nil {
		kv = map[string]string{}
	}
	return &fakeStore{rows: kv}
}

func (f *fakeStore) GetSetting(_ context.Context, id string) (setting.Setting, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.rows[id]
	if !ok {
		return setting.Setting{}, fmt.Errorf("failed to get setting: %w", sql.ErrNoRows)
	}
	return setting.Setting{ID: id, SettingText: v}, nil
}

func (f *fakeStore) AddSetting(_ context.Context, s setting.Setting) (setting.Setting, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.rows[s.ID]; ok {
		return setting.Setting{}, fmt.Errorf("duplicate key %s", s.ID)
	}
	f.rows[s.ID] = s.SettingText
	return s, nil
}

func (f *fakeStore) EditSetting(_ context.Context, s setting.Setting) (setting.Setting, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rows[s.ID] = s.SettingText
	return s, nil
}

func (f *fakeStore) DeleteSetting(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, id)
	return nil
}
```

`server/internal/setting/service_test.go`:
```go
package setting_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/setting"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

var ctx = context.Background()

func TestInfo(t *testing.T) {
	store := newFakeStore(nil)
	svc := setting.NewService(store)

	got, err := svc.Info(ctx)
	require.NoError(t, err)
	assert.Empty(t, got, "unset is empty, not an error")

	saved, err := svc.SetInfo(ctx, `<div>History</div><script>x</script>`)
	require.NoError(t, err)
	assert.Equal(t, "<div>History</div>", saved)
	saved, err = svc.SetInfo(ctx, `<div>Updated</div>`) // second write edits rather than inserts
	require.NoError(t, err)
	assert.Equal(t, "<div>Updated</div>", store.rows["infoContent"])
	assert.Equal(t, saved, store.rows["infoContent"])
}

func TestDisplayEmail(t *testing.T) {
	store := newFakeStore(map[string]string{"displayEmail": "old@example.test"})
	svc := setting.NewService(store)

	_, err := svc.SetDisplayEmail(ctx, "not-an-email")
	se, ok := svcerr.As(err)
	require.True(t, ok)
	assert.Contains(t, se.Fields, "email")
	assert.Equal(t, "old@example.test", store.rows["displayEmail"])

	_, err = svc.SetDisplayEmail(ctx, "hello@example.test")
	require.NoError(t, err)
	got, _ := svc.DisplayEmail(ctx)
	assert.Equal(t, "hello@example.test", got)

	_, err = svc.SetDisplayEmail(ctx, "")
	require.NoError(t, err)
	_, present := store.rows["displayEmail"]
	assert.False(t, present, "empty removes the setting, as legacy does")
}

func TestSettingRoutes(t *testing.T) {
	svc := setting.NewService(newFakeStore(map[string]string{"infoContent": "<div>Hi</div>"}))
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	setting.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	editor := anon.As(authtest.Cookie(t, sessions, authtest.Treasurer))
	secretary := anon.As(authtest.Cookie(t, sessions, authtest.Webmaster))

	rec := anon.Get(t, "/api/v1/info")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "<div>Hi</div>", apitest.Decode[setting.InfoContent](t, rec).Content)

	assert.Equal(t, http.StatusUnauthorized, anon.JSON(t, http.MethodPut, "/api/v1/info", setting.InfoContent{Content: "x"}).Code)
	assert.Equal(t, http.StatusOK, editor.JSON(t, http.MethodPut, "/api/v1/info", setting.InfoContent{Content: "<p>x</p>"}).Code)

	assert.Equal(t, http.StatusForbidden, editor.JSON(t, http.MethodPut, "/api/v1/settings/display-email", setting.DisplayEmail{Email: "a@b.test"}).Code,
		"treasurer is an editor but not club secretary or higher")
	rec = secretary.JSON(t, http.MethodPut, "/api/v1/settings/display-email", setting.DisplayEmail{Email: "a@b.test"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "a@b.test", apitest.Decode[setting.DisplayEmail](t, rec).Email)
}
```
Run: `go test ./server/internal/setting/`. Expected: compile errors.

- [ ] **Step 2: Implement**

`server/internal/setting/types.go`:
```go
package setting

// InfoContent is the body of GET and PUT /info.
type InfoContent struct {
	Content string `json:"content"`
}

// DisplayEmail is the body of PUT /settings/display-email.
type DisplayEmail struct {
	Email string `json:"email"`
}
```

`server/internal/setting/service.go`:
```go
package setting

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	emailverifier "github.com/AfterShip/email-verifier"

	"github.com/COMTOP1/AFC-GO/server/internal/sanitize"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
)

const (
	infoContentID  = "infoContent"
	displayEmailID = "displayEmail"
)

type store interface {
	GetSetting(ctx context.Context, settingID string) (Setting, error)
	AddSetting(ctx context.Context, settingParam Setting) (Setting, error)
	EditSetting(ctx context.Context, settingParam Setting) (Setting, error)
	DeleteSetting(ctx context.Context, settingID string) error
}

// Service manages the editable site settings.
type Service struct {
	store store
}

func NewService(store store) *Service {
	return &Service{store: store}
}

func (s *Service) get(ctx context.Context, id string) (string, error) {
	v, err := s.store.GetSetting(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get setting %s: %w", id, err)
	}
	return v.SettingText, nil
}

// put inserts or updates a setting.
func (s *Service) put(ctx context.Context, id, text string) error {
	_, err := s.store.GetSetting(ctx, id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		_, err = s.store.AddSetting(ctx, Setting{ID: id, SettingText: text})
	case err == nil:
		_, err = s.store.EditSetting(ctx, Setting{ID: id, SettingText: text})
	}
	if err != nil {
		return fmt.Errorf("failed to save setting %s: %w", id, err)
	}
	return nil
}

// Info returns the "about the club" page HTML ("" when never set).
func (s *Service) Info(ctx context.Context) (string, error) {
	ctx, span := tracer.Start(ctx, "setting.Service.Info")
	defer span.End()
	return s.get(ctx, infoContentID)
}

// SetInfo sanitises and saves the info page HTML, returning what was stored.
func (s *Service) SetInfo(ctx context.Context, html string) (string, error) {
	ctx, span := tracer.Start(ctx, "setting.Service.SetInfo")
	defer span.End()
	safe := sanitize.HTML(html)
	if err := s.put(ctx, infoContentID, safe); err != nil {
		return "", err
	}
	return safe, nil
}

// DisplayEmail returns the public contact email ("" when not set).
func (s *Service) DisplayEmail(ctx context.Context) (string, error) {
	ctx, span := tracer.Start(ctx, "setting.Service.DisplayEmail")
	defer span.End()
	return s.get(ctx, displayEmailID)
}

// SetDisplayEmail validates and saves the public contact email; an empty
// email removes it.
func (s *Service) SetDisplayEmail(ctx context.Context, email string) (string, error) {
	ctx, span := tracer.Start(ctx, "setting.Service.SetDisplayEmail")
	defer span.End()
	email = strings.TrimSpace(email)
	if email == "" {
		if err := s.store.DeleteSetting(ctx, displayEmailID); err != nil {
			return "", fmt.Errorf("failed to delete display email: %w", err)
		}
		return "", nil
	}
	if !emailverifier.IsAddressValid(email) {
		return "", svcerr.InvalidField("email", "email address is not valid")
	}
	if err := s.put(ctx, displayEmailID, email); err != nil {
		return "", err
	}
	return email, nil
}
```

`server/internal/setting/handlers.go`:
```go
package setting

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /info and /settings.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the settings routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/info", h.info)
	g.PUT("/info", h.setInfo, guards.Editor)
	g.PUT("/settings/display-email", h.setDisplayEmail, guards.ClubSecretaryHigher)
}

// info returns the club information page HTML.
//
//	@Summary	Get the info page
//	@Tags		settings
//	@Produce	json
//	@Success	200	{object}	setting.InfoContent
//	@Router		/info [get]
func (h *Handlers) info(c echo.Context) error {
	content, err := h.svc.Info(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, InfoContent{Content: content})
}

// setInfo replaces the info page HTML (sanitised).
//
//	@Summary	Update the info page
//	@Tags		settings
//	@Accept		json
//	@Produce	json
//	@Param		body	body		setting.InfoContent	true	"HTML content"
//	@Success	200		{object}	setting.InfoContent
//	@Router		/info [put]
func (h *Handlers) setInfo(c echo.Context) error {
	var in InfoContent
	if err := web.BindJSON(c, &in); err != nil {
		return err
	}
	saved, err := h.svc.SetInfo(c.Request().Context(), in.Content)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, InfoContent{Content: saved})
}

// setDisplayEmail sets (or, with an empty email, clears) the public contact email.
//
//	@Summary	Set the public contact email
//	@Tags		settings
//	@Accept		json
//	@Produce	json
//	@Param		body	body		setting.DisplayEmail	true	"Email; empty removes it"
//	@Success	200		{object}	setting.DisplayEmail
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/settings/display-email [put]
func (h *Handlers) setDisplayEmail(c echo.Context) error {
	var in DisplayEmail
	if err := web.BindJSON(c, &in); err != nil {
		return err
	}
	saved, err := h.svc.SetDisplayEmail(c.Request().Context(), in.Email)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, DisplayEmail{Email: saved})
}
```
Run: `go test ./server/internal/setting/`. Expected: PASS.

- [ ] **Step 3: Legacy writes → service**

Add `settingSvc` / `SettingService` to `Views`/`Deps`/`New`.

In `server/internal/legacy/views/info.go` replace `_infoEditPost`:
```go
func (v *Views) _infoEditPost(c echo.Context) error {
	if _, err := v.settingSvc.SetInfo(c.Request().Context(), c.FormValue("htmlContent")); err != nil {
		return fmt.Errorf("failed to save info content, error: %w", err)
	}
	return c.Redirect(http.StatusFound, "/info/edit")
}
```
In `server/internal/legacy/views/user.go` replace `UsersSetDisplayEmailFunc`:
```go
func (v *Views) UsersSetDisplayEmailFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.UsersSetDisplayEmailFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	var data struct {
		Error string `json:"error"`
	}
	if _, err := v.settingSvc.SetDisplayEmail(c.Request().Context(), c.FormValue("email")); err != nil {
		slog.Info(fmt.Sprintf("failed to set display email, error: %+v", err))
		data.Error = fmt.Sprintf("failed to set display email: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, "successfully edited display email")
	return c.JSON(http.StatusOK, data)
}
```

- [ ] **Step 4: Wire, regenerate, run, commit**

In `app.Build`:
- Add `settingSvc := setting.NewService(s.Setting)`.
- Add `SettingService: settingSvc,` to `views.Deps`.
- Add `setting.NewHandlers(settingSvc).Register(api, guards)`.
```bash
go generate ./server/internal/docs
goimports -w server/internal/legacy/views
gofmt -l server; go build ./... && go vet ./... && go test ./... && go tool golangci-lint run ./...
git add -A server
git commit -m "Add settings service (/info, display email); legacy writes use the service

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---
### Task 18: `site`: layout data, home, contact, team detail

`site` sits above the domain packages and serves read-only aggregates that span them:
- `GET /site`: layout data (nav teams, display email, visitor count, year, version)
- `GET /home`
- `GET /contact`
- `GET /teams/{id}`: team + managers + sponsors + squad; see Task 15's import-cycle note

The home page stays *fail-soft* like legacy `HomeFunc`: if one panel's query fails, it is logged and left out rather than failing the page.

**Files:**
- Create: `server/internal/site/{types,service,handlers,service_test}.go`
- Modify: `server/internal/app/app.go`

**Interfaces:**
- Consumes:
  - `news.Service.Latest`, `whatson.Service.Next`
  - `sponsor.Service.ListMinimal/ForTeam`, `affiliation.Service.ListMinimal`
  - `team.Service.List/Get`, `player.Service.Squad`
  - `setting.Service.DisplayEmail`
  - `user.Store.GetUsersContact/GetUsersManagersTeam`
  - `visitors.Counter.Count`
- Produces (`site`):
  - `type Deps struct{News NewsSource; WhatsOn WhatsOnSource; Sponsors SponsorSource; Affiliations AffiliationSource; Teams TeamSource; Players SquadSource; Settings SettingSource; Users UserSource; Visitors CountSource; Files *upload.Files; Version string}` (every field an interface, declared in `service.go`)
  - `NewService(Deps) *Service`, with methods `Site(ctx) (Info, error)`, `Home(ctx) Home`, `Contact(ctx) (Contact, error)`, `Team(ctx, id) (TeamDetail, error)`
  - `NewHandlers`, `Register`
  - Types: `Info`, `Home`, `Contact`, `ContactPerson`, `TeamDetail`, `Manager`

- [ ] **Step 1: Failing tests**

`server/internal/site/service_test.go`:
```go
package site_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/site"
	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

var ctx = context.Background()

type fakes struct {
	newsErr      error
	sponsorsErr  error
	squadForTeam map[int][]player.Member
}

func (f *fakes) Latest(context.Context) (news.Article, bool, error) {
	return news.Article{ID: 1, Title: "Opener"}, true, f.newsErr
}
func (f *fakes) Next(context.Context) (whatson.Event, bool, error) { return whatson.Event{}, false, nil }
func (f *fakes) ListMinimal(context.Context) ([]sponsor.Public, error) {
	return []sponsor.Public{{ID: 1, Name: "Club Sponsor"}}, f.sponsorsErr
}
func (f *fakes) ForTeam(_ context.Context, teamID int) ([]sponsor.Public, error) {
	return []sponsor.Public{{ID: 2, Name: "Team Sponsor", Team: "1"}}, nil
}

type affiliations struct{}

func (affiliations) ListMinimal(context.Context) ([]affiliation.Public, error) {
	return []affiliation.Public{{ID: 1, Name: "County FA"}}, nil
}

type teams struct{}

func (teams) List(_ context.Context, includeInactive bool) ([]team.Public, error) {
	return []team.Public{{ID: 1, Name: "First Team", IsActive: true}}, nil
}

func (teams) Get(_ context.Context, id int) (team.Public, error) {
	switch id {
	case 1:
		return team.Public{ID: 1, Name: "First Team", IsActive: true}, nil
	case 2:
		return team.Public{ID: 2, Name: "Under 12s", IsYouth: true}, nil
	}
	return team.Public{}, svcerr.NotFound("team not found", nil)
}

func (f *fakes) Squad(_ context.Context, t team.Team) ([]player.Member, error) {
	return f.squadForTeam[t.ID], nil
}

type settings struct{}

func (settings) DisplayEmail(context.Context) (string, error) { return "hello@example.test", nil }

type users struct{}

func (users) GetUsersContact(context.Context) ([]user.User, error) {
	return []user.User{{ID: 3, Name: "Club Secretary", Email: "secretary@example.test", Role: role.ClubSecretary,
		FileName: null.StringFrom("user/sec.png"), Hash: null.StringFrom("secret-hash")}}, nil
}

func (users) GetUsersManagersTeam(_ context.Context, t team.Team) ([]user.User, error) {
	if t.ID == 1 {
		return []user.User{{ID: 2, Name: "Team Manager", Email: "manager@example.test", Role: role.Manager}}, nil
	}
	return nil, nil
}

type visitors struct{}

func (visitors) Count() int { return 42 }

func newService(f *fakes) *site.Service {
	return site.NewService(site.Deps{
		News: f, WhatsOn: f, Sponsors: f, Affiliations: affiliations{}, Teams: teams{}, Players: f,
		Settings: settings{}, Users: users{}, Visitors: visitors{},
		Files: upload.New(uploadtest.New()), Version: "test",
	})
}

func TestSite(t *testing.T) {
	info, err := newService(&fakes{}).Site(ctx)
	require.NoError(t, err)
	assert.Equal(t, 42, info.VisitorCount)
	assert.Equal(t, "hello@example.test", info.DisplayEmail)
	assert.Len(t, info.Teams, 1)
	assert.Positive(t, info.Year)
}

func TestHomeIsFailSoft(t *testing.T) {
	home := newService(&fakes{sponsorsErr: errors.New("db down")}).Home(ctx)
	require.NotNil(t, home.LatestNews)
	assert.Equal(t, "Opener", home.LatestNews.Title)
	assert.Nil(t, home.NextEvent)
	assert.Empty(t, home.Sponsors, "failed panel is omitted")
	assert.Len(t, home.Affiliations, 1)

	home = newService(&fakes{newsErr: errors.New("db down")}).Home(ctx)
	assert.Nil(t, home.LatestNews)
}

func TestContactNeverLeaksSecrets(t *testing.T) {
	c, err := newService(&fakes{}).Contact(ctx)
	require.NoError(t, err)
	require.Len(t, c.People, 1)
	assert.Equal(t, "Club Secretary", c.People[0].Role)
	assert.Equal(t, "https://cdn.test/user/sec.png", c.People[0].ImageURL)
}

// TestTeamDetailHidesMinorPhotos pins Review Focus #1 for /teams/{id}: the
// squad is built by player.Service.Squad, which applies PhotoVisible with the
// team's youth flag. This test proves site passes that flag through.
func TestTeamDetailHidesMinorPhotos(t *testing.T) {
	f := &fakes{squadForTeam: map[int][]player.Member{2: {{ID: 9, Name: "Youth"}}}}
	var seen []team.Team
	svc := site.NewService(site.Deps{
		News: f, WhatsOn: f, Sponsors: f, Affiliations: affiliations{}, Teams: teams{},
		Players: squadSpy{f: f, seen: &seen}, Settings: settings{}, Users: users{}, Visitors: visitors{},
		Files: upload.New(uploadtest.New()),
	})
	detail, err := svc.Team(ctx, 2)
	require.NoError(t, err)
	require.Len(t, seen, 1)
	assert.True(t, seen[0].IsYouth, "youth flag must reach player.Squad")
	assert.Len(t, detail.Players, 1)
}

type squadSpy struct {
	f    *fakes
	seen *[]team.Team
}

func (s squadSpy) Squad(ctx context.Context, t team.Team) ([]player.Member, error) {
	*s.seen = append(*s.seen, t)
	return s.f.Squad(ctx, t)
}

func TestTeamDetail(t *testing.T) {
	detail, err := newService(&fakes{}).Team(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "First Team", detail.Team.Name)
	require.Len(t, detail.Managers, 1)
	assert.Equal(t, "manager@example.test", detail.Managers[0].Email)
	assert.Len(t, detail.Sponsors, 1)

	_, err = newService(&fakes{}).Team(ctx, 99)
	se, ok := svcerr.As(err)
	require.True(t, ok)
	assert.Equal(t, svcerr.KindNotFound, se.Kind)
}

func TestSiteRoutes(t *testing.T) {
	e := apitest.NewEcho()
	site.NewHandlers(newService(&fakes{})).Register(web.NewAPI(e, false), authtest.Everyone().Guards())
	c := apitest.New(e)
	for _, path := range []string{"/api/v1/site", "/api/v1/home", "/api/v1/contact", "/api/v1/teams/1"} {
		rec := c.Get(t, path)
		assert.Equal(t, http.StatusOK, rec.Code, path)
		assert.NotContains(t, rec.Body.String(), "secret-hash", path)
	}
	assert.Equal(t, http.StatusNotFound, c.Get(t, "/api/v1/teams/99").Code)
}
```
Run: `go test ./server/internal/site/`. Expected: compile errors (package missing).

- [ ] **Step 2: Implement**

`server/internal/site/types.go`:
```go
// Package site serves read-only views that combine several domains: shared
// layout data, the home page, the contact page and the team page.
package site

import (
	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

// Info is the data every page's layout needs.
type Info struct {
	Year         int           `json:"year"`
	VisitorCount int           `json:"visitorCount"`
	DisplayEmail string        `json:"displayEmail,omitempty"`
	Version      string        `json:"version"`
	Teams        []team.Public `json:"teams"`
}

// Home is the home page. Panels whose data failed to load are omitted.
type Home struct {
	LatestNews   *news.Article        `json:"latestNews,omitempty"`
	NextEvent    *whatson.Event       `json:"nextEvent,omitempty"`
	Sponsors     []sponsor.Public     `json:"sponsors"`
	Affiliations []affiliation.Public `json:"affiliations"`
}

// Contact is the contact page.
type Contact struct {
	DisplayEmail string          `json:"displayEmail,omitempty"`
	People       []ContactPerson `json:"people"`
}

// ContactPerson is a club official listed on the contact page.
type ContactPerson struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	ImageURL string `json:"imageUrl,omitempty"`
}

// TeamDetail is the public team page.
type TeamDetail struct {
	Team     team.Public      `json:"team"`
	Managers []Manager        `json:"managers"`
	Sponsors []sponsor.Public `json:"sponsors"`
	Players  []player.Member  `json:"players"`
}

// Manager is a team manager's public contact.
type Manager struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
```

`server/internal/site/service.go`:
```go
package site

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"

	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

var tracer = otel.Tracer("github.com/COMTOP1/AFC-GO/server/internal/site")

type (
	NewsSource interface {
		Latest(ctx context.Context) (news.Article, bool, error)
	}
	WhatsOnSource interface {
		Next(ctx context.Context) (whatson.Event, bool, error)
	}
	SponsorSource interface {
		ListMinimal(ctx context.Context) ([]sponsor.Public, error)
		ForTeam(ctx context.Context, teamID int) ([]sponsor.Public, error)
	}
	AffiliationSource interface {
		ListMinimal(ctx context.Context) ([]affiliation.Public, error)
	}
	TeamSource interface {
		List(ctx context.Context, includeInactive bool) ([]team.Public, error)
		Get(ctx context.Context, id int) (team.Public, error)
	}
	SquadSource interface {
		Squad(ctx context.Context, t team.Team) ([]player.Member, error)
	}
	SettingSource interface {
		DisplayEmail(ctx context.Context) (string, error)
	}
	UserSource interface {
		GetUsersContact(ctx context.Context) ([]user.User, error)
		GetUsersManagersTeam(ctx context.Context, teamParam team.Team) ([]user.User, error)
	}
	CountSource interface {
		Count() int
	}
)

// Deps are the services site reads from.
type Deps struct {
	News         NewsSource
	WhatsOn      WhatsOnSource
	Sponsors     SponsorSource
	Affiliations AffiliationSource
	Teams        TeamSource
	Players      SquadSource
	Settings     SettingSource
	Users        UserSource
	Visitors     CountSource
	Files        *upload.Files
	Version      string
}

// Service builds the cross-domain page data.
type Service struct {
	d Deps
}

func NewService(d Deps) *Service {
	return &Service{d: d}
}

// Site returns the layout data shared by every page.
func (s *Service) Site(ctx context.Context) (Info, error) {
	ctx, span := tracer.Start(ctx, "site.Service.Site")
	defer span.End()
	teams, err := s.d.Teams.List(ctx, false)
	if err != nil {
		return Info{}, err
	}
	email, err := s.d.Settings.DisplayEmail(ctx)
	if err != nil {
		return Info{}, err
	}
	return Info{
		Year:         time.Now().Year(),
		VisitorCount: s.d.Visitors.Count(),
		DisplayEmail: email,
		Version:      s.d.Version,
		Teams:        teams,
	}, nil
}

// Home returns the home page; failed panels are logged and omitted.
func (s *Service) Home(ctx context.Context) Home {
	ctx, span := tracer.Start(ctx, "site.Service.Home")
	defer span.End()
	home := Home{Sponsors: []sponsor.Public{}, Affiliations: []affiliation.Public{}}
	soft := func(what string, err error) bool {
		if err != nil {
			span.RecordError(err)
			slog.InfoContext(ctx, fmt.Sprintf("home: failed to load %s: %+v", what, err))
			return false
		}
		return true
	}
	if a, ok, err := s.d.News.Latest(ctx); soft("latest news", err) && ok {
		home.LatestNews = &a
	}
	if e, ok, err := s.d.WhatsOn.Next(ctx); soft("next event", err) && ok {
		home.NextEvent = &e
	}
	if sponsors, err := s.d.Sponsors.ListMinimal(ctx); soft("sponsors", err) {
		home.Sponsors = sponsors
	}
	if affiliations, err := s.d.Affiliations.ListMinimal(ctx); soft("affiliations", err) {
		home.Affiliations = affiliations
	}
	return home
}

// Contact returns the club officials and public contact email.
func (s *Service) Contact(ctx context.Context) (Contact, error) {
	ctx, span := tracer.Start(ctx, "site.Service.Contact")
	defer span.End()
	people, err := s.d.Users.GetUsersContact(ctx)
	if err != nil {
		return Contact{}, fmt.Errorf("failed to get contacts: %w", err)
	}
	email, err := s.d.Settings.DisplayEmail(ctx)
	if err != nil {
		return Contact{}, err
	}
	out := Contact{DisplayEmail: email, People: make([]ContactPerson, 0, len(people))}
	for _, u := range people {
		out.People = append(out.People, ContactPerson{
			ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role.String(), ImageURL: s.d.Files.URL(u.FileName.String),
		})
	}
	return out, nil
}

// Team returns the public team page.
func (s *Service) Team(ctx context.Context, id int) (TeamDetail, error) {
	ctx, span := tracer.Start(ctx, "site.Service.Team")
	defer span.End()
	t, err := s.d.Teams.Get(ctx, id)
	if err != nil {
		return TeamDetail{}, err
	}
	ref := team.Team{ID: t.ID, IsYouth: t.IsYouth}
	managers, err := s.d.Users.GetUsersManagersTeam(ctx, ref)
	if err != nil {
		return TeamDetail{}, fmt.Errorf("failed to get managers: %w", err)
	}
	sponsors, err := s.d.Sponsors.ForTeam(ctx, t.ID)
	if err != nil {
		return TeamDetail{}, err
	}
	players, err := s.d.Players.Squad(ctx, ref)
	if err != nil {
		return TeamDetail{}, err
	}
	out := TeamDetail{Team: t, Managers: make([]Manager, 0, len(managers)), Sponsors: sponsors, Players: players}
	for _, m := range managers {
		out.Managers = append(out.Managers, Manager{Name: m.Name, Email: m.Email})
	}
	return out, nil
}
```

`server/internal/site/handlers.go`:
```go
package site

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /site, /home, /contact and /teams/{id}.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the site routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, _ web.Guards) {
	g.GET("/site", h.site)
	g.GET("/home", h.home)
	g.GET("/contact", h.contact)
	g.GET("/teams/:id", h.team)
}

// site returns layout data shared by every page.
//
//	@Summary	Site layout data
//	@Tags		site
//	@Produce	json
//	@Success	200	{object}	site.Info
//	@Router		/site [get]
func (h *Handlers) site(c echo.Context) error {
	out, err := h.svc.Site(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// home returns the home page.
//
//	@Summary	Home page
//	@Tags		site
//	@Produce	json
//	@Success	200	{object}	site.Home
//	@Router		/home [get]
func (h *Handlers) home(c echo.Context) error {
	return c.JSON(http.StatusOK, h.svc.Home(c.Request().Context()))
}

// contact returns the contact page.
//
//	@Summary	Contact page
//	@Tags		site
//	@Produce	json
//	@Success	200	{object}	site.Contact
//	@Router		/contact [get]
func (h *Handlers) contact(c echo.Context) error {
	out, err := h.svc.Contact(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// team returns a team's public page: details, managers, sponsors and squad.
//
//	@Summary	Team page
//	@Tags		teams
//	@Produce	json
//	@Param		id	path		int	true	"Team ID"
//	@Success	200	{object}	site.TeamDetail
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/teams/{id} [get]
func (h *Handlers) team(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	out, err := h.svc.Team(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}
```
Run: `go test ./server/internal/site/`. Expected: PASS.

- [ ] **Step 3: Wire into `app.Build`, regenerate, run, commit**

After the other services in `Build`, add:
```go
	siteSvc := site.NewService(site.Deps{
		News: newsSvc, WhatsOn: whatsOnSvc, Sponsors: sponsorSvc, Affiliations: affiliationSvc,
		Teams: teamSvc, Players: playerSvc, Settings: settingSvc, Users: s.User, Visitors: counter,
		Files: uploads, Version: conf.Version,
	})
```
`counter` is declared above the services, so move `counter := visitors.New(…)` before this block if it isn't already. After the other handlers add `site.NewHandlers(siteSvc).Register(api, guards)`.
```bash
go generate ./server/internal/docs
gofmt -l server; go build ./... && go vet ./... && go test ./... && go tool golangci-lint run ./...
git add -A server
git commit -m "Add site aggregates: /site, /home, /contact, /teams/{id}

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---
### Task 19: `files`: `GET /files/{kind}/{id}` (and the legacy `/download` alias)

This replaces `views/download.go`. It looks up the object key for `(kind, id)`, checks the object exists and 302-redirects to its CDN URL with the same `Cache-Control` header as today. Player photos go through `player.Service.PhotoKey`, so youth-team and under-18 photos are a **404** (legacy returned an empty 200). Legacy `/download?s=<code>&id=<n>` is rewritten to call the same service.

**Files:**
- Create: `server/internal/files/{service,handlers,service_test}.go`
- Modify: `server/internal/legacy/views/download.go` (becomes a thin alias), `views.go`, `server/internal/app/app.go`

**Interfaces:**
- Consumes: `player.Service.PhotoKey` (Task 16), every domain store's single-row getter (existing), `upload.Files.Exists/URL` (Task 3).
- Produces (`files`):
  - `type Resolver func(ctx, id int) (key string, err error)`
  - `type Sources struct{Affiliation, Document, Image, News, Programme, Sponsor, Team, User, WhatsOn <store interfaces>; Players PhotoKeyer}`
  - `Resolvers(Sources) map[string]Resolver`
  - `NewService(*upload.Files, map[string]Resolver) *Service`, with methods `URL(ctx, kind string, id int) (string, error)` and `LegacyKind(code string) (string, bool)`
  - `NewHandlers`, `Register`
  - Kinds: `affiliation, document, gallery, news, player, programme, sponsor, team, user, whatson`
- Produces (legacy): `Deps.FileService`.

- [ ] **Step 1: Failing tests**

`server/internal/files/service_test.go`:
```go
package files_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/files"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

var ctx = context.Background()

func newService() *files.Service {
	objects := uploadtest.New()
	objects.Objects["document/rules.pdf"] = "PDF"
	objects.Objects["player/adult.png"] = "IMG"
	return files.NewService(upload.New(objects), map[string]files.Resolver{
		"document": func(_ context.Context, id int) (string, error) {
			switch id {
			case 1:
				return "document/rules.pdf", nil
			case 2:
				return "document/missing.pdf", nil // row exists, object doesn't
			case 3:
				return "", nil // row exists with no file
			}
			return "", fmt.Errorf("failed to get document: %w", sql.ErrNoRows)
		},
		// Stands in for player.Service.PhotoKey, which returns NotFound for
		// youth-team and under-18 players (tested in the player package).
		"player": func(_ context.Context, id int) (string, error) {
			if id == 1 {
				return "player/adult.png", nil
			}
			return "", svcerr.NotFound("player photo not found", nil)
		},
	})
}

func kindOf(t *testing.T, err error) svcerr.Kind {
	t.Helper()
	se, ok := svcerr.As(err)
	require.True(t, ok, "want svcerr, got %v", err)
	return se.Kind
}

func TestURL(t *testing.T) {
	svc := newService()
	u, err := svc.URL(ctx, "document", 1)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.test/document/rules.pdf", u)

	for name, tc := range map[string]struct {
		kind string
		id   int
	}{
		"unknown kind":      {"passwords", 1},
		"unknown row":       {"document", 99},
		"object missing":    {"document", 2},
		"row without file":  {"document", 3},
	} {
		_, err = svc.URL(ctx, tc.kind, tc.id)
		assert.Equal(t, svcerr.KindNotFound, kindOf(t, err), name)
	}
}

// TestPlayerFileHiddenForMinor pins Review Focus #1 for /files/player.
func TestPlayerFileHiddenForMinor(t *testing.T) {
	e := apitest.NewEcho()
	files.NewHandlers(newService()).Register(web.NewAPI(e, false), authtest.Everyone().Guards())
	c := apitest.New(e)

	rec := c.Get(t, "/api/v1/files/player/1")
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "https://cdn.test/player/adult.png", rec.Header().Get("Location"))
	assert.Equal(t, "public, max-age=31536000, immutable", rec.Header().Get("Cache-Control"))

	rec = c.Get(t, "/api/v1/files/player/2")
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Empty(t, rec.Header().Get("Location"))
}

func TestLegacyKind(t *testing.T) {
	for code, kind := range map[string]string{
		"a": "affiliation", "d": "document", "g": "gallery", "l": "player", "n": "news",
		"p": "programme", "s": "sponsor", "t": "team", "u": "user", "w": "whatson",
	} {
		got, ok := files.LegacyKind(code)
		assert.True(t, ok, code)
		assert.Equal(t, kind, got, code)
	}
	_, ok := files.LegacyKind("x")
	assert.False(t, ok)
}
```
Run: compile errors expected.

- [ ] **Step 2: Implement**

`server/internal/files/service.go`:
```go
// Package files redirects to stored files by resource kind and ID, applying
// per-kind access rules (player photos respect the safeguarding rule).
package files

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"

	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
	"github.com/COMTOP1/AFC-GO/server/internal/document"
	"github.com/COMTOP1/AFC-GO/server/internal/image"
	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/programme"
	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

var tracer = otel.Tracer("github.com/COMTOP1/AFC-GO/server/internal/files")

// Resolver returns the storage key for a resource ID ("" when it has no file).
type Resolver func(ctx context.Context, id int) (string, error)

// PhotoKeyer resolves player photos, enforcing the safeguarding rule
// (satisfied by *player.Service).
type PhotoKeyer interface {
	PhotoKey(ctx context.Context, id int) (string, error)
}

// Sources are the lookups behind each kind (satisfied by the domain stores).
type Sources struct {
	Affiliation interface {
		GetAffiliation(ctx context.Context, a affiliation.Affiliation) (affiliation.Affiliation, error)
	}
	Document interface {
		GetDocument(ctx context.Context, d document.Document) (document.Document, error)
	}
	Image interface {
		GetImage(ctx context.Context, i image.Image) (image.Image, error)
	}
	News interface {
		GetNewsArticle(ctx context.Context, n news.News) (news.News, error)
	}
	Programme interface {
		GetProgramme(ctx context.Context, p programme.Programme) (programme.Programme, error)
	}
	Sponsor interface {
		GetSponsor(ctx context.Context, s sponsor.Sponsor) (sponsor.Sponsor, error)
	}
	Team interface {
		GetTeam(ctx context.Context, t team.Team) (team.Team, error)
	}
	User interface {
		GetUser(ctx context.Context, u user.User) (user.User, error)
	}
	WhatsOn interface {
		GetWhatsOnArticle(ctx context.Context, w whatson.WhatsOn) (whatson.WhatsOn, error)
	}
	Players PhotoKeyer
}

// Resolvers builds the kind → lookup table.
func Resolvers(src Sources) map[string]Resolver {
	return map[string]Resolver{
		"affiliation": func(ctx context.Context, id int) (string, error) {
			a, err := src.Affiliation.GetAffiliation(ctx, affiliation.Affiliation{ID: id})
			return a.FileName.String, svcerr.FromStore(err, "affiliation")
		},
		"document": func(ctx context.Context, id int) (string, error) {
			d, err := src.Document.GetDocument(ctx, document.Document{ID: id})
			return d.FileName, svcerr.FromStore(err, "document")
		},
		"gallery": func(ctx context.Context, id int) (string, error) {
			i, err := src.Image.GetImage(ctx, image.Image{ID: id})
			return i.FileName, svcerr.FromStore(err, "gallery image")
		},
		"news": func(ctx context.Context, id int) (string, error) {
			n, err := src.News.GetNewsArticle(ctx, news.News{ID: id})
			return n.FileName.String, svcerr.FromStore(err, "news article")
		},
		"player": src.Players.PhotoKey,
		"programme": func(ctx context.Context, id int) (string, error) {
			p, err := src.Programme.GetProgramme(ctx, programme.Programme{ID: id})
			return p.FileName, svcerr.FromStore(err, "programme")
		},
		"sponsor": func(ctx context.Context, id int) (string, error) {
			s, err := src.Sponsor.GetSponsor(ctx, sponsor.Sponsor{ID: id})
			return s.FileName.String, svcerr.FromStore(err, "sponsor")
		},
		"team": func(ctx context.Context, id int) (string, error) {
			t, err := src.Team.GetTeam(ctx, team.Team{ID: id})
			return t.FileName.String, svcerr.FromStore(err, "team")
		},
		"user": func(ctx context.Context, id int) (string, error) {
			u, err := src.User.GetUser(ctx, user.User{ID: id})
			if err == nil && u.ID != id {
				// GetUser matches email OR id; never hand out someone else's photo.
				return "", svcerr.NotFound("user not found", nil)
			}
			return u.FileName.String, svcerr.FromStore(err, "user")
		},
		"whatson": func(ctx context.Context, id int) (string, error) {
			w, err := src.WhatsOn.GetWhatsOnArticle(ctx, whatson.WhatsOn{ID: id})
			return w.FileName.String, svcerr.FromStore(err, "whats on event")
		},
	}
}

var legacyKinds = map[string]string{
	"a": "affiliation", "d": "document", "g": "gallery", "l": "player", "n": "news",
	"p": "programme", "s": "sponsor", "t": "team", "u": "user", "w": "whatson",
}

// LegacyKind maps the old /download?s=<code> codes to kinds.
func LegacyKind(code string) (string, bool) {
	kind, ok := legacyKinds[code]
	return kind, ok
}

// Service resolves files to public URLs.
type Service struct {
	files     *upload.Files
	resolvers map[string]Resolver
}

func NewService(files *upload.Files, resolvers map[string]Resolver) *Service {
	return &Service{files: files, resolvers: resolvers}
}

// URL returns the CDN URL for the file behind (kind, id), or NotFound.
func (s *Service) URL(ctx context.Context, kind string, id int) (string, error) {
	ctx, span := tracer.Start(ctx, "files.Service.URL")
	defer span.End()
	resolve, ok := s.resolvers[kind]
	if !ok {
		return "", svcerr.NotFound("unknown file kind", nil)
	}
	key, err := resolve(ctx, id)
	if err != nil {
		return "", err
	}
	if key == "" {
		return "", svcerr.NotFound(kind+" has no file", nil)
	}
	exists, err := s.files.Exists(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to check %s file: %w", kind, err)
	}
	if !exists {
		return "", svcerr.NotFound(kind+" file not found", nil)
	}
	return s.files.URL(key), nil
}
```

`server/internal/files/handlers.go`:
```go
package files

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /files.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the file route on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, _ web.Guards) {
	g.GET("/files/:kind/:id", h.get)
}

// get redirects to a stored file.
//
//	@Summary	Download a file
//	@Tags		files
//	@Param		kind	path	string	true	"affiliation, document, gallery, news, player, programme, sponsor, team, user or whatson"
//	@Param		id		path	int		true	"Resource ID"
//	@Success	302
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/files/{kind}/{id} [get]
func (h *Handlers) get(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	u, err := h.svc.URL(c.Request().Context(), c.Param("kind"), id)
	if err != nil {
		return err
	}
	c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	return c.Redirect(http.StatusFound, u)
}
```
Run: `go test ./server/internal/files/`. Expected: PASS.

- [ ] **Step 3: Make legacy `/download` an alias**

Add `fileSvc *files.Service` / `FileService *files.Service` to `Views`/`Deps`/`New`. Replace the **entire** contents of `server/internal/legacy/views/download.go` with:
```go
package views

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/files"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
)

// DownloadFunc keeps /download?s=<code>&id=<n> working for old links and
// templates; it delegates to the same service as GET /api/v1/files.
func (v *Views) DownloadFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.DownloadFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))

	id, err := strconv.Atoi(c.QueryParam("id"))
	if err != nil || id < 1 {
		return v.error(http.StatusBadRequest, "id must be a positive integer",
			fmt.Errorf("download: invalid id %q", c.QueryParam("id")))
	}
	kind, ok := files.LegacyKind(c.QueryParam("s"))
	if !ok {
		return v.error(http.StatusBadRequest, "unknown download source",
			fmt.Errorf("download: unknown source %q", c.QueryParam("s")))
	}
	u, err := v.fileSvc.URL(c.Request().Context(), kind, id)
	if err != nil {
		if se, isSvc := svcerr.As(err); isSvc && se.Kind == svcerr.KindNotFound {
			return c.String(http.StatusNotFound, se.Message)
		}
		return errors.Join(errors.New("download failed"), err)
	}
	c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	return c.Redirect(http.StatusFound, u)
}
```

- [ ] **Step 4: Wire, regenerate, run, commit**

In `app.Build`, after `playerSvc`:
```go
	fileSvc := files.NewService(uploads, files.Resolvers(files.Sources{
		Affiliation: s.Affiliation, Document: s.Document, Image: s.Image, News: s.News,
		Programme: s.Programme, Sponsor: s.Sponsor, Team: s.Team, User: s.User, WhatsOn: s.WhatsOn,
		Players: playerSvc,
	}))
```
Add `FileService: fileSvc,` to `views.Deps` and `files.NewHandlers(fileSvc).Register(api, guards)` after the other handlers.
```bash
go generate ./server/internal/docs
goimports -w server/internal/legacy/views
gofmt -l server; go build ./... && go vet ./... && go test ./... && go tool golangci-lint run ./...
git add -A server
git commit -m "Add /api/v1/files/{kind}/{id}; legacy /download delegates to it

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---
### Task 20: Users (and the `emails` package)

User administration: list, create (generated password plus signup email), edit, delete and admin-triggered password reset. All of it is 🛡 `ClubSecretaryHigher`. Emails move out of the legacy `templates` package, which gets deleted in sub-project 5, into `emails`.

When no mail server is configured, or sending fails, the responses carry what the admin needs to pass on by hand, exactly as legacy flash messages do today:
- create: `tempPassword`
- reset: `resetUrl`

**Files:**
- Move: `server/internal/legacy/templates/signupEmail.tmpl`, `resetEmail.tmpl` and `server/internal/legacy/templates/mjml/` → `server/internal/emails/`
- Create: `server/internal/emails/emails.go`, `emails_test.go`
- Modify: `server/internal/legacy/templates/template.go` (drop email templates), `server/internal/user/user.go` (`EditUser` can clear phone), `server/internal/user/store.go` (`addUser` returns ID)
- Create: `server/internal/user/{store_test,types,service,handlers,fake_test,service_test,handlers_test}.go`
- Modify: `server/internal/legacy/views/user.go`, `server/internal/legacy/views/reset.go` (`ResetUserPasswordFunc` only), `views.go`, `server/internal/app/app.go`, `server/internal/app/helpers_test.go`

**Interfaces:**
- Produces (`emails`):
  - `var ErrNoMailer`
  - `type Sender interface{Send(ctx, mail.Mail) error}`
  - `NewSMTP(*mail.MailerInit) *SMTP`, which implements `Sender`
  - `Signup(to, name, password, domain string) (mail.Mail, error)` and `Reset(to, resetURL string) (mail.Mail, error)`
- Produces (`user`):
  - Types:
    - `Admin{ID int; Name, Email, Phone, Role, RoleCode string; TeamID int; ImageURL string}`
    - `CreateInput{Name, Email, Phone, Role string; TeamID int}`
    - `UpdateInput{Name, Email, Phone, Role *string; TeamID *int; RemoveImage bool}`
    - `Created{User Admin; EmailSent bool; TempPassword string}`
    - `ResetResult{EmailSent bool; ResetURL string}`
  - `type HashParams struct{WorkFactor, BlockSize, Parallelism, KeyLength int}`
  - `type TokenSetter interface{Set(ctx, token string, userID int, ttl time.Duration) error}`
  - `NewService(store, TeamGetter, *upload.Files, emails.Sender, TokenSetter, HashParams, domain string) *Service`, with methods `List`, `Get`, `Create`, `Update`, `Delete`, `ResetPassword(ctx, id) (ResetResult, error)`
  - `NewHandlers`, `Register`
- Produces (legacy): `Deps.UserService`.

- [ ] **Step 1: Move the email templates and write `emails`**

```bash
mkdir -p server/internal/emails
git mv server/internal/legacy/templates/signupEmail.tmpl server/internal/legacy/templates/resetEmail.tmpl server/internal/emails/
git mv server/internal/legacy/templates/mjml server/internal/emails/mjml
```
In `server/internal/legacy/templates/template.go`, delete the `SignupEmailTemplate` and `ResetEmailTemplate` constants, `GetEmailTemplate`, and the two `AllTemplates` rows for `signupEmail.tmpl` and `resetEmail.tmpl`.

`server/internal/emails/emails_test.go`:
```go
package emails_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/emails"
)

func TestSignupRendersCredentials(t *testing.T) {
	m, err := emails.Signup("new@example.test", "New Person", "Tmp-Pa55!", "afc.example.test")
	require.NoError(t, err)
	assert.Equal(t, "new@example.test", m.To)
	var body bytes.Buffer
	require.NoError(t, m.Tpl.Execute(&body, m.TplData))
	for _, want := range []string{"New Person", "new@example.test", "Tmp-Pa55!", "https://afc.example.test"} {
		assert.Contains(t, body.String(), want)
	}
}

func TestResetRendersLink(t *testing.T) {
	m, err := emails.Reset("user@example.test", "https://afc.example.test/reset/abc")
	require.NoError(t, err)
	var body bytes.Buffer
	require.NoError(t, m.Tpl.Execute(&body, m.TplData))
	assert.Contains(t, body.String(), "https://afc.example.test/reset/abc")
}
```

`server/internal/emails/emails.go`:
```go
// Package emails renders and sends the site's transactional emails.
package emails

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"

	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
)

//go:embed signupEmail.tmpl resetEmail.tmpl
var tmpls embed.FS

// ErrNoMailer means no SMTP server could be reached; callers fall back to
// showing the admin what would have been emailed.
var ErrNoMailer = errors.New("no mailer available")

// Sender sends one email.
type Sender interface {
	Send(ctx context.Context, m mail.Mail) error
}

// SMTP sends via the configured SMTP server, connecting per message.
type SMTP struct {
	init *mail.MailerInit
}

func NewSMTP(init *mail.MailerInit) *SMTP {
	return &SMTP{init: init}
}

func (s *SMTP) Send(ctx context.Context, m mail.Mail) error {
	mailer := s.init.ConnectMailer(ctx)
	if mailer == nil {
		return ErrNoMailer
	}
	defer func() { _ = mailer.Close() }()
	return mailer.SendMail(ctx, m)
}

func parse(name string) (*template.Template, error) {
	t, err := template.New(name).ParseFS(tmpls, name)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", name, err)
	}
	return t, nil
}

// Signup is the welcome email carrying a new user's temporary password.
func Signup(to, name, password, domain string) (mail.Mail, error) {
	tpl, err := parse("signupEmail.tmpl")
	if err != nil {
		return mail.Mail{}, err
	}
	return mail.Mail{
		Subject: "Welcome to AFC Aldermaston!",
		Tpl:     tpl,
		To:      to,
		From:    "Aldermaston AFC No-Reply <no-reply.afc@bswdi.co.uk>",
		TplData: struct {
			Name, Email, Password, Domain string
		}{Name: name, Email: to, Password: password, Domain: domain},
	}, nil
}

// Reset is the password reset email.
func Reset(to, resetURL string) (mail.Mail, error) {
	tpl, err := parse("resetEmail.tmpl")
	if err != nil {
		return mail.Mail{}, err
	}
	return mail.Mail{
		Subject: "AFC Security - Reset Password",
		Tpl:     tpl,
		To:      to,
		From:    "AFC Security <no-reply.afc@bswdi.co.uk>",
		TplData: struct {
			Email, URL string
		}{Email: to, URL: resetURL},
	}, nil
}
```
Run: `go test ./server/internal/emails/`. Expected: PASS. (The templates use `html/template` escaping. `https://` URLs pass through unescaped. If an assertion fails on escaping, compare with what legacy `GetEmailTemplate` produced, which also used `html/template`.)

- [ ] **Step 2: Store tests, then fix `addUser` and `EditUser`**

`server/internal/user/store_test.go`:
```go
package user_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

func TestAddUserReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	added, err := user.NewUserRepo(db).AddUser(context.Background(),
		user.User{Name: "New", Email: "new@example.test", Role: role.Treasurer, ResetPassword: true})
	require.NoError(t, err)
	require.Positive(t, added.ID)
}

func TestEditUserClearsPhone(t *testing.T) {
	db, _ := testdb.Open(t)
	store := user.NewUserRepo(db)
	ctx := context.Background()
	u, err := store.GetUser(ctx, user.User{ID: 1})
	require.NoError(t, err)
	require.True(t, u.Phone.Valid)

	u.Phone = null.String{}
	_, err = store.EditUser(ctx, u)
	require.NoError(t, err)
	got, err := store.GetUser(ctx, user.User{ID: 1})
	require.NoError(t, err)
	assert.False(t, got.Phone.Valid)
}
```
Run (FAIL). Then:
- In `store.go` `addUser`: add `.Suffix("RETURNING id")` and replace the `ExecContext`/`RowsAffected` block with `err = s.db.GetContext(ctx, &userParam.ID, sql, args...)`, keeping the existing error wrapping and returning `userParam, nil`.
- In `user.go` `EditUser`, replace the phone block:
```go
	userDB.Phone = userParam.Phone
```
  Every caller passes a user loaded from the store (the only one left after Task 15 is `ResetPassword`), so the phone is only cleared when a caller clears it on purpose.

Re-run: PASS.

- [ ] **Step 3: Fakes and failing service tests**

`server/internal/user/fake_test.go`:
```go
package user_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/lib/pq"

	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

type fakeStore struct {
	mu     sync.Mutex
	rows   map[int]user.User
	nextID int
}

func newFakeStore(rows ...user.User) *fakeStore {
	f := &fakeStore{rows: map[int]user.User{}, nextID: 100}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) GetUsers(context.Context) ([]user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]user.User, 0, len(f.rows))
	for _, r := range f.rows {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (f *fakeStore) GetUser(_ context.Context, u user.User) (user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r, ok := f.rows[u.ID]; ok {
		return r, nil
	}
	return user.User{}, fmt.Errorf("failed to get user: %w", sql.ErrNoRows)
}

func (f *fakeStore) AddUser(_ context.Context, u user.User) (user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, r := range f.rows {
		if r.Email == u.Email {
			return user.User{}, fmt.Errorf("failed to add user: %w", &pq.Error{Code: "23505"})
		}
	}
	f.nextID++
	u.ID = f.nextID
	f.rows[u.ID] = u
	return u, nil
}

func (f *fakeStore) EditUser(_ context.Context, u user.User) (user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.rows[u.ID]; !ok {
		return user.User{}, errors.New("no such row")
	}
	f.rows[u.ID] = u
	return u, nil
}

func (f *fakeStore) DeleteUser(_ context.Context, u user.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rows, u.ID)
	return nil
}

func (f *fakeStore) row(id int) user.User {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.rows[id]
}

type fakeTeams map[int]team.Team

func (f fakeTeams) GetTeam(_ context.Context, t team.Team) (team.Team, error) {
	if x, ok := f[t.ID]; ok {
		return x, nil
	}
	return team.Team{}, fmt.Errorf("failed to get team: %w", sql.ErrNoRows)
}

type fakeSender struct {
	err  error
	sent []mail.Mail
}

func (f *fakeSender) Send(_ context.Context, m mail.Mail) error {
	if f.err != nil {
		return f.err
	}
	f.sent = append(f.sent, m)
	return nil
}

type fakeTokens struct {
	tokens map[string]int
	ttl    time.Duration
}

func (f *fakeTokens) Set(_ context.Context, token string, userID int, ttl time.Duration) error {
	f.tokens[token] = userID
	f.ttl = ttl
	return nil
}
```

`server/internal/user/service_test.go`:
```go
package user_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/emails"
	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

var ctx = context.Background()

func secretary() user.User {
	return user.User{ID: 3, Name: "Club Secretary", Email: "secretary@example.test", Phone: null.StringFrom("0123"),
		Role: role.ClubSecretary, FileName: null.StringFrom("user/sec.png"),
		Hash: null.StringFrom("keep-me"), Salt: null.StringFrom("keep-me-too")}
}

type harness struct {
	svc     *user.Service
	store   *fakeStore
	objects *uploadtest.Storage
	sender  *fakeSender
	tokens  *fakeTokens
}

func newHarness(senderErr error) harness {
	h := harness{
		store:   newFakeStore(secretary()),
		objects: uploadtest.New(),
		sender:  &fakeSender{err: senderErr},
		tokens:  &fakeTokens{tokens: map[string]int{}},
	}
	h.objects.Objects["user/sec.png"] = "IMG"
	teams := fakeTeams{1: team.Team{ID: 1, Name: "First Team"}}
	h.svc = user.NewService(h.store, teams, upload.New(h.objects), h.sender, h.tokens,
		user.HashParams{WorkFactor: 2, BlockSize: 1, Parallelism: 1, KeyLength: 32}, "afc.example.test")
	return h
}

func fieldsOf(t *testing.T, err error) map[string]string {
	t.Helper()
	se, ok := svcerr.As(err)
	require.True(t, ok, "want svcerr, got %v", err)
	return se.Fields
}

func TestCreateValidates(t *testing.T) {
	h := newHarness(nil)
	for field, in := range map[string]user.CreateInput{
		"name":   {Email: "a@b.test", Role: "treasurer"},
		"email":  {Name: "x", Email: "not-email", Role: "treasurer"},
		"role":   {Name: "x", Email: "a@b.test", Role: "president"},
		"teamId": {Name: "x", Email: "a@b.test", Role: "manager", TeamID: 9},
	} {
		_, err := h.svc.Create(ctx, in, nil)
		assert.Contains(t, fieldsOf(t, err), field)
	}
}

func TestCreateSendsSignupEmail(t *testing.T) {
	h := newHarness(nil)
	created, err := h.svc.Create(ctx, user.CreateInput{Name: "New", Email: "new@example.test", Role: "manager", TeamID: 1}, nil)
	require.NoError(t, err)
	assert.True(t, created.EmailSent)
	assert.Empty(t, created.TempPassword, "never return the password when it was emailed")
	require.Len(t, h.sender.sent, 1)
	assert.Equal(t, "new@example.test", h.sender.sent[0].To)

	row := h.store.row(created.User.ID)
	assert.True(t, row.ResetPassword, "new users must reset their password on first login")
	assert.True(t, row.Hash.Valid)
	assert.Equal(t, 1, row.TeamID)
}

func TestCreateWithoutMailerReturnsPassword(t *testing.T) {
	h := newHarness(emails.ErrNoMailer)
	created, err := h.svc.Create(ctx, user.CreateInput{Name: "New", Email: "new@example.test", Role: "treasurer", TeamID: 1}, nil)
	require.NoError(t, err)
	assert.False(t, created.EmailSent)
	assert.NotEmpty(t, created.TempPassword)
	assert.Zero(t, h.store.row(created.User.ID).TeamID, "only managers keep a team")
}

func TestCreateDuplicateEmailIsConflict(t *testing.T) {
	h := newHarness(nil)
	_, err := h.svc.Create(ctx, user.CreateInput{Name: "Dup", Email: "secretary@example.test", Role: "treasurer"}, nil)
	se, ok := svcerr.As(err)
	require.True(t, ok)
	assert.Equal(t, svcerr.KindConflict, se.Kind)
}

func TestUpdatePreservesPasswordAndClearsPhone(t *testing.T) {
	h := newHarness(nil)
	empty := ""
	_, err := h.svc.Update(ctx, 3, user.UpdateInput{Phone: &empty}, nil)
	require.NoError(t, err)
	row := h.store.row(3)
	assert.False(t, row.Phone.Valid)
	assert.Equal(t, "keep-me", row.Hash.String, "profile edits never touch credentials")
	assert.Equal(t, role.ClubSecretary, row.Role)
}

func TestDelete(t *testing.T) {
	h := newHarness(nil)
	deleted, err := h.svc.Delete(ctx, 3)
	require.NoError(t, err)
	assert.Equal(t, "Club Secretary", deleted.Name)
	assert.Equal(t, []string{"user/sec.png"}, h.objects.Deleted)
}

func TestResetPassword(t *testing.T) {
	h := newHarness(nil)
	res, err := h.svc.ResetPassword(ctx, 3)
	require.NoError(t, err)
	assert.True(t, res.EmailSent)
	assert.Empty(t, res.ResetURL)
	assert.True(t, h.store.row(3).ResetPassword)
	require.Len(t, h.tokens.tokens, 1)
	assert.Equal(t, 7*24*time.Hour, h.tokens.ttl)
	for token, id := range h.tokens.tokens {
		assert.Equal(t, 3, id)
		assert.Contains(t, fmt.Sprint(h.sender.sent[0].TplData), token)
	}

	h = newHarness(emails.ErrNoMailer)
	res, err = h.svc.ResetPassword(ctx, 3)
	require.NoError(t, err)
	assert.False(t, res.EmailSent)
	assert.True(t, strings.HasPrefix(res.ResetURL, "https://afc.example.test/reset/"))
}
```
Run: `go test ./server/internal/user/`. Expected: compile errors.

- [ ] **Step 4: Implement `types.go` and `service.go`**

`server/internal/user/types.go`:
```go
package user

// Admin is a user as the user-management API returns it. It never contains
// password material.
type Admin struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone,omitempty"`
	Role     string `json:"role"`     // display name, e.g. "Club Secretary"
	RoleCode string `json:"roleCode"` // input code, e.g. "club_secretary"
	TeamID   int    `json:"teamId,omitempty"`
	ImageURL string `json:"imageUrl,omitempty"`
}

// CreateInput is a new user. Role is a role code (e.g. "manager"); TeamID
// is kept only for managers.
type CreateInput struct {
	Name   string
	Email  string
	Phone  string
	Role   string
	TeamID int
}

// UpdateInput changes a user; nil fields are left unchanged.
type UpdateInput struct {
	Name        *string
	Email       *string
	Phone       *string
	Role        *string
	TeamID      *int
	RemoveImage bool
}

// Created is the result of creating a user. TempPassword is set only when
// the signup email could not be sent, so the admin can pass it on.
type Created struct {
	User         Admin  `json:"user"`
	EmailSent    bool   `json:"emailSent"`
	TempPassword string `json:"tempPassword,omitempty"`
}

// ResetResult is the result of an admin-triggered password reset. ResetURL
// is set only when the email could not be sent.
type ResetResult struct {
	EmailSent bool   `json:"emailSent"`
	ResetURL  string `json:"resetUrl,omitempty"`
}
```

`server/internal/user/service.go`:
```go
package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/emails"
	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/utils"
)

const (
	uploadCategory = "user"
	resetTokenTTL  = 7 * 24 * time.Hour
)

type store interface {
	GetUsers(ctx context.Context) ([]User, error)
	GetUser(ctx context.Context, userParam User) (User, error)
	AddUser(ctx context.Context, userParam User) (User, error)
	EditUser(ctx context.Context, userParam User) (User, error)
	DeleteUser(ctx context.Context, userParam User) error
}

// TeamGetter checks manager teams (satisfied by *team.Store).
type TeamGetter interface {
	GetTeam(ctx context.Context, teamParam team.Team) (team.Team, error)
}

// TokenSetter issues password reset tokens (satisfied by *auth.Tokens).
type TokenSetter interface {
	Set(ctx context.Context, token string, userID int, ttl time.Duration) error
}

// HashParams are the scrypt parameters for new users' initial passwords.
type HashParams struct {
	WorkFactor  int
	BlockSize   int
	Parallelism int
	KeyLength   int
}

// Service is user administration, shared by the API and legacy views.
type Service struct {
	store  store
	teams  TeamGetter
	files  *upload.Files
	mailer emails.Sender
	tokens TokenSetter
	hash   HashParams
	domain string
}

func NewService(store store, teams TeamGetter, files *upload.Files, mailer emails.Sender, tokens TokenSetter,
	hash HashParams, domain string) *Service {
	return &Service{store: store, teams: teams, files: files, mailer: mailer, tokens: tokens, hash: hash, domain: domain}
}

func (s *Service) admin(u User) Admin {
	return Admin{
		ID:       u.ID,
		Name:     u.Name,
		Email:    u.Email,
		Phone:    u.Phone.String,
		Role:     u.Role.String(),
		RoleCode: strings.ToLower(u.Role.DBString()),
		TeamID:   u.TeamID,
		ImageURL: s.files.URL(u.FileName.String),
	}
}

func (s *Service) get(ctx context.Context, id int) (User, error) {
	u, err := s.store.GetUser(ctx, User{ID: id})
	if err == nil && u.ID != id {
		// GetUser matches email OR id; a blank-email row must not stand in.
		return User{}, svcerr.NotFound("user not found", nil)
	}
	if err != nil {
		return User{}, svcerr.FromStore(err, "user")
	}
	return u, nil
}

func (s *Service) List(ctx context.Context) ([]Admin, error) {
	ctx, span := tracer.Start(ctx, "user.Service.List")
	defer span.End()
	rows, err := s.store.GetUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	out := make([]Admin, 0, len(rows))
	for _, u := range rows {
		out = append(out, s.admin(u))
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, id int) (Admin, error) {
	ctx, span := tracer.Start(ctx, "user.Service.Get")
	defer span.End()
	u, err := s.get(ctx, id)
	if err != nil {
		return Admin{}, err
	}
	return s.admin(u), nil
}

// validate checks u and normalises its team: only managers keep one.
func (s *Service) validate(ctx context.Context, u *User, roleCode string) error {
	f := svcerr.Fields{}
	u.Name = strings.TrimSpace(u.Name)
	u.Email = strings.TrimSpace(u.Email)
	if u.Name == "" {
		f.Add("name", "name is required")
	}
	if !emailverifier.IsAddressValid(u.Email) {
		f.Add("email", "email address is not valid")
	}
	if roleCode != "" {
		r, err := role.GetRole(roleCode)
		if err != nil {
			f.Add("role", "unknown role")
		} else {
			u.Role = r
		}
	}
	if u.Role == role.Manager {
		if _, err := s.teams.GetTeam(ctx, team.Team{ID: u.TeamID}); err != nil {
			f.Add("teamId", "managers need an existing team")
		}
	} else {
		u.TeamID = 0
	}
	return f.Err()
}

func conflictOnDuplicate(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return svcerr.Conflict("email address is already in use", err)
	}
	return err
}

// Create adds a user with a generated password they must reset on first
// login, and emails it to them. If the email can't be sent the password is
// returned so the admin can pass it on.
func (s *Service) Create(ctx context.Context, in CreateInput, image *upload.File) (Created, error) {
	ctx, span := tracer.Start(ctx, "user.Service.Create")
	defer span.End()

	phone := strings.TrimSpace(in.Phone)
	u := User{Name: in.Name, Email: in.Email, Phone: null.NewString(phone, phone != ""), TeamID: in.TeamID, ResetPassword: true}
	if in.Role == "" {
		return Created{}, svcerr.InvalidField("role", "role is required")
	}
	if err := s.validate(ctx, &u, in.Role); err != nil {
		return Created{}, err
	}

	password, err := utils.GenerateRandom(utils.GeneratePassword)
	if err != nil {
		return Created{}, fmt.Errorf("failed to generate password: %w", err)
	}
	salt, err := utils.GenerateRandom(utils.GenerateSalt)
	if err != nil {
		return Created{}, fmt.Errorf("failed to generate salt: %w", err)
	}
	// Same hashing as legacy UserAddFunc (raw salt string); the user must
	// reset before this hash is ever checked.
	hash, err := utils.HashPassScrypt([]byte(password), []byte(salt), s.hash.WorkFactor, s.hash.BlockSize, s.hash.Parallelism, s.hash.KeyLength)
	if err != nil {
		return Created{}, fmt.Errorf("failed to hash password: %w", err)
	}
	u.Hash = null.StringFrom(hash)
	u.Salt = null.NewString(salt, salt != "")

	if image != nil {
		key, saveErr := s.files.Save(ctx, image, uploadCategory)
		if saveErr != nil {
			return Created{}, saveErr
		}
		u.FileName = null.StringFrom(key)
	}
	added, err := s.store.AddUser(ctx, u)
	if err != nil {
		s.files.Remove(ctx, u.FileName.String)
		return Created{}, conflictOnDuplicate(fmt.Errorf("failed to add user: %w", err))
	}

	out := Created{User: s.admin(added)}
	msg, err := emails.Signup(added.Email, added.Name, password, s.domain)
	if err == nil {
		err = s.mailer.Send(ctx, msg)
	}
	if err != nil {
		slog.InfoContext(ctx, fmt.Sprintf("signup email not sent to user %d: %+v", added.ID, err))
		out.TempPassword = password
		return out, nil
	}
	out.EmailSent = true
	return out, nil
}

func (s *Service) Update(ctx context.Context, id int, in UpdateInput, image *upload.File) (Admin, error) {
	ctx, span := tracer.Start(ctx, "user.Service.Update")
	defer span.End()
	u, err := s.get(ctx, id)
	if err != nil {
		return Admin{}, err
	}
	if in.Name != nil {
		u.Name = *in.Name
	}
	if in.Email != nil {
		u.Email = *in.Email
	}
	if in.Phone != nil {
		phone := strings.TrimSpace(*in.Phone)
		u.Phone = null.NewString(phone, phone != "")
	}
	if in.TeamID != nil {
		u.TeamID = *in.TeamID
	}
	roleCode := ""
	if in.Role != nil {
		roleCode = *in.Role
	}
	if err = s.validate(ctx, &u, roleCode); err != nil {
		return Admin{}, err
	}

	oldKey, newKey := u.FileName.String, ""
	switch {
	case image != nil:
		if newKey, err = s.files.Save(ctx, image, uploadCategory); err != nil {
			return Admin{}, err
		}
		u.FileName = null.StringFrom(newKey)
	case in.RemoveImage:
		u.FileName = null.String{}
	}
	if _, err = s.store.EditUser(ctx, u); err != nil {
		s.files.Remove(ctx, newKey)
		return Admin{}, conflictOnDuplicate(fmt.Errorf("failed to edit user: %w", err))
	}
	if oldKey != "" && oldKey != u.FileName.String {
		s.files.Remove(ctx, oldKey)
	}
	return s.admin(u), nil
}

// Delete removes the user and their photo, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Admin, error) {
	ctx, span := tracer.Start(ctx, "user.Service.Delete")
	defer span.End()
	u, err := s.get(ctx, id)
	if err != nil {
		return Admin{}, err
	}
	deleted := s.admin(u)
	if err = s.store.DeleteUser(ctx, u); err != nil {
		return Admin{}, fmt.Errorf("failed to delete user: %w", err)
	}
	s.files.Remove(ctx, u.FileName.String)
	return deleted, nil
}

// ResetPassword forces a reset on next login and emails a 7-day reset link.
// If the email can't be sent, the link is returned for the admin to pass on.
func (s *Service) ResetPassword(ctx context.Context, id int) (ResetResult, error) {
	ctx, span := tracer.Start(ctx, "user.Service.ResetPassword")
	defer span.End()
	u, err := s.get(ctx, id)
	if err != nil {
		return ResetResult{}, err
	}
	u.ResetPassword = true
	if _, err = s.store.EditUser(ctx, u); err != nil {
		return ResetResult{}, fmt.Errorf("failed to flag user for reset: %w", err)
	}
	token := uuid.NewString()
	if err = s.tokens.Set(ctx, token, u.ID, resetTokenTTL); err != nil {
		return ResetResult{}, fmt.Errorf("failed to store reset token: %w", err)
	}
	link := fmt.Sprintf("https://%s/reset/%s", s.domain, token)

	msg, err := emails.Reset(u.Email, link)
	if err == nil {
		err = s.mailer.Send(ctx, msg)
	}
	if err != nil {
		slog.InfoContext(ctx, fmt.Sprintf("reset email not sent to user %d: %+v", u.ID, err))
		return ResetResult{ResetURL: link}, nil
	}
	return ResetResult{EmailSent: true}, nil
}
```
Run: `go test ./server/internal/user/`. Expected: PASS.

- [ ] **Step 5: Failing handler tests, then `handlers.go`**

`server/internal/user/handlers_test.go`:
```go
package user_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestUserRoutes(t *testing.T) {
	h := newHarness(nil)
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	user.NewHandlers(h.svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	treasurer := anon.As(authtest.Cookie(t, sessions, authtest.Treasurer))
	admin := anon.As(authtest.Cookie(t, sessions, authtest.Webmaster))

	for _, path := range []string{"/api/v1/users", "/api/v1/users/3"} {
		assert.Equal(t, http.StatusUnauthorized, anon.Get(t, path).Code, path)
		assert.Equal(t, http.StatusForbidden, treasurer.Get(t, path).Code, path)
	}

	rec := admin.Get(t, "/api/v1/users")
	require.Equal(t, http.StatusOK, rec.Code)
	for _, secret := range []string{"keep-me", "hash", "salt", "password"} {
		assert.NotContains(t, rec.Body.String(), secret)
	}

	rec = admin.Multipart(t, http.MethodPost, "/api/v1/users",
		map[string]string{"name": "New", "email": "new@example.test", "role": "treasurer"})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.True(t, apitest.Decode[user.Created](t, rec).EmailSent)

	rec = admin.Multipart(t, http.MethodPost, "/api/v1/users",
		map[string]string{"name": "Dup", "email": "secretary@example.test", "role": "treasurer"})
	assert.Equal(t, http.StatusConflict, rec.Code)

	rec = admin.Multipart(t, http.MethodPatch, "/api/v1/users/3", map[string]string{"phone": ""})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	rec = admin.JSON(t, http.MethodPost, "/api/v1/users/3/reset", nil)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, apitest.Decode[user.ResetResult](t, rec).EmailSent)

	assert.Equal(t, http.StatusNoContent, admin.JSON(t, http.MethodDelete, "/api/v1/users/3", nil).Code)
}
```

`server/internal/user/handlers.go`:
```go
package user

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /users. Every route needs Club Secretary or higher.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the user-management routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	admin := guards.ClubSecretaryHigher
	g.GET("/users", h.list, admin)
	g.POST("/users", h.create, admin)
	g.GET("/users/:id", h.get, admin)
	g.PATCH("/users/:id", h.update, admin)
	g.DELETE("/users/:id", h.remove, admin)
	g.POST("/users/:id/reset", h.reset, admin)
}

// list returns every user.
//
//	@Summary	List users
//	@Tags		users
//	@Produce	json
//	@Success	200	{array}		user.Admin
//	@Failure	403	{object}	web.ErrorResponse
//	@Router		/users [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// get returns one user.
//
//	@Summary	Get a user
//	@Tags		users
//	@Produce	json
//	@Param		id	path		int	true	"User ID"
//	@Success	200	{object}	user.Admin
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/users/{id} [get]
func (h *Handlers) get(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	out, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// create adds a user and emails them a temporary password.
//
//	@Summary	Create a user
//	@Tags		users
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		name	formData	string	true	"Name"
//	@Param		email	formData	string	true	"Email"
//	@Param		phone	formData	string	false	"Phone"
//	@Param		role	formData	string	true	"Role code, e.g. manager, club_secretary"
//	@Param		teamId	formData	int		false	"Team (managers only)"
//	@Param		image	formData	file	false	"Photo"
//	@Success	201		{object}	user.Created
//	@Failure	409		{object}	web.ErrorResponse
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/users [post]
func (h *Handlers) create(c echo.Context) error {
	teamID, err := web.FormInt(c, "teamId")
	if err != nil {
		return err
	}
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	in := CreateInput{Name: c.FormValue("name"), Email: c.FormValue("email"), Phone: c.FormValue("phone"), Role: c.FormValue("role")}
	if teamID != nil {
		in.TeamID = *teamID
	}
	out, err := h.svc.Create(c.Request().Context(), in, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// update changes a user; omitted fields are left as they are.
//
//	@Summary	Update a user
//	@Tags		users
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		id			path		int		true	"User ID"
//	@Param		name		formData	string	false	"Name"
//	@Param		email		formData	string	false	"Email"
//	@Param		phone		formData	string	false	"Phone; empty clears"
//	@Param		role		formData	string	false	"Role code"
//	@Param		teamId		formData	int		false	"Team (managers only)"
//	@Param		image		formData	file	false	"Replacement photo"
//	@Param		removeImage	formData	bool	false	"Remove the current photo"
//	@Success	200			{object}	user.Admin
//	@Failure	404			{object}	web.ErrorResponse
//	@Failure	409			{object}	web.ErrorResponse
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/users/{id} [patch]
func (h *Handlers) update(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	in := UpdateInput{
		Name:  web.FormString(c, "name"),
		Email: web.FormString(c, "email"),
		Phone: web.FormString(c, "phone"),
		Role:  web.FormString(c, "role"),
	}
	if in.TeamID, err = web.FormInt(c, "teamId"); err != nil {
		return err
	}
	remove, err := web.FormBool(c, "removeImage")
	if err != nil {
		return err
	}
	in.RemoveImage = remove != nil && *remove
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	out, err := h.svc.Update(c.Request().Context(), id, in, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// remove deletes a user.
//
//	@Summary	Delete a user
//	@Tags		users
//	@Param		id	path	int	true	"User ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/users/{id} [delete]
func (h *Handlers) remove(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if _, err = h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// reset forces a password reset and emails the user a link.
//
//	@Summary	Reset a user's password
//	@Tags		users
//	@Produce	json
//	@Param		id	path		int	true	"User ID"
//	@Success	200	{object}	user.ResetResult
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/users/{id}/reset [post]
func (h *Handlers) reset(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	out, err := h.svc.ResetPassword(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}
```
Run: `go test ./server/internal/user/`. Expected: PASS.

- [ ] **Step 6: Legacy writes → service**

Add `userSvc` / `UserService` to `Views`/`Deps`/`New`. In `server/internal/legacy/views/user.go` keep `UsersFunc` (and `UsersSetDisplayEmailFunc` from Task 17). Replace `UserAddFunc`, `UserEditFunc` and `UserDeleteFunc`:
```go
func (v *Views) UserAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.UserAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	var data struct {
		Error string `json:"error"`
	}
	teamID, err := strconv.Atoi(c.FormValue("userTeam"))
	if err != nil || teamID < 0 {
		teamID = 0
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for user add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	created, err := v.userSvc.Create(c.Request().Context(), user.CreateInput{
		Name: c.FormValue("name"), Email: c.FormValue("email"), Phone: c.FormValue("phone"),
		Role: c.FormValue("role"), TeamID: teamID,
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add user for user add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add user for user add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	msg := fmt.Sprintf("successfully created user, sent signup email to: \"%s\"", created.User.Email)
	if !created.EmailSent {
		msg = html.UnescapeString(fmt.Sprintf("successfully created user - failed to send email. Please send the username and password to this email: %s, password: %s",
			created.User.Email, created.TempPassword))
	}
	v.flash(c, c1, msg)
	return c.JSON(http.StatusOK, data)
}

func (v *Views) UserEditFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.UserEditFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to parse id for user edit, error: %w", err)
	}
	var data struct {
		Error string `json:"error"`
	}
	remove := c.FormValue("removeUserImage")
	if remove != "" && remove != "Y" {
		data.Error = "failed to parse removeUserImage for user edit: " + remove
		return c.JSON(http.StatusOK, data)
	}
	teamID, err := strconv.Atoi(c.FormValue("userTeam"))
	if err != nil || teamID < 0 {
		teamID = 0
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for user edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	name, email, phone, roleCode := c.FormValue("name"), c.FormValue("email"), c.FormValue("phone"), c.FormValue("role")
	updated, err := v.userSvc.Update(c.Request().Context(), userID, user.UpdateInput{
		Name: &name, Email: &email, Phone: &phone, Role: &roleCode, TeamID: &teamID, RemoveImage: remove == "Y",
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to edit user for user edit, user id: %d, error: %+v", userID, err))
		data.Error = fmt.Sprintf("failed to edit user for user edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully edited \"%s\"", updated.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) UserDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.UserDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for user delete, error: %w", err)
	}
	deleted, err := v.userSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete user for user delete, user id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", deleted.Name))
	return c.Redirect(http.StatusFound, "/users")
}
```
In `server/internal/legacy/views/reset.go`, keep `ResetURLFunc` (Task 22 handles it) and replace `ResetUserPasswordFunc`:
```go
func (v *Views) ResetUserPasswordFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ResetUserPasswordFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to parse user id for reset, error: %w", err)
	}
	u, err := v.userSvc.Get(c.Request().Context(), userID)
	if err != nil {
		return fmt.Errorf("failed to get user for reset, user id: %d, error: %w", userID, err)
	}
	res, err := v.userSvc.ResetPassword(c.Request().Context(), userID)
	if err != nil {
		return fmt.Errorf("failed to reset password, user id: %d, error: %w", userID, err)
	}
	var message struct {
		Message string `json:"message"`
		Error   error  `json:"error"`
	}
	if res.EmailSent {
		message.Message = fmt.Sprintf("Reset email sent to: \"%s\"", u.Email)
	} else {
		message.Message = fmt.Sprintf("Please forward the link to this email: %s, reset link: %s", u.Email, res.ResetURL)
		message.Error = errors.New("failed to send reset email")
	}
	return c.JSON(http.StatusOK, message)
}
```
`legacy/views` no longer uses `v.mailer` after this task. Leave the field until Task 23 removes it.

- [ ] **Step 7: Wire, add guarded GETs, regenerate, run, commit**

In `app.Build`:
```go
	sender := emails.NewSMTP(mailer)
	userSvc := user.NewService(s.User, s.Team, uploads, sender, tokens,
		user.HashParams{
			WorkFactor:  conf.Passwords.ScryptWorkFactor,
			BlockSize:   conf.Passwords.ScryptBlockSize,
			Parallelism: conf.Passwords.ScryptParallelismFactor,
			KeyLength:   conf.Passwords.KeyLength,
		}, conf.DomainName)
```
`tokens` must be declared before this line; move `tokens := auth.NewTokens(conf.Redis)` up if needed. Add `UserService: userSvc,` to `views.Deps` and `user.NewHandlers(userSvc).Register(api, guards)`.

In `server/internal/app/helpers_test.go` append `"/api/v1/users", "/api/v1/users/1",` to `guardedGets`.
```bash
go generate ./server/internal/docs
goimports -w server/internal/legacy/views server/internal/legacy/templates
gofmt -l server; go build ./... && go vet ./... && go test ./... && go tool golangci-lint run ./...
git add -A server
git commit -m "Add user administration service and /api/v1/users; move emails out of legacy templates

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---
### Task 21: Security fix: the forced-reset login hands out reset links without a password

**Found while planning; not in the spec.** `user.Store.VerifyUser` starts with:
```go
	if user.ResetPassword {
		userParam.ID = user.ID
		return userParam, true, errors.New("password reset required")
	}
```
That runs **before any password check**. Legacy `LoginFunc` (and the spec's API login) answers `resetPw == true` by issuing a 1-hour reset token and returning `/reset/<token>`. So anyone who types the email address of a reset-flagged account (every newly created user, and anyone an admin just reset) can set that account's password.

A second bug hides the first. New users are hashed as `scrypt(password, []byte(saltHexString))`, but verification uses `hex.DecodeString(salt)`. A new user's emailed temporary password therefore never verifies, and the leaky early return is the only thing that let new users in.

The fix:
1. Remove the early return.
2. Recognise the new-user hash format.
3. Report "reset required" only after the password verified.

Every existing path already returns `(user, true, …)` for reset-flagged users after a successful match. This fixes the live template site as well as the API.

**This is live on `main` today.** Recommend shipping this task on its own as a hotfix PR to `main` too; it only touches `user.go` and one test file, and has no dependency on the rest of the plan.

**Files:**
- Modify: `server/internal/user/user.go` (`VerifyUser`)
- Create: `server/internal/user/verify_test.go`

**Interfaces:**
- `VerifyUser` signature unchanged. Behaviour change: for reset-flagged users it returns `(user, true, "password reset required")` only when the password matches, and `(userParam, false, "invalid credentials")` otherwise.

- [ ] **Step 1: Failing DB-backed tests**

`server/internal/user/verify_test.go`:
```go
package user_test

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/utils"
)

const (
	iter, workFactor, blockSize, parallelism, keyLen = 1, 2, 1, 1, 32
)

func verify(t *testing.T, store *user.Store, email, password string) (user.User, bool, error) {
	t.Helper()
	return store.VerifyUser(context.Background(), user.User{Email: email, Password: null.StringFrom(password)},
		iter, workFactor, blockSize, parallelism, keyLen)
}

// newUserLikeSignup stores a user exactly as legacy UserAddFunc and
// user.Service.Create do: raw hex-salt bytes, reset_password = true.
func newUserLikeSignup(t *testing.T, store *user.Store, email, password string) {
	t.Helper()
	salt, err := utils.GenerateRandom(utils.GenerateSalt)
	require.NoError(t, err)
	hash, err := utils.HashPassScrypt([]byte(password), []byte(salt), workFactor, blockSize, parallelism, keyLen)
	require.NoError(t, err)
	_, err = store.AddUser(context.Background(), user.User{Name: "New", Email: email, Role: role.Treasurer,
		ResetPassword: true, Hash: null.StringFrom(hash), Salt: null.StringFrom(salt)})
	require.NoError(t, err)
}

func TestResetFlaggedUserNeedsTheRightPassword(t *testing.T) {
	db, _ := testdb.Open(t)
	store := user.NewUserRepo(db)
	newUserLikeSignup(t, store, "new@example.test", "Temp-Pa55!")

	_, reset, err := verify(t, store, "new@example.test", "a guess")
	require.Error(t, err)
	assert.False(t, reset, "a wrong password must never be told 'reset required'")

	u, reset, err := verify(t, store, "new@example.test", "Temp-Pa55!")
	require.Error(t, err, "reset-required is still reported as an error to callers")
	assert.True(t, reset, "the emailed temporary password must now work")
	assert.Positive(t, u.ID)
}

func TestAdminResetUserKeepsWorkingPassword(t *testing.T) {
	db, _ := testdb.Open(t)
	store := user.NewUserRepo(db)
	ctx := context.Background()

	// An established account: hex-decoded salt, as EditUserPassword writes.
	saltHex, err := utils.GenerateRandom(utils.GenerateSalt)
	require.NoError(t, err)
	saltBytes, err := hex.DecodeString(saltHex)
	require.NoError(t, err)
	hash, err := utils.HashPassScrypt([]byte("Known-Pa55!"), saltBytes, workFactor, blockSize, parallelism, keyLen)
	require.NoError(t, err)
	_, err = store.AddUser(ctx, user.User{Name: "Old", Email: "old@example.test", Role: role.Treasurer,
		Hash: null.StringFrom(hash), Salt: null.StringFrom(saltHex)})
	require.NoError(t, err)

	_, reset, err := verify(t, store, "old@example.test", "Known-Pa55!")
	require.NoError(t, err)
	assert.False(t, reset)

	u, err := store.GetUser(ctx, user.User{Email: "old@example.test"})
	require.NoError(t, err)
	u.ResetPassword = true // admin clicked "reset password"
	_, err = store.EditUser(ctx, u)
	require.NoError(t, err)

	_, reset, _ = verify(t, store, "old@example.test", "wrong")
	assert.False(t, reset)
	_, reset, _ = verify(t, store, "old@example.test", "Known-Pa55!")
	assert.True(t, reset)
}
```
Run with `AFC_TEST_DB`: `go test -run 'TestResetFlaggedUser|TestAdminResetUser' -v ./server/internal/user/`
Expected: FAIL. `a guess` gets `reset == true` (the vulnerability).

- [ ] **Step 2: Fix `VerifyUser`**

In `server/internal/user/user.go` `VerifyUser`:
1. Delete the early-return block:
```go
	if user.ResetPassword {
		userParam.ID = user.ID
		return userParam, true, errors.New("password reset required")
	}
```
2. Replace the final scrypt comparison (from `scryptHash, err := utils.HashPassScrypt(…saltDecode…)` to the end of the function) with:
```go
	scryptHash, err := utils.HashPassScrypt([]byte(userParam.Password.String), saltDecode, workFactor, blockSize, parallelismFactor, keyLen)
	if err != nil {
		return userParam, false, fmt.Errorf("failed to generate password hash verify: %w", err)
	}
	match := scryptHash == user.Hash.String
	if !match && user.Salt.Valid {
		// Accounts created by the signup flow hash with the salt's hex text
		// as raw bytes rather than the decoded salt; accept that form too.
		var rawHash string
		rawHash, err = utils.HashPassScrypt([]byte(userParam.Password.String), []byte(user.Salt.String), workFactor, blockSize, parallelismFactor, keyLen)
		if err != nil {
			return userParam, false, fmt.Errorf("failed to generate password hash verify: %w", err)
		}
		match = rawHash == user.Hash.String
	}
	if !match {
		return userParam, false, errors.New("invalid credentials")
	}
	user.Hash = null.NewString("", false)
	user.Salt = null.NewString("", false)
	if user.ResetPassword {
		return user, true, errors.New("password reset required")
	}
	return user, false, nil
}
```
3. `hex.DecodeString(user.Salt.String)` above this block never fails for signup salts, because they are hex. Leave it as it is.

Run the two tests again: PASS. Then run `go test ./...`: all green.

- [ ] **Step 3: Commit (and consider a separate hotfix)**

```bash
git add server/internal/user/user.go server/internal/user/verify_test.go
git commit -m "Fix forced-reset login issuing reset links without checking the password

VerifyUser returned 'reset required' for reset-flagged accounts before
verifying the password, and login then issued a reset link, so anyone who
knew the email of a new or admin-reset account could set its password.
Verify first; also accept the raw-salt hash format signup has always
written, which is why new users' temporary passwords never verified.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```
To hotfix production independently: `git switch main && git switch -c fix-reset-login && git cherry-pick <this commit>` then adjust paths (`user/user.go`, and move the test to `user/verify_test.go` using the root-layout import paths, or ship without the test if `testdb` isn't on `main`).

---
### Task 22: Auth endpoints (login, logout, password, reset) and account

Depends on Task 21. Without it, `POST /auth/login` would reproduce the reset-link takeover.

**Files:**
- Create: `server/internal/auth/service.go`, `server/internal/auth/service_test.go`, `server/internal/auth/handlers_test.go`
- Modify: `server/internal/auth/handlers.go`, `server/internal/auth/types.go`, `server/internal/auth/sessions_test.go` (constructor call)
- Create: `server/internal/account/{service,handlers,service_test}.go`
- Modify: `server/internal/legacy/views/{login,changePassword,reset,account,helpers}.go`, `views.go`, `server/internal/app/app.go`, `server/internal/app/helpers_test.go`

**Interfaces:**
- Consumes: `auth.Sessions`, `auth.Tokens` (Tasks 5–6); `user.Store.VerifyUser/EditUserPassword/GetUser/EditUserImage` (fixed in Task 21).
- Produces (`auth`):
  - `type UserStore interface{GetUser; VerifyUser; EditUserPassword}`
  - `var ErrInvalidCredentials`
  - `type LoginResult struct{User user.User; ResetRequired bool; ResetURL string}`
  - `PasswordProblems(pw string) string` (the legacy `minRequirementsMet` rules)
  - `NewService(UserStore, *Tokens, PasswordConfig) *Service`, with methods:
    - `Login(ctx, email, password) (LoginResult, error)`
    - `ChangePassword(ctx, u user.User, old, new, confirm string) error`
    - `CheckResetToken(ctx, token) error`
    - `ResetPassword(ctx, token, new, confirm string) error`
  - `NewHandlers(*Sessions, *Service, *upload.Files) *Handlers` (**signature change** from Task 5)
  - JSON types: `LoginInput{Email, Password string; Remember bool}`, `LoginResponse{User *CurrentUser; ResetRequired bool; ResetURL string}`, `PasswordInput{OldPassword, NewPassword, ConfirmationPassword string}`, `ResetInput{NewPassword, ConfirmationPassword string}`
- Produces (`account`):
  - `NewService(Store, *upload.Files) *Service`, with methods `SetImage(ctx, u user.User, *upload.File) (user.User, error)` and `RemoveImage(ctx, u user.User) (user.User, error)`
  - `NewHandlers(*Service, *upload.Files) *Handlers`, `Register`
- Produces (legacy): `Deps.AuthService`, `Deps.AccountService`.

- [ ] **Step 1: Failing service tests for `auth.Service`**

`server/internal/auth/service_test.go`:
```go
package auth_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

var ctx = context.Background()

// fakeUsers accepts password "Right-Pa55" for known emails; reset-flagged
// accounts only report reset-required after a correct password (Task 21).
type fakeUsers struct {
	byEmail map[string]user.User
	changed map[int]string
}

func newFakeUsers() *fakeUsers {
	return &fakeUsers{
		byEmail: map[string]user.User{
			"ok@example.test":    {ID: 1, Email: "ok@example.test", Role: role.Treasurer},
			"reset@example.test": {ID: 2, Email: "reset@example.test", Role: role.Treasurer, ResetPassword: true},
		},
		changed: map[int]string{},
	}
}

func (f *fakeUsers) GetUser(_ context.Context, u user.User) (user.User, error) {
	for _, x := range f.byEmail {
		if x.ID == u.ID {
			return x, nil
		}
	}
	return user.User{}, fmt.Errorf("failed to get user: %w", sql.ErrNoRows)
}

func (f *fakeUsers) VerifyUser(_ context.Context, u user.User, _, _, _, _, _ int) (user.User, bool, error) {
	x, ok := f.byEmail[u.Email]
	if !ok || u.Password.String != "Right-Pa55" {
		return u, false, errors.New("invalid credentials")
	}
	if x.ResetPassword {
		return x, true, errors.New("password reset required")
	}
	return x, false, nil
}

func (f *fakeUsers) EditUserPassword(_ context.Context, u user.User, _, _, _, _ int) error {
	f.changed[u.ID] = u.Password.String
	return nil
}

func newAuthService() (*auth.Service, *fakeUsers, *auth.Tokens) {
	users := newFakeUsers()
	tokens := auth.NewTokens(auth.RedisConfig{})
	return auth.NewService(users, tokens, auth.PasswordConfig{}), users, tokens
}

func TestPasswordProblems(t *testing.T) {
	assert.Empty(t, auth.PasswordProblems("Str0ng-Pass!"))
	for _, weak := range []string{"short1!A", "alllowercase1!", "ALLUPPERCASE1!", "NoDigits!!aa", "NoSpecial1aaA"} {
		assert.NotEmpty(t, auth.PasswordProblems(weak), weak)
	}
}

func TestLogin(t *testing.T) {
	svc, _, tokens := newAuthService()

	res, err := svc.Login(ctx, "ok@example.test", "Right-Pa55")
	require.NoError(t, err)
	assert.Equal(t, 1, res.User.ID)
	assert.False(t, res.ResetRequired)

	_, err = svc.Login(ctx, "ok@example.test", "wrong")
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	_, err = svc.Login(ctx, "reset@example.test", "wrong")
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials, "no reset link without the password")

	res, err = svc.Login(ctx, "reset@example.test", "Right-Pa55")
	require.NoError(t, err)
	assert.True(t, res.ResetRequired)
	require.True(t, strings.HasPrefix(res.ResetURL, "/reset/"))
	id, ok := tokens.Get(ctx, strings.TrimPrefix(res.ResetURL, "/reset/"))
	assert.True(t, ok)
	assert.Equal(t, 2, id)
}

func TestChangePassword(t *testing.T) {
	svc, users, _ := newAuthService()
	me := users.byEmail["ok@example.test"]
	fields := func(err error) map[string]string {
		se, ok := svcerr.As(err)
		require.True(t, ok, "want svcerr, got %v", err)
		return se.Fields
	}
	assert.Contains(t, fields(svc.ChangePassword(ctx, me, "wrong", "Str0ng-Pass!", "Str0ng-Pass!")), "oldPassword")
	assert.Contains(t, fields(svc.ChangePassword(ctx, me, "Right-Pa55", "Str0ng-Pass!", "different")), "confirmationPassword")
	assert.Contains(t, fields(svc.ChangePassword(ctx, me, "Right-Pa55", "weak", "weak")), "newPassword")
	require.NoError(t, svc.ChangePassword(ctx, me, "Right-Pa55", "Str0ng-Pass!", "Str0ng-Pass!"))
	assert.Equal(t, "Str0ng-Pass!", users.changed[1])
}

func TestResetFlow(t *testing.T) {
	svc, users, tokens := newAuthService()
	require.NoError(t, tokens.Set(ctx, "tok", 2, time.Hour))

	require.NoError(t, svc.CheckResetToken(ctx, "tok"))
	se, ok := svcerr.As(svc.CheckResetToken(ctx, "nope"))
	require.True(t, ok)
	assert.Equal(t, svcerr.KindNotFound, se.Kind)

	se, ok = svcerr.As(svc.ResetPassword(ctx, "tok", "Str0ng-Pass!", "mismatch"))
	require.True(t, ok)
	assert.Contains(t, se.Fields, "confirmationPassword")

	require.NoError(t, svc.ResetPassword(ctx, "tok", "Str0ng-Pass!", "Str0ng-Pass!"))
	assert.Equal(t, "Str0ng-Pass!", users.changed[2])
	_, still := tokens.Get(ctx, "tok")
	assert.False(t, still, "tokens are single use")
}
```
Run: `go test ./server/internal/auth/`. Expected: compile errors.

- [ ] **Step 2: Implement `auth/service.go`**

```go
package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

// ErrInvalidCredentials is returned for any failed login, without saying why.
var ErrInvalidCredentials = errors.New("invalid email or password")

const loginResetTTL = time.Hour

// UserStore is the credential side of the user store (satisfied by *user.Store).
type UserStore interface {
	GetUser(ctx context.Context, userParam user.User) (user.User, error)
	VerifyUser(ctx context.Context, userParam user.User, iter, workFactor, blockSize, parallelismFactor, keyLen int) (user.User, bool, error)
	EditUserPassword(ctx context.Context, userParam user.User, workFactor, blockSize, parallelismFactor, keyLen int) error
}

// LoginResult is a successful credential check. When ResetRequired is set
// no session should be created; the user must visit ResetURL.
type LoginResult struct {
	User          user.User
	ResetRequired bool
	ResetURL      string
}

// Service handles logins and password changes.
type Service struct {
	users     UserStore
	tokens    *Tokens
	passwords PasswordConfig
}

func NewService(users UserStore, tokens *Tokens, passwords PasswordConfig) *Service {
	return &Service{users: users, tokens: tokens, passwords: passwords}
}

var passwordRules = []struct {
	re  *regexp.Regexp
	msg string
}{
	{regexp.MustCompile(`[a-z]`), "at least 1 lower case letter"},
	{regexp.MustCompile(`[A-Z]`), "at least 1 upper case letter"},
	{regexp.MustCompile(`\d`), "at least 1 number"},
	{regexp.MustCompile(`[@$!%*?&|^£;:/.,<>()_=+~§±#{}-]`), "at least 1 special character"},
}

// PasswordProblems describes what a new password is missing ("" when it is
// acceptable). Rules match legacy minRequirementsMet, including its
// "longer than 8 characters" length rule.
func PasswordProblems(pw string) string {
	var missing []string
	for _, r := range passwordRules {
		if !r.re.MatchString(pw) {
			missing = append(missing, r.msg)
		}
	}
	if len(pw) <= 8 {
		missing = append(missing, "more than 8 characters")
	}
	if len(missing) == 0 {
		return ""
	}
	return "password needs " + strings.Join(missing, ", ")
}

// Login checks credentials. A reset-flagged account (Task 21: only after a
// correct password) gets a one-hour reset link instead of a session.
func (s *Service) Login(ctx context.Context, email, password string) (LoginResult, error) {
	ctx, span := tracer.Start(ctx, "auth.Service.Login")
	defer span.End()
	p := s.passwords
	u, reset, err := s.users.VerifyUser(ctx, user.User{Email: strings.TrimSpace(email), Password: null.StringFrom(password)},
		p.Iterations, p.ScryptWorkFactor, p.ScryptBlockSize, p.ScryptParallelismFactor, p.KeyLength)
	if err != nil && !reset {
		slog.InfoContext(ctx, fmt.Sprintf("failed login for %q: %v", email, err))
		return LoginResult{}, ErrInvalidCredentials
	}
	if reset {
		token := uuid.NewString()
		if err = s.tokens.Set(ctx, token, u.ID, loginResetTTL); err != nil {
			return LoginResult{}, err
		}
		return LoginResult{ResetRequired: true, ResetURL: "/reset/" + token}, nil
	}
	return LoginResult{User: u}, nil
}

func checkNew(newPassword, confirm string) error {
	f := svcerr.Fields{}
	if newPassword != confirm {
		f.Add("confirmationPassword", "passwords do not match")
	}
	if msg := PasswordProblems(newPassword); msg != "" {
		f.Add("newPassword", msg)
	}
	return f.Err()
}

func (s *Service) setPassword(ctx context.Context, u user.User, newPassword string) error {
	p := s.passwords
	u.Password = null.StringFrom(newPassword)
	if err := s.users.EditUserPassword(ctx, u, p.ScryptWorkFactor, p.ScryptBlockSize, p.ScryptParallelismFactor, p.KeyLength); err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}
	return nil
}

// ChangePassword changes the logged-in user's password after re-checking the old one.
func (s *Service) ChangePassword(ctx context.Context, u user.User, oldPassword, newPassword, confirm string) error {
	ctx, span := tracer.Start(ctx, "auth.Service.ChangePassword")
	defer span.End()
	p := s.passwords
	check := u
	check.Password = null.StringFrom(oldPassword)
	if _, _, err := s.users.VerifyUser(ctx, check, p.Iterations, p.ScryptWorkFactor, p.ScryptBlockSize, p.ScryptParallelismFactor, p.KeyLength); err != nil {
		return svcerr.InvalidField("oldPassword", "old password is not correct")
	}
	if err := checkNew(newPassword, confirm); err != nil {
		return err
	}
	return s.setPassword(ctx, u, newPassword)
}

func (s *Service) tokenUser(ctx context.Context, token string) (user.User, error) {
	id, ok := s.tokens.Get(ctx, token)
	if !ok {
		return user.User{}, svcerr.NotFound("reset link is invalid or has expired", nil)
	}
	u, err := s.users.GetUser(ctx, user.User{ID: id})
	if err != nil || u.ID != id {
		s.tokens.Delete(ctx, token)
		return user.User{}, svcerr.NotFound("reset link is invalid or has expired", err)
	}
	return u, nil
}

// CheckResetToken reports NotFound for an unknown or expired token.
func (s *Service) CheckResetToken(ctx context.Context, token string) error {
	ctx, span := tracer.Start(ctx, "auth.Service.CheckResetToken")
	defer span.End()
	_, err := s.tokenUser(ctx, token)
	return err
}

// ResetPassword sets a new password using a single-use reset token.
func (s *Service) ResetPassword(ctx context.Context, token, newPassword, confirm string) error {
	ctx, span := tracer.Start(ctx, "auth.Service.ResetPassword")
	defer span.End()
	u, err := s.tokenUser(ctx, token)
	if err != nil {
		return err
	}
	if err = checkNew(newPassword, confirm); err != nil {
		return err
	}
	if err = s.setPassword(ctx, u, newPassword); err != nil {
		return err
	}
	s.tokens.Delete(ctx, token)
	return nil
}
```
`EditUserPassword` already clears `reset_password`. Run: `go test ./server/internal/auth/`. Expected: the service tests PASS. The Task 5 tests don't compile yet, because `NewHandlers` changes in the next step.

- [ ] **Step 3: Failing handler tests, then the new routes**

In `server/internal/auth/sessions_test.go`, change `newMeAPI` to:
```go
func newMeAPI(s *auth.Sessions) *apitest.Client {
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	auth.NewHandlers(s, nil, authtest.Files()).Register(api, s.Guards())
	return apitest.New(e)
}
```

`server/internal/auth/handlers_test.go`:
```go
package auth_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestAuthRoutes(t *testing.T) {
	svc, users, tokens := newAuthService()
	sessions := authtest.New(users.byEmail["ok@example.test"], users.byEmail["reset@example.test"])
	e := apitest.NewEcho()
	auth.NewHandlers(sessions, svc, authtest.Files()).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)

	rec := anon.JSON(t, http.MethodPost, "/api/v1/auth/login", auth.LoginInput{Email: "ok@example.test", Password: "wrong"})
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "invalid email or password", apitest.ErrorOf(t, rec).Message)

	rec = anon.JSON(t, http.MethodPost, "/api/v1/auth/login", auth.LoginInput{Email: "reset@example.test", Password: "Right-Pa55"})
	require.Equal(t, http.StatusOK, rec.Code)
	res := apitest.Decode[auth.LoginResponse](t, rec)
	assert.True(t, res.ResetRequired)
	assert.Nil(t, res.User)
	assert.Empty(t, rec.Result().Cookies(), "no session for a reset-required login")

	rec = anon.JSON(t, http.MethodPost, "/api/v1/auth/login", auth.LoginInput{Email: "ok@example.test", Password: "Right-Pa55", Remember: true})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.NotNil(t, apitest.Decode[auth.LoginResponse](t, rec).User)
	var session *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessions.Name() {
			session = c
		}
	}
	require.NotNil(t, session)
	me := anon.As(session)
	assert.Equal(t, http.StatusOK, me.Get(t, "/api/v1/auth/me").Code)

	rec = me.JSON(t, http.MethodPost, "/api/v1/auth/password",
		auth.PasswordInput{OldPassword: "Right-Pa55", NewPassword: "weak", ConfirmationPassword: "weak"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	rec = me.JSON(t, http.MethodPost, "/api/v1/auth/password",
		auth.PasswordInput{OldPassword: "Right-Pa55", NewPassword: "Str0ng-Pass!", ConfirmationPassword: "Str0ng-Pass!"})
	assert.Equal(t, http.StatusNoContent, rec.Code)

	rec = me.JSON(t, http.MethodPost, "/api/v1/auth/logout", nil)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	require.NoError(t, tokens.Set(ctx, "tok", 2, time.Hour))
	assert.Equal(t, http.StatusNoContent, anon.Get(t, "/api/v1/auth/reset/tok").Code)
	assert.Equal(t, http.StatusNotFound, anon.Get(t, "/api/v1/auth/reset/bad").Code)
	rec = anon.JSON(t, http.MethodPost, "/api/v1/auth/reset/tok",
		auth.ResetInput{NewPassword: "Str0ng-Pass!", ConfirmationPassword: "Str0ng-Pass!"})
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, http.StatusNotFound, anon.Get(t, "/api/v1/auth/reset/tok").Code, "single use")
}
```

Add to `server/internal/auth/types.go`:
```go
// LoginInput is the body of POST /auth/login.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

// LoginResponse is either the logged-in user or a reset instruction.
type LoginResponse struct {
	User          *CurrentUser `json:"user,omitempty"`
	ResetRequired bool         `json:"resetRequired"`
	ResetURL      string       `json:"resetUrl,omitempty"`
}

// PasswordInput is the body of POST /auth/password.
type PasswordInput struct {
	OldPassword          string `json:"oldPassword"`
	NewPassword          string `json:"newPassword"`
	ConfirmationPassword string `json:"confirmationPassword"`
}

// ResetInput is the body of POST /auth/reset/{token}.
type ResetInput struct {
	NewPassword          string `json:"newPassword"`
	ConfirmationPassword string `json:"confirmationPassword"`
}
```

Replace `server/internal/auth/handlers.go` with:
```go
package auth

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves the /auth endpoints.
type Handlers struct {
	sessions *Sessions
	svc      *Service
	files    *upload.Files
}

func NewHandlers(sessions *Sessions, svc *Service, files *upload.Files) *Handlers {
	return &Handlers{sessions: sessions, svc: svc, files: files}
}

// Register mounts the /auth routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/auth/me", h.me, guards.Login)
	g.POST("/auth/login", h.login)
	g.POST("/auth/logout", h.logout, guards.Login)
	g.POST("/auth/password", h.password, guards.Login)
	g.GET("/auth/reset/:token", h.checkReset)
	g.POST("/auth/reset/:token", h.reset)
}

// me returns the logged-in user.
//
//	@Summary	Current user
//	@Tags		auth
//	@Produce	json
//	@Success	200	{object}	auth.CurrentUser
//	@Failure	401	{object}	web.ErrorResponse
//	@Router		/auth/me [get]
func (h *Handlers) me(c echo.Context) error {
	u, _ := Current(c)
	return c.JSON(http.StatusOK, NewCurrentUser(u, h.files))
}

// login starts a session, or returns a reset link for reset-flagged accounts.
//
//	@Summary	Log in
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		auth.LoginInput	true	"Credentials"
//	@Success	200		{object}	auth.LoginResponse
//	@Failure	401		{object}	web.ErrorResponse
//	@Router		/auth/login [post]
func (h *Handlers) login(c echo.Context) error {
	var in LoginInput
	if err := web.BindJSON(c, &in); err != nil {
		return err
	}
	res, err := h.svc.Login(c.Request().Context(), in.Email, in.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		return echo.NewHTTPError(http.StatusUnauthorized, ErrInvalidCredentials.Error())
	}
	if err != nil {
		return err
	}
	if res.ResetRequired {
		return c.JSON(http.StatusOK, LoginResponse{ResetRequired: true, ResetURL: res.ResetURL})
	}
	if err = h.sessions.Login(c.Response(), c.Request(), res.User, in.Remember); err != nil {
		return err
	}
	me := NewCurrentUser(res.User, h.files)
	return c.JSON(http.StatusOK, LoginResponse{User: &me})
}

// logout ends the session.
//
//	@Summary	Log out
//	@Tags		auth
//	@Success	204
//	@Router		/auth/logout [post]
func (h *Handlers) logout(c echo.Context) error {
	if err := h.sessions.Logout(c.Response(), c.Request()); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// password changes the logged-in user's password.
//
//	@Summary	Change password
//	@Tags		auth
//	@Accept		json
//	@Param		body	body	auth.PasswordInput	true	"Old and new passwords"
//	@Success	204
//	@Failure	422	{object}	web.ErrorResponse
//	@Router		/auth/password [post]
func (h *Handlers) password(c echo.Context) error {
	var in PasswordInput
	if err := web.BindJSON(c, &in); err != nil {
		return err
	}
	u, _ := Current(c)
	if err := h.svc.ChangePassword(c.Request().Context(), u, in.OldPassword, in.NewPassword, in.ConfirmationPassword); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// checkReset reports whether a reset link is still valid.
//
//	@Summary	Check a reset link
//	@Tags		auth
//	@Param		token	path	string	true	"Reset token"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/auth/reset/{token} [get]
func (h *Handlers) checkReset(c echo.Context) error {
	if err := h.svc.CheckResetToken(c.Request().Context(), c.Param("token")); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// reset sets a new password with a reset link.
//
//	@Summary	Reset password
//	@Tags		auth
//	@Accept		json
//	@Param		token	path	string			true	"Reset token"
//	@Param		body	body	auth.ResetInput	true	"New password"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Failure	422	{object}	web.ErrorResponse
//	@Router		/auth/reset/{token} [post]
func (h *Handlers) reset(c echo.Context) error {
	var in ResetInput
	if err := web.BindJSON(c, &in); err != nil {
		return err
	}
	if err := h.svc.ResetPassword(c.Request().Context(), c.Param("token"), in.NewPassword, in.ConfirmationPassword); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
```
Run: `go test ./server/internal/auth/`. Expected: PASS.

- [ ] **Step 4: `account` package**

`server/internal/account/service_test.go`:
```go
package account_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/account"
	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

type fakeStore struct{ rows map[int]user.User }

func (f *fakeStore) EditUserImage(_ context.Context, u user.User) error {
	if _, ok := f.rows[u.ID]; !ok {
		return fmt.Errorf("failed to get user: %w", sql.ErrNoRows)
	}
	row := f.rows[u.ID]
	row.FileName = u.FileName
	f.rows[u.ID] = row
	return nil
}

func TestAccountImage(t *testing.T) {
	me := authtest.Treasurer
	me.FileName = null.StringFrom("user/old.png")
	store := &fakeStore{rows: map[int]user.User{me.ID: me}}
	objects := uploadtest.New()
	objects.Objects["user/old.png"] = "OLD"
	files := upload.New(objects)
	sessions := authtest.New(me)

	e := apitest.NewEcho()
	account.NewHandlers(account.NewService(store, files), files).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	client := anon.As(authtest.Cookie(t, sessions, me))

	assert.Equal(t, http.StatusUnauthorized, anon.Get(t, "/api/v1/account").Code)
	rec := client.Get(t, "/api/v1/account")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "https://cdn.test/user/old.png", apitest.Decode[auth.CurrentUser](t, rec).ImageURL)

	rec = client.Multipart(t, http.MethodPut, "/api/v1/account/image", nil,
		apitest.FilePart{Field: "image", Name: "me.png", ContentType: "image/png", Body: "NEW"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	newKey := store.rows[me.ID].FileName.String
	assert.NotEqual(t, "user/old.png", newKey)
	assert.Equal(t, []string{"user/old.png"}, objects.Deleted, "old image removed only after the new one is saved")

	rec = client.JSON(t, http.MethodDelete, "/api/v1/account/image", nil)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.False(t, store.rows[me.ID].FileName.Valid)
	assert.Contains(t, objects.Deleted, newKey)

	rec = client.Multipart(t, http.MethodPut, "/api/v1/account/image", nil)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "image is required")
}
```
The `authtest` fake user getter returns the persona as registered, so the `GET /account` image comes from `me.FileName`, which the test set before `authtest.New(me)`.

`server/internal/account/service.go`:
```go
// Package account lets a logged-in user manage their own profile photo.
package account

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

var tracer = otel.Tracer("github.com/COMTOP1/AFC-GO/server/internal/account")

// Store updates a user's photo (satisfied by *user.Store).
type Store interface {
	EditUserImage(ctx context.Context, userParam user.User) error
}

// Service changes the current user's photo.
type Service struct {
	store Store
	files *upload.Files
}

func NewService(store Store, files *upload.Files) *Service {
	return &Service{store: store, files: files}
}

// SetImage replaces u's photo; the old object is deleted after the DB write.
func (s *Service) SetImage(ctx context.Context, u user.User, image *upload.File) (user.User, error) {
	ctx, span := tracer.Start(ctx, "account.Service.SetImage")
	defer span.End()
	if image == nil {
		return user.User{}, svcerr.InvalidField("image", "image is required")
	}
	key, err := s.files.Save(ctx, image, "user")
	if err != nil {
		return user.User{}, err
	}
	old := u.FileName.String
	u.FileName = null.StringFrom(key)
	if err = s.store.EditUserImage(ctx, u); err != nil {
		s.files.Remove(ctx, key)
		return user.User{}, fmt.Errorf("failed to update image: %w", err)
	}
	s.files.Remove(ctx, old)
	return u, nil
}

// RemoveImage clears u's photo.
func (s *Service) RemoveImage(ctx context.Context, u user.User) (user.User, error) {
	ctx, span := tracer.Start(ctx, "account.Service.RemoveImage")
	defer span.End()
	old := u.FileName.String
	u.FileName = null.String{}
	if err := s.store.EditUserImage(ctx, u); err != nil {
		return user.User{}, fmt.Errorf("failed to remove image: %w", err)
	}
	s.files.Remove(ctx, old)
	return u, nil
}
```

`server/internal/account/handlers.go`:
```go
package account

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /account.
type Handlers struct {
	svc   *Service
	files *upload.Files
}

func NewHandlers(svc *Service, files *upload.Files) *Handlers {
	return &Handlers{svc: svc, files: files}
}

// Register mounts the account routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/account", h.get, guards.Login)
	g.PUT("/account/image", h.setImage, guards.Login)
	g.DELETE("/account/image", h.removeImage, guards.Login)
}

// get returns the logged-in user's account.
//
//	@Summary	My account
//	@Tags		account
//	@Produce	json
//	@Success	200	{object}	auth.CurrentUser
//	@Failure	401	{object}	web.ErrorResponse
//	@Router		/account [get]
func (h *Handlers) get(c echo.Context) error {
	u, _ := auth.Current(c)
	return c.JSON(http.StatusOK, auth.NewCurrentUser(u, h.files))
}

// setImage replaces my photo.
//
//	@Summary	Set my photo
//	@Tags		account
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		image	formData	file	true	"Photo"
//	@Success	200		{object}	auth.CurrentUser
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/account/image [put]
func (h *Handlers) setImage(c echo.Context) error {
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	u, _ := auth.Current(c)
	updated, err := h.svc.SetImage(c.Request().Context(), u, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, auth.NewCurrentUser(updated, h.files))
}

// removeImage clears my photo.
//
//	@Summary	Remove my photo
//	@Tags		account
//	@Produce	json
//	@Success	200	{object}	auth.CurrentUser
//	@Router		/account/image [delete]
func (h *Handlers) removeImage(c echo.Context) error {
	u, _ := auth.Current(c)
	updated, err := h.svc.RemoveImage(c.Request().Context(), u)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, auth.NewCurrentUser(updated, h.files))
}
```
Run: `go test ./server/internal/account/ ./server/internal/auth/`. Expected: PASS.

- [ ] **Step 5: Legacy login, password, reset and account → services**

Add `authSvc *auth.Service` and `accountSvc *account.Service` to `Views`, `AuthService`/`AccountService` to `Deps`, and the assignments in `New`.

In `server/internal/legacy/views/login.go`, keep the session-writing code, the flash messages and the JSON shape. Replace only the credential check and the token issue. The body of the `if c.Request().Method == http.MethodPost` branch becomes:
```go
		res, err := v.authSvc.Login(c.Request().Context(), c.FormValue("email"), c.FormValue("password"))
		if err != nil {
			if saveErr := session.Save(c.Request(), c.Response()); saveErr != nil {
				return fmt.Errorf("failed to save session for login: %w", saveErr)
			}
			ctx := v.getSessionData(c)
			ctx.Message = "Invalid email or password"
			ctx.MsgType = "is-danger"
			if err = v.setMessagesInSession(c, ctx); err != nil {
				return fmt.Errorf("failed to set message for login: %w", err)
			}
			return c.JSON(http.StatusOK, struct {
				Error         string `json:"error"`
				ResetPassword bool   `json:"resetPassword"`
			}{Error: "Invalid email or password"})
		}
		if res.ResetRequired {
			ctx := v.getSessionData(c)
			ctx.Message = "Password reset required"
			ctx.MsgType = "is-danger"
			if err = v.setMessagesInSession(c, ctx); err != nil {
				return fmt.Errorf("failed to set message for login: %w", err)
			}
			return c.JSON(http.StatusOK, struct {
				Error         string `json:"error"`
				ResetPassword bool   `json:"resetPassword"`
				URL           string `json:"url"`
			}{ResetPassword: true, URL: res.ResetURL})
		}
		u := res.User
		u.Authenticated = true
		// … the existing lines from `err = v.clearMessagesInSession(c)` to the
		//   final `return c.JSON(http.StatusOK, data)` stay exactly as they are,
		//   including the legacy `remember != "on"` cookie lifetime …
```
Remove the now-unused `uuid`, `null` and `time` imports if goimports reports them.

In `server/internal/legacy/views/changePassword.go`, the POST branch becomes:
```go
		c1 := v.getSessionData(c)
		data := struct {
			Error string `json:"error"`
		}{}
		if err := v.authSvc.ChangePassword(c.Request().Context(), c1.User,
			c.FormValue("oldPassword"), c.FormValue("newPassword"), c.FormValue("confirmationPassword")); err != nil {
			data.Error = err.Error()
			return c.JSON(http.StatusOK, data)
		}
		v.flash(c, c1, "successfully changed password")
		return c.JSON(http.StatusOK, data)
```

In `server/internal/legacy/views/reset.go`, `ResetURLFunc` becomes:
```go
func (v *Views) ResetURLFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ResetURLFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	c1 := v.getSessionData(c)
	token := c.Param("url")
	if err := v.authSvc.CheckResetToken(c.Request().Context(), token); err != nil {
		return v.error(http.StatusBadRequest, "failed to get url for reset", err)
	}
	switch c.Request().Method {
	case http.MethodGet:
		data := struct {
			Year    int
			Context *Context
			User    user.User
			URL     string
		}{Year: time.Now().Year(), Context: c1, URL: token}
		return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.ResetTemplate, templates.NoNavType)
	case http.MethodPost:
		data := struct {
			Error string `json:"error"`
		}{}
		if err := v.authSvc.ResetPassword(c.Request().Context(), token,
			c.FormValue("newPassword"), c.FormValue("confirmationPassword")); err != nil {
			data.Error = err.Error()
			return c.JSON(http.StatusOK, data)
		}
		if err := v.clearMessagesInSession(c); err != nil {
			slog.Info(fmt.Sprintf("failed to clear messages for reset: %+v", err))
		}
		v.flash(c, c1, "successfully reset password")
		return c.JSON(http.StatusOK, data)
	default:
		return v.invalidMethodUsed(c)
	}
}
```
Before replacing it, check the old GET branch's `data` struct: if it passed `VisitorCount`, keep that field in the new struct (`VisitorCount: v.GetVisitorCount()`).

In `server/internal/legacy/views/account.go`, keep `AccountFunc`. Replace the bodies of `UploadImageFunc` and `RemoveImageFunc` with calls to `v.accountSvc.SetImage(ctx, c1.User, image)` (image via `legacyUpload(c, "upload")`; a nil image → `data.Error = "failed to get file for upload image"`) and `v.accountSvc.RemoveImage(ctx, c1.User)`. On success, flash `"successfully uploaded image"` / `"successfully removed image"` and return `c.JSON(http.StatusOK, data)`, keeping each handler's existing `data` struct.

Delete `minRequirementsMet` from `helpers.go`; it is now `auth.PasswordProblems`.

- [ ] **Step 6: Wire, extend guarded GETs, regenerate, run, commit**

In `app.Build`:
- `authSvc := auth.NewService(s.User, tokens, conf.Passwords)`
- `accountSvc := account.NewService(s.User, uploads)`
- add `AuthService: authSvc, AccountService: accountSvc,` to `views.Deps`
- change the auth handler line to `auth.NewHandlers(sessions, authSvc, uploads).Register(api, guards)`
- add `account.NewHandlers(accountSvc, uploads).Register(api, guards)`

In `server/internal/app/helpers_test.go` append `"/api/v1/account",` to `guardedGets`.
```bash
go generate ./server/internal/docs
goimports -w server/internal/legacy/views
gofmt -l server; go build ./... && go vet ./... && go test ./... && go tool golangci-lint run ./...
git add -A server
git commit -m "Add auth (login/logout/password/reset) and account APIs; legacy flows use the services

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```
`TestUnsafeAPIRoutesRequireLogin` now exercises its `publicWrites` allow-list (`/auth/login`, `/auth/reset/:token`).

---

### Task 23: Clean-up and route coverage check

**Files:**
- Modify: `server/internal/legacy/views/views.go`, `helpers.go` (remove dead fields/helpers)
- Create: `server/internal/app/coverage_test.go`

- [ ] **Step 1: Remove dead legacy plumbing**

After Tasks 8–22 no legacy view writes through a store or uploads directly. Delete `fileUpload` from `helpers.go`. Then run:
```bash
go vet ./server/internal/legacy/... && go tool golangci-lint run ./server/internal/legacy/...
grep -n "v\.fileUpload\|v\.mailer\|v\.storage\.Put\|v\.storage\.Delete" -r server/internal/legacy/views || echo "none left"
```
Expected: `none left`. Remove the `mailer` field from `Views` and `Mailer` from `Deps` (and its line in `app.Build`); keep `storage`, because templates still need `PublicURL`. `unused` lint findings for other now-dead helpers (e.g. `DBUserToTemplateFormat` if unused) are removed the same way. Anything still referenced stays.

- [ ] **Step 2: Route coverage test**

`server/internal/app/coverage_test.go`:
```go
package app_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEveryLegacyActionHasAnAPIRoute is the spec's cleanup checklist: every
// action the template site offers has an /api/v1 equivalent (the two
// form-redirect helpers, programmeselect and whatsonselect, are replaced by
// query parameters and intentionally have none).
func TestEveryLegacyActionHasAnAPIRoute(t *testing.T) {
	want := []string{
		"GET /api/v1/site", "GET /api/v1/home", "GET /api/v1/contact",
		"POST /api/v1/auth/login", "POST /api/v1/auth/logout", "GET /api/v1/auth/me",
		"POST /api/v1/auth/password", "GET /api/v1/auth/reset/:token", "POST /api/v1/auth/reset/:token",
		"GET /api/v1/account", "PUT /api/v1/account/image", "DELETE /api/v1/account/image",
		"GET /api/v1/news", "POST /api/v1/news", "GET /api/v1/news/:id", "PATCH /api/v1/news/:id", "DELETE /api/v1/news/:id",
		"GET /api/v1/whatson", "POST /api/v1/whatson", "GET /api/v1/whatson/:id", "PATCH /api/v1/whatson/:id", "DELETE /api/v1/whatson/:id",
		"GET /api/v1/teams", "POST /api/v1/teams", "GET /api/v1/teams/:id", "PATCH /api/v1/teams/:id", "DELETE /api/v1/teams/:id",
		"GET /api/v1/players", "POST /api/v1/players", "PATCH /api/v1/players/:id", "DELETE /api/v1/players/:id",
		"GET /api/v1/programmes", "POST /api/v1/programmes", "DELETE /api/v1/programmes/:id",
		"GET /api/v1/seasons", "POST /api/v1/seasons", "PATCH /api/v1/seasons/:id", "DELETE /api/v1/seasons/:id",
		"GET /api/v1/sponsors", "POST /api/v1/sponsors", "DELETE /api/v1/sponsors/:id",
		"GET /api/v1/affiliations", "POST /api/v1/affiliations", "DELETE /api/v1/affiliations/:id",
		"GET /api/v1/documents", "POST /api/v1/documents", "DELETE /api/v1/documents/:id",
		"GET /api/v1/gallery", "POST /api/v1/gallery", "DELETE /api/v1/gallery/:id",
		"GET /api/v1/info", "PUT /api/v1/info",
		"GET /api/v1/users", "POST /api/v1/users", "GET /api/v1/users/:id", "PATCH /api/v1/users/:id",
		"DELETE /api/v1/users/:id", "POST /api/v1/users/:id/reset",
		"PUT /api/v1/settings/display-email",
		"GET /api/v1/files/:kind/:id",
		"GET /api/v1/health", "GET /api/health",
	}
	have := map[string]bool{}
	for _, r := range bareApp(t).Echo.Routes() {
		have[r.Method+" "+r.Path] = true
	}
	for _, route := range want {
		assert.True(t, have[route], "missing %s", route)
	}
}
```
Run: `go test ./server/internal/app/`. Expected: PASS.

- [ ] **Step 3: Full verification**

```bash
gofmt -l server; go build ./... && go vet ./...
go test ./...                                   # with AFC_TEST_DB exported
go tool golangci-lint run ./...
go install go.uber.org/nilaway/cmd/nilaway@latest && nilaway ./server/cmd/afc
go generate ./server/internal/docs && git diff --exit-code server/internal/docs   # docs up to date
docker build -t afc-api-split .                # Dockerfile still builds
```
Then run the server locally and click through the legacy site. Log in, add and delete a news article with an image, edit a team, and open `/api/v1/swagger/index.html`.

- [ ] **Step 4: Commit**

```bash
git add -A server
git commit -m "Remove dead legacy plumbing; add API route coverage check

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```
