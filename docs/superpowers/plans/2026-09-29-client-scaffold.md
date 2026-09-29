# Client Scaffold & Build Pipeline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a React/Vite/TypeScript client under `client/`, embed its build in the Go binary, and serve it at `/app/*`. Include a Go→Vite dev proxy, a typed `/api/v1` data layer with a proof page, and Docker/CI support.

**Architecture:** `web.MountSPA` serves `/app/*` from an embedded `fs.FS` (index fallback for client routes, immutable hashed assets, a 503 "not built" page) or reverse-proxies to Vite when `AFC_UI_PROXY_URL` is set. The client is plain React 19 + React Router (`basename="/app"`) + TanStack Query, with a `fetch` wrapper that understands the API's error envelope. Yarn 4 scripts mirror MV-Controller: build the client, copy it into `server/cmd/afc/ui`, `go build`.

**Tech Stack:** Go 1.26 / Echo v4; Node 24, Yarn 4 (corepack), Vite, React 19, React Router 7, TanStack Query 5, TypeScript, ESLint + Prettier, Vitest + Testing Library + jsdom.

**Spec:** `docs/superpowers/specs/2026-09-29-client-scaffold-design.md`. Read it before starting any task.

**Working directory:** `/Users/liam/Code/Go/AFC-client-scaffold` (git worktree, branch `client-scaffold`, based on `main` after PR #11). Every command runs from there.

## Global Constraints

- The SPA is served only under `/app` and `/app/*`. Legacy routes, `/api/*` and `/public/*` behave exactly as before.
- Vite `base: '/app/'`; `BrowserRouter basename="/app"`; Vite output in `build/client`; embed target `server/cmd/afc/ui` via `//go:embed all:ui`; only `server/cmd/afc/ui/.keep` is committed.
- Cache headers:
  - `index.html`: `Cache-Control: no-cache`.
  - `/app/assets/*`: `Cache-Control: public, max-age=31536000, immutable`.
  - Other root files with an extension: `no-cache`.
- A missing `/app/assets/*` file is a `404` (never the index fallback). With no built `index.html`, `/app*` returns `503` with the text "The web client has not been built. Run `yarn build:client`." Methods other than GET/HEAD under `/app` return `405`.
- Dev proxy: env `AFC_UI_PROXY_URL` (e.g. `http://localhost:5173`); an invalid value is fatal at startup.
- API calls: base `/api/v1`, `credentials: 'same-origin'`, no CSRF token handling (browsers send `Sec-Fetch-Site`). Errors surface as `ApiError {status, message, fields}`; a network failure is `status 0`.
- Query retry: never for `ApiError` with 400 ≤ status < 500; otherwise up to 2 retries. `staleTime` 30 000 ms. A `401` from `/auth/me` means anonymous (`null`), not an error.
- No UI library, CSS framework or styling decisions (sub-project 3).
- Toolchain: Yarn 4 via corepack (`packageManager` field), `nodeLinker: node-modules`; Prettier `printWidth 100, singleQuote, semi, tabWidth 2`.
- Go: `gofmt`, `go vet`, `go test ./...` (with `AFC_TEST_DB` for DB tests) and `golangci-lint run ./...` stay green after every task. Client: `yarn lint:client`, `yarn typecheck` and `yarn test:client` stay green from Task 3 on.
- Commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Plan Decisions (beyond the spec)

1. **`MountSPA` takes a parsed `*url.URL`** for the proxy, and `main.go` validates `AFC_UI_PROXY_URL`. The proxy uses `httputil.ReverseProxy.Rewrite` with `SetURL`, so the `Host` header matches Vite (Vite rejects unknown hosts).
2. **`app.Config.UI` may be nil** (tests, `bareApp`). That is treated as "not built", giving a 503.
3. **Hand-written TS types** mirror the Go JSON tags for `site.Info`, `team.Public` and `auth.CurrentUser` only.
4. **`useAuth` lives in its own file.** It is split from `AuthProvider` so `react-refresh/only-export-components` stays quiet.
5. **`build:emails` regenerated output is committed only if `go test ./server/internal/emails/` passes** on it. Otherwise the script is kept, the templates are reverted and the difference is reported.
6. **Test DB:** there is no Docker on the dev machine. Use the local throwaway Postgres: `AFC_TEST_DB='postgres://postgres@localhost:55432/postgres?sslmode=disable'` (never port 5432).

## Review Focus

1. **Refreshing a deep client URL** (`/app/news/5`) must return the SPA, not the legacy 404 page. Pinned in Task 1 (`TestSPAIndexFallback`) and Task 2 (`TestAppServesSPAFallback`).
2. **After a deploy, browsers must not keep an old `index.html`** that points at deleted hashed assets. Pinned in Task 1 (`TestSPACacheHeaders`).
3. **A server built without the client** (Go-only build, CI, or a forgotten `yarn build:client`) must say so clearly, not show a blank page or a legacy 404. Pinned in Task 1 (`TestSPANotBuilt`) and Task 2 (`TestAppSPANotBuiltByDefault`).
4. **An API response that isn't JSON** (an HTML error page, a proxy error, an empty 200) must become a readable `ApiError`, not a JSON parse crash. Pinned in Task 4 (`non-JSON error body`, `empty 200 body`).
5. **Path traversal under `/app`** (`/app/../server/...`, encoded `..`) must never read outside the embedded `ui` files. Pinned in Task 1 (`TestSPARejectsTraversal`).

## File Structure

```
.dockerignore, .gitignore(mod), Dockerfile(mod), README.md(mod)
.github/workflows/client.yaml
package.json, yarn.lock, .yarnrc.yml, vite.config.ts, tsconfig.json, tsconfig.app.json, tsconfig.node.json,
eslint.config.js, .prettierrc, .prettierignore
scripts/{clean,build-client,build-server,dev-server,build-emails}.js
client/
  index.html, main.tsx, App.tsx, App.test.tsx, vite-env.d.ts
  public/favicon.svg
  api/{client.ts,client.test.ts,types.ts,queryClient.ts,queries.ts,queries.test.tsx}
  auth/{context.ts,AuthProvider.tsx,useAuth.ts}
  components/Layout.tsx
  pages/{HomePage.tsx,HomePage.test.tsx,NotFoundPage.tsx}
  test/{setup.ts,render.tsx,mockFetch.ts}
server/cmd/afc/{ui.go, ui/.keep, main.go(mod)}
server/internal/web/{spa.go, spa_test.go}
server/internal/app/{app.go(mod), spa_test.go}
```

---

### Task 1: `web.MountSPA`

**Files:**
- Create: `server/internal/web/spa.go`, `server/internal/web/spa_test.go`

**Interfaces:**
- Produces: `web.MountSPA(e *echo.Echo, files fs.FS, proxy *url.URL)` and the constant `web.SPAPrefix = "/app"`. `files` is the root of the built client (containing `index.html` and `assets/`) and may be nil; `proxy` may be nil.

- [ ] **Step 1: Write the failing tests**

`server/internal/web/spa_test.go`:
```go
package web_test

import (
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

const indexHTML = `<!doctype html><html><body><div id="root"></div><script type="module" src="/app/assets/index-abc123.js"></script></body></html>`

func builtFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html":             {Data: []byte(indexHTML)},
		"assets/index-abc123.js": {Data: []byte("console.log('app')")},
		"assets/style-9f8e.css":  {Data: []byte("body{}")},
		"favicon.svg":            {Data: []byte("<svg/>")},
		".keep":                  {Data: nil},
	}
}

func spaEcho(files fs.FS, proxy *url.URL) *echo.Echo {
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	e.RouteNotFound("/*", func(c echo.Context) error { return c.HTML(http.StatusNotFound, "legacy 404") })
	web.MountSPA(e, files, proxy)
	return e
}

func get(e *echo.Echo, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestSPAIndexFallback(t *testing.T) {
	e := spaEcho(builtFS(), nil)
	for _, path := range []string{"/app", "/app/", "/app/news", "/app/news/5", "/app/teams/2/players"} {
		rec := get(e, http.MethodGet, path)
		assert.Equal(t, http.StatusOK, rec.Code, path)
		assert.Equal(t, indexHTML, rec.Body.String(), path)
		assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "text/html", path)
	}
}

func TestSPAAssets(t *testing.T) {
	e := spaEcho(builtFS(), nil)

	rec := get(e, http.MethodGet, "/app/assets/index-abc123.js")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "console.log('app')", rec.Body.String())
	assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "javascript")

	rec = get(e, http.MethodGet, "/app/assets/style-9f8e.css")
	assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "text/css")

	for _, missing := range []string{"/app/assets/index-old999.js", "/app/assets", "/app/favicon.png"} {
		rec = get(e, http.MethodGet, missing)
		assert.Equal(t, http.StatusNotFound, rec.Code, missing)
		assert.NotContains(t, rec.Body.String(), `<div id="root">`, "must never fall back to index: %s", missing)
	}

	rec = get(e, http.MethodGet, "/app/favicon.svg")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "<svg/>", rec.Body.String())
}

func TestSPACacheHeaders(t *testing.T) {
	e := spaEcho(builtFS(), nil)
	assert.Equal(t, "no-cache", get(e, http.MethodGet, "/app/news/5").Header().Get("Cache-Control"))
	assert.Equal(t, "public, max-age=31536000, immutable",
		get(e, http.MethodGet, "/app/assets/index-abc123.js").Header().Get("Cache-Control"))
	assert.Equal(t, "no-cache", get(e, http.MethodGet, "/app/favicon.svg").Header().Get("Cache-Control"))
}

func TestSPANotBuilt(t *testing.T) {
	for name, files := range map[string]fs.FS{
		"only .keep": fstest.MapFS{".keep": {Data: nil}},
		"nil fs":     nil, // a true nil interface, as app.Config.UI is in tests
	} {
		t.Run(name, func(t *testing.T) {
			e := spaEcho(files, nil)
			for _, path := range []string{"/app", "/app/news/5", "/app/assets/index-abc123.js"} {
				rec := get(e, http.MethodGet, path)
				assert.Equal(t, http.StatusServiceUnavailable, rec.Code, path)
				assert.Contains(t, rec.Body.String(), "The web client has not been built", path)
				assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"), path)
			}
		})
	}
}

func TestSPAMethods(t *testing.T) {
	e := spaEcho(builtFS(), nil)
	assert.Equal(t, http.StatusMethodNotAllowed, get(e, http.MethodPost, "/app/news").Code)
	assert.Equal(t, http.StatusMethodNotAllowed, get(e, http.MethodDelete, "/app").Code)
	head := get(e, http.MethodHead, "/app/news")
	assert.Equal(t, http.StatusOK, head.Code)
}

func TestSPARejectsTraversal(t *testing.T) {
	e := spaEcho(builtFS(), nil)
	for _, path := range []string{"/app/../go.mod", "/app/assets/../../go.mod", "/app/%2e%2e/go.mod", "/app/assets/%2e%2e/index.html"} {
		rec := get(e, http.MethodGet, path)
		assert.NotContains(t, rec.Body.String(), "module github.com", path)
		assert.NotEqual(t, "console.log('app')", rec.Body.String(), path)
	}
}

func TestSPADoesNotShadowOtherRoutes(t *testing.T) {
	e := spaEcho(builtFS(), nil)
	e.GET("/news", func(c echo.Context) error { return c.String(http.StatusOK, "legacy news") })
	assert.Equal(t, "legacy news", get(e, http.MethodGet, "/news").Body.String())
	assert.Equal(t, "legacy 404", get(e, http.MethodGet, "/application").Body.String())
	assert.Equal(t, "legacy 404", get(e, http.MethodGet, "/apps/x").Body.String())
}

func TestSPAProxy(t *testing.T) {
	var gotPath, gotMethod, gotHost, gotBody string
	vite := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod, gotHost = r.URL.Path, r.Method, r.Host
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = w.Write([]byte("from vite"))
	}))
	defer vite.Close()
	target, err := url.Parse(vite.URL)
	require.NoError(t, err)

	e := spaEcho(builtFS(), target)
	rec := get(e, http.MethodGet, "/app/src/main.tsx")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "from vite", rec.Body.String())
	assert.Equal(t, "/app/src/main.tsx", gotPath)
	assert.Equal(t, target.Host, gotHost, "Host must match Vite so its host check passes")

	req := httptest.NewRequest(http.MethodPost, "/app/@vite/ping", strings.NewReader("payload"))
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.MethodPost, gotMethod, "proxy mode forwards every method")
	assert.Equal(t, "payload", gotBody)

	rec = get(e, http.MethodGet, "/app")
	assert.Equal(t, "/app", gotPath)
	assert.Equal(t, "from vite", rec.Body.String())
}
```
- [ ] **Step 2: Run to confirm they fail**

Run: `go test ./server/internal/web/ -run SPA`
Expected: compile error `undefined: web.MountSPA`.

- [ ] **Step 3: Implement `spa.go`**

`server/internal/web/spa.go`:
```go
package web

import (
	"errors"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
)

// SPAPrefix is where the React client is served until it replaces the legacy
// site (sub-project 5 moves it to /).
const SPAPrefix = "/app"

const (
	cacheImmutable = "public, max-age=31536000, immutable"
	cacheNone      = "no-cache"
	notBuiltPage   = `<!doctype html><html lang="en"><head><meta charset="utf-8"><title>AFC Aldermaston</title></head>` +
		`<body><p>The web client has not been built. Run <code>yarn build:client</code>.</p></body></html>`
)

// MountSPA serves the React client under /app. With proxy set (development)
// every /app request is reverse-proxied to the Vite dev server, including
// HMR websockets. Otherwise files (the built client; may be nil) are served
// with an index.html fallback for client-side routes.
func MountSPA(e *echo.Echo, files fs.FS, proxy *url.URL) {
	var h echo.HandlerFunc
	if proxy != nil {
		slog.Info("serving " + SPAPrefix + " from the Vite dev server at " + proxy.String())
		rp := &httputil.ReverseProxy{Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(proxy)
			r.SetXForwarded()
		}}
		h = echo.WrapHandler(rp)
	} else {
		s := &spa{files: files}
		s.built = files != nil && isFile(files, "index.html")
		if !s.built {
			slog.Warn("web client not built: " + SPAPrefix + " will answer 503 until `yarn build` is run")
		}
		h = s.serve
	}
	e.Any(SPAPrefix, h)
	e.Any(SPAPrefix+"/*", h)
}

type spa struct {
	files fs.FS
	built bool
}

func (s *spa) serve(c echo.Context) error {
	if m := c.Request().Method; m != http.MethodGet && m != http.MethodHead {
		return c.String(http.StatusMethodNotAllowed, "method not allowed")
	}
	if !s.built {
		c.Response().Header().Set("Cache-Control", cacheNone)
		return c.HTML(http.StatusServiceUnavailable, notBuiltPage)
	}

	rel := strings.TrimPrefix(strings.TrimPrefix(c.Request().URL.Path, SPAPrefix), "/")
	switch {
	case rel == "":
		return s.file(c, "index.html", cacheNone)
	case !fs.ValidPath(rel) || strings.Contains(rel, "\\"):
		return c.String(http.StatusNotFound, "not found")
	case rel == "assets" || strings.HasPrefix(rel, "assets/"):
		return s.file(c, rel, cacheImmutable)
	case path.Ext(rel) != "":
		return s.file(c, rel, cacheNone)
	default:
		return s.file(c, "index.html", cacheNone)
	}
}

// file writes name from the build, or a plain 404. It never falls back to
// index.html, so a missing hashed asset is visible rather than served as HTML.
func (s *spa) file(c echo.Context, name, cacheControl string) error {
	b, err := fs.ReadFile(s.files, name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) || isDir(s.files, name) {
			return c.String(http.StatusNotFound, "not found")
		}
		return err
	}
	ctype := mime.TypeByExtension(path.Ext(name))
	if ctype == "" {
		ctype = http.DetectContentType(b)
	}
	c.Response().Header().Set("Cache-Control", cacheControl)
	return c.Blob(http.StatusOK, ctype, b)
}

func isFile(fsys fs.FS, name string) bool {
	info, err := fs.Stat(fsys, name)
	return err == nil && !info.IsDir()
}

func isDir(fsys fs.FS, name string) bool {
	info, err := fs.Stat(fsys, name)
	return err == nil && info.IsDir()
}
```
Notes:
- `c.Request().URL.Path` is already percent-decoded, so `%2e%2e` becomes `..` and `fs.ValidPath` rejects it. Echo's router may also normalise some traversal paths before they reach the handler; the test only asserts nothing outside the FS is served.
- `mime.TypeByExtension(".js")` returns `text/javascript; charset=utf-8` on Go 1.26.

- [ ] **Step 4: Run tests**

Run: `go test ./server/internal/web/ -run SPA -v`
Expected: PASS. If `/app/assets` (a directory) returns something other than 404, check the `isDir` branch in `file`: `fs.ReadFile` on a `MapFS` directory returns an error that is not `ErrNotExist`.

- [ ] **Step 5: Lint and commit**

```bash
gofmt -l server; go vet ./server/internal/web/ && golangci-lint run ./server/internal/web/...
git add server/internal/web/spa.go server/internal/web/spa_test.go
git commit -m "Add web.MountSPA: serve the React client under /app or proxy to Vite

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 2: Embed the client and wire `/app` into the app

**Files:**
- Create: `server/cmd/afc/ui.go`, `server/cmd/afc/ui/.keep` (empty), `server/internal/app/spa_test.go`
- Modify: `server/internal/app/app.go` (Config + Build), `server/cmd/afc/main.go`, `.gitignore`

**Interfaces:**
- Consumes: `web.MountSPA` (Task 1).
- Produces: `app.Config.UI fs.FS` and `app.Config.UIProxy *url.URL`; env var `AFC_UI_PROXY_URL`; package-main `uiFiles embed.FS`.

- [ ] **Step 1: Write the failing app-level tests**

`server/internal/app/spa_test.go`:
```go
package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/app"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

func serve(a *app.App, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	a.Echo.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestAppSPANotBuiltByDefault(t *testing.T) {
	a := bareApp(t)
	rec := serve(a, "/app")
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), "The web client has not been built")
}

