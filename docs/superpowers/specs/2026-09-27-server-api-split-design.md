# Server Restructure & JSON API Design

Sub-project 1 of 5 in moving AFC-GO to the MV-Controller layout
(`/Users/liam/Code/YSTV/MV-Controller`): a Go backend in `server/` exposing a JSON API, and a
React/Vite frontend in `client/` embedded into the Go binary.

| # | Sub-project | Status |
|---|---|---|
| 1 | **Server restructure + JSON API** (this spec) | Designing |
| 2 | Client scaffold + build pipeline (Vite, React, TS, BrowserRouter, `go:embed`, dev proxy, Docker/Jenkins) | Later |
| 3 | Design system (fresh visual design: palette, type, shell, shared components) | Later |
| 4 | Page ports (public pages, then logged-in/admin pages) | Later |
| 5 | Cutover (delete templates, legacy views, jQuery/Bulma assets) | Later |

## Context

AFC-GO is a server-rendered Echo app. `router.go` maps ~60 routes to methods on a single
`views.Views` struct (~5,000 lines across `views/*.go`), each of which fetches data, checks
permissions, validates form input, uploads files and renders one of 25 Go templates
(~5,000 lines in `templates/`). Frontend behaviour lives in jQuery, Bulma and small scripts
in `public/`. A few handlers (login, display email) already reply with JSON to AJAX calls.

Domain data access is already separated: each of `affiliation`, `document`, `image`, `news`,
`player`, `programme`, `setting`, `sponsor`, `team`, `user`, `whatson` has a `db.go` with a
concrete `*Store` over `sqlx`/`squirrel`. Infrastructure (`db`, `mail`, `storage` (S3),
`telemetry` (OTel)) lives under `infrastructure/`.

MV-Controller's layout, which this adopts:

- `go.mod` at the repo root; Go code under `server/cmd/<binary>/` and `server/internal/...`.
- Per-domain packages under `server/internal/<domain>/` containing routes/handlers, DB access and
  structs together.
- A shared `server/internal/web/` for the router, JSON helpers and static hosting.
- JSON API under `/api/v1`, documented with swaggo into `server/internal/docs/`, with `/api` and
  `/api/v1` redirecting to the swagger UI.

## Decisions

- **Full rewrite target**: all pages eventually move to React (sub-project 4); this spec only
  covers the backend.
- **Keep Echo.** `otelecho` tracing and existing middleware are retained. MV-Controller uses chi;
  only its *layout* is copied, not its router.
- **Branch**: `server-api-split`, off `main` (which already contains the OTel work).
- **Shared service layer.** Business logic is extracted from `views/` into per-domain services.
  Both the new API handlers and the existing template views call these services, so the live
  site keeps working and remains deployable after every step. Templates are deleted in
  sub-project 5.
- **Vertical slices** (one package per domain containing store, service and handlers), matching
  MV-Controller.
- **Cookie session auth is kept.** The React app will be served same-origin (sub-project 2), so
  the existing gorilla `sessions.CookieStore` works unchanged. No tokens/JWT.
- **No DB schema changes.**
- **Client routing will use BrowserRouter** (sub-project 2). That's why the API lives
  exclusively under `/api/v1`: every other path can later fall through to the SPA's
  `index.html`.

## Directory layout

