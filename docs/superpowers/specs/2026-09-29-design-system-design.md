# Design System Design

Sub-project 3 of 5 in moving AFC-GO to the MV-Controller layout.

| # | Sub-project | Status |
|---|---|---|
| 1 | Server restructure + JSON API | Done (PR #11) |
| 2 | Client scaffold + build pipeline | Done (PR #12) |
| 3 | **Design system** (this spec) | Designing |
| 4 | Page ports (public, then logged-in/admin) | Later |
| 5 | Cutover (SPA takes over `/`; delete templates) | Later |

## Context

`main` serves a React client at `/app`: React 19, Vite 8, TypeScript 6, React Router 7 and TanStack Query 5. It has typed `apiFetch`, `useSite`/`useMe` and `AuthProvider`/`useAuth`. The shell (`client/components/Layout.tsx`) is bare and unstyled, and there is no UI library.

The legacy site uses the following look:
- a Bulma red navbar;
- Arial and Allerta fonts;
- near-black text `#1F1F1F` and dark red `#d30e27`;
- club theme colour `#fc0f2b`;
- a masthead with the AFC crest on the left and the FA Charter Standard logo on the right (`server/internal/legacy/public/AFC-256.png`, `facs.jpeg`).

The crest is red and blue on white, with "Facta Non Verba" and "Founded 1952".

This sub-project gives the client its visual identity and the shared building blocks that sub-project 4's pages are built from. The mockups agreed during brainstorming live in `.superpowers/brainstorm/` (untracked).

## Decisions

- **Crest-led direction.** Club red is the accent and action colour, crest blue is the secondary, and the page is mostly white with near-black text.
- **Tailwind CSS v4.** Components use utility classes, with a typed variant map per component. Design tokens live in Tailwind's `@theme` as CSS variables. `clsx` is the only helper library; there is no `cva` or `tailwind-variants`.
- **No headless UI library.** Modals use native `<dialog>`. Menus, the mobile nav and toasts use small in-house hooks.
- **Dark mode follows the system, with a toggle.** The setting is `system` (the default), `light` or `dark`, is remembered per browser, and is applied before first paint.
- **Typography: Barlow Condensed and Inter.** Headings use Barlow Condensed 600–800, uppercase. Body text, forms and tables use Inter (variable). Both are self-hosted via Fontsource, with no request to Google.
- **Header, based on option C of the mockups.** The masthead has the crest on the left, a large centred "AFC ALDERMASTON" with "Facta Non Verba · Founded 1952" below it, and the FA logo on the right. A full-width nav bar sits underneath with its links right-aligned.
- **News and What's On cards** show the uploaded image when there is one, and a blue→red gradient otherwise.
- **Nav links point at legacy pages until they are ported.** Each nav entry names its legacy URL and, once sub-project 4 ports it, an SPA route.
- **Sign-in and sign-out belong to the shell,** like the legacy login and logout modals. The reset and account pages stay in sub-project 4.

## Foundations

### Tailwind and tokens

- Add `tailwindcss` and `@tailwindcss/vite`, registered in `vite.config.ts`.
- `client/styles/app.css` is the one stylesheet, imported by `main.tsx`. It contains, in order:
  - `@import 'tailwindcss'`;
  - the Fontsource imports;
  - `@custom-variant dark (&:where([data-theme=dark], [data-theme=dark] *));`
  - `@theme` holding the light token values;
  - a `[data-theme='dark']` block overriding the colour variables;
  - base styles: body font and colour, the focus ring, and reduced motion.
- Components use token classes only (`bg-surface`, `text-muted`, `border-line`, `text-red`), never raw hex values.

| Token | Light | Dark | Use |
|---|---|---|---|
| `--color-red` | `#d30e27` | `#ff4d5e` | Buttons, links, active nav, errors. White text on light red is 5.4:1. |
| `--color-red-hover` | `#b10c20` | `#ff6b79` | Hover and pressed states |
| `--color-club-red` | `#fc0f2b` | `#fc0f2b` | Decoration only (rules, accents); never behind text |
| `--color-blue` | `#233a8f` | `#8ea2f0` | Secondary: scores, badges, focus ring |
| `--color-footer` | `#233a8f` | `#1b2550` | Footer background (white text) |
| `--color-ink` | `#1f1f1f` | `#eceef2` | Body text |
| `--color-muted` | `#5b6170` | `#a2a8b5` | Secondary text |
| `--color-line` | `#e4e6eb` | `#2a2e37` | Borders and dividers |
| `--color-bg` | `#ffffff` | `#121419` | Page background |
| `--color-surface` | `#f5f6f8` | `#1a1d24` | Bands, hovered rows, skeletons |
| `--color-field` | `#ffffff` | `#0e1015` | Input backgrounds |
| `--color-success` | `#1f8a4c` | `#4ade80` | Success notices |
| `--color-warning` | `#b45309` | `#fbbf24` | Warning notices |
| `--color-on-red` | `#ffffff` | `#1a0005` | Text on a red background |
| `--font-display` | `'Barlow Condensed', sans-serif` | same | Headings |
| `--font-sans` | `'Inter Variable', system-ui, sans-serif` | same | Everything else |

Radii are Tailwind's defaults: `rounded-md` for buttons and inputs, `rounded-lg` for cards and notices, and `rounded-xl` for dialogs.

### Dark mode

- `client/theme/ThemeProvider.tsx` exposes `useTheme(): { setting: 'system' | 'light' | 'dark', resolved: 'light' | 'dark', setSetting }`.
  - It stores the setting in `localStorage` under `afc-theme`. Every read and write is wrapped in `try/catch`, and it falls back to `system`.
  - It sets `document.documentElement.dataset.theme` to the resolved value.
  - While the setting is `system`, it follows `prefers-color-scheme` live via a `matchMedia` change listener.
- `client/public/theme-init.js` is a small synchronous script loaded in `<head>` before the stylesheet. It uses the same key and rules and sets `data-theme` before first paint, so a dark-mode visitor never sees a white flash. It lives in its own file, rather than inline, so that a future Content-Security-Policy doesn't need a hash.
- `<meta name="color-scheme" content="light dark">` and `<meta name="theme-color" content="#fc0f2b">` go in `client/index.html`.

### Fonts

- `@fontsource-variable/inter` provides Inter.
- `@fontsource/barlow-condensed` provides Barlow Condensed at weights 600, 700 and 800, Latin subset.
- Vite bundles the font files into `/app/assets/`, where they get the immutable cache headers.

### Brand assets

- `client/assets/crest.png` is a copy of `AFC-256.png`, and `client/assets/fa-logo.jpeg` is a copy of `facs.jpeg`. Both are imported from components, so their filenames are hashed.
- Replacing the FA logo means replacing `client/assets/fa-logo.jpeg`; no code changes.
- The favicon becomes the crest: `client/public/favicon.png` (32×32) and `client/public/apple-touch-icon.png` (180×180), both taken from the legacy favicon set. `client/public/favicon.svg` (the placeholder) is deleted.
- The legacy public files stay where they are, because the legacy site still uses them.

### Accessibility baseline

- A 3px focus-visible outline in the blue token, with an offset.
- A "Skip to content" link, visible when focused, targets `<main id="content" tabindex="-1">`.
- Motion (the mobile menu and toasts) is disabled under `prefers-reduced-motion`.
- Colour pairs were chosen for WCAG AA. Club red is never behind text.

## Layout shell (`client/components/layout/`)

`Layout.tsx` renders `SkipLink`, `Masthead`, `NavBar`, `<main id="content">` holding `<Outlet/>` inside a `Container`, and then `Footer`.

### Masthead

- A three-column grid: the crest (96px; 52px on phones), a centred title block, and the FA logo (84px tall; 46px on phones).
- The title is an `<h1>`-styled "AFC ALDERMASTON" in `font-display`, extra-bold and uppercase. Below it is "Facta Non Verba · Founded 1952", uppercase with wide letter spacing, in the muted colour. On phones it shortens to "Facta Non Verba · 1952".
- The masthead title is a `<p>` styled as a heading, because each page owns its own `<h1>` via `PageHeader`.
- The crest links to `/app`, with alt text "AFC Aldermaston home". The FA logo's alt text is "The FA Charter Standard".
- In dark mode both logos sit on white rounded tiles, because the images have white backgrounds.

### NavBar

- A full-width strip with a top border in the line colour and a 3px bottom border in the red token. Its contents are right-aligned.
- Links come from `client/components/layout/navItems.ts`:

  ```ts
  export interface NavItem { label: string; legacyHref: string; to?: string }
  export const navItems: NavItem[] = [
    { label: 'Home', legacyHref: '/', to: '/' },
    { label: 'Teams', legacyHref: '/teams' },
    { label: 'News', legacyHref: '/news' },
    { label: "What's On", legacyHref: '/whatson' },
    { label: 'Gallery', legacyHref: '/gallery' },
    { label: 'Documents', legacyHref: '/documents' },
    { label: 'Programmes', legacyHref: '/programmes' },
    { label: 'Sponsors', legacyHref: '/sponsors' },
    { label: 'Info', legacyHref: '/info' },
    { label: 'Contact', legacyHref: '/contact' },
  ];
  ```

  An item with `to` renders a React Router `NavLink`, which is relative to the `/app` basename; the active one gets `aria-current="page"`, red text and a 3px red underline. An item without `to` renders a plain `<a href={legacyHref}>`, a full page load into the legacy site. When sub-project 4 ports a page, it adds `to`.
- The right end of the nav holds `ThemeToggle` and then `AccountControl`.
- Below the `md` breakpoint:
  - the strip shows the current page's label on the left, and the theme toggle and a "Menu" button on the right;
  - "Menu" (`aria-expanded`, `aria-controls`) opens a full-width panel under the bar listing every item plus the account entries;
  - the panel closes on Esc, on an outside click, and on route change.

### ThemeToggle

A small icon button that cycles System → Light → Dark. Its `aria-label` names the current setting and the next one, for example "Theme: system (switch to light)". The icons are inline SVG, a monitor, sun and moon, so no icon library is needed.

### AccountControl

- **Signed out** (`useAuth().user === null`): a "Sign in" secondary button opens `SignInDialog`.
- **While `useAuth().isLoading`:** a fixed-width placeholder, so the layout doesn't shift.
- **Signed in:** a button showing the user's name opens a `Menu` with these items:
  - Players (`/players`), for everyone;
  - Account (`/account`), for everyone;
  - Edit info (`/info/edit`), only if `permissions.canEdit`;
  - Users (`/users`), only if `permissions.canManageUsers`;
  - Sign out.

  These are legacy links for now. Managers get this menu too. The legacy site hides it from them, which leaves them no way to reach Account or log out.
- **Sign out** opens a `ConfirmDialog` ("Sign out?"). Confirming sends `POST /api/v1/auth/logout`, runs `queryClient.clear()` and then `refresh()`.

### SignInDialog

- A `Modal` titled "Sign in" containing a `<form>` with Email (`type=email`, `autocomplete=username`, required), Password (`autocomplete=current-password`, required), a "Remember me" checkbox, a "Cancel" button, and a "Sign in" submit button (primary, with a loading state).
- Submitting sends `POST /api/v1/auth/login` with JSON `{ email, password, remember }`.
- A response of `{ resetRequired: true, resetUrl }` sends the browser to `resetUrl` with `window.location.assign`. That page is still the legacy reset page.
- A response of `{ user }` closes the dialog, calls `refresh()`, and shows a toast "Signed in as <name>".
- A 401 shows an inline error `Alert`, "Incorrect email or password.", and keeps the email and clears the password.
- Any other `ApiError` shows its message in the `Alert`.
- Focus starts on Email.
- The client types gain `LoginResponse { user?: CurrentUser; resetRequired: boolean; resetUrl?: string }`. This matches the JSON tags of `auth.LoginResponse` in `server/internal/auth/types.go` (`user,omitempty`, `resetRequired`, `resetUrl,omitempty`).

### Footer

- Crest-blue background (the `footer` token) with white text.
- **Left:** "© 2020–<SiteInfo.year> AFC Aldermaston · Website provided by BSWDI", where BSWDI links to `https://bswdi.co.uk` (new tab, `rel="noopener"`). Below it, "Visitor count: <n>" appears only when signed in, matching the legacy footer.
- **Right:** round icon links to Facebook (`https://www.facebook.com/AFC-Aldermaston-114651238068/`) and X (`https://x.com/afcaldermaston`), using inline SVG, with `aria-label`s.
- If `useSite` hasn't loaded yet, the year falls back to the current year and the visitor count is omitted.

## Components (`client/components/ui/`)

Each component is one file with typed props, forwards `className` (merged with `clsx`), and spreads native props where it wraps a native element. `client/components/ui/index.ts` re-exports them all.

| Component | Behaviour |
|---|---|
| `Button` | `variant: 'primary' \| 'secondary' \| 'ghost' \| 'danger'` (default primary). `size: 'sm' \| 'md'`. `loading` shows a spinner and sets `aria-busy` and `disabled`. Defaults to `type="button"`. |
| `ButtonLink` | The same styles on a router `Link` (`to`) or a plain `<a>` (`href`). |
| `Badge` | `tone: 'red' \| 'blue' \| 'neutral'`. A small uppercase pill. |
| `Alert` | `tone: 'success' \| 'error' \| 'warning' \| 'info'`, with a coloured left border on the surface background. `error` gets `role="alert"`; the others get `role="status"`. |
| `Card` | A bordered, rounded container. `CardBody` provides padding. |
| `CardMedia` | Takes `src?: string`, `alt: string` and `aspect` (default `16/9`). It renders the image lazy-loaded with `object-cover`. It renders the gradient (`linear-gradient(135deg, var(--color-blue), var(--color-red))`, `aria-hidden`) when `src` is missing or empty, or after the image's `onError` fires. |
| `Skeleton` | A pulsing surface bar (the pulse is disabled under reduced motion). |
| `Spinner` | A spinning ring, with `role="status"` and a visually hidden "Loading". |
| `EmptyState` | A title, an optional message and an optional action, centred in muted text. |
| `Field` | Takes `label`, `help?`, `error?` and `id?` (generated with `useId` when absent), with one control as its child. It wires `htmlFor`, `aria-describedby` (help and error ids) and `aria-invalid` onto the control through a small context that `Input`, `Textarea`, `Select` and `FileInput` read. The error text replaces the help text and is shown in red. |
| `Input`, `Textarea`, `Select`, `FileInput` | Native elements with shared field styles, and a red border when invalid. |
| `Checkbox` | A native checkbox with its label to the right; takes `label`. |
| `fieldError(err, name)` | A helper returning `err.fields?.[name]` when `err` is an `ApiError`, so pages can put server validation messages straight into `Field`. |
| `Table` | `Table` (with a horizontal-scroll wrapper), `THead`, `TBody`, `Tr` (hover surface), `Th` (small uppercase muted text), `Td`. Styling only. |
| `Modal` | Takes `open`, `onClose`, `title`, `children` and an `actions?` slot. It uses native `<dialog>`: `showModal()` when `open` becomes true and `close()` when it becomes false. The dialog's `cancel` (Esc) and `close` events call `onClose`, and a click on the backdrop also closes it. `aria-labelledby` points at the title. Focus returns to the element that had it before opening. |
| `ConfirmDialog` | A `Modal` with `title`, `message`, `confirmLabel`, `tone: 'danger' \| 'primary'`, `onConfirm` (may return a promise, which shows a loading state until it settles) and `onCancel`. |
| `useDisclosure` | Returns `{ open, setOpen, toggle, close }`. The menus use it. |
| `Menu` | A trigger button plus a popover list. It uses `aria-haspopup="menu"` and `aria-expanded`, and items have `role="menuitem"`. Esc and outside clicks close it. ArrowUp and ArrowDown move focus between items, and choosing an item closes the menu. Items can be links or buttons. |
| `ToastProvider` / `useToast` | `show({ tone, message })`. Toasts are stacked bottom-right in an `aria-live="polite"` region, with at most 3 visible (the oldest is dropped). Each auto-dismisses after 5s or on click. |
| `PageHeader` | An `<h1>` in `font-display`, uppercase, with an optional `subtitle` and `actions` slot. |
| `Container` | `max-w-6xl`, centred, with `px-4` (16px gutters on phones) and more padding from `md` up. |

`main.tsx` gains `ThemeProvider` and `ToastProvider`. The order is StrictMode → QueryClientProvider → BrowserRouter → ThemeProvider → AuthProvider → ToastProvider → App.

## Pages

- **`/app/design` (`client/pages/DesignPage.tsx`):** a showcase rendering every component in every state (all button variants, sizes, loading and disabled; badges; alerts; cards with and without images and with a broken image; skeletons; fields with help and with an error; a table; buttons that open a modal, a confirm dialog and each toast tone; a menu) plus the palette swatches. It uses static sample content only, isn't linked in the nav, and is safe to be public. Sub-project 4 uses it as the reference, and reviewers use it to check the look.
- **`HomePage`:** the proof page restyled with `PageHeader`, `Card`, `Alert`, `Skeleton`, `Badge` and `Table`. Its content and behaviour are unchanged.
- **`NotFoundPage`:** restyled with `EmptyState` and a `ButtonLink` home.

## Testing

Vitest, Testing Library and jsdom, as in sub-project 2. `client/test/setup.ts` gains:
- a `HTMLDialogElement.prototype.showModal` and `close` polyfill (jsdom lacks them) that toggles the `open` attribute and dispatches `close`;
- a controllable `matchMedia` mock;
- `vitest-axe` matchers.

Tests to add:
- **Theme:**
  - the default is `system`, and resolves from `matchMedia`;
  - a stored `dark` is applied;
  - a `matchMedia` change updates `system`;
  - the toggle cycles through the settings and persists them;
  - a throwing `localStorage` still renders;
  - `theme-init.js`, run in jsdom, sets `data-theme` from storage and falls back safely when storage throws.
- **NavBar:**
  - Home is a router link with `aria-current` on `/`;
  - News is a plain `<a href="/news">`;
  - the mobile Menu opens and closes on click, Esc and route change.
- **AccountControl:**
  - signed out, the button shows "Sign in";
  - signed in, the user's name shows;
  - the menu items for a Manager fixture (Players, Account, Sign out), an editor (plus Edit info) and a user admin (plus Users);
  - sign-out confirms, POSTs to `/auth/logout`, and the name disappears.
- **SignInDialog:**
  - success refreshes and closes;
  - `resetRequired` calls `window.location.assign` with `resetUrl`;
  - 401 shows "Incorrect email or password." and clears the password;
  - the submit button shows loading while the request is pending.
- **Footer:** the visitor count appears only when signed in; the social links have labels.
- **Components:**
  - `Button` in the loading state is disabled and sets `aria-busy`;
  - `Field` wires `aria-describedby` and `aria-invalid`, and shows the error in place of the help text;
  - `CardMedia` shows the image when `src` is set, the gradient when it's missing, and the gradient after `onError`;
  - `Modal` opens, `onClose` fires on Esc (the `cancel` event), and focus returns to the trigger;
  - `ConfirmDialog` shows a loading state while `onConfirm` is pending;
  - `Menu` opens and closes, moves between items with arrow keys, and closes on Esc;
  - `Toast` auto-dismisses at 5s (fake timers) and keeps at most 3;
  - `fieldError` maps `ApiError.fields`.
- **Accessibility smoke:** `axe` finds no violations on the `DesignPage` and on the shell signed out and signed in (both themes). jsdom can't check colour contrast; the palette above was chosen for AA contrast by hand.
- **Existing tests:** the `HomePage` and route tests are updated for the new markup, keeping their assertions about behaviour.
- `yarn lint`, `yarn typecheck`, `yarn test` and `yarn build` stay green, and the Go tests are unaffected (there are no server changes).

## Out of scope

- Real page content, including Home: sub-project 4.
- Reset, account and admin pages, and nav targets moving into the SPA: sub-project 4.
- Moving the SPA to `/`, and legacy redirects: sub-project 5.
- The updated FA logo, which is a file swap whenever it arrives.
- Visual regression screenshots, Storybook, icon libraries, and any sorting or pagination inside `Table`.