func TestAppServesSPAFallback(t *testing.T) {
	conf := testConfig()
	conf.UI = fstest.MapFS{"index.html": {Data: []byte(`<div id="root"></div>`)}}
	a := app.Build(conf, app.Stores{}, uploadtest.New(), mail.NewMailer(mail.Config{}))
	t.Cleanup(a.Stop)

	for _, path := range []string{"/app", "/app/news/5"} {
		rec := serve(a, path)
		assert.Equal(t, http.StatusOK, rec.Code, path)
		assert.Contains(t, rec.Body.String(), `<div id="root">`, path)
	}
	// The API and legacy 404 handling are untouched.
	assert.Equal(t, http.StatusOK, serve(a, "/api/v1/health").Code)
	assert.Equal(t, http.StatusNotFound, serve(a, "/api/v1/nope").Code)
}
```

- [ ] **Step 2: Run to confirm they fail**

Run: `go test ./server/internal/app/ -run SPA`
Expected: compile error (`conf.UI undefined`).

- [ ] **Step 3: Wire it in**

In `server/internal/app/app.go`:
- Add imports `io/fs` and `net/url`.
- Add to `Config`, after `Redis`:
```go
	// UI is the built React client (index.html + assets/); nil means not built.
	UI fs.FS
	// UIProxy, when set, proxies /app to the Vite dev server instead of UI.
	UIProxy *url.URL
