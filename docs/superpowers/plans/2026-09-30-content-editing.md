# Content Editing Implementation Plan (Sub-project 4c-1)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let editors add, edit and delete every kind of site content from the React client at `/app`. That covers:
- News, What's On and Info, written in a Tiptap rich-text editor;
- Teams;
- Documents, Programmes and seasons, Sponsors, Affiliations and Gallery photos.

It also fixes two 4a display bugs: event times, and sponsor team codes.

**Architecture:**
- A small editing toolkit in `client/components/edit/`: `useCanEdit`, `RequireEditor`, `useSaveForm`, `ImageField`, `FileField`, `DeleteButton` and a lazily loaded `RichTextEditor`.
- Helpers in `client/lib/` for multipart bodies, dates, images and sponsor labels.
- Write calls added to each section's `client/api/*.ts` module.
- Long forms get their own lazily loaded routes; short ones open in dialogs. Controls only show for users with permission.

**Tech Stack:** React 19, React Router 7, TanStack Query 5, TypeScript 6, Tailwind 4, Tiptap 3 (`@tiptap/react`, `@tiptap/pm`, `@tiptap/starter-kit`), Vitest 5, Testing Library and jsdom, axe-core.

**Spec:** `docs/superpowers/specs/2026-09-30-content-editing-design.md`

## Global Constraints

- **Where to work:** the worktree `/Users/liam/Code/Go/AFC-design-system`, branch `content-editing`. Never touch `/Users/liam/Code/Go/AFC`, never read `postgres_*.sql`, and never `rm -rf` the current working directory.
- **Commits:** every commit message ends with a blank line and `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
- **Scope:** no Go changes.
- **New dependencies:** only `@tiptap/react`, `@tiptap/pm` and `@tiptap/starter-kit`, all `^3`. StarterKit 3 already includes Underline and Link (verified), so no separate extension packages.
- **Tiptap must never be in the entry chunk or any public page chunk.** It's loaded only via `React.lazy` in `RichTextEditor`, on edit routes.
- **Multipart field names are exactly the API's:**
  - `title`, `content`, `image`, `removeImage`, `dateOfEvent`;
  - `name`, `ages`, `description`, `league`, `division`, `leagueTable`, `fixtures`, `coach`, `physio`, `isActive`, `isYouth`;
  - `file`, `date`, `seasonId`, `website`, `purpose`, `team`, `caption`.
  - JSON bodies: `{ content }` for info, `{ name }` for seasons.
- **Value formats:**
  - booleans are sent as `'true'`/`'false'`;
  - dates as `YYYY-MM-DD`;
  - `ages` as an integer from 6 to 19 ("Under 6" … "Under 18", and "Over 18" = 19);
  - sponsor `team` as `''`, `A`, `O`, `Y` or a team id.
- **Permissions:**
  - Add, Edit and Delete controls show only when `user.permissions.canEdit`;
  - the gallery uses `canManageGallery`.
- **Toasts:** "Article saved", "Article deleted", "Event saved", "Event deleted", "Club information saved", "Team saved", "Team deleted", "Document added", "Document deleted", "Programme added", "Programme deleted", "Season added", "Season renamed", "Season deleted", "Sponsor added", "Sponsor deleted", "Affiliation added", "Affiliation deleted", "Photo added", "Photo deleted".
- **Other copy:**
  - "You don't have permission to edit this"
  - "Remove image"
  - "Choose an image file (JPEG, PNG, GIF, WebP, AVIF, APNG or SVG)."
  - "That file is too large (15 MB maximum)."
  - "Couldn't delete: <message>"
- **Style and lint:** token classes only, one component per `.tsx`, and no barrel files. Run `yarn eslint --fix <files>` **before** a task's tests, then commit.
- **Test helpers:**
  - `mockFetch` routes are keyed by path, including the query string;
  - route functions are evaluated per request, so a test can switch responses;
  - `fetchMock.mock.calls[i][1]` is the `RequestInit`, with `body` as a `FormData` or a JSON string.

## Review Focus

1. **Editing an article or team and saving without touching the image:** no `image` and no `removeImage` are sent, so the current image stays. Pinned in Task 4 (`newsForm`/`teamForm` omit them) and Task 5.
2. **Unticking "Active team" on edit:** `isActive=false` is sent explicitly, never omitted. Pinned in Tasks 4 and 8.
3. **A Manager or a signed-out visitor who knows an edit URL** (`/news/1/edit`): they see the permission message and no form. Pinned in Tasks 2 and 5.
4. **A photographer-style user** (`canManageGallery` without `canEdit`): they can add and delete photos but see no other edit controls. Pinned in Task 10.
5. **Deleting a season that has programmes:** the confirmation warns that the programmes stay, and the list refreshes to show them under "No season". Pinned in Task 9.

---

### Task 1: Helpers and the two 4a display fixes

**Files:**
- Create: `client/lib/images.ts`, `client/lib/editForm.ts`, `client/lib/sponsorTeam.ts`, `client/lib/saveErrors.ts`
- Modify:
  - `client/lib/format.ts` (adds the `'dayDate'` style)
  - `client/pages/account/PhotoCard.tsx` (imports from `lib/images`)
  - `client/pages/home/HomePage.tsx`, `client/pages/whatson/WhatsOnPage.tsx`, `client/pages/whatson/EventPage.tsx` (use `'dayDate'`)
  - `client/pages/sponsors/SponsorsPage.tsx` (`sponsorTeamLabel`)
  - `client/test/fixtures.ts` (realistic `dateOfEvent`, sponsor team as an id)
  - the tests that assert the old strings: `HomePage.test.tsx`, `whatson.test.tsx`
- Test: `client/lib/editing-helpers.test.ts`, `client/lib/helpers.test.ts` (adds `dayDate`)

**Interfaces:**
- Produces:
  - `IMAGE_TYPE_LIST: string[]`, `IMAGE_TYPES: string`, `IMAGE_TYPE_MESSAGE: string`
  - `isAcceptedImage(f: File): boolean`, `previewUrl(f: File): string | null`
  - `interface ImageValue { file: File | null; remove: boolean }`, `emptyImage: ImageValue`
  - `formData(fields: Record<string, string | number | boolean | File | null | undefined>): FormData`
  - `toDateInput(iso: string): string` (London date as `YYYY-MM-DD`, or `''`)
  - `formatSize(bytes: number): string`
  - `SPONSOR_TEAM_CHOICES: { value: string; label: string }[]` (the fixed codes)
  - `sponsorTeamLabel(team: string | undefined, teams: { id: number; name: string }[]): string | undefined`
  - `describeSaveError(err: unknown): string`
  - `formatDate(iso, 'dayDate')`, e.g. "Fri 16 Oct 2026"

- [ ] **Step 1: Write the failing tests**

`client/lib/editing-helpers.test.ts`:

```ts
import { describe, expect, it } from 'vitest';

import { ApiError } from '../api/client';
import { formatSize, formData, toDateInput } from './editForm';
import { IMAGE_TYPE_MESSAGE, isAcceptedImage } from './images';
import { describeSaveError } from './saveErrors';
import { SPONSOR_TEAM_CHOICES, sponsorTeamLabel } from './sponsorTeam';

describe('formData', () => {
  it('sends strings (even empty), numbers, booleans as true/false and files; skips null/undefined', () => {
    const file = new File(['x'], 'a.png', { type: 'image/png' });
    const fd = formData({
      title: 'Hi',
      content: '',
      ages: 12,
      isActive: false,
      isYouth: true,
      image: file,
      removeImage: undefined,
      seasonId: null,
    });
    expect(fd.get('title')).toBe('Hi');
    expect(fd.get('content')).toBe('');
    expect(fd.get('ages')).toBe('12');
    expect(fd.get('isActive')).toBe('false');
    expect(fd.get('isYouth')).toBe('true');
    expect(fd.get('image')).toBe(file);
    expect(fd.has('removeImage')).toBe(false);
    expect(fd.has('seasonId')).toBe(false);
  });
});

describe('dates and sizes', () => {
  it('turns an API date into a date-input value in UK time', () => {
    expect(toDateInput('2026-10-16T00:00:00Z')).toBe('2026-10-16');
    expect(toDateInput('2026-06-30T23:30:00Z')).toBe('2026-07-01');
    expect(toDateInput('nonsense')).toBe('');
  });

  it('formats file sizes', () => {
    expect(formatSize(512)).toBe('512 B');
    expect(formatSize(2048)).toBe('2 KB');
    expect(formatSize(1_572_864)).toBe('1.5 MB');
  });
});

describe('images', () => {
  it('accepts only the server image types', () => {
    expect(isAcceptedImage(new File([''], 'a.png', { type: 'image/png' }))).toBe(true);
    expect(isAcceptedImage(new File([''], 'a.html', { type: 'text/html' }))).toBe(false);
    expect(IMAGE_TYPE_MESSAGE).toBe(
      'Choose an image file (JPEG, PNG, GIF, WebP, AVIF, APNG or SVG).',
    );
  });
});

describe('sponsor team', () => {
  const teams = [{ id: 7, name: 'First Team' }];

  it('labels the codes and team ids', () => {
    expect(sponsorTeamLabel('A', teams)).toBe('All teams');
    expect(sponsorTeamLabel('O', teams)).toBe('Adult teams');
    expect(sponsorTeamLabel('Y', teams)).toBe('Youth teams');
    expect(sponsorTeamLabel('7', teams)).toBe('First Team');
    expect(sponsorTeamLabel('99', teams)).toBeUndefined();
    expect(sponsorTeamLabel('', teams)).toBeUndefined();
    expect(sponsorTeamLabel(undefined, teams)).toBeUndefined();
  });

  it('offers None, All, Adult and Youth as the fixed choices', () => {
    expect(SPONSOR_TEAM_CHOICES.map((c) => c.value)).toEqual(['', 'A', 'O', 'Y']);
  });
});

describe('describeSaveError', () => {
  it('explains too-large uploads and passes other messages through', () => {
    expect(describeSaveError(new ApiError(413, 'Request Entity Too Large'))).toBe(
      'That file is too large (15 MB maximum).',
    );
    expect(describeSaveError(new ApiError(500, 'boom'))).toBe('boom');
    expect(describeSaveError('?')).toBe('Something went wrong. Please try again.');
  });
});
```

Add to `client/lib/helpers.test.ts`, inside `describe('formatDate', …)`:

```ts
  it('formats a date with its weekday and no time', () => {
    expect(formatDate('2026-10-16T00:00:00Z', 'dayDate')).toBe('Fri 16 Oct 2026');
    expect(formatDate('2026-12-05T00:00:00Z', 'dayDate')).toBe('Sat 5 Dec 2026');
  });
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/lib`
Expected: FAIL. `./editForm`, `./images`, `./saveErrors` and `./sponsorTeam` can't be resolved; `'dayDate'` is not handled.

- [ ] **Step 3: Implement the helpers**

`client/lib/images.ts`:

```ts
// The server's accepted image types (server/internal/upload).
export const IMAGE_TYPE_LIST = [
  'image/jpeg',
  'image/png',
  'image/gif',
  'image/webp',
  'image/avif',
  'image/apng',
  'image/svg+xml',
];
export const IMAGE_TYPES = IMAGE_TYPE_LIST.join(',');
export const IMAGE_TYPE_MESSAGE = 'Choose an image file (JPEG, PNG, GIF, WebP, AVIF, APNG or SVG).';

export function isAcceptedImage(file: File): boolean {
  return IMAGE_TYPE_LIST.includes(file.type);
}

/**
 * A local preview URL for a chosen image. Only a browser-issued blob: URL is
 * ever used as an <img> source; anything else is released and not shown.
 */
export function previewUrl(file: File): string | null {
  const url = URL.createObjectURL(file);
  if (url.startsWith('blob:')) {
    return url;
  }
  URL.revokeObjectURL(url);
  return null;
}

/** An image form field's value: a newly chosen file, and/or "remove the current one". */
export interface ImageValue {
  file: File | null;
  remove: boolean;
}

export const emptyImage: ImageValue = { file: null, remove: false };
```

`client/lib/editForm.ts`:

```ts
type FieldValue = string | number | boolean | File | null | undefined;

/**
 * A multipart body in the API's conventions: strings (including empty, which
 * clears a field on PATCH) and numbers as-is, booleans as "true"/"false",
 * files as files; null/undefined fields are left out entirely.
 */
export function formData(fields: Record<string, FieldValue>): FormData {
  const fd = new FormData();
  for (const [name, value] of Object.entries(fields)) {
    if (value === null || value === undefined) {
      continue;
    }
    if (value instanceof File) {
      fd.append(name, value);
    } else if (typeof value === 'boolean') {
      fd.append(name, value ? 'true' : 'false');
    } else {
      fd.append(name, String(value));
    }
  }
  return fd;
}

const dateParts = new Intl.DateTimeFormat('en-GB', {
  timeZone: 'Europe/London',
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
});

/** An API date as an <input type="date"> value (YYYY-MM-DD, UK time), or '' if invalid. */
export function toDateInput(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) {
    return '';
  }
  const p: Record<string, string> = {};
  for (const part of dateParts.formatToParts(d)) {
    p[part.type] = part.value;
  }
  return `${p.year}-${p.month}-${p.day}`;
}

export function formatSize(bytes: number): string {
  if (bytes < 1024) {
    return `${bytes} B`;
  }
  if (bytes < 1024 * 1024) {
    return `${Math.round(bytes / 1024)} KB`;
  }
  return `${(bytes / (1024 * 1024)).toFixed(1).replace(/\.0$/, '')} MB`;
}
```

`client/lib/sponsorTeam.ts`:

```ts
/** The API's fixed sponsor "team" codes, plus "no team". Team ids follow in the select. */
export const SPONSOR_TEAM_CHOICES = [
  { value: '', label: 'None' },
  { value: 'A', label: 'All teams' },
  { value: 'O', label: 'Adult teams' },
  { value: 'Y', label: 'Youth teams' },
];

const CODES: Record<string, string> = { A: 'All teams', O: 'Adult teams', Y: 'Youth teams' };

/** What a sponsor's raw `team` value means, for display; undefined when there's nothing to show. */
export function sponsorTeamLabel(
  team: string | undefined,
  teams: { id: number; name: string }[],
): string | undefined {
  if (!team) {
    return undefined;
  }
  if (CODES[team]) {
    return CODES[team];
  }
  return teams.find((t) => String(t.id) === team)?.name;
}
```

`client/lib/saveErrors.ts`:

```ts
import { ApiError } from '../api/client';

/** A save/delete failure as a sentence for the person who pressed the button. */
export function describeSaveError(err: unknown): string {
  if (err instanceof ApiError && err.status === 413) {
    return 'That file is too large (15 MB maximum).';
  }
  return err instanceof Error ? err.message : 'Something went wrong. Please try again.';
}
```

`client/lib/format.ts`:
- widen the signature to `style: 'date' | 'dateTime' | 'dayDate' = 'date'`;
- after `if (style === 'date') { return date; }` add:

```ts
  if (style === 'dayDate') {
    return `${p.weekday} ${date}`;
  }
```

- add `'dayDate'` → "Fri 16 Oct 2026" to the doc comment.

`client/pages/account/PhotoCard.tsx`: delete its local `IMAGE_TYPE_LIST`, `IMAGE_TYPES` and `previewUrl` declarations and the image-type message literal. Import them instead: `import { IMAGE_TYPE_LIST, IMAGE_TYPE_MESSAGE, IMAGE_TYPES, previewUrl } from '../../lib/images';`, and use `IMAGE_TYPE_MESSAGE` where the string was.

- [ ] **Step 4: Apply the display fixes and update fixtures and tests**

- `client/pages/home/HomePage.tsx`, `client/pages/whatson/WhatsOnPage.tsx` and `client/pages/whatson/EventPage.tsx`: change every `formatDate(…dateOfEvent, 'dateTime')` to `formatDate(…dateOfEvent, 'dayDate')`.
- `client/pages/sponsors/SponsorsPage.tsx`:
  - import `useSite` from `../../api/queries` and `sponsorTeamLabel` from `../../lib/sponsorTeam`;
  - add `const site = useSite();` after `const sponsors = useSponsors();`;
  - replace `{s.team && <p className="text-sm text-muted">Sponsor of {s.team}</p>}` with:

```tsx
                  {(() => {
                    const label = sponsorTeamLabel(s.team, site.data?.teams ?? []);
                    return label && <p className="text-sm text-muted">Sponsor of {label}</p>;
                  })()}
```

- `client/test/fixtures.ts`:
  - set the `event` fixture's `dateOfEvent` to `'2026-10-16T00:00:00Z'` (events store a date only);
  - set the `sponsor` fixture's `team` to `'7'` (the First Team's id).
- Tests: `client/pages/home/HomePage.test.tsx` and `client/pages/whatson/whatson.test.tsx` currently expect `'Fri 16 Oct 2026, 7pm'`; change every occurrence to `'Fri 16 Oct 2026'`. `client/pages/sponsors/sponsors.test.tsx` keeps expecting "Sponsor of First Team", which now comes from the id mapping.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `yarn eslint --fix client/lib client/pages client/test && yarn vitest run client/lib && yarn test:client`
Expected: PASS. All the new helper tests pass, and the whole suite is green.

- [ ] **Step 6: Commit**

```bash
yarn typecheck
git add -A client
git commit -q -m "Add editing helpers; show event dates without a time and sponsor teams by name" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 2: Editing toolkit

**Files:**
- Create in `client/components/edit/`: `useCanEdit.ts`, `RequireEditor.tsx`, `useSaveForm.ts`, `ImageField.tsx`, `FileField.tsx`, `DeleteButton.tsx`
- Create: `client/test/Location.tsx`
- Modify: `client/test/fixtures.ts` (a `photographer` fixture)
- Test: `client/components/edit/toolkit.test.tsx`

**Interfaces:**
- Consumes (Task 1): `ImageValue`, `IMAGE_TYPES`, `IMAGE_TYPE_MESSAGE`, `isAcceptedImage`, `previewUrl`, `describeSaveError`, `formatSize`; `isSessionExpired` (4b, `client/lib/session.ts`).
- Produces:
  - `useCanEdit(): { canEdit: boolean; canManageGallery: boolean }`
  - `RequireEditor({ permission?: 'canEdit' | 'canManageGallery'; children })`
  - `useSaveForm<T>({ submit: () => Promise<T>; invalidate?: QueryKey[]; onSaved?: (r: T) => void }): { run(): Promise<void>; busy: boolean; fieldErrors: Record<string, string>; formError: string | null }`
  - `ImageField({ label, currentUrl?, required?, allowRemove?, value: ImageValue, onChange(v: ImageValue), error? })`
  - `FileField({ label, value: File | null, onChange(f: File | null), error?, accept? })`
  - `DeleteButton({ confirmTitle, confirmMessage, onDelete, invalidate?, successMessage, after?, children?, ariaLabel?, size? })`
  - the test component `Location` (renders `pathname + search` in `data-testid="location"`)
  - the fixture `photographer`

- [ ] **Step 1: Write the failing tests**

`client/test/Location.tsx`:

```tsx
import { useLocation } from 'react-router';

/** Test helper: shows the router's current path so tests can assert navigation. */
export function Location() {
  const l = useLocation();
  return <output data-testid="location">{l.pathname + l.search}</output>;
}
```

