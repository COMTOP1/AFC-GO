# Public Pages Implementation Plan (Sub-project 4a)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port every public legacy page (Home, Teams, Team, News and article, What's On and event, Gallery, Documents, Programmes, Sponsors, Info, Contact) to the React client at `/app`. They're read-only, use the sub-project 3 design system, add four client-only upgrades, and the whole nav points into the SPA.

**Architecture:**
- One route component per page, in `client/pages/<section>/`.
- One data module per section, in `client/api/<section>.ts`, with hand-written types and TanStack Query hooks built on `apiFetch`.
- Shared page pieces in `client/components/page/`: cleaned rich text, a card grid with "Show more", tabs and search kept in the address, query states, and an editor link.
- Helpers in `client/lib/`.
- Every page except Home is loaded on first visit (`React.lazy`), behind a `Suspense` inside the layout.

**Tech Stack:** React 19, React Router 7, TanStack Query 5, TypeScript 6, Vite 8, Tailwind CSS 4, DOMPurify 3, Vitest 5 with Testing Library and jsdom, axe-core. Yarn 4 via corepack.

**Spec:** `docs/superpowers/specs/2026-09-29-public-pages-design.md`

## Global Constraints

- **Where to work:** the worktree `/Users/liam/Code/Go/AFC-design-system`, branch `public-pages`. Never touch `/Users/liam/Code/Go/AFC`, never read any `postgres_*.sql` file, and never `rm -rf` the current working directory.
- **Commits:** every commit message ends with a blank line and `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
- **Server:** no Go or server changes. The only new dependency is `dompurify`.
- **Styling:** use the sub-project 3 token classes only (`bg-surface`, `text-muted`, `border-line`, `text-red`, …), never raw hex. The exceptions are `black` and `white`, which already exist as tokens.
- **Imports:** there's no barrel file. Import each component from its own file, and keep one component (plus its own types) per `.tsx` file (`react-refresh/only-export-components`).
- **Page titles:** `"<Title> · AFC Aldermaston"`, or `"AFC Aldermaston"` for Home.
- **Dates:** `'date'` gives "28 Sep 2026" and `'dateTime'` gives "Fri 16 Oct 2026, 7pm" (minutes shown only when not :00, lower-case am/pm), always in the `Europe/London` time zone.
- **Addresses:**
  - `?period=future|past|all` (default `future`) on What's On;
  - `?season=<id>` on Programmes (absent = all seasons);
  - `?q=` on Documents and Programmes.
  - Search and season changes replace the current address entry; tab clicks add one.
- **"Show more":** pages of 12, with the label `Show more (N more)`.
- **Editor link:** text `Manage this on the classic site ↗`. It shows on `permissions.canEdit`, or on `permissions.canManageGallery` for the Gallery.
- **Contact:** the satnav notice reads "If you're using a satnav, use the postcode RG26 4QP — the postcode listed takes you some distance away."
- **Lint:** Prettier `printWidth` is 100 with single quotes. Run `yarn eslint --fix <files>` **before** running a task's tests, then commit.
- **Test helpers:** `mockFetch` keys include the query string, e.g. `'/api/v1/whatson?period=past'`. Unknown paths return 599.

## Review Focus

1. **A list item with no image or a broken image URL** (news, events, teams, sponsors, gallery, players, contacts): each shows its fallback (gradient, crest or name tile), never a broken image. Pinned by the tests in Tasks 5, 6, 10 and 11 that render an item without `imageUrl`.
2. **Stored HTML with scripts, event handlers, `javascript:` links or iframes:** these never reach the DOM. Links pointing off-site open in a new tab safely. Pinned in Task 2.
3. **A detail URL with a non-numeric or unknown ID** (`/news/abc`, `/team/999`): shows "Page not found" and doesn't loop or retry. Pinned in Tasks 6, 7 and 8.
4. **Switching What's On tabs or programme seasons after pressing "Show more":** the list resets to the first page, and the Back button restores the previous tab. Pinned in Tasks 3 and 8.
5. **Typing a search term with odd casing or extra spaces** ("  CLUB  const"): it still matches, and clearing it shows everything again. Pinned in Tasks 1 and 10.

---

## File map

| File | Task |
|---|---|
| `client/lib/format.ts`, `client/lib/text.ts`, `client/lib/ids.ts`, `client/components/page/usePageTitle.ts` (+ tests) | 1 |
| `package.json`, `yarn.lock`, `client/lib/sanitize.ts`, `client/components/page/RichText.tsx`, `client/styles/app.css` (+ tests) | 2 |
| `client/components/page/QueryState.tsx`, `CardGrid.tsx`, `TabsNav.tsx`, `useTabParam.ts`, `SearchInput.tsx`, `useSearchQuery.ts`, `EditorLink.tsx`, `client/lib/notFound.ts`, `client/pages/NotFoundPage.tsx` (+ tests) | 3 |
| `client/api/queries.ts`, `home.ts`, `teams.ts`, `news.ts`, `whatson.ts`, `gallery.ts`, `documents.ts`, `programmes.ts`, `sponsors.ts`, `pages.ts`, `client/test/fixtures.ts`, `client/test/render.tsx` (+ tests) | 4 |
| `client/components/page/LogoRow.tsx`, `client/pages/home/HomePage.tsx`; delete `client/pages/HomePage.tsx` and its test; `client/App.tsx` import (+ tests) | 5 |
| `client/components/page/LinkCard.tsx`, `client/pages/teams/TeamsPage.tsx`, `TeamPage.tsx` (+ tests) | 6 |
| `client/components/page/ArticleView.tsx`, `client/pages/news/NewsListPage.tsx`, `NewsArticlePage.tsx` (+ tests) | 7 |
| `client/pages/whatson/WhatsOnPage.tsx`, `EventPage.tsx` (+ tests) | 8 |
| `client/pages/gallery/GalleryPage.tsx`, `Lightbox.tsx` (+ tests) | 9 |
| `client/pages/documents/DocumentsPage.tsx`, `client/pages/programmes/ProgrammesPage.tsx`, `client/pages/sponsors/SponsorsPage.tsx` (+ tests) | 10 |
| `client/pages/info/InfoPage.tsx`, `fallback.tsx`, `client/pages/contact/ContactPage.tsx` (+ tests) | 11 |
| `client/App.tsx`, `client/components/layout/Layout.tsx`, `navItems.ts`, `NavBar.test.tsx`, `client/test/axe.ts`, `client/App.test.tsx`, `client/a11y.test.tsx`, `README.md` | 12 |

---

### Task 1: Formatting, text and ID helpers, and page titles

**Files:**
- Create: `client/lib/format.ts`, `client/lib/text.ts`, `client/lib/ids.ts`, `client/components/page/usePageTitle.ts`
- Test: `client/lib/helpers.test.ts`, `client/components/page/usePageTitle.test.tsx`

**Interfaces:**
- Produces:
  - `formatDate(iso: string, style?: 'date' | 'dateTime'): string` (an invalid date gives `''`)
  - `matchesQuery(text: string, q: string): boolean`
  - `parseId(raw: string | undefined): number | null`
  - `usePageTitle(title?: string): void`

- [ ] **Step 1: Write the failing tests**

`client/lib/helpers.test.ts`:

```ts
import { describe, expect, it } from 'vitest';

import { formatDate } from './format';
import { parseId } from './ids';
import { matchesQuery } from './text';

describe('formatDate', () => {
  it('formats a date in UK style', () => {
    expect(formatDate('2026-09-28T10:00:00Z')).toBe('28 Sep 2026');
  });

  it('formats a date and time in UK summer time, dropping :00', () => {
    expect(formatDate('2026-10-16T18:00:00Z', 'dateTime')).toBe('Fri 16 Oct 2026, 7pm');
  });

  it('keeps minutes and uses GMT in winter', () => {
    expect(formatDate('2026-12-19T19:30:00Z', 'dateTime')).toBe('Sat 19 Dec 2026, 7:30pm');
  });

  it('shows midnight as 12am', () => {
    expect(formatDate('2026-12-19T00:00:00Z', 'dateTime')).toBe('Sat 19 Dec 2026, 12am');
  });

  it('returns an empty string for an invalid date', () => {
    expect(formatDate('not a date')).toBe('');
  });
});

describe('matchesQuery', () => {
  it('matches every word, ignoring case and spacing', () => {
    expect(matchesQuery('Club constitution', '  CLUB  const')).toBe(true);
    expect(matchesQuery('Club constitution', 'club policy')).toBe(false);
  });

  it('matches everything for an empty query', () => {
    expect(matchesQuery('Anything', '')).toBe(true);
    expect(matchesQuery('Anything', '   ')).toBe(true);
  });
});

describe('parseId', () => {
  it('accepts positive integers', () => {
    expect(parseId('42')).toBe(42);
  });

  it('rejects everything else', () => {
    for (const raw of [undefined, '', 'abc', '0', '-1', '1.5', '12abc', '99999999999999999999']) {
      expect(parseId(raw)).toBeNull();
    }
  });
});
```

`client/components/page/usePageTitle.test.tsx`:

```tsx
import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { usePageTitle } from './usePageTitle';

function Probe({ title }: { title?: string }) {
  usePageTitle(title);
  return null;
}