```
- In `Build`, right after `legacy.Mount(e, legacyViews)`, add:
```go
	web.MountSPA(e, conf.UI, conf.UIProxy)
```

Create `server/cmd/afc/ui.go`:
```go
package main

import "embed"

// uiFiles is the built React client. `yarn build:server` and the Dockerfile
// copy build/client into ui/ before compiling; only ui/.keep is committed, so
// a Go-only build serves a "client not built" page under /app.
//
//go:embed all:ui
var uiFiles embed.FS
```
Create the empty file `server/cmd/afc/ui/.keep`.

In `server/cmd/afc/main.go`:
- Add imports `io/fs` and `net/url`.
- Just before `a := app.New(app.Config{`, add:
```go
	uiRoot, err := fs.Sub(uiFiles, "ui")
	if err != nil {
		fatal(fmt.Sprintf("failed to open embedded web client: %+v", err))
	}
	var uiProxy *url.URL
	if raw := os.Getenv("AFC_UI_PROXY_URL"); raw != "" {
		uiProxy, err = url.Parse(raw)
		if err != nil || uiProxy.Scheme == "" || uiProxy.Host == "" {
			fatal(fmt.Sprintf("invalid AFC_UI_PROXY_URL %q: want e.g. http://localhost:5173", raw))
		}
	}
```
- Add to the `app.Config{…}` literal:
```go
		UI:      uiRoot,
		UIProxy: uiProxy,
```

In `.gitignore`, append:
```gitignore

# Built web client (copied in by `yarn build:server`); keep the embed target
server/cmd/afc/ui/*
!server/cmd/afc/ui/.keep
```

- [ ] **Step 4: Run tests**

```bash
export AFC_TEST_DB='postgres://postgres@localhost:55432/postgres?sslmode=disable'
go build ./... && go vet ./... && go test ./...
```
Expected: all pass. That includes the new tests, the legacy smoke test, `TestUnsafeAPIRoutesRequireLogin` (`/app` isn't under `/api/v1`) and `TestEveryLegacyActionHasAnAPIRoute`.

- [ ] **Step 5: Lint and commit**

```bash
gofmt -l server; golangci-lint run ./...
git add .gitignore server/cmd/afc/ui.go server/cmd/afc/ui/.keep server/cmd/afc/main.go server/internal/app/app.go server/internal/app/spa_test.go
git commit -m "Embed the web client and serve it under /app (AFC_UI_PROXY_URL for dev)

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 3: JS toolchain and client skeleton

**Files:**
- Create: `package.json`, `yarn.lock` (generated), `.yarnrc.yml`, `vite.config.ts`, `tsconfig.json`, `tsconfig.app.json`, `tsconfig.node.json`, `eslint.config.js`, `.prettierrc`, `.prettierignore`
- Create: `client/index.html`, `client/main.tsx`, `client/App.tsx`, `client/App.test.tsx`, `client/vite-env.d.ts`, `client/public/favicon.svg`, `client/components/Layout.tsx`, `client/pages/HomePage.tsx` (placeholder, replaced in Task 4), `client/pages/NotFoundPage.tsx`, `client/test/setup.ts`
- Create: `scripts/clean.js`, `scripts/build-client.js`
- Modify: `.gitignore`

**Interfaces:**
- Produces:
  - Scripts: `yarn lint:client`, `yarn typecheck`, `yarn test:client`, `yarn build:client` (outputs `build/client/index.html` + `build/client/assets/*`), `yarn clean`.
  - `client/App.tsx` default export `App`: a `<Routes>` tree with `Layout` as the parent route; `"/"` → `HomePage`, `"*"` → `NotFoundPage`.
  - `client/components/Layout.tsx` default export.

- [ ] **Step 1: Initialise Yarn 4 and add dependencies**

```bash
corepack enable
cat > package.json <<'EOF'
{
  "name": "afc",
  "private": true,
  "version": "0.0.0",
  "type": "module",
  "scripts": {
    "dev:client": "vite --host",
    "build:client": "node scripts/build-client.js",
    "clean": "node scripts/clean.js",
    "lint:client": "eslint .",
    "typecheck": "tsc -b",
    "test:client": "vitest run"
  }
}
EOF
corepack use yarn@4
cat > .yarnrc.yml <<'EOF'
enableScripts: true

nodeLinker: node-modules

npmMinimalAgeGate: 4d
EOF
yarn add react react-dom react-router @tanstack/react-query
yarn add -D typescript vite @vitejs/plugin-react @types/react @types/react-dom @types/node \
  eslint @eslint/js typescript-eslint globals eslint-plugin-react eslint-plugin-react-hooks \
  eslint-plugin-react-refresh eslint-plugin-prettier eslint-config-prettier prettier \
  vitest jsdom @testing-library/react @testing-library/jest-dom
```
`corepack use yarn@4` writes the `packageManager` field. Keep whatever current versions resolve; the lockfile pins them. If `npmMinimalAgeGate` blocks a freshly published version, Yarn picks the newest one old enough, which is fine.

- [ ] **Step 2: Config files**

`tsconfig.json`:
```json
{
  "files": [],
  "references": [{ "path": "./tsconfig.app.json" }, { "path": "./tsconfig.node.json" }]
}
```

`tsconfig.app.json`:
```json
{
  "compilerOptions": {
    "tsBuildInfoFile": "./node_modules/.tmp/tsconfig.app.tsbuildinfo",
    "target": "ES2023",
    "lib": ["ES2023", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "types": ["vite/client"],
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "verbatimModuleSyntax": true,
    "moduleDetection": "force",
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "erasableSyntaxOnly": true,
    "noFallthroughCasesInSwitch": true
  },
  "include": ["client"]
}
```

`tsconfig.node.json`:
```json
{
  "compilerOptions": {
    "tsBuildInfoFile": "./node_modules/.tmp/tsconfig.node.tsbuildinfo",
    "target": "ES2023",
    "lib": ["ES2023"],
    "module": "ESNext",
    "types": ["node"],
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "verbatimModuleSyntax": true,
    "moduleDetection": "force",
    "noEmit": true,
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true
  },
  "include": ["vite.config.ts"]
}
```

`vite.config.ts`:
```ts
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

const root = fileURLToPath(new URL('.', import.meta.url));

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  root: resolve(root, 'client'),
  // Served by the Go server under /app until the SPA replaces the legacy site.
  base: '/app/',
  build: {
    outDir: resolve(root, 'build/client'),
    emptyOutDir: true,
    sourcemap: false,
  },
  server: {
    port: 5173,
    strictPort: true,
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./test/setup.ts'],
    css: false,
  },
});
```

`eslint.config.js`:
```js
import js from '@eslint/js';
import pluginPrettier from 'eslint-plugin-prettier';
import pluginReact from 'eslint-plugin-react';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';
import { defineConfig, globalIgnores } from 'eslint/config';
import globals from 'globals';
import tseslint from 'typescript-eslint';

export default defineConfig([
  globalIgnores(['build', 'node_modules', 'server', '.yarn', '**/*.d.ts']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
    ],
    languageOptions: {
      ecmaVersion: 'latest',
      globals: { ...globals.browser, ...globals.node },
    },
    settings: { react: { version: '19.0' } },
    plugins: { react: pluginReact, prettier: pluginPrettier },
    rules: {
      ...pluginReact.configs.recommended.rules,
      ...pluginReact.configs['jsx-runtime'].rules,
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
      'prettier/prettier': 'error',
    },
  },
]);
```
If the installed `eslint-plugin-react-hooks` has no `configs.flat.recommended`, use `reactHooks.configs['recommended-latest']`. This is the only permitted adaptation.

`.prettierrc`:
```json
{
  "printWidth": 100,
  "tabWidth": 2,
  "useTabs": false,
  "semi": true,
  "singleQuote": true,
  "bracketSpacing": true
}
```

`.prettierignore`:
```
build
node_modules
server
.yarn
yarn.lock
```

Append to `.gitignore`:
```gitignore