In `client/test/fixtures.ts`, add after `manager`:

```ts
export const photographer: MockResponse = signedIn('Pat Photographer', 'Photographer', false, true);
```

`client/components/edit/toolkit.test.tsx`:

```tsx
import { fireEvent, screen, waitFor } from '@testing-library/react';
import { useState } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { ApiError } from '../../api/client';
import { anonymous, editor, manager, photographer, publicRoutes } from '../../test/fixtures';
import type { MockResponse } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import { emptyImage, type ImageValue } from '../../lib/images';
import { DeleteButton } from './DeleteButton';
import { FileField } from './FileField';
import { ImageField } from './ImageField';
import { RequireEditor } from './RequireEditor';
import { useCanEdit } from './useCanEdit';
import { useSaveForm } from './useSaveForm';

let revokeObjectURL: ReturnType<typeof vi.fn>;
beforeEach(() => {
  let n = 0;
  revokeObjectURL = vi.fn();
  Object.assign(URL, { createObjectURL: vi.fn(() => `blob:p-${++n}`), revokeObjectURL });
});
afterEach(() => {
  delete (URL as unknown as Record<string, unknown>).createObjectURL;
  delete (URL as unknown as Record<string, unknown>).revokeObjectURL;
});

function CanEditProbe() {
  const { canEdit, canManageGallery } = useCanEdit();
  return <p>{`edit:${canEdit} gallery:${canManageGallery}`}</p>;
}

describe('useCanEdit', () => {
  it.each([
    [anonymous, 'edit:false gallery:false'],
    [manager, 'edit:false gallery:false'],
    [photographer, 'edit:false gallery:true'],
    [editor, 'edit:true gallery:true'],
  ] as [MockResponse, string][])('reflects permissions', async (me, text) => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': me }));
    renderWithProviders(<CanEditProbe />);
    expect(await screen.findByText(text)).toBeInTheDocument();
  });
});

describe('RequireEditor', () => {
  it('shows the permission message to non-editors', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': manager }));
    renderWithProviders(
      <RequireEditor>
        <p>secret form</p>
      </RequireEditor>,
    );
    expect(await screen.findByText("You don't have permission to edit this")).toBeInTheDocument();
    expect(screen.queryByText('secret form')).toBeNull();
  });

  it('renders the children for editors, and gallery access for photographers', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': photographer }));
    renderWithProviders(
      <>
        <RequireEditor permission="canManageGallery">
          <p>photo form</p>
        </RequireEditor>
        <RequireEditor>
          <p>article form</p>
        </RequireEditor>
      </>,
    );
    expect(await screen.findByText('photo form')).toBeInTheDocument();
    expect(screen.queryByText('article form')).toBeNull();
  });
});

describe('useSaveForm', () => {
  function Form({ submit, onSaved }: { submit: () => Promise<unknown>; onSaved?: () => void }) {
    const save = useSaveForm({ submit, invalidate: [['news']], onSaved });
    return (
      <>
        <button onClick={() => void save.run()}>Save</button>
        <p>busy:{String(save.busy)}</p>
        <p>field:{save.fieldErrors.title ?? ''}</p>
        <p>form:{save.formError ?? ''}</p>
      </>
    );
  }

  it('invalidates and calls onSaved on success', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': editor }));
    const onSaved = vi.fn();
    const { queryClient } = renderWithProviders(
      <Form submit={async () => ({ id: 1 })} onSaved={onSaved} />,
    );
    const spy = vi.spyOn(queryClient, 'invalidateQueries');
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    await waitFor(() => expect(onSaved).toHaveBeenCalledWith({ id: 1 }));
    expect(spy).toHaveBeenCalledWith({ queryKey: ['news'] });
  });

  it('puts field errors on fields and explains 413s', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': editor }));
    let n = 0;
    renderWithProviders(
      <Form
        submit={async () => {
          n++;
          if (n === 1) throw new ApiError(422, 'invalid', { title: 'title is required' });
          throw new ApiError(413, 'Request Entity Too Large');
        }}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    expect(await screen.findByText('field:title is required')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    expect(
      await screen.findByText('form:That file is too large (15 MB maximum).'),
    ).toBeInTheDocument();
  });

  it('re-reads the session on a 401', async () => {
    let me: MockResponse = editor;
    const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': () => me }));
    renderWithProviders(
      <Form
        submit={async () => {
          me = anonymous;
          throw new ApiError(401, 'login required');
        }}
      />,
    );
    await screen.findByText('busy:false');
    const before = fetchMock.mock.calls.filter(([u]) => String(u).endsWith('/auth/me')).length;
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    await waitFor(() =>
      expect(
        fetchMock.mock.calls.filter(([u]) => String(u).endsWith('/auth/me')).length,
      ).toBeGreaterThan(before),
    );
    expect(screen.getByText('form:')).toBeInTheDocument();
  });
});

describe('ImageField', () => {
  function Harness({ current }: { current?: string }) {
    const [value, setValue] = useState<ImageValue>(emptyImage);
    return (
      <>
        <ImageField label="Image" currentUrl={current} allowRemove value={value} onChange={setValue} />
        <p>file:{value.file?.name ?? ''} remove:{String(value.remove)}</p>
      </>
    );
  }

  function pick(file: File) {
    fireEvent.change(screen.getByLabelText('Image'), { target: { files: [file] } });
  }

  it('previews a chosen image and reports it', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<Harness current="/img/1" />);
    pick(new File(['x'], 'a.png', { type: 'image/png' }));
    expect(screen.getByRole('img', { name: 'Preview of the new image' })).toHaveAttribute(
      'src',
      'blob:p-1',
    );
    expect(screen.getByText('file:a.png remove:false')).toBeInTheDocument();
  });

  it('refuses a non-image without previewing it', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<Harness />);
    pick(new File(['<b>'], 'a.html', { type: 'text/html' }));
    expect(screen.getByLabelText('Image')).toHaveAccessibleDescription(
      'Choose an image file (JPEG, PNG, GIF, WebP, AVIF, APNG or SVG).',
    );
    expect(screen.queryByRole('img', { name: 'Preview of the new image' })).toBeNull();
    expect(URL.createObjectURL).not.toHaveBeenCalled();
  });

  it('offers Remove image for a current image, and a new file clears it', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<Harness current="/img/1" />);
    fireEvent.click(screen.getByRole('checkbox', { name: 'Remove image' }));
    expect(screen.getByText('file: remove:true')).toBeInTheDocument();
    pick(new File(['x'], 'b.png', { type: 'image/png' }));
    expect(screen.getByText('file:b.png remove:false')).toBeInTheDocument();
    pick(new File(['x'], 'c.png', { type: 'image/png' }));
    expect(revokeObjectURL).toHaveBeenCalledWith('blob:p-1');
  });

  it('has no Remove image without a current image', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<Harness />);
    expect(screen.queryByRole('checkbox', { name: 'Remove image' })).toBeNull();
  });
});

describe('FileField', () => {
  it('shows the chosen file name and size', async () => {
    function Harness() {
      const [file, setFile] = useState<File | null>(null);
      return <FileField label="File" value={file} onChange={setFile} />;
    }
    mockFetch(publicRoutes());
    renderWithProviders(<Harness />);
    fireEvent.change(screen.getByLabelText('File'), {
      target: { files: [new File(['x'.repeat(2048)], 'rules.pdf', { type: 'application/pdf' })] },
    });
    expect(screen.getByLabelText('File')).toHaveAccessibleDescription('rules.pdf (2 KB)');
  });
});

describe('DeleteButton', () => {
  it('confirms, deletes, toasts and calls after', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': editor }));
    const onDelete = vi.fn(async () => undefined);
    const after = vi.fn();
    renderWithProviders(
      <DeleteButton
        ariaLabel="Delete Club rules"
        confirmTitle="Delete Club rules?"
        confirmMessage="This can't be undone."
        onDelete={onDelete}
        successMessage="Document deleted"
        after={after}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Delete Club rules' }));
    const dialog = screen.getByRole('dialog', { name: 'Delete Club rules?' });
    fireEvent.click(screen.getAllByRole('button', { name: 'Delete' }).at(-1) as HTMLElement);
    expect(await screen.findByRole('button', { name: 'Document deleted' })).toBeInTheDocument();
    expect(onDelete).toHaveBeenCalledOnce();
    expect(after).toHaveBeenCalledOnce();
    expect(dialog).not.toBeInTheDocument();
  });

  it('does nothing on Cancel, and toasts a failure', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': editor }));
    const onDelete = vi.fn(async () => {
      throw new ApiError(500, 'boom');
    });
    renderWithProviders(
      <DeleteButton
        confirmTitle="Delete it?"
        confirmMessage="Sure?"
        onDelete={onDelete}
        successMessage="Deleted"
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Delete' }));
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(onDelete).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: 'Delete' }));
    fireEvent.click(screen.getAllByRole('button', { name: 'Delete' }).at(-1) as HTMLElement);
    expect(await screen.findByRole('button', { name: "Couldn't delete: boom" })).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/components/edit`
Expected: FAIL. The toolkit modules can't be resolved.

- [ ] **Step 3: Implement**

`client/components/edit/useCanEdit.ts`:

```ts
import { useAuth } from '../../auth/useAuth';

/** What the signed-in user may edit (all false when signed out). The server enforces the same rules. */
export function useCanEdit(): { canEdit: boolean; canManageGallery: boolean } {
  const { user } = useAuth();
  return {
    canEdit: user?.permissions.canEdit ?? false,
    canManageGallery: user?.permissions.canManageGallery ?? false,
  };
}
```

`client/components/edit/RequireEditor.tsx`:

```tsx
import type { ReactNode } from 'react';
import { useNavigate } from 'react-router';

import { useAuth } from '../../auth/useAuth';
import { PageSkeleton } from '../page/QueryState';
import { Button } from '../ui/Button';
import { EmptyState } from '../ui/EmptyState';

export interface RequireEditorProps {
  permission?: 'canEdit' | 'canManageGallery';
  children: ReactNode;
}

/** Gate for edit routes: loading placeholder, a permission message, or the form. */
export function RequireEditor({ permission = 'canEdit', children }: RequireEditorProps) {
  const { user, isLoading } = useAuth();
  const navigate = useNavigate();
  if (isLoading) {
    return <PageSkeleton />;
  }
  if (!user?.permissions[permission]) {
    return (
      <EmptyState
        title="You don't have permission to edit this"
        message="Sign in with an editor account, or go back."
        action={
          <Button
            variant="secondary"
            onClick={() => (window.history.length > 1 ? navigate(-1) : navigate('/'))}
          >
            Back
          </Button>
        }
      />
    );
  }
  return <>{children}</>;
}
```

`client/components/edit/useSaveForm.ts`:

```ts
import { useQueryClient, type QueryKey } from '@tanstack/react-query';
import { useState } from 'react';

import { ApiError } from '../../api/client';
import { useAuth } from '../../auth/useAuth';
import { describeSaveError } from '../../lib/saveErrors';
import { isSessionExpired } from '../../lib/session';

export interface SaveFormOptions<T> {
  submit: () => Promise<T>;
  /** Query keys (or prefixes) to refresh after a successful save. */
  invalidate?: QueryKey[];
  onSaved?: (result: T) => void;
}

/** Busy state, server field errors, and the after-save refresh for an edit form. */
export function useSaveForm<T>({ submit, invalidate = [], onSaved }: SaveFormOptions<T>) {
  const queryClient = useQueryClient();
  const { refresh } = useAuth();
  const [busy, setBusy] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState<string | null>(null);

  async function run() {
    setBusy(true);
    setFieldErrors({});
    setFormError(null);
    try {
      const result = await submit();
      for (const queryKey of invalidate) {
        void queryClient.invalidateQueries({ queryKey });
      }
      onSaved?.(result);
    } catch (err) {
      if (isSessionExpired(err)) {
        await refresh();
      } else if (err instanceof ApiError && Object.keys(err.fields).length > 0) {
        setFieldErrors(err.fields);
      } else {
        setFormError(describeSaveError(err));
      }
    } finally {
      setBusy(false);
    }
  }

  return { run, busy, fieldErrors, formError };
}
```

`client/components/edit/ImageField.tsx`:

```tsx
import { useEffect, useRef, useState } from 'react';

import {
  IMAGE_TYPE_MESSAGE,
  IMAGE_TYPES,
  isAcceptedImage,
  previewUrl,
  type ImageValue,
} from '../../lib/images';
import { ImageWithFallback } from '../page/ImageWithFallback';
import { Checkbox } from '../ui/Checkbox';
import { FileInput } from '../ui/controls';
import { Field } from '../ui/Field';

export interface ImageFieldProps {
  label: string;
  /** The image already saved, if any. */
  currentUrl?: string;
  required?: boolean;
  /** Offer "Remove image" when there's a current image (edit forms). */
  allowRemove?: boolean;
  value: ImageValue;
  onChange: (value: ImageValue) => void;
  error?: string;
}

const box = 'h-40 w-full max-w-sm rounded-lg border border-line';

export function ImageField({
  label,
  currentUrl,
  required,
  allowRemove,
  value,
  onChange,
  error,
}: ImageFieldProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const [typeError, setTypeError] = useState<string | undefined>();

  useEffect(() => {
    if (!preview) {
      return;
    }
    return () => URL.revokeObjectURL(preview);
  }, [preview]);

  function pick(file: File | null) {
    setTypeError(undefined);
    if (file && !isAcceptedImage(file)) {
      setTypeError(IMAGE_TYPE_MESSAGE);
      setPreview(null);
      if (inputRef.current) {
        inputRef.current.value = '';
      }
      onChange({ file: null, remove: value.remove });
      return;
    }
    setPreview(file ? previewUrl(file) : null);
    onChange({ file, remove: file ? false : value.remove });
  }

  const shown = preview ?? (value.remove ? undefined : currentUrl);
  return (
    <div className="space-y-2">
      <ImageWithFallback
        src={shown}
        alt={preview ? 'Preview of the new image' : ''}
        className={`${box} object-cover`}
        fallback={
          <div aria-hidden="true" data-fallback="" className={`${box} bg-linear-135 from-blue to-red`} />
        }
      />
      <Field label={label} error={typeError ?? error}>
        <FileInput
          ref={inputRef}
          accept={IMAGE_TYPES}
          required={required}
          onChange={(e) => pick(e.target.files?.[0] ?? null)}
        />
      </Field>
      {allowRemove && currentUrl && !value.file && (
        <Checkbox
          label="Remove image"
          checked={value.remove}
          onChange={(e) => onChange({ file: null, remove: e.target.checked })}
        />
      )}
    </div>
  );
}
```

`client/components/edit/FileField.tsx`:

```tsx
import { formatSize } from '../../lib/editForm';
import { FileInput } from '../ui/controls';
import { Field } from '../ui/Field';

export interface FileFieldProps {
  label: string;
  value: File | null;
  onChange: (file: File | null) => void;
  error?: string;
  accept?: string;
}

export function FileField({ label, value, onChange, error, accept }: FileFieldProps) {
  return (
    <Field
      label={label}
      error={error}
      help={value ? `${value.name} (${formatSize(value.size)})` : undefined}
    >
      <FileInput accept={accept} onChange={(e) => onChange(e.target.files?.[0] ?? null)} />
    </Field>
  );
}
```

`client/components/edit/DeleteButton.tsx`:

```tsx
import { useQueryClient, type QueryKey } from '@tanstack/react-query';
import { useState, type ReactNode } from 'react';

import { useAuth } from '../../auth/useAuth';
import { describeSaveError } from '../../lib/saveErrors';
import { isSessionExpired } from '../../lib/session';
import { Button } from '../ui/Button';
import type { ButtonSize } from '../ui/buttonStyles';
import { ConfirmDialog } from '../ui/ConfirmDialog';
import { useToast } from '../ui/toast/useToast';

export interface DeleteButtonProps {
  confirmTitle: string;
  confirmMessage: ReactNode;
  onDelete: () => Promise<unknown>;
  invalidate?: QueryKey[];
  successMessage: string;
  after?: () => void;
  /** Visible text; defaults to "Delete". */
  children?: ReactNode;
  /** Accessible name when the text alone is ambiguous (e.g. "Delete Club rules"). */
  ariaLabel?: string;
  size?: ButtonSize;
}

export function DeleteButton({
  confirmTitle,
  confirmMessage,
  onDelete,
  invalidate = [],
  successMessage,
  after,
  children = 'Delete',
  ariaLabel,
  size = 'sm',
}: DeleteButtonProps) {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();
  const toast = useToast();
  const { refresh } = useAuth();

  async function confirm() {
    try {
      await onDelete();
      for (const queryKey of invalidate) {
        void queryClient.invalidateQueries({ queryKey });
      }
      setOpen(false);
      toast.show({ tone: 'success', message: successMessage });
      after?.();
    } catch (err) {
      setOpen(false);
      if (isSessionExpired(err)) {
        await refresh();
        return;
      }
      toast.show({ tone: 'error', message: `Couldn't delete: ${describeSaveError(err)}` });
    }
  }

  return (
    <>
      <Button variant="danger" size={size} aria-label={ariaLabel} onClick={() => setOpen(true)}>
        {children}
      </Button>
      <ConfirmDialog
        open={open}
        title={confirmTitle}
        message={confirmMessage}
        confirmLabel="Delete"
        tone="danger"
        onConfirm={confirm}
        onCancel={() => setOpen(false)}
      />
    </>
  );
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/components/edit client/test && yarn vitest run client/components/edit`
Expected: PASS, 15 tests.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add client/components/edit client/test
git commit -q -m "Add the editing toolkit: permissions, save hook, image/file fields, delete button" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 3: Rich-text editor

**Files:**
- Modify: `package.json`, `yarn.lock` (`yarn add @tiptap/react @tiptap/pm @tiptap/starter-kit`), `client/test/setup.ts` (Range polyfill)
- Create: `client/lib/links.ts`, `client/components/edit/RichTextEditor.tsx`, `client/components/edit/RichTextEditorImpl.tsx`, `client/components/edit/richText.ts`, `client/test/editor.ts`
- Test: `client/components/edit/RichTextEditor.test.tsx`

**Interfaces:**
- Produces:
  - `normalizeLink(input: string): string | null`
  - `RichTextEditor({ label: string; value: string; onChange(html: string): void; error?: string })`, lazily loaded
  - the test helpers `editorFor(label)` (async) and `typeInEditor(label, html)` (async; replaces the content and emits an update)

- [ ] **Step 1: Add the dependency and the jsdom polyfill**

```bash
cd /Users/liam/Code/Go/AFC-design-system && yarn add @tiptap/react @tiptap/pm @tiptap/starter-kit
```

In `client/test/setup.ts`, after `installDialogPolyfill();`, add the lines below. ProseMirror measures ranges when it scrolls the selection into view, and jsdom has no layout:

```ts
// ProseMirror measures ranges when scrolling the selection into view; jsdom has no layout.
Range.prototype.getClientRects = () =>
  ({ length: 0, item: () => null, [Symbol.iterator]: [][Symbol.iterator] }) as unknown as DOMRectList;
