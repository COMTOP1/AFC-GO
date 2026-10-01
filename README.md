# AFC-GO
Conversion of AFC from Java to go

## Development

Requirements: Go 1.26, Node 24 with corepack (`corepack enable`), and a `.env` (see `example.env`).

- `yarn install` installs the client toolchain.
- `yarn dev` runs Vite and the Go server together. Browse the Go server's address: the server proxies every page to Vite (via `AFC_UI_PROXY_URL`) for hot reload, and answers `/api` and `/download` itself.
- `yarn build` lints, type-checks and builds the client, then builds `build/afc` with the client embedded.
- `yarn test` runs the client (Vitest) and server (`go test ./server/...`) tests. DB-backed Go tests need `AFC_TEST_DB`, a Postgres URL.
- `yarn build:docs` regenerates the swagger docs; `yarn build:emails` recompiles the mjml email templates.

The React client lives in `client/` and is served at the site root.

### Design system

The client's look lives in `client/styles/app.css`. That file holds the Tailwind v4 theme: colours and fonts as CSS variables, with a `[data-theme='dark']` override. Shared components are in `client/components/ui/`, one file each, and the layout shell is in `client/components/layout/`. Run `yarn dev` and open `/design` to see every component in every state, in light and dark. The FA logo in the header is `client/assets/fa-logo.jpeg`; replace that file to update it.

### Public pages

Every page is in the React client: `/`, `/teams`, `/news`, `/whatson`, `/gallery`, `/documents`, `/programmes`, `/sponsors`, `/info` and `/contact`, with detail pages for teams, articles and events. Editors add, edit and delete content in place: News, What's On, Info and Teams have their own edit pages (with a rich-text editor for articles, events and the club information); documents, programmes and seasons, sponsors, affiliations and gallery photos use dialogs. Signed-in members can browse players at `/players` (editors add, edit and delete them), and user admins manage accounts, password resets and the public contact email at `/users`. The server sends a real 404 status for unknown pages (its route list, `web.SPARoutes`, is checked against `client/App.tsx` by a test).

The classic template site has been removed. Old links still work: `/app/...` redirects to the same path without `/app`, `/programmes/<season>` to `/programmes?season=<season>`, `/whatson/period/<p>` to `/whatson?period=<p>`, `/changepassword` to `/account`, and `/login`, `/logout`, `/programmeselect` and `/whatsonselect` to `/`. Old `/download?s=…&id=…` file links keep redirecting to the file.

### Account pages

`/account` shows the signed-in user's details and lets them change their photo and password. `/reset/<token>` is the password reset page; sign-in sends reset-flagged accounts there, and the reset email links to it.