# Node / Yarn
node_modules/
build/
.yarn/*
!.yarn/patches
!.yarn/plugins
!.yarn/releases
!.yarn/sdks
!.yarn/versions
```

- [ ] **Step 3: Write the failing client test**

`client/test/setup.ts`:
```ts
import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterEach, vi } from 'vitest';

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});
```

`client/App.test.tsx`:
```tsx
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { describe, expect, it } from 'vitest';

import App from './App';

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <App />
    </MemoryRouter>,
  );
}

describe('App routes', () => {
  it('renders the home page at /', () => {
    renderAt('/');
    expect(screen.getByRole('heading', { level: 1, name: 'AFC Aldermaston' })).toBeInTheDocument();
  });

  it('renders the not-found page for unknown routes', () => {
    renderAt('/no/such/page');
    expect(screen.getByRole('heading', { level: 1, name: 'Page not found' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Go to the start' })).toHaveAttribute('href', '/');
  });

  it('wraps pages in the layout', () => {
    renderAt('/');
    expect(screen.getByRole('banner')).toBeInTheDocument();
    expect(screen.getByRole('main')).toBeInTheDocument();
    expect(screen.getByRole('contentinfo')).toHaveTextContent('AFC Aldermaston');
  });
});
```
Run: `yarn test:client`. Expected: FAIL (`./App` not found).

- [ ] **Step 4: Client skeleton**

`client/index.html`:
```html
<!doctype html>
<html lang="en-GB">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/app/favicon.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta name="description" content="AFC Aldermaston Football Club" />
    <title>AFC Aldermaston</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/main.tsx"></script>
  </body>
</html>
```

`client/public/favicon.svg`:
```svg
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><circle cx="16" cy="16" r="15" fill="#fc0f2b"/><text x="16" y="21" font-size="12" text-anchor="middle" fill="#fff" font-family="Arial, sans-serif">AFC</text></svg>
```

`client/vite-env.d.ts`:
```ts
/// <reference types="vite/client" />
```

`client/main.tsx` (Task 4 adds the providers):
```tsx
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router';

import App from './App';

const root = document.getElementById('root');
if (!root) {
  throw new Error('missing #root element');
}

createRoot(root).render(
  <StrictMode>
    <BrowserRouter basename="/app">
      <App />
    </BrowserRouter>
  </StrictMode>,
);
```

`client/App.tsx`:
```tsx
import { Route, Routes } from 'react-router';

import Layout from './components/Layout';
import HomePage from './pages/HomePage';
import NotFoundPage from './pages/NotFoundPage';

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<HomePage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  );
}
```

`client/components/Layout.tsx`:
```tsx
import { Outlet } from 'react-router';

// Bare page shell. Visual design is sub-project 3.
export default function Layout() {
  return (
    <>
      <header>
        <a href="/">AFC Aldermaston</a>
      </header>
      <main>
        <Outlet />
      </main>
      <footer>© {new Date().getFullYear()} AFC Aldermaston</footer>
    </>
  );
}
```
The header link is a plain `<a href="/">`, not a router `Link`, because `/` is the legacy site outside the SPA's `/app` basename.

`client/pages/HomePage.tsx` (placeholder; Task 4 replaces it):
```tsx
export default function HomePage() {
  return (
    <section>
      <h1>AFC Aldermaston</h1>
    </section>
  );
}
```

`client/pages/NotFoundPage.tsx`:
```tsx
import { Link } from 'react-router';

export default function NotFoundPage() {
  return (
    <section>
      <h1>Page not found</h1>
      <p>
        <Link to="/">Go to the start</Link>
      </p>
    </section>
  );
}
```

- [ ] **Step 5: Scripts**

`scripts/clean.js`:
```js
// Removes build output and resets the Go embed target to just .keep.
import { mkdirSync, rmSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

const root = process.cwd();
rmSync(join(root, 'build'), { recursive: true, force: true });

const ui = join(root, 'server', 'cmd', 'afc', 'ui');
rmSync(ui, { recursive: true, force: true });
mkdirSync(ui, { recursive: true });
writeFileSync(join(ui, '.keep'), '');
```

`scripts/build-client.js`:
```js
// Lints (unless BUILD_CLIENT_SKIP_LINT=true), type-checks and builds the client into build/client.
import { execSync } from 'node:child_process';

if (process.env.BUILD_CLIENT_SKIP_LINT !== 'true') {
  execSync('yarn lint:client', { stdio: 'inherit' });
}
execSync('tsc -b && vite build', { stdio: 'inherit' });
```

- [ ] **Step 6: Run everything**

```bash
yarn exec prettier --write client vite.config.ts eslint.config.js scripts
yarn lint:client && yarn typecheck && yarn test:client && yarn build:client
ls build/client build/client/assets
grep -o 'src="/app/assets/[^"]*"' build/client/index.html
```
Expected:
- Lint is clean (0 errors).
- Typecheck is clean.
- 3 tests pass.
- `build/client/index.html` references `/app/assets/…js`, and `build/client/favicon.svg` exists.

Also confirm Go is unaffected: `go build ./...`.

- [ ] **Step 7: Commit**

```bash
git add package.json yarn.lock .yarnrc.yml vite.config.ts tsconfig.json tsconfig.app.json tsconfig.node.json \
  eslint.config.js .prettierrc .prettierignore .gitignore client scripts
git status --short   # no node_modules/, build/ or .yarn/cache
git commit -m "Add Vite/React/TypeScript client skeleton served at /app

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 4: Data layer, auth context and proof page

**Files:**
- Create: `client/api/{client.ts,client.test.ts,types.ts,queryClient.ts,queries.ts,queries.test.tsx}`
- Create: `client/auth/{context.ts,AuthProvider.tsx,useAuth.ts}`
- Create: `client/test/{render.tsx,mockFetch.ts}`, `client/pages/HomePage.test.tsx`
- Modify: `client/pages/HomePage.tsx`, `client/main.tsx`, `client/App.test.tsx`

**Interfaces:**
- Consumes: the skeleton from Task 3; API endpoints `GET /api/v1/site` and `GET /api/v1/auth/me` (existing Go code).
- Produces:
  - `apiFetch<T>(path: string, req?: ApiRequest): Promise<T>`, `ApiError` (`status`, `message`, `fields`), `API_BASE = '/api/v1'`.
  - Types: `SiteInfo`, `TeamSummary`, `CurrentUser`, `Permissions`, `ErrorEnvelope`.
  - Query client: `createQueryClient()`, `shouldRetry(failureCount, error)`.
  - Queries: `queryKeys`, `useSite()`, `useMe()`, `fetchMe(signal?)`.
  - Auth: `AuthProvider`, `useAuth(): AuthState {user, isLoading, refresh}`.
  - Test helpers: `renderWithProviders(ui, {route})`, `mockFetch(routes)`.

- [ ] **Step 1: Test helpers**

`client/test/mockFetch.ts`:
```ts
import { vi } from 'vitest';

export interface MockResponse {
  status?: number;
  body?: unknown; // objects are JSON-encoded; strings are sent as-is
  contentType?: string;
}

/**
 * Replaces global fetch with a stub keyed by path (e.g. '/api/v1/site').
 * Unknown paths fail the test loudly with a 599.
 */