Range.prototype.getBoundingClientRect = () => new DOMRect();
```

- [ ] **Step 2: Write the failing tests**

`client/test/editor.ts`:

```ts
import { act, screen } from '@testing-library/react';
import type { Editor } from '@tiptap/react';

/** The Tiptap editor behind a labelled editable area (Tiptap puts it on the DOM node). */
export async function editorFor(label: string): Promise<Editor> {
  const el = await screen.findByLabelText(label, undefined, { timeout: 3000 });
  return (el as unknown as { editor: Editor }).editor;
}

/** Replaces the editor's content, as if typed, so onUpdate fires. */
export async function typeInEditor(label: string, html: string): Promise<void> {
  const editor = await editorFor(label);
  await act(async () => {
    editor.commands.selectAll();
    editor.commands.insertContent(html);
  });
}
```

`client/components/edit/RichTextEditor.test.tsx`:

```tsx
import { act, fireEvent, screen, within } from '@testing-library/react';
import { useState } from 'react';
import { describe, expect, it } from 'vitest';

import { normalizeLink } from '../../lib/links';
import { editorFor, typeInEditor } from '../../test/editor';
import { publicRoutes } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import { RichTextEditor } from './RichTextEditor';

function Harness({ initial = '' }: { initial?: string }) {
  const [html, setHtml] = useState(initial);
  return (
    <>
      <RichTextEditor label="Content" value={initial} onChange={setHtml} />
      <output data-testid="html">{html}</output>
    </>
  );
}

function renderEditor(initial?: string) {
  mockFetch(publicRoutes());
  renderWithProviders(<Harness initial={initial} />);
}

describe('normalizeLink', () => {
  it('accepts web and email links and adds https:// to bare domains', () => {
    expect(normalizeLink('https://league.example/table')).toBe('https://league.example/table');
    expect(normalizeLink('  league.example  ')).toBe('https://league.example');
    expect(normalizeLink('mailto:sec@example.test')).toBe('mailto:sec@example.test');
  });

  it('rejects other schemes and junk', () => {
    for (const bad of ['javascript:alert(1)', 'data:text/html,x', '', 'not a url']) {
      expect(normalizeLink(bad)).toBeNull();
    }
  });
});

