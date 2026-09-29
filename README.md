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