export function mockFetch(routes: Record<string, MockResponse | (() => never)>) {
  const fn = vi.fn(async (input: RequestInfo | URL) => {
    const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url;
    const path = url.replace(/^https?:\/\/[^/]+/, '');
    const route = routes[path];
    if (!route) {
      return new Response(`no mock for ${path}`, { status: 599 });
    }
    if (typeof route === 'function') {
      return route();
    }
    const status = route.status ?? 200;
    if (status === 204) {
      return new Response(null, { status });
    }
    const isText = typeof route.body === 'string';
    return new Response(
      route.body === undefined ? '' : isText ? (route.body as string) : JSON.stringify(route.body),
      {
        status,
        headers: {
          'Content-Type': route.contentType ?? (isText ? 'text/html' : 'application/json'),
        },
      },
    );
  });
  vi.stubGlobal('fetch', fn);
  return fn;
}
```

`client/test/render.tsx`:
```tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render } from '@testing-library/react';
import type { ReactElement } from 'react';
import { MemoryRouter } from 'react-router';

import { AuthProvider } from '../auth/AuthProvider';

/** Renders ui inside the same providers as main.tsx, with retries off. */
export function renderWithProviders(ui: ReactElement, { route = '/' }: { route?: string } = {}) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: Infinity } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[route]}>
        <AuthProvider>{ui}</AuthProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}
```

- [ ] **Step 2: Failing tests for `apiFetch` and the retry policy**

`client/api/client.test.ts`:
```ts
import { describe, expect, it, vi } from 'vitest';

import { mockFetch } from '../test/mockFetch';
import { ApiError, apiFetch } from './client';
import { shouldRetry } from './queryClient';

describe('apiFetch', () => {
  it('returns parsed JSON and calls /api/v1 same-origin', async () => {
    const fetchMock = mockFetch({ '/api/v1/site': { body: { year: 2026 } } });
    await expect(apiFetch<{ year: number }>('/site')).resolves.toEqual({ year: 2026 });
    const [, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
    expect(init.credentials).toBe('same-origin');
    expect(init.method).toBe('GET');
  });

  it('sends JSON bodies with a content type and defaults to POST', async () => {
    const fetchMock = mockFetch({ '/api/v1/auth/login': { body: { ok: true } } });
    await apiFetch('/auth/login', { json: { email: 'a@b.test' } });
    const [, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
    expect(init.method).toBe('POST');
    expect((init.headers as Record<string, string>)['Content-Type']).toBe('application/json');
    expect(init.body).toBe('{"email":"a@b.test"}');
  });

  it('sends FormData without setting a content type', async () => {
    const fetchMock = mockFetch({ '/api/v1/news': { status: 201, body: { id: 1 } } });
    const form = new FormData();
    form.set('title', 'x');
    await apiFetch('/news', { form });
    const [, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
    expect(init.body).toBe(form);
    expect((init.headers as Record<string, string>)['Content-Type']).toBeUndefined();
  });

  it('returns undefined for 204', async () => {
    mockFetch({ '/api/v1/news/1': { status: 204 } });
    await expect(apiFetch('/news/1', { method: 'DELETE' })).resolves.toBeUndefined();
  });

  it('turns the error envelope into an ApiError with fields', async () => {
    mockFetch({
      '/api/v1/news': {
        status: 422,
        body: { error: { code: 422, message: 'validation failed', fields: { title: 'title is required' } } },
      },
    });
    const err = await apiFetch('/news', { json: {} }).catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect(err).toMatchObject({ status: 422, message: 'validation failed', fields: { title: 'title is required' } });
  });

  it('non-JSON error body becomes an ApiError with the status text', async () => {
    mockFetch({ '/api/v1/site': { status: 502, body: '<html>Bad gateway</html>' } });
    const err = await apiFetch('/site').catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect((err as ApiError).status).toBe(502);
    expect((err as ApiError).message).not.toContain('<html>');
    expect((err as ApiError).fields).toEqual({});
  });

  it('empty 200 body is an error, not undefined data', async () => {
    mockFetch({ '/api/v1/site': { status: 200, body: '', contentType: 'application/json' } });
    const err = await apiFetch('/site').catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect((err as ApiError).message).toBe('The server sent an unexpected response.');
  });

  it('network failure is ApiError status 0', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => Promise.reject(new TypeError('Failed to fetch'))));
    const err = await apiFetch('/site').catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect((err as ApiError).status).toBe(0);
  });

  it('lets aborts propagate unchanged', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => Promise.reject(new DOMException('aborted', 'AbortError'))));
    const err = await apiFetch('/site').catch((e: unknown) => e);
    expect(err).toBeInstanceOf(DOMException);
  });
});