```
AFC/
├── go.mod, go.sum                  # stay at root; module path github.com/COMTOP1/AFC-GO unchanged
├── Dockerfile, Jenkinsfile          # build path updated to ./server/cmd/afc
├── holding/                         # unchanged (separate Go module)
├── docs/                            # unchanged
└── server/
    ├── cmd/
    │   ├── afc/main.go              # from main.go; env/config loading unchanged, now wires services
    │   ├── migrates3/main.go        # from cmd/migrates3
    │   └── offline/                 # from offline/ (standalone binary + its templates/public)
    └── internal/
        ├── web/
        │   ├── router.go            # Echo setup, global middleware, mounts legacy + api
        │   ├── session.go           # CookieStore construction, current-user load/save
        │   ├── auth.go              # RequireLogin / role guards, API (JSON 401/403) + legacy (redirect) variants
        │   ├── csrf.go              # CSRF config for /api
        │   ├── errors.go            # JSON error envelope + svcerr → HTTP mapping
        │   ├── json.go              # bind/write helpers
        │   └── static.go            # serves embedded public/ (SPA hosting added in sub-project 2)
        ├── svcerr/                  # typed service errors (NotFound, Forbidden, Invalid, Conflict)
        ├── affiliation/ document/ image/ news/ player/ programme/
        │   setting/ sponsor/ team/ user/ whatson/
        │   ├── store.go             # today's <domain>/db.go, unchanged apart from package move
        │   ├── <domain>.go          # today's model types, unchanged
        │   ├── service.go           # logic extracted from views/*
        │   ├── types.go             # API response/input types
        │   ├── handlers.go          # /api/v1 handlers with swag annotations
        │   └── *_test.go
        ├── auth/                    # login, logout, change password, reset flow, reset tokens (Redis or in-process cache)
        ├── account/                 # the logged-in user's own profile + image
        ├── files/                   # GET /files/{kind}/{id} (replaces /download)
        ├── site/                    # /site, /home, /contact aggregate endpoints
        ├── visitors/                # visitor tracking middleware + count flusher (from views.go)
        ├── role/ utils/
        ├── infrastructure/{db,mail,storage,telemetry}/
        ├── docs/                    # swag output
        └── legacy/
            ├── views/               # today's views/, rewritten to call services
            ├── templates/           # today's templates/
            └── public/              # today's public/ (embedded)
```

Notes:

- The root `middleware/` package is not imported anywhere (its CORS config is dead code). It is
  deleted rather than moved; no CORS is needed because the client will be same-origin.
- `main.go.old`, `test/` (gitignored) and the root mjml `package.json` are left alone here.
  `package.json` gets replaced by the Vite one in sub-project 2.
- `views.Views` stops being a god-struct. `server/cmd/afc/main.go` constructs each store and
  service once and injects them into both `legacy/views` and the API handler groups.

## API

### Conventions

- **Base**: `/api/v1`. REST with JSON request bodies; `multipart/form-data` wherever a file is
  uploaded (fields plus a `file` part).
- **Success**: the resource itself, e.g. an object or array, with no wrapper. `201` for creates
  (with the created resource), `204` for deletes.
- **Errors**: always
  ```json
  { "error": { "code": 422, "message": "validation failed", "fields": { "email": "invalid email" } } }
  ```
  `fields` appears only on 422. One Echo `HTTPErrorHandler` renders this for any path under
  `/api/`; other paths keep the legacy HTML error page. Internal error detail is logged via
  `slog` and recorded on the OTel span, never returned to the client. 5xx messages are generic.