describe('usePageTitle', () => {
  it('sets a page title with the club suffix', () => {
    render(<Probe title="News" />);
    expect(document.title).toBe('News · AFC Aldermaston');
  });

  it('uses the bare club name without a title', () => {
    render(<Probe />);
    expect(document.title).toBe('AFC Aldermaston');
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/lib client/components/page/usePageTitle.test.tsx`
Expected: FAIL. `./format`, `./ids`, `./text` and `./usePageTitle` can't be resolved.

- [ ] **Step 3: Implement**

`client/lib/format.ts`:

```ts
const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

const partsFormat = new Intl.DateTimeFormat('en-GB', {
  timeZone: 'Europe/London',
  weekday: 'short',
  day: 'numeric',
  month: 'numeric',
  year: 'numeric',
  hour: 'numeric',
  minute: '2-digit',
  hourCycle: 'h12',
});

/**
 * UK-style dates in the club's time zone: 'date' → "28 Sep 2026",
 * 'dateTime' → "Fri 16 Oct 2026, 7pm" (minutes only when not :00).
 * Month names are fixed so ICU's "Sept" never appears.
 */
export function formatDate(iso: string, style: 'date' | 'dateTime' = 'date'): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) {
    return '';
  }
  const p: Record<string, string> = {};
  for (const part of partsFormat.formatToParts(d)) {
    p[part.type] = part.value;
  }
  const date = `${p.day} ${MONTHS[Number(p.month) - 1]} ${p.year}`;
  if (style === 'date') {
    return date;
  }
  const minutes = p.minute === '00' ? '' : `:${p.minute}`;
  return `${p.weekday} ${date}, ${p.hour}${minutes}${(p.dayPeriod ?? '').toLowerCase()}`;
}
```

`client/lib/text.ts`:

```ts
/** True when every word of q appears in text, ignoring case and extra spaces. */
export function matchesQuery(text: string, q: string): boolean {
  const haystack = text.toLowerCase();
  return q
    .toLowerCase()
    .split(/\s+/)
    .filter(Boolean)
    .every((word) => haystack.includes(word));
}
```

`client/lib/ids.ts`:

```ts
/** A route :id as a positive integer, or null when it isn't one. */
export function parseId(raw: string | undefined): number | null {
  if (!raw || !/^[1-9]\d*$/.test(raw)) {
    return null;
  }
  const id = Number(raw);
  return Number.isSafeInteger(id) ? id : null;
}
```

`client/components/page/usePageTitle.ts`:

```ts
import { useEffect } from 'react';

const CLUB = 'AFC Aldermaston';

export function usePageTitle(title?: string): void {
  useEffect(() => {
    document.title = title ? `${title} · ${CLUB}` : CLUB;
  }, [title]);
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/lib client/components/page && yarn vitest run client/lib client/components/page/usePageTitle.test.tsx`
Expected: PASS, 11 tests.

If `formatDate` gives `PM` or a different hour format on this ICU, the `.toLowerCase()` and `hourCycle: 'h12'` should already cover it. Debug any remaining mismatch by printing `formatToParts`; don't weaken the assertions.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add client/lib client/components/page
git commit -q -m "Add date, search, id helpers and usePageTitle" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 2: Cleaned rich text

**Files:**
- Modify: `package.json`, `yarn.lock` (via `yarn add dompurify`), `client/styles/app.css`
- Create: `client/lib/sanitize.ts`, `client/components/page/RichText.tsx`
- Test: `client/lib/sanitize.test.ts`, `client/components/page/RichText.test.tsx`

**Interfaces:**
- Produces:
  - `cleanHtml(html: string): string`
  - `plainText(html: string, maxChars: number): string`
  - `RichText({ html: string; className?: string })`
  - the `.prose` CSS class

- [ ] **Step 1: Add the dependency**

```bash
cd /Users/liam/Code/Go/AFC-design-system && yarn add dompurify
```

Expected: `package.json` gains `"dompurify": "^3…"` under dependencies. DOMPurify 3 ships its own types.

- [ ] **Step 2: Write the failing tests**

`client/lib/sanitize.test.ts`:

```ts
import { describe, expect, it } from 'vitest';

import { cleanHtml, plainText } from './sanitize';

describe('cleanHtml', () => {
  it('removes scripts, handlers, javascript: links and iframes', () => {
    const out = cleanHtml(
      '<p onclick="steal()">Hi<script>bad()</script></p>' +
        '<a href="javascript:alert(1)">x</a><iframe src="https://evil.example"></iframe>' +
        '<img src="x" onerror="steal()">',
    );
    expect(out).not.toMatch(/script|onclick|onerror|javascript:|iframe/i);
    expect(out).toContain('<p>Hi</p>');
  });

  it('keeps the formatting the editor produces', () => {
    const html =
      '<h2>Title</h2><p><b>bold</b> <i>it</i> <u>u</u> <strong>s</strong> <em>e</em></p>' +
      '<ul><li>one</li></ul><ol><li>two</li></ol><blockquote>q</blockquote><a href="/news/1">in</a>';
    const out = cleanHtml(html);
    for (const tag of ['<h2>', '<b>', '<i>', '<u>', '<strong>', '<em>', '<ul>', '<ol>', '<li>', '<blockquote>']) {
      expect(out).toContain(tag);
    }
  });

  it('opens off-site links in a new tab, safely', () => {
    const out = cleanHtml('<a href="https://league.example/table">table</a>');
    expect(out).toContain('target="_blank"');
    expect(out).toContain('rel="noopener noreferrer"');
  });

  it('leaves same-site links alone', () => {
    const out = cleanHtml('<a href="/news/1">news</a>');
    expect(out).not.toContain('target=');
  });
});

describe('plainText', () => {
  it('strips tags and separates blocks with spaces', () => {
    expect(plainText('<p>A late <b>winner</b>.</p><p>Record crowd.</p>', 200)).toBe(
      'A late winner. Record crowd.',
    );
  });

  it('truncates at a word boundary with an ellipsis', () => {
    expect(plainText('<p>one two three four five</p>', 12)).toBe('one two…');
  });

  it('never runs scripts or keeps markup', () => {
    expect(plainText('<img src=x onerror="steal()"><p>safe</p>', 50)).toBe('safe');
  });
});
```

`client/components/page/RichText.test.tsx`:

```tsx
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { RichText } from './RichText';

describe('RichText', () => {
  it('renders cleaned HTML inside a prose wrapper', () => {
    const { container } = render(
      <RichText html={'<h2>Welcome</h2><p>Hello<script>bad()</script></p>'} />,
    );
    expect(screen.getByRole('heading', { level: 2, name: 'Welcome' })).toBeInTheDocument();
    expect(container.querySelector('script')).toBeNull();
    expect(container.firstElementChild).toHaveClass('prose');
  });
});
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `yarn vitest run client/lib/sanitize.test.ts client/components/page/RichText.test.tsx`
Expected: FAIL. `./sanitize` and `./RichText` can't be resolved.

- [ ] **Step 4: Implement**

`client/lib/sanitize.ts`:

```ts
import DOMPurify from 'dompurify';

// Off-site links in stored content open in a new tab without leaking the opener.
DOMPurify.addHook('afterSanitizeAttributes', (node) => {
  if (node.tagName !== 'A' || !node.hasAttribute('href')) {
    return;
  }
  try {
    const url = new URL(node.getAttribute('href') ?? '', window.location.href);
    if (url.origin !== window.location.origin) {
      node.setAttribute('target', '_blank');
      node.setAttribute('rel', 'noopener noreferrer');
    }
  } catch {
    // Unparseable href: leave it as DOMPurify cleaned it.
  }
});

/**
 * Cleans stored article/info HTML before it is rendered. The server only
 * sanitises on save, so older rows may contain anything.
 */
export function cleanHtml(html: string): string {
  return DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
    FORBID_TAGS: ['style', 'form', 'iframe'],
  });
}

const BLOCKS = 'p,div,br,li,h1,h2,h3,h4,h5,h6,blockquote,tr';

/** Cleaned HTML as plain text, cut at a word boundary to at most maxChars (plus "…"). */
export function plainText(html: string, maxChars: number): string {
  const doc = new DOMParser().parseFromString(cleanHtml(html), 'text/html');
  doc.body.querySelectorAll(BLOCKS).forEach((el) => el.after(' '));
  const text = (doc.body.textContent ?? '').replace(/\s+/g, ' ').trim();
  if (text.length <= maxChars) {
    return text;
  }
  const cut = text.slice(0, maxChars);
  const lastSpace = cut.lastIndexOf(' ');
  return `${(lastSpace > 0 ? cut.slice(0, lastSpace) : cut).replace(/[\s,.;:]+$/, '')}…`;
}
```

`client/components/page/RichText.tsx`:

```tsx
import { clsx } from 'clsx';
import { useMemo } from 'react';

import { cleanHtml } from '../../lib/sanitize';

export function RichText({ html, className }: { html: string; className?: string }) {
  const clean = useMemo(() => cleanHtml(html), [html]);
  return <div className={clsx('prose', className)} dangerouslySetInnerHTML={{ __html: clean }} />;
}
```

Append to `client/styles/app.css`, after the `@layer base { … }` block:

```css
@layer components {
  .prose {
    line-height: 1.6;
  }
  .prose > * + * {
    margin-top: 0.75rem;
  }
  .prose h1,
  .prose h2,
  .prose h3 {
    margin-top: 1.5rem;
    font-family: var(--font-display);
    font-weight: 800;
    line-height: 1.1;
    letter-spacing: 0.02em;
    text-transform: uppercase;
  }
  .prose h1 {
    font-size: 1.75rem;
  }
  .prose h2 {
    font-size: 1.5rem;
  }
  .prose h3 {
    font-size: 1.25rem;
  }
  .prose ul {
    list-style: disc;
    padding-left: 1.5rem;
  }
  .prose ol {
    list-style: decimal;
    padding-left: 1.5rem;
  }
  .prose a {
    color: var(--color-red);
    text-decoration: underline;
  }
  .prose img {
    max-width: 100%;
    border-radius: 0.5rem;
  }
  .prose blockquote {
    border-left: 4px solid var(--color-line);
    padding-left: 1rem;
    color: var(--color-muted);
  }
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `yarn eslint --fix client/lib client/components/page && yarn vitest run client/lib/sanitize.test.ts client/components/page/RichText.test.tsx`
Expected: PASS, 8 tests.

- [ ] **Step 6: Commit**

```bash
yarn typecheck && BUILD_CLIENT_SKIP_LINT=true yarn build:client > /dev/null
git add package.json yarn.lock client/lib client/components/page client/styles/app.css
git commit -q -m "Add DOMPurify-cleaned RichText, plainText and prose styles" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 3: Shared page pieces

**Files:**
- Create in `client/components/page/`: `QueryState.tsx`, `CardGrid.tsx`, `TabsNav.tsx`, `useTabParam.ts`, `SearchInput.tsx`, `useSearchQuery.ts`, `EditorLink.tsx`
- Create: `client/lib/notFound.ts`
- Modify: `client/pages/NotFoundPage.tsx` (adds `usePageTitle`)
- Test: `client/components/page/pieces.test.tsx`

**Interfaces:**
- Consumes: `Skeleton`, `Alert`, `Button`, `EmptyState`, `Field`, `Input` (sub-project 3); `ApiError` (`client/api/client.ts`); `useAuth`; `usePageTitle` (Task 1).
- Produces:
  - `QueryState<T>({ query: UseQueryResult<T>; children: (data: T) => ReactNode; isEmpty?: (data: T) => boolean; emptyTitle?: string; emptyMessage?: ReactNode })`
  - `PageSkeleton()`
  - `CardGrid<T>({ items: T[]; render: (item: T) => ReactNode; getKey: (item: T) => string | number; pageSize?: number; resetKey?: string; emptyTitle: string; emptyMessage?: ReactNode })`
  - `interface Tab { value: string; label: string }`
  - `TabsNav({ param: string; tabs: Tab[]; defaultValue: string; label: string })`
  - `useTabParam(param: string, values: string[], defaultValue: string): string`
  - `SearchInput({ label: string })`
  - `useSearchQuery(): string`
  - `EditorLink({ legacyHref: string; permission?: 'canEdit' | 'canManageGallery' })`
  - `isNotFound(error: unknown): boolean`

  Ruling vs the spec: `CardGrid` resets "Show more" when `resetKey` changes, not when `items` changes identity. Filtered arrays are new on every render, which would reset constantly or loop.

- [ ] **Step 1: Write the failing tests**

`client/components/page/pieces.test.tsx`:

```tsx
import { useQuery } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import { Route, Routes, useLocation, useNavigate } from 'react-router';
import { describe, expect, it, vi } from 'vitest';

import { ApiError } from '../../api/client';
import { isNotFound } from '../../lib/notFound';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import { CardGrid } from './CardGrid';
import { EditorLink } from './EditorLink';
import { QueryState } from './QueryState';
import { SearchInput } from './SearchInput';
import { TabsNav } from './TabsNav';
import { useSearchQuery } from './useSearchQuery';
import { useTabParam } from './useTabParam';

const anonymous = { status: 401, body: { error: { code: 401, message: 'login required' } } };

function Location() {
  const location = useLocation();
  return <output data-testid="location">{location.pathname + location.search}</output>;
}

describe('QueryState', () => {
  function Probe({ fn }: { fn: () => Promise<string[]> }) {
    const query = useQuery({ queryKey: ['probe'], queryFn: fn });
    return (
      <QueryState query={query} isEmpty={(d) => d.length === 0} emptyTitle="Nothing yet">
        {(data) => <p>{data.join(', ')}</p>}
      </QueryState>
    );
  }

  it('shows loading, then the data', async () => {
    mockFetch({ '/api/v1/auth/me': anonymous });
    renderWithProviders(<Probe fn={async () => ['a', 'b']} />);
    expect(screen.getByText('Loading')).toBeInTheDocument();
    expect(await screen.findByText('a, b')).toBeInTheDocument();
  });

  it('shows the empty state', async () => {
    mockFetch({ '/api/v1/auth/me': anonymous });
    renderWithProviders(<Probe fn={async () => []} />);
    expect(await screen.findByText('Nothing yet')).toBeInTheDocument();
  });

  it('shows the error and retries', async () => {
    mockFetch({ '/api/v1/auth/me': anonymous });
    const fn = vi
      .fn<() => Promise<string[]>>()
      .mockRejectedValueOnce(new ApiError(500, 'internal server error'))
      .mockResolvedValueOnce(['ok']);
    renderWithProviders(<Probe fn={fn} />);
    expect(await screen.findByRole('alert')).toHaveTextContent(
      "Couldn't load this: internal server error",
    );
    fireEvent.click(screen.getByRole('button', { name: 'Retry' }));
    expect(await screen.findByText('ok')).toBeInTheDocument();
  });
});

describe('CardGrid', () => {
  const items = Array.from({ length: 30 }, (_, i) => ({ id: i + 1 }));
  function Grid({ resetKey }: { resetKey?: string }) {
    return (
      <CardGrid
        items={items}
        getKey={(i) => i.id}
        render={(i) => <span>Item {i.id}</span>}
        emptyTitle="None"
        resetKey={resetKey}
      />
    );
  }

  it('shows 12, then 12 more, then the rest, then no button', () => {
    render(<Grid />);
    expect(screen.getAllByText(/^Item /)).toHaveLength(12);
    fireEvent.click(screen.getByRole('button', { name: 'Show more (18 more)' }));
    expect(screen.getAllByText(/^Item /)).toHaveLength(24);
    fireEvent.click(screen.getByRole('button', { name: 'Show more (6 more)' }));
    expect(screen.getAllByText(/^Item /)).toHaveLength(30);
    expect(screen.queryByRole('button', { name: /Show more/ })).toBeNull();
  });

  it('resets to the first page when resetKey changes', () => {
    const { rerender } = render(<Grid resetKey="future" />);
    fireEvent.click(screen.getByRole('button', { name: 'Show more (18 more)' }));
    expect(screen.getAllByText(/^Item /)).toHaveLength(24);
    rerender(<Grid resetKey="future" />);
    expect(screen.getAllByText(/^Item /)).toHaveLength(24);
    rerender(<Grid resetKey="past" />);
    expect(screen.getAllByText(/^Item /)).toHaveLength(12);
  });

  it('shows the empty state for no items', () => {
    render(<CardGrid items={[]} getKey={() => 1} render={() => null} emptyTitle="No news yet" />);
    expect(screen.getByText('No news yet')).toBeInTheDocument();
  });
});

describe('TabsNav and useTabParam', () => {
  const tabs = [
    { value: 'future', label: 'Upcoming' },
    { value: 'past', label: 'Past' },
    { value: 'all', label: 'All' },
  ];

  function Page() {
    const period = useTabParam('period', ['future', 'past', 'all'], 'future');
    const navigate = useNavigate();
    return (
      <>
        <TabsNav param="period" tabs={tabs} defaultValue="future" label="Events" />
        <p>Current: {period}</p>
        <button onClick={() => navigate(-1)}>Back</button>
        <Location />
      </>
    );
  }

  function renderAt(route: string) {
    mockFetch({ '/api/v1/auth/me': anonymous });
    renderWithProviders(
      <Routes>
        <Route path="/whatson" element={<Page />} />
      </Routes>,
      { route },
    );
  }

  it('reads the active tab from the address and marks it', () => {
    renderAt('/whatson?period=past');
    expect(screen.getByText('Current: past')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Past' })).toHaveAttribute('aria-current', 'page');
    expect(screen.getByRole('link', { name: 'Upcoming' })).not.toHaveAttribute('aria-current');
  });

  it('falls back to the default for unknown values', () => {
    renderAt('/whatson?period=nonsense');
    expect(screen.getByText('Current: future')).toBeInTheDocument();
  });

  it('updates the address on click and Back restores it', () => {
    renderAt('/whatson?q=x');
    fireEvent.click(screen.getByRole('link', { name: 'All' }));
    expect(screen.getByText('Current: all')).toBeInTheDocument();
    expect(screen.getByTestId('location')).toHaveTextContent('/whatson?q=x&period=all');
    fireEvent.click(screen.getByRole('button', { name: 'Back' }));
    expect(screen.getByText('Current: future')).toBeInTheDocument();
  });
});

describe('SearchInput and useSearchQuery', () => {
  function Page() {
    const q = useSearchQuery();
    return (
      <>
        <SearchInput label="Search documents" />
        <p>Query: [{q}]</p>
        <Location />
      </>
    );
  }

  it('starts from ?q= and writes changes back to the address', () => {
    mockFetch({ '/api/v1/auth/me': anonymous });
    renderWithProviders(
      <Routes>
        <Route path="/documents" element={<Page />} />
      </Routes>,
      { route: '/documents?q=club' },
    );
    const box = screen.getByLabelText('Search documents');
    expect(box).toHaveValue('club');
    fireEvent.change(box, { target: { value: 'policy' } });
    expect(screen.getByText('Query: [policy]')).toBeInTheDocument();
    expect(screen.getByTestId('location')).toHaveTextContent('/documents?q=policy');
    fireEvent.change(box, { target: { value: '' } });
    expect(screen.getByTestId('location')).toHaveTextContent(/^\/documents$/);
  });
});

describe('EditorLink', () => {
  function user(perms: { canEdit: boolean; canManageGallery: boolean }) {
    return {
      body: {
        id: 1,
        name: 'Someone',
        email: 's@example.test',
        role: 'Manager',
        permissions: { ...perms, canManageUsers: false },
      },
    };
  }

  it('is hidden when signed out', async () => {
    mockFetch({ '/api/v1/auth/me': anonymous });
    renderWithProviders(<EditorLink legacyHref="/news" />);
    await new Promise((r) => setTimeout(r, 0));
    expect(screen.queryByRole('link')).toBeNull();
  });

  it('is hidden for a user who cannot edit', async () => {
    mockFetch({ '/api/v1/auth/me': user({ canEdit: false, canManageGallery: false }) });
    renderWithProviders(<EditorLink legacyHref="/news" />);
    await new Promise((r) => setTimeout(r, 0));
    expect(screen.queryByRole('link')).toBeNull();
  });

  it('links editors to the classic page', async () => {
    mockFetch({ '/api/v1/auth/me': user({ canEdit: true, canManageGallery: true }) });
    renderWithProviders(<EditorLink legacyHref="/news" />);
    expect(
      await screen.findByRole('link', { name: 'Manage this on the classic site ↗' }),
    ).toHaveAttribute('href', '/news');
  });

  it('follows canManageGallery for the gallery', async () => {
    mockFetch({ '/api/v1/auth/me': user({ canEdit: false, canManageGallery: true }) });
    renderWithProviders(<EditorLink legacyHref="/gallery" permission="canManageGallery" />);
    expect(await screen.findByRole('link')).toHaveAttribute('href', '/gallery');
  });
});

describe('isNotFound', () => {
  it('is true only for a 404 ApiError', () => {
    expect(isNotFound(new ApiError(404, 'not found'))).toBe(true);
    expect(isNotFound(new ApiError(500, 'boom'))).toBe(false);
    expect(isNotFound(new Error('x'))).toBe(false);
    expect(isNotFound(null)).toBe(false);
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/components/page/pieces.test.tsx`
Expected: FAIL. `../../lib/notFound` can't be resolved.

- [ ] **Step 3: Implement**

`client/lib/notFound.ts`:

```ts
import { ApiError } from '../api/client';

export function isNotFound(error: unknown): boolean {
  return error instanceof ApiError && error.status === 404;
}
```

`client/components/page/QueryState.tsx`:

```tsx
import type { UseQueryResult } from '@tanstack/react-query';
import type { ReactNode } from 'react';

import { Alert } from '../ui/Alert';
import { Button } from '../ui/Button';
import { EmptyState } from '../ui/EmptyState';
import { Skeleton } from '../ui/Skeleton';

export function PageSkeleton() {
  return (
    <div className="space-y-3" aria-busy="true">
      <span className="sr-only">Loading</span>
      <Skeleton className="w-1/3" />
      <Skeleton />
      <Skeleton className="w-2/3" />
    </div>
  );
}

export interface QueryStateProps<T> {
  query: UseQueryResult<T>;
  children: (data: T) => ReactNode;
  isEmpty?: (data: T) => boolean;
  emptyTitle?: string;
  emptyMessage?: ReactNode;
}

/** One loading / error-with-retry / empty / content flow for every page. */
export function QueryState<T>({
  query,
  children,
  isEmpty,
  emptyTitle = 'Nothing here yet',
  emptyMessage,
}: QueryStateProps<T>) {
  if (query.isPending) {
    return <PageSkeleton />;
  }
  if (query.isError) {
    return (
      <Alert tone="error" className="flex flex-wrap items-center justify-between gap-3">
        <span>Couldn&apos;t load this: {query.error.message}</span>
        <Button size="sm" variant="secondary" onClick={() => void query.refetch()}>
          Retry
        </Button>
      </Alert>
    );
  }
  if (isEmpty?.(query.data)) {
    return <EmptyState title={emptyTitle} message={emptyMessage} />;
  }
  return <>{children(query.data)}</>;
}
```

`client/components/page/CardGrid.tsx`:

```tsx
import { useState, type ReactNode } from 'react';

import { Button } from '../ui/Button';
import { EmptyState } from '../ui/EmptyState';

export interface CardGridProps<T> {
  items: T[];
  render: (item: T) => ReactNode;
  getKey: (item: T) => string | number;
  /** Cards per "Show more" step; Infinity shows everything. */
  pageSize?: number;
  /** Changing this (e.g. a tab value) goes back to the first page. */
  resetKey?: string;
  emptyTitle: string;
  emptyMessage?: ReactNode;
}

export function CardGrid<T>({
  items,
  render,
  getKey,
  pageSize = 12,
  resetKey,
  emptyTitle,
  emptyMessage,
}: CardGridProps<T>) {
  const [shown, setShown] = useState(pageSize);
  const [lastKey, setLastKey] = useState(resetKey);
  if (resetKey !== lastKey) {
    setLastKey(resetKey);
    setShown(pageSize);
  }

  if (items.length === 0) {
    return <EmptyState title={emptyTitle} message={emptyMessage} />;
  }
  const remaining = items.length - shown;
  return (
    <>
      <ul className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {items.slice(0, shown).map((item) => (
          <li key={getKey(item)}>{render(item)}</li>
        ))}
      </ul>
      {remaining > 0 && (
        <div className="mt-6 text-center">
          <Button variant="secondary" onClick={() => setShown((s) => s + pageSize)}>
            Show more ({remaining} more)
          </Button>
        </div>
      )}
    </>
  );
}
```

`client/components/page/useTabParam.ts`:

```ts
import { useSearchParams } from 'react-router';

/** The current value of a tab-like search param, or the default when absent/unknown. */
export function useTabParam(param: string, values: string[], defaultValue: string): string {
  const [params] = useSearchParams();
  const value = params.get(param);
  return value !== null && values.includes(value) ? value : defaultValue;
}
```

`client/components/page/TabsNav.tsx`:

```tsx
import { clsx } from 'clsx';
import { Link, useSearchParams } from 'react-router';

import { useTabParam } from './useTabParam';

export interface Tab {
  value: string;
  label: string;
}

export interface TabsNavProps {
  param: string;
  tabs: Tab[];
  defaultValue: string;
  /** Accessible name of the tab navigation. */
  label: string;
}

export function TabsNav({ param, tabs, defaultValue, label }: TabsNavProps) {
  const [params] = useSearchParams();
  const current = useTabParam(
    param,
    tabs.map((t) => t.value),
    defaultValue,
  );

  return (
    <nav aria-label={label} className="mb-5 flex gap-1 border-b border-line">
      {tabs.map((tab) => {
        const next = new URLSearchParams(params);
        next.set(param, tab.value);
        const active = tab.value === current;
        return (
          <Link
            key={tab.value}
            to={`?${next.toString()}`}
            aria-current={active ? 'page' : undefined}
            className={clsx(
              '-mb-px border-b-3 px-3 py-2 text-sm font-semibold',
              active ? 'border-red text-red' : 'border-transparent text-muted hover:text-ink',
            )}
          >
            {tab.label}
          </Link>
        );
      })}
    </nav>
  );
}
```

`client/components/page/useSearchQuery.ts`:

```ts
import { useSearchParams } from 'react-router';

export function useSearchQuery(): string {
  const [params] = useSearchParams();
  return params.get('q') ?? '';
}
```

`client/components/page/SearchInput.tsx`:

```tsx
import { useSearchParams } from 'react-router';

import { Input } from '../ui/controls';
import { Field } from '../ui/Field';

/** A search box bound to ?q=; edits replace the history entry so Back skips keystrokes. */
export function SearchInput({ label }: { label: string }) {
  const [params, setParams] = useSearchParams();
  return (
    <Field label={label} className="max-w-md">
      <Input
        type="search"
        value={params.get('q') ?? ''}
        onChange={(e) => {
          const next = new URLSearchParams(params);
          if (e.target.value) {
            next.set('q', e.target.value);
          } else {
            next.delete('q');
          }
          setParams(next, { replace: true });
        }}
      />
    </Field>
  );
}
```

`client/components/page/EditorLink.tsx`:

```tsx
import { useAuth } from '../../auth/useAuth';

export interface EditorLinkProps {
  /** The legacy page whose edit controls cover this content. */
  legacyHref: string;
  permission?: 'canEdit' | 'canManageGallery';
}

/** Until sub-project 4c, editors manage content on the legacy pages. */
export function EditorLink({ legacyHref, permission = 'canEdit' }: EditorLinkProps) {
  const { user } = useAuth();
  if (!user?.permissions[permission]) {
    return null;
  }
  return (
    <a href={legacyHref} className="text-sm font-semibold text-red hover:underline">
      Manage this on the classic site ↗
    </a>
  );
}
```

`client/pages/NotFoundPage.tsx`: add the import `import { usePageTitle } from '../components/page/usePageTitle';` and call `usePageTitle('Page not found');` as the first line of the component body.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/components/page client/lib client/pages/NotFoundPage.tsx && yarn vitest run client/components/page && yarn test:client`
Expected: PASS. 15 tests in `pieces.test.tsx`, and the whole suite is green.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add client/components/page client/lib client/pages/NotFoundPage.tsx
git commit -q -m "Add QueryState, CardGrid, URL tabs, search input and editor link" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 4: Data modules and test fixtures

**Files:**
- Modify: `client/api/queries.ts` (the `queryKeys` additions), `client/test/render.tsx` (returns `queryClient`)
- Create in `client/api/`: `home.ts`, `teams.ts`, `news.ts`, `whatson.ts`, `gallery.ts`, `documents.ts`, `programmes.ts`, `sponsors.ts`, `pages.ts`
- Create: `client/test/fixtures.ts`
- Test: `client/api/sections.test.tsx`

**Interfaces:**
- Consumes:
  - `apiFetch<T>(path, { signal })`;
  - `TeamSummary`, `SiteInfo` (`client/api/types.ts`);
  - `useAuth()`;
  - `mockFetch`, `MockRoute`, `MockResponse` (`client/test/mockFetch.ts`).
- Produces the types:

  ```ts
  NewsArticle { id: number; title: string; content: string; date: string; imageUrl?: string }
  WhatsOnEvent { id: number; title: string; content: string; date: string; dateOfEvent: string; imageUrl?: string }
  type WhatsOnPeriod = 'future' | 'past' | 'all'
  Sponsor { id: number; name: string; website?: string; purpose?: string; team?: string; imageUrl?: string }
  Affiliation { id: number; name: string; website?: string; imageUrl?: string }
  HomeData { latestNews?: NewsArticle; nextEvent?: WhatsOnEvent; sponsors: Sponsor[]; affiliations: Affiliation[] }
  Manager { name: string; email: string }
  SquadMember { id: number; name: string; position?: string; isCaptain: boolean; imageUrl?: string }
  TeamDetail { team: TeamSummary; managers: Manager[]; sponsors: Sponsor[]; players: SquadMember[] }
  GalleryImage { id: number; caption?: string; imageUrl: string }
  ClubDocument { id: number; name: string; fileUrl: string }
  Season { id: number; name: string }
  Programme { id: number; name: string; date: string; fileUrl: string; season?: Season }
  InfoContent { content: string }
  ContactPerson { id: number; name: string; email: string; role: string; imageUrl?: string }
  ContactData { displayEmail?: string; people: ContactPerson[] }
  ```

- Produces the hooks:
  - `useHome()`
  - `useTeams()`: the key includes whether someone is signed in
  - `useTeam(id: number | null)`
  - `useNewsList()`, `useNewsArticle(id: number | null)`
  - `useWhatsOnList(period: WhatsOnPeriod)`, `useWhatsOnEvent(id: number | null)`
  - `useGallery()`, `useDocuments()`
  - `useProgrammes(seasonId: number)`: 0 means all seasons
  - `useSeasons()`, `useSponsors()`, `useInfo()`, `useContact()`
  - detail hooks are disabled when `id` is `null`
- Also produces:
  - `renderWithProviders(ui, { route?, queryClient? })`, which returns `{ ...renderResult, queryClient }`
  - fixtures in `client/test/fixtures.ts` (below)

- [ ] **Step 1: Check the Go types**

Before writing the TS types, read these and confirm every JSON tag matches the Interfaces block above (`omitempty` becomes `?`):
- `server/internal/news/types.go`, `whatson/types.go`, `sponsor/types.go`, `affiliation/types.go`
- `server/internal/player/types.go` (`Member`)
- `server/internal/image/types.go`, `document/types.go`, `programme/types.go`
- `server/internal/setting/types.go`
- `server/internal/site/types.go` (`Home`, `Contact`, `ContactPerson`, `TeamDetail`, `Manager`)

If a tag differs, the Go file wins. Record a ruling and use the Go name.

- [ ] **Step 2: Write the failing tests**

`client/api/sections.test.tsx`:

```tsx
import { screen, waitFor } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { anonymous, editor, events, programmes, teams } from '../test/fixtures';
import { mockFetch } from '../test/mockFetch';
import { renderWithProviders } from '../test/render';
import { useNewsArticle } from './news';
import { useProgrammes } from './programmes';
import { queryKeys } from './queries';
import { useTeams } from './teams';
import { useWhatsOnList } from './whatson';

describe('section hooks', () => {
  it('useWhatsOnList asks the API for the period', async () => {
    const fetchMock = mockFetch({
      '/api/v1/auth/me': anonymous,
      '/api/v1/whatson?period=past': { body: events },
    });
    function Probe() {
      const q = useWhatsOnList('past');
      return <p>{q.data ? `${q.data.length} events` : 'loading'}</p>;
    }
    renderWithProviders(<Probe />);
    expect(await screen.findByText(`${events.length} events`)).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([u]) => String(u) === '/api/v1/whatson?period=past')).toBe(
      true,
    );
  });

  it('useProgrammes omits season for all, and sends it otherwise', async () => {
    const fetchMock = mockFetch({
      '/api/v1/auth/me': anonymous,
      '/api/v1/programmes': { body: programmes },
      '/api/v1/programmes?season=2': { body: programmes.slice(0, 1) },
    });
    function Probe({ season }: { season: number }) {
      const q = useProgrammes(season);
      return <p>{q.data ? `season ${season}: ${q.data.length}` : 'loading'}</p>;
    }
    renderWithProviders(
      <>
        <Probe season={0} />
        <Probe season={2} />
      </>,
    );
    expect(await screen.findByText(`season 0: ${programmes.length}`)).toBeInTheDocument();
    expect(await screen.findByText('season 2: 1')).toBeInTheDocument();
    const urls = fetchMock.mock.calls.map(([u]) => String(u));
    expect(urls).toContain('/api/v1/programmes');
    expect(urls).toContain('/api/v1/programmes?season=2');
  });

  it('useTeams caches signed-in and anonymous lists separately', async () => {
    mockFetch({ '/api/v1/auth/me': editor, '/api/v1/teams': { body: teams } });
    function Probe() {
      const q = useTeams();
      return <p>{q.data ? `${q.data.length} teams` : 'loading'}</p>;
    }
    const { queryClient } = renderWithProviders(<Probe />);
    expect(await screen.findByText(`${teams.length} teams`)).toBeInTheDocument();
    await waitFor(() => expect(queryClient.getQueryData(queryKeys.teams(true))).toEqual(teams));
    expect(queryClient.getQueryData(queryKeys.teams(false))).toBeUndefined();
  });

  it('detail hooks do not fetch without an id', async () => {
    const fetchMock = mockFetch({ '/api/v1/auth/me': anonymous });
    function Probe() {
      const q = useNewsArticle(null);
      return <p>{q.fetchStatus}</p>;
    }
    renderWithProviders(<Probe />);
    expect(await screen.findByText('idle')).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([u]) => String(u).startsWith('/api/v1/news'))).toBe(false);
  });
});
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `yarn vitest run client/api/sections.test.tsx`
Expected: FAIL. `../test/fixtures` (and the modules) can't be resolved.