describe('shouldRetry', () => {
  it('never retries 4xx ApiErrors', () => {
    expect(shouldRetry(0, new ApiError(401, 'login required'))).toBe(false);
    expect(shouldRetry(0, new ApiError(404, 'not found'))).toBe(false);
  });
  it('retries other errors up to twice', () => {
    expect(shouldRetry(0, new ApiError(503, 'down'))).toBe(true);
    expect(shouldRetry(1, new ApiError(0, 'offline'))).toBe(true);
    expect(shouldRetry(2, new ApiError(500, 'boom'))).toBe(false);
    expect(shouldRetry(0, new Error('x'))).toBe(true);
  });
});
```
Run: `yarn test:client`. Expected: FAIL (modules missing).

- [ ] **Step 3: Implement the API layer**

`client/api/types.ts`:
```ts
// Hand-written mirrors of the Go API's JSON (server/internal/*/types.go).

export interface ErrorEnvelope {
  error: {
    code: number;
    message: string;
    fields?: Record<string, string>;
  };
}

/** team.Public */
export interface TeamSummary {
  id: number;
  name: string;
  description?: string;
  league?: string;
  division?: string;
  leagueTableUrl?: string;
  fixturesUrl?: string;
  coach?: string;
  physio?: string;
  imageUrl?: string;
  isActive: boolean;
  isYouth: boolean;
  ages: number;
}

/** site.Info — GET /site */
export interface SiteInfo {
  year: number;
  visitorCount: number;
  displayEmail?: string;
  version: string;
  teams: TeamSummary[];
}

/** auth.Permissions */
export interface Permissions {
  canEdit: boolean;
  canManageGallery: boolean;
  canManageUsers: boolean;
}

/** auth.CurrentUser — GET /auth/me */
export interface CurrentUser {
  id: number;
  name: string;
  email: string;
  phone?: string;
  role: string;
  teamId?: number;
  imageUrl?: string;
  permissions: Permissions;
}
```
Before relying on these, check each field name against the Go JSON tags in `server/internal/team/types.go`, `server/internal/site/types.go` and `server/internal/auth/types.go`, and fix any mismatch here, not in Go.

`client/api/client.ts`:
```ts
import type { ErrorEnvelope } from './types';

export const API_BASE = '/api/v1';

/** A failed API call. status 0 means the server could not be reached. */
export class ApiError extends Error {
  readonly status: number;
  readonly fields: Record<string, string>;

  constructor(status: number, message: string, fields: Record<string, string> = {}) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.fields = fields;
  }
}

export interface ApiRequest {
  method?: string;
  json?: unknown;
  form?: FormData;
  signal?: AbortSignal;
}

function isEnvelope(data: unknown): data is ErrorEnvelope {
  if (typeof data !== 'object' || data === null || !('error' in data)) {
    return false;
  }
  const error = (data as { error: unknown }).error;
  return (
    typeof error === 'object' &&
    error !== null &&
    typeof (error as { message?: unknown }).message === 'string'
  );
}

/**
 * Calls the JSON API. CSRF needs no handling: browsers send
 * Sec-Fetch-Site: same-origin, which the server accepts.
 */
export async function apiFetch<T>(path: string, req: ApiRequest = {}): Promise<T> {
  const headers: Record<string, string> = { Accept: 'application/json' };
  let body: BodyInit | undefined;
  if (req.json !== undefined) {
    headers['Content-Type'] = 'application/json';
    body = JSON.stringify(req.json);
  } else if (req.form) {
    body = req.form; // the browser sets the multipart boundary
  }

  let res: Response;
  try {
    res = await fetch(API_BASE + path, {
      method: req.method ?? (body === undefined ? 'GET' : 'POST'),
      headers,
      body,
      credentials: 'same-origin',
      signal: req.signal,
    });
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      throw err;
    }
    throw new ApiError(0, 'Could not reach the server. Check your connection and try again.');
  }

  if (res.status === 204) {
    return undefined as T;
  }

  const text = await res.text();
  let data: unknown;
  try {
    data = text === '' ? undefined : JSON.parse(text);
  } catch {
    data = undefined;
  }

  if (!res.ok) {
    if (isEnvelope(data)) {
      throw new ApiError(res.status, data.error.message, data.error.fields ?? {});
    }
    throw new ApiError(res.status, res.statusText || `Request failed with status ${res.status}`);
  }
  if (data === undefined) {
    throw new ApiError(res.status, 'The server sent an unexpected response.');
  }
  return data as T;
}
```

`client/api/queryClient.ts`:
```ts
import { QueryClient } from '@tanstack/react-query';

import { ApiError } from './client';

/** Client errors (4xx) never succeed on retry; anything else gets two more tries. */
export function shouldRetry(failureCount: number, error: unknown): boolean {
  if (error instanceof ApiError && error.status >= 400 && error.status < 500) {
    return false;
  }
  return failureCount < 2;
}

export function createQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: shouldRetry, staleTime: 30_000 },
    },
  });
}
```

`client/api/queries.ts`:
```ts
import { useQuery } from '@tanstack/react-query';

import { ApiError, apiFetch } from './client';
import type { CurrentUser, SiteInfo } from './types';

export const queryKeys = {
  site: ['site'] as const,
  me: ['auth', 'me'] as const,
};

export function useSite() {
  return useQuery({
    queryKey: queryKeys.site,
    queryFn: ({ signal }) => apiFetch<SiteInfo>('/site', { signal }),
  });
}

/** The signed-in user, or null when nobody is (a 401 is not an error here). */
export async function fetchMe(signal?: AbortSignal): Promise<CurrentUser | null> {
  try {
    return await apiFetch<CurrentUser>('/auth/me', { signal });
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      return null;
    }
    throw err;
  }
}

export function useMe() {
  return useQuery({
    queryKey: queryKeys.me,
    queryFn: ({ signal }) => fetchMe(signal),
  });
}
```

`client/auth/context.ts`:
```ts
import { createContext } from 'react';

import type { CurrentUser } from '../api/types';

export interface AuthState {
  user: CurrentUser | null;
  isLoading: boolean;
  /** Re-checks who is signed in (after login/logout in sub-project 4). */
  refresh: () => Promise<void>;
}

export const AuthContext = createContext<AuthState | null>(null);
```

`client/auth/AuthProvider.tsx`:
```tsx
import { useQueryClient } from '@tanstack/react-query';
import { useMemo, type ReactNode } from 'react';

import { queryKeys, useMe } from '../api/queries';
import { AuthContext, type AuthState } from './context';

export function AuthProvider({ children }: { children: ReactNode }) {
  const me = useMe();
  const queryClient = useQueryClient();

  const value = useMemo<AuthState>(
    () => ({
      user: me.data ?? null,
      isLoading: me.isPending,
      refresh: () => queryClient.invalidateQueries({ queryKey: queryKeys.me }),
    }),
    [me.data, me.isPending, queryClient],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
```

`client/auth/useAuth.ts`:
```ts
import { useContext } from 'react';

import { AuthContext, type AuthState } from './context';

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used inside <AuthProvider>');
  }
  return ctx;
}
```

Run: `yarn test:client`. Expected: `client.test.ts` passes.

- [ ] **Step 4: Failing tests for `fetchMe` and the proof page**

`client/api/queries.test.tsx`:
```tsx
import { describe, expect, it } from 'vitest';

import { mockFetch } from '../test/mockFetch';
import { ApiError } from './client';
import { fetchMe } from './queries';

const me = {
  id: 4,
  name: 'Trea Surer',
  email: 'treasurer@example.test',
  role: 'Treasurer',
  permissions: { canEdit: true, canManageGallery: true, canManageUsers: false },
};