- **Response types** replace the `*Template` view structs in `views/helpers.go`:
  - Dates and times are RFC 3339 strings (`time.Time` marshalled). Clients format them.
  - File fields are exposed as ready-to-use URLs (`imageUrl`, `fileUrl`) built by
    `storage.Store`, and are omitted when there is no file.
  - IDs are numbers; foreign keys are included, and embedded summaries are added where a page
    needs them (e.g. a player's `team: {id, name}`).
  - `user.User` is never serialised directly. A `user.Public` / `user.Admin` projection controls
    which fields are exposed; password hashes, salts and reset flags are never included.
- **HTML content** (news/what's on bodies, info page) is sanitised by the service with the
  existing bluemonday policy on write, and returned as an HTML string.

### Auth, sessions and CSRF

- The existing encrypted cookie session (`WAUTH_SESSION_COOKIE_NAME`) is kept, with the same keys,
  so logged-in users stay logged in across the deploy.
- The cookie gains `SameSite=Lax`, and `Secure` when `DOMAIN_NAME` is not `localhost`.
- The API guards mirror today's:

  | Guard | Allows | Used by |
  |---|---|---|
  | `RequireLogin` | any authenticated user | account, players list, change password, logout |
  | `RequireEditor` | logged in, not Manager, not Photographer | news, whatson, teams, players, programmes, seasons, sponsors, affiliations, documents, info writes |
  | `RequireNotManager` | logged in, not Manager | gallery writes |
  | `RequireClubSecretaryHigher` | SafeguardingOfficer, ClubSecretary, Chairperson, Webmaster | users, display email |

  `RequireUserManagement` exists today but is not wired to any route. It is carried over
  unwired.
- API guards return JSON `401` (not logged in) or `403` (wrong role). Legacy guards keep
  redirecting to `/`.
- **CSRF**: Echo's `CSRFWithConfig` on `/api/v1`, applied to unsafe methods only. The token is
  exposed in a readable `_csrf` cookie (`SameSite=Strict`) and must be sent back in the
  `X-CSRF-Token` header. Missing or invalid tokens get `403`. Legacy form routes are unchanged.
- Session flash messages (`Context.Message`/`MsgType`) are not used by the API. Clients show
  feedback from responses. Legacy views keep them.
- Login keeps the forced-reset behaviour: when `VerifyUser` reports a reset is required, the
  response is `200 {"resetRequired": true, "resetUrl": "/reset/<token>"}` and no session is
  created.

### Endpoints

🔓 = public, 🔒 = `RequireLogin`, ✏️ = `RequireEditor`, 📷 = `RequireNotManager`,
🛡 = `RequireClubSecretaryHigher`.

| Method & path | Guard | Replaces |
|---|---|---|
| `GET /health` | 🔓 | `/api/health` (kept as alias, untraced) |
| `GET /site` | 🔓 | shared layout data: nav teams, displayEmail, visitorCount, year, version |
| `GET /home` | 🔓 | `HomeFunc`: latest news, latest what's on, sponsors, affiliations |
| `GET /contact` | 🔓 | `ContactFunc` (read-only; there is no form submission today) |
| `POST /auth/login` | 🔓 | `LoginFunc` (`{email, password, remember}`) |
| `POST /auth/logout` | 🔒 | `LogoutFunc` |
| `GET /auth/me` | 🔓 | current user + role, or `401` |
| `POST /auth/password` | 🔒 | `ChangePasswordFunc` |
| `GET /auth/reset/{token}` | 🔓 | `ResetURLFunc` GET (token validity) |
| `POST /auth/reset/{token}` | 🔓 | `ResetURLFunc` POST |
| `GET /account` / `PATCH /account` | 🔒 | `AccountFunc` |
| `PUT /account/image` / `DELETE /account/image` | 🔒 | `UploadImageFunc` / `RemoveImageFunc` |
| `GET /news` / `POST /news` | 🔓 / ✏️ | `NewsFunc` / `NewsAddFunc` |
| `GET /news/{id}` / `PATCH` / `DELETE` | 🔓 / ✏️ / ✏️ | `NewsArticleFunc` / `NewsEditFunc` / `NewsDeleteFunc` |
| `GET /whatson?period=all\|future\|past` / `POST /whatson` | 🔓 / ✏️ | `WhatsOnFunc`, `WhatsOnTomePeriodFunc` / `WhatsOnAddFunc` |
| `GET /whatson/{id}` / `PATCH` / `DELETE` | 🔓 / ✏️ / ✏️ | article / edit / delete |
| `GET /teams` / `POST /teams` | 🔓 / ✏️ | `TeamsFunc` / `TeamAddFunc` |
| `GET /teams/{id}` / `PATCH` / `DELETE` | 🔓 / ✏️ / ✏️ | `TeamFunc` (includes managers and players, respecting today's youth-team hiding rules) / edit / delete |
| `GET /players` / `POST /players` | 🔒 / ✏️ | `PlayersFunc` / `PlayerAddFunc` |
| `PATCH /players/{id}` / `DELETE` | ✏️ | edit / delete |
| `GET /programmes?season={id}` / `POST /programmes` | 🔓 / ✏️ | `ProgrammesFunc`, `ProgrammesSeasonsFunc` / `ProgrammeAddFunc` |
| `DELETE /programmes/{id}` | ✏️ | `ProgrammeDeleteFunc` |
| `GET /seasons` / `POST /seasons` | 🔓 / ✏️ | season list / `ProgrammeSeasonAddFunc` |
| `PATCH /seasons/{id}` / `DELETE` | ✏️ | edit / delete |
| `GET /sponsors` / `POST /sponsors` / `DELETE /sponsors/{id}` | 🔓 / ✏️ / ✏️ | sponsors |
| `GET /affiliations` / `POST` / `DELETE /affiliations/{id}` | 🔓 / ✏️ / ✏️ | affiliations |
| `GET /documents` / `POST` / `DELETE /documents/{id}` | 🔓 / ✏️ / ✏️ | documents |
| `GET /gallery` / `POST` / `DELETE /gallery/{id}` | 🔓 / 📷 / 📷 | gallery / image add / delete |
| `GET /info` / `PUT /info` | 🔓 / ✏️ | `InfoFunc` / `InfoEditFunc` |
| `GET /users` / `POST /users` | 🛡 | `UsersFunc` / `UserAddFunc` |
| `GET /users/{id}` / `PATCH` / `DELETE` | 🛡 | edit / delete |
| `POST /users/{id}/reset` | 🛡 | `ResetUserPasswordFunc` |
| `PUT /settings/display-email` | 🛡 | `UsersSetDisplayEmailFunc` |
| `GET /files/{kind}/{id}` | 🔓 | `DownloadFunc` (`/download?s=&id=` stays as a legacy alias) |

`kind` ∈ `affiliation, document, gallery, news, player, programme, sponsor, team, user, whatson`.
`/files` behaves exactly like `_downloadFunc` today (redirect to the object's CDN URL), including
any existing access checks for kinds that need them.

`programmeselect` and `whatsonselect` are form-redirect helpers for the template UI and have no
API equivalent; query parameters replace them.

The exact request and response fields for each endpoint are defined by the `types.go` structs
and published through swagger. Each struct is derived from the fields its template uses today,
so nothing a page currently shows is lost.

### Documentation

`swag init` (run with `go tool swag`, added as a tool dependency as in MV-Controller) generates
into `server/internal/docs`. The swagger UI is served at `/api/v1/swagger/*`, and `GET /api` and
`GET /api/v1` redirect to it.

## Services

Each domain gets:

```go
// server/internal/news/service.go
type store interface { // satisfied by *Store
    GetNews(ctx context.Context) ([]News, error)
    GetNewsArticle(ctx context.Context, n News) (News, error)
    AddNews(ctx context.Context, n News) (News, error)
    // ...only the methods this service calls
}

type Service struct {
    store store
    files *storage.Store
    now   func() time.Time
}

func (s *Service) List(ctx context.Context) ([]Article, error)
func (s *Service) Get(ctx context.Context, id int) (Article, error)
func (s *Service) Create(ctx context.Context, actor user.User, in CreateInput, image *multipart.FileHeader) (Article, error)
func (s *Service) Update(ctx context.Context, actor user.User, id int, in UpdateInput, image *multipart.FileHeader) (Article, error)
func (s *Service) Delete(ctx context.Context, actor user.User, id int) error
```

Rules:

- **No `echo.Context` in services.** Inputs are plain structs. API handlers bind JSON or
  multipart into them; legacy views bind form values into them.
- **Services own validation and side effects**: password rules (`minRequirementsMet`), email
  verification (`emailverifier`), bluemonday sanitising, the content-type allow-list plus the S3
  upload (today's `fileUpload`; size stays enforced by the global 15MB `BodyLimit`), deleting old objects on replace or delete, and sending the
  password-reset email. When no mailer is configured, today's fallback of returning the reset
  link to the admin is kept (`POST /users/{id}/reset` responds `{"emailSent": false, "resetUrl": ...}`).
  The `signupEmail` template exists but no code sends it; that stays as it is.
- **Authorisation**: route guards remain the only authorisation, exactly as today. Behaviour is
  preserved, not tightened: `views/user.go` has no self-delete or role-hierarchy checks, and
  none are added in this sub-project.
- **Typed errors** from `svcerr`: `NotFound`, `Forbidden`, `Invalid{Fields map[string]string}`,
  `Conflict`, wrapping the cause. `web/errors.go` maps them to 404/403/422/409; anything else
  becomes 500. Legacy views map them onto today's `v.error(code, msg, err)`.
- **Cross-domain reads are injected explicitly**, e.g. `team.Service` gets `player` and `user`
  store interfaces for managers and squads; `site.Service` aggregates news, whatson, sponsor,
  affiliation, team and setting.
- **Tracing**: the `tracer.Start(...)` spans currently opened in each view move into service
  methods, named `<domain>.Service.<Method>`. `otelecho` continues to create the request span,
  so trace depth is preserved.
- `templates.Templater`'s `getTeamName` func (a DB call made from inside a template) moves to
  the services that need team names, so API responses include them.

## Request flow

```
Echo → RemoveTrailingSlash → Recover → otelecho → BodyLimit(15M) → Gzip (skips /download, /api/v1/files)
     → visitors.Middleware
     → /api/v1:  CSRF → [guard] → handler → service → store / S3 / mail → JSON
     → legacy:   [guard] → view → service → store / S3 / mail → template
```

## Migration order

Each step compiles, passes lint and tests, and leaves the site deployable.

1. **Move only.** `git mv` everything into the new layout, rename `db.go` → `store.go`, fix
   imports, update Dockerfile / Jenkinsfile / workflows for `./server/cmd/afc`. No behaviour change.
   Verify with the legacy smoke test (below), which is written as the first task.
2. **Web foundation.** `svcerr`, `web/errors.go`, `web/json.go`, `web/session.go`,
   `web/auth.go`, CSRF, `/api/v1/health`, `/api/v1/site`, `/api/v1/auth/me`, swagger wiring.
3. **Domains, one at a time.** For each: write service tests → extract service → point the legacy
   view at it → add API handlers and handler tests → regenerate swagger. The order goes from
   simplest to most coupled:
   news → whatson → sponsor → affiliation → document → gallery (image) → programme + seasons →
   team → player → info + settings → site (home/contact) → files → users → auth + account.
4. **Clean up.** Delete the root `middleware/` package and any now-empty `views` helpers.
   Confirm every route in the old `router.go` has either an API equivalent or a documented reason
   not to (the select helpers).

## Testing

- **Legacy smoke test (written first)**: boots the real router against a Postgres loaded from
  a fixture and asserts that the key public pages (`/`, `/news`, `/news/{id}`,
  `/whatson`, `/teams`, `/team/{id}`, `/sponsors`, `/gallery`, `/documents`, `/programmes`,
  `/info`, `/contact`) return `200`, and that guarded pages redirect when logged out. It runs
  when `AFC_TEST_DB` is set (Docker Postgres) and is skipped otherwise. It guards against
  regressions throughout the move.
  The repo has no schema file today. The fixture is `server/internal/testdata/schema.sql`
  (DDL only, extracted from a production `pg_dump --schema-only`) plus a hand-written
  `seed.sql` with synthetic rows. Real database dumps contain member personal data and are never
  committed.
- **Service unit tests**: table-driven, using hand-written fake stores (no mocking library),
  covering validation errors, permission decisions, sanitising, and file replace/delete calling
  storage. `storage.Store` sits behind a small interface so it can be faked.
- **Handler tests**: `httptest` through the real `web` router with fake services, covering
  status codes, JSON shape, the error envelope, `401`/`403` on every guarded route (driven from
  a table of routes and guards), and CSRF rejection of unsafe methods without a token.
- **Existing**: `infrastructure/storage` tests move with the package. `golangci-lint` (repo
  `.golangci.yml`), `nilaway` and `go-template-lint` workflows keep passing, with their paths
  updated.

## Out of scope

- Anything in `client/`, Vite, React, and embedding a SPA (sub-project 2).
- Visual design (sub-project 3) and page ports (sub-project 4).
- Deleting templates, legacy views or `public/` (sub-project 5).
- DB schema changes, auth model changes (no JWT/OAuth), pagination (lists are small today and
  unpaginated).
- `holding/` and `main.go.old`.
