# AFC-GO
Conversion of AFC from Java to go

## Development

Requirements: Go 1.26, Node 24 with corepack (`corepack enable`), and a `.env` (see `example.env`).

- `yarn install` installs the client toolchain.
- `yarn dev` runs Vite and the Go server together. Browse the Go server's address: legacy pages as usual, and the new React client at `/app` with hot reload (the server proxies `/app` to Vite via `AFC_UI_PROXY_URL`).
- `yarn build` lints, type-checks and builds the client, then builds `build/afc` with the client embedded.
- `yarn test` runs the client (Vitest) and server (`go test ./server/...`) tests. DB-backed Go tests need `AFC_TEST_DB`, a Postgres URL.
- `yarn build:docs` regenerates the swagger docs; `yarn build:emails` recompiles the mjml email templates.

The React client lives in `client/` and is served under `/app` until it replaces the template site.

### Design system

The client's look lives in `client/styles/app.css`. That file holds the Tailwind v4 theme: colours and fonts as CSS variables, with a `[data-theme='dark']` override. Shared components are in `client/components/ui/`, one file each, and the layout shell is in `client/components/layout/`. Run `yarn dev` and open `/app/design` to see every component in every state, in light and dark. The FA logo in the header is `client/assets/fa-logo.jpeg`; replace that file to update it.

### Public pages

Every public page is now in the React client: `/app`, `/app/teams`, `/app/news`, `/app/whatson`, `/app/gallery`, `/app/documents`, `/app/programmes`, `/app/sponsors`, `/app/info` and `/app/contact`, with detail pages for teams, articles and events. Editors add, edit and delete content in place: News, What's On, Info and Teams have their own edit pages (with a rich-text editor for articles, events and the club information); documents, programmes and seasons, sponsors, affiliations and gallery photos use dialogs. The Players and Users admin pages are still on the classic site.

### Account pages

`/app/account` shows the signed-in user's details and lets them change their photo and password. `/app/reset/<token>` is the password reset page. Sign-in sends reset-flagged accounts there; the admin reset email still links to the classic `/reset/<token>` page until the cutover, when that path becomes the SPA's.
