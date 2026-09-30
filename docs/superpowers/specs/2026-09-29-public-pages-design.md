# Public Pages Design (Sub-project 4a)

Sub-project 4 of 5 in moving AFC-GO to the MV-Controller layout. Sub-project 4 is split into three PRs, and this spec covers the first.

| # | Sub-project | Status |
|---|---|---|
| 1 | Server restructure + JSON API | Done (#11) |
| 2 | Client scaffold + build pipeline | Done (#12) |
| 3 | Design system | Done (#13) |
| **4a** | **Public pages (read-only)** (this spec) | Designing |
| 4b | Account pages: account, change password, reset link | Later |
| 4c | Editing on every page, plus the Users, Players and Edit info pages | Later |
| 5 | Cutover (SPA takes over `/`; delete templates) | Later |

## Context

`main` serves the React client at `/app` with:
- the sub-project 3 shell (masthead, nav, sign-in and account menu, footer);
- the components in `client/components/ui/`;
- the `/app/design` showcase;
- a proof home page.

Nav items other than Home still link to the legacy pages. The JSON API (`/api/v1`) already has a read endpoint for every public page. The route-walk tests from sub-project 1 check that every legacy action has an API route, so this sub-project changes the client only.

In the legacy site, the edit controls sit inside the public pages (add/edit/delete modals on News, Teams and so on). In 4a the new pages are read-only; editing arrives in 4c.

## Decisions

- **Same content as the legacy pages, new look, plus four client-only upgrades:**
  1. What's On tabs, the Programmes season, and search terms are kept in the address, so links can be shared and Back works.
  2. "Show more" on long card lists (12 at a time).
  3. A better gallery viewer: keyboard, swipe, counter, captions.
  4. Filter-as-you-type search on Documents and Programmes.
- **Home layout, option B from the mockups.** The latest article is a large image hero with its opening lines; the next event sits in a narrower column beside it; then sponsor and affiliation logo rows.
- **List, article and team page patterns** follow the mockups approved in brainstorming (`.superpowers/brainstorm/`, untracked).
- **One route component per page, and one data module per section**, with shared page pieces for the repeated patterns. There is no generic config-driven engine and no new API endpoints.
- **Stored HTML is cleaned on the client with DOMPurify before display.** The server cleans (bluemonday) only when saving, so rows written before that existed may be unsafe.
- **Every page except Home is loaded only when first visited** (`React.lazy`).
- **Editors keep editing on the legacy site until 4c.** On pages that have legacy edit controls, signed-in editors see a "Manage this on the classic site" link.

## Routes and navigation

The routes mirror the legacy paths, under the `/app` basename:

| Route | Page | API |
|---|---|---|
| `/` | Home | `GET /home` |
| `/teams` | Teams | `GET /teams` |
| `/team/:id` | Team | `GET /teams/:id` |
| `/news` | News list | `GET /news` |
| `/news/:id` | News article | `GET /news/:id` |
| `/whatson` | What's On (`?period=future\|past\|all`) | `GET /whatson?period=` |
| `/whatson/:id` | Event | `GET /whatson/:id` |
| `/gallery` | Gallery | `GET /gallery` |
| `/documents` | Documents (`?q=`) | `GET /documents` |
| `/programmes` | Programmes (`?season=`, `?q=`) | `GET /programmes?season=`, `GET /seasons` |
| `/sponsors` | Sponsors | `GET /sponsors` |
| `/info` | Info | `GET /info` |
| `/contact` | Contact | `GET /contact` |
| `/design`, `*` | unchanged (showcase, not found) | — |

`client/components/layout/navItems.ts` gains a `to` for every item, so the whole nav stays in the SPA. `legacyHref` is kept: it's the target of the editor link on each page.

`App.tsx` wraps the lazy routes in one `<Suspense>`, whose fallback is the page loading placeholder. `HomePage` is imported normally.

## Data layer

There is one module per section in `client/api/`. Each has hand-written types that mirror the Go JSON, and TanStack Query hooks built on `apiFetch`. The keys come from the `queryKeys` object in `client/api/queries.ts`, which is extended.

| Module | Types (Go source) | Hooks |
|---|---|---|
| `home.ts` | `HomeData` (`site.Home`: `latestNews?`, `nextEvent?`, `sponsors`, `affiliations`), `Affiliation` (`affiliation.Public`) | `useHome()` |
| `teams.ts` | `TeamDetail` (`site.TeamDetail`: `team`, `managers`, `sponsors`, `players`), `Manager`, `SquadMember` (`player.Member`); `TeamSummary` already exists | `useTeams()`, `useTeam(id)` |
| `news.ts` | `NewsArticle` (`news.Article`: `id`, `title`, `content`, `date`, `imageUrl?`) | `useNewsList()`, `useNewsArticle(id)` |
| `whatson.ts` | `WhatsOnEvent` (`whatson.Event`: `…`, `dateOfEvent`), `WhatsOnPeriod = 'future' \| 'past' \| 'all'` | `useWhatsOnList(period)`, `useWhatsOnEvent(id)` |
| `gallery.ts` | `GalleryImage` (`image.Public`: `id`, `caption?`, `imageUrl`) | `useGallery()` |
| `documents.ts` | `ClubDocument` (`document.Public`: `id`, `name`, `fileUrl`) | `useDocuments()` |
| `programmes.ts` | `Programme` (`programme.Public`: `id`, `name`, `date`, `fileUrl`, `season?`), `Season` (`programme.PublicSeason`) | `useProgrammes(seasonId)`, `useSeasons()` |
| `sponsors.ts` | `Sponsor` (`sponsor.Public`: `id`, `name`, `website?`, `purpose?`, `team?`, `imageUrl?`) | `useSponsors()` |
| `pages.ts` | `InfoContent` (`setting.InfoContent`: `content`), `ContactData` (`site.Contact`: `displayEmail?`, `people`), `ContactPerson` (`id`, `name`, `email`, `role`, `imageUrl?`) | `useInfo()`, `useContact()` |

Rules for the data layer:
- **Dates:** JSON dates are ISO strings in the types, and are formatted only at display time.
- **`useTeams()`:** the cache key includes whether someone is signed in. The server returns inactive teams to signed-in users, so the anonymous and signed-in lists must not share a cache entry.
- **Detail hooks** take a numeric ID. Parsing happens in the page (see Missing items).
- **Before writing the types:** read each Go type (`server/internal/*/types.go`, `server/internal/site/types.go`, `server/internal/setting/types.go`) and mirror its JSON tags exactly, including `omitempty`, which becomes `?`.

## Shared page pieces (`client/components/page/`)

| Piece | Behaviour |
|---|---|
| `RichText({ html })` | Cleans `html` with DOMPurify, then renders it with `dangerouslySetInnerHTML` inside a `prose`-style wrapper (heading, list, link and paragraph styles in `app.css`). Links from the content that point off-site open in a new tab with `rel="noopener noreferrer"`, using a DOMPurify `afterSanitizeAttributes` hook. |
| `plainText(html, maxChars)` | Turns HTML into plain text using DOMPurify with no tags allowed, then cuts it at a word boundary and adds "…". Used for the Home hero's opening lines. |
| `CardGrid({ items, render, pageSize = 12, emptyTitle })` | A responsive grid (1, 2 or 3 columns). It shows the first `pageSize` items and then a "Show more (N more)" button that reveals the next `pageSize`; the button disappears when everything is shown. When `items` is empty it shows `EmptyState` with `emptyTitle`. The count shown resets when `items` changes identity (for example, a new tab). |
| `TabsNav({ param, tabs, defaultValue })` | Tabs styled like the What's On mockup, as a `nav` of links. Each link sets `?param=value` and keeps the other parameters. The active tab has `aria-current="page"`. An unknown value falls back to `defaultValue`. |
| `SearchInput({ label })` | A labelled search input bound to `?q=`. Each change replaces the address entry rather than adding one, so Back doesn't step through every keystroke. It's paired with `matchesQuery(text, q)`, which matches case-insensitively on words. |
| `QueryState({ query, children, emptyTitle?, isEmpty? })` | Renders `Skeleton` rows while the data loads. On error it shows an `Alert` with the `ApiError` message and a Retry button (`query.refetch()`). If `isEmpty(data)` is true it shows `EmptyState`. Otherwise it renders `children(data)`. |
| `usePageTitle(title)` | Sets `document.title` to `"<title> · AFC Aldermaston"`, or `"AFC Aldermaston"` for Home. |
| `EditorLink({ legacyHref })` | Shows only when `useAuth().user?.permissions.canEdit` is true: a small ghost link, "Manage this on the classic site ↗", pointing at the legacy page. On Gallery the condition is `canManageGallery`. |
| `formatDate(iso, style)` | Wraps `Intl.DateTimeFormat('en-GB', { timeZone: 'Europe/London' })`. `style` is `'date'` ("28 Sep 2026") or `'dateTime'` ("Fri 16 Oct 2026, 7pm"): the weekday comes first, minutes are dropped when they're :00, and the time uses 12-hour am/pm. |

**Missing items:** detail routes parse `:id` with `/^\d+$/`. A non-numeric ID renders `NotFoundPage` without calling the API. An `ApiError` with status 404 from a detail hook also renders `NotFoundPage`. Other errors go through `QueryState`.

## Pages (`client/pages/<section>/`)

- **Home.** Replaces the proof page, which is deleted.
  - The **hero** has `latestNews`'s `CardMedia` image, the "Latest news" kicker and title over a bottom gradient, then `plainText(content, 180)` and "Read more →" linking to `/news/:id`.
  - The **event column** has the "Next event" kicker, title, `formatDate(dateOfEvent, 'dateTime')`, `plainText(content, 120)` and "All events →" (`/whatson`).
  - A **missing panel** is skipped, and the other one spans the full width. With neither, there's no hero row at all.
  - **Logo rows:** "Our sponsors" and "Affiliations" show logo tiles (image, or the name when there's none). A tile links to its website in a new tab when it has one. Each row is hidden when empty.
  - It shows only the club name as the title (the masthead carries the rest) and uses `usePageTitle` with no page name.
- **Teams.** A `CardGrid` of team cards: `CardMedia` photo, name, league and division, and a `Badge` (Youth in red, Adult in blue). Inactive teams (only returned to signed-in users) get a neutral "Inactive" badge. Each card links to `/team/:id`. An `EditorLink` points to `/teams`.
- **Team.**
  - **Breadcrumb:** "Teams / <name>", with the title as `PageHeader`.
  - **Two columns:** the team photo on the left. On the right, a definition list of League, Division, Coach and Physio (each row only when set), and Managers ("Name (email link)" per manager, from `managers`). Below that, "League table ↗" and "Fixtures ↗" `ButtonLink`s (new tab), each only when its URL is set.
  - **"Squad":** a grid of player tiles with a round photo (or the crest when there's no `imageUrl`), name, position, and a Captain badge. The section is rendered only when `players` is non-empty. The API returns no players for youth teams, so youth teams show no squad section at all, and nothing says a squad is hidden.
  - **"Team sponsors":** logo tiles, hidden when there are none.
  - `description` is shown as plain text under the title when set.
- **News.** A `CardGrid` of cards (image, title, `formatDate(date, 'date')`) linking to `/news/:id`, newest first as the API returns them. Empty: "No news yet". Editor link: `/news`.
- **News article.**
  - The `CardMedia` image across the top, and a "News / <date>" breadcrumb.
  - The `PageHeader` title, then `RichText(content)` and "← All news".
  - Editor link: `/news/:id`.
- **What's On.**
  - `TabsNav` on `period` with Upcoming (`future`, the default), Past (`past`) and All (`all`).
  - A `CardGrid` of cards (image, title, `formatDate(dateOfEvent, 'dateTime')`) linking to `/whatson/:id`, in the API's order.
  - Empty titles: "No upcoming events", "No past events", "No events yet".
  - Editor link: `/whatson`.
- **Event.** Like the article, with the "What's On / <event date>" breadcrumb, the event date and time under the title, and "← All events".
- **Gallery.**
  - A square-thumbnail grid of buttons (`aria-label` = the caption, or "Photo N of M"). Clicking one opens `Lightbox` at that index.
  - `Lightbox` (`client/pages/gallery/Lightbox.tsx`) is built on `Modal`'s native `<dialog>` and shows:
    - the large image with its caption;
    - an "N / M" counter;
    - previous and next buttons, which wrap around;
    - ← and → keys, and Esc to close;
    - touch swipe: a horizontal movement of 50px or more between `touchstart` and `touchend` moves one photo.
  - Focus returns to the thumbnail that opened it.
  - Empty: "No photos yet". Editor link (on `canManageGallery`): `/gallery`.
- **Documents.** `SearchInput` ("Search documents"), then a list with one row per document (name, plus a "Download" `ButtonLink` to `fileUrl`), filtered with `matchesQuery(name, q)`. Empty (no documents): "No documents yet". No matches: "No documents match '<q>'". Editor link: `/documents`.
- **Programmes.**
  - A season `Select` bound to `?season=`: "All seasons" plus one option per season from `useSeasons()`. It updates the address, replacing the current entry.
  - A `SearchInput` bound to `?q=`, matching the programme name.
  - The list is grouped by season name ("No season" last), and each row has the name, `formatDate(date, 'date')` and "View" (`fileUrl`, new tab).
  - Empty and no-match states work as on Documents.
  - Editor link: `/programmes`.
- **Sponsors.** A `CardGrid` (`pageSize` large enough to show everything, with no "Show more") of cards: logo, name, purpose, team (when set), and a "Website ↗" link when set. Editor link: `/sponsors`.
- **Info.**
  - `RichText` of the stored info HTML.
  - When it's empty, the fallback copy is shown instead: the static wording in `server/internal/legacy/templates/info.tmpl` (the non-commented `{{else}}` branch), moved into `client/pages/info/fallback.tsx` as JSX.
  - Editor link: `/info/edit`.
- **Contact.**
  - One card per person: round photo (or the crest), name, role, and an email link. The address shown is `displayEmail` when set, otherwise the person's own `email`, as the legacy site does.
  - Then a notice `Alert` (info tone): "If you're using a satnav, use the postcode RG26 4QP — the postcode listed takes you some distance away."
  - Then the Google Maps embed from the legacy template, as an `<iframe>` with `title="Map to Aldermaston Recreational Society"`, `loading="lazy"`, `referrerpolicy="no-referrer-when-downgrade"`, full width, height 400px, rounded, with a line border.
  - There's no editor link: the contacts come from user roles, which are edited on the Users page.

Every page calls `usePageTitle` and wraps its data in `QueryState`.

## Styles

`client/styles/app.css` gains a `.prose` component class in `@layer components`, using the tokens only:
- paragraph spacing and a line height of 1.6;
- `h2` and `h3` in `font-display`, uppercase;
- lists with bullets and numbers;
- `a` in the red token, underlined;
- `img` with `max-width: 100%` and rounded corners;
- `blockquote` with a left border in the line colour.

This is instead of adding `@tailwindcss/typography`, to avoid a dependency.

## Testing

Vitest, Testing Library and jsdom, using the mocked `fetch` and `renderWithProviders` from earlier sub-projects:

- **Every page:**
  - loaded content;
  - the empty state (where it has one);
  - an error with Retry that refetches;
  - `document.title`.
- **Detail pages** (Team, News article, Event):
  - an API 404 shows "Page not found";
  - `/news/abc` shows it without calling `fetch`.
- **`RichText`:**
  - strips `<script>`, `onerror`, `javascript:` hrefs and `<iframe>`;
  - keeps `b`, `i`, `u`, `strong`, `em`, `p`, `h1`–`h3`, `ul`, `ol`, `li`, `a` and `blockquote`;
  - off-site links get `target="_blank"` and `rel="noopener noreferrer"`.
- **`plainText`:** strips tags and truncates at a word boundary with "…".
- **`CardGrid`:**
  - shows 12;
  - "Show more (N more)" reveals 12 more and disappears at the end;
  - resets when the items change.
- **`TabsNav`:**
  - the active tab comes from the address;
  - clicking a tab updates `?period=`;
  - the router's Back restores the previous tab;
  - an unknown value falls back to the default.
- **`SearchInput` and `matchesQuery`:**
  - typing filters the list and sets `?q=`;
  - rendering with `?q=` already set starts filtered;
  - the no-match state appears.
- **What's On:** each tab requests `/whatson?period=<value>`.
- **Programmes:** choosing a season requests `/programmes?season=<id>` and sets `?season=`; items are grouped by season.
- **Team:**
  - a youth team (no players) renders no "Squad" heading;
  - there are no League table or Fixtures buttons when those URLs are absent;
  - manager email links are present.
- **Home:**
  - both panels;
  - only one panel;
  - neither panel;
  - the logo rows are hidden when empty.
- **Gallery `Lightbox`:**
  - opens at the clicked photo;
  - next and previous wrap;
  - ← and → move and Esc closes;
  - the "N / M" counter;
  - a swipe (simulated touch events) moves;
  - focus returns to the thumbnail.
- **Contact:** `displayEmail` overrides the person's own email; the map iframe has a title.
- **`EditorLink`:**
  - hidden when signed out and for Managers;
  - shown for editors;
  - the gallery version follows `canManageGallery`.
- **Nav:** every item renders as an in-app link.
- **Accessibility:** `client/a11y.test.tsx` is extended to every new route, signed in and signed out, in both themes, with no axe violations.
- **The usual checks:** `yarn lint`, `yarn typecheck`, `yarn test` and `yarn build` stay green, and the Go tests are unchanged.

## Out of scope

- 4b: the Account page, change password, and the SPA reset-link page. The sign-in dialog keeps sending reset-flagged accounts to the legacy reset page.
- 4c: all add, edit and delete controls; the Users, Players and Edit info pages; removing the editor links.
- Sub-project 5: redirecting legacy URLs, and moving the SPA to `/`.
- API changes, including server-side paging and search.