describe('fetchMe', () => {
  it('returns the current user', async () => {
    mockFetch({ '/api/v1/auth/me': { body: me } });
    await expect(fetchMe()).resolves.toEqual(me);
  });

  it('treats 401 as anonymous', async () => {
    mockFetch({
      '/api/v1/auth/me': { status: 401, body: { error: { code: 401, message: 'login required' } } },
    });
    await expect(fetchMe()).resolves.toBeNull();
  });

  it('still throws other errors', async () => {
    mockFetch({
      '/api/v1/auth/me': { status: 500, body: { error: { code: 500, message: 'internal server error' } } },
    });
    await expect(fetchMe()).rejects.toBeInstanceOf(ApiError);
  });
});
```

`client/pages/HomePage.test.tsx`:
```tsx
import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { mockFetch } from '../test/mockFetch';
import { renderWithProviders } from '../test/render';
import HomePage from './HomePage';

const site = {
  year: 2026,
  visitorCount: 1234,
  version: 'test',
  teams: [
    { id: 1, name: 'First Team', isActive: true, isYouth: false, ages: 99 },
    { id: 2, name: 'Under 12s', isActive: true, isYouth: true, ages: 12 },
  ],
};

const anonymous = {
  status: 401,
  body: { error: { code: 401, message: 'login required' } },
};

describe('HomePage', () => {
  it('shows site info for an anonymous visitor', async () => {
    mockFetch({ '/api/v1/site': { body: site }, '/api/v1/auth/me': anonymous });
    renderWithProviders(<HomePage />);
    expect(await screen.findByText('Not signed in')).toBeInTheDocument();
    expect(await screen.findByText('Visitors: 1234')).toBeInTheDocument();
    expect(screen.getByText('First Team')).toBeInTheDocument();
    expect(screen.getByText('Under 12s')).toBeInTheDocument();
  });

  it('shows who is signed in', async () => {
    mockFetch({
      '/api/v1/site': { body: site },
      '/api/v1/auth/me': {
        body: {
          id: 1,
          name: 'Web Master',
          email: 'webmaster@example.test',
          role: 'Webmaster',
          permissions: { canEdit: true, canManageGallery: true, canManageUsers: true },
        },
      },
    });
    renderWithProviders(<HomePage />);
    expect(await screen.findByText('Signed in as Web Master (Webmaster)')).toBeInTheDocument();
  });

  it('shows a readable error when the API fails', async () => {
    mockFetch({
      '/api/v1/site': { status: 500, body: { error: { code: 500, message: 'internal server error' } } },
      '/api/v1/auth/me': anonymous,
    });
    renderWithProviders(<HomePage />);
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Could not load site information: internal server error',
    );
  });

  it('says so when there are no teams', async () => {
    mockFetch({ '/api/v1/site': { body: { ...site, teams: [] } }, '/api/v1/auth/me': anonymous });
    renderWithProviders(<HomePage />);
    expect(await screen.findByText('No teams yet.')).toBeInTheDocument();
  });
});
```
Update `client/App.test.tsx` so `/` renders inside the providers with a mocked API. Replace its `renderAt` helper and the first test with:
```tsx
import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import App from './App';
import { mockFetch } from './test/mockFetch';
import { renderWithProviders } from './test/render';

function renderAt(path: string) {
  mockFetch({
    '/api/v1/site': { body: { year: 2026, visitorCount: 1, version: 'test', teams: [] } },
    '/api/v1/auth/me': { status: 401, body: { error: { code: 401, message: 'login required' } } },
  });
  return renderWithProviders(<App />, { route: path });
}
```
Keep the three `it` blocks unchanged (they use `renderAt`).

Run: `yarn test:client`. Expected: the HomePage tests fail (the placeholder has no data).

- [ ] **Step 5: Proof page and providers**

`client/pages/HomePage.tsx`:
```tsx
import { useSite } from '../api/queries';
import { useAuth } from '../auth/useAuth';

// Proof that the client, API, session cookie and data layer work end to end.
// Real pages arrive in sub-project 4.
export default function HomePage() {
  const site = useSite();
  const { user, isLoading } = useAuth();

  let signedIn = 'Not signed in';
  if (isLoading) {
    signedIn = 'Checking sign-in…';
  } else if (user) {
    signedIn = `Signed in as ${user.name} (${user.role})`;
  }

  return (
    <section>
      <h1>AFC Aldermaston</h1>
      <p>{signedIn}</p>
      {site.isPending && <p>Loading…</p>}
      {site.isError && (
        <p role="alert">Could not load site information: {site.error.message}</p>
      )}
      {site.data && (
        <>
          <p>Visitors: {site.data.visitorCount}</p>
          <h2>Teams</h2>
          {site.data.teams.length === 0 ? (
            <p>No teams yet.</p>
          ) : (
            <ul>
              {site.data.teams.map((team) => (
                <li key={team.id}>{team.name}</li>
              ))}
            </ul>
          )}
        </>
      )}
    </section>
  );
}
```

`client/main.tsx`:
```tsx
import { QueryClientProvider } from '@tanstack/react-query';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router';

import { createQueryClient } from './api/queryClient';
import App from './App';
import { AuthProvider } from './auth/AuthProvider';

const root = document.getElementById('root');
if (!root) {
  throw new Error('missing #root element');
}

const queryClient = createQueryClient();

createRoot(root).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter basename="/app">
        <AuthProvider>
          <App />
        </AuthProvider>
      </BrowserRouter>
    </QueryClientProvider>
  </StrictMode>,
);
```

- [ ] **Step 6: Run everything**

```bash
yarn exec prettier --write client
yarn lint:client && yarn typecheck && yarn test:client && yarn build:client
```
Expected: all green, and every client test passes (App 3, client 11, queries 3, HomePage 4).

- [ ] **Step 7: Commit**

```bash
git add client
git commit -m "Add typed API client, TanStack Query, auth context and /app proof page

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 5: Build, dev, Docker and CI pipeline

**Files:**
- Create: `scripts/build-server.js`, `scripts/dev-server.js`, `scripts/build-emails.js`, `.dockerignore`, `.github/workflows/client.yaml`
- Modify: `package.json` (scripts, devDeps `concurrently`, `mjml`), `Dockerfile`, `README.md`

**Interfaces:**
- Consumes: `yarn build:client` (Task 3), Go embed target `server/cmd/afc/ui` (Task 2), `AFC_UI_PROXY_URL` (Task 2).
- Produces:
  - `yarn build`, which gives `build/afc` with the client embedded.
  - `yarn dev`, `yarn build:server`, `yarn build:docs`, `yarn build:emails`, `yarn lint`, `yarn lint:server`, `yarn test`, `yarn test:server`.
  - A Docker image that serves `/app`.

- [ ] **Step 1: Dependencies and scripts**

```bash
yarn add -D concurrently mjml@^4
```
Set the `"scripts"` block of `package.json` to:
```json
  "scripts": {
    "dev": "concurrently -k -n client,server -c blue,green \"yarn dev:client\" \"yarn dev:server\"",
    "dev:client": "vite --host",
    "dev:server": "node scripts/dev-server.js",
    "build": "yarn clean && yarn build:client && yarn build:server",
    "build:client": "node scripts/build-client.js",
    "build:server": "node scripts/build-server.js",
    "build:docs": "go generate ./server/internal/docs",
    "build:emails": "node scripts/build-emails.js",
    "clean": "node scripts/clean.js",
    "lint": "yarn lint:client && yarn lint:server",
    "lint:client": "eslint .",
    "lint:server": "golangci-lint run ./...",
    "typecheck": "tsc -b",
    "test": "yarn test:client && yarn test:server",
    "test:client": "vitest run",
    "test:server": "go test ./..."
  },
```

