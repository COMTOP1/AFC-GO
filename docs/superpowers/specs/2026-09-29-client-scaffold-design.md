# Client Scaffold & Build Pipeline Design

Sub-project 2 of 5 in moving AFC-GO to the MV-Controller layout (`/Users/liam/Code/YSTV/MV-Controller`).

| # | Sub-project | Status |
|---|---|---|
| 1 | Server restructure + JSON API | Done (PR #11) |
| 2 | **Client scaffold + build pipeline** (this spec) | Designing |
| 3 | Design system (fresh visual design) | Later |
| 4 | Page ports (public, then logged-in/admin) | Later |
| 5 | Cutover (SPA takes over `/`; delete templates) | Later |

## Context

`main` now has the Go code in `server/` and a JSON API under `/api/v1`: cookie sessions, CSRF via `Sec-Fetch-Site` or the `X-CSRF-Token` double-submit, and swagger at `/api/v1/swagger/`. The legacy template site still owns every page URL (`/`, `/news`, `/team/{id}`, …) plus `/public/*`.

MV-Controller, the reference:
- Yarn 4 with Vite, React 19 and TypeScript; the client lives in `client/`, with `vite.config.ts` and `package.json` at the repo root.
- `vite build` writes to `build/client`. A Node script copies that into `server/cmd/<app>/ui/` before `go build`, and the server embeds it with `//go:embed all:ui`.
- `server/cmd/<app>/ui/.keep` keeps the embed valid when nothing is built.
- In development, a `--ProxyURL` makes the Go server reverse-proxy page requests to the Vite dev server.
- ESLint (typescript-eslint, react, react-hooks, react-refresh, prettier) and Prettier (`printWidth 100`, single quotes).

## Decisions

- **Served under `/app` until cutover.** The SPA lives at `/app/*` (`BrowserRouter basename="/app"`, Vite `base: '/app/'`). Legacy keeps every existing URL. Sub-project 5 drops the prefix and hands `/` to the SPA. The site stays deployable at every step, and ported pages can be tried live.
- **Mirror MV-Controller's toolchain.** Yarn 4 via corepack (`packageManager` field), Vite, React 19, TypeScript, ESLint and Prettier with the same configs (adapted paths), and Node build scripts.
- **Dev workflow: Go proxies to Vite.** `yarn dev` runs Vite on :5173 and the Go server with `AFC_UI_PROXY_URL=http://localhost:5173`. You browse the Go port, and Echo proxies `/app/*` (including the HMR websocket) to Vite. Cookies, CSRF and legacy links behave exactly as in production.
- **Scope: plumbing plus the data layer.** A typed fetch wrapper, TanStack Query, an auth context and one proof page. No UI library, styling or real pages; those are sub-projects 3 and 4.
- **Types are hand-written for now.** Only what the proof page needs. Generating types from `swagger.json` is a possible later step, not part of this spec.
- **mjml becomes a dev dependency** with a `build:emails` script that compiles `server/internal/emails/mjml/*.mjml` into the matching `.tmpl` files. `main` has no committed `package.json` today; the mjml sources were compiled by hand.

## Layout

```
AFC/
├── package.json            Yarn 4; packageManager "yarn@4.x"; scripts below
├── yarn.lock, .yarnrc.yml  nodeLinker: node-modules (as MV-Controller)
├── vite.config.ts          root: client/, base: '/app/', build.outDir: ../build/client, manifest: false
├── tsconfig.json, tsconfig.app.json, tsconfig.node.json
├── eslint.config.js, .prettierrc
├── scripts/                Node ESM scripts (see Build)
├── client/
│   ├── index.html
│   ├── main.tsx            StrictMode → QueryClientProvider → BrowserRouter(basename="/app") → AuthProvider → App
│   ├── App.tsx             <Routes>: "/" → HomePage, "*" → NotFoundPage, wrapped in <Layout>
│   ├── api/
│   │   ├── client.ts       apiFetch<T>(), ApiError
│   │   ├── types.ts        SiteInfo, TeamSummary, CurrentUser, Permissions, ErrorEnvelope
│   │   ├── queryClient.ts  QueryClient with the retry policy
│   │   └── queries.ts      useSite(), useMe()
│   ├── auth/
│   │   └── AuthProvider.tsx  useAuth(): { user: CurrentUser | null, isLoading, refresh }
│   ├── components/
│   │   └── Layout.tsx      bare semantic shell (header/main/footer); no styling decisions
│   ├── pages/
│   │   ├── HomePage.tsx    proof page (see below)
│   │   └── NotFoundPage.tsx
│   └── test/setup.ts       Vitest + Testing Library setup
└── server/
    ├── cmd/afc/ui/.keep    embed target (ui/* gitignored except .keep)
    ├── cmd/afc/ui.go       //go:embed all:ui → passed to app.Config
    └── internal/web/spa.go  SPA handler: embedded files, or proxy to Vite
```

## Serving `/app` (`web.SPA`)

`web.MountSPA(e *echo.Echo, files fs.FS, proxyURL string)` is called from `app.Build`. `files` is the `ui` sub-directory of the embedded FS, and `proxyURL` comes from `AFC_UI_PROXY_URL`, where empty means serve the embedded files.

**Embedded mode:**

| Request | Response |
|---|---|
| `GET /app`, `GET /app/<anything not under assets>` | `index.html`, `200`, `Cache-Control: no-cache` |
| `GET /app/assets/<existing file>` | the file, correct `Content-Type`, `Cache-Control: public, max-age=31536000, immutable` |
| `GET /app/assets/<missing file>` | `404` (never the index fallback, so broken deploys are visible) |
| any of the above when `ui/` has no `index.html` | `503` with a small HTML page: "The web client has not been built. Run `yarn build:client`." |
| non-GET/HEAD under `/app` | `405` |

Other static files Vite copies from `client/public/` (e.g. `favicon.svg`) sit at the build root. `GET /app/<file with an extension that exists in ui/>` serves the file with `Cache-Control: no-cache`; a path with an extension that doesn't exist is a `404`.

**Proxy mode** (`AFC_UI_PROXY_URL` set): every `/app` and `/app/*` request is reverse-proxied to that URL, path unchanged, including websocket upgrades (Vite HMR). A startup log line says proxy mode is on.

`RemoveTrailingSlash` already runs first (`/app/` → `/app`). Gzip, OTel and visitor counting apply as for other routes. `/api`, `/public` and the legacy routes are unaffected; `/app` doesn't collide with any legacy route.

## Data layer

`client/api/client.ts`:
- `apiFetch<T>(path: string, init?: { method?, json?, form?: FormData, signal? }): Promise<T>`
  - Base URL is `/api/v1`, with `credentials: 'same-origin'`.
  - A `json` body is sent as `application/json`. A `form` body is sent as `FormData`; the browser sets the multipart boundary.
  - `2xx` returns the parsed JSON, or `undefined` for `204`.
  - Non-2xx throws `ApiError` with `status`, `message` and `fields` taken from `{error:{code,message,fields}}`. If the body isn't the envelope, `message` is the HTTP status text.
  - A network failure throws `ApiError` with `status: 0`.
  - CSRF: browsers send `Sec-Fetch-Site: same-origin`, which the server accepts, so no token handling is needed.

`client/api/queryClient.ts`: `retry` is off for `ApiError` with `400 ≤ status < 500`, up to 2 retries otherwise; `staleTime` 30 s.

`client/api/queries.ts`:
- `useSite()` → `GET /site` → `SiteInfo`.
- `useMe()` → `GET /auth/me` → `CurrentUser | null`. A `401` resolves to `null` (anonymous); it is not an error.

`client/auth/AuthProvider.tsx`: wraps `useMe()`. `useAuth()` returns `{ user, isLoading, refresh }`, where `refresh` invalidates the `me` query. Login and logout UI come in sub-project 4.

`client/pages/HomePage.tsx` is the proof page. Unstyled, it shows:
- "Signed in as <name> (<role>)" or "Not signed in";
- the visitor count;
- the list of team names from `/site`;
- a loading state and an error state that shows `ApiError.message`.

## Build and scripts

`package.json` scripts, following MV-Controller:

| Script | Does |
|---|---|
| `dev` | runs `dev:client` and `dev:server` together (`concurrently`) |
| `dev:client` | `vite --host` |
| `dev:server` | `node scripts/dev-server.js` → `go run ./server/cmd/afc` with `AFC_UI_PROXY_URL=http://localhost:5173` (reads `.env` as today) |
| `build` | `clean` → `build:client` → `build:server` |
| `build:client` | `tsc -b && vite build` (after `lint:client`, unless `BUILD_CLIENT_SKIP_LINT=true`) |
| `build:server` | copy `build/client` → `server/cmd/afc/ui`, `go build -o build/afc ./server/cmd/afc`, then restore `ui/` to just `.keep` |
| `build:docs` | `go generate ./server/internal/docs` |
| `build:emails` | compile each `server/internal/emails/mjml/<name>.mjml` to `server/internal/emails/<name>.tmpl` |
| `lint`, `lint:client`, `lint:server` | `eslint .`; `golangci-lint run ./...` |
| `typecheck` | `tsc -b --noEmit` |
| `test`, `test:client`, `test:server` | `vitest run`; `go test ./...` |
| `clean` | remove `build/`, reset `server/cmd/afc/ui` to `.keep` |

**Dockerfile:** a new first stage, `node:24-alpine` with corepack, runs `yarn install --immutable` and `yarn build:client` (with `BUILD_CLIENT_SKIP_LINT=true`, because CI lints). The Go stage copies `build/client/` into `server/cmd/afc/ui/` before its existing `go build`. The ldflags, `migrates3` binary and final image stay the same. **Jenkinsfile:** unchanged (`docker build .`).

**`.dockerignore` (new):** `node_modules`, `build`, `.yarn/cache`, `.git`, `.env*`, `*.sql`, `FileStore`, so the Docker context stays small and never ships local secrets or dumps.

**`.gitignore`:** add `node_modules/`, `build/`, `server/cmd/afc/ui/*` with `!server/cmd/afc/ui/.keep`, and `.yarn/*` with the usual Yarn 4 exceptions.

**CI:** a new `.github/workflows/client.yaml` on push and pull request: setup-node with corepack, `yarn install --immutable`, `yarn lint:client`, `yarn typecheck`, `yarn test:client`, `yarn build:client`. The Go workflows are unchanged and pass with only `.keep` in `ui/`.

## Testing

- **Go (`server/internal/web/spa_test.go`):** tests use `fstest.MapFS` and cover:
  - index fallback for `/app` and a deep client route;
  - asset hit with immutable cache;
  - asset miss → 404;
  - root static file with an extension;
  - the no-build 503 page;
  - 405 for POST;
  - proxy mode forwarding path and body to an `httptest` server.

  The existing legacy smoke test and route-walk tests stay green (`/app` is not under `/api/v1`).
- **Client (Vitest + Testing Library + jsdom):** tests cover:
  - `apiFetch`: JSON success, 204, error envelope with fields, a non-envelope error, and network failure;
  - `useMe` treating a 401 as `null`;
  - `HomePage` rendering signed-in and anonymous states from a mocked `fetch`;
  - `NotFoundPage` for an unknown route.
- **Manual:** `yarn build` produces `build/afc`. Running it with a real `.env` serves `/app` from the embedded build, and legacy pages are unchanged. `yarn dev` gives HMR at `http://localhost:<ADDRESS port>/app`.

## Out of scope

- UI library, styling, fonts, icons and components (sub-project 3).
- Real pages, navigation, the login form and admin screens (sub-project 4).
- Moving the SPA to `/`, redirects from legacy URLs, and deleting templates (sub-project 5).
- Generating types from swagger, SSR, and SEO work.