- [ ] **Step 4: Implement the query keys, the modules, the render helper and the fixtures**

Replace the `queryKeys` object in `client/api/queries.ts`:

```ts
export const queryKeys = {
  site: ['site'] as const,
  me: ['auth', 'me'] as const,
  home: ['home'] as const,
  // The server includes inactive teams for signed-in users.
  teams: (signedIn: boolean) => ['teams', { signedIn }] as const,
  team: (id: number) => ['team', id] as const,
  news: ['news'] as const,
  newsArticle: (id: number) => ['news', id] as const,
  whatson: (period: string) => ['whatson', period] as const,
  whatsonEvent: (id: number) => ['whatson', 'event', id] as const,
  gallery: ['gallery'] as const,
  documents: ['documents'] as const,
  programmes: (seasonId: number) => ['programmes', seasonId] as const,
  seasons: ['seasons'] as const,
  sponsors: ['sponsors'] as const,
  info: ['info'] as const,
  contact: ['contact'] as const,
};
```

`client/api/news.ts`:

```ts
import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';

/** news.Article — GET /news, /news/:id */
export interface NewsArticle {
  id: number;
  title: string;
  content: string;
  date: string;
  imageUrl?: string;
}

export function useNewsList() {
  return useQuery({
    queryKey: queryKeys.news,
    queryFn: ({ signal }) => apiFetch<NewsArticle[]>('/news', { signal }),
  });
}

export function useNewsArticle(id: number | null) {
  return useQuery({
    queryKey: queryKeys.newsArticle(id ?? 0),
    queryFn: ({ signal }) => apiFetch<NewsArticle>(`/news/${id}`, { signal }),
    enabled: id !== null,
  });
}
```

`client/api/whatson.ts`:

```ts
import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';

/** whatson.Event — GET /whatson, /whatson/:id */
export interface WhatsOnEvent {
  id: number;
  title: string;
  content: string;
  date: string;
  dateOfEvent: string;
  imageUrl?: string;
}

export type WhatsOnPeriod = 'future' | 'past' | 'all';

export function useWhatsOnList(period: WhatsOnPeriod) {
  return useQuery({
    queryKey: queryKeys.whatson(period),
    queryFn: ({ signal }) => apiFetch<WhatsOnEvent[]>(`/whatson?period=${period}`, { signal }),
  });
}

export function useWhatsOnEvent(id: number | null) {
  return useQuery({
    queryKey: queryKeys.whatsonEvent(id ?? 0),
    queryFn: ({ signal }) => apiFetch<WhatsOnEvent>(`/whatson/${id}`, { signal }),
    enabled: id !== null,
  });
}
```

`client/api/sponsors.ts`:

```ts
import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';

/** sponsor.Public — GET /sponsors */
export interface Sponsor {
  id: number;
  name: string;
  website?: string;
  purpose?: string;
  team?: string;
  imageUrl?: string;
}

export function useSponsors() {
  return useQuery({
    queryKey: queryKeys.sponsors,
    queryFn: ({ signal }) => apiFetch<Sponsor[]>('/sponsors', { signal }),
  });
}
```

`client/api/home.ts`:

```ts
import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import type { NewsArticle } from './news';
import { queryKeys } from './queries';
import type { Sponsor } from './sponsors';
import type { WhatsOnEvent } from './whatson';

/** affiliation.Public */
export interface Affiliation {
  id: number;
  name: string;
  website?: string;
  imageUrl?: string;
}

/** site.Home — GET /home. Panels whose data failed to load are omitted. */
export interface HomeData {
  latestNews?: NewsArticle;
  nextEvent?: WhatsOnEvent;
  sponsors: Sponsor[];
  affiliations: Affiliation[];
}

export function useHome() {
  return useQuery({
    queryKey: queryKeys.home,
    queryFn: ({ signal }) => apiFetch<HomeData>('/home', { signal }),
  });
}
```

`client/api/teams.ts`:

```ts
import { useQuery } from '@tanstack/react-query';

import { useAuth } from '../auth/useAuth';
import { apiFetch } from './client';
import { queryKeys } from './queries';
import type { Sponsor } from './sponsors';
import type { TeamSummary } from './types';

/** site.Manager */
export interface Manager {
  name: string;
  email: string;
}

/** player.Member — youth teams never return any. */
export interface SquadMember {
  id: number;
  name: string;
  position?: string;
  isCaptain: boolean;
  imageUrl?: string;
}

/** site.TeamDetail — GET /teams/:id */
export interface TeamDetail {
  team: TeamSummary;
  managers: Manager[];
  sponsors: Sponsor[];
  players: SquadMember[];
}

export function useTeams() {
  const { user } = useAuth();
  return useQuery({
    queryKey: queryKeys.teams(user !== null),
    queryFn: ({ signal }) => apiFetch<TeamSummary[]>('/teams', { signal }),
  });
}

export function useTeam(id: number | null) {
  return useQuery({
    queryKey: queryKeys.team(id ?? 0),
    queryFn: ({ signal }) => apiFetch<TeamDetail>(`/teams/${id}`, { signal }),
    enabled: id !== null,
  });
}
```

`client/api/gallery.ts`:

```ts
import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';

/** image.Public — GET /gallery */
export interface GalleryImage {
  id: number;
  caption?: string;
  imageUrl: string;
}

export function useGallery() {
  return useQuery({
    queryKey: queryKeys.gallery,
    queryFn: ({ signal }) => apiFetch<GalleryImage[]>('/gallery', { signal }),
  });
}
```

`client/api/documents.ts`:

```ts
import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';

/** document.Public — GET /documents */
export interface ClubDocument {
  id: number;
  name: string;
  fileUrl: string;
}

export function useDocuments() {
  return useQuery({
    queryKey: queryKeys.documents,
    queryFn: ({ signal }) => apiFetch<ClubDocument[]>('/documents', { signal }),
  });
}
```

`client/api/programmes.ts`:

```ts
import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';

/** programme.PublicSeason — GET /seasons */
export interface Season {
  id: number;
  name: string;
}

/** programme.Public — GET /programmes */
export interface Programme {
  id: number;
  name: string;
  date: string;
  fileUrl: string;
  season?: Season;
}

/** seasonId 0 means every season (the API's default when the param is absent). */
export function useProgrammes(seasonId: number) {
  return useQuery({
    queryKey: queryKeys.programmes(seasonId),
    queryFn: ({ signal }) =>
      apiFetch<Programme[]>(seasonId ? `/programmes?season=${seasonId}` : '/programmes', {
        signal,
      }),
  });
}

export function useSeasons() {
  return useQuery({
    queryKey: queryKeys.seasons,
    queryFn: ({ signal }) => apiFetch<Season[]>('/seasons', { signal }),
  });
}
```

`client/api/pages.ts`:

```ts
import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';

/** setting.InfoContent — GET /info */
export interface InfoContent {
  content: string;
}

/** site.ContactPerson */
export interface ContactPerson {
  id: number;
  name: string;
  email: string;
  role: string;
  imageUrl?: string;
}

/** site.Contact — GET /contact */
export interface ContactData {
  displayEmail?: string;
  people: ContactPerson[];
}

export function useInfo() {
  return useQuery({
    queryKey: queryKeys.info,
    queryFn: ({ signal }) => apiFetch<InfoContent>('/info', { signal }),
  });
}

export function useContact() {
  return useQuery({
    queryKey: queryKeys.contact,
    queryFn: ({ signal }) => apiFetch<ContactData>('/contact', { signal }),
  });
}
```

Replace `client/test/render.tsx` so callers can reach the query cache:

```tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render } from '@testing-library/react';
import type { ReactElement } from 'react';
import { MemoryRouter } from 'react-router';

import { AuthProvider } from '../auth/AuthProvider';
import { ToastProvider } from '../components/ui/toast/ToastProvider';
import { ThemeProvider } from '../theme/ThemeProvider';

export interface RenderOptions {
  route?: string;
  queryClient?: QueryClient;
}

/** Renders ui inside the same providers as main.tsx, with retries off. */
export function renderWithProviders(
  ui: ReactElement,
  {
    route = '/',
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false, gcTime: Infinity } },
    }),
  }: RenderOptions = {},
) {
  const wrap = (node: ReactElement) => (
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[route]}>
        <ThemeProvider>
          <AuthProvider>
            <ToastProvider>{node}</ToastProvider>
          </AuthProvider>
        </ThemeProvider>
      </MemoryRouter>
    </QueryClientProvider>
  );
  const result = render(wrap(ui));
  return { ...result, queryClient, rerender: (next: ReactElement) => result.rerender(wrap(next)) };
}
```

The `rerender` override keeps the providers when a test rerenders, so state survives in later tests that rerender a wrapped tree.

`client/test/fixtures.ts`:

```ts
import type { ClubDocument } from '../api/documents';
import type { GalleryImage } from '../api/gallery';
import type { Affiliation, HomeData } from '../api/home';
import type { NewsArticle } from '../api/news';
import type { ContactData, InfoContent } from '../api/pages';
import type { Programme, Season } from '../api/programmes';
import type { Sponsor } from '../api/sponsors';
import type { TeamDetail } from '../api/teams';
import type { SiteInfo, TeamSummary } from '../api/types';
import type { WhatsOnEvent } from '../api/whatson';
import type { MockResponse, MockRoute } from './mockFetch';

export const anonymous: MockResponse = {
  status: 401,
  body: { error: { code: 401, message: 'login required' } },
};

function signedIn(name: string, role: string, canEdit: boolean, canManageGallery: boolean) {
  return {
    body: {
      id: 1,
      name,
      email: 'someone@example.test',
      role,
      permissions: { canEdit, canManageGallery, canManageUsers: false },
    },
  };
}

export const editor: MockResponse = signedIn('Ed Editor', 'Treasurer', true, true);
export const manager: MockResponse = signedIn('Mo Manager', 'Manager', false, false);

export const newsArticle: NewsArticle = {
  id: 1,
  title: 'First team win the cup',
  content: '<p>A late winner at <b>Aldermaston</b> in front of a record crowd.</p>',
  date: '2026-09-28T10:00:00Z',
  imageUrl: '/api/v1/files/news/1',
};
export const newsNoImage: NewsArticle = {
  id: 2,
  title: 'AGM announced',
  content: '<p>All members welcome.</p>',
  date: '2026-09-20T10:00:00Z',
};
export const news: NewsArticle[] = [newsArticle, newsNoImage];

export const event: WhatsOnEvent = {
  id: 3,
  title: 'Presentation evening',
  content: '<p>Clubhouse, all welcome.</p>',
  date: '2026-09-01T10:00:00Z',
  dateOfEvent: '2026-10-16T18:00:00Z',
};
export const events: WhatsOnEvent[] = [event];

export const sponsor: Sponsor = {
  id: 4,
  name: 'Acme Ltd',
  website: 'https://acme.example',
  purpose: 'Kit sponsor',
  team: 'First Team',
  imageUrl: '/api/v1/files/sponsor/4',
};
export const sponsorPlain: Sponsor = { id: 5, name: 'Corner Shop' };
export const sponsors: Sponsor[] = [sponsor, sponsorPlain];

export const affiliation: Affiliation = { id: 6, name: 'The FA', website: 'https://www.thefa.com' };

export const team: TeamSummary = {
  id: 7,
  name: 'First Team',
  league: 'Thames Valley Premier',
  division: 'Division 1',
  isActive: true,
  isYouth: false,
  ages: 99,
};
export const youthTeam: TeamSummary = {
  id: 8,
  name: 'Under 12s',
  isActive: true,
  isYouth: true,
  ages: 12,
};
export const teams: TeamSummary[] = [team, youthTeam];

export const teamDetail: TeamDetail = {
  team: {
    ...team,
    description: 'Our senior side.',
    coach: 'Sam Patel',
    physio: 'Alex Lee',
    leagueTableUrl: 'https://league.example/table',
    fixturesUrl: 'https://league.example/fixtures',
  },
  managers: [{ name: 'Jo Smith', email: 'jo@example.test' }],
  sponsors: [sponsor],
  players: [
    { id: 9, name: 'Chris Captain', position: 'Defender', isCaptain: true },
    { id: 10, name: 'Pat Player', position: 'Forward', isCaptain: false, imageUrl: '/p/10' },
  ],
};
export const youthTeamDetail: TeamDetail = {
  team: youthTeam,
  managers: [],
  sponsors: [],
  players: [],
};

export const galleryImages: GalleryImage[] = [
  { id: 11, caption: 'Cup final', imageUrl: '/api/v1/files/gallery/11' },
  { id: 12, imageUrl: '/api/v1/files/gallery/12' },
  { id: 13, caption: 'Presentation', imageUrl: '/api/v1/files/gallery/13' },
];

export const documents: ClubDocument[] = [
  { id: 14, name: 'Club constitution', fileUrl: '/api/v1/files/document/14' },
  { id: 15, name: 'Safeguarding policy', fileUrl: '/api/v1/files/document/15' },
];

export const seasons: Season[] = [
  { id: 1, name: '2025-26' },
  { id: 2, name: '2026-27' },
];
export const programmes: Programme[] = [
  {
    id: 16,
    name: 'vs Downton',
    date: '2026-09-05T13:00:00Z',
    fileUrl: '/api/v1/files/programme/16',
    season: seasons[1],
  },
  {
    id: 17,
    name: 'vs Marlow',
    date: '2026-03-01T15:00:00Z',
    fileUrl: '/api/v1/files/programme/17',
    season: seasons[0],
  },
  { id: 18, name: 'Pre-season friendly', date: '2025-07-20T13:00:00Z', fileUrl: '/api/v1/files/programme/18' },
];

export const info: InfoContent = { content: '<h2>Welcome</h2><p>All about the club.</p>' };

export const contact: ContactData = {
  people: [
    { id: 19, name: 'Sam Sec', email: 'sam@example.test', role: 'Club Secretary' },
    { id: 20, name: 'Cara Chair', email: 'cara@example.test', role: 'Chairperson', imageUrl: '/c/20' },
  ],
};

export const home: HomeData = {
  latestNews: newsArticle,
  nextEvent: event,
  sponsors,
  affiliations: [affiliation],
};

export const site: SiteInfo = { year: 2026, visitorCount: 42, version: 'test', teams };

/** Every public endpoint with fixture data; override entries per test. */
export function publicRoutes(overrides: Record<string, MockRoute> = {}): Record<string, MockRoute> {
  return {
    '/api/v1/site': { body: site },
    '/api/v1/auth/me': anonymous,
    '/api/v1/home': { body: home },
    '/api/v1/teams': { body: teams },
    [`/api/v1/teams/${team.id}`]: { body: teamDetail },
    [`/api/v1/teams/${youthTeam.id}`]: { body: youthTeamDetail },
    '/api/v1/news': { body: news },
    [`/api/v1/news/${newsArticle.id}`]: { body: newsArticle },
    '/api/v1/whatson?period=future': { body: events },
    '/api/v1/whatson?period=past': { body: [] },
    '/api/v1/whatson?period=all': { body: events },
    [`/api/v1/whatson/${event.id}`]: { body: event },
    '/api/v1/gallery': { body: galleryImages },
    '/api/v1/documents': { body: documents },
    '/api/v1/programmes': { body: programmes },
    '/api/v1/seasons': { body: seasons },
    '/api/v1/sponsors': { body: sponsors },
    '/api/v1/info': { body: info },
    '/api/v1/contact': { body: contact },
    ...overrides,
  };
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `yarn eslint --fix client/api client/test && yarn vitest run client/api && yarn test:client`
Expected: PASS. The 4 new hook tests pass, and the whole suite is green.

- [ ] **Step 6: Commit**

```bash
yarn typecheck
git add client/api client/test
git commit -q -m "Add public section data modules, query keys and test fixtures" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 5: Home page

**Files:**
- Create: `client/components/page/LogoRow.tsx`, `client/pages/home/HomePage.tsx`, `client/pages/home/HomePage.test.tsx`
- Delete: `client/pages/HomePage.tsx`, `client/pages/HomePage.test.tsx`
- Modify: `client/App.tsx` (the HomePage import path only)

**Interfaces:**
- Consumes:
  - `useHome`, `HomeData`, `Affiliation` (Task 4);
  - `QueryState`, `usePageTitle` (Tasks 1 and 3);
  - `plainText` (Task 2);
  - `formatDate` (Task 1);
  - `Card`, `CardBody`, `CardMedia`;
  - fixtures (Task 4).
- Produces:
  - `interface LogoItem { id: number; name: string; website?: string; imageUrl?: string }`
  - `LogoRow({ title: string; items: LogoItem[] })`, which renders nothing when `items` is empty
  - the default export `HomePage`

- [ ] **Step 1: Write the failing tests**

`client/pages/home/HomePage.test.tsx`:

```tsx
import { screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { event, home, newsArticle, publicRoutes } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import HomePage from './HomePage';

describe('HomePage', () => {
  it('shows the news hero, next event, sponsors and affiliations', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<HomePage />);
    expect(screen.getByRole('heading', { level: 1, name: 'AFC Aldermaston' })).toBeInTheDocument();
    const heroLink = await screen.findByRole('link', { name: newsArticle.title });
    expect(heroLink).toHaveAttribute('href', `/news/${newsArticle.id}`);
    expect(
      screen.getByText('A late winner at Aldermaston in front of a record crowd.'),
    ).toBeInTheDocument();
    expect(screen.getByRole('link', { name: event.title })).toHaveAttribute(
      'href',
      `/whatson/${event.id}`,
    );
    expect(screen.getByText('Fri 16 Oct 2026, 7pm')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'All events →' })).toHaveAttribute('href', '/whatson');

    const sponsors = screen.getByRole('region', { name: 'Our sponsors' });
    expect(within(sponsors).getByRole('link', { name: 'Acme Ltd' })).toHaveAttribute(
      'href',
      'https://acme.example',
    );
    expect(within(sponsors).getByText('Corner Shop')).toBeInTheDocument();
    expect(screen.getByRole('region', { name: 'Affiliations' })).toBeInTheDocument();
    expect(document.title).toBe('AFC Aldermaston');
  });

  it('lets the event span the row when there is no news', async () => {
    mockFetch(publicRoutes({ '/api/v1/home': { body: { ...home, latestNews: undefined } } }));
    renderWithProviders(<HomePage />);
    expect(await screen.findByRole('link', { name: event.title })).toBeInTheDocument();
    expect(screen.queryByText('Latest news')).toBeNull();
  });

  it('skips the hero row and empty logo rows when there is nothing to show', async () => {
    mockFetch(
      publicRoutes({ '/api/v1/home': { body: { sponsors: [], affiliations: [] } } }),
    );
    renderWithProviders(<HomePage />);
    await screen.findByRole('heading', { level: 1 });
    await new Promise((r) => setTimeout(r, 0));
    expect(screen.queryByText('Latest news')).toBeNull();
    expect(screen.queryByText('Next event')).toBeNull();
    expect(screen.queryByRole('region', { name: 'Our sponsors' })).toBeNull();
    expect(screen.queryByRole('region', { name: 'Affiliations' })).toBeNull();
  });

  it('shows a retryable error when /home fails', async () => {
    mockFetch(
      publicRoutes({
        '/api/v1/home': { status: 500, body: { error: { code: 500, message: 'boom' } } },
      }),
    );
    renderWithProviders(<HomePage />);
    expect(await screen.findByRole('alert')).toHaveTextContent("Couldn't load this: boom");
    expect(screen.getByRole('button', { name: 'Retry' })).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/home`
Expected: FAIL. `./HomePage` can't be resolved.

- [ ] **Step 3: Implement**

`client/components/page/LogoRow.tsx`:

```tsx
import { useId } from 'react';

export interface LogoItem {
  id: number;
  name: string;
  website?: string;
  imageUrl?: string;
}

function Logo({ item }: { item: LogoItem }) {
  const inner = item.imageUrl ? (
    <img src={item.imageUrl} alt={item.name} loading="lazy" className="max-h-14 w-auto" />
  ) : (
    <span className="text-center text-sm font-semibold">{item.name}</span>
  );
  const box =
    'flex h-20 w-36 items-center justify-center rounded-lg border border-line bg-white p-2 text-black';
  if (item.website) {
    return (
      <a href={item.website} target="_blank" rel="noopener noreferrer" className={box}>
        {inner}
      </a>
    );
  }
  return <div className={box}>{inner}</div>;
}

/** A titled row of sponsor/affiliation logos; renders nothing when empty. */
export function LogoRow({ title, items }: { title: string; items: LogoItem[] }) {
  const headingId = useId();
  if (items.length === 0) {
    return null;
  }
  return (
    <section aria-labelledby={headingId}>
      <h2
        id={headingId}
        className="mb-3 font-display text-2xl font-extrabold tracking-wide uppercase"
      >
        {title}
      </h2>
      <ul className="flex flex-wrap gap-3">
        {items.map((item) => (
          <li key={item.id}>
            <Logo item={item} />
          </li>
        ))}
      </ul>
    </section>
  );
}
```

`client/pages/home/HomePage.tsx`:

```tsx
import { clsx } from 'clsx';
import { Link } from 'react-router';

import type { HomeData } from '../../api/home';
import { useHome } from '../../api/home';
import type { NewsArticle } from '../../api/news';
import type { WhatsOnEvent } from '../../api/whatson';
import { LogoRow } from '../../components/page/LogoRow';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Card, CardBody, CardMedia } from '../../components/ui/Card';
import { formatDate } from '../../lib/format';
import { plainText } from '../../lib/sanitize';

const kicker = 'text-xs font-bold tracking-widest uppercase';

function NewsHero({ article, className }: { article: NewsArticle; className?: string }) {
  const href = `/news/${article.id}`;
  return (
    <Card className={className}>
      <Link to={href} className="relative block">
        <CardMedia src={article.imageUrl} alt="" />
        <div className="absolute inset-x-0 bottom-0 bg-linear-to-t from-black/75 to-transparent p-4 text-white">
          <p className={kicker}>Latest news</p>
          <h2 className="font-display text-2xl leading-none font-extrabold uppercase md:text-3xl">
            {article.title}
          </h2>
        </div>
      </Link>
      <CardBody>
        <p className="text-muted">{plainText(article.content, 180)}</p>
        <Link to={href} className="mt-2 inline-block font-semibold text-red" tabIndex={-1}>
          Read more →
        </Link>
      </CardBody>
    </Card>
  );
}

function NextEvent({ event }: { event: WhatsOnEvent }) {
  return (
    <Card>
      <CardBody className="flex h-full flex-col gap-2">
        <p className={clsx(kicker, 'text-red')}>Next event</p>
        <h2 className="font-display text-2xl leading-none font-extrabold uppercase">
          <Link to={`/whatson/${event.id}`} className="hover:text-red">
            {event.title}
          </Link>
        </h2>
        <p className="text-sm font-semibold">{formatDate(event.dateOfEvent, 'dateTime')}</p>
        <p className="text-sm text-muted">{plainText(event.content, 120)}</p>
        <Link to="/whatson" className="mt-auto font-semibold text-red">
          All events →
        </Link>
      </CardBody>
    </Card>
  );
}

function HomeContent({ data }: { data: HomeData }) {
  const { latestNews, nextEvent, sponsors, affiliations } = data;
  return (
    <div className="space-y-10">
      {(latestNews || nextEvent) && (
        <div className={clsx('grid gap-4', latestNews && nextEvent && 'md:grid-cols-3')}>
          {latestNews && (
            <NewsHero article={latestNews} className={nextEvent ? 'md:col-span-2' : undefined} />
          )}
          {nextEvent && <NextEvent event={nextEvent} />}
        </div>
      )}
      <LogoRow title="Our sponsors" items={sponsors} />
      <LogoRow title="Affiliations" items={affiliations} />
    </div>
  );
}

export default function HomePage() {
  usePageTitle();
  const home = useHome();
  return (
    <>
      {/* The masthead shows the club name; the page still needs its own h1. */}
      <h1 className="sr-only">AFC Aldermaston</h1>
      <QueryState query={home}>{(data) => <HomeContent data={data} />}</QueryState>
    </>
  );
}
```

Ruling vs the spec ("shows only the club name as the title"): the `h1` is visually hidden, because the masthead already shows the name in large type.

The "Read more →" link has `tabIndex={-1}` because it duplicates the hero link, which already carries the title as its name. Keyboard users get one stop per article.

Remove the old proof page and repoint the App import:

```bash
git rm -q client/pages/HomePage.tsx client/pages/HomePage.test.tsx
```

In `client/App.tsx`, change `import HomePage from './pages/HomePage';` to `import HomePage from './pages/home/HomePage';`.

`client/App.test.tsx`'s `renderAt` mocks only `/site` and `/auth/me`, so `/home` gets a 599 there. That's fine: the `h1` renders outside `QueryState`, and the App tests only check the h1, the layout and the not-found page. They're rewritten in Task 12.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/pages client/components/page client/App.tsx && yarn vitest run client/pages/home && yarn test:client`
Expected: PASS. 4 Home tests pass, and the suite is green. The a11y test's "home" cases now render the new Home page with a 599 `/home` error, which is fine until Task 12 gives them full fixtures.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add -A client/pages client/components/page client/App.tsx
git commit -q -m "Port the home page: news hero, next event, sponsors and affiliations" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 6: Teams and Team pages

**Files:**
- Create: `client/components/page/LinkCard.tsx`, `client/pages/teams/TeamsPage.tsx`, `client/pages/teams/TeamPage.tsx`
- Test: `client/pages/teams/teams.test.tsx`

**Interfaces:**
- Consumes:
  - `useTeams`, `useTeam`, `TeamDetail`, `SquadMember` (Task 4);
  - `QueryState`, `CardGrid`, `EditorLink`, `usePageTitle`, `isNotFound`, `parseId`, `LogoRow` (Tasks 1, 3 and 5);
  - `NotFoundPage`;
  - `PageHeader`, `Badge`, `ButtonLink`, `CardMedia`;
  - `crest` from `client/assets/crest.png`.
- Produces:
  - `LinkCard({ to: string; imageUrl?: string; title: string; meta?: ReactNode; children?: ReactNode })`
  - default exports `TeamsPage` and `TeamPage`

- [ ] **Step 1: Write the failing tests**

`client/pages/teams/teams.test.tsx`:

```tsx
import { screen, within } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import {
  editor,
  publicRoutes,
  team,
  teamDetail,
  teams,
  youthTeam,
} from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import TeamPage from './TeamPage';
import TeamsPage from './TeamsPage';

function renderTeam(route: string, overrides = {}) {
  const fetchMock = mockFetch(publicRoutes(overrides));
  renderWithProviders(
    <Routes>
      <Route path="/team/:id" element={<TeamPage />} />
    </Routes>,
    { route },
  );
  return fetchMock;
}

describe('TeamsPage', () => {
  it('lists teams as cards with league and badge', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<TeamsPage />);
    const card = await screen.findByRole('link', { name: /First Team/ });
    expect(card).toHaveAttribute('href', `/team/${team.id}`);
    expect(within(card).getByText('Thames Valley Premier · Division 1')).toBeInTheDocument();
    expect(within(card).getByText('Adult')).toBeInTheDocument();
    expect(
      within(screen.getByRole('link', { name: /Under 12s/ })).getByText('Youth'),
    ).toBeInTheDocument();
    expect(document.title).toBe('Teams · AFC Aldermaston');
  });

  it('marks inactive teams (returned to signed-in users)', async () => {
    mockFetch(
      publicRoutes({
        '/api/v1/auth/me': editor,
        '/api/v1/teams': { body: [...teams, { ...team, id: 99, name: 'Vets', isActive: false }] },
      }),
    );
    renderWithProviders(<TeamsPage />);
    expect(
      within(await screen.findByRole('link', { name: /Vets/ })).getByText('Inactive'),
    ).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Manage this on the classic site ↗' })).toHaveAttribute(
      'href',
      '/teams',
    );
  });

  it('shows the empty state', async () => {
    mockFetch(publicRoutes({ '/api/v1/teams': { body: [] } }));
    renderWithProviders(<TeamsPage />);
    expect(await screen.findByText('No teams yet')).toBeInTheDocument();
  });
});

describe('TeamPage', () => {
  it('shows details, links, managers, squad and sponsors', async () => {
    renderTeam(`/team/${team.id}`);
    expect(await screen.findByRole('heading', { level: 1, name: 'First Team' })).toBeInTheDocument();
    expect(screen.getByText('Our senior side.')).toBeInTheDocument();
    expect(screen.getByText('Sam Patel')).toBeInTheDocument();
    expect(screen.getByText('Alex Lee')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'jo@example.test' })).toHaveAttribute(
      'href',
      'mailto:jo@example.test',
    );
    expect(screen.getByRole('link', { name: 'League table ↗' })).toHaveAttribute(
      'href',
      teamDetail.team.leagueTableUrl,
    );
    expect(screen.getByRole('link', { name: 'Fixtures ↗' })).toHaveAttribute(
      'href',
      teamDetail.team.fixturesUrl,
    );
    const squad = screen.getByRole('region', { name: 'Squad' });
    expect(within(squad).getByText('Chris Captain')).toBeInTheDocument();
    expect(within(squad).getByText('Captain')).toBeInTheDocument();
    // No photo → the crest stands in.
    expect(within(squad).getAllByRole('img')[0]).toHaveAttribute('src', expect.stringContaining('crest'));
    expect(screen.getByRole('region', { name: 'Team sponsors' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Teams' })).toHaveAttribute('href', '/teams');
    expect(document.title).toBe('First Team · AFC Aldermaston');
  });

  it('hides the squad entirely for youth teams and omits missing links', async () => {
    renderTeam(`/team/${youthTeam.id}`);
    expect(await screen.findByRole('heading', { level: 1, name: 'Under 12s' })).toBeInTheDocument();
    expect(screen.queryByRole('region', { name: 'Squad' })).toBeNull();
    expect(screen.queryByText(/squad/i)).toBeNull();
    expect(screen.queryByRole('link', { name: 'League table ↗' })).toBeNull();
    expect(screen.queryByRole('link', { name: 'Fixtures ↗' })).toBeNull();
    expect(screen.queryByRole('region', { name: 'Team sponsors' })).toBeNull();
  });

  it('shows not found for an unknown team', async () => {
    renderTeam('/team/999', {
      '/api/v1/teams/999': { status: 404, body: { error: { code: 404, message: 'not found' } } },
    });
    expect(await screen.findByRole('heading', { name: 'Page not found' })).toBeInTheDocument();
  });

  it('shows not found for a malformed id without calling the API', async () => {
    const fetchMock = renderTeam('/team/abc');
    expect(await screen.findByRole('heading', { name: 'Page not found' })).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([u]) => String(u).startsWith('/api/v1/teams'))).toBe(false);
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/teams`
Expected: FAIL. `./TeamPage` and `./TeamsPage` can't be resolved.

- [ ] **Step 3: Implement**

`client/components/page/LinkCard.tsx`:

```tsx
import type { ReactNode } from 'react';
import { Link } from 'react-router';

import { Card, CardBody, CardMedia } from '../ui/Card';

export interface LinkCardProps {
  to: string;
  imageUrl?: string;
  title: string;
  meta?: ReactNode;
  children?: ReactNode;
}

/** An image card that links to a detail page; the gradient stands in for a missing image. */
export function LinkCard({ to, imageUrl, title, meta, children }: LinkCardProps) {
  return (
    <Card className="h-full transition-shadow hover:shadow-md">
      <Link to={to} className="block h-full">
        <CardMedia src={imageUrl} alt="" />
        <CardBody className="space-y-1">
          <h2 className="font-display text-xl leading-tight font-extrabold uppercase">{title}</h2>
          {meta && <div className="text-sm text-muted">{meta}</div>}
          {children}
        </CardBody>
      </Link>
    </Card>
  );
}
```

`client/pages/teams/TeamsPage.tsx`:

```tsx
import { useTeams } from '../../api/teams';
import type { TeamSummary } from '../../api/types';
import { CardGrid } from '../../components/page/CardGrid';
import { EditorLink } from '../../components/page/EditorLink';
import { LinkCard } from '../../components/page/LinkCard';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Badge } from '../../components/ui/Badge';
import { PageHeader } from '../../components/ui/PageHeader';

function leagueLine(team: TeamSummary): string | undefined {
  const parts = [team.league, team.division].filter(Boolean);
  return parts.length ? parts.join(' · ') : undefined;
}

export default function TeamsPage() {
  usePageTitle('Teams');
  const teams = useTeams();
  return (
    <>
      <PageHeader title="Teams" actions={<EditorLink legacyHref="/teams" />} />
      <QueryState query={teams}>
        {(list) => (
          <CardGrid
            items={list}
            getKey={(t) => t.id}
            pageSize={Number.POSITIVE_INFINITY}
            emptyTitle="No teams yet"
            render={(t) => (
              <LinkCard to={`/team/${t.id}`} imageUrl={t.imageUrl} title={t.name} meta={leagueLine(t)}>
                <div className="flex flex-wrap gap-2 pt-1">
                  <Badge tone={t.isYouth ? 'red' : 'blue'}>{t.isYouth ? 'Youth' : 'Adult'}</Badge>
                  {!t.isActive && <Badge>Inactive</Badge>}
                </div>
              </LinkCard>
            )}
          />
        )}
      </QueryState>
    </>
  );
}
```

`client/pages/teams/TeamPage.tsx`:

```tsx
import { useId } from 'react';
import { Link, useParams } from 'react-router';

import { useTeam, type SquadMember, type TeamDetail } from '../../api/teams';
import crest from '../../assets/crest.png';
import { LogoRow } from '../../components/page/LogoRow';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Badge } from '../../components/ui/Badge';
import { ButtonLink } from '../../components/ui/ButtonLink';
import { CardMedia } from '../../components/ui/Card';
import { PageHeader } from '../../components/ui/PageHeader';
import { parseId } from '../../lib/ids';
import { isNotFound } from '../../lib/notFound';
import NotFoundPage from '../NotFoundPage';

function Player({ player }: { player: SquadMember }) {
  return (
    <li className="text-center">
      <img
        src={player.imageUrl || crest}
        alt=""
        loading="lazy"
        className="mx-auto mb-2 size-20 rounded-full border border-line bg-white object-cover"
      />
      <p className="font-semibold">{player.name}</p>
      {player.position && <p className="text-sm text-muted">{player.position}</p>}
      {player.isCaptain && <Badge tone="blue">Captain</Badge>}
    </li>
  );
}

function Squad({ players }: { players: SquadMember[] }) {
  const headingId = useId();
  return (
    <section aria-labelledby={headingId} className="mt-10">
      <h2 id={headingId} className="mb-4 font-display text-2xl font-extrabold tracking-wide uppercase">
        Squad
      </h2>
      <ul className="grid grid-cols-2 gap-6 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
        {players.map((p) => (
          <Player key={p.id} player={p} />
        ))}
      </ul>
    </section>
  );
}

function TeamContent({ detail }: { detail: TeamDetail }) {
  const { team, managers, sponsors, players } = detail;
  usePageTitle(team.name);
  const facts: [string, string | undefined][] = [
    ['League', team.league],
    ['Division', team.division],
    ['Coach', team.coach],
    ['Physio', team.physio],
  ];
  return (
    <>
      <nav aria-label="Breadcrumb" className="mb-2 text-sm text-muted">
        <Link to="/teams" className="font-semibold text-red">
          Teams
        </Link>{' '}
        / {team.name}
      </nav>
      <PageHeader title={team.name} subtitle={team.description} />
      <div className="grid gap-6 md:grid-cols-2">
        <div className="overflow-hidden rounded-lg border border-line">
          <CardMedia src={team.imageUrl} alt={`${team.name} team photo`} />
        </div>
        <div className="space-y-4">
          <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 text-sm">
            {facts
              .filter(([, value]) => value)
              .map(([label, value]) => (
                <div key={label} className="contents">
                  <dt className="text-muted">{label}</dt>
                  <dd>{value}</dd>
                </div>
              ))}
            {managers.length > 0 && (
              <div className="contents">
                <dt className="text-muted">{managers.length > 1 ? 'Managers' : 'Manager'}</dt>
                <dd className="space-y-1">
                  {managers.map((m) => (
                    <p key={m.email}>
                      {m.name} (
                      <a href={`mailto:${m.email}`} className="text-red underline">
                        {m.email}
                      </a>
                      )
                    </p>
                  ))}
                </dd>
              </div>
            )}
          </dl>
          <div className="flex flex-wrap gap-2">
            {team.leagueTableUrl && (
              <ButtonLink href={team.leagueTableUrl} variant="secondary" target="_blank" rel="noopener noreferrer">
                League table ↗
              </ButtonLink>
            )}
            {team.fixturesUrl && (
              <ButtonLink href={team.fixturesUrl} variant="secondary" target="_blank" rel="noopener noreferrer">
                Fixtures ↗
              </ButtonLink>
            )}
          </div>
        </div>
      </div>
      {/* The API returns no players for youth teams; the whole section is then omitted. */}
      {players.length > 0 && <Squad players={players} />}
      <div className="mt-10">
        <LogoRow title="Team sponsors" items={sponsors} />
      </div>
    </>
  );
}

export default function TeamPage() {
  const id = parseId(useParams().id);
  const team = useTeam(id);
  if (id === null || isNotFound(team.error)) {
    return <NotFoundPage />;
  }
  return <QueryState query={team}>{(detail) => <TeamContent detail={detail} />}</QueryState>;
}
```

`usePageTitle` in `TeamContent` runs only once the data has loaded. Before that, the tab keeps the previous title, which is acceptable.

`<div className="contents">` wrapping a `dt`/`dd` pair is valid HTML (`div` groups are allowed inside `dl`). If axe flags `definition-list`, render the `dt` and `dd` as keyed fragments instead, and record a ruling.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/pages/teams client/components/page && yarn vitest run client/pages/teams`
Expected: PASS, 7 tests.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add client/pages/teams client/components/page/LinkCard.tsx
git commit -q -m "Port the teams list and team page (squad hidden for youth teams)" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 7: News list and article

**Files:**
- Create: `client/components/page/ArticleView.tsx`, `client/pages/news/NewsListPage.tsx`, `client/pages/news/NewsArticlePage.tsx`
- Test: `client/pages/news/news.test.tsx`

**Interfaces:**
- Consumes:
  - `useNewsList`, `useNewsArticle` (Task 4);
  - `LinkCard` (Task 6);
  - `CardGrid`, `QueryState`, `EditorLink`, `usePageTitle`, `isNotFound`, `parseId`, `RichText`, `formatDate`;
  - `NotFoundPage`, `PageHeader`, `CardMedia`, `ButtonLink`.
- Produces: `ArticleView({ section: string; sectionHref: string; crumb: string; title: string; imageUrl?: string; html: string; subtitle?: ReactNode; backLabel: string; editorHref: string })`, plus default exports `NewsListPage` and `NewsArticlePage`.

- [ ] **Step 1: Write the failing tests**

`client/pages/news/news.test.tsx`:

```tsx
import { screen } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import type { NewsArticle } from '../../api/news';
import { editor, newsArticle, newsNoImage, publicRoutes } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import NewsArticlePage from './NewsArticlePage';
import NewsListPage from './NewsListPage';

function renderArticle(route: string, overrides = {}) {
  const fetchMock = mockFetch(publicRoutes(overrides));
  renderWithProviders(
    <Routes>
      <Route path="/news/:id" element={<NewsArticlePage />} />
    </Routes>,
    { route },
  );
  return fetchMock;
}

describe('NewsListPage', () => {
  it('lists articles newest first with dates, falling back to the gradient', async () => {
    mockFetch(publicRoutes());
    const { container } = renderWithProviders(<NewsListPage />);
    const first = await screen.findByRole('link', { name: new RegExp(newsArticle.title) });
    expect(first).toHaveAttribute('href', `/news/${newsArticle.id}`);
    expect(screen.getByText('28 Sep 2026')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: new RegExp(newsNoImage.title) })).toBeInTheDocument();
    expect(container.querySelectorAll('[data-fallback]')).toHaveLength(1);
    expect(document.title).toBe('News · AFC Aldermaston');
  });

  it('pages long lists with Show more', async () => {
    const many: NewsArticle[] = Array.from({ length: 14 }, (_, i) => ({
      ...newsNoImage,
      id: 100 + i,
      title: `Story ${i + 1}`,
    }));
    mockFetch(publicRoutes({ '/api/v1/news': { body: many } }));
    renderWithProviders(<NewsListPage />);
    expect(await screen.findByRole('button', { name: 'Show more (2 more)' })).toBeInTheDocument();
    expect(screen.getAllByRole('link', { name: /Story/ })).toHaveLength(12);
  });

  it('shows the empty state and the editor link', async () => {
    mockFetch(publicRoutes({ '/api/v1/news': { body: [] }, '/api/v1/auth/me': editor }));
    renderWithProviders(<NewsListPage />);
    expect(await screen.findByText('No news yet')).toBeInTheDocument();
    expect(
      await screen.findByRole('link', { name: 'Manage this on the classic site ↗' }),
    ).toHaveAttribute('href', '/news');
  });
});

describe('NewsArticlePage', () => {
  it('renders the article with cleaned content', async () => {
    renderArticle(`/news/${newsArticle.id}`);
    expect(
      await screen.findByRole('heading', { level: 1, name: newsArticle.title }),
    ).toBeInTheDocument();
    expect(screen.getByText('Aldermaston').tagName).toBe('B');
    expect(screen.getByRole('link', { name: 'News' })).toHaveAttribute('href', '/news');
    expect(screen.getByText(/28 Sep 2026/)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: '← All news' })).toHaveAttribute('href', '/news');
    expect(document.title).toBe(`${newsArticle.title} · AFC Aldermaston`);
  });

  it('shows not found for a missing article', async () => {
    renderArticle('/news/404', {
      '/api/v1/news/404': { status: 404, body: { error: { code: 404, message: 'not found' } } },
    });
    expect(await screen.findByRole('heading', { name: 'Page not found' })).toBeInTheDocument();
  });

  it('shows not found for a malformed id without calling the API', async () => {
    const fetchMock = renderArticle('/news/abc');
    expect(await screen.findByRole('heading', { name: 'Page not found' })).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([u]) => String(u).startsWith('/api/v1/news'))).toBe(false);
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/news`
Expected: FAIL. The page modules can't be resolved.

- [ ] **Step 3: Implement**

`client/components/page/ArticleView.tsx`:

```tsx
import type { ReactNode } from 'react';
import { Link } from 'react-router';

import { ButtonLink } from '../ui/ButtonLink';
import { CardMedia } from '../ui/Card';
import { PageHeader } from '../ui/PageHeader';
import { EditorLink } from './EditorLink';
import { RichText } from './RichText';
import { usePageTitle } from './usePageTitle';

export interface ArticleViewProps {
  section: string;
  sectionHref: string;
  /** Text after the section in the breadcrumb, e.g. the date. */
  crumb: string;
  title: string;
  imageUrl?: string;
  html: string;
  subtitle?: ReactNode;
  backLabel: string;
  editorHref: string;
}

/** A news article or event: image, breadcrumb, title, cleaned body, back link. */
export function ArticleView({
  section,
  sectionHref,
  crumb,
  title,
  imageUrl,
  html,
  subtitle,
  backLabel,
  editorHref,
}: ArticleViewProps) {
  usePageTitle(title);
  return (
    <article className="mx-auto max-w-3xl">
      <div className="mb-6 overflow-hidden rounded-lg border border-line">
        <CardMedia src={imageUrl} alt="" aspect="21 / 9" />
      </div>
      <nav aria-label="Breadcrumb" className="mb-2 text-sm text-muted">
        <Link to={sectionHref} className="font-semibold text-red">
          {section}
        </Link>{' '}
        / {crumb}
      </nav>
      <PageHeader title={title} subtitle={subtitle} actions={<EditorLink legacyHref={editorHref} />} />
      <RichText html={html} />
      <div className="mt-8">
        <ButtonLink to={sectionHref} variant="secondary">
          {backLabel}
        </ButtonLink>
      </div>
    </article>
  );
}
```

`client/pages/news/NewsListPage.tsx`:

```tsx
import { useNewsList } from '../../api/news';
import { CardGrid } from '../../components/page/CardGrid';
import { EditorLink } from '../../components/page/EditorLink';
import { LinkCard } from '../../components/page/LinkCard';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { PageHeader } from '../../components/ui/PageHeader';
import { formatDate } from '../../lib/format';

export default function NewsListPage() {
  usePageTitle('News');
  const news = useNewsList();
  return (
    <>
      <PageHeader title="News" actions={<EditorLink legacyHref="/news" />} />
      <QueryState query={news}>
        {(list) => (
          <CardGrid
            items={list}
            getKey={(a) => a.id}
            emptyTitle="No news yet"
            render={(a) => (
              <LinkCard
                to={`/news/${a.id}`}
                imageUrl={a.imageUrl}
                title={a.title}
                meta={formatDate(a.date)}
              />
            )}
          />
        )}
      </QueryState>
    </>
  );
}
```

`client/pages/news/NewsArticlePage.tsx`:

```tsx
import { useParams } from 'react-router';

import { useNewsArticle } from '../../api/news';
import { ArticleView } from '../../components/page/ArticleView';
import { QueryState } from '../../components/page/QueryState';
import { formatDate } from '../../lib/format';
import { parseId } from '../../lib/ids';
import { isNotFound } from '../../lib/notFound';
import NotFoundPage from '../NotFoundPage';

