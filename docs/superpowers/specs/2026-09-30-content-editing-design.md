# Content Editing Design (Sub-project 4c-1)

Part of sub-project 4 in moving AFC-GO to the MV-Controller layout. 4c is split into two PRs: **4c-1** (this spec) covers content editing, and **4c-2** covers the Players and Users admin pages.

| # | Sub-project | Status |
|---|---|---|
| 1–3 | Server/API, client scaffold, design system | Done (#11–#13) |
| 4a, 4b | Public pages; account and reset pages | Done (#14, #15) |
| **4c-1** | **Content editing** (this spec) | Designing |
| 4c-2 | Players and Users admin (including display email and admin password resets) | Later |
| 5 | Cutover (SPA takes over `/`; delete templates) | Later |

## Context

The React client at `/app` shows every public page read-only. Editors currently follow "Manage this on the classic site ↗" links to add, edit or delete content on the legacy pages. The API already has every write endpoint needed, so there are no server changes:

| Endpoint | Body | Guard |
|---|---|---|
| `POST /news`, `PATCH /news/:id`, `DELETE /news/:id` | multipart: `title`*, `content`, `image`, and on PATCH `removeImage` | Editor |
| `POST /whatson`, `PATCH /whatson/:id`, `DELETE /whatson/:id` | multipart: `title`*, `content`, `dateOfEvent`* (`YYYY-MM-DD`), `image`, and on PATCH `removeImage` | Editor |
| `PUT /info` | JSON `{ content }` | Editor |
| `POST /teams`, `PATCH /teams/:id`, `DELETE /teams/:id` | multipart: `name`*, `ages`* (int), `description`, `league`, `division`, `leagueTable`, `fixtures`, `coach`, `physio`, `isActive`, `isYouth` (`true`/`false`), `image`, and on PATCH `removeImage` | Editor |
| `POST /documents`, `DELETE /documents/:id` | multipart: `name`*, `file`* | Editor |
| `POST /programmes`, `DELETE /programmes/:id` | multipart: `name`*, `date`* (`YYYY-MM-DD`), `seasonId` (0 or omitted = none), `file`* | Editor |
| `POST /seasons`, `PATCH /seasons/:id`, `DELETE /seasons/:id` | JSON `{ name }` | Editor |
| `POST /sponsors`, `DELETE /sponsors/:id` | multipart: `name`*, `website`, `purpose`, `team` (`""`, `A` all, `O` adult, `Y` youth, or a team id), `image`* | Editor |
| `POST /affiliations`, `DELETE /affiliations/:id` | multipart: `name`*, `website`, `image`* | Editor |
| `POST /gallery`, `DELETE /gallery/:id` | multipart: `caption`, `image`* | NotManager (`canManageGallery`) |

(`*` = required.)

**Server behaviour the client relies on:**
- PATCH changes only the fields that are present. An empty string clears a text field.
- `FormBool` accepts `true`/`false`.
- Deleting a team unlinks its players, sponsors and managers.
- Deleting a season unlinks its programmes (they keep existing, with no season).
- Validation errors are `422` with `fields`. Upload type errors use field `file`.
- The body limit is 15 MB (`413` beyond that).
- Stored HTML is cleaned (bluemonday) on save.

## Decisions

- **Split:** two PRs, content first (this spec), then admin (4c-2).
- **Editor:** Tiptap (StarterKit, Underline, Link), loaded only on edit screens.
- **Where editing happens:**
  - Add, Edit and Delete controls appear on the public pages for users allowed to use them.
  - Long forms (article, event, team, info) get their own routes.
  - Short forms (document, programme, season, sponsor, affiliation, photo) open in a dialog.
  - Delete always asks for confirmation.
- **Structure:** a small shared editing toolkit, with one plain form per section. There's no generic CRUD engine and no form library.
- **Classic-site links:** the "Manage this on the classic site" links are removed from every content page, and **Edit info** in the account menu becomes an in-app link.
- **Two 4a display bugs are fixed here:** sponsor cards show the team code instead of a name (see Sponsors), and
- **What's On dates:** event dates show the date only. `dateOfEvent` is a date without a time, so the 4a `'dateTime'` format showed a meaningless time ("1am" in summer).

## Editing toolkit (`client/components/edit/`)

- **`useCanEdit(): { canEdit: boolean; canManageGallery: boolean }`:** reads `useAuth().user?.permissions`, returning false when signed out.
- **`RequireEditor({ permission = 'canEdit', children })`:**
  - while the sign-in check loads: the page loading placeholder;
  - without the permission: an `EmptyState` titled "You don't have permission to edit this", with a `ButtonLink` back ("Back") to `-1`, or to `/` when there's no history;
  - otherwise: `children`.
- **`useSaveForm<T>({ submit: () => Promise<T>, invalidate: QueryKey[], onSaved(result: T) })`:** returns `{ run, busy, fieldErrors, formError, setFieldErrors }`.
  - `run()` sets busy and clears previous errors, then calls `submit`.
  - On success it invalidates every key in `invalidate` and calls `onSaved`.
  - On a `422` with `fields` it sets `fieldErrors`.
  - On a `401` it calls `useAuth().refresh()`; the page then shows `RequireEditor`'s message or the header's Sign in.
  - On a `413` it sets `formError` to "That file is too large (15 MB maximum)."
  - Any other error sets `formError` to the error message.
- **`ImageField({ label, currentUrl?, required?, allowRemove?, value, onChange, error })`:**
  - `value` is `{ file: File | null; remove: boolean }`.
  - It shows the current image (`ImageWithFallback`, with the gradient as fallback) or a preview of the chosen file.
  - Accepted types match the server's image list. A wrong type shows "Choose an image file (JPEG, PNG, GIF, WebP, AVIF, APNG or SVG)." and no preview.
  - Previews use the 4b `previewUrl` rules: only `blob:` URLs, revoked when replaced or on unmount. `previewUrl` and the type list move to `client/lib/images.ts`, and 4b's `PhotoCard` uses them from there.
  - **Remove image** is a checkbox, shown when `allowRemove` is set and there's a current image. Choosing a new file clears it.
- **`FileField({ label, required?, value: File | null, onChange, error, accept? })`:** a file input that shows the chosen file's name and size.
- **`RichTextEditor({ label, value: string, onChange(html: string), error? })`:** lazy-loaded (`React.lazy` around `RichTextEditorImpl`), with a skeleton while loading.
  - Tiptap `useEditor` with StarterKit (headings limited to levels 2–3), Underline, and Link (`openOnClick: false`, `autolink: true`, `protocols: ['http','https','mailto']`).
  - **Toolbar:** Bold, Italic, Underline, Strikethrough, Heading 2, Heading 3, Bullet list, Numbered list, Quote, Link, Remove link, Horizontal rule, Undo, Redo.
  - Toolbar buttons have `aria-label`, `aria-pressed` for toggles, and `role="toolbar"` with the editor's label.
  - **Link** opens a small dialog asking for a URL. It accepts `http(s)://` and `mailto:` URLs; a bare domain gets `https://` added.
  - The editable area has `aria-label`, and an error (if any) through `aria-describedby`.
  - Output is `editor.getHTML()`. An empty document (`<p></p>`) is emitted as `''`.
- **`DeleteButton({ label, confirmTitle, confirmMessage, onDelete: () => Promise<unknown>, invalidate, successMessage, after? })`:**
  - a danger button that opens a `ConfirmDialog` (danger tone);
  - on confirm it calls `onDelete`, invalidates the keys, shows `successMessage` as a toast, then calls `after`, which is usually a navigate;
  - on an error it shows the toast "Couldn't delete: <message>", and a 401 re-reads the session.
- **`editForm.ts` helpers:**
  - `formData(fields: Record<string, string | number | boolean | File | null | undefined>)`: strings (including empty) and numbers are appended; booleans become `'true'`/`'false'`; `File` is appended; `null`/`undefined` are skipped.
  - `toDateInput(iso)` and `fromDateInput(value)` for `YYYY-MM-DD`, in `Europe/London`.

## Pages

Controls use `useCanEdit()`. "Add" buttons sit in the page header's `actions`. Toast messages are "<Thing> saved", "<Thing> added" and "<Thing> deleted".

### News
- **List:** "Add article" (links to `/news/new`).
- **Article:** "Edit" (links to `/news/:id/edit`) and a `DeleteButton`. Confirm: "Delete this article?" / "This can't be undone." Afterwards it navigates to `/news` with the toast "Article deleted".
- **`NewsFormPage`** (`/news/new` and `/news/:id/edit`, inside `RequireEditor`):
  - Fields: Title* (`Input`), Image (`ImageField`, `allowRemove` on edit), Content (`RichTextEditor`).
  - Edit loads the article through `useNewsArticle`; a missing or bad ID shows `NotFoundPage`.
  - **Save** is `POST /news` on create, or `PATCH /news/:id` on edit with `title`, `content`, `image` (if chosen) and `removeImage` (if ticked).
  - Invalidates `news`, `newsArticle(id)` and `home`. Afterwards it navigates to `/news/:id` with the toast "Article saved".
  - An empty title shows "Enter a title" and nothing is sent.
  - **Cancel** goes back.

### What's On
- **List and event:** as for News: "Add event", "Edit" and `DeleteButton` ("Delete this event?").
- **`EventFormPage`** (`/whatson/new` and `/whatson/:id/edit`):
  - Fields: Title*, **Date of event*** (`<Input type="date">`), Image, Content.
  - Sent as `dateOfEvent=YYYY-MM-DD`. An empty date shows "Choose the date of the event".
  - Invalidates `whatson(*)` (all periods), `whatsonEvent(id)` and `home`.
- **Date display fix:** everywhere 4a showed `formatDate(dateOfEvent, 'dateTime')` (What's On cards, the event page subtitle, and Home's next event) now uses `'date'`, with the weekday added. That needs a new `formatDate` style `'dayDate'`, giving "Fri 16 Oct 2026".

### Info
- **Info page:** "Edit" (links to `/info/edit`).
- **Account menu:** **Edit info** becomes `{ to: '/info/edit' }`.
- **`InfoEditPage`** (`/info/edit`):
  - one `RichTextEditor` labelled "Club information";
  - it starts from the stored content. When that's empty, it starts from the `InfoFallback` copy rendered to HTML (`renderToStaticMarkup(<InfoFallback />)`), so saving without changes keeps the fallback wording.
  - **Save** sends `PUT /info` `{ content }`, invalidates `info`, and navigates to `/info` with the toast "Club information saved".

### Teams
- **List:** "Add team" (links to `/teams/new`).
- **Team page:** "Edit" (links to `/team/:id/edit`) and a `DeleteButton`. Confirm: "Delete this team?" / "Its players, sponsors and managers will be unlinked from it. This can't be undone." Afterwards it navigates to `/teams`.
- **`TeamFormPage`** (`/teams/new` and `/team/:id/edit`):
  - Fields: Name*, **Age group*** (`Select`: "Under 6" … "Under 18" = 6…18, and "Over 18" = 19, matching the legacy form's options), Description (`Textarea`), League, Division, League table URL, Fixtures URL, Coach, Physio, Image (`allowRemove` on edit), **Active team** and **Youth team** checkboxes.
  - New teams default to Active checked, Youth unchecked.
  - Every field is always sent; the booleans are sent as `true`/`false`.
  - Invalidates `teams(true)`, `teams(false)`, `team(id)` and `site`. Afterwards it navigates to `/team/:id`.
  - Empty required fields show "Enter a name" / "Choose the age group". The age group is sent as an integer, `ages`.

### Documents
- "Add document" opens the **Add document** dialog: Name* and File* (`FileField`). It sends `POST /documents`, invalidates `documents`, and shows the toast "Document added".
- Each row gets a small `DeleteButton` ("Delete <name>?").

### Programmes and seasons
- "Add programme" opens a dialog: Name*, Date* (`type=date`), Season (`Select` with "No season" plus the seasons), File*. It sends `POST /programmes` (`seasonId` only when chosen) and invalidates `programmes(*)`.
- Each row gets a `DeleteButton`.
- "Manage seasons" opens the **Seasons** dialog:
  - a list of seasons, each with **Rename** (switches the row into an input with Save/Cancel; `PATCH /seasons/:id` `{ name }`) and a `DeleteButton` ("Delete season <name>?" / "Its programmes will stay, with no season.");
  - an **Add season** input and button (`POST /seasons` `{ name }`);
  - every change invalidates `seasons` and `programmes(*)`.

### Sponsors
- "Add sponsor" opens a dialog: Name*, Logo* (`ImageField`, `required`), Website, Purpose, and **Sponsors** (`Select`):
  - "None" `""`, "All teams" `A`, "Adult teams" `O`, "Youth teams" `Y`;
  - then one option per team from `useTeams()` (value = id).
- It sends `POST /sponsors` and invalidates `sponsors`, `home` and `team(*)`.
- Each card gets a `DeleteButton`.
- **Display fix:** the API returns the raw `team` code (`A`, `O`, `Y` or a team id), so the 4a card currently shows "Sponsor of A". A new `sponsorTeamLabel(team, teams)` helper maps it:
  - `A` → "All teams", `O` → "Adult teams", `Y` → "Youth teams";
  - an id → that team's name (from `useSite().data.teams`);
  - unknown or empty → nothing shown.

  The Sponsors page uses it ("Sponsor of Youth teams", "Sponsor of First Team").

### Affiliations (Home)
- Editors see "Add affiliation" under the Affiliations row, even when the row is empty; the empty row renders the title and "No affiliations yet" for editors only.
- The dialog has Name*, Logo* and Website, and sends `POST /affiliations`.
- Each logo tile gets a small delete control: an icon button labelled "Delete <name>" with a confirmation.
- Invalidates `home`.

### Gallery
- "Add photo" (for `canManageGallery`) opens a dialog: Photo* (`ImageField`, `required`) and Caption. It sends `POST /gallery` and invalidates `gallery`.
- Each thumbnail gets a `DeleteButton` (icon, labelled "Delete photo <caption or number>") beside it, not inside the thumbnail button.

### Routes

New lazy routes, all wrapped in `RequireEditor` (Gallery has no form route):
- `news/new`, `news/:id/edit`
- `whatson/new`, `whatson/:id/edit`
- `info/edit`
- `teams/new`, `team/:id/edit`

## Data layer

Each section module gains write calls. They use `apiFetch` with `form: formData({...})` or `json`, and `method` `POST`/`PATCH`/`PUT`/`DELETE`:

- `news.ts`: `createNews`, `updateNews`, `deleteNews`
- `whatson.ts`: `createEvent`, `updateEvent`, `deleteEvent`
- `pages.ts`: `setInfo`
- `teams.ts`: `createTeam`, `updateTeam`, `deleteTeam`
- `documents.ts`: `createDocument`, `deleteDocument`
- `programmes.ts`: `createProgramme`, `deleteProgramme`, `createSeason`, `renameSeason`, `deleteSeason`
- `sponsors.ts`: `createSponsor`, `deleteSponsor`
- `home.ts`: `createAffiliation`, `deleteAffiliation`
- `gallery.ts`: `createPhoto`, `deletePhoto`

Creates return the created item where the API does (used for navigation); deletes return `void`. Invalidation targets use the existing `queryKeys`. A "prefix" invalidation such as `['whatson']` covers every What's On period.

## Dependencies

`@tiptap/react`, `@tiptap/pm`, `@tiptap/starter-kit`, `@tiptap/extension-underline` and `@tiptap/extension-link` (current majors). A build check asserts that no chunk other than the editor's contains Tiptap/ProseMirror code. Concretely: the entry chunk's size doesn't include `prosemirror`, grep-checked in the plan's verification step.

## Testing

Vitest, Testing Library and jsdom, using `mockFetch`, fixtures and `renderWithProviders`.

- **Toolkit:**
  - `useCanEdit` (signed out, Manager, editor, photographer-style);
  - `RequireEditor` (loading, no permission, allowed);
  - `useSaveForm` (success invalidates and calls `onSaved`; `422` field errors; `401` re-reads `me`; `413` message; other errors);
  - `formData` (skips null/undefined, sends empty strings, booleans as `'true'`/`'false'`);
  - `ImageField` (preview, wrong type, `blob:` only, revoke, remove checkbox, new file clears remove);
  - `FileField`;
  - `DeleteButton` (confirm, cancel, error toast, `after` called);
  - `RichTextEditor` (loads HTML; typing emits HTML; Bold/Bullet toolbar buttons set `aria-pressed` and change output; the Link dialog adds `https://` to a bare domain and rejects `javascript:`; an empty document emits `''`).
- **Per page:**
  - editors see the controls, and a Manager and a signed-out visitor don't (the gallery follows `canManageGallery`);
  - each form sends exactly the expected multipart fields or JSON;
  - required-field messages show and nothing is sent;
  - `422` field errors show on their inputs;
  - success navigates or closes the dialog and shows the toast;
  - edit forms load existing values;
  - unticking a checkbox sends `false`;
  - removing an image sends `removeImage=true`;
  - delete confirms, calls `DELETE` and redirects;
  - edit routes for a non-editor show the permission message.
- **Specific cases:**
  - the sponsor team select maps to `''`/`A`/`O`/`Y`/id, and `sponsorTeamLabel` maps codes and ids to labels on the Sponsors page;
  - the team form's age group select sends 6–19;
  - seasons: add, rename and delete (with the unlink warning), each refreshing the programmes;
  - a programme without a season doesn't send `seasonId`;
  - Info edit starts from the fallback when the stored content is empty;
  - What's On, the event page and Home show the date with no time.
- **Accessibility:** axe checks cover the new routes (as an editor) and an open add dialog, in both themes.
- **The usual checks:** lint, typecheck, test and build green; the Go tests are unchanged.

## Out of scope

- 4c-2: the Players and Users admin pages, the site-wide display email (`PUT /settings/display-email`) and admin password resets.
- Editing documents, programmes, sponsors, affiliations or photos after creation (the API has no update for these; delete and re-add, as on the legacy site).
- Image cropping or resizing, drafts, and scheduled publishing.
- Removing legacy templates (sub-project 5).