describe('RichTextEditor', () => {
  it('loads existing HTML and emits HTML as you type', async () => {
    renderEditor('<p>Hello</p>');
    const editor = await editorFor('Content');
    expect(editor.getHTML()).toBe('<p>Hello</p>');
    await typeInEditor('Content', '<p>New <strong>text</strong></p>');
    expect(screen.getByTestId('html')).toHaveTextContent('<p>New <strong>text</strong></p>');
  });

  it('emits an empty string for an empty document', async () => {
    renderEditor('<p>Hello</p>');
    const editor = await editorFor('Content');
    await act(async () => {
      editor.commands.clearContent(true);
    });
    expect(screen.getByTestId('html')).toHaveTextContent(/^$/);
  });

  it('has a labelled toolbar whose buttons toggle formatting', async () => {
    renderEditor('<p>Hello</p>');
    const editor = await editorFor('Content');
    const toolbar = screen.getByRole('toolbar', { name: 'Content formatting' });
    await act(async () => {
      editor.commands.selectAll();
    });
    const bold = within(toolbar).getByRole('button', { name: 'Bold' });
    expect(bold).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(bold);
    expect(bold).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('html')).toHaveTextContent('<p><strong>Hello</strong></p>');
    fireEvent.click(within(toolbar).getByRole('button', { name: 'Bullet list' }));
    expect(screen.getByTestId('html')).toHaveTextContent('<ul><li><p><strong>Hello</strong></p></li></ul>');
  });

  it('adds a link through the link dialog, fixing a bare domain and refusing javascript:', async () => {
    renderEditor('<p>Table</p>');
    const editor = await editorFor('Content');
    await act(async () => {
      editor.commands.selectAll();
    });
    fireEvent.click(screen.getByRole('button', { name: 'Link' }));
    const dialog = screen.getByRole('dialog', { name: 'Add a link' });
    const input = within(dialog).getByLabelText('Link address');
    fireEvent.change(input, { target: { value: 'javascript:alert(1)' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add link' }));
    expect(input).toHaveAccessibleDescription(
      'Enter a web address (https://…) or an email link (mailto:…).',
    );
    fireEvent.change(input, { target: { value: 'league.example' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add link' }));
    expect(screen.queryByRole('dialog', { name: 'Add a link' })).toBeNull();
    expect(screen.getByTestId('html').textContent).toContain('href="https://league.example"');
  });

  it('shows an error under the editor', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(
      <RichTextEditor label="Content" value="" onChange={() => {}} error="content is too long" />,
    );
    const area = await screen.findByLabelText('Content', undefined, { timeout: 3000 });
    expect(area).toHaveAccessibleDescription('content is too long');
  });
});
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `yarn vitest run client/components/edit/RichTextEditor.test.tsx`
Expected: FAIL. `../../lib/links` and `./RichTextEditor` can't be resolved.

- [ ] **Step 4: Implement**

`client/lib/links.ts`:

```ts
/**
 * A link address as typed by an editor, made safe: http(s) and mailto only;
 * a bare domain ("league.example/table") gets https://. Anything else is null.
 */
export function normalizeLink(input: string): string | null {
  const value = input.trim();
  if (!value) {
    return null;
  }
  const hasScheme = /^[a-z][a-z0-9+.-]*:/i.test(value);
  const candidate = hasScheme ? value : `https://${value}`;
  try {
    const url = new URL(candidate);
    if (url.protocol === 'mailto:') {
      return url.href;
    }
    if ((url.protocol === 'https:' || url.protocol === 'http:') && url.hostname.includes('.')) {
      return hasScheme ? url.href : candidate;
    }
  } catch {
    return null;
  }
  return null;
}
```

`client/components/edit/richText.ts`:

```ts
export interface RichTextEditorProps {
  label: string;
  /** Initial HTML (read once when the editor mounts). */
  value: string;
  /** Receives the editor's HTML ('' for an empty document). */
  onChange: (html: string) => void;
  error?: string;
}
```

`client/components/edit/RichTextEditor.tsx`:

```tsx
import { lazy, Suspense } from 'react';

import { Skeleton } from '../ui/Skeleton';
import type { RichTextEditorProps } from './richText';

// Tiptap/ProseMirror are only downloaded when an edit form opens.
const RichTextEditorImpl = lazy(() => import('./RichTextEditorImpl'));

export function RichTextEditor(props: RichTextEditorProps) {
  return (
    <Suspense fallback={<Skeleton className="h-48" />}>
      <RichTextEditorImpl {...props} />
    </Suspense>
  );
}
```

`client/components/edit/RichTextEditorImpl.tsx`:

```tsx
import { EditorContent, useEditor, type Editor } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import { clsx } from 'clsx';
import { useEffect, useId, useState, type FormEvent } from 'react';

import { normalizeLink } from '../../lib/links';
import { Button } from '../ui/Button';
import { buttonClasses } from '../ui/buttonStyles';
import { Input } from '../ui/controls';
import { Field } from '../ui/Field';
import { Modal } from '../ui/Modal';
import type { RichTextEditorProps } from './richText';

interface Tool {
  label: string;
  text: string;
  isActive?: (e: Editor) => boolean;
  run: (e: Editor) => void;
}

const tools: Tool[] = [
  { label: 'Bold', text: 'B', isActive: (e) => e.isActive('bold'), run: (e) => e.chain().focus().toggleBold().run() },
  { label: 'Italic', text: 'I', isActive: (e) => e.isActive('italic'), run: (e) => e.chain().focus().toggleItalic().run() },
  { label: 'Underline', text: 'U', isActive: (e) => e.isActive('underline'), run: (e) => e.chain().focus().toggleUnderline().run() },
  { label: 'Strikethrough', text: 'S', isActive: (e) => e.isActive('strike'), run: (e) => e.chain().focus().toggleStrike().run() },
  { label: 'Heading 2', text: 'H2', isActive: (e) => e.isActive('heading', { level: 2 }), run: (e) => e.chain().focus().toggleHeading({ level: 2 }).run() },
  { label: 'Heading 3', text: 'H3', isActive: (e) => e.isActive('heading', { level: 3 }), run: (e) => e.chain().focus().toggleHeading({ level: 3 }).run() },
  { label: 'Bullet list', text: '•', isActive: (e) => e.isActive('bulletList'), run: (e) => e.chain().focus().toggleBulletList().run() },
  { label: 'Numbered list', text: '1.', isActive: (e) => e.isActive('orderedList'), run: (e) => e.chain().focus().toggleOrderedList().run() },
  { label: 'Quote', text: '❝', isActive: (e) => e.isActive('blockquote'), run: (e) => e.chain().focus().toggleBlockquote().run() },
  { label: 'Horizontal rule', text: '―', run: (e) => e.chain().focus().setHorizontalRule().run() },
  { label: 'Undo', text: '↶', run: (e) => e.chain().focus().undo().run() },
  { label: 'Redo', text: '↷', run: (e) => e.chain().focus().redo().run() },
];

const toolClass = (active: boolean) =>
  clsx(buttonClasses('ghost', 'sm', 'min-w-8 px-2 text-ink'), active && 'bg-surface text-red');

export default function RichTextEditorImpl({ label, value, onChange, error }: RichTextEditorProps) {
  const errorId = useId();
  const [linkOpen, setLinkOpen] = useState(false);
  const [linkValue, setLinkValue] = useState('');
  const [linkError, setLinkError] = useState<string | undefined>();

  const editor = useEditor({
    extensions: [
      StarterKit.configure({
        heading: { levels: [2, 3] },
        code: false,
        codeBlock: false,
        link: { openOnClick: false, autolink: true, protocols: ['http', 'https', 'mailto'] },
      }),
    ],
    content: value,
    immediatelyRender: true,
    // Re-render on every change so the toolbar's pressed states stay current.
    shouldRerenderOnTransaction: true,
    editorProps: {
      attributes: {
        'aria-label': label,
        class:
          'prose min-h-48 rounded-b-md border border-line bg-field px-3 py-2 text-ink focus:outline-none focus-visible:border-blue',
      },
    },
    onUpdate: ({ editor: e }) => onChange(e.isEmpty ? '' : e.getHTML()),
  });

  // Keep the error wired to the editable area for screen readers.
  useEffect(() => {
    const dom = editor?.view.dom;
    if (!dom) {
      return;
    }
    if (error) {
      dom.setAttribute('aria-describedby', errorId);
      dom.setAttribute('aria-invalid', 'true');
    } else {
      dom.removeAttribute('aria-describedby');
      dom.removeAttribute('aria-invalid');
    }
  }, [editor, error, errorId]);

  if (!editor) {
    return null;
  }

  function openLink() {
    setLinkValue(editor?.getAttributes('link').href ?? '');
    setLinkError(undefined);
    setLinkOpen(true);
  }

  function applyLink(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const href = normalizeLink(linkValue);
    if (!href) {
      setLinkError('Enter a web address (https://…) or an email link (mailto:…).');
      return;
    }
    editor?.chain().focus().extendMarkRange('link').setLink({ href }).run();
    setLinkOpen(false);
  }

  return (
    <div className="flex flex-col gap-1.5">
      <span className="text-sm font-semibold">{label}</span>
      <div
        role="toolbar"
        aria-label={`${label} formatting`}
        className="flex flex-wrap gap-1 rounded-t-md border border-b-0 border-line bg-surface p-1"
      >
        {tools.map((tool) => {
          const active = tool.isActive?.(editor) ?? false;
          return (
            <button
              key={tool.label}
              type="button"
              aria-label={tool.label}
              aria-pressed={tool.isActive ? active : undefined}
              title={tool.label}
              onClick={() => tool.run(editor)}
              className={toolClass(active)}
            >
              {tool.text}
            </button>
          );
        })}
        <button
          type="button"
          aria-label="Link"
          aria-pressed={editor.isActive('link')}
          title="Link"
          onClick={openLink}
          className={toolClass(editor.isActive('link'))}
        >
          🔗
        </button>
        <button
          type="button"
          aria-label="Remove link"
          title="Remove link"
          disabled={!editor.isActive('link')}
          onClick={() => editor.chain().focus().unsetLink().run()}
          className={toolClass(false)}
        >
          ⛓
        </button>
      </div>
      <EditorContent editor={editor} />
      {error && (
        <p id={errorId} className="text-sm text-red">
          {error}
        </p>
      )}
      <Modal open={linkOpen} onClose={() => setLinkOpen(false)} title="Add a link">
        <form onSubmit={applyLink} className="flex flex-col gap-4" noValidate>
          <Field label="Link address" error={linkError} help="For example https://thefa.com or mailto:someone@example.com">
            <Input value={linkValue} onChange={(e) => setLinkValue(e.target.value)} autoFocus />
          </Field>
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => setLinkOpen(false)}>
              Cancel
            </Button>
            <Button type="submit">Add link</Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `yarn eslint --fix client/components/edit client/lib client/test && yarn vitest run client/components/edit && yarn test:client`
Expected: PASS. 7 editor tests pass, and the suite is green with no unhandled errors.

If the "empty document" test fails because Tiptap 3's `clearContent` signature differs, check `node_modules/@tiptap/core` for the current `clearContent(emitUpdate?)` or `clearContent({ emitUpdate })` form and use it. Don't weaken the assertion.

If an emitted HTML string differs only in harmless structure (for example `<li><p>` nesting), match Tiptap's actual output and record a ruling.

- [ ] **Step 6: Commit**

```bash
yarn typecheck
git add package.json yarn.lock client/components/edit client/lib/links.ts client/test
git commit -q -m "Add the lazily loaded Tiptap rich-text editor with a safe link dialog" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 4: API write calls

**Files:**
- Modify: `client/api/news.ts`, `whatson.ts`, `pages.ts`, `teams.ts`, `documents.ts`, `programmes.ts`, `sponsors.ts`, `home.ts`, `gallery.ts`
- Test: `client/api/writes.test.ts`

**Interfaces:**
- Consumes: `formData`, `ImageValue` (Task 1); `apiFetch`.
- Produces:
  - **News:** `NewsInput { title; content; image: ImageValue }`; `createNews(i): Promise<NewsArticle>`, `updateNews(id, i): Promise<NewsArticle>`, `deleteNews(id)`
  - **What's On:** `EventInput { title; content; dateOfEvent /* YYYY-MM-DD */; image }`; `createEvent`, `updateEvent`, `deleteEvent`
  - **Info:** `setInfo(content: string)`
  - **Teams:** `TeamInput { name; ages: number; description; league; division; leagueTable; fixtures; coach; physio; isActive; isYouth; image }`; `createTeam(i): Promise<TeamSummary>`, `updateTeam(id, i)`, `deleteTeam(id)`
  - **Documents:** `createDocument({ name, file })`, `deleteDocument(id)`
  - **Programmes and seasons:** `createProgramme({ name, date, seasonId: number | null, file })`, `deleteProgramme(id)`; `createSeason(name)`, `renameSeason(id, name)`, `deleteSeason(id)`
  - **Sponsors:** `createSponsor({ name, website, purpose, team, image: File })`, `deleteSponsor(id)`
  - **Affiliations:** `createAffiliation({ name, website, image: File })`, `deleteAffiliation(id)`
  - **Gallery:** `createPhoto({ caption, image: File })`, `deletePhoto(id)`

- [ ] **Step 1: Write the failing tests**

`client/api/writes.test.ts`:

```ts
import { describe, expect, it } from 'vitest';

import { emptyImage } from '../lib/images';
import { mockFetch } from '../test/mockFetch';
import { deleteDocument, createDocument } from './documents';
import { createPhoto } from './gallery';
import { createAffiliation } from './home';
import { createNews, deleteNews, updateNews } from './news';
import { setInfo } from './pages';
import { createProgramme, createSeason, deleteSeason, renameSeason } from './programmes';
import { createSponsor } from './sponsors';
import { createTeam, updateTeam } from './teams';
import { updateEvent } from './whatson';

const png = () => new File(['x'], 'a.png', { type: 'image/png' });
const pdf = () => new File(['x'], 'a.pdf', { type: 'application/pdf' });

function body(fetchMock: ReturnType<typeof mockFetch>, i = 0) {
  return fetchMock.mock.calls[i][1]?.body as FormData;
}

describe('write calls', () => {
  it('creates news with title/content/image and no removeImage', async () => {
    const fetchMock = mockFetch({ '/api/v1/news': { status: 201, body: { id: 5 } } });
    const file = png();
    await expect(createNews({ title: 'T', content: '<p>c</p>', image: { file, remove: false } })).resolves.toEqual({ id: 5 });
    const fd = body(fetchMock);
    expect(fetchMock.mock.calls[0][1]?.method).toBe('POST');
    expect([...fd.keys()].sort()).toEqual(['content', 'image', 'title']);
    expect(fd.get('image')).toBe(file);
  });

  it('updates news: no image fields when the image is untouched, removeImage when ticked', async () => {
    const fetchMock = mockFetch({ '/api/v1/news/5': { body: { id: 5 } } });
    await updateNews(5, { title: 'T', content: '', image: emptyImage });
    expect(fetchMock.mock.calls[0][1]?.method).toBe('PATCH');
    expect([...body(fetchMock).keys()].sort()).toEqual(['content', 'title']);
    await updateNews(5, { title: 'T', content: '', image: { file: null, remove: true } });
    expect(body(fetchMock, 1).get('removeImage')).toBe('true');
  });

  it('deletes with DELETE', async () => {
    const fetchMock = mockFetch({ '/api/v1/news/5': { status: 204 } });
    await deleteNews(5);
    expect(fetchMock.mock.calls[0][1]?.method).toBe('DELETE');
  });

  it('updates an event with its date', async () => {
    const fetchMock = mockFetch({ '/api/v1/whatson/3': { body: { id: 3 } } });
    await updateEvent(3, { title: 'Quiz', content: '', dateOfEvent: '2026-11-07', image: emptyImage });
    expect(body(fetchMock).get('dateOfEvent')).toBe('2026-11-07');
  });

  it('saves info as JSON with PUT', async () => {
    const fetchMock = mockFetch({ '/api/v1/info': { status: 204 } });
    await setInfo('<p>x</p>');
    expect(fetchMock.mock.calls[0][1]?.method).toBe('PUT');
    expect(JSON.parse(String(fetchMock.mock.calls[0][1]?.body))).toEqual({ content: '<p>x</p>' });
  });

  it('always sends team booleans and every text field', async () => {
    const fetchMock = mockFetch({
      '/api/v1/teams': { status: 201, body: { id: 9 } },
      '/api/v1/teams/9': { body: { id: 9 } },
    });
    const input = {
      name: 'Vets',
      ages: 19,
      description: '',
      league: '',
      division: '',
      leagueTable: '',
      fixtures: '',
      coach: '',
      physio: '',
      isActive: false,
      isYouth: false,
      image: emptyImage,
    };
    await createTeam(input);
    await updateTeam(9, input);
    for (const i of [0, 1]) {
      const fd = body(fetchMock, i);
      expect(fd.get('isActive')).toBe('false');
      expect(fd.get('isYouth')).toBe('false');
      expect(fd.get('ages')).toBe('19');
      expect(fd.get('coach')).toBe('');
      expect(fd.has('image')).toBe(false);
    }
    expect(fetchMock.mock.calls[1][1]?.method).toBe('PATCH');
  });

  it('creates a document and a programme (seasonId only when chosen)', async () => {
    const fetchMock = mockFetch({
      '/api/v1/documents': { status: 201, body: { id: 1 } },
      '/api/v1/programmes': { status: 201, body: { id: 2 } },
      '/api/v1/documents/1': { status: 204 },
    });
    await createDocument({ name: 'Rules', file: pdf() });
    expect(body(fetchMock).get('name')).toBe('Rules');
    await createProgramme({ name: 'vs X', date: '2026-10-03', seasonId: null, file: pdf() });
    expect(body(fetchMock, 1).has('seasonId')).toBe(false);
    await createProgramme({ name: 'vs Y', date: '2026-10-10', seasonId: 2, file: pdf() });
    expect(body(fetchMock, 2).get('seasonId')).toBe('2');
    await deleteDocument(1);
    expect(fetchMock.mock.calls[3][1]?.method).toBe('DELETE');
  });

  it('manages seasons with JSON', async () => {
    const fetchMock = mockFetch({
      '/api/v1/seasons': { status: 201, body: { id: 3, name: '2027-28' } },
      '/api/v1/seasons/3': { body: { id: 3, name: '2027/28' } },
    });
    await createSeason('2027-28');
    await renameSeason(3, '2027/28');
    await deleteSeason(3);
    const calls = fetchMock.mock.calls.map(([, init]) => [init?.method, init?.body]);
    expect(calls).toEqual([
      ['POST', JSON.stringify({ name: '2027-28' })],
      ['PATCH', JSON.stringify({ name: '2027/28' })],
      ['DELETE', undefined],
    ]);
  });

  it('creates a sponsor, an affiliation and a photo', async () => {
    const fetchMock = mockFetch({
      '/api/v1/sponsors': { status: 201, body: { id: 1 } },
      '/api/v1/affiliations': { status: 201, body: { id: 1 } },
      '/api/v1/gallery': { status: 201, body: { id: 1 } },
    });
    await createSponsor({ name: 'Acme', website: '', purpose: 'Kit', team: 'Y', image: png() });
    expect(body(fetchMock).get('team')).toBe('Y');
    await createAffiliation({ name: 'FA', website: 'https://thefa.com', image: png() });
    expect(body(fetchMock, 1).get('website')).toBe('https://thefa.com');
    await createPhoto({ caption: '', image: png() });
    expect(body(fetchMock, 2).get('caption')).toBe('');
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/api/writes.test.ts`
Expected: FAIL. The imports don't exist yet (for example, "does not provide an export named 'createDocument'").

- [ ] **Step 3: Implement.** Add the calls below to each module, together with the imports they need (`formData` from `../lib/editForm`, `ImageValue` from `../lib/images`).

`client/api/news.ts`:

```ts
export interface NewsInput {
  title: string;
  content: string;
  image: ImageValue;
}

function newsForm(input: NewsInput, isUpdate: boolean): FormData {
  return formData({
    title: input.title,
    content: input.content,
    image: input.image.file,
    removeImage: isUpdate && input.image.remove ? true : undefined,
  });
}

export function createNews(input: NewsInput): Promise<NewsArticle> {
  return apiFetch<NewsArticle>('/news', { form: newsForm(input, false) });
}

export function updateNews(id: number, input: NewsInput): Promise<NewsArticle> {
  return apiFetch<NewsArticle>(`/news/${id}`, { method: 'PATCH', form: newsForm(input, true) });
}

export function deleteNews(id: number): Promise<void> {
  return apiFetch<void>(`/news/${id}`, { method: 'DELETE' });
}
```

`client/api/whatson.ts`:

```ts
export interface EventInput {
  title: string;
  content: string;
  /** YYYY-MM-DD */
  dateOfEvent: string;
  image: ImageValue;
}

function eventForm(input: EventInput, isUpdate: boolean): FormData {
  return formData({
    title: input.title,
    content: input.content,
    dateOfEvent: input.dateOfEvent,
    image: input.image.file,
    removeImage: isUpdate && input.image.remove ? true : undefined,
  });
}

export function createEvent(input: EventInput): Promise<WhatsOnEvent> {
  return apiFetch<WhatsOnEvent>('/whatson', { form: eventForm(input, false) });
}

export function updateEvent(id: number, input: EventInput): Promise<WhatsOnEvent> {
  return apiFetch<WhatsOnEvent>(`/whatson/${id}`, { method: 'PATCH', form: eventForm(input, true) });
}

export function deleteEvent(id: number): Promise<void> {
  return apiFetch<void>(`/whatson/${id}`, { method: 'DELETE' });
}
```

`client/api/pages.ts`:

```ts
export function setInfo(content: string): Promise<void> {
  return apiFetch<void>('/info', { method: 'PUT', json: { content } });
}
```

`client/api/teams.ts`:

```ts
export interface TeamInput {
  name: string;
  /** 6–18 for "Under N", 19 for "Over 18". */
  ages: number;
  description: string;
  league: string;
  division: string;
  leagueTable: string;
  fixtures: string;
  coach: string;
  physio: string;
  isActive: boolean;
  isYouth: boolean;
  image: ImageValue;
}

function teamForm(input: TeamInput, isUpdate: boolean): FormData {
  return formData({
    name: input.name,
    ages: input.ages,
    description: input.description,
    league: input.league,
    division: input.division,
    leagueTable: input.leagueTable,
    fixtures: input.fixtures,
    coach: input.coach,
    physio: input.physio,
    isActive: input.isActive,
    isYouth: input.isYouth,
    image: input.image.file,
    removeImage: isUpdate && input.image.remove ? true : undefined,
  });
}

export function createTeam(input: TeamInput): Promise<TeamSummary> {
  return apiFetch<TeamSummary>('/teams', { form: teamForm(input, false) });
}

export function updateTeam(id: number, input: TeamInput): Promise<TeamSummary> {
  return apiFetch<TeamSummary>(`/teams/${id}`, { method: 'PATCH', form: teamForm(input, true) });
}

export function deleteTeam(id: number): Promise<void> {
  return apiFetch<void>(`/teams/${id}`, { method: 'DELETE' });
}
```

`client/api/documents.ts`:

```ts
export function createDocument(input: { name: string; file: File }): Promise<ClubDocument> {
  return apiFetch<ClubDocument>('/documents', { form: formData(input) });
}

export function deleteDocument(id: number): Promise<void> {
  return apiFetch<void>(`/documents/${id}`, { method: 'DELETE' });
}
```

`client/api/programmes.ts`:

```ts
export function createProgramme(input: {
  name: string;
  /** YYYY-MM-DD */
  date: string;
  seasonId: number | null;
  file: File;
}): Promise<Programme> {
  return apiFetch<Programme>('/programmes', {
    form: formData({ name: input.name, date: input.date, seasonId: input.seasonId, file: input.file }),
  });
}

export function deleteProgramme(id: number): Promise<void> {
  return apiFetch<void>(`/programmes/${id}`, { method: 'DELETE' });
}

export function createSeason(name: string): Promise<Season> {
  return apiFetch<Season>('/seasons', { json: { name } });
}

export function renameSeason(id: number, name: string): Promise<Season> {
  return apiFetch<Season>(`/seasons/${id}`, { method: 'PATCH', json: { name } });
}

export function deleteSeason(id: number): Promise<void> {
  return apiFetch<void>(`/seasons/${id}`, { method: 'DELETE' });
}
```

`client/api/sponsors.ts`:

```ts
export function createSponsor(input: {
  name: string;
  website: string;
  purpose: string;
  /** '', 'A', 'O', 'Y' or a team id */
  team: string;
  image: File;
}): Promise<Sponsor> {
  return apiFetch<Sponsor>('/sponsors', { form: formData(input) });
}

export function deleteSponsor(id: number): Promise<void> {
  return apiFetch<void>(`/sponsors/${id}`, { method: 'DELETE' });
}
```

`client/api/home.ts`:

```ts
export function createAffiliation(input: {
  name: string;
  website: string;
  image: File;
}): Promise<Affiliation> {
  return apiFetch<Affiliation>('/affiliations', { form: formData(input) });
}

export function deleteAffiliation(id: number): Promise<void> {
  return apiFetch<void>(`/affiliations/${id}`, { method: 'DELETE' });
}
```

`client/api/gallery.ts`:

```ts
export function createPhoto(input: { caption: string; image: File }): Promise<GalleryImage> {
  return apiFetch<GalleryImage>('/gallery', { form: formData(input) });
}

export function deletePhoto(id: number): Promise<void> {
  return apiFetch<void>(`/gallery/${id}`, { method: 'DELETE' });
}
```

- [ ] **Step 4: Check the delete and create responses against the Go handlers**

Run: `grep -n 'NoContent\|StatusCreated\|StatusOK' server/internal/{news,whatson,team,document,programme,sponsor,affiliation,image,setting}/handlers.go`

For every delete that returns `200` with a body (rather than `204`), `apiFetch<void>` still resolves; the parsed body is ignored. Nothing to change, but note it in the ledger if any differ from the spec.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `yarn eslint --fix client/api && yarn vitest run client/api/writes.test.ts && yarn test:client`
Expected: PASS. 9 write tests pass, and the suite is green.

- [ ] **Step 6: Commit**

```bash
yarn typecheck
git add client/api
git commit -q -m "Add content write calls to each API module" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 5: News editing

**Files:**
- Create: `client/pages/news/NewsFormPage.tsx`
- Modify:
  - `client/components/page/ArticleView.tsx` (`editorHref` is replaced by `actions?: ReactNode`)
  - `client/pages/news/NewsListPage.tsx`, `client/pages/news/NewsArticlePage.tsx`
  - `client/pages/whatson/EventPage.tsx` (drops `editorHref`)
  - `client/App.tsx` (routes)
  - `client/pages/news/news.test.tsx` (the editor-link assertion)
- Test: `client/pages/news/newsEdit.test.tsx`

**Interfaces:**
- Consumes:
  - `createNews`, `updateNews`, `deleteNews`, `NewsInput` (Task 4);
  - `useCanEdit`, `RequireEditor`, `useSaveForm`, `ImageField`, `DeleteButton` (Task 2);
  - `RichTextEditor` (Task 3);
  - `emptyImage` (Task 1);
  - `typeInEditor` and `Location` (test helpers).
- Produces: `ArticleView` with `actions?: ReactNode` (no `editorHref`); routes `news/new` and `news/:id/edit`.

- [ ] **Step 1: Write the failing tests**

`client/pages/news/newsEdit.test.tsx`:

```tsx
import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import { typeInEditor } from '../../test/editor';
import { anonymous, editor, manager, newsArticle, publicRoutes } from '../../test/fixtures';
import { Location } from '../../test/Location';
import type { MockResponse, MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import NewsArticlePage from './NewsArticlePage';
import NewsFormPage from './NewsFormPage';
import NewsListPage from './NewsListPage';

function renderNews(route: string, overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(
    <>
      <Routes>
        <Route path="/news" element={<NewsListPage />} />
        <Route path="/news/new" element={<NewsFormPage />} />
        <Route path="/news/:id" element={<NewsArticlePage />} />
        <Route path="/news/:id/edit" element={<NewsFormPage />} />
      </Routes>
      <Location />
    </>,
    { route },
  );
  return fetchMock;
}

const saved = { ...newsArticle, id: 42, title: 'New title' };

describe('News editing controls', () => {
  it.each([[anonymous], [manager]] as [MockResponse][])('are hidden from non-editors', async (me) => {
    renderNews('/news', { '/api/v1/auth/me': me });
    await screen.findByRole('link', { name: new RegExp(newsArticle.title) });
    expect(screen.queryByRole('link', { name: 'Add article' })).toBeNull();
  });

  it('show Add on the list and Edit/Delete on an article for editors', async () => {
    renderNews(`/news/${newsArticle.id}`);
    expect(await screen.findByRole('link', { name: 'Edit' })).toHaveAttribute(
      'href',
      `/news/${newsArticle.id}/edit`,
    );
    expect(screen.getByRole('button', { name: 'Delete' })).toBeInTheDocument();
    expect(screen.queryByText(/classic site/)).toBeNull();
  });

  it('delete confirms, calls DELETE and returns to the list', async () => {
    const fetchMock = renderNews(`/news/${newsArticle.id}`, {
      [`/api/v1/news/${newsArticle.id}`]: () => ({ body: newsArticle }),
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Delete' }));
    const dialog = screen.getByRole('dialog', { name: 'Delete this article?' });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Delete' }));
    await waitFor(() => expect(screen.getByTestId('location')).toHaveTextContent(/^\/news$/));
    expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'DELETE')).toBe(true);
    expect(screen.getByRole('button', { name: 'Article deleted' })).toBeInTheDocument();
  });
});

describe('NewsFormPage', () => {
  it('shows the permission message to non-editors', async () => {
    renderNews('/news/new', { '/api/v1/auth/me': manager });
    expect(await screen.findByText("You don't have permission to edit this")).toBeInTheDocument();
  });

  it('asks for a title without sending anything', async () => {
    const fetchMock = renderNews('/news/new');
    fireEvent.click(await screen.findByRole('button', { name: 'Save article' }));
    expect(screen.getByLabelText('Title')).toHaveAccessibleDescription('Enter a title');
    expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(false);
  });

  it('creates an article and opens it', async () => {
    const fetchMock = renderNews('/news/new', { '/api/v1/news': { status: 201, body: saved } });
    fireEvent.change(await screen.findByLabelText('Title'), { target: { value: 'New title' } });
    await typeInEditor('Content', '<p>Body</p>');
    fireEvent.click(screen.getByRole('button', { name: 'Save article' }));
    await waitFor(() => expect(screen.getByTestId('location')).toHaveTextContent('/news/42'));
    const post = fetchMock.mock.calls.find(([, init]) => init?.method === 'POST');
    const fd = post?.[1]?.body as FormData;
    expect(fd.get('title')).toBe('New title');
    expect(fd.get('content')).toBe('<p>Body</p>');
    expect(await screen.findByRole('button', { name: 'Article saved' })).toBeInTheDocument();
  });

  it('loads an article for editing and can remove its image', async () => {
    const fetchMock = renderNews(`/news/${newsArticle.id}/edit`, {
      [`/api/v1/news/${newsArticle.id}`]: () => ({ body: newsArticle }),
    });
    expect(await screen.findByLabelText('Title')).toHaveValue(newsArticle.title);
    fireEvent.click(screen.getByRole('checkbox', { name: 'Remove image' }));
    fireEvent.click(screen.getByRole('button', { name: 'Save article' }));
    await waitFor(() =>
      expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'PATCH')).toBe(true),
    );
    const patch = fetchMock.mock.calls.find(([, init]) => init?.method === 'PATCH');
    const fd = patch?.[1]?.body as FormData;
    expect(fd.get('removeImage')).toBe('true');
    expect(fd.get('content')).toBe(newsArticle.content);
  });

  it('shows server field errors', async () => {
    renderNews('/news/new', {
      '/api/v1/news': {
        status: 422,
        body: { error: { code: 422, message: 'invalid', fields: { title: 'title is too long' } } },
      },
    });
    fireEvent.change(await screen.findByLabelText('Title'), { target: { value: 'x' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save article' }));
    await waitFor(() =>
      expect(screen.getByLabelText('Title')).toHaveAccessibleDescription('title is too long'),
    );
  });
});
```

In `client/pages/news/news.test.tsx`, the list test `'shows the empty state and the editor link'` expects the classic-site link. Change its last assertion to:

```tsx
    expect(await screen.findByRole('link', { name: 'Add article' })).toHaveAttribute(
      'href',
      '/news/new',
    );
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/news`
Expected: FAIL. `./NewsFormPage` can't be resolved.

- [ ] **Step 3: Implement**

`client/components/page/ArticleView.tsx`:
- remove the `EditorLink` import, and the `editorHref` prop from `ArticleViewProps` and the destructuring;
- add `actions?: ReactNode;` to the props, with the doc comment `/** Editor controls (Edit/Delete) for the page header. */`;
- change `actions={<EditorLink legacyHref={editorHref} />}` to `actions={actions}`.

`client/pages/whatson/EventPage.tsx`: delete the line `editorHref={`/whatson/${e.id}`}`. Task 6 adds actions.

`client/pages/news/NewsListPage.tsx`: replace the `EditorLink` import with `import { useCanEdit } from '../../components/edit/useCanEdit';` and `import { ButtonLink } from '../../components/ui/ButtonLink';`, add `const { canEdit } = useCanEdit();`, and set the header to:

```tsx
      <PageHeader
        title="News"
        actions={canEdit && <ButtonLink to="/news/new">Add article</ButtonLink>}
      />
```

`client/pages/news/NewsArticlePage.tsx`:
- add the imports `useNavigate` (from `react-router`), `deleteNews` (from `../../api/news`), `queryKeys` (from `../../api/queries`), `DeleteButton`, `useCanEdit` and `ButtonLink`;
- in the component, add `const { canEdit } = useCanEdit();` and `const navigate = useNavigate();`;
- remove `editorHref={…}` from `ArticleView` and pass:

```tsx
          actions={
            canEdit && (
              <>
                <ButtonLink to={`/news/${a.id}/edit`} variant="secondary" size="sm">
                  Edit
                </ButtonLink>
                <DeleteButton
                  confirmTitle="Delete this article?"
                  confirmMessage="This can't be undone."
                  onDelete={() => deleteNews(a.id)}
                  invalidate={[queryKeys.news, queryKeys.home]}
                  successMessage="Article deleted"
                  after={() => navigate('/news')}
                />
              </>
            )
          }
```

`client/pages/news/NewsFormPage.tsx`:

```tsx
import { useState, type FormEvent } from 'react';
import { useNavigate, useParams } from 'react-router';

import { createNews, updateNews, useNewsArticle, type NewsArticle } from '../../api/news';
import { queryKeys } from '../../api/queries';
import { ImageField } from '../../components/edit/ImageField';
import { RequireEditor } from '../../components/edit/RequireEditor';
import { RichTextEditor } from '../../components/edit/RichTextEditor';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Input } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { PageHeader } from '../../components/ui/PageHeader';
import { useToast } from '../../components/ui/toast/useToast';
import { parseId } from '../../lib/ids';
import { emptyImage, type ImageValue } from '../../lib/images';
import { isNotFound } from '../../lib/notFound';
import NotFoundPage from '../NotFoundPage';

function NewsForm({ article }: { article?: NewsArticle }) {
  const heading = article ? 'Edit article' : 'New article';
  usePageTitle(heading);
  const navigate = useNavigate();
  const toast = useToast();
  const [title, setTitle] = useState(article?.title ?? '');
  const [content, setContent] = useState(article?.content ?? '');
  const [image, setImage] = useState<ImageValue>(emptyImage);
  const [missing, setMissing] = useState<string | undefined>();

  const save = useSaveForm({
    submit: () =>
      article
        ? updateNews(article.id, { title, content, image })
        : createNews({ title, content, image }),
    invalidate: [queryKeys.news, queryKeys.home],
    onSaved: (a) => {
      toast.show({ tone: 'success', message: 'Article saved' });
      navigate(`/news/${a.id}`);
    },
  });

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!title.trim()) {
      setMissing('Enter a title');
      return;
    }
    setMissing(undefined);
    void save.run();
  }

  return (
    <>
      <PageHeader title={heading} />
      <form onSubmit={onSubmit} noValidate className="flex max-w-3xl flex-col gap-5">
        {save.formError && <Alert tone="error">{save.formError}</Alert>}
        <Field label="Title" error={missing ?? save.fieldErrors.title}>
          <Input value={title} onChange={(e) => setTitle(e.target.value)} />
        </Field>
        <ImageField
          label="Image"
          currentUrl={article?.imageUrl}
          allowRemove={Boolean(article)}
          value={image}
          onChange={setImage}
          error={save.fieldErrors.file ?? save.fieldErrors.image}
        />
        <RichTextEditor
          label="Content"
          value={content}
          onChange={setContent}
          error={save.fieldErrors.content}
        />
        <div className="flex gap-2">
          <Button type="submit" loading={save.busy}>
            Save article
          </Button>
          <Button variant="secondary" onClick={() => navigate(-1)} disabled={save.busy}>
            Cancel
          </Button>
        </div>
      </form>
    </>
  );
}

function EditNews({ id }: { id: number | null }) {
  const article = useNewsArticle(id);
  if (id === null || isNotFound(article.error)) {
    return <NotFoundPage />;
  }
  return <QueryState query={article}>{(a) => <NewsForm article={a} />}</QueryState>;
}

export default function NewsFormPage() {
  const { id } = useParams();
  return (
    <RequireEditor>{id === undefined ? <NewsForm /> : <EditNews id={parseId(id)} />}</RequireEditor>
  );
}
```

`client/App.tsx`: add `const NewsFormPage = lazy(() => import('./pages/news/NewsFormPage'));`, and add the routes **before** `news/:id`:

```tsx
        <Route path="news/new" element={<NewsFormPage />} />
        <Route path="news/:id/edit" element={<NewsFormPage />} />
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/pages client/components/page client/App.tsx && yarn vitest run client/pages/news client/pages/whatson && yarn test:client`
Expected: PASS. 9 new News-editing tests pass, and the suite is green.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add -A client
git commit -q -m "Edit news in the app: add, edit and delete articles" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 6: What's On editing

**Files:**
- Create: `client/pages/whatson/EventFormPage.tsx`
- Modify: `client/pages/whatson/WhatsOnPage.tsx`, `client/pages/whatson/EventPage.tsx`, `client/App.tsx`
- Test: `client/pages/whatson/eventEdit.test.tsx`

**Interfaces:**
- Consumes: `createEvent`, `updateEvent`, `deleteEvent` (Task 4); `toDateInput` (Task 1); the toolkit and `RichTextEditor`.
- Produces: routes `whatson/new` and `whatson/:id/edit`.

- [ ] **Step 1: Write the failing tests**

`client/pages/whatson/eventEdit.test.tsx`:

```tsx
import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import { typeInEditor } from '../../test/editor';
import { editor, event, manager, publicRoutes } from '../../test/fixtures';
import { Location } from '../../test/Location';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import EventFormPage from './EventFormPage';
import EventPage from './EventPage';
import WhatsOnPage from './WhatsOnPage';

function renderEvents(route: string, overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(
    <>
      <Routes>
        <Route path="/whatson" element={<WhatsOnPage />} />
        <Route path="/whatson/new" element={<EventFormPage />} />
        <Route path="/whatson/:id" element={<EventPage />} />
        <Route path="/whatson/:id/edit" element={<EventFormPage />} />
      </Routes>
      <Location />
    </>,
    { route },
  );
  return fetchMock;
}

describe("What's On editing", () => {
  it('shows Add event to editors only', async () => {
    renderEvents('/whatson');
    expect(await screen.findByRole('link', { name: 'Add event' })).toHaveAttribute(
      'href',
      '/whatson/new',
    );
  });

  it('hides the controls from a Manager', async () => {
    renderEvents(`/whatson/${event.id}`, { '/api/v1/auth/me': manager });
    await screen.findByRole('heading', { level: 1, name: event.title });
    expect(screen.queryByRole('link', { name: 'Edit' })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Delete' })).toBeNull();
  });

  it('requires a title and a date', async () => {
    const fetchMock = renderEvents('/whatson/new');
    fireEvent.click(await screen.findByRole('button', { name: 'Save event' }));
    expect(screen.getByLabelText('Title')).toHaveAccessibleDescription('Enter a title');
    expect(screen.getByLabelText('Date of event')).toHaveAccessibleDescription(
      'Choose the date of the event',
    );
    expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(false);
  });

  it('creates an event with its date and opens it', async () => {
    const fetchMock = renderEvents('/whatson/new', {
      '/api/v1/whatson': { status: 201, body: { ...event, id: 77 } },
    });
    fireEvent.change(await screen.findByLabelText('Title'), { target: { value: 'Quiz night' } });
    fireEvent.change(screen.getByLabelText('Date of event'), { target: { value: '2026-11-07' } });
    await typeInEditor('Content', '<p>Teams of four</p>');
    fireEvent.click(screen.getByRole('button', { name: 'Save event' }));
    await waitFor(() => expect(screen.getByTestId('location')).toHaveTextContent('/whatson/77'));
    const fd = fetchMock.mock.calls.find(([, i]) => i?.method === 'POST')?.[1]?.body as FormData;
    expect(fd.get('dateOfEvent')).toBe('2026-11-07');
    expect(fd.get('title')).toBe('Quiz night');
  });

  it('loads the existing date when editing', async () => {
    renderEvents(`/whatson/${event.id}/edit`);
    expect(await screen.findByLabelText('Date of event')).toHaveValue('2026-10-16');
  });

  it('deletes an event after confirming', async () => {
    const fetchMock = renderEvents(`/whatson/${event.id}`, {
      [`/api/v1/whatson/${event.id}`]: () => ({ body: event }),
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Delete' }));
    fireEvent.click(
      within(screen.getByRole('dialog', { name: 'Delete this event?' })).getByRole('button', {
        name: 'Delete',
      }),
    );
    await waitFor(() => expect(screen.getByTestId('location')).toHaveTextContent(/^\/whatson$/));
    expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true);
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/whatson/eventEdit.test.tsx`
Expected: FAIL. `./EventFormPage` can't be resolved.

- [ ] **Step 3: Implement**

`client/pages/whatson/WhatsOnPage.tsx`:
- replace the `EditorLink` import with `useCanEdit` and `ButtonLink`;
- add `const { canEdit } = useCanEdit();`;
- set the header's actions to `actions={canEdit && <ButtonLink to="/whatson/new">Add event</ButtonLink>}`.

`client/pages/whatson/EventPage.tsx`:
- add the imports `useNavigate`, `deleteEvent`, `DeleteButton`, `useCanEdit` and `ButtonLink`;
- add `const { canEdit } = useCanEdit();` and `const navigate = useNavigate();`;
- pass this to `ArticleView`:

```tsx
          actions={
            canEdit && (
              <>
                <ButtonLink to={`/whatson/${e.id}/edit`} variant="secondary" size="sm">
                  Edit
                </ButtonLink>
                <DeleteButton
                  confirmTitle="Delete this event?"
                  confirmMessage="This can't be undone."
                  onDelete={() => deleteEvent(e.id)}
                  invalidate={[['whatson'], ['home']]}
                  successMessage="Event deleted"
                  after={() => navigate('/whatson')}
                />
              </>
            )
          }
```

`client/pages/whatson/EventFormPage.tsx`:

```tsx
import { useState, type FormEvent } from 'react';
import { useNavigate, useParams } from 'react-router';

import { createEvent, updateEvent, useWhatsOnEvent, type WhatsOnEvent } from '../../api/whatson';
import { ImageField } from '../../components/edit/ImageField';
import { RequireEditor } from '../../components/edit/RequireEditor';
import { RichTextEditor } from '../../components/edit/RichTextEditor';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Input } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { PageHeader } from '../../components/ui/PageHeader';
import { useToast } from '../../components/ui/toast/useToast';
import { toDateInput } from '../../lib/editForm';
import { parseId } from '../../lib/ids';
import { emptyImage, type ImageValue } from '../../lib/images';
import { isNotFound } from '../../lib/notFound';
import NotFoundPage from '../NotFoundPage';

function EventForm({ event }: { event?: WhatsOnEvent }) {
  const heading = event ? 'Edit event' : 'New event';
  usePageTitle(heading);
  const navigate = useNavigate();
  const toast = useToast();
  const [title, setTitle] = useState(event?.title ?? '');
  const [date, setDate] = useState(event ? toDateInput(event.dateOfEvent) : '');
  const [content, setContent] = useState(event?.content ?? '');
  const [image, setImage] = useState<ImageValue>(emptyImage);
  const [missing, setMissing] = useState<{ title?: string; date?: string }>({});

  const save = useSaveForm({
    submit: () => {
      const input = { title, content, dateOfEvent: date, image };
      return event ? updateEvent(event.id, input) : createEvent(input);
    },
    invalidate: [['whatson'], ['home']],
    onSaved: (e) => {
      toast.show({ tone: 'success', message: 'Event saved' });
      navigate(`/whatson/${e.id}`);
    },
  });

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const next = {
      title: title.trim() ? undefined : 'Enter a title',
      date: date ? undefined : 'Choose the date of the event',
    };
    setMissing(next);
    if (next.title || next.date) {
      return;
    }
    void save.run();
  }

  return (
    <>
      <PageHeader title={heading} />
      <form onSubmit={onSubmit} noValidate className="flex max-w-3xl flex-col gap-5">
        {save.formError && <Alert tone="error">{save.formError}</Alert>}
        <Field label="Title" error={missing.title ?? save.fieldErrors.title}>
          <Input value={title} onChange={(e) => setTitle(e.target.value)} />
        </Field>
        <Field label="Date of event" error={missing.date ?? save.fieldErrors.dateOfEvent}>
          <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
        </Field>
        <ImageField
          label="Image"
          currentUrl={event?.imageUrl}
          allowRemove={Boolean(event)}
          value={image}
          onChange={setImage}
          error={save.fieldErrors.file ?? save.fieldErrors.image}
        />
        <RichTextEditor
          label="Content"
          value={content}
          onChange={setContent}
          error={save.fieldErrors.content}
        />
        <div className="flex gap-2">
          <Button type="submit" loading={save.busy}>
            Save event
          </Button>
          <Button variant="secondary" onClick={() => navigate(-1)} disabled={save.busy}>
            Cancel
          </Button>
        </div>
      </form>
    </>
  );
}

function EditEvent({ id }: { id: number | null }) {
  const event = useWhatsOnEvent(id);
  if (id === null || isNotFound(event.error)) {
    return <NotFoundPage />;
  }
  return <QueryState query={event}>{(e) => <EventForm event={e} />}</QueryState>;
}

export default function EventFormPage() {
  const { id } = useParams();
  return (
    <RequireEditor>
      {id === undefined ? <EventForm /> : <EditEvent id={parseId(id)} />}
    </RequireEditor>
  );
}
```

`client/App.tsx`: add `const EventFormPage = lazy(() => import('./pages/whatson/EventFormPage'));` and, before `whatson/:id`:

```tsx
        <Route path="whatson/new" element={<EventFormPage />} />
        <Route path="whatson/:id/edit" element={<EventFormPage />} />
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/pages/whatson client/App.tsx && yarn vitest run client/pages/whatson && yarn test:client`
Expected: PASS. 6 new tests pass, and the suite is green.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add -A client
git commit -q -m "Edit What's On events in the app" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 7: Editing the club information

**Files:**
- Create: `client/pages/info/InfoEditPage.tsx`
- Modify: `client/pages/info/InfoPage.tsx`, `client/components/layout/AccountControl.tsx`, `client/App.tsx`, `client/pages/info/info.test.tsx`
- Test: `client/pages/info/infoEdit.test.tsx`

**Interfaces:**
- Consumes: `setInfo` (Task 4); `useInfo`; `InfoFallback` (4a); the toolkit and `RichTextEditor`.
- Produces: the route `info/edit`. The account menu's "Edit info" becomes `{ label: 'Edit info', to: '/info/edit' }`.

- [ ] **Step 1: Write the failing tests**

`client/pages/info/infoEdit.test.tsx`:

```tsx
import { fireEvent, screen, waitFor } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import { editorFor, typeInEditor } from '../../test/editor';
import { editor, info, manager, publicRoutes } from '../../test/fixtures';
import { Location } from '../../test/Location';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import InfoEditPage from './InfoEditPage';
import InfoPage from './InfoPage';

function renderInfo(route: string, overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(
    <>
      <Routes>
        <Route path="/info" element={<InfoPage />} />
        <Route path="/info/edit" element={<InfoEditPage />} />
      </Routes>
      <Location />
    </>,
    { route },
  );
  return fetchMock;
}

describe('Info editing', () => {
  it('links editors to the in-app editor', async () => {
    renderInfo('/info');
    expect(await screen.findByRole('link', { name: 'Edit' })).toHaveAttribute('href', '/info/edit');
  });

  it('hides Edit from a Manager', async () => {
    renderInfo('/info', { '/api/v1/auth/me': manager });
    await screen.findByRole('heading', { level: 2, name: 'Welcome' });
    expect(screen.queryByRole('link', { name: 'Edit' })).toBeNull();
  });

  it('starts from the stored content and saves it with PUT', async () => {
    const fetchMock = renderInfo('/info/edit', { '/api/v1/info': () => ({ body: info }) });
    expect((await editorFor('Club information')).getHTML()).toContain('Welcome');
    await typeInEditor('Club information', '<p>Updated</p>');
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    await waitFor(() => expect(screen.getByTestId('location')).toHaveTextContent(/^\/info$/));
    const put = fetchMock.mock.calls.find(([, i]) => i?.method === 'PUT');
    expect(JSON.parse(String(put?.[1]?.body))).toEqual({ content: '<p>Updated</p>' });
    expect(screen.getByRole('button', { name: 'Club information saved' })).toBeInTheDocument();
  });

  it('starts from the club history when nothing is stored', async () => {
    renderInfo('/info/edit', { '/api/v1/info': { body: { content: '' } } });
    const html = (await editorFor('Club information')).getHTML();
    expect(html).toContain('Club history');
    expect(html).toContain('AWRE Football Club');
  });
});
```

In `client/pages/info/info.test.tsx`, change the test `'links editors to the classic info editor'` to expect `screen.findByRole('link', { name: 'Edit' })` with href `/info/edit`, and rename it to `'links editors to the in-app editor'`.

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/info`
Expected: FAIL. `./InfoEditPage` can't be resolved.

- [ ] **Step 3: Implement**

`client/pages/info/InfoPage.tsx`:
- replace the `EditorLink` import with `useCanEdit` and `ButtonLink`;
- add `const { canEdit } = useCanEdit();`;
- set the header's actions to:

```tsx
        actions={
          canEdit && (
            <ButtonLink to="/info/edit" variant="secondary" size="sm">
              Edit
            </ButtonLink>
          )
        }
```

`client/components/layout/AccountControl.tsx`: change `{ label: 'Edit info', href: '/info/edit' }` to `{ label: 'Edit info', to: '/info/edit' }`.

`client/pages/info/InfoEditPage.tsx`:

```tsx
import { useState } from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import { useNavigate } from 'react-router';

import { setInfo, useInfo } from '../../api/pages';
import { queryKeys } from '../../api/queries';
import { RequireEditor } from '../../components/edit/RequireEditor';
import { RichTextEditor } from '../../components/edit/RichTextEditor';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { PageHeader } from '../../components/ui/PageHeader';
import { useToast } from '../../components/ui/toast/useToast';
import { InfoFallback } from './fallback';

function InfoForm({ initial }: { initial: string }) {
  const navigate = useNavigate();
  const toast = useToast();
  const [content, setContent] = useState(initial);
  const save = useSaveForm({
    submit: () => setInfo(content),
    invalidate: [queryKeys.info],
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Club information saved' });
      navigate('/info');
    },
  });
  return (
    <form
      noValidate
      className="flex max-w-3xl flex-col gap-5"
      onSubmit={(e) => {
        e.preventDefault();
        void save.run();
      }}
    >
      {save.formError && <Alert tone="error">{save.formError}</Alert>}
      <RichTextEditor
        label="Club information"
        value={initial}
        onChange={setContent}
        error={save.fieldErrors.content}
      />
      <div className="flex gap-2">
        <Button type="submit" loading={save.busy}>
          Save
        </Button>
        <Button variant="secondary" onClick={() => navigate('/info')} disabled={save.busy}>
          Cancel
        </Button>
      </div>
    </form>
  );
}

export default function InfoEditPage() {
  usePageTitle('Edit information');
  const info = useInfo();
  return (
    <RequireEditor>
      <PageHeader title="Edit information" />
      <QueryState query={info}>
        {(data) => (
          <InfoForm
            // With nothing saved yet, start from the wording the page already shows.
            initial={data.content.trim() ? data.content : renderToStaticMarkup(<InfoFallback />)}
          />
        )}
      </QueryState>
    </RequireEditor>
  );
}
```

`client/App.tsx`: add `const InfoEditPage = lazy(() => import('./pages/info/InfoEditPage'));` and `<Route path="info/edit" element={<InfoEditPage />} />`.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/pages/info client/components/layout client/App.tsx && yarn vitest run client/pages/info client/components/layout && yarn test:client`
Expected: PASS. 4 new tests pass, and the suite is green.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add -A client
git commit -q -m "Edit the club information in the app" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 8: Team editing

**Files:**
- Create: `client/pages/teams/TeamFormPage.tsx`, `client/lib/ageGroups.ts`
- Modify: `client/pages/teams/TeamsPage.tsx`, `client/pages/teams/TeamPage.tsx`, `client/App.tsx`, `client/pages/teams/teams.test.tsx`
- Test: `client/pages/teams/teamEdit.test.tsx`

**Interfaces:**
- Consumes: `createTeam`, `updateTeam`, `deleteTeam`, `TeamInput` (Task 4); `useTeam`; the toolkit.
- Produces:
  - `AGE_GROUPS: { value: number; label: string }[]` (6–18 "Under N", 19 "Over 18");
  - routes `teams/new` and `team/:id/edit`.

- [ ] **Step 1: Write the failing tests**

`client/pages/teams/teamEdit.test.tsx`:

```tsx
import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import { editor, manager, publicRoutes, team, teamDetail } from '../../test/fixtures';
import { Location } from '../../test/Location';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import TeamFormPage from './TeamFormPage';
import TeamPage from './TeamPage';
import TeamsPage from './TeamsPage';

function renderTeams(route: string, overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(
    <>
      <Routes>
        <Route path="/teams" element={<TeamsPage />} />
        <Route path="/teams/new" element={<TeamFormPage />} />
        <Route path="/team/:id" element={<TeamPage />} />
        <Route path="/team/:id/edit" element={<TeamFormPage />} />
      </Routes>
      <Location />
    </>,
    { route },
  );
  return fetchMock;
}

describe('Team editing', () => {
  it('shows Add team to editors and hides it from a Manager', async () => {
    renderTeams('/teams');
    expect(await screen.findByRole('link', { name: 'Add team' })).toHaveAttribute(
      'href',
      '/teams/new',
    );
  });

  it('hides Edit/Delete on a team from a Manager', async () => {
    renderTeams(`/team/${team.id}`, { '/api/v1/auth/me': manager });
    await screen.findByRole('heading', { level: 1, name: team.name });
    expect(screen.queryByRole('link', { name: 'Edit' })).toBeNull();
  });

  it('asks for a name and age group; new teams default to active', async () => {
    const fetchMock = renderTeams('/teams/new');
    fireEvent.click(await screen.findByRole('button', { name: 'Save team' }));
    expect(screen.getByLabelText('Name')).toHaveAccessibleDescription('Enter a name');
    expect(screen.getByLabelText('Age group')).toHaveAccessibleDescription('Choose the age group');
    expect(screen.getByRole('checkbox', { name: 'Active team' })).toBeChecked();
    expect(screen.getByRole('checkbox', { name: 'Youth team' })).not.toBeChecked();
    expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'POST')).toBe(false);
  });

  it('creates a team with every field and opens it', async () => {
    const fetchMock = renderTeams('/teams/new', {
      '/api/v1/teams': { status: 201, body: { ...team, id: 30, name: 'Under 10s' } },
    });
    fireEvent.change(await screen.findByLabelText('Name'), { target: { value: 'Under 10s' } });
    fireEvent.change(screen.getByLabelText('Age group'), { target: { value: '10' } });
    fireEvent.change(screen.getByLabelText('Coach'), { target: { value: 'Sam' } });
    fireEvent.click(screen.getByRole('checkbox', { name: 'Youth team' }));
    fireEvent.click(screen.getByRole('button', { name: 'Save team' }));
    await waitFor(() => expect(screen.getByTestId('location')).toHaveTextContent('/team/30'));
    const fd = fetchMock.mock.calls.find(([, i]) => i?.method === 'POST')?.[1]?.body as FormData;
    expect(fd.get('name')).toBe('Under 10s');
    expect(fd.get('ages')).toBe('10');
    expect(fd.get('coach')).toBe('Sam');
    expect(fd.get('isYouth')).toBe('true');
    expect(fd.get('isActive')).toBe('true');
    expect(fd.get('physio')).toBe('');
  });

  it('loads a team for editing and sends an unticked Active as false', async () => {
    const fetchMock = renderTeams(`/team/${team.id}/edit`, {
      [`/api/v1/teams/${team.id}`]: () => ({ body: teamDetail }),
    });
    expect(await screen.findByLabelText('Name')).toHaveValue(team.name);
    expect(screen.getByLabelText('League table URL')).toHaveValue(teamDetail.team.leagueTableUrl);
    fireEvent.click(screen.getByRole('checkbox', { name: 'Active team' }));
    fireEvent.click(screen.getByRole('button', { name: 'Save team' }));
    await waitFor(() =>
      expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'PATCH')).toBe(true),
    );
    const fd = fetchMock.mock.calls.find(([, i]) => i?.method === 'PATCH')?.[1]?.body as FormData;
    expect(fd.get('isActive')).toBe('false');
    expect(fd.get('leagueTable')).toBe(teamDetail.team.leagueTableUrl);
    expect(fd.has('image')).toBe(false);
  });

  it('warns that deleting unlinks players, sponsors and managers', async () => {
    const fetchMock = renderTeams(`/team/${team.id}`, {
      [`/api/v1/teams/${team.id}`]: () => ({ body: teamDetail }),
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Delete' }));
    const dialog = screen.getByRole('dialog', { name: 'Delete this team?' });
    expect(dialog).toHaveTextContent('Its players, sponsors and managers will be unlinked from it.');
    fireEvent.click(within(dialog).getByRole('button', { name: 'Delete' }));
    await waitFor(() => expect(screen.getByTestId('location')).toHaveTextContent(/^\/teams$/));
    expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true);
  });
});
```

In `client/pages/teams/teams.test.tsx`, the test `'marks inactive teams (returned to signed-in users)'` asserts the classic-site link. Replace that assertion with `expect(await screen.findByRole('link', { name: 'Add team' })).toHaveAttribute('href', '/teams/new');`.

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/teams`
Expected: FAIL. `./TeamFormPage` can't be resolved.

- [ ] **Step 3: Implement**

`client/lib/ageGroups.ts`:

```ts
/** The age groups the legacy team form offered: "Under 6"…"Under 18", and 19 = "Over 18". */
export const AGE_GROUPS: { value: number; label: string }[] = [
  ...Array.from({ length: 13 }, (_, i) => ({ value: i + 6, label: `Under ${i + 6}` })),
  { value: 19, label: 'Over 18' },
];
```

`client/pages/teams/TeamsPage.tsx`:
- replace the `EditorLink` import with `useCanEdit` and `ButtonLink`;
- add `const { canEdit } = useCanEdit();`;
- set the header's actions to `actions={canEdit && <ButtonLink to="/teams/new">Add team</ButtonLink>}`.

`client/pages/teams/TeamPage.tsx`:
- in `TeamContent`, add `const { canEdit } = useCanEdit();` and `const navigate = useNavigate();`;
- add the imports `useNavigate` (react-router), `deleteTeam` (`../../api/teams`), `useCanEdit`, `DeleteButton` and `ButtonLink`;
- change `<PageHeader title={team.name} subtitle={team.description} />` to:

```tsx
      <PageHeader
        title={team.name}
        subtitle={team.description}
        actions={
          canEdit && (
            <>
              <ButtonLink to={`/team/${team.id}/edit`} variant="secondary" size="sm">
                Edit
              </ButtonLink>
              <DeleteButton
                confirmTitle="Delete this team?"
                confirmMessage="Its players, sponsors and managers will be unlinked from it. This can't be undone."
                onDelete={() => deleteTeam(team.id)}
                invalidate={[['teams'], ['team'], ['site']]}
                successMessage="Team deleted"
                after={() => navigate('/teams')}
              />
            </>
          )
        }
      />
```

`client/pages/teams/TeamFormPage.tsx`:

```tsx
import { useState, type FormEvent } from 'react';
import { useNavigate, useParams } from 'react-router';

import { createTeam, updateTeam, useTeam } from '../../api/teams';
import type { TeamSummary } from '../../api/types';
import { ImageField } from '../../components/edit/ImageField';
import { RequireEditor } from '../../components/edit/RequireEditor';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Checkbox } from '../../components/ui/Checkbox';
import { Input, Select, Textarea } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { PageHeader } from '../../components/ui/PageHeader';
import { useToast } from '../../components/ui/toast/useToast';
import { AGE_GROUPS } from '../../lib/ageGroups';
import { parseId } from '../../lib/ids';
import { emptyImage, type ImageValue } from '../../lib/images';
import { isNotFound } from '../../lib/notFound';
import NotFoundPage from '../NotFoundPage';

type TextKey =
  | 'name'
  | 'description'
  | 'league'
  | 'division'
  | 'leagueTable'
  | 'fixtures'
  | 'coach'
  | 'physio';

function TeamForm({ team }: { team?: TeamSummary }) {
  const heading = team ? 'Edit team' : 'New team';
  usePageTitle(heading);
  const navigate = useNavigate();
  const toast = useToast();
  const [text, setText] = useState<Record<TextKey, string>>({
    name: team?.name ?? '',
    description: team?.description ?? '',
    league: team?.league ?? '',
    division: team?.division ?? '',
    leagueTable: team?.leagueTableUrl ?? '',
    fixtures: team?.fixturesUrl ?? '',
    coach: team?.coach ?? '',
    physio: team?.physio ?? '',
  });
  const [ages, setAges] = useState(team ? String(team.ages) : '');
  const [isActive, setIsActive] = useState(team?.isActive ?? true);
  const [isYouth, setIsYouth] = useState(team?.isYouth ?? false);
  const [image, setImage] = useState<ImageValue>(emptyImage);
  const [missing, setMissing] = useState<{ name?: string; ages?: string }>({});

  const save = useSaveForm({
    submit: () => {
      const input = { ...text, ages: Number(ages), isActive, isYouth, image };
      return team ? updateTeam(team.id, input) : createTeam(input);
    },
    invalidate: [['teams'], ['team'], ['site']],
    onSaved: (t) => {
      toast.show({ tone: 'success', message: 'Team saved' });
      navigate(`/team/${t.id}`);
    },
  });

  const set = (key: TextKey) => (value: string) => setText((t) => ({ ...t, [key]: value }));
  const err = (key: string) => save.fieldErrors[key];

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const next = {
      name: text.name.trim() ? undefined : 'Enter a name',
      ages: ages ? undefined : 'Choose the age group',
    };
    setMissing(next);
    if (next.name || next.ages) {
      return;
    }
    void save.run();
  }

  const textField = (key: TextKey, label: string, type = 'text') => (
    <Field label={label} error={err(key)}>
      <Input type={type} value={text[key]} onChange={(e) => set(key)(e.target.value)} />
    </Field>
  );

  return (
    <>
      <PageHeader title={heading} />
      <form onSubmit={onSubmit} noValidate className="grid max-w-3xl gap-5 md:grid-cols-2">
        {save.formError && (
          <Alert tone="error" className="md:col-span-2">
            {save.formError}
          </Alert>
        )}
        <Field label="Name" error={missing.name ?? err('name')}>
          <Input value={text.name} onChange={(e) => set('name')(e.target.value)} />
        </Field>
        <Field label="Age group" error={missing.ages ?? err('ages')}>
          <Select value={ages} onChange={(e) => setAges(e.target.value)}>
            <option value="">Choose…</option>
            {AGE_GROUPS.map((g) => (
              <option key={g.value} value={String(g.value)}>
                {g.label}
              </option>
            ))}
          </Select>
        </Field>
        <Field label="Description" error={err('description')} className="md:col-span-2">
          <Textarea value={text.description} onChange={(e) => set('description')(e.target.value)} />
        </Field>
        {textField('league', 'League')}
        {textField('division', 'Division')}
        {textField('leagueTable', 'League table URL', 'url')}
        {textField('fixtures', 'Fixtures URL', 'url')}
        {textField('coach', 'Coach')}
        {textField('physio', 'Physio')}
        <div className="md:col-span-2">
          <ImageField
            label="Team photo"
            currentUrl={team?.imageUrl}
            allowRemove={Boolean(team)}
            value={image}
            onChange={setImage}
            error={err('file') ?? err('image')}
          />
        </div>
        <Checkbox label="Active team" checked={isActive} onChange={(e) => setIsActive(e.target.checked)} />
        <Checkbox label="Youth team" checked={isYouth} onChange={(e) => setIsYouth(e.target.checked)} />
        <div className="flex gap-2 md:col-span-2">
          <Button type="submit" loading={save.busy}>
            Save team
          </Button>
          <Button variant="secondary" onClick={() => navigate(-1)} disabled={save.busy}>
            Cancel
          </Button>
        </div>
      </form>
    </>
  );
}

function EditTeam({ id }: { id: number | null }) {
  const detail = useTeam(id);
  if (id === null || isNotFound(detail.error)) {
    return <NotFoundPage />;
  }
  return <QueryState query={detail}>{(d) => <TeamForm team={d.team} />}</QueryState>;
}

export default function TeamFormPage() {
  const { id } = useParams();
  return (
    <RequireEditor>{id === undefined ? <TeamForm /> : <EditTeam id={parseId(id)} />}</RequireEditor>
  );
}
```

`client/App.tsx`: add `const TeamFormPage = lazy(() => import('./pages/teams/TeamFormPage'));` and the routes `<Route path="teams/new" element={<TeamFormPage />} />` and `<Route path="team/:id/edit" element={<TeamFormPage />} />`.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/pages/teams client/lib client/App.tsx && yarn vitest run client/pages/teams && yarn test:client`
Expected: PASS. 6 new tests pass, and the suite is green.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add -A client
git commit -q -m "Add, edit and delete teams in the app" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 9: Documents, programmes and seasons

**Files:**
- Create: `client/pages/documents/AddDocumentDialog.tsx`, `client/pages/programmes/AddProgrammeDialog.tsx`, `client/pages/programmes/SeasonsDialog.tsx`
- Modify: `client/pages/documents/DocumentsPage.tsx`, `client/pages/programmes/ProgrammesPage.tsx`
- Test: `client/pages/documents/documentsEdit.test.tsx`, `client/pages/programmes/programmesEdit.test.tsx`

**Interfaces:**
- Consumes: `createDocument`, `deleteDocument`, `createProgramme`, `deleteProgramme`, `createSeason`, `renameSeason`, `deleteSeason` (Task 4); `useSeasons`; the toolkit (`FileField`, `DeleteButton`, `useSaveForm`, `useCanEdit`).
- Produces:
  - `AddDocumentDialog({ open, onClose })`
  - `AddProgrammeDialog({ open, onClose })`
  - `SeasonsDialog({ open, onClose })`

- [ ] **Step 1: Write the failing tests**

`client/pages/documents/documentsEdit.test.tsx`:

```tsx
import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import { documents, editor, manager, publicRoutes } from '../../test/fixtures';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import DocumentsPage from './DocumentsPage';

function renderDocs(overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(
    <Routes>
      <Route path="/documents" element={<DocumentsPage />} />
    </Routes>,
    { route: '/documents' },
  );
  return fetchMock;
}

describe('Documents editing', () => {
  it('hides the controls from a Manager', async () => {
    renderDocs({ '/api/v1/auth/me': manager });
    await screen.findByText(documents[0].name);
    expect(screen.queryByRole('button', { name: 'Add document' })).toBeNull();
    expect(screen.queryByRole('button', { name: `Delete ${documents[0].name}` })).toBeNull();
  });

  it('adds a document from the dialog', async () => {
    const fetchMock = renderDocs({ '/api/v1/documents': () => ({ status: 201, body: { id: 99, name: 'Rules', fileUrl: '/f' } }) });
    fireEvent.click(await screen.findByRole('button', { name: 'Add document' }));
    const dialog = screen.getByRole('dialog', { name: 'Add document' });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add document' }));
    expect(within(dialog).getByLabelText('Name')).toHaveAccessibleDescription('Enter a name');
    expect(within(dialog).getByLabelText('File')).toHaveAccessibleDescription('Choose a file');
    fireEvent.change(within(dialog).getByLabelText('Name'), { target: { value: 'Rules' } });
    const file = new File(['x'], 'rules.pdf', { type: 'application/pdf' });
    fireEvent.change(within(dialog).getByLabelText('File'), { target: { files: [file] } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add document' }));
    expect(await screen.findByRole('button', { name: 'Document added' })).toBeInTheDocument();
    expect(screen.queryByRole('dialog', { name: 'Add document' })).toBeNull();
    const fd = fetchMock.mock.calls.find(([, i]) => i?.method === 'POST')?.[1]?.body as FormData;
    expect(fd.get('name')).toBe('Rules');
    expect(fd.get('file')).toBe(file);
  });

  it('deletes a document after confirming', async () => {
    const fetchMock = renderDocs({ [`/api/v1/documents/${documents[0].id}`]: { status: 204 } });
    fireEvent.click(await screen.findByRole('button', { name: `Delete ${documents[0].name}` }));
    fireEvent.click(
      within(screen.getByRole('dialog', { name: `Delete ${documents[0].name}?` })).getByRole('button', { name: 'Delete' }),
    );
    expect(await screen.findByRole('button', { name: 'Document deleted' })).toBeInTheDocument();
    await waitFor(() => expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true));
  });
});
```

`client/pages/programmes/programmesEdit.test.tsx`:

```tsx
import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import { editor, programmes, publicRoutes, seasons } from '../../test/fixtures';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import ProgrammesPage from './ProgrammesPage';

function renderProgrammes(overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(
    <Routes>
      <Route path="/programmes" element={<ProgrammesPage />} />
    </Routes>,
    { route: '/programmes' },
  );
  return fetchMock;
}

const pdf = () => new File(['x'], 'p.pdf', { type: 'application/pdf' });

describe('Programmes editing', () => {
  it('adds a programme without a season (no seasonId sent)', async () => {
    const fetchMock = renderProgrammes({
      '/api/v1/programmes': () => ({ status: 201, body: programmes[0] }),
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Add programme' }));
    const dialog = screen.getByRole('dialog', { name: 'Add programme' });
    fireEvent.change(within(dialog).getByLabelText('Name'), { target: { value: 'vs Z' } });
    fireEvent.change(within(dialog).getByLabelText('Date'), { target: { value: '2026-10-03' } });
    fireEvent.change(within(dialog).getByLabelText('File'), { target: { files: [pdf()] } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add programme' }));
    expect(await screen.findByRole('button', { name: 'Programme added' })).toBeInTheDocument();
    const post = fetchMock.mock.calls.find(([u, i]) => i?.method === 'POST' && String(u).endsWith('/programmes'));
    const fd = post?.[1]?.body as FormData;
    expect(fd.get('date')).toBe('2026-10-03');
    expect(fd.has('seasonId')).toBe(false);
  });

  it('adds a programme to a chosen season', async () => {
    const fetchMock = renderProgrammes({
      '/api/v1/programmes': () => ({ status: 201, body: programmes[0] }),
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Add programme' }));
    const dialog = screen.getByRole('dialog', { name: 'Add programme' });
    await within(dialog).findByRole('option', { name: seasons[1].name });
    fireEvent.change(within(dialog).getByLabelText('Name'), { target: { value: 'vs Z' } });
    fireEvent.change(within(dialog).getByLabelText('Date'), { target: { value: '2026-10-03' } });
    fireEvent.change(within(dialog).getByLabelText('Season'), { target: { value: String(seasons[1].id) } });
    fireEvent.change(within(dialog).getByLabelText('File'), { target: { files: [pdf()] } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add programme' }));
    await screen.findByRole('button', { name: 'Programme added' });
    const post = fetchMock.mock.calls.find(([u, i]) => i?.method === 'POST' && String(u).endsWith('/programmes'));
    expect((post?.[1]?.body as FormData).get('seasonId')).toBe(String(seasons[1].id));
  });

  it('manages seasons: add, rename, and delete with the unlink warning', async () => {
    const fetchMock = renderProgrammes({
      '/api/v1/seasons': () => ({ body: seasons }),
      [`/api/v1/seasons/${seasons[0].id}`]: () => ({ body: { ...seasons[0], name: '2025/26' } }),
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Manage seasons' }));
    const dialog = screen.getByRole('dialog', { name: 'Seasons' });
    fireEvent.change(await within(dialog).findByLabelText('New season'), { target: { value: '2027-28' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add season' }));
    expect(await screen.findByRole('button', { name: 'Season added' })).toBeInTheDocument();
    expect(JSON.parse(String(fetchMock.mock.calls.find(([u, i]) => i?.method === 'POST' && String(u).endsWith('/seasons'))?.[1]?.body))).toEqual({ name: '2027-28' });

    fireEvent.click(within(dialog).getByRole('button', { name: `Rename ${seasons[0].name}` }));
    const input = within(dialog).getByLabelText(`New name for ${seasons[0].name}`);
    fireEvent.change(input, { target: { value: '2025/26' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Save name' }));
    expect(await screen.findByRole('button', { name: 'Season renamed' })).toBeInTheDocument();

    fireEvent.click(within(dialog).getByRole('button', { name: `Delete ${seasons[1].name}` }));
    const confirm = screen.getByRole('dialog', { name: `Delete season ${seasons[1].name}?` });
    expect(confirm).toHaveTextContent('Its programmes will stay, with no season.');
    fireEvent.click(within(confirm).getByRole('button', { name: 'Delete' }));
    expect(await screen.findByRole('button', { name: 'Season deleted' })).toBeInTheDocument();
    await waitFor(() => expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true));
  });

  it('deletes a programme from its row', async () => {
    const fetchMock = renderProgrammes({ [`/api/v1/programmes/${programmes[0].id}`]: { status: 204 } });
    fireEvent.click(await screen.findByRole('button', { name: `Delete ${programmes[0].name}` }));
    fireEvent.click(
      within(screen.getByRole('dialog', { name: `Delete ${programmes[0].name}?` })).getByRole('button', { name: 'Delete' }),
    );
    expect(await screen.findByRole('button', { name: 'Programme deleted' })).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true);
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/documents client/pages/programmes`
Expected: FAIL. There are no "Add document", "Add programme" or "Manage seasons" controls.

- [ ] **Step 3: Implement**

`client/pages/documents/AddDocumentDialog.tsx`:

```tsx
import { useState, type FormEvent } from 'react';

import { createDocument } from '../../api/documents';
import { queryKeys } from '../../api/queries';
import { FileField } from '../../components/edit/FileField';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Input } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';

export function AddDocumentDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const toast = useToast();
  const [name, setName] = useState('');
  const [file, setFile] = useState<File | null>(null);
  const [missing, setMissing] = useState<{ name?: string; file?: string }>({});
  const save = useSaveForm({
    submit: () => createDocument({ name, file: file as File }),
    invalidate: [queryKeys.documents],
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Document added' });
      setName('');
      setFile(null);
      onClose();
    },
  });

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const next = {
      name: name.trim() ? undefined : 'Enter a name',
      file: file ? undefined : 'Choose a file',
    };
    setMissing(next);
    if (!next.name && !next.file) {
      void save.run();
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Add document">
      <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
        {save.formError && <Alert tone="error">{save.formError}</Alert>}
        <Field label="Name" error={missing.name ?? save.fieldErrors.name}>
          <Input value={name} onChange={(e) => setName(e.target.value)} />
        </Field>
        <FileField label="File" value={file} onChange={setFile} error={missing.file ?? save.fieldErrors.file} />
        <div className="flex justify-end gap-2">
          <Button variant="secondary" onClick={onClose} disabled={save.busy}>
            Cancel
          </Button>
          <Button type="submit" loading={save.busy}>
            Add document
          </Button>
        </div>
      </form>
    </Modal>
  );
}
```

`client/pages/documents/DocumentsPage.tsx`:
- replace the `EditorLink` import with `useState` (react), `deleteDocument` (`../../api/documents`), `queryKeys` (`../../api/queries`), `useCanEdit`, `DeleteButton`, `Button` and `AddDocumentDialog` (`./AddDocumentDialog`);
- add `const { canEdit } = useCanEdit();` and `const [adding, setAdding] = useState(false);`;
- set the header to:

```tsx
      <PageHeader
        title="Documents"
        actions={canEdit && <Button onClick={() => setAdding(true)}>Add document</Button>}
      />
```

- inside each `<li>`, after the Download `ButtonLink`, add:

```tsx
                      {canEdit && (
                        <DeleteButton
                          ariaLabel={`Delete ${d.name}`}
                          confirmTitle={`Delete ${d.name}?`}
                          confirmMessage="This can't be undone."
                          onDelete={() => deleteDocument(d.id)}
                          invalidate={[queryKeys.documents]}
                          successMessage="Document deleted"
                        />
                      )}
```

  and wrap the Download link and the delete button together in `<div className="flex gap-2">…</div>`;
- before the closing fragment, add `<AddDocumentDialog open={adding} onClose={() => setAdding(false)} />`.

`client/pages/programmes/AddProgrammeDialog.tsx`:

```tsx
import { useState, type FormEvent } from 'react';

import { createProgramme, useSeasons } from '../../api/programmes';
import { FileField } from '../../components/edit/FileField';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Input, Select } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';

type Missing = { name?: string; date?: string; file?: string };

export function AddProgrammeDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const toast = useToast();
  const seasons = useSeasons();
  const [name, setName] = useState('');
  const [date, setDate] = useState('');
  const [seasonId, setSeasonId] = useState('');
  const [file, setFile] = useState<File | null>(null);
  const [missing, setMissing] = useState<Missing>({});
  const save = useSaveForm({
    submit: () =>
      createProgramme({ name, date, seasonId: seasonId ? Number(seasonId) : null, file: file as File }),
    invalidate: [['programmes']],
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Programme added' });
      setName('');
      setDate('');
      setSeasonId('');
      setFile(null);
      onClose();
    },
  });

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const next: Missing = {
      name: name.trim() ? undefined : 'Enter a name',
      date: date ? undefined : 'Choose the date',
      file: file ? undefined : 'Choose a file',
    };
    setMissing(next);
    if (!next.name && !next.date && !next.file) {
      void save.run();
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Add programme">
      <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
        {save.formError && <Alert tone="error">{save.formError}</Alert>}
        <Field label="Name" error={missing.name ?? save.fieldErrors.name}>
          <Input value={name} onChange={(e) => setName(e.target.value)} />
        </Field>
        <Field label="Date" error={missing.date ?? save.fieldErrors.date}>
          <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
        </Field>
        <Field label="Season" error={save.fieldErrors.seasonId}>
          <Select value={seasonId} onChange={(e) => setSeasonId(e.target.value)}>
            <option value="">No season</option>
            {seasons.data?.map((s) => (
              <option key={s.id} value={String(s.id)}>
                {s.name}
              </option>
            ))}
          </Select>
        </Field>
        <FileField label="File" value={file} onChange={setFile} error={missing.file ?? save.fieldErrors.file} />
        <div className="flex justify-end gap-2">
          <Button variant="secondary" onClick={onClose} disabled={save.busy}>
            Cancel
          </Button>
          <Button type="submit" loading={save.busy}>
            Add programme
          </Button>
        </div>
      </form>
    </Modal>
  );
}
```

`client/pages/programmes/SeasonsDialog.tsx`:

```tsx
import { useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';

import { createSeason, deleteSeason, renameSeason, useSeasons, type Season } from '../../api/programmes';
import { DeleteButton } from '../../components/edit/DeleteButton';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Input } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';

const refresh = [['seasons'], ['programmes']];

function SeasonRow({ season }: { season: Season }) {
  const toast = useToast();
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(season.name);
  const save = useSaveForm({
    submit: () => renameSeason(season.id, name),
    invalidate: refresh,
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Season renamed' });
      setEditing(false);
    },
  });
  if (editing) {
    return (
      <li className="flex flex-wrap items-end gap-2 py-2">
        <Field label={`New name for ${season.name}`} error={save.fieldErrors.name ?? save.formError ?? undefined} className="flex-1">
          <Input value={name} onChange={(e) => setName(e.target.value)} />
        </Field>
        <Button size="sm" loading={save.busy} onClick={() => void save.run()}>
          Save name
        </Button>
        <Button size="sm" variant="secondary" onClick={() => setEditing(false)}>
          Cancel
        </Button>
      </li>
    );
  }
  return (
    <li className="flex items-center justify-between gap-2 py-2">
      <span>{season.name}</span>
      <span className="flex gap-2">
        <Button size="sm" variant="secondary" aria-label={`Rename ${season.name}`} onClick={() => setEditing(true)}>
          Rename
        </Button>
        <DeleteButton
          ariaLabel={`Delete ${season.name}`}
          confirmTitle={`Delete season ${season.name}?`}
          confirmMessage="Its programmes will stay, with no season."
          onDelete={() => deleteSeason(season.id)}
          invalidate={refresh}
          successMessage="Season deleted"
        />
      </span>
    </li>
  );
}

export function SeasonsDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const toast = useToast();
  const seasons = useSeasons();
  const queryClient = useQueryClient();
  const [name, setName] = useState('');
  const [missing, setMissing] = useState<string | undefined>();
  const add = useSaveForm({
    submit: () => createSeason(name),
    invalidate: refresh,
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Season added' });
      setName('');
    },
  });

  return (
    <Modal open={open} onClose={onClose} title="Seasons">
      <div className="flex flex-col gap-4">
        {seasons.isError && (
          <Alert tone="error">
            Couldn&apos;t load the seasons.{' '}
            <Button size="sm" variant="secondary" onClick={() => void queryClient.refetchQueries({ queryKey: ['seasons'] })}>
              Retry
            </Button>
          </Alert>
        )}
        <ul className="divide-y divide-line">
          {seasons.data?.map((s) => <SeasonRow key={`${s.id}-${s.name}`} season={s} />)}
        </ul>
        <form
          noValidate
          className="flex flex-wrap items-end gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            if (!name.trim()) {
              setMissing('Enter a season name');
              return;
            }
            setMissing(undefined);
            void add.run();
          }}
        >
          <Field label="New season" error={missing ?? add.fieldErrors.name ?? add.formError ?? undefined} className="flex-1">
            <Input value={name} placeholder="e.g. 2027-28" onChange={(e) => setName(e.target.value)} />
          </Field>
          <Button type="submit" loading={add.busy}>
            Add season
          </Button>
        </form>
        <div className="flex justify-end">
          <Button variant="secondary" onClick={onClose}>
            Done
          </Button>
        </div>
      </div>
    </Modal>
  );
}
```

`client/pages/programmes/ProgrammesPage.tsx`:
- replace the `EditorLink` import with `useState`, `deleteProgramme`, `useCanEdit`, `DeleteButton`, `Button`, `AddProgrammeDialog` and `SeasonsDialog`;
- in `ProgrammesPage`, add `const { canEdit } = useCanEdit();`, `const [adding, setAdding] = useState(false);` and `const [managing, setManaging] = useState(false);`;
- set the header to:

```tsx
      <PageHeader
        title="Programmes"
        actions={
          canEdit && (
            <>
              <Button onClick={() => setAdding(true)}>Add programme</Button>
              <Button variant="secondary" onClick={() => setManaging(true)}>
                Manage seasons
              </Button>
            </>
          )
        }
      />
```

- change `SeasonGroup`'s props to `{ name: string; items: Programme[]; canEdit: boolean }` and pass `canEdit={canEdit}`;
- in its row, wrap the View link in `<span className="flex gap-2">…</span>` and add:

```tsx
              {canEdit && (
                <DeleteButton
                  ariaLabel={`Delete ${p.name}`}
                  confirmTitle={`Delete ${p.name}?`}
                  confirmMessage="This can't be undone."
                  onDelete={() => deleteProgramme(p.id)}
                  invalidate={[['programmes']]}
                  successMessage="Programme deleted"
                />
              )}
```

- before the closing fragment, add `<AddProgrammeDialog open={adding} onClose={() => setAdding(false)} />` and `<SeasonsDialog open={managing} onClose={() => setManaging(false)} />`.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/pages/documents client/pages/programmes && yarn vitest run client/pages/documents client/pages/programmes && yarn test:client`
Expected: PASS. 7 new tests pass, and the suite is green. `SeasonRow` is keyed by id and name, so a renamed season re-mounts with its new name.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add -A client
git commit -q -m "Add and delete documents and programmes, and manage seasons, in the app" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 10: Sponsors, affiliations and gallery photos

**Files:**
- Create: `client/pages/sponsors/AddSponsorDialog.tsx`, `client/pages/home/AddAffiliationDialog.tsx`, `client/pages/gallery/AddPhotoDialog.tsx`
- Modify:
  - `client/pages/sponsors/SponsorsPage.tsx`
  - `client/components/page/LogoRow.tsx` (optional per-item actions, and showing the row empty for editors)
  - `client/pages/home/HomePage.tsx`
  - `client/pages/gallery/GalleryPage.tsx`
  - `client/pages/gallery/gallery.test.tsx` (the editor-link assertions)
- Test: `client/pages/sponsors/sponsorsEdit.test.tsx`, `client/pages/home/affiliationsEdit.test.tsx`, `client/pages/gallery/galleryEdit.test.tsx`

**Interfaces:**
- Consumes: `createSponsor`, `deleteSponsor`, `createAffiliation`, `deleteAffiliation`, `createPhoto`, `deletePhoto` (Task 4); `SPONSOR_TEAM_CHOICES` (Task 1); the toolkit.
- Produces: `LogoRow({ title, items, itemAction?: (item: LogoItem) => ReactNode, showWhenEmpty?: boolean, emptyText?: string, headerAction?: ReactNode })`.

- [ ] **Step 1: Write the failing tests**

`client/pages/sponsors/sponsorsEdit.test.tsx`:

```tsx
import { fireEvent, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { editor, manager, publicRoutes, sponsor, team } from '../../test/fixtures';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import SponsorsPage from './SponsorsPage';

beforeEach(() => {
  Object.assign(URL, { createObjectURL: vi.fn(() => 'blob:x'), revokeObjectURL: vi.fn() });
});
afterEach(() => {
  delete (URL as unknown as Record<string, unknown>).createObjectURL;
  delete (URL as unknown as Record<string, unknown>).revokeObjectURL;
});

function renderSponsors(overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(<SponsorsPage />);
  return fetchMock;
}

describe('Sponsors editing', () => {
  it('hides the controls from a Manager', async () => {
    renderSponsors({ '/api/v1/auth/me': manager });
    await screen.findByRole('heading', { name: sponsor.name });
    expect(screen.queryByRole('button', { name: 'Add sponsor' })).toBeNull();
  });

  it('offers the team choices and sends the chosen team id', async () => {
    // First call is the list GET, second the create POST, then the refetch.
    let n = 0;
    const fetchMock = renderSponsors({
      '/api/v1/sponsors': () => (n++ === 1 ? { status: 201, body: sponsor } : { body: [sponsor] }),
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Add sponsor' }));
    const dialog = screen.getByRole('dialog', { name: 'Add sponsor' });
    const select = within(dialog).getByLabelText('Sponsors');
    await within(dialog).findByRole('option', { name: team.name });
    expect(
      within(select)
        .getAllByRole('option')
        .map((o) => (o as HTMLOptionElement).value)
        .slice(0, 4),
    ).toEqual(['', 'A', 'O', 'Y']);
    fireEvent.change(within(dialog).getByLabelText('Name'), { target: { value: 'Corner Shop' } });
    fireEvent.change(within(dialog).getByLabelText('Logo'), {
      target: { files: [new File(['x'], 'l.png', { type: 'image/png' })] },
    });
    fireEvent.change(select, { target: { value: String(team.id) } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add sponsor' }));
    expect(await screen.findByRole('button', { name: 'Sponsor added' })).toBeInTheDocument();
    const fd = fetchMock.mock.calls.find(([, i]) => i?.method === 'POST')?.[1]?.body as FormData;
    expect(fd.get('team')).toBe(String(team.id));
    expect(fd.get('name')).toBe('Corner Shop');
  });

  it('requires a name and a logo', async () => {
    renderSponsors();
    fireEvent.click(await screen.findByRole('button', { name: 'Add sponsor' }));
    const dialog = screen.getByRole('dialog', { name: 'Add sponsor' });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add sponsor' }));
    expect(within(dialog).getByLabelText('Name')).toHaveAccessibleDescription('Enter a name');
    expect(within(dialog).getByLabelText('Logo')).toHaveAccessibleDescription('Choose a logo image');
  });

  it('deletes a sponsor after confirming', async () => {
    const fetchMock = renderSponsors({ [`/api/v1/sponsors/${sponsor.id}`]: { status: 204 } });
    fireEvent.click(await screen.findByRole('button', { name: `Delete ${sponsor.name}` }));
    fireEvent.click(
      within(screen.getByRole('dialog', { name: `Delete ${sponsor.name}?` })).getByRole('button', { name: 'Delete' }),
    );
    expect(await screen.findByRole('button', { name: 'Sponsor deleted' })).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true);
  });
});
```

`client/pages/home/affiliationsEdit.test.tsx`:

```tsx
import { fireEvent, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { affiliation, editor, home, manager, publicRoutes } from '../../test/fixtures';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import HomePage from './HomePage';

beforeEach(() => {
  Object.assign(URL, { createObjectURL: vi.fn(() => 'blob:x'), revokeObjectURL: vi.fn() });
});
afterEach(() => {
  delete (URL as unknown as Record<string, unknown>).createObjectURL;
  delete (URL as unknown as Record<string, unknown>).revokeObjectURL;
});

function renderHome(overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(<HomePage />);
  return fetchMock;
}

describe('Affiliations editing', () => {
  it('hides the controls from a Manager', async () => {
    renderHome({ '/api/v1/auth/me': manager });
    await screen.findByRole('region', { name: 'Affiliations' });
    expect(screen.queryByRole('button', { name: 'Add affiliation' })).toBeNull();
  });

  it('shows an empty Affiliations row to editors so they can add one', async () => {
    renderHome({ '/api/v1/home': { body: { ...home, affiliations: [] } } });
    const row = await screen.findByRole('region', { name: 'Affiliations' });
    expect(within(row).getByText('No affiliations yet')).toBeInTheDocument();
    expect(within(row).getByRole('button', { name: 'Add affiliation' })).toBeInTheDocument();
  });

  it('adds and deletes an affiliation', async () => {
    let n = 0;
    const fetchMock = renderHome({
      '/api/v1/affiliations': () => ({ status: 201, body: { id: 50, name: 'League' } }),
      [`/api/v1/affiliations/${affiliation.id}`]: () => (n++, { status: 204 }),
    });
    const row = await screen.findByRole('region', { name: 'Affiliations' });
    fireEvent.click(within(row).getByRole('button', { name: 'Add affiliation' }));
    const dialog = screen.getByRole('dialog', { name: 'Add affiliation' });
    fireEvent.change(within(dialog).getByLabelText('Name'), { target: { value: 'League' } });
    fireEvent.change(within(dialog).getByLabelText('Logo'), {
      target: { files: [new File(['x'], 'l.png', { type: 'image/png' })] },
    });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add affiliation' }));
    expect(await screen.findByRole('button', { name: 'Affiliation added' })).toBeInTheDocument();

    fireEvent.click(within(row).getByRole('button', { name: `Delete ${affiliation.name}` }));
    fireEvent.click(
      within(screen.getByRole('dialog', { name: `Delete ${affiliation.name}?` })).getByRole('button', { name: 'Delete' }),
    );
    expect(await screen.findByRole('button', { name: 'Affiliation deleted' })).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true);
    expect(n).toBe(1);
  });
});
```

`client/pages/gallery/galleryEdit.test.tsx`:

```tsx
import { fireEvent, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { editor, galleryImages, manager, photographer, publicRoutes } from '../../test/fixtures';
import type { MockResponse, MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import GalleryPage from './GalleryPage';

beforeEach(() => {
  Object.assign(URL, { createObjectURL: vi.fn(() => 'blob:x'), revokeObjectURL: vi.fn() });
});
afterEach(() => {
  delete (URL as unknown as Record<string, unknown>).createObjectURL;
  delete (URL as unknown as Record<string, unknown>).revokeObjectURL;
});

function renderGallery(me: MockResponse, overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': me, ...overrides }));
  renderWithProviders(<GalleryPage />);
  return fetchMock;
}

describe('Gallery editing', () => {
  it('hides the controls from a Manager', async () => {
    renderGallery(manager);
    await screen.findByRole('button', { name: 'Cup final' });
    expect(screen.queryByRole('button', { name: 'Add photo' })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Delete photo Cup final' })).toBeNull();
  });

  it('lets a photographer add a photo with a caption', async () => {
    let n = 0;
    const fetchMock = renderGallery(photographer, {
      '/api/v1/gallery': () => (n++ === 1 ? { status: 201, body: galleryImages[0] } : { body: galleryImages }),
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Add photo' }));
    const dialog = screen.getByRole('dialog', { name: 'Add photo' });
    fireEvent.change(within(dialog).getByLabelText('Photo'), {
      target: { files: [new File(['x'], 'p.png', { type: 'image/png' })] },
    });
    fireEvent.change(within(dialog).getByLabelText('Caption'), { target: { value: 'Awards' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add photo' }));
    expect(await screen.findByRole('button', { name: 'Photo added' })).toBeInTheDocument();
    const fd = fetchMock.mock.calls.find(([, i]) => i?.method === 'POST')?.[1]?.body as FormData;
    expect(fd.get('caption')).toBe('Awards');
  });

  it('requires a photo', async () => {
    renderGallery(editor);
    fireEvent.click(await screen.findByRole('button', { name: 'Add photo' }));
    const dialog = screen.getByRole('dialog', { name: 'Add photo' });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add photo' }));
    expect(within(dialog).getByLabelText('Photo')).toHaveAccessibleDescription('Choose a photo');
  });

  it('deletes a photo after confirming, from a control beside the thumbnail', async () => {
    const fetchMock = renderGallery(editor, {
      [`/api/v1/gallery/${galleryImages[0].id}`]: { status: 204 },
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Delete photo Cup final' }));
    fireEvent.click(
      within(screen.getByRole('dialog', { name: 'Delete this photo?' })).getByRole('button', { name: 'Delete' }),
    );
    expect(await screen.findByRole('button', { name: 'Photo deleted' })).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true);
  });
});
```

In `client/pages/gallery/gallery.test.tsx`, replace the two editor-link tests with:

```tsx
  it('shows Add photo to people who can manage the gallery', async () => {
    await renderGallery({ '/api/v1/auth/me': editor });
    expect(await screen.findByRole('button', { name: 'Add photo' })).toBeInTheDocument();
  });

  it('hides Add photo from managers', async () => {
    await renderGallery({ '/api/v1/auth/me': manager });
    await new Promise((r) => setTimeout(r, 0));
    expect(screen.queryByRole('button', { name: 'Add photo' })).toBeNull();
  });
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/sponsors client/pages/home client/pages/gallery`
Expected: FAIL. There are no Add or Delete controls yet.

- [ ] **Step 3: Implement**

`client/components/page/LogoRow.tsx`: extend it as follows. The row is still hidden when it's empty and `showWhenEmpty` is false, so public behaviour doesn't change.

```tsx
export interface LogoRowProps {
  title: string;
  items: LogoItem[];
  /** Extra control under each logo (e.g. an editor's delete button). */
  itemAction?: (item: LogoItem) => ReactNode;
  /** Render the row even with no items (editors, so they can add one). */
  showWhenEmpty?: boolean;
  emptyText?: string;
  /** A control beside the row's heading (e.g. "Add affiliation"). */
  headerAction?: ReactNode;
}

export function LogoRow({ title, items, itemAction, showWhenEmpty, emptyText, headerAction }: LogoRowProps) {
  const headingId = useId();
  if (items.length === 0 && !showWhenEmpty) {
    return null;
  }
  return (
    <section aria-labelledby={headingId}>
      <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
        <h2 id={headingId} className="font-display text-2xl font-extrabold tracking-wide uppercase">
          {title}
        </h2>
        {headerAction}
      </div>
      {items.length === 0 ? (
        <p className="text-sm text-muted">{emptyText}</p>
      ) : (
        <ul className="flex flex-wrap gap-3">
          {items.map((item) => (
            <li key={item.id} className="flex flex-col items-center gap-1">
              <Logo item={item} />
              {itemAction?.(item)}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
```

Add `import type { ReactNode } from 'react';` to it (`useId` stays).

`client/pages/home/AddAffiliationDialog.tsx`:

```tsx
import { useState, type FormEvent } from 'react';

import { createAffiliation } from '../../api/home';
import { queryKeys } from '../../api/queries';
import { ImageField } from '../../components/edit/ImageField';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Input } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';
import { emptyImage, type ImageValue } from '../../lib/images';

export function AddAffiliationDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const toast = useToast();
  const [name, setName] = useState('');
  const [website, setWebsite] = useState('');
  const [image, setImage] = useState<ImageValue>(emptyImage);
  const [missing, setMissing] = useState<{ name?: string; image?: string }>({});
  const save = useSaveForm({
    submit: () => createAffiliation({ name, website, image: image.file as File }),
    invalidate: [queryKeys.home],
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Affiliation added' });
      setName('');
      setWebsite('');
      setImage(emptyImage);
      onClose();
    },
  });

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const next = {
      name: name.trim() ? undefined : 'Enter a name',
      image: image.file ? undefined : 'Choose a logo image',
    };
    setMissing(next);
    if (!next.name && !next.image) {
      void save.run();
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Add affiliation">
      <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
        {save.formError && <Alert tone="error">{save.formError}</Alert>}
        <Field label="Name" error={missing.name ?? save.fieldErrors.name}>
          <Input value={name} onChange={(e) => setName(e.target.value)} />
        </Field>
        <ImageField label="Logo" value={image} onChange={setImage} error={missing.image ?? save.fieldErrors.file ?? save.fieldErrors.image} />
        <Field label="Website" error={save.fieldErrors.website}>
          <Input type="url" value={website} onChange={(e) => setWebsite(e.target.value)} />
        </Field>
        <div className="flex justify-end gap-2">
          <Button variant="secondary" onClick={onClose} disabled={save.busy}>
            Cancel
          </Button>
          <Button type="submit" loading={save.busy}>
            Add affiliation
          </Button>
        </div>
      </form>
    </Modal>
  );
}
```

`client/pages/home/HomePage.tsx`:
- in `HomeContent`, add `const { canEdit } = useCanEdit();` and `const [adding, setAdding] = useState(false);`;
- import `useState`, `useCanEdit`, `Button`, `DeleteButton`, `deleteAffiliation` (`../../api/home`), `queryKeys` and `AddAffiliationDialog`;
- replace `<LogoRow title="Affiliations" items={affiliations} />` with:

```tsx
      <LogoRow
        title="Affiliations"
        items={affiliations}
        showWhenEmpty={canEdit}
        emptyText="No affiliations yet"
        headerAction={
          canEdit && (
            <Button size="sm" onClick={() => setAdding(true)}>
              Add affiliation
            </Button>
          )
        }
        itemAction={
          canEdit
            ? (item) => (
                <DeleteButton
                  ariaLabel={`Delete ${item.name}`}
                  confirmTitle={`Delete ${item.name}?`}
                  confirmMessage="This can't be undone."
                  onDelete={() => deleteAffiliation(item.id)}
                  invalidate={[queryKeys.home]}
                  successMessage="Affiliation deleted"
                />
              )
            : undefined
        }
      />
      <AddAffiliationDialog open={adding} onClose={() => setAdding(false)} />
```

`client/pages/sponsors/AddSponsorDialog.tsx`:

```tsx
import { useState, type FormEvent } from 'react';

import { queryKeys, useSite } from '../../api/queries';
import { createSponsor } from '../../api/sponsors';
import { ImageField } from '../../components/edit/ImageField';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Input, Select } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';
import { emptyImage, type ImageValue } from '../../lib/images';
import { SPONSOR_TEAM_CHOICES } from '../../lib/sponsorTeam';

export function AddSponsorDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const toast = useToast();
  const site = useSite();
  const [name, setName] = useState('');
  const [website, setWebsite] = useState('');
  const [purpose, setPurpose] = useState('');
  const [team, setTeam] = useState('');
  const [image, setImage] = useState<ImageValue>(emptyImage);
  const [missing, setMissing] = useState<{ name?: string; image?: string }>({});
  const save = useSaveForm({
    submit: () => createSponsor({ name, website, purpose, team, image: image.file as File }),
    invalidate: [queryKeys.sponsors, queryKeys.home, ['team']],
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Sponsor added' });
      setName('');
      setWebsite('');
      setPurpose('');
      setTeam('');
      setImage(emptyImage);
      onClose();
    },
  });

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const next = {
      name: name.trim() ? undefined : 'Enter a name',
      image: image.file ? undefined : 'Choose a logo image',
    };
    setMissing(next);
    if (!next.name && !next.image) {
      void save.run();
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Add sponsor">
      <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
        {save.formError && <Alert tone="error">{save.formError}</Alert>}
        <Field label="Name" error={missing.name ?? save.fieldErrors.name}>
          <Input value={name} onChange={(e) => setName(e.target.value)} />
        </Field>
        <ImageField label="Logo" value={image} onChange={setImage} error={missing.image ?? save.fieldErrors.file ?? save.fieldErrors.image} />
        <Field label="Website" error={save.fieldErrors.website}>
          <Input type="url" value={website} onChange={(e) => setWebsite(e.target.value)} />
        </Field>
        <Field label="Purpose" error={save.fieldErrors.purpose}>
          <Input value={purpose} placeholder="e.g. Kit sponsor" onChange={(e) => setPurpose(e.target.value)} />
        </Field>
        <Field label="Sponsors" error={save.fieldErrors.team}>
          <Select value={team} onChange={(e) => setTeam(e.target.value)}>
            {SPONSOR_TEAM_CHOICES.map((c) => (
              <option key={c.value} value={c.value}>
                {c.label}
              </option>
            ))}
            {site.data?.teams.map((t) => (
              <option key={t.id} value={String(t.id)}>
                {t.name}
              </option>
            ))}
          </Select>
        </Field>
        <div className="flex justify-end gap-2">
          <Button variant="secondary" onClick={onClose} disabled={save.busy}>
            Cancel
          </Button>
          <Button type="submit" loading={save.busy}>
            Add sponsor
          </Button>
        </div>
      </form>
    </Modal>
  );
}
```

`client/pages/sponsors/SponsorsPage.tsx`:
- replace the `EditorLink` import with `useState`, `deleteSponsor`, `queryKeys` (already imported with `useSite`), `useCanEdit`, `Button`, `DeleteButton` and `AddSponsorDialog`;
- add `const { canEdit } = useCanEdit();` and `const [adding, setAdding] = useState(false);`;
- set the header's actions to `actions={canEdit && <Button onClick={() => setAdding(true)}>Add sponsor</Button>}`;
- at the end of each card's `CardBody`, add:

```tsx
                  {canEdit && (
                    <div className="pt-2">
                      <DeleteButton
                        ariaLabel={`Delete ${s.name}`}
                        confirmTitle={`Delete ${s.name}?`}
                        confirmMessage="This can't be undone."
                        onDelete={() => deleteSponsor(s.id)}
                        invalidate={[queryKeys.sponsors, queryKeys.home, ['team']]}
                        successMessage="Sponsor deleted"
                      />
                    </div>
                  )}
```

- before the closing fragment, add `<AddSponsorDialog open={adding} onClose={() => setAdding(false)} />`.

`client/pages/gallery/AddPhotoDialog.tsx`:

```tsx
import { useState, type FormEvent } from 'react';

import { createPhoto } from '../../api/gallery';
import { queryKeys } from '../../api/queries';
import { ImageField } from '../../components/edit/ImageField';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Input } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';
import { emptyImage, type ImageValue } from '../../lib/images';

export function AddPhotoDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const toast = useToast();
  const [caption, setCaption] = useState('');
  const [image, setImage] = useState<ImageValue>(emptyImage);
  const [missing, setMissing] = useState<string | undefined>();
  const save = useSaveForm({
    submit: () => createPhoto({ caption, image: image.file as File }),
    invalidate: [queryKeys.gallery],
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Photo added' });
      setCaption('');
      setImage(emptyImage);
      onClose();
    },
  });

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!image.file) {
      setMissing('Choose a photo');
      return;
    }
    setMissing(undefined);
    void save.run();
  }

  return (
    <Modal open={open} onClose={onClose} title="Add photo">
      <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
        {save.formError && <Alert tone="error">{save.formError}</Alert>}
        <ImageField label="Photo" value={image} onChange={setImage} error={missing ?? save.fieldErrors.file ?? save.fieldErrors.image} />
        <Field label="Caption" error={save.fieldErrors.caption}>
          <Input value={caption} onChange={(e) => setCaption(e.target.value)} />
        </Field>
        <div className="flex justify-end gap-2">
          <Button variant="secondary" onClick={onClose} disabled={save.busy}>
            Cancel
          </Button>
          <Button type="submit" loading={save.busy}>
            Add photo
          </Button>
        </div>
      </form>
    </Modal>
  );
}
```

`client/pages/gallery/GalleryPage.tsx`:
- replace the `EditorLink` import with `useCanEdit`, `Button`, `DeleteButton`, `deletePhoto` (`../../api/gallery`), `queryKeys` and `AddPhotoDialog`;
- add `const { canManageGallery } = useCanEdit();` and `const [adding, setAdding] = useState(false);`;
- set the header's actions to `actions={canManageGallery && <Button onClick={() => setAdding(true)}>Add photo</Button>}`;
- in each `<li>`, after the thumbnail `<button>…</button>` (as a sibling, not inside it), add:

```tsx
                  {canManageGallery && (
                    <div className="mt-1 flex justify-end">
                      <DeleteButton
                        ariaLabel={`Delete photo ${img.caption || i + 1}`}
                        confirmTitle="Delete this photo?"
                        confirmMessage="This can't be undone."
                        onDelete={() => deletePhoto(img.id)}
                        invalidate={[queryKeys.gallery]}
                        successMessage="Photo deleted"
                      />
                    </div>
                  )}
```

- before the closing fragment, add `<AddPhotoDialog open={adding} onClose={() => setAdding(false)} />`.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/pages client/components/page && yarn vitest run client/pages/sponsors client/pages/home client/pages/gallery && yarn test:client`
Expected: PASS. 11 new tests pass, and the suite is green.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add -A client
git commit -q -m "Add and delete sponsors, affiliations and gallery photos in the app" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 11: Retire the classic-site links, then accessibility, the bundle check and docs

**Files:**
- Delete: `client/components/page/EditorLink.tsx` (and its tests in `client/components/page/pieces.test.tsx`)
- Modify: `client/a11y.test.tsx`, `client/App.test.tsx`, `README.md`

- [ ] **Step 1: Check that nothing still uses `EditorLink`, then delete it**

Run: `grep -rn "EditorLink" client | grep -v pieces.test.tsx`
Expected: no output. Every page switched to in-app controls in Tasks 5–10.

Delete `client/components/page/EditorLink.tsx`, and in `client/components/page/pieces.test.tsx` delete the `describe('EditorLink', …)` block and its import.

- [ ] **Step 2: Extend the route and accessibility tests**

In `client/App.test.tsx`, add route cases. They render as anonymous by default, which shows the permission message's heading instead of a page `h1`, so add a separate `it.each` for editors:

```tsx
  it.each([
    ['/news/new', 'New article'],
    [`/news/${newsArticle.id}/edit`, 'Edit article'],
    ['/whatson/new', 'New event'],
    [`/whatson/${event.id}/edit`, 'Edit event'],
    ['/info/edit', 'Edit information'],
    ['/teams/new', 'New team'],
    [`/team/${team.id}/edit`, 'Edit team'],
  ])('routes %s to its edit page for editors', async (path, heading) => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': editor }));
    renderWithProviders(<App />, { route: path });
    expect(await screen.findByRole('heading', { level: 1, name: heading }, lazy)).toBeInTheDocument();
  });
```

Import `editor` from the fixtures and `mockFetch` as already used.

In `client/a11y.test.tsx`:
- add the seven edit routes above to `routes`, so they're audited signed out (permission message) and signed in as `editor` (the form);
- the wait for an `h1` still works: the permission message has no `h1`, so the check becomes: for anonymous users, wait for `"You don't have permission to edit this"` instead. Implement this by replacing the heading wait with:

```tsx
    if (route.endsWith('/new') || route.endsWith('/edit')) {
      await screen.findByText(
        me === anonymous ? "You don't have permission to edit this" : /Save|Edit information/,
        undefined,
        { timeout: 3000 },
      );
    } else {
      await screen.findByRole('heading', { level: 1 }, { timeout: 3000 });
    }
```

Then add one more test at the end of the file (outside `describe.each`) for an open dialog:

```tsx
it('has no axe violations with the Add document dialog open', async () => {
  mockFetch(publicRoutes({ '/api/v1/auth/me': editor }));
  const { container } = renderWithProviders(<App />, { route: '/documents' });
  fireEvent.click(await screen.findByRole('button', { name: 'Add document' }, { timeout: 3000 }));
  await screen.findByRole('dialog', { name: 'Add document' });
  expect(await axeViolations(container)).toEqual([]);
});
```

(import `fireEvent`).

- [ ] **Step 3: Run the tests**

Run: `yarn eslint --fix client README.md 2>/dev/null; yarn vitest run client/App.test.tsx client/a11y.test.tsx client/components/page`
Expected: PASS. If axe reports a violation on an edit page (for example a toolbar or a file input without a name), fix the markup; don't disable rules.

- [ ] **Step 4: Check the editor stays out of the public bundle**

```bash
BUILD_CLIENT_SKIP_LINT=true yarn build:client > /tmp/claude-501/c1-build.log 2>&1; echo "build $?"
for f in build/client/assets/*.js; do grep -l 'prosemirror\|ProseMirror' "$f"; done | xargs -n1 basename | sort
grep -o '<script type="module"[^>]*src="[^"]*"' build/client/index.html
```

Expected:
- the build succeeds;
- only chunks named like `RichTextEditorImpl-*.js` (and Tiptap/ProseMirror vendor chunks imported only by it) contain ProseMirror;
- the entry chunk referenced by `index.html` is not in that list.

If the entry chunk contains ProseMirror, find the static import of `@tiptap/*` outside `RichTextEditorImpl.tsx` and remove it.

- [ ] **Step 5: README and full verification**

In `README.md`, replace the sentence "Editing still happens on the classic pages; signed-in editors see a 'Manage this on the classic site' link on each page." with:

```markdown
Editors add, edit and delete content in place: News, What's On, Info and Teams have their own edit pages (with a rich-text editor for articles, events and the club information); documents, programmes and seasons, sponsors, affiliations and gallery photos use dialogs. The Players and Users admin pages are still on the classic site.
```

Then run:

```bash
yarn lint
yarn typecheck
yarn test:client
yarn build:client
yarn test:server
```

Expected: all exit 0.

- [ ] **Step 6: Commit**

```bash
git add -A client README.md
git commit -q -m "Retire the classic-site editor links, extend accessibility checks, and document in-app editing" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```