export default function NewsArticlePage() {
  const id = parseId(useParams().id);
  const article = useNewsArticle(id);
  if (id === null || isNotFound(article.error)) {
    return <NotFoundPage />;
  }
  return (
    <QueryState query={article}>
      {(a) => (
        <ArticleView
          section="News"
          sectionHref="/news"
          crumb={formatDate(a.date)}
          title={a.title}
          imageUrl={a.imageUrl}
          html={a.content}
          backLabel="← All news"
          editorHref={`/news/${a.id}`}
        />
      )}
    </QueryState>
  );
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/pages/news client/components/page && yarn vitest run client/pages/news`
Expected: PASS, 6 tests.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add client/pages/news client/components/page/ArticleView.tsx
git commit -q -m "Port the news list and article pages" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 8: What's On list and event

**Files:**
- Create: `client/pages/whatson/WhatsOnPage.tsx`, `client/pages/whatson/EventPage.tsx`
- Test: `client/pages/whatson/whatson.test.tsx`

**Interfaces:**
- Consumes:
  - `useWhatsOnList(period)`, `useWhatsOnEvent(id)`, `WhatsOnPeriod` (Task 4);
  - `TabsNav`, `useTabParam`, `CardGrid` (with `resetKey`), `LinkCard`, `ArticleView`, `QueryState`, `EditorLink`, `usePageTitle`;
  - `formatDate`, `parseId`, `isNotFound`, `NotFoundPage`, `PageHeader`.
- Produces: default exports `WhatsOnPage` and `EventPage`.

- [ ] **Step 1: Write the failing tests**

`client/pages/whatson/whatson.test.tsx`:

```tsx
import { fireEvent, screen } from '@testing-library/react';
import { Route, Routes, useNavigate } from 'react-router';
import { describe, expect, it } from 'vitest';

import type { WhatsOnEvent } from '../../api/whatson';
import { event, publicRoutes } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import EventPage from './EventPage';
import WhatsOnPage from './WhatsOnPage';

function BackButton() {
  const navigate = useNavigate();
  return <button onClick={() => navigate(-1)}>Go back</button>;
}

function renderList(route = '/whatson', overrides = {}) {
  const fetchMock = mockFetch(publicRoutes(overrides));
  renderWithProviders(
    <Routes>
      <Route
        path="/whatson"
        element={
          <>
            <WhatsOnPage />
            <BackButton />
          </>
        }
      />
    </Routes>,
    { route },
  );
  return fetchMock;
}

describe('WhatsOnPage', () => {
  it('defaults to upcoming events', async () => {
    const fetchMock = renderList();
    expect(await screen.findByRole('link', { name: new RegExp(event.title) })).toHaveAttribute(
      'href',
      `/whatson/${event.id}`,
    );
    expect(screen.getByText('Fri 16 Oct 2026, 7pm')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Upcoming' })).toHaveAttribute('aria-current', 'page');
    expect(fetchMock.mock.calls.some(([u]) => String(u) === '/api/v1/whatson?period=future')).toBe(
      true,
    );
    expect(document.title).toBe("What's On · AFC Aldermaston");
  });

  it('switches tabs, requests that period and Back restores the tab', async () => {
    const fetchMock = renderList();
    await screen.findByRole('link', { name: new RegExp(event.title) });
    fireEvent.click(screen.getByRole('link', { name: 'Past' }));
    expect(await screen.findByText('No past events')).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([u]) => String(u) === '/api/v1/whatson?period=past')).toBe(
      true,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Go back' }));
    expect(await screen.findByRole('link', { name: new RegExp(event.title) })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Upcoming' })).toHaveAttribute('aria-current', 'page');
  });

  it('goes back to the first page of cards when the tab changes', async () => {
    const many: WhatsOnEvent[] = Array.from({ length: 20 }, (_, i) => ({
      ...event,
      id: 200 + i,
      title: `Event ${i + 1}`,
    }));
    renderList('/whatson?period=all', {
      '/api/v1/whatson?period=all': { body: many },
      '/api/v1/whatson?period=future': { body: many },
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Show more (8 more)' }));
    expect(screen.getAllByRole('link', { name: /^Event / })).toHaveLength(20);
    fireEvent.click(screen.getByRole('link', { name: 'Upcoming' }));
    expect(await screen.findByRole('button', { name: 'Show more (8 more)' })).toBeInTheDocument();
    expect(screen.getAllByRole('link', { name: /^Event / })).toHaveLength(12);
  });

  it('uses a period-specific empty title', async () => {
    renderList('/whatson?period=all', { '/api/v1/whatson?period=all': { body: [] } });
    expect(await screen.findByText('No events yet')).toBeInTheDocument();
  });
});

describe('EventPage', () => {
  function renderEvent(route: string, overrides = {}) {
    const fetchMock = mockFetch(publicRoutes(overrides));
    renderWithProviders(
      <Routes>
        <Route path="/whatson/:id" element={<EventPage />} />
      </Routes>,
      { route },
    );
    return fetchMock;
  }

  it('renders the event with its date and time', async () => {
    renderEvent(`/whatson/${event.id}`);
    expect(await screen.findByRole('heading', { level: 1, name: event.title })).toBeInTheDocument();
    expect(screen.getAllByText('Fri 16 Oct 2026, 7pm').length).toBeGreaterThan(0);
    expect(screen.getByText('Clubhouse, all welcome.')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: '← All events' })).toHaveAttribute('href', '/whatson');
  });

  it('shows not found for a missing or malformed event', async () => {
    renderEvent('/whatson/404', {
      '/api/v1/whatson/404': { status: 404, body: { error: { code: 404, message: 'not found' } } },
    });
    expect(await screen.findByRole('heading', { name: 'Page not found' })).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/whatson`
Expected: FAIL. The page modules can't be resolved.

- [ ] **Step 3: Implement**

`client/pages/whatson/WhatsOnPage.tsx`:

```tsx
import { useWhatsOnList, type WhatsOnPeriod } from '../../api/whatson';
import { CardGrid } from '../../components/page/CardGrid';
import { EditorLink } from '../../components/page/EditorLink';
import { LinkCard } from '../../components/page/LinkCard';
import { QueryState } from '../../components/page/QueryState';
import { TabsNav } from '../../components/page/TabsNav';
import { usePageTitle } from '../../components/page/usePageTitle';
import { useTabParam } from '../../components/page/useTabParam';
import { PageHeader } from '../../components/ui/PageHeader';
import { formatDate } from '../../lib/format';

const tabs: { value: WhatsOnPeriod; label: string }[] = [
  { value: 'future', label: 'Upcoming' },
  { value: 'past', label: 'Past' },
  { value: 'all', label: 'All' },
];

const emptyTitles: Record<WhatsOnPeriod, string> = {
  future: 'No upcoming events',
  past: 'No past events',
  all: 'No events yet',
};

export default function WhatsOnPage() {
  usePageTitle("What's On");
  const period = useTabParam(
    'period',
    tabs.map((t) => t.value),
    'future',
  ) as WhatsOnPeriod;
  const events = useWhatsOnList(period);
  return (
    <>
      <PageHeader title="What's On" actions={<EditorLink legacyHref="/whatson" />} />
      <TabsNav param="period" tabs={tabs} defaultValue="future" label="Event period" />
      <QueryState query={events}>
        {(list) => (
          <CardGrid
            items={list}
            getKey={(e) => e.id}
            resetKey={period}
            emptyTitle={emptyTitles[period]}
            render={(e) => (
              <LinkCard
                to={`/whatson/${e.id}`}
                imageUrl={e.imageUrl}
                title={e.title}
                meta={formatDate(e.dateOfEvent, 'dateTime')}
              />
            )}
          />
        )}
      </QueryState>
    </>
  );
}
```

`client/pages/whatson/EventPage.tsx`:

```tsx
import { useParams } from 'react-router';

import { useWhatsOnEvent } from '../../api/whatson';
import { ArticleView } from '../../components/page/ArticleView';
import { QueryState } from '../../components/page/QueryState';
import { formatDate } from '../../lib/format';
import { parseId } from '../../lib/ids';
import { isNotFound } from '../../lib/notFound';
import NotFoundPage from '../NotFoundPage';

export default function EventPage() {
  const id = parseId(useParams().id);
  const event = useWhatsOnEvent(id);
  if (id === null || isNotFound(event.error)) {
    return <NotFoundPage />;
  }
  return (
    <QueryState query={event}>
      {(e) => (
        <ArticleView
          section="What's On"
          sectionHref="/whatson"
          crumb={formatDate(e.dateOfEvent)}
          title={e.title}
          subtitle={formatDate(e.dateOfEvent, 'dateTime')}
          imageUrl={e.imageUrl}
          html={e.content}
          backLabel="← All events"
          editorHref={`/whatson/${e.id}`}
        />
      )}
    </QueryState>
  );
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/pages/whatson && yarn vitest run client/pages/whatson`
Expected: PASS, 6 tests.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add client/pages/whatson
git commit -q -m "Port What's On with period tabs in the URL, and the event page" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 9: Gallery and lightbox

**Files:**
- Create: `client/pages/gallery/GalleryPage.tsx`, `client/pages/gallery/Lightbox.tsx`
- Test: `client/pages/gallery/gallery.test.tsx`

**Interfaces:**
- Consumes:
  - `useGallery`, `GalleryImage` (Task 4);
  - `Modal`, `Button` (sub-project 3);
  - `QueryState`, `EditorLink`, `usePageTitle`, `PageHeader`.
- Produces:
  - `Lightbox({ images: GalleryImage[]; index: number | null; onIndexChange: (i: number) => void; onClose: () => void })`
  - the default export `GalleryPage`

- [ ] **Step 1: Write the failing tests**

`client/pages/gallery/gallery.test.tsx`:

```tsx
import { act, fireEvent, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { editor, galleryImages, manager, publicRoutes } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import GalleryPage from './GalleryPage';

async function renderGallery(overrides = {}) {
  mockFetch(publicRoutes(overrides));
  renderWithProviders(<GalleryPage />);
  return screen.findByRole('button', { name: 'Cup final' });
}

function swipe(el: Element, fromX: number, toX: number) {
  const start = new Event('touchstart', { bubbles: true });
  Object.defineProperty(start, 'touches', { value: [{ clientX: fromX }] });
  const end = new Event('touchend', { bubbles: true });
  Object.defineProperty(end, 'changedTouches', { value: [{ clientX: toX }] });
  act(() => {
    el.dispatchEvent(start);
    el.dispatchEvent(end);
  });
}

describe('GalleryPage', () => {
  it('shows thumbnails named by caption, or by position', async () => {
    await renderGallery();
    expect(screen.getByRole('button', { name: 'Photo 2 of 3' })).toBeInTheDocument();
    expect(document.title).toBe('Gallery · AFC Aldermaston');
  });

  it('opens the viewer at the clicked photo with a counter and caption', async () => {
    const thumb = await renderGallery();
    fireEvent.click(thumb);
    const dialog = screen.getByRole('dialog');
    expect(within(dialog).getByText('1 / 3')).toBeInTheDocument();
    expect(within(dialog).getByRole('img', { name: 'Cup final' })).toHaveAttribute(
      'src',
      galleryImages[0].imageUrl,
    );
  });

  it('moves with the buttons and wraps around', async () => {
    fireEvent.click(await renderGallery());
    const dialog = screen.getByRole('dialog');
    fireEvent.click(within(dialog).getByRole('button', { name: 'Previous photo' }));
    expect(within(dialog).getByText('3 / 3')).toBeInTheDocument();
    fireEvent.click(within(dialog).getByRole('button', { name: 'Next photo' }));
    expect(within(dialog).getByText('1 / 3')).toBeInTheDocument();
  });

  it('moves with the arrow keys', async () => {
    fireEvent.click(await renderGallery());
    fireEvent.keyDown(document, { key: 'ArrowRight' });
    expect(screen.getByText('2 / 3')).toBeInTheDocument();
    fireEvent.keyDown(document, { key: 'ArrowLeft' });
    expect(screen.getByText('1 / 3')).toBeInTheDocument();
  });

  it('moves with a swipe of 50px or more, but not a small drag', async () => {
    fireEvent.click(await renderGallery());
    const stage = screen.getByTestId('lightbox-stage');
    swipe(stage, 300, 280);
    expect(screen.getByText('1 / 3')).toBeInTheDocument();
    swipe(stage, 300, 200);
    expect(screen.getByText('2 / 3')).toBeInTheDocument();
    swipe(stage, 100, 200);
    expect(screen.getByText('1 / 3')).toBeInTheDocument();
  });

  it('closes on Esc and returns focus to the thumbnail', async () => {
    const thumb = await renderGallery();
    thumb.focus();
    fireEvent.click(thumb);
    fireEvent(screen.getByRole('dialog'), new Event('cancel', { cancelable: true }));
    expect(screen.queryByText('1 / 3')).toBeNull();
    expect(document.activeElement).toBe(thumb);
  });

  it('closes with the Close button', async () => {
    fireEvent.click(await renderGallery());
    fireEvent.click(screen.getByRole('button', { name: 'Close' }));
    expect(screen.queryByText('1 / 3')).toBeNull();
  });

  it('shows the empty state', async () => {
    mockFetch(publicRoutes({ '/api/v1/gallery': { body: [] } }));
    renderWithProviders(<GalleryPage />);
    expect(await screen.findByText('No photos yet')).toBeInTheDocument();
  });

  it('shows the editor link on canManageGallery only', async () => {
    await renderGallery({ '/api/v1/auth/me': editor });
    expect(
      await screen.findByRole('link', { name: 'Manage this on the classic site ↗' }),
    ).toHaveAttribute('href', '/gallery');
  });

  it('hides the editor link from managers', async () => {
    await renderGallery({ '/api/v1/auth/me': manager });
    await new Promise((r) => setTimeout(r, 0));
    expect(screen.queryByRole('link', { name: /classic site/ })).toBeNull();
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/gallery`
Expected: FAIL. `./GalleryPage` can't be resolved.

- [ ] **Step 3: Implement**

`client/pages/gallery/Lightbox.tsx`:

```tsx
import { useEffect, useRef } from 'react';

import type { GalleryImage } from '../../api/gallery';
import { Button } from '../../components/ui/Button';
import { Modal } from '../../components/ui/Modal';

export interface LightboxProps {
  images: GalleryImage[];
  /** The open photo, or null when closed. */
  index: number | null;
  onIndexChange: (index: number) => void;
  onClose: () => void;
}

const SWIPE_PX = 50;

export function Lightbox({ images, index, onIndexChange, onClose }: LightboxProps) {
  const touchStartX = useRef<number | null>(null);
  const count = images.length;

  useEffect(() => {
    if (index === null) {
      return;
    }
    function onKey(e: KeyboardEvent) {
      if (index === null) {
        return;
      }
      if (e.key === 'ArrowRight') {
        e.preventDefault();
        onIndexChange((index + 1) % count);
      } else if (e.key === 'ArrowLeft') {
        e.preventDefault();
        onIndexChange((index - 1 + count) % count);
      }
    }
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [index, count, onIndexChange]);

  const image = index !== null ? images[index] : undefined;
  const go = (delta: number) => {
    if (index !== null) {
      onIndexChange((index + delta + count) % count);
    }
  };

  return (
    <Modal
      open={index !== null}
      onClose={onClose}
      title="Gallery"
      className="w-[min(64rem,calc(100vw-2rem))]"
    >
      {image && index !== null && (
        <div
          data-testid="lightbox-stage"
          onTouchStart={(e) => {
            touchStartX.current = e.touches[0]?.clientX ?? null;
          }}
          onTouchEnd={(e) => {
            const start = touchStartX.current;
            touchStartX.current = null;
            const end = e.changedTouches[0]?.clientX;
            if (start === null || end === undefined) {
              return;
            }
            const dx = end - start;
            if (Math.abs(dx) >= SWIPE_PX) {
              go(dx < 0 ? 1 : -1);
            }
          }}
        >
          <figure>
            <img
              src={image.imageUrl}
              alt={image.caption ?? ''}
              className="mx-auto max-h-[70vh] w-auto rounded-md object-contain"
            />
            {image.caption && (
              <figcaption className="mt-2 text-center text-sm">{image.caption}</figcaption>
            )}
          </figure>
          <div className="mt-4 flex items-center justify-between gap-2">
            <Button variant="secondary" size="sm" aria-label="Previous photo" onClick={() => go(-1)}>
              ← Previous
            </Button>
            <span aria-live="polite" className="text-sm text-muted">
              {index + 1} / {count}
            </span>
            <div className="flex gap-2">
              <Button variant="secondary" size="sm" aria-label="Next photo" onClick={() => go(1)}>
                Next →
              </Button>
              <Button variant="ghost" size="sm" onClick={onClose}>
                Close
              </Button>
            </div>
          </div>
        </div>
      )}
    </Modal>
  );
}
```

`client/pages/gallery/GalleryPage.tsx`:

```tsx
import { useState } from 'react';

import { useGallery } from '../../api/gallery';
import { EditorLink } from '../../components/page/EditorLink';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { PageHeader } from '../../components/ui/PageHeader';
import { Lightbox } from './Lightbox';

export default function GalleryPage() {
  usePageTitle('Gallery');
  const gallery = useGallery();
  const [openIndex, setOpenIndex] = useState<number | null>(null);
  return (
    <>
      <PageHeader
        title="Gallery"
        actions={<EditorLink legacyHref="/gallery" permission="canManageGallery" />}
      />
      <QueryState query={gallery} isEmpty={(l) => l.length === 0} emptyTitle="No photos yet">
        {(images) => (
          <>
            <ul className="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-4">
              {images.map((img, i) => (
                <li key={img.id}>
                  <button
                    type="button"
                    aria-label={img.caption || `Photo ${i + 1} of ${images.length}`}
                    onClick={() => setOpenIndex(i)}
                    className="block aspect-square w-full overflow-hidden rounded-lg border border-line"
                  >
                    <img
                      src={img.imageUrl}
                      alt=""
                      loading="lazy"
                      className="size-full object-cover transition-transform hover:scale-105"
                    />
                  </button>
                </li>
              ))}
            </ul>
            <Lightbox
              images={images}
              index={openIndex}
              onIndexChange={setOpenIndex}
              onClose={() => setOpenIndex(null)}
            />
          </>
        )}
      </QueryState>
    </>
  );
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/pages/gallery && yarn vitest run client/pages/gallery`
Expected: PASS, 10 tests.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add client/pages/gallery
git commit -q -m "Port the gallery with a keyboard and swipe lightbox" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 10: Documents, Programmes and Sponsors

**Files:**
- Create: `client/pages/documents/DocumentsPage.tsx`, `client/pages/programmes/ProgrammesPage.tsx`, `client/pages/sponsors/SponsorsPage.tsx`
- Test: `client/pages/documents/documents.test.tsx`, `client/pages/programmes/programmes.test.tsx`, `client/pages/sponsors/sponsors.test.tsx`

**Interfaces:**
- Consumes:
  - `useDocuments`, `useProgrammes`, `useSeasons`, `useSponsors`, `Programme` (Task 4);
  - `SearchInput`, `useSearchQuery`, `matchesQuery`, `CardGrid`, `QueryState`, `EditorLink`, `usePageTitle`;
  - `EmptyState`, `ButtonLink`, `Card`, `CardBody`, `Field`, `Select`, `PageHeader`, `formatDate`.
- Produces: default exports `DocumentsPage`, `ProgrammesPage` and `SponsorsPage`.

- [ ] **Step 1: Write the failing tests**

`client/pages/documents/documents.test.tsx`:

```tsx
import { fireEvent, screen } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import { publicRoutes } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import DocumentsPage from './DocumentsPage';

function renderDocs(route = '/documents', overrides = {}) {
  mockFetch(publicRoutes(overrides));
  renderWithProviders(
    <Routes>
      <Route path="/documents" element={<DocumentsPage />} />
    </Routes>,
    { route },
  );
}

describe('DocumentsPage', () => {
  it('lists documents with labelled download links', async () => {
    renderDocs();
    expect(
      await screen.findByRole('link', { name: 'Download Club constitution' }),
    ).toHaveAttribute('href', '/api/v1/files/document/14');
    expect(screen.getByRole('link', { name: 'Download Safeguarding policy' })).toBeInTheDocument();
    expect(document.title).toBe('Documents · AFC Aldermaston');
  });

  it('filters as you type, ignoring case and spacing', async () => {
    renderDocs();
    const box = await screen.findByLabelText('Search documents');
    fireEvent.change(box, { target: { value: '  CLUB  const' } });
    expect(screen.getByText('Club constitution')).toBeInTheDocument();
    expect(screen.queryByText('Safeguarding policy')).toBeNull();
    fireEvent.change(box, { target: { value: '' } });
    expect(screen.getByText('Safeguarding policy')).toBeInTheDocument();
  });

  it('starts filtered from ?q= and says when nothing matches', async () => {
    renderDocs('/documents?q=zzz');
    expect(await screen.findByText("No documents match 'zzz'")).toBeInTheDocument();
  });

  it('shows the empty state', async () => {
    renderDocs('/documents', { '/api/v1/documents': { body: [] } });
    expect(await screen.findByText('No documents yet')).toBeInTheDocument();
  });
});
```

`client/pages/programmes/programmes.test.tsx`:

```tsx
import { fireEvent, screen, within } from '@testing-library/react';
import { Route, Routes, useLocation } from 'react-router';
import { describe, expect, it } from 'vitest';

import { programmes, publicRoutes } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import ProgrammesPage from './ProgrammesPage';

function Location() {
  const l = useLocation();
  return <output data-testid="location">{l.pathname + l.search}</output>;
}

function renderProgrammes(route = '/programmes', overrides = {}) {
  const fetchMock = mockFetch(
    publicRoutes({ '/api/v1/programmes?season=2': { body: [programmes[0]] }, ...overrides }),
  );
  renderWithProviders(
    <Routes>
      <Route
        path="/programmes"
        element={
          <>
            <ProgrammesPage />
            <Location />
          </>
        }
      />
    </Routes>,
    { route },
  );
  return fetchMock;
}

describe('ProgrammesPage', () => {
  it('groups programmes by season with "No season" last', async () => {
    renderProgrammes();
    const headings = await screen.findAllByRole('heading', { level: 2 });
    expect(headings.map((h) => h.textContent)).toEqual(['2026-27', '2025-26', 'No season']);
    const group = screen.getByRole('region', { name: '2026-27' });
    expect(within(group).getByText('vs Downton')).toBeInTheDocument();
    expect(within(group).getByText('5 Sep 2026')).toBeInTheDocument();
    expect(within(group).getByRole('link', { name: 'View vs Downton' })).toHaveAttribute(
      'href',
      '/api/v1/files/programme/16',
    );
    expect(document.title).toBe('Programmes · AFC Aldermaston');
  });

  it('filters by season through the address', async () => {
    const fetchMock = renderProgrammes();
    const select = await screen.findByLabelText('Season');
    await screen.findByRole('option', { name: '2026-27' });
    fireEvent.change(select, { target: { value: '2' } });
    expect(await screen.findByTestId('location')).toHaveTextContent('/programmes?season=2');
    expect(await screen.findByText('vs Downton')).toBeInTheDocument();
    expect(screen.queryByText('vs Marlow')).toBeNull();
    expect(fetchMock.mock.calls.some(([u]) => String(u) === '/api/v1/programmes?season=2')).toBe(
      true,
    );
    fireEvent.change(select, { target: { value: '0' } });
    expect(screen.getByTestId('location')).toHaveTextContent(/^\/programmes$/);
  });

  it('searches by name', async () => {
    renderProgrammes('/programmes?q=marlow');
    expect(await screen.findByText('vs Marlow')).toBeInTheDocument();
    expect(screen.queryByText('vs Downton')).toBeNull();
  });

  it('says when nothing matches, and when there are none', async () => {
    renderProgrammes('/programmes?q=zzz');
    expect(await screen.findByText("No programmes match 'zzz'")).toBeInTheDocument();
  });
});
```

`client/pages/sponsors/sponsors.test.tsx`:

```tsx
import { screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { publicRoutes } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import SponsorsPage from './SponsorsPage';

describe('SponsorsPage', () => {
  it('shows every sponsor with purpose, team and website', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<SponsorsPage />);
    const acme = (await screen.findByRole('heading', { name: 'Acme Ltd' })).closest('li');
    expect(acme).not.toBeNull();
    const card = within(acme as HTMLElement);
    expect(card.getByText('Kit sponsor')).toBeInTheDocument();
    expect(card.getByText('Sponsor of First Team')).toBeInTheDocument();
    expect(card.getByRole('link', { name: 'Acme Ltd website ↗' })).toHaveAttribute(
      'href',
      'https://acme.example',
    );
    // No logo: the name tile stands in, and no website link.
    const plain = within(screen.getByRole('heading', { name: 'Corner Shop' }).closest('li') as HTMLElement);
    expect(plain.queryByRole('link')).toBeNull();
    expect(screen.queryByRole('button', { name: /Show more/ })).toBeNull();
    expect(document.title).toBe('Sponsors · AFC Aldermaston');
  });

  it('shows the empty state', async () => {
    mockFetch(publicRoutes({ '/api/v1/sponsors': { body: [] } }));
    renderWithProviders(<SponsorsPage />);
    expect(await screen.findByText('No sponsors yet')).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/documents client/pages/programmes client/pages/sponsors`
Expected: FAIL. The page modules can't be resolved.

- [ ] **Step 3: Implement**

`client/pages/documents/DocumentsPage.tsx`:

```tsx
import { useDocuments } from '../../api/documents';
import { EditorLink } from '../../components/page/EditorLink';
import { QueryState } from '../../components/page/QueryState';
import { SearchInput } from '../../components/page/SearchInput';
import { usePageTitle } from '../../components/page/usePageTitle';
import { useSearchQuery } from '../../components/page/useSearchQuery';
import { ButtonLink } from '../../components/ui/ButtonLink';
import { EmptyState } from '../../components/ui/EmptyState';
import { PageHeader } from '../../components/ui/PageHeader';
import { matchesQuery } from '../../lib/text';

export default function DocumentsPage() {
  usePageTitle('Documents');
  const docs = useDocuments();
  const q = useSearchQuery();
  return (
    <>
      <PageHeader title="Documents" actions={<EditorLink legacyHref="/documents" />} />
      <QueryState query={docs} isEmpty={(l) => l.length === 0} emptyTitle="No documents yet">
        {(list) => {
          const shown = list.filter((d) => matchesQuery(d.name, q));
          return (
            <>
              <SearchInput label="Search documents" />
              {shown.length === 0 ? (
                <EmptyState title={`No documents match '${q}'`} />
              ) : (
                <ul className="mt-4 divide-y divide-line rounded-lg border border-line">
                  {shown.map((d) => (
                    <li key={d.id} className="flex items-center justify-between gap-3 px-4 py-3">
                      <span className="font-medium">{d.name}</span>
                      <ButtonLink
                        href={d.fileUrl}
                        variant="secondary"
                        size="sm"
                        aria-label={`Download ${d.name}`}
                      >
                        Download
                      </ButtonLink>
                    </li>
                  ))}
                </ul>
              )}
            </>
          );
        }}
      </QueryState>
    </>
  );
}
```

`client/pages/programmes/ProgrammesPage.tsx`:

```tsx
import { useId } from 'react';
import { useSearchParams } from 'react-router';

import { useProgrammes, useSeasons, type Programme } from '../../api/programmes';
import { EditorLink } from '../../components/page/EditorLink';
import { QueryState } from '../../components/page/QueryState';
import { SearchInput } from '../../components/page/SearchInput';
import { usePageTitle } from '../../components/page/usePageTitle';
import { useSearchQuery } from '../../components/page/useSearchQuery';
import { ButtonLink } from '../../components/ui/ButtonLink';
import { Select } from '../../components/ui/controls';
import { EmptyState } from '../../components/ui/EmptyState';
import { Field } from '../../components/ui/Field';
import { PageHeader } from '../../components/ui/PageHeader';
import { formatDate } from '../../lib/format';
import { parseId } from '../../lib/ids';
import { matchesQuery } from '../../lib/text';

const NO_SEASON = 'No season';

/** Groups in API order, with programmes that have no season last. */
function groupBySeason(list: Programme[]): [string, Programme[]][] {
  const groups = new Map<string, Programme[]>();
  for (const p of list) {
    const key = p.season?.name ?? NO_SEASON;
    groups.set(key, [...(groups.get(key) ?? []), p]);
  }
  const entries = [...groups.entries()];
  return [...entries.filter(([k]) => k !== NO_SEASON), ...entries.filter(([k]) => k === NO_SEASON)];
}

function SeasonGroup({ name, items }: { name: string; items: Programme[] }) {
  const headingId = useId();
  return (
    <section aria-labelledby={headingId}>
      <h2 id={headingId} className="mb-2 font-display text-2xl font-extrabold tracking-wide uppercase">
        {name}
      </h2>
      <ul className="divide-y divide-line rounded-lg border border-line">
        {items.map((p) => (
          <li key={p.id} className="flex items-center justify-between gap-3 px-4 py-3">
            <span>
              <span className="font-medium">{p.name}</span>{' '}
              <span className="text-sm text-muted">{formatDate(p.date)}</span>
            </span>
            <ButtonLink
              href={p.fileUrl}
              variant="secondary"
              size="sm"
              target="_blank"
              rel="noopener noreferrer"
              aria-label={`View ${p.name}`}
            >
              View
            </ButtonLink>
          </li>
        ))}
      </ul>
    </section>
  );
}

export default function ProgrammesPage() {
  usePageTitle('Programmes');
  const [params, setParams] = useSearchParams();
  const seasonId = parseId(params.get('season') ?? undefined) ?? 0;
  const programmes = useProgrammes(seasonId);
  const seasons = useSeasons();
  const q = useSearchQuery();

  return (
    <>
      <PageHeader title="Programmes" actions={<EditorLink legacyHref="/programmes" />} />
      <div className="mb-6 flex flex-wrap items-end gap-4">
        <Field label="Season" className="w-56">
          <Select
            value={String(seasonId)}
            onChange={(e) => {
              const next = new URLSearchParams(params);
              if (e.target.value === '0') {
                next.delete('season');
              } else {
                next.set('season', e.target.value);
              }
              setParams(next, { replace: true });
            }}
          >
            <option value="0">All seasons</option>
            {seasons.data?.map((s) => (
              <option key={s.id} value={String(s.id)}>
                {s.name}
              </option>
            ))}
          </Select>
        </Field>
        <SearchInput label="Search programmes" />
      </div>
      <QueryState query={programmes} isEmpty={(l) => l.length === 0} emptyTitle="No programmes yet">
        {(list) => {
          const shown = list.filter((p) => matchesQuery(p.name, q));
          if (shown.length === 0) {
            return <EmptyState title={`No programmes match '${q}'`} />;
          }
          return (
            <div className="space-y-8">
              {groupBySeason(shown).map(([name, items]) => (
                <SeasonGroup key={name} name={name} items={items} />
              ))}
            </div>
          );
        }}
      </QueryState>
    </>
  );
}
```

`client/pages/sponsors/SponsorsPage.tsx`:

```tsx
import { useSponsors } from '../../api/sponsors';
import { CardGrid } from '../../components/page/CardGrid';
import { EditorLink } from '../../components/page/EditorLink';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Card, CardBody } from '../../components/ui/Card';
import { PageHeader } from '../../components/ui/PageHeader';

export default function SponsorsPage() {
  usePageTitle('Sponsors');
  const sponsors = useSponsors();
  return (
    <>
      <PageHeader title="Sponsors" actions={<EditorLink legacyHref="/sponsors" />} />
      <QueryState query={sponsors}>
        {(list) => (
          <CardGrid
            items={list}
            getKey={(s) => s.id}
            pageSize={Number.POSITIVE_INFINITY}
            emptyTitle="No sponsors yet"
            render={(s) => (
              <Card className="h-full">
                <div className="flex h-32 items-center justify-center border-b border-line bg-white p-4 text-black">
                  {s.imageUrl ? (
                    <img src={s.imageUrl} alt="" loading="lazy" className="max-h-24 w-auto" />
                  ) : (
                    <span className="text-lg font-semibold">{s.name}</span>
                  )}
                </div>
                <CardBody className="space-y-1">
                  <h2 className="font-display text-xl font-extrabold uppercase">{s.name}</h2>
                  {s.purpose && <p className="text-sm">{s.purpose}</p>}
                  {s.team && <p className="text-sm text-muted">Sponsor of {s.team}</p>}
                  {s.website && (
                    <a
                      href={s.website}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-block text-sm font-semibold text-red"
                    >
                      {s.name} website ↗
                    </a>
                  )}
                </CardBody>
              </Card>
            )}
          />
        )}
      </QueryState>
    </>
  );
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/pages/documents client/pages/programmes client/pages/sponsors && yarn vitest run client/pages/documents client/pages/programmes client/pages/sponsors`
Expected: PASS, 10 tests.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add client/pages/documents client/pages/programmes client/pages/sponsors
git commit -q -m "Port documents and programmes with search, season filter in the URL, and sponsors" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 11: Info and Contact

**Files:**
- Create: `client/pages/info/InfoPage.tsx`, `client/pages/info/fallback.tsx`, `client/pages/contact/ContactPage.tsx`
- Test: `client/pages/info/info.test.tsx`, `client/pages/contact/contact.test.tsx`

**Interfaces:**
- Consumes:
  - `useInfo`, `useContact` (Task 4);
  - `RichText`, `QueryState`, `EditorLink`, `usePageTitle`;
  - `Alert`, `PageHeader`;
  - `crest`.
- Produces: default exports `InfoPage` and `ContactPage`; `InfoFallback()`.

- [ ] **Step 1: Write the failing tests**

`client/pages/info/info.test.tsx`:

```tsx
import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { editor, publicRoutes } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import InfoPage from './InfoPage';

describe('InfoPage', () => {
  it('shows the stored info, cleaned', async () => {
    mockFetch(
      publicRoutes({
        '/api/v1/info': { body: { content: '<h2>Welcome</h2><p>Hi<script>x()</script></p>' } },
      }),
    );
    const { container } = renderWithProviders(<InfoPage />);
    expect(await screen.findByRole('heading', { level: 2, name: 'Welcome' })).toBeInTheDocument();
    expect(container.querySelector('script')).toBeNull();
    expect(document.title).toBe('Information · AFC Aldermaston');
  });

  it('falls back to the club history when nothing is stored', async () => {
    mockFetch(publicRoutes({ '/api/v1/info': { body: { content: '  ' } } }));
    renderWithProviders(<InfoPage />);
    expect(await screen.findByRole('heading', { name: 'Club history' })).toBeInTheDocument();
    expect(screen.getByText(/founded as AWRE Football Club in 1952/)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'secretary@afcaldermaston.co.uk' })).toHaveAttribute(
      'href',
      'mailto:secretary@afcaldermaston.co.uk',
    );
  });

  it('links editors to the classic info editor', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': editor }));
    renderWithProviders(<InfoPage />);
    expect(
      await screen.findByRole('link', { name: 'Manage this on the classic site ↗' }),
    ).toHaveAttribute('href', '/info/edit');
  });
});
```

`client/pages/contact/contact.test.tsx`:

```tsx
import { screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { contact, publicRoutes } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import ContactPage from './ContactPage';

describe('ContactPage', () => {
  it('lists officials with their own email when no site email is set', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<ContactPage />);
    const sam = (await screen.findByRole('heading', { name: 'Sam Sec' })).closest('li') as HTMLElement;
    expect(within(sam).getByText('Club Secretary')).toBeInTheDocument();
    expect(within(sam).getByRole('link', { name: 'sam@example.test' })).toHaveAttribute(
      'href',
      'mailto:sam@example.test',
    );
    // No photo → crest.
    expect(within(sam).getByRole('img')).toHaveAttribute('src', expect.stringContaining('crest'));
    expect(document.title).toBe('Contact · AFC Aldermaston');
  });

  it('uses the site-wide email for everyone when set', async () => {
    mockFetch(
      publicRoutes({
        '/api/v1/contact': { body: { ...contact, displayEmail: 'club@example.test' } },
      }),
    );
    renderWithProviders(<ContactPage />);
    await screen.findByRole('heading', { name: 'Sam Sec' });
    expect(screen.getAllByRole('link', { name: 'club@example.test' })).toHaveLength(2);
    expect(screen.queryByRole('link', { name: 'sam@example.test' })).toBeNull();
  });

  it('shows the satnav notice and a titled, lazy map', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<ContactPage />);
    await screen.findByRole('heading', { name: 'Sam Sec' });
    expect(screen.getByText(/use the postcode RG26 4QP/)).toBeInTheDocument();
    const map = screen.getByTitle('Map to Aldermaston Recreational Society');
    expect(map.tagName).toBe('IFRAME');
    expect(map).toHaveAttribute('loading', 'lazy');
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/info client/pages/contact`
Expected: FAIL. The page modules can't be resolved.

- [ ] **Step 3: Implement**

`client/pages/info/fallback.tsx` holds the legacy fixed wording, copied **verbatim** from the `{{else}}` branch of `server/internal/legacy/templates/info.tmpl`. Copy the exact paragraph text from that file (four paragraphs, which are separated there by `<br><br>`). Don't retype it from memory. Its shape:

```tsx
const heading = 'mt-6 mb-2 font-display text-2xl font-extrabold tracking-wide uppercase';

/** Shown when no info has been saved: the legacy site's built-in copy. */
export function InfoFallback() {
  return (
    <div className="prose">
      <h2 className={heading}>Club history</h2>
      <p>{/* paragraph 1 from info.tmpl: "Our club was founded as AWRE Football Club in 1952 … Operation Hurricane project." */}</p>
      <p>{/* paragraph 2: "In the late 1960s/early 1970s, we changed our name … placed in the Premier Division." */}</p>
      <p>{/* paragraph 3: "When the Hampshire and Wessex Leagues merged in 2004 … at that difficult time." */}</p>
      <p>{/* paragraph 4: "Unsurprisingly, we finished bottom of Division One … 1-0 to Cookham Dean." */}</p>
      <p>{/* paragraph 5: "We made a welcome return to Step 6 football … towards the top half of the table." */}</p>
      <h2 className={heading}>Join us @</h2>
      <p>
        <a href="mailto:secretary@afcaldermaston.co.uk">secretary@afcaldermaston.co.uk</a>
        <br />
        <a href="mailto:safeguardingofficer@afcaldermaston.co.uk">
          safeguardingofficer@afcaldermaston.co.uk
        </a>
      </p>
    </div>
  );
}
```

Replace each `{/* paragraph N … */}` with that paragraph's full text as a JSX string. The template has five text blocks separated by `<br><br>`. Keep the original wording and spelling exactly, and replace `&amp;` with `&`. Write the text as JSX string literals, e.g. `{'Reading & District League'}`, or as plain JSX text with `&apos;`/`&amp;` where ESLint's `react/no-unescaped-entities` requires. The test asserts the "founded as AWRE Football Club in 1952" phrase and the secretary mailto.

`client/pages/info/InfoPage.tsx`:

```tsx
import { useInfo } from '../../api/pages';
import { EditorLink } from '../../components/page/EditorLink';
import { QueryState } from '../../components/page/QueryState';
import { RichText } from '../../components/page/RichText';
import { usePageTitle } from '../../components/page/usePageTitle';
import { PageHeader } from '../../components/ui/PageHeader';
import { InfoFallback } from './fallback';

export default function InfoPage() {
  usePageTitle('Information');
  const info = useInfo();
  return (
    <>
      <PageHeader title="Information" actions={<EditorLink legacyHref="/info/edit" />} />
      <QueryState query={info}>
        {(data) =>
          data.content.trim() ? <RichText html={data.content} /> : <InfoFallback />
        }
      </QueryState>
    </>
  );
}
```

`client/pages/contact/ContactPage.tsx`:

```tsx
import { useContact } from '../../api/pages';
import crest from '../../assets/crest.png';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Alert } from '../../components/ui/Alert';
import { PageHeader } from '../../components/ui/PageHeader';

const MAP_SRC =
  'https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d8939.983502968527!2d-1.1655729266615455!3d51.361551652298395!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x4876a0226ede6355%3A0xefc85c4cd2bbf09b!2sAldermaston%20Recreational%20Society!5e1!3m2!1sen!2suk!4v1587086773128!5m2!1sen!2suk';

export default function ContactPage() {
  usePageTitle('Contact');
  const contact = useContact();
  return (
    <>
      <PageHeader title="Contact" />
      <QueryState query={contact}>
        {({ displayEmail, people }) => (
          <ul className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {people.map((p) => {
              // Legacy rule: a site-wide contact address, when set, replaces personal ones.
              const email = displayEmail || p.email;
              return (
                <li key={p.id} className="rounded-lg border border-line p-4 text-center">
                  <img
                    src={p.imageUrl || crest}
                    alt=""
                    loading="lazy"
                    className="mx-auto mb-3 size-28 rounded-full border border-line bg-white object-cover"
                  />
                  <h2 className="font-display text-xl font-extrabold uppercase">{p.name}</h2>
                  <p className="text-sm text-muted">{p.role}</p>
                  <a href={`mailto:${email}`} className="mt-1 inline-block text-sm text-red underline">
                    {email}
                  </a>
                </li>
              );
            })}
          </ul>
        )}
      </QueryState>
      <Alert tone="info" className="mt-8">
        If you&apos;re using a satnav, use the postcode <strong>RG26 4QP</strong> — the postcode
        listed takes you some distance away.
      </Alert>
      <iframe
        src={MAP_SRC}
        title="Map to Aldermaston Recreational Society"
        loading="lazy"
        referrerPolicy="no-referrer-when-downgrade"
        className="mt-4 h-[400px] w-full rounded-lg border border-line"
      />
    </>
  );
}
```

The test regex is `/use the postcode RG26 4QP/`. `<strong>` splits the text nodes, and `getByText` with a regex matches an element's own text nodes only. If that fails, match with a function matcher on the Alert's `textContent` (`screen.getByRole('status')`) instead, and record a ruling. The copy stays the same either way.

The Google Maps URL is copied from `server/internal/legacy/templates/contact.tmpl`. Check it character for character against that file.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/pages/info client/pages/contact && yarn vitest run client/pages/info client/pages/contact`
Expected: PASS, 6 tests.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add client/pages/info client/pages/contact
git commit -q -m "Port the info page (with legacy fallback copy) and the contact page" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 12: Routes, lazy loading, nav, and accessibility across the site

**Files:**
- Modify: `client/App.tsx`, `client/components/layout/Layout.tsx`, `client/components/layout/navItems.ts`, `client/components/layout/NavBar.test.tsx`, `client/test/axe.ts`, `client/App.test.tsx`, `client/a11y.test.tsx`, `README.md`

**Interfaces:**
- Consumes: every page default export (Tasks 5–11), `PageSkeleton` (Task 3), and `publicRoutes`, `editor`, `team`, `newsArticle`, `event` (Task 4).
- Produces: the final route table. `navItems` gains `to` on every item.

- [ ] **Step 1: Write the failing tests**

In `client/components/layout/NavBar.test.tsx`, replace the test `'uses a plain legacy link for pages not yet ported'` with:

```tsx
  it('keeps every nav item inside the app', () => {
    renderNav('/');
    fireEvent.click(screen.getByRole('link', { name: 'News' }));
    expect(screen.getByRole('link', { name: 'News' })).toHaveAttribute('aria-current', 'page');
    expect(screen.getByRole('link', { name: 'Home' })).not.toHaveAttribute('aria-current');
    for (const item of navItems) {
      expect(item.to).toBeDefined();
    }
  });
```

Replace `client/App.test.tsx`:

```tsx
import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import App from './App';
import { event, newsArticle, publicRoutes, team } from './test/fixtures';
import { mockFetch } from './test/mockFetch';
import { renderWithProviders } from './test/render';

function renderAt(path: string) {
  mockFetch(publicRoutes());
  return renderWithProviders(<App />, { route: path });
}

const lazy = { timeout: 3000 };

describe('App routes', () => {
  it('renders the home page at /', () => {
    renderAt('/');
    expect(screen.getByRole('heading', { level: 1, name: 'AFC Aldermaston' })).toBeInTheDocument();
  });

  it.each([
    ['/teams', 'Teams'],
    [`/team/${team.id}`, team.name],
    ['/news', 'News'],
    [`/news/${newsArticle.id}`, newsArticle.title],
    ['/whatson', "What's On"],
    [`/whatson/${event.id}`, event.title],
    ['/gallery', 'Gallery'],
    ['/documents', 'Documents'],
    ['/programmes', 'Programmes'],
    ['/sponsors', 'Sponsors'],
    ['/info', 'Information'],
    ['/contact', 'Contact'],
    ['/design', 'Design system'],
  ])('routes %s to its page', async (path, heading) => {
    renderAt(path);
    expect(await screen.findByRole('heading', { level: 1, name: heading }, lazy)).toBeInTheDocument();
  });

  it('renders the not-found page for unknown routes', async () => {
    renderAt('/no/such/page');
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Page not found' }, lazy),
    ).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Go to the start' })).toHaveAttribute('href', '/');
  });

  it('wraps pages in the layout', () => {
    renderAt('/');
    const banner = screen.getByRole('banner');
    expect(banner).toHaveTextContent('AFC Aldermaston');
    expect(banner).toHaveTextContent('Facta Non Verba');
    expect(screen.getByRole('link', { name: 'AFC Aldermaston home' })).toHaveAttribute('href', '/');
    expect(screen.getByRole('img', { name: 'The FA Charter Standard' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Skip to content' })).toHaveAttribute(
      'href',
      '#content',
    );
    expect(screen.getByRole('main')).toHaveAttribute('id', 'content');
    expect(screen.getByRole('navigation', { name: 'Main' })).toBeInTheDocument();
    expect(screen.getByRole('contentinfo')).toHaveTextContent('AFC Aldermaston');
  });
});
```

Replace `client/a11y.test.tsx`:

```tsx
import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import App from './App';
import { axeViolations } from './test/axe';
import { anonymous, editor, event, newsArticle, publicRoutes, team } from './test/fixtures';
import type { MockResponse } from './test/mockFetch';
import { mockFetch } from './test/mockFetch';
import { renderWithProviders } from './test/render';

const routes = [
  '/',
  '/teams',
  `/team/${team.id}`,
  '/news',
  `/news/${newsArticle.id}`,
  '/whatson',
  `/whatson/${event.id}`,
  '/gallery',
  '/documents',
  '/programmes',
  '/sponsors',
  '/info',
  '/contact',
  '/design',
  '/nope',
];

const cases: [string, string, MockResponse][] = routes.flatMap((route) => [
  [`${route}, signed out`, route, anonymous] as [string, string, MockResponse],
  [`${route}, signed in`, route, editor] as [string, string, MockResponse],
]);

describe.each(['light', 'dark'] as const)('accessibility (%s theme)', (theme) => {
  it.each(cases)('%s has no axe violations', async (_name, route, me) => {
    localStorage.setItem('afc-theme', theme);
    mockFetch(publicRoutes({ '/api/v1/auth/me': me }));
    const { container } = renderWithProviders(<App />, { route });
    await screen.findByRole('button', { name: me === anonymous ? 'Sign in' : /Ed Editor/ });
    await screen.findByRole('heading', { level: 1 }, { timeout: 3000 });
    // Let the page's own query settle so we audit content, not the skeleton.
    await screen.findByRole('contentinfo');
    await new Promise((r) => setTimeout(r, 50));
    expect(document.documentElement.dataset.theme).toBe(theme);
    expect(await axeViolations(container)).toEqual([]);
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/App.test.tsx client/components/layout/NavBar.test.tsx client/a11y.test.tsx`
Expected: FAIL. The new routes render "Page not found", the News nav item is a plain link without `aria-current` after the click, and `item.to` is undefined.

- [ ] **Step 3: Implement**

`client/components/layout/navItems.ts`: give every item a `to` equal to its legacy path (for example `{ label: 'Teams', legacyHref: '/teams', to: '/teams' }`), keeping `legacyHref`, and update the `to` comment to say every item is ported:

```ts
export const navItems: NavItem[] = [
  { label: 'Home', legacyHref: '/', to: '/' },
  { label: 'Teams', legacyHref: '/teams', to: '/teams' },
  { label: 'News', legacyHref: '/news', to: '/news' },
  { label: "What's On", legacyHref: '/whatson', to: '/whatson' },
  { label: 'Gallery', legacyHref: '/gallery', to: '/gallery' },
  { label: 'Documents', legacyHref: '/documents', to: '/documents' },
  { label: 'Programmes', legacyHref: '/programmes', to: '/programmes' },
  { label: 'Sponsors', legacyHref: '/sponsors', to: '/sponsors' },
  { label: 'Info', legacyHref: '/info', to: '/info' },
  { label: 'Contact', legacyHref: '/contact', to: '/contact' },
];
```

`NavBar` renders `NavLink` with `end`. For a section's detail pages (`/news/1`) the News item isn't highlighted, which matches sub-project 3's behaviour, so it's left as is.

`client/components/layout/Layout.tsx`: wrap the outlet in `Suspense`:

```tsx
import { Suspense } from 'react';
import { Outlet } from 'react-router';

import { PageSkeleton } from '../page/QueryState';
import { Container } from '../ui/Container';
import { Footer } from './Footer';
import { Masthead } from './Masthead';
import { NavBar } from './NavBar';
import { SkipLink } from './SkipLink';

export default function Layout() {
  return (
    <div className="flex min-h-screen flex-col bg-bg text-ink">
      <SkipLink />
      <header>
        <Masthead />
        <NavBar />
      </header>
      <main id="content" tabIndex={-1} className="flex-1 py-8 outline-none">
        <Container>
          <Suspense fallback={<PageSkeleton />}>
            <Outlet />
          </Suspense>
        </Container>
      </main>
      <Footer />
    </div>
  );
}
```

`client/App.tsx`:

```tsx
import { lazy } from 'react';
import { Route, Routes } from 'react-router';

import Layout from './components/layout/Layout';
import HomePage from './pages/home/HomePage';

const TeamsPage = lazy(() => import('./pages/teams/TeamsPage'));
const TeamPage = lazy(() => import('./pages/teams/TeamPage'));
const NewsListPage = lazy(() => import('./pages/news/NewsListPage'));
const NewsArticlePage = lazy(() => import('./pages/news/NewsArticlePage'));
const WhatsOnPage = lazy(() => import('./pages/whatson/WhatsOnPage'));
const EventPage = lazy(() => import('./pages/whatson/EventPage'));
const GalleryPage = lazy(() => import('./pages/gallery/GalleryPage'));
const DocumentsPage = lazy(() => import('./pages/documents/DocumentsPage'));
const ProgrammesPage = lazy(() => import('./pages/programmes/ProgrammesPage'));
const SponsorsPage = lazy(() => import('./pages/sponsors/SponsorsPage'));
const InfoPage = lazy(() => import('./pages/info/InfoPage'));
const ContactPage = lazy(() => import('./pages/contact/ContactPage'));
const DesignPage = lazy(() => import('./pages/DesignPage'));
const NotFoundPage = lazy(() => import('./pages/NotFoundPage'));

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<HomePage />} />
        <Route path="teams" element={<TeamsPage />} />
        <Route path="team/:id" element={<TeamPage />} />
        <Route path="news" element={<NewsListPage />} />
        <Route path="news/:id" element={<NewsArticlePage />} />
        <Route path="whatson" element={<WhatsOnPage />} />
        <Route path="whatson/:id" element={<EventPage />} />
        <Route path="gallery" element={<GalleryPage />} />
        <Route path="documents" element={<DocumentsPage />} />
        <Route path="programmes" element={<ProgrammesPage />} />
        <Route path="sponsors" element={<SponsorsPage />} />
        <Route path="info" element={<InfoPage />} />
        <Route path="contact" element={<ContactPage />} />
        <Route path="design" element={<DesignPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  );
}
```

`client/test/axe.ts`: add `iframes: false` to the `axe.run` options, because the Contact map is a cross-origin frame that axe can't audit in jsdom:

```ts
  const results = await axe.run(node, {
    iframes: false,
    rules: { 'color-contrast': { enabled: false } },
  });
```

`client/pages/DesignPage.test.tsx` renders `<App />` at `/design`, which is now lazy. Change each of its synchronous `getByRole('heading', { level: 1, name: 'Design system' })` and first-interaction `getByRole` calls to `await screen.findByRole(…, undefined, { timeout: 3000 })`. Make the tests `async` as needed, then use `getBy…` for the rest. Its `beforeEach` mock can become `mockFetch(publicRoutes())`.

`README.md`, under "Design system", add:

```markdown
### Public pages

Every public page is now in the React client: `/app`, `/app/teams`, `/app/news`, `/app/whatson`, `/app/gallery`, `/app/documents`, `/app/programmes`, `/app/sponsors`, `/app/info` and `/app/contact`, with detail pages for teams, articles and events. Editing still happens on the classic pages; signed-in editors see a "Manage this on the classic site" link on each page.
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client README.md 2>/dev/null; yarn vitest run client/App.test.tsx client/components/layout/NavBar.test.tsx client/pages/DesignPage.test.tsx client/a11y.test.tsx`
Expected: PASS. 16 App tests, 8 NavBar tests, 4 DesignPage tests, and 60 a11y tests (15 routes × signed in/out × 2 themes).

If axe reports violations, fix the markup in the page or piece responsible (names, landmarks, heading order). Don't disable rules, apart from the existing `color-contrast` switch and `iframes: false`.

- [ ] **Step 5: Full verification**

Run each of these and read the output:

```bash
yarn lint
yarn typecheck
yarn test:client
yarn build:client
yarn test:server
```

Expected:
- lint exits 0;
- typecheck exits 0;
- every Vitest file passes;
- the build succeeds and the output lists separate lazy chunks (e.g. `TeamsPage-*.js`, `GalleryPage-*.js`);
- the Go tests pass (they're untouched).

Do a visual check with `vite preview` (see the sub-project 3 notes): `yarn vite preview --port 4179 --strictPort` in the background, then take headless Chrome screenshots of `/app/` and `/app/news`. API calls fail without the Go server, so pages show the error-with-retry state. The purpose is to confirm the shell and lazy loading work in a real browser. Stop the preview process afterwards.

- [ ] **Step 6: Commit**

```bash
git add -A client README.md
git commit -q -m "Route every public page in the SPA with lazy loading, and extend axe checks" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```