`scripts/build-server.js`:
```js
// Copies build/client into the Go embed target, builds build/afc, then resets
// the embed target to just .keep so the working tree stays clean.
import { execSync } from 'node:child_process';
import { cpSync, existsSync, mkdirSync, rmSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

const root = process.cwd();
const client = join(root, 'build', 'client');
const ui = join(root, 'server', 'cmd', 'afc', 'ui');

function resetUI() {
  rmSync(ui, { recursive: true, force: true });
  mkdirSync(ui, { recursive: true });
  writeFileSync(join(ui, '.keep'), '');
}

if (!existsSync(join(client, 'index.html'))) {
  console.warn('build/client not found: building the server without the web client (/app will answer 503).');
}

resetUI();
try {
  if (existsSync(client)) {
    cpSync(client, ui, { recursive: true });
  }
  execSync('go build -o build/afc ./server/cmd/afc', { stdio: 'inherit' });
  console.log('Built build/afc');
} finally {
  resetUI();
}
```

`scripts/dev-server.js`:
```js
// Runs the Go server with /app proxied to the Vite dev server (yarn dev:client).
import { spawn } from 'node:child_process';

const env = {
  ...process.env,
  AFC_UI_PROXY_URL: process.env.AFC_UI_PROXY_URL ?? 'http://localhost:5173',
};

const server = spawn('go', ['run', './server/cmd/afc'], { stdio: 'inherit', env });
server.on('exit', (code) => process.exit(code ?? 0));
for (const signal of ['SIGINT', 'SIGTERM']) {
  process.on(signal, () => server.kill(signal));
}
```

`scripts/build-emails.js`:
```js
// Compiles server/internal/emails/mjml/<name>.mjml to server/internal/emails/<name>.tmpl.
import { readdirSync, readFileSync, writeFileSync } from 'node:fs';
import { basename, join } from 'node:path';
import mjml2html from 'mjml';

const dir = join(process.cwd(), 'server', 'internal', 'emails');
const src = join(dir, 'mjml');

let failed = false;
for (const file of readdirSync(src).filter((f) => f.endsWith('.mjml'))) {
  const { html, errors } = mjml2html(readFileSync(join(src, file), 'utf8'), {
    validationLevel: 'soft',
  });
  for (const e of errors) {
    console.error(`${file}: line ${e.line}: ${e.message}`);
  }
  if (errors.length > 0) {
    failed = true;
    continue;
  }
  const out = join(dir, `${basename(file, '.mjml')}.tmpl`);
  writeFileSync(out, html);
  console.log(`wrote ${out}`);
}
process.exit(failed ? 1 : 0);
```

- [ ] **Step 2: Verify the local build end to end**

```bash
export AFC_TEST_DB='postgres://postgres@localhost:55432/postgres?sslmode=disable'
yarn build
ls -la build/afc
git status --short server/cmd/afc/ui   # expect nothing: the embed target is reset to .keep
strings build/afc | grep -c '/app/assets/index-' # >= 1: index.html is embedded
```
Then prove the built client is served correctly: extend `server/internal/app/spa_test.go` with a test that only runs when `AFC_BUILT_UI` points at a built client:
```go
func TestAppServesBuiltClient(t *testing.T) {
	dir := os.Getenv("AFC_BUILT_UI")
	if dir == "" {
		t.Skip("AFC_BUILT_UI not set; run after `yarn build:client` with AFC_BUILT_UI=$PWD/build/client")
	}
	conf := testConfig()
	conf.UI = os.DirFS(dir)
	a := app.Build(conf, app.Stores{}, uploadtest.New(), mail.NewMailer(mail.Config{}))
	t.Cleanup(a.Stop)

	rec := serve(a, "/app/news/5")
	require.Equal(t, http.StatusOK, rec.Code)
	m := regexp.MustCompile(`/app/assets/[^"]+\.js`).FindString(rec.Body.String())
	require.NotEmpty(t, m, "index.html should reference a hashed script under /app/assets")
	asset := serve(a, m)
	assert.Equal(t, http.StatusOK, asset.Code)
	assert.Equal(t, "public, max-age=31536000, immutable", asset.Header().Get("Cache-Control"))
}
```
(Add `os`, `regexp` and `github.com/stretchr/testify/require` imports.) Run:
```bash
AFC_BUILT_UI=$PWD/build/client go test ./server/internal/app/ -run BuiltClient -v
```
Expected: PASS. Without the env var it skips.

- [ ] **Step 3: Emails**

```bash
cp server/internal/emails/resetEmail.tmpl /tmp/resetEmail.before.tmpl
yarn build:emails
git diff --stat server/internal/emails
go test ./server/internal/emails/
```
If the tests pass, keep the regenerated templates (Plan Decision 5). If they fail, run `git checkout server/internal/emails/*.tmpl`, keep the script, and record the failure in the report.

- [ ] **Step 4: Docker and CI**

`.dockerignore`:
```
.git
node_modules
build
.yarn/cache
.yarn/install-state.gz
.env
.env.*
*.sql
FileStore
.superpowers
.idea
.DS_Store
```

`Dockerfile`: add this stage **before** the existing `FROM golang…` line:
```dockerfile
FROM node:24-alpine AS client

WORKDIR /src/
RUN corepack enable

COPY package.json yarn.lock .yarnrc.yml ./
RUN yarn install --immutable

COPY tsconfig.json tsconfig.app.json tsconfig.node.json vite.config.ts eslint.config.js .prettierrc ./
COPY scripts ./scripts
COPY client ./client
# CI lints; the image build only compiles.
RUN BUILD_CLIENT_SKIP_LINT=true yarn build:client

```
In the Go stage, right after `COPY . .`, add:
```dockerfile
# Embed the web client built in the first stage
COPY --from=client /src/build/client/ ./server/cmd/afc/ui/
```
Nothing else in the Dockerfile changes.

`.github/workflows/client.yaml`:
```yaml
name: client
on: [push, pull_request]

permissions:
  contents: read

jobs:
  client:
    name: lint, typecheck, test, build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-node@v5
        with:
          node-version: '24'
      - run: corepack enable
      - run: yarn install --immutable
      - run: yarn lint:client
      - run: yarn typecheck
      - run: yarn test:client
      - run: BUILD_CLIENT_SKIP_LINT=true yarn build:client
```

`README.md`: append a "Development" section:
```markdown
## Development

Requirements: Go 1.26, Node 24 with corepack (`corepack enable`), and a `.env` (see `example.env`).

- `yarn install` installs the client toolchain.
- `yarn dev` runs Vite and the Go server together. Browse the Go server's address: legacy pages as usual, and the new React client at `/app` with hot reload (the server proxies `/app` to Vite via `AFC_UI_PROXY_URL`).
- `yarn build` lints, type-checks and builds the client, then builds `build/afc` with the client embedded.
- `yarn test` runs the client (Vitest) and server (`go test ./...`) tests. DB-backed Go tests need `AFC_TEST_DB`, a Postgres URL.
- `yarn build:docs` regenerates the swagger docs; `yarn build:emails` recompiles the mjml email templates.

The React client lives in `client/` and is served under `/app` until it replaces the template site.
```

- [ ] **Step 5: Final verification**

```bash
export AFC_TEST_DB='postgres://postgres@localhost:55432/postgres?sslmode=disable'
yarn lint && yarn typecheck && yarn test
AFC_BUILT_UI=$PWD/build/client go test ./server/internal/app/ -run BuiltClient
go vet ./... && gofmt -l server
git status --short   # only intended files; no build/, node_modules/, ui/ contents
```
If Docker is available, also run `docker build -t afc-client-scaffold .` and `docker run --rm --entrypoint ls afc-client-scaffold /bin`. Otherwise say it was skipped. Manual dev check, only if a real `.env` exists: `yarn dev`, open `http://localhost:<port>/app`, edit `HomePage.tsx`, and see it hot-reload.

- [ ] **Step 6: Commit**

```bash
git add package.json yarn.lock scripts .dockerignore Dockerfile .github/workflows/client.yaml README.md \
  server/internal/app/spa_test.go server/internal/emails
git commit -m "Add yarn build/dev pipeline, Docker client stage and client CI

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```
