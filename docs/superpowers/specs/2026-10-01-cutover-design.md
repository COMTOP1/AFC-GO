# Cutover: the React app takes over `/` (sub-project 5)

**Status:** approved in conversation on 2026-10-01.

**Context:** every page now has an in-app version, after sub-projects 4a, 4b, 4c-1 and 4c-2. The React client is still served under `/app`, and the server-rendered classic site still answers at `/`. This sub-project moves the app to `/`, deletes the classic site, and keeps old links working.

**Decisions:**
- **Rollout:** one PR. Rollback means redeploying the previous image. This is safe because the session cookie is already shared and there are no data changes.
- **Unknown URLs:** they get a real **404** status, with the app's own "Page not found" page.

## Goals

- `https://<domain>/` serves the React app. Every app route works at its path without `/app`.
- Classic URLs and `/app/...` bookmarks keep working, either because they are already the same path or through redirects.
- Old `/download?s=…&id=…` file links keep working.
- The classic site's code, templates and static files are gone from the repo.
- Emails show the club logo again, and their links open the app.
- Nobody is signed out by the switch.

## Non-goals

- Database or API changes, apart from moving `/download` within the server.
- Server-side rendering or prerendering for SEO.
- 404 status for detail pages whose item doesn't exist (`/news/99999`). The pattern matches, so the server sends 200 and the app shows its not-found state.
- Deleting `holding/` and `main.go.old`, which are unrelated.

## Server

### `web.MountSPA` (root)

- Served at `/`, alongside the API group at `/api`. API paths are never handled by the SPA handler; `/api/...` 404s stay JSON.
- With a built client (`files` not nil and `index.html` present), the handler checks paths in this order:
  1. **`/assets/*`:** the file, with `Cache-Control: public, max-age=31536000, immutable`. A missing file is a plain 404, never `index.html`.
  2. **Any other path with a file extension** (`/favicon.ico`, `/robots.txt`, `/site.webmanifest`, `/AFC.png`, `/theme-init.js` and so on): the file, with `no-cache`, or a plain 404.
  3. **A path matching a known app route:** `index.html` with **200** and `no-cache`.
  4. **Anything else:** `index.html` with **404** and `no-cache`. The app renders "Page not found".
- **Methods:** only GET and HEAD. Anything else gets 405, as now.
- **Not built:** without a build, every page path answers 503 with the existing "not built" page.
- **Development:** with `UIProxy` set (`yarn dev`), every non-API path is reverse-proxied to Vite at base `/`, including the HMR websocket. `/download` and the redirects below are still answered by Go.
- **Route list:** `web.SPARoutes`, matching Echo-style patterns:
  - `/`
  - `/teams`, `/teams/new`, `/team/:id`, `/team/:id/edit`
  - `/news`, `/news/new`, `/news/:id`, `/news/:id/edit`
  - `/whatson`, `/whatson/new`, `/whatson/:id`, `/whatson/:id/edit`
  - `/gallery`, `/documents`, `/programmes`, `/sponsors`, `/players`, `/users`
  - `/info`, `/info/edit`, `/contact`, `/account`, `/reset/:token`, `/design`

  `:param` matches one non-empty path segment. A trailing slash is already removed by the existing middleware.
- **Drift check:** a Go test reads every `path="…"` in `client/App.tsx`, plus the `index` route as `/`, ignoring `*`. It fails if that set differs from `SPARoutes`.

### Redirects (GET and HEAD only; the query string is kept)

| From | To | Status |
|---|---|---|
| `/app`, `/app/*` | the same path without `/app` (`/app` → `/`) | 301 |
| `/programmes/:id` (numeric) | `/programmes?season=:id` | 301 |
| `/whatson/period/:p` (`future`, `past` or `all`) | `/whatson?period=:p`. Any other value goes to `/whatson` | 301 |
| `/changepassword` | `/account` | 301 |
| `/login`, `/logout`, `/programmeselect`, `/whatsonselect` | `/` | 302 |

`/logout` **does not** sign anyone out: signing out from a GET link would let any site sign people out. In the app, signing out stays a POST to `/api/v1/auth/logout`.

### `/download`

- `GET /download?s=<code>&id=<n>` keeps its current behaviour, moved from the legacy views into `files.Handlers`:
  - it maps the old source code with `files.LegacyKind`;
  - it redirects (302) to the file URL, with `files.CacheControl(kind)`;
  - a bad `id` or unknown `s` gives 400 (plain text);
  - an unknown item gives 404 (plain text).
- The gzip skipper keeps excluding `/download`.

### Errors outside `/api`

`web.ErrorHandler` no longer takes a legacy handler. For non-API paths it writes a small HTML page: status code, "Something went wrong" or "Not found", and a link home. It records the error on the trace, logs 5xx errors, and skips writing when the response is already committed. HEAD requests get the status with no body. API errors are unchanged.

### Deleted

- `server/internal/legacy/`, all of it: routes, views, templates, `assets.go` and `public/`. The `public/` folder holds jQuery, Bulma, Font Awesome, SimpleMDE, pdf.js, the old CSS and images.
- In `app.go`: the views wiring (`views.New`, `legacy.Mount`, `legacyViews`) and its imports, plus anything only it used. Comments that mention the legacy views are updated.
- In `auth.Sessions`: `CookieStore()`, if nothing else uses it once the views are gone. The doc comments that mention the legacy views are updated.
- `.github/workflows/go-template-lint.yaml`.
- `web.SPAPrefix`.
- The Go tests that only cover deleted code.

## Client

- **Paths:**
  - `vite.config.ts`: `base: '/'`;
  - `client/main.tsx`: `<BrowserRouter>` with no basename;
  - `client/index.html`: `/theme-init.js`, `/favicon.png`, `/apple-touch-icon.png`;
  - `DesignPage.tsx`: the example image paths drop `/app`.
- **Copied into `client/public/`** from `server/internal/legacy/public/` before it's deleted:
  - `favicon.ico`, `robots.txt`, `site.webmanifest`, `browserconfig.xml`;
  - the icons they refer to (`android-chrome-192x192.png`, `android-chrome-384x384.png`, `mstile-150x150.png`, `safari-pinned-tab.svg`);
  - `AFC.png`, for the emails.

  Icon paths inside `site.webmanifest` and `browserconfig.xml` become root paths (`/android-chrome-192x192.png` and so on). `index.html` links the manifest.
- **Comments:** comments that mention `/app` or "the classic site" (`resetLink.ts`, `ProgrammePreview.tsx`, `ProgrammesPage.tsx`) are updated. `appResetPath` behaves the same.
- **`/design`:** stays reachable by URL and out of the navigation, as now.

## Emails

- The logo `src` in `signupEmail.tmpl` and `resetEmail.tmpl` becomes `https://{{.Domain}}/AFC.png`, replacing the GitHub `public/AFC.png?raw=true` path, which no longer exists.
- `emails.Reset(to, resetURL)` becomes `emails.Reset(to, resetURL, domain string)`. Its callers (the user service, and anything else) pass their configured domain.
- Reset links stay `https://<domain>/reset/<token>`, which is now the app's own page.

## Tests

- **Go:**
  - **SPA handler:**
    - each route class gets the status and body above, including 404 with `index.html` for unknown paths;
    - asset caching, and missing assets as a plain 404;
    - a file at the root such as `/robots.txt`;
    - `/api/...` is not handled;
    - HEAD, and 405 for POST;
    - not built gives 503;
    - in proxy mode, a non-API path is proxied.
  - **Route drift check** against `client/App.tsx`.
  - **Each redirect:** its target, its status, and that the query string is kept. `/logout` leaves the session cookie alone.
  - **`/download`:** a redirect and cache header for a known item; 400 for a bad id or source; 404 for an unknown item.
  - **Non-API error page:** HTML with the status; HEAD gives no body.
  - **Emails:** both logos point at `https://<domain>/AFC.png`.
  - The existing app-level tests (`server/internal/app`) are updated for root paths.
- **Client:** the existing Vitest suite, including axe, passes unchanged apart from path updates.
- **Build:**
  - `yarn build` produces `build/client/index.html`, which references `/assets/…` and `/theme-init.js`, plus the copied public files;
  - `go build ./...` and `go vet ./...` pass;
  - no Go file imports `internal/legacy`;
  - no client or server file contains an `/app/` path except the redirect code and its tests.

## Docs

- **README:** the app is served at `/`. It lists the redirects and states that the classic site has been removed. Dev instructions are updated.
- **Script and `example.env` text** about `/app` is updated (`scripts/dev-server.js`, `scripts/build-server.js`).
- **The PR description** includes the rollback note.
