# Design System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give the React client at `/app` its visual identity (crest-led palette, Barlow Condensed and Inter, light and dark themes), a styled layout shell with sign-in and sign-out, a shared component set, and a `/app/design` showcase.

**Architecture:**
- Tailwind CSS v4 via `@tailwindcss/vite`. Tokens are CSS variables in `@theme static`, with a `[data-theme='dark']` override block.
- Components live in `client/components/ui/`, one file each. Each keeps its own utility classes, with a typed variant map joined by `clsx`.
- Theme state comes from a `ThemeProvider`, plus a pre-paint script in `client/public/theme-init.js`.
- The shell lives in `client/components/layout/`. Nav links point at legacy URLs until sub-project 4 gives them SPA routes.

**Tech Stack:** React 19, React Router 7, TanStack Query 5, TypeScript 6, Vite 8, Tailwind CSS 4.3, clsx 2, Fontsource (Inter Variable, Barlow Condensed), Vitest 5, Testing Library, jsdom and axe-core 4. Yarn 4 runs through corepack.

**Spec:** `docs/superpowers/specs/2026-09-29-design-system-design.md`

## Global Constraints

- All work happens in the worktree `/Users/liam/Code/Go/AFC-design-system` on branch `design-system`. Never touch `/Users/liam/Code/Go/AFC`, and never read any `postgres_*.sql` file.
- Every commit message ends with a blank line and then `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
- Components use token classes only (`bg-surface`, `text-muted`, `border-line`, `text-red`, `bg-footer`, `text-on-red`…), never raw hex values. The one exception is hex inside `app.css` and the swatch table in `DesignPage`.
- Tokens, with light values then dark values:
  - `red` `#d30e27` / `#ff4d5e`
  - `red-hover` `#b10c20` / `#ff6b79`
  - `club-red` `#fc0f2b` / `#fc0f2b`
  - `blue` `#233a8f` / `#8ea2f0`
  - `footer` `#233a8f` / `#1b2550`
  - `ink` `#1f1f1f` / `#eceef2`
  - `muted` `#5b6170` / `#a2a8b5`
  - `line` `#e4e6eb` / `#2a2e37`
  - `bg` `#ffffff` / `#121419`
  - `surface` `#f5f6f8` / `#1a1d24`
  - `field` `#ffffff` / `#0e1015`
  - `success` `#1f8a4c` / `#4ade80`
  - `warning` `#b45309` / `#fbbf24`
  - `on-red` `#ffffff` / `#1a0005`
- Fonts:
  - `--font-display: 'Barlow Condensed', sans-serif` (weights 600, 700, 800, Latin);
  - `--font-sans: 'Inter Variable', system-ui, sans-serif`.
- Theme storage key `afc-theme`. Values `'light' | 'dark'`; when absent, the setting is `'system'`. The toggle cycles System → Light → Dark → System.
- Exact copy used in the shell:
  - "AFC Aldermaston"
  - "Facta Non Verba · Founded 1952" (phones: "Facta Non Verba · 1952")
  - "Sign in", "Sign out?", "Incorrect email or password.", "Signed in as <name>"
  - "© 2020–<year> AFC Aldermaston · Website provided by BSWDI"
  - "Visitor count: <n>"
  - "Skip to content"
- Social URLs: Facebook `https://www.facebook.com/AFC-Aldermaston-114651238068/` and X `https://x.com/afcaldermaston`. The BSWDI link is `https://bswdi.co.uk`.
- The nav order is fixed: Home, Teams, News, What's On, Gallery, Documents, Programmes, Sponsors, Info, Contact.
- There are no server (Go) changes. `yarn lint`, `yarn typecheck`, `yarn test:client` and `yarn build:client` must stay green.
- There is no barrel `index.ts`. `react-refresh/only-export-components` warns on files that mix component and non-component exports, so import each component from its own file.
- Prettier settings: `printWidth` 100, single quotes. Run `yarn eslint --fix <files>` before each commit.

## Review Focus

1. **Blocked `localStorage` or a missing `matchMedia`** (private mode, old browsers): the theme falls back to `system`/light and nothing throws. Tests are pinned in Task 1 (`theme-init.js`) and Task 2 (`ThemeProvider`).
2. **A card image URL that 404s:** `CardMedia` swaps to the gradient instead of showing a broken image. Pinned in Task 3.
3. **Sign-out after the session has already expired** (logout returns 401): the UI still ends up signed out, not stuck. Pinned in Task 9.
4. **A modal unmounted while open** (a successful sign-in unmounts `SignInDialog`): no errors, the dialog closes, and focus goes back. Pinned in Task 6.
5. **Long user names and labels on a 360px phone:** the account button truncates (`max-w-40 truncate`) rather than pushing the Menu button off-screen. jsdom can't measure this, so check it by hand at `/app/design` and on the shell in Task 11.

---

## File map

| File | Task | Responsibility |
|---|---|---|
| `package.json`, `yarn.lock`, `vite.config.ts` | 1 | Dependencies and the Tailwind plugin |
| `client/styles/app.css` | 1 | Tailwind import, fonts, tokens, dark overrides, base styles |
| `client/assets/crest.png`, `client/assets/fa-logo.jpeg` | 1 | Brand images |
| `client/public/favicon.png`, `client/public/apple-touch-icon.png` | 1 | Crest favicons (replacing `favicon.svg`) |
| `client/public/theme-init.js`, `client/theme/themeInit.test.ts` | 1 | Pre-paint theme script |
| `client/index.html`, `client/main.tsx` | 1, 2, 8 | Metadata, stylesheet, providers |
| `client/theme/setting.ts`, `context.ts`, `ThemeProvider.tsx`, `useTheme.ts`, `ThemeToggle.tsx` (+ tests) | 2 | Theme state and toggle |
| `client/test/matchMedia.ts`, `client/test/dialog.ts`, `client/test/axe.ts`, `client/test/setup.ts`, `client/test/render.tsx`, `client/test/mockFetch.ts` | 2, 6, 8, 9, 11 | Test support |
| `client/components/ui/buttonStyles.ts`, `Button.tsx`, `ButtonLink.tsx`, `Spinner.tsx`, `Badge.tsx`, `Alert.tsx`, `Card.tsx`, `Skeleton.tsx`, `EmptyState.tsx`, `PageHeader.tsx`, `Container.tsx` (+ tests) | 3 | Basic components |
| `client/components/ui/fieldContext.ts`, `Field.tsx`, `controls.tsx`, `Checkbox.tsx`, `fieldError.ts` (+ tests) | 4 | Form components |
| `client/components/ui/Table.tsx` (+ test) | 5 | Table primitives |
| `client/components/ui/Modal.tsx`, `ConfirmDialog.tsx` (+ tests) | 6 | Dialogs |
| `client/components/ui/useDisclosure.ts`, `useDismiss.ts`, `Menu.tsx` (+ test) | 7 | Popovers |
| `client/components/ui/toast/context.ts`, `ToastProvider.tsx`, `useToast.ts` (+ test) | 8 | Toasts |
| `client/api/auth.ts`, `client/api/types.ts`, `client/lib/navigation.ts`, `client/components/layout/SignInDialog.tsx`, `AccountControl.tsx` (+ test) | 9 | Sign-in and account menu |
| `client/components/layout/navItems.ts`, `Masthead.tsx`, `NavBar.tsx`, `Footer.tsx`, `SkipLink.tsx`, `Layout.tsx` (+ tests); delete `client/components/Layout.tsx` | 10 | Shell |
| `client/pages/DesignPage.tsx`, `HomePage.tsx`, `NotFoundPage.tsx`, `client/App.tsx` (+ tests), `README.md` | 11 | Pages, accessibility smoke tests, docs |

---

### Task 1: Tailwind, tokens, fonts, brand assets and the pre-paint theme script

**Files:**
- Modify: `package.json`, `yarn.lock` (via `yarn add`), `vite.config.ts`, `client/index.html`, `client/main.tsx`
- Create: `client/styles/app.css`, `client/assets/crest.png`, `client/assets/fa-logo.jpeg`, `client/public/favicon.png`, `client/public/apple-touch-icon.png`, `client/public/theme-init.js`
- Delete: `client/public/favicon.svg`
- Test: `client/theme/themeInit.test.ts`

**Interfaces:**
- Produces:
  - Tailwind utilities for every token (`bg-red`, `text-muted`, `border-line`, `bg-footer`, `text-on-red`, `font-display`, …);
  - the `dark:` variant bound to `[data-theme=dark]`;
  - `client/public/theme-init.js`, which sets `document.documentElement.dataset.theme` to `'light' | 'dark'`;
  - image modules `client/assets/crest.png` and `client/assets/fa-logo.jpeg`.

- [ ] **Step 1: Install dependencies**

```bash
cd /Users/liam/Code/Go/AFC-design-system
corepack enable
yarn install --immutable
yarn add clsx @fontsource-variable/inter @fontsource/barlow-condensed
yarn add -D tailwindcss @tailwindcss/vite axe-core
```

Expected: `package.json` gains `clsx`, `@fontsource-variable/inter` and `@fontsource/barlow-condensed` under dependencies, and `tailwindcss`, `@tailwindcss/vite` and `axe-core` under devDependencies.

- [ ] **Step 2: Write the failing test for the pre-paint script**

Create `client/theme/themeInit.test.ts`:

```ts
import { afterEach, describe, expect, it, vi } from 'vitest';

import src from '../public/theme-init.js?raw';
import { installMatchMedia } from '../test/matchMedia';

function runInit() {
  new Function(src)();
}

describe('theme-init.js', () => {
  afterEach(() => {
    localStorage.clear();
    delete document.documentElement.dataset.theme;
  });

  it('uses a stored dark setting', () => {
    installMatchMedia(false);
    localStorage.setItem('afc-theme', 'dark');
    runInit();
    expect(document.documentElement.dataset.theme).toBe('dark');
  });

  it('uses a stored light setting even when the system is dark', () => {
    installMatchMedia(true);
    localStorage.setItem('afc-theme', 'light');
    runInit();
    expect(document.documentElement.dataset.theme).toBe('light');
  });

  it('follows the system when nothing is stored', () => {
    installMatchMedia(true);
    runInit();
    expect(document.documentElement.dataset.theme).toBe('dark');
  });

  it('ignores junk in storage', () => {
    installMatchMedia(false);
    localStorage.setItem('afc-theme', 'purple');
    runInit();
    expect(document.documentElement.dataset.theme).toBe('light');
  });

  it('falls back to the system when storage throws', () => {
    installMatchMedia(true);
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('blocked');
    });
    runInit();
    expect(document.documentElement.dataset.theme).toBe('dark');
  });

  it('chooses light when matchMedia is missing', () => {
    vi.stubGlobal('matchMedia', undefined);
    runInit();
    expect(document.documentElement.dataset.theme).toBe('light');
  });
});
```

Create `client/test/matchMedia.ts` (Task 2 reuses it):

```ts
import { vi } from 'vitest';

export const DARK_QUERY = '(prefers-color-scheme: dark)';

export interface MediaController {
  /** Flips the OS preference and notifies every change listener. */
  setDark(dark: boolean): void;
}

/** Stubs window.matchMedia; only the dark-scheme query ever matches. */
export function installMatchMedia(initialDark = false): MediaController {
  let dark = initialDark;
  const listeners = new Set<(e: MediaQueryListEvent) => void>();
  vi.stubGlobal(
    'matchMedia',
    vi.fn((query: string) => ({
      get matches() {
        return query === DARK_QUERY && dark;
      },
      media: query,
      onchange: null,
      addEventListener: (_type: string, fn: (e: MediaQueryListEvent) => void) => listeners.add(fn),
      removeEventListener: (_type: string, fn: (e: MediaQueryListEvent) => void) =>
        listeners.delete(fn),
      addListener: (fn: (e: MediaQueryListEvent) => void) => listeners.add(fn),
      removeListener: (fn: (e: MediaQueryListEvent) => void) => listeners.delete(fn),
      dispatchEvent: () => false,
    })),
  );
  return {
    setDark(next: boolean) {
      dark = next;
      listeners.forEach((fn) => fn({ matches: next, media: DARK_QUERY } as MediaQueryListEvent));
    },
  };
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `yarn vitest run client/theme/themeInit.test.ts`
Expected: FAIL. The import `../public/theme-init.js?raw` can't be resolved, because the file doesn't exist.

- [ ] **Step 4: Write the pre-paint script**

Create `client/public/theme-init.js`:

```js
// Applies the saved light/dark theme before first paint so there is no flash.
// Keep in step with client/theme/setting.ts (same key, same rules).
(function () {
  var setting = 'system';
  try {
    var saved = window.localStorage.getItem('afc-theme');
    if (saved === 'light' || saved === 'dark') {
      setting = saved;
    }
  } catch (e) {
    // Storage blocked (private mode, policy): follow the system.
  }
  var dark =
    setting === 'dark' ||
    (setting === 'system' &&
      typeof window.matchMedia === 'function' &&
      window.matchMedia('(prefers-color-scheme: dark)').matches);
  document.documentElement.dataset.theme = dark ? 'dark' : 'light';
})();
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `yarn vitest run client/theme/themeInit.test.ts`
Expected: PASS, 6 tests.

- [ ] **Step 6: Add the Tailwind plugin, stylesheet, assets and HTML metadata**

`vite.config.ts`: add the import and the plugin.

```ts
import tailwindcss from '@tailwindcss/vite';
```

```ts
  plugins: [react(), tailwindcss()],
```

Create `client/styles/app.css`:

```css
@import 'tailwindcss';
@import '@fontsource-variable/inter/wght.css';
@import '@fontsource/barlow-condensed/latin-600.css';
@import '@fontsource/barlow-condensed/latin-700.css';
@import '@fontsource/barlow-condensed/latin-800.css';

@custom-variant dark (&:where([data-theme=dark], [data-theme=dark] *));

/*
 * Design tokens. `static` emits every variable even when no utility uses it,
 * because some are read with var() from inline styles (card gradient, swatches).
 * The default palette is cleared so only club tokens exist as utilities.
 */
@theme static {
  --color-*: initial;
  --color-white: #ffffff;
  --color-black: #000000;
  --color-red: #d30e27;
  --color-red-hover: #b10c20;
  --color-club-red: #fc0f2b;
  --color-blue: #233a8f;
  --color-footer: #233a8f;
  --color-ink: #1f1f1f;
  --color-muted: #5b6170;
  --color-line: #e4e6eb;
  --color-bg: #ffffff;
  --color-surface: #f5f6f8;
  --color-field: #ffffff;
  --color-success: #1f8a4c;
  --color-warning: #b45309;
  --color-on-red: #ffffff;

  --font-display: 'Barlow Condensed', sans-serif;
  --font-sans: 'Inter Variable', system-ui, sans-serif;
}

[data-theme='dark'] {
  color-scheme: dark;
  --color-red: #ff4d5e;
  --color-red-hover: #ff6b79;
  --color-blue: #8ea2f0;
  --color-footer: #1b2550;
  --color-ink: #eceef2;
  --color-muted: #a2a8b5;
  --color-line: #2a2e37;
  --color-bg: #121419;
  --color-surface: #1a1d24;
  --color-field: #0e1015;
  --color-success: #4ade80;
  --color-warning: #fbbf24;
  --color-on-red: #1a0005;
}

@layer base {
  body {
    background-color: var(--color-bg);
    color: var(--color-ink);
    font-family: var(--font-sans);
  }

  :focus-visible {
    outline: 3px solid var(--color-blue);
    outline-offset: 2px;
  }

  @media (prefers-reduced-motion: reduce) {
    *,
    *::before,
    *::after {
      animation-duration: 0.01ms !important;
      animation-iteration-count: 1 !important;
      transition-duration: 0.01ms !important;
      scroll-behavior: auto !important;
    }
  }
}
```

`client/main.tsx`: add the stylesheet import directly after the React/router imports block and before `import { createQueryClient } …`:

```ts
import './styles/app.css';
```

Copy the assets and remove the placeholder favicon:

```bash
mkdir -p client/assets
cp server/internal/legacy/public/AFC-256.png client/assets/crest.png
cp server/internal/legacy/public/facs.jpeg client/assets/fa-logo.jpeg
cp server/internal/legacy/public/favicon-32x32.png client/public/favicon.png
cp server/internal/legacy/public/apple-touch-icon.png client/public/apple-touch-icon.png
git rm -q client/public/favicon.svg
grep -rn 'favicon.svg' client server scripts
```

Expected: the grep prints only `client/index.html`, which the next edit changes.

Replace `client/index.html` with:

```html
<!doctype html>
<html lang="en-GB">
  <head>
    <meta charset="UTF-8" />
    <!-- Sets data-theme before first paint; not bundled on purpose (vite-ignore). -->
    <script vite-ignore src="/app/theme-init.js"></script>
    <link rel="icon" type="image/png" sizes="32x32" href="/app/favicon.png" />
    <link rel="apple-touch-icon" href="/app/apple-touch-icon.png" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta name="color-scheme" content="light dark" />
    <meta name="theme-color" content="#fc0f2b" />
    <meta name="description" content="AFC Aldermaston Football Club" />
    <title>AFC Aldermaston</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/main.tsx"></script>
  </body>
</html>
```

- [ ] **Step 7: Verify the build emits the tokens, fonts and script**

Run:

```bash
BUILD_CLIENT_SKIP_LINT=true yarn build:client 2>&1 | tail -5
grep -c -- '--color-club-red:#fc0f2b' build/client/assets/*.css
grep -c "Barlow Condensed" build/client/assets/*.css
grep -o '<script[^>]*theme-init[^>]*>' build/client/index.html
ls build/client/theme-init.js build/client/favicon.png build/client/apple-touch-icon.png
```

Expected:
- the build succeeds with no "can't be bundled" warning;
- both CSS greps print `1` or more;
- the script tag is printed without the `vite-ignore` attribute;
- all three files exist.

Also run `yarn lint:client` and `yarn typecheck`, and expect both to exit 0. If ESLint reports `client/public/theme-init.js`, add `'client/public'` to `globalIgnores` in `eslint.config.js` and record that as a ruling.

- [ ] **Step 8: Commit**

```bash
yarn eslint --fix client/theme/themeInit.test.ts client/test/matchMedia.ts client/main.tsx vite.config.ts
git add -A package.json yarn.lock vite.config.ts client/index.html client/main.tsx client/styles client/assets client/public client/theme client/test/matchMedia.ts
git commit -q -m "Add Tailwind v4, design tokens, fonts, crest assets and pre-paint theme script" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 2: Theme provider and toggle

**Files:**
- Create: `client/theme/setting.ts`, `client/theme/context.ts`, `client/theme/ThemeProvider.tsx`, `client/theme/useTheme.ts`, `client/theme/ThemeToggle.tsx`
- Modify: `client/test/setup.ts`, `client/test/render.tsx`, `client/main.tsx`
- Test: `client/theme/ThemeProvider.test.tsx`, `client/theme/ThemeToggle.test.tsx`

**Interfaces:**
- Consumes: `installMatchMedia(initialDark?: boolean): MediaController` and `DARK_QUERY` from `client/test/matchMedia.ts` (Task 1).
- Produces:
  - `type ThemeSetting = 'system' | 'light' | 'dark'`
  - `type ResolvedTheme = 'light' | 'dark'`
  - `THEME_STORAGE_KEY = 'afc-theme'`
  - `nextSetting: Record<ThemeSetting, ThemeSetting>`
  - `<ThemeProvider>`
  - `useTheme(): { setting: ThemeSetting; resolved: ResolvedTheme; setSetting(s: ThemeSetting): void }`
  - `<ThemeToggle />`
  - `renderWithProviders` now wraps content in `ThemeProvider`.

- [ ] **Step 1: Write the failing tests**

Create `client/theme/ThemeProvider.test.tsx`:

```tsx
import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { installMatchMedia } from '../test/matchMedia';
import { ThemeProvider } from './ThemeProvider';
import { useTheme } from './useTheme';

function Probe() {
  const { setting, resolved, setSetting } = useTheme();
  return (
    <>
      <p>setting:{setting}</p>
      <p>resolved:{resolved}</p>
      <button onClick={() => setSetting('dark')}>dark</button>
      <button onClick={() => setSetting('system')}>system</button>
    </>
  );
}

function renderProbe() {
  return render(
    <ThemeProvider>
      <Probe />
    </ThemeProvider>,
  );
}

describe('ThemeProvider', () => {
  it('defaults to system and resolves light on a light OS', () => {
    installMatchMedia(false);
    renderProbe();
    expect(screen.getByText('setting:system')).toBeInTheDocument();
    expect(document.documentElement.dataset.theme).toBe('light');
  });

  it('resolves dark on a dark OS', () => {
    installMatchMedia(true);
    renderProbe();
    expect(screen.getByText('resolved:dark')).toBeInTheDocument();
    expect(document.documentElement.dataset.theme).toBe('dark');
  });

  it('applies a stored setting', () => {
    installMatchMedia(false);
    localStorage.setItem('afc-theme', 'dark');
    renderProbe();
    expect(screen.getByText('setting:dark')).toBeInTheDocument();
    expect(document.documentElement.dataset.theme).toBe('dark');
  });

  it('follows OS changes live while on system', () => {
    const media = installMatchMedia(false);
    renderProbe();
    act(() => media.setDark(true));
    expect(document.documentElement.dataset.theme).toBe('dark');
  });

  it('persists a chosen setting and forgets it again for system', () => {
    installMatchMedia(false);
    renderProbe();
    fireEvent.click(screen.getByRole('button', { name: 'dark' }));
    expect(localStorage.getItem('afc-theme')).toBe('dark');
    expect(document.documentElement.dataset.theme).toBe('dark');
    fireEvent.click(screen.getByRole('button', { name: 'system' }));
    expect(localStorage.getItem('afc-theme')).toBeNull();
  });

  it('still renders when storage throws', () => {
    installMatchMedia(false);
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('blocked');
    });
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('blocked');
    });
    renderProbe();
    expect(screen.getByText('setting:system')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'dark' }));
    expect(screen.getByText('setting:dark')).toBeInTheDocument();
  });

  it('resolves light when matchMedia is missing', () => {
    vi.stubGlobal('matchMedia', undefined);
    renderProbe();
    expect(screen.getByText('resolved:light')).toBeInTheDocument();
  });
});
```

Create `client/theme/ThemeToggle.test.tsx`:

```tsx
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ThemeProvider } from './ThemeProvider';
import { ThemeToggle } from './ThemeToggle';

describe('ThemeToggle', () => {
  it('cycles system → light → dark → system and remembers the choice', () => {
    render(
      <ThemeProvider>
        <ThemeToggle />
      </ThemeProvider>,
    );
    const button = () => screen.getByRole('button', { name: /^Theme:/ });
    expect(button()).toHaveAccessibleName('Theme: system (switch to light)');

    fireEvent.click(button());
    expect(button()).toHaveAccessibleName('Theme: light (switch to dark)');
    expect(localStorage.getItem('afc-theme')).toBe('light');

    fireEvent.click(button());
    expect(button()).toHaveAccessibleName('Theme: dark (switch to system)');
    expect(document.documentElement.dataset.theme).toBe('dark');

    fireEvent.click(button());
    expect(button()).toHaveAccessibleName('Theme: system (switch to light)');
    expect(localStorage.getItem('afc-theme')).toBeNull();
  });
});
```

Update `client/test/setup.ts` so every test starts with a light OS and clean storage:

```ts
import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterEach, beforeEach, vi } from 'vitest';

import { installMatchMedia } from './matchMedia';

beforeEach(() => {
  installMatchMedia(false);
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
  localStorage.clear();
  delete document.documentElement.dataset.theme;
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/theme`
Expected: FAIL. `./ThemeProvider`, `./useTheme` and `./ThemeToggle` can't be resolved. (`themeInit.test.ts` still passes.)

- [ ] **Step 3: Implement the theme modules**

`client/theme/setting.ts`:

```ts
// Keep in step with client/public/theme-init.js (same key, same rules).
export type ThemeSetting = 'system' | 'light' | 'dark';
export type ResolvedTheme = 'light' | 'dark';

export const THEME_STORAGE_KEY = 'afc-theme';
export const DARK_QUERY = '(prefers-color-scheme: dark)';

export const nextSetting: Record<ThemeSetting, ThemeSetting> = {
  system: 'light',
  light: 'dark',
  dark: 'system',
};

export function readSetting(): ThemeSetting {
  try {
    const saved = window.localStorage.getItem(THEME_STORAGE_KEY);
    if (saved === 'light' || saved === 'dark') {
      return saved;
    }
  } catch {
    // Storage blocked: follow the system.
  }
  return 'system';
}

export function writeSetting(setting: ThemeSetting): void {
  try {
    if (setting === 'system') {
      window.localStorage.removeItem(THEME_STORAGE_KEY);
    } else {
      window.localStorage.setItem(THEME_STORAGE_KEY, setting);
    }
  } catch {
    // Storage blocked: the choice lasts for this page only.
  }
}

export function systemPrefersDark(): boolean {
  return typeof window.matchMedia === 'function' && window.matchMedia(DARK_QUERY).matches;
}

export function resolveTheme(setting: ThemeSetting, prefersDark: boolean): ResolvedTheme {
  if (setting === 'system') {
    return prefersDark ? 'dark' : 'light';
  }
  return setting;
}
```

`client/theme/context.ts`:

```ts
import { createContext } from 'react';

import type { ResolvedTheme, ThemeSetting } from './setting';

export interface ThemeState {
  setting: ThemeSetting;
  resolved: ResolvedTheme;
  setSetting: (setting: ThemeSetting) => void;
}

export const ThemeContext = createContext<ThemeState | null>(null);
```

`client/theme/useTheme.ts`:

```ts
import { useContext } from 'react';

import { ThemeContext, type ThemeState } from './context';

export function useTheme(): ThemeState {
  const ctx = useContext(ThemeContext);
  if (!ctx) {
    throw new Error('useTheme must be used inside <ThemeProvider>');
  }
  return ctx;
}
```

`client/theme/ThemeProvider.tsx`:

```tsx
import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react';

import { ThemeContext, type ThemeState } from './context';
import {
  DARK_QUERY,
  readSetting,
  resolveTheme,
  systemPrefersDark,
  writeSetting,
  type ThemeSetting,
} from './setting';

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [setting, setSettingState] = useState<ThemeSetting>(readSetting);
  const [prefersDark, setPrefersDark] = useState(systemPrefersDark);

  useEffect(() => {
    if (typeof window.matchMedia !== 'function') {
      return;
    }
    const mql = window.matchMedia(DARK_QUERY);
    const onChange = (e: MediaQueryListEvent) => setPrefersDark(e.matches);
    mql.addEventListener('change', onChange);
    return () => mql.removeEventListener('change', onChange);
  }, []);

  const resolved = resolveTheme(setting, prefersDark);

  useEffect(() => {
    document.documentElement.dataset.theme = resolved;
  }, [resolved]);

  const setSetting = useCallback((next: ThemeSetting) => {
    writeSetting(next);
    setSettingState(next);
  }, []);

  const value = useMemo<ThemeState>(
    () => ({ setting, resolved, setSetting }),
    [setting, resolved, setSetting],
  );

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}
```

`client/theme/ThemeToggle.tsx`:

```tsx
import type { ReactNode } from 'react';

import { buttonClasses } from '../components/ui/buttonStyles';
import { nextSetting, type ThemeSetting } from './setting';
import { useTheme } from './useTheme';

const iconProps = {
  width: 18,
  height: 18,
  viewBox: '0 0 24 24',
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 2,
  strokeLinecap: 'round' as const,
  strokeLinejoin: 'round' as const,
  'aria-hidden': true,
};

const icons: Record<ThemeSetting, ReactNode> = {
  system: (
    <svg {...iconProps}>
      <rect x="3" y="4" width="18" height="12" rx="2" />
      <path d="M8 20h8M12 16v4" />
    </svg>
  ),
  light: (
    <svg {...iconProps}>
      <circle cx="12" cy="12" r="4" />
      <path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" />
    </svg>
  ),
  dark: (
    <svg {...iconProps}>
      <path d="M20 14.5A8 8 0 0 1 9.5 4a8 8 0 1 0 10.5 10.5z" />
    </svg>
  ),
};

export function ThemeToggle() {
  const { setting, setSetting } = useTheme();
  const next = nextSetting[setting];
  return (
    <button
      type="button"
      onClick={() => setSetting(next)}
      aria-label={`Theme: ${setting} (switch to ${next})`}
      title={`Theme: ${setting}`}
      className={buttonClasses('secondary', 'sm', 'px-2')}
    >
      {icons[setting]}
    </button>
  );
}
```

`ThemeToggle` imports `buttonClasses` from `client/components/ui/buttonStyles.ts`, which is formally part of Task 3. Create it now, exactly as Task 3 Step 3 shows it, so this task compiles. Task 3 then leaves it unchanged.

`client/components/ui/buttonStyles.ts`:

```ts
import { clsx } from 'clsx';

export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';
export type ButtonSize = 'sm' | 'md';

const base =
  'inline-flex items-center justify-center gap-2 rounded-md border font-semibold transition-colors disabled:cursor-not-allowed disabled:opacity-50';

const variants: Record<ButtonVariant, string> = {
  primary: 'border-transparent bg-red text-on-red hover:bg-red-hover',
  secondary: 'border-line bg-bg text-ink hover:bg-surface',
  ghost: 'border-transparent bg-transparent text-red hover:bg-surface',
  danger: 'border-red bg-transparent text-red hover:bg-red hover:text-on-red',
};

const sizes: Record<ButtonSize, string> = {
  sm: 'px-2.5 py-1 text-[13px]',
  md: 'px-3.5 py-2 text-sm',
};

export function buttonClasses(
  variant: ButtonVariant = 'primary',
  size: ButtonSize = 'md',
  className?: string,
): string {
  return clsx(base, variants[variant], sizes[size], className);
}
```

Update `client/test/render.tsx` to wrap in `ThemeProvider` (between `MemoryRouter` and `AuthProvider`):

```tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render } from '@testing-library/react';
import type { ReactElement } from 'react';
import { MemoryRouter } from 'react-router';

import { AuthProvider } from '../auth/AuthProvider';
import { ThemeProvider } from '../theme/ThemeProvider';

/** Renders ui inside the same providers as main.tsx, with retries off. */
export function renderWithProviders(ui: ReactElement, { route = '/' }: { route?: string } = {}) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: Infinity } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[route]}>
        <ThemeProvider>
          <AuthProvider>{ui}</AuthProvider>
        </ThemeProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}
```

In `client/main.tsx`, import `ThemeProvider` from `./theme/ThemeProvider` and wrap it around `AuthProvider`, inside `BrowserRouter`:

```tsx
      <BrowserRouter basename="/app">
        <ThemeProvider>
          <AuthProvider>
            <App />
          </AuthProvider>
        </ThemeProvider>
      </BrowserRouter>
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn vitest run client/theme && yarn test:client`
Expected: PASS. All theme tests pass (6 + 7 + 1), and the existing suite is still green.

- [ ] **Step 5: Commit**

```bash
yarn eslint --fix client/theme client/components/ui/buttonStyles.ts client/test client/main.tsx
yarn typecheck
git add client/theme client/components/ui/buttonStyles.ts client/test client/main.tsx
git commit -q -m "Add theme provider, dark mode setting and theme toggle" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 3: Basic components

**Files:**
- Create in `client/components/ui/`: `Button.tsx`, `ButtonLink.tsx`, `Spinner.tsx`, `Badge.tsx`, `Alert.tsx`, `Card.tsx`, `Skeleton.tsx`, `EmptyState.tsx`, `PageHeader.tsx`, `Container.tsx` (`buttonStyles.ts` already exists from Task 2)
- Test: `client/components/ui/basic.test.tsx`

**Interfaces:**
- Consumes: `buttonClasses(variant?, size?, className?)`, `ButtonVariant` and `ButtonSize` from `buttonStyles.ts`.
- Produces:
  - `Button(props: ComponentProps<'button'> & { variant?: ButtonVariant; size?: ButtonSize; loading?: boolean })`
  - `ButtonLink(props: Omit<ComponentProps<'a'>, 'href'> & { to?: string; href?: string; variant?; size? })`
  - `Spinner({ label?: string; className?: string })`
  - `Badge(ComponentProps<'span'> & { tone?: 'red' | 'blue' | 'neutral' })`
  - `Alert(ComponentProps<'div'> & { tone?: 'success' | 'error' | 'warning' | 'info' })`
  - `Card(ComponentProps<'div'>)`
  - `CardBody(ComponentProps<'div'>)`
  - `CardMedia({ src?: string; alt: string; aspect?: string; className?: string })`
  - `Skeleton({ className?: string })`
  - `EmptyState({ title: string; message?: ReactNode; action?: ReactNode })`
  - `PageHeader({ title: string; subtitle?: ReactNode; actions?: ReactNode })`
  - `Container(ComponentProps<'div'>)`

- [ ] **Step 1: Write the failing tests**

Create `client/components/ui/basic.test.tsx`:

```tsx
import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { describe, expect, it, vi } from 'vitest';

import { Alert } from './Alert';
import { Badge } from './Badge';
import { Button } from './Button';
import { ButtonLink } from './ButtonLink';
import { Card, CardBody, CardMedia } from './Card';
import { EmptyState } from './EmptyState';
import { PageHeader } from './PageHeader';
import { Spinner } from './Spinner';

describe('Button', () => {
  it('defaults to type=button and fires onClick', () => {
    const onClick = vi.fn();
    render(<Button onClick={onClick}>Save</Button>);
    const button = screen.getByRole('button', { name: 'Save' });
    expect(button).toHaveAttribute('type', 'button');
    fireEvent.click(button);
    expect(onClick).toHaveBeenCalledOnce();
  });

  it('is busy and disabled while loading', () => {
    const onClick = vi.fn();
    render(
      <Button loading onClick={onClick}>
        Save
      </Button>,
    );
    const button = screen.getByRole('button', { name: 'Save' });
    expect(button).toBeDisabled();
    expect(button).toHaveAttribute('aria-busy', 'true');
    fireEvent.click(button);
    expect(onClick).not.toHaveBeenCalled();
  });
});

describe('ButtonLink', () => {
  it('renders a router link for `to` and a plain link for `href`', () => {
    render(
      <MemoryRouter>
        <ButtonLink to="/design">Design</ButtonLink>
        <ButtonLink href="/news">News</ButtonLink>
      </MemoryRouter>,
    );
    expect(screen.getByRole('link', { name: 'Design' })).toHaveAttribute('href', '/design');
    expect(screen.getByRole('link', { name: 'News' })).toHaveAttribute('href', '/news');
  });
});

describe('Spinner', () => {
  it('announces loading', () => {
    render(<Spinner />);
    expect(screen.getByRole('status')).toHaveTextContent('Loading');
  });
});

describe('Badge and Alert', () => {
  it('renders badge text', () => {
    render(<Badge tone="red">U12s</Badge>);
    expect(screen.getByText('U12s')).toBeInTheDocument();
  });

  it('uses role=alert for errors and role=status otherwise', () => {
    render(
      <>
        <Alert tone="error">Broken</Alert>
        <Alert tone="success">Saved</Alert>
      </>,
    );
    expect(screen.getByRole('alert')).toHaveTextContent('Broken');
    expect(screen.getByRole('status')).toHaveTextContent('Saved');
  });
});

describe('CardMedia', () => {
  it('shows the uploaded image', () => {
    render(<CardMedia src="/files/news/1" alt="Cup final" />);
    expect(screen.getByRole('img', { name: 'Cup final' })).toHaveAttribute('src', '/files/news/1');
  });

  it('shows the gradient when there is no image', () => {
    const { container } = render(<CardMedia alt="Cup final" />);
    expect(screen.queryByRole('img')).toBeNull();
    expect(container.querySelector('[data-fallback]')).toBeInTheDocument();
  });

  it('treats an empty src as no image', () => {
    const { container } = render(<CardMedia src="" alt="Cup final" />);
    expect(container.querySelector('[data-fallback]')).toBeInTheDocument();
  });

  it('falls back to the gradient when the image fails to load', () => {
    const { container } = render(<CardMedia src="/files/news/404" alt="Cup final" />);
    fireEvent.error(screen.getByRole('img', { name: 'Cup final' }));
    expect(screen.queryByRole('img')).toBeNull();
    expect(container.querySelector('[data-fallback]')).toBeInTheDocument();
  });

  it('composes inside a card', () => {
    render(
      <Card>
        <CardMedia alt="" />
        <CardBody>First team win the cup</CardBody>
      </Card>,
    );
    expect(screen.getByText('First team win the cup')).toBeInTheDocument();
  });
});

describe('EmptyState and PageHeader', () => {
  it('renders the empty state parts', () => {
    render(<EmptyState title="No news yet" message="Check back soon." action={<a href="/">Home</a>} />);
    expect(screen.getByText('No news yet')).toBeInTheDocument();
    expect(screen.getByText('Check back soon.')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Home' })).toBeInTheDocument();
  });

  it('renders the page title as the only h1', () => {
    render(<PageHeader title="Teams" subtitle="All club teams" actions={<button>Add</button>} />);
    expect(screen.getByRole('heading', { level: 1, name: 'Teams' })).toBeInTheDocument();
    expect(screen.getByText('All club teams')).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/components/ui/basic.test.tsx`
Expected: FAIL. `./Alert` (and the others) can't be resolved.

- [ ] **Step 3: Implement the components**

`client/components/ui/Spinner.tsx`:

```tsx
import { clsx } from 'clsx';

export function Spinner({ label = 'Loading', className }: { label?: string; className?: string }) {
  return (
    <span role="status" className={clsx('inline-flex items-center', className)}>
      <span
        aria-hidden="true"
        className="size-5 rounded-full border-2 border-line border-t-red motion-safe:animate-spin"
      />
      <span className="sr-only">{label}</span>
    </span>
  );
}
```

`client/components/ui/Button.tsx`:

```tsx
import type { ComponentProps } from 'react';

import { buttonClasses, type ButtonSize, type ButtonVariant } from './buttonStyles';

export interface ButtonProps extends ComponentProps<'button'> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  loading?: boolean;
}

export function Button({
  variant,
  size,
  loading = false,
  disabled,
  className,
  type = 'button',
  children,
  ...props
}: ButtonProps) {
  return (
    <button
      type={type}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      className={buttonClasses(variant, size, className)}
      {...props}
    >
      {loading && (
        <span
          aria-hidden="true"
          className="size-3 rounded-full border-2 border-current border-t-transparent motion-safe:animate-spin"
        />
      )}
      {children}
    </button>
  );
}
```

`client/components/ui/ButtonLink.tsx`:

```tsx
import type { ComponentProps } from 'react';
import { Link } from 'react-router';

import { buttonClasses, type ButtonSize, type ButtonVariant } from './buttonStyles';

export interface ButtonLinkProps extends Omit<ComponentProps<'a'>, 'href'> {
  /** An SPA route (relative to /app). */
  to?: string;
  /** A full-page URL, e.g. a legacy page. */
  href?: string;
  variant?: ButtonVariant;
  size?: ButtonSize;
}

export function ButtonLink({ to, href, variant, size, className, ...props }: ButtonLinkProps) {
  const classes = buttonClasses(variant, size, className);
  if (to !== undefined) {
    return <Link to={to} className={classes} {...props} />;
  }
  return <a href={href} className={classes} {...props} />;
}
```

`client/components/ui/Badge.tsx`:

```tsx
import { clsx } from 'clsx';
import type { ComponentProps } from 'react';

export type BadgeTone = 'red' | 'blue' | 'neutral';

const tones: Record<BadgeTone, string> = {
  red: 'bg-red/12 text-red',
  blue: 'bg-blue/14 text-blue',
  neutral: 'border border-line bg-surface text-muted',
};

export function Badge({
  tone = 'neutral',
  className,
  ...props
}: ComponentProps<'span'> & { tone?: BadgeTone }) {
  return (
    <span
      className={clsx(
        'inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-bold tracking-wider uppercase',
        tones[tone],
        className,
      )}
      {...props}
    />
  );
}
```

`client/components/ui/Alert.tsx`:

```tsx
import { clsx } from 'clsx';
import type { ComponentProps } from 'react';

export type AlertTone = 'success' | 'error' | 'warning' | 'info';

const tones: Record<AlertTone, string> = {
  success: 'border-success',
  error: 'border-red',
  warning: 'border-warning',
  info: 'border-blue',
};

export function Alert({
  tone = 'info',
  className,
  ...props
}: ComponentProps<'div'> & { tone?: AlertTone }) {
  return (
    <div
      role={tone === 'error' ? 'alert' : 'status'}
      className={clsx(
        'rounded-lg border-l-4 bg-surface px-3 py-2.5 text-sm text-ink',
        tones[tone],
        className,
      )}
      {...props}
    />
  );
}
```

`client/components/ui/Card.tsx`:

```tsx
import { clsx } from 'clsx';
import { useState, type ComponentProps } from 'react';

export function Card({ className, ...props }: ComponentProps<'div'>) {
  return (
    <div
      className={clsx('overflow-hidden rounded-lg border border-line bg-bg', className)}
      {...props}
    />
  );
}

export function CardBody({ className, ...props }: ComponentProps<'div'>) {
  return <div className={clsx('p-4', className)} {...props} />;
}

export interface CardMediaProps {
  /** The uploaded image; missing or empty shows the club gradient instead. */
  src?: string;
  alt: string;
  /** CSS aspect-ratio, default 16 / 9. */
  aspect?: string;
  className?: string;
}

export function CardMedia({ src, alt, aspect = '16 / 9', className }: CardMediaProps) {
  const [failedSrc, setFailedSrc] = useState<string | null>(null);
  const style = { aspectRatio: aspect };

  if (!src || failedSrc === src) {
    return (
      <div
        aria-hidden="true"
        data-fallback=""
        className={clsx('w-full bg-linear-135 from-blue to-red', className)}
        style={style}
      />
    );
  }
  return (
    <img
      src={src}
      alt={alt}
      loading="lazy"
      onError={() => setFailedSrc(src)}
      className={clsx('w-full object-cover', className)}
      style={style}
    />
  );
}
```

`client/components/ui/Skeleton.tsx`:

```tsx
import { clsx } from 'clsx';

export function Skeleton({ className }: { className?: string }) {
  return (
    <div
      aria-hidden="true"
      className={clsx('h-3 rounded-md bg-line motion-safe:animate-pulse', className)}
    />
  );
}
```

`client/components/ui/EmptyState.tsx`:

```tsx
import type { ReactNode } from 'react';

export interface EmptyStateProps {
  title: string;
  message?: ReactNode;
  action?: ReactNode;
}

export function EmptyState({ title, message, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center gap-2 py-10 text-center text-muted">
      <p className="font-display text-xl font-extrabold tracking-wide text-ink uppercase">{title}</p>
      {message && <div className="text-sm">{message}</div>}
      {action && <div className="mt-2">{action}</div>}
    </div>
  );
}
```

`client/components/ui/PageHeader.tsx`:

```tsx
import type { ReactNode } from 'react';

export interface PageHeaderProps {
  title: string;
  subtitle?: ReactNode;
  actions?: ReactNode;
}

export function PageHeader({ title, subtitle, actions }: PageHeaderProps) {
  return (
    <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 className="font-display text-3xl leading-none font-extrabold tracking-wide uppercase md:text-4xl">
          {title}
        </h1>
        {subtitle && <p className="mt-2 text-muted">{subtitle}</p>}
      </div>
      {actions && <div className="flex flex-wrap gap-2">{actions}</div>}
    </div>
  );
}
```

`client/components/ui/Container.tsx`:

```tsx
import { clsx } from 'clsx';
import type { ComponentProps } from 'react';

export function Container({ className, ...props }: ComponentProps<'div'>) {
  return <div className={clsx('mx-auto w-full max-w-6xl px-4 md:px-6', className)} {...props} />;
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn vitest run client/components/ui/basic.test.tsx`
Expected: PASS, 13 tests.

- [ ] **Step 5: Commit**

```bash
yarn eslint --fix client/components/ui
yarn typecheck
git add client/components/ui
git commit -q -m "Add button, badge, alert, card, skeleton, empty state and page header components" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 4: Form components

**Files:**
- Create in `client/components/ui/`: `fieldContext.ts`, `Field.tsx`, `controls.tsx`, `Checkbox.tsx`, `fieldError.ts`
- Test: `client/components/ui/form.test.tsx`

**Interfaces:**
- Consumes: `ApiError` from `client/api/client.ts` (`fields: Record<string, string>`).
- Produces:
  - `Field({ label: string; help?: string; error?: string; id?: string; children: ReactNode; className?: string })`
  - `Input`, `Textarea`, `Select` and `FileInput`, which take the native props of `input`, `textarea`, `select` and `input` (minus `type`) respectively
  - `Checkbox(Omit<ComponentProps<'input'>, 'type'> & { label: ReactNode })`
  - `fieldError(err: unknown, name: string): string | undefined`

- [ ] **Step 1: Write the failing tests**

Create `client/components/ui/form.test.tsx`:

```tsx
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ApiError } from '../../api/client';
import { Checkbox } from './Checkbox';
import { FileInput, Input, Select, Textarea } from './controls';
import { Field } from './Field';
import { fieldError } from './fieldError';

describe('Field', () => {
  it('labels the control and describes it with the help text', () => {
    render(
      <Field label="Email" help="We never share it.">
        <Input type="email" />
      </Field>,
    );
    const input = screen.getByLabelText('Email');
    expect(input).toHaveAccessibleDescription('We never share it.');
    expect(input).not.toHaveAttribute('aria-invalid');
  });

  it('shows the error instead of the help and marks the control invalid', () => {
    render(
      <Field label="Email" help="We never share it." error="Enter a valid email address.">
        <Input type="email" />
      </Field>,
    );
    const input = screen.getByLabelText('Email');
    expect(input).toHaveAccessibleDescription('Enter a valid email address.');
    expect(input).toHaveAttribute('aria-invalid', 'true');
    expect(screen.queryByText('We never share it.')).toBeNull();
  });

  it('keeps an explicit id', () => {
    render(
      <Field label="Name" id="player-name">
        <Input />
      </Field>,
    );
    expect(screen.getByLabelText('Name')).toHaveAttribute('id', 'player-name');
  });

  it('wires textarea, select and file inputs too', () => {
    render(
      <>
        <Field label="Bio">
          <Textarea />
        </Field>
        <Field label="Team">
          <Select>
            <option>Under 12s</option>
          </Select>
        </Field>
        <Field label="Photo" error="Too big">
          <FileInput />
        </Field>
      </>,
    );
    expect(screen.getByLabelText('Bio').tagName).toBe('TEXTAREA');
    expect(screen.getByLabelText('Team').tagName).toBe('SELECT');
    const file = screen.getByLabelText('Photo');
    expect(file).toHaveAttribute('type', 'file');
    expect(file).toHaveAccessibleDescription('Too big');
  });

  it('works outside a Field with plain props', () => {
    render(<Input aria-label="Search" />);
    expect(screen.getByRole('textbox', { name: 'Search' })).toBeInTheDocument();
  });
});

describe('Checkbox', () => {
  it('toggles from its label', () => {
    render(<Checkbox label="Youth team" />);
    const box = screen.getByRole('checkbox', { name: 'Youth team' });
    fireEvent.click(screen.getByText('Youth team'));
    expect(box).toBeChecked();
  });
});

describe('fieldError', () => {
  it('reads a field message from an ApiError', () => {
    const err = new ApiError(422, 'invalid', { email: 'already in use' });
    expect(fieldError(err, 'email')).toBe('already in use');
    expect(fieldError(err, 'name')).toBeUndefined();
  });

  it('ignores other errors', () => {
    expect(fieldError(new Error('boom'), 'email')).toBeUndefined();
    expect(fieldError(null, 'email')).toBeUndefined();
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/components/ui/form.test.tsx`
Expected: FAIL. `./Checkbox` (and the others) can't be resolved.

- [ ] **Step 3: Implement the components**

`client/components/ui/fieldContext.ts`:

```ts
import { createContext, useContext } from 'react';

/** Props a Field hands to its control so label, help and error are wired up. */
export interface FieldControlProps {
  id: string;
  'aria-describedby'?: string;
  'aria-invalid'?: true;
}

export const FieldContext = createContext<FieldControlProps | null>(null);

export function useFieldControl(): FieldControlProps | null {
  return useContext(FieldContext);
}

export const controlClass =
  'w-full rounded-md border border-line bg-field px-2.5 py-2 text-sm text-ink placeholder:text-muted focus-visible:border-blue disabled:opacity-60 aria-[invalid=true]:border-red';
```

`client/components/ui/Field.tsx`:

```tsx
import { clsx } from 'clsx';
import { useId, type ReactNode } from 'react';

import { FieldContext, type FieldControlProps } from './fieldContext';

export interface FieldProps {
  label: string;
  help?: string;
  error?: string;
  id?: string;
  children: ReactNode;
  className?: string;
}

export function Field({ label, help, error, id, children, className }: FieldProps) {
  const autoId = useId();
  const controlId = id ?? autoId;
  const messageId = `${controlId}-message`;
  const message = error ?? help;
  const control: FieldControlProps = {
    id: controlId,
    'aria-describedby': message ? messageId : undefined,
    'aria-invalid': error ? true : undefined,
  };

  return (
    <div className={clsx('flex flex-col gap-1.5', className)}>
      <label htmlFor={controlId} className="text-sm font-semibold">
        {label}
      </label>
      <FieldContext.Provider value={control}>{children}</FieldContext.Provider>
      {message && (
        <p id={messageId} className={error ? 'text-sm text-red' : 'text-xs text-muted'}>
          {message}
        </p>
      )}
    </div>
  );
}
```

`client/components/ui/controls.tsx`:

```tsx
import { clsx } from 'clsx';
import type { ComponentProps } from 'react';

import { controlClass, useFieldControl } from './fieldContext';

export function Input({ className, ...props }: ComponentProps<'input'>) {
  const field = useFieldControl();
  return <input {...field} {...props} className={clsx(controlClass, className)} />;
}

export function Textarea({ className, ...props }: ComponentProps<'textarea'>) {
  const field = useFieldControl();
  return <textarea rows={4} {...field} {...props} className={clsx(controlClass, className)} />;
}

export function Select({ className, ...props }: ComponentProps<'select'>) {
  const field = useFieldControl();
  return <select {...field} {...props} className={clsx(controlClass, className)} />;
}

export function FileInput({ className, ...props }: Omit<ComponentProps<'input'>, 'type'>) {
  const field = useFieldControl();
  return (
    <input
      type="file"
      {...field}
      {...props}
      className={clsx(
        controlClass,
        'file:mr-3 file:rounded file:border-0 file:bg-surface file:px-2 file:py-1 file:text-ink',
        className,
      )}
    />
  );
}
```

`client/components/ui/Checkbox.tsx`:

```tsx
import { clsx } from 'clsx';
import type { ComponentProps, ReactNode } from 'react';

export interface CheckboxProps extends Omit<ComponentProps<'input'>, 'type'> {
  label: ReactNode;
}

export function Checkbox({ label, className, ...props }: CheckboxProps) {
  return (
    <label className={clsx('inline-flex items-center gap-2 text-sm', className)}>
      <input type="checkbox" className="size-4 accent-red" {...props} />
      {label}
    </label>
  );
}
```

`client/components/ui/fieldError.ts`:

```ts
import { ApiError } from '../../api/client';

/** The server's validation message for one form field, if the error carries one. */
export function fieldError(err: unknown, name: string): string | undefined {
  return err instanceof ApiError ? err.fields[name] : undefined;
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn vitest run client/components/ui/form.test.tsx`
Expected: PASS, 8 tests.

- [ ] **Step 5: Commit**

```bash
yarn eslint --fix client/components/ui
yarn typecheck
git add client/components/ui
git commit -q -m "Add form field, controls, checkbox and fieldError helper" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 5: Table primitives

**Files:**
- Create: `client/components/ui/Table.tsx`
- Test: `client/components/ui/Table.test.tsx`

**Interfaces:**
- Produces: `Table`, `THead`, `TBody`, `Tr`, `Th` and `Td`, each taking the native props of its element.

- [ ] **Step 1: Write the failing test**

Create `client/components/ui/Table.test.tsx`:

```tsx
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { Table, TBody, Td, Th, THead, Tr } from './Table';

describe('Table', () => {
  it('renders an accessible table inside a horizontal scroller', () => {
    const { container } = render(
      <Table>
        <THead>
          <Tr>
            <Th>Name</Th>
            <Th>Role</Th>
          </Tr>
        </THead>
        <TBody>
          <Tr>
            <Td>Jo Smith</Td>
            <Td>Manager</Td>
          </Tr>
        </TBody>
      </Table>,
    );
    expect(screen.getByRole('table')).toBeInTheDocument();
    expect(screen.getAllByRole('columnheader').map((th) => th.textContent)).toEqual(['Name', 'Role']);
    expect(screen.getByRole('cell', { name: 'Jo Smith' })).toBeInTheDocument();
    expect(container.firstElementChild).toHaveClass('overflow-x-auto');
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `yarn vitest run client/components/ui/Table.test.tsx`
Expected: FAIL. `./Table` can't be resolved.

- [ ] **Step 3: Implement**

`client/components/ui/Table.tsx`:

```tsx
import { clsx } from 'clsx';
import type { ComponentProps } from 'react';

export function Table({ className, ...props }: ComponentProps<'table'>) {
  return (
    <div className="w-full overflow-x-auto">
      <table className={clsx('w-full border-collapse text-sm', className)} {...props} />
    </div>
  );
}

export function THead(props: ComponentProps<'thead'>) {
  return <thead {...props} />;
}

export function TBody({ className, ...props }: ComponentProps<'tbody'>) {
  return <tbody className={clsx('[&>tr:hover]:bg-surface', className)} {...props} />;
}

export function Tr(props: ComponentProps<'tr'>) {
  return <tr {...props} />;
}

export function Th({ className, ...props }: ComponentProps<'th'>) {
  return (
    <th
      className={clsx(
        'border-b-2 border-line px-3 py-2 text-left text-[11px] font-bold tracking-wider text-muted uppercase',
        className,
      )}
      {...props}
    />
  );
}

export function Td({ className, ...props }: ComponentProps<'td'>) {
  return <td className={clsx('border-b border-line px-3 py-2.5', className)} {...props} />;
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `yarn vitest run client/components/ui/Table.test.tsx`
Expected: PASS, 1 test.

- [ ] **Step 5: Commit**

```bash
yarn eslint --fix client/components/ui/Table.tsx client/components/ui/Table.test.tsx
git add client/components/ui/Table.tsx client/components/ui/Table.test.tsx
git commit -q -m "Add table primitives" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 6: Modal and ConfirmDialog

**Files:**
- Create: `client/components/ui/Modal.tsx`, `client/components/ui/ConfirmDialog.tsx`, `client/test/dialog.ts`
- Modify: `client/test/setup.ts`
- Test: `client/components/ui/Modal.test.tsx`

**Interfaces:**
- Consumes: `Button` (Task 3).
- Produces:
  - `Modal({ open: boolean; onClose: () => void; title: string; children: ReactNode; actions?: ReactNode; className?: string })`
  - `ConfirmDialog({ open: boolean; title: string; message: ReactNode; confirmLabel: string; cancelLabel?: string; tone?: 'danger' | 'primary'; onConfirm: () => void | Promise<void>; onCancel: () => void })`

  `onConfirm` must handle its own errors; ConfirmDialog only shows the pending state.

- [ ] **Step 1: Add the jsdom dialog polyfill**

jsdom reflects `dialog.open` but has no `showModal` or `close`. Create `client/test/dialog.ts`:

```ts
/** jsdom has no showModal/close; emulate the parts our Modal relies on. */
export function installDialogPolyfill(): void {
  const proto = HTMLDialogElement.prototype;
  if (typeof proto.showModal !== 'function') {
    proto.showModal = function showModal(this: HTMLDialogElement) {
      this.setAttribute('open', '');
    };
  }
  if (typeof proto.show !== 'function') {
    proto.show = function show(this: HTMLDialogElement) {
      this.setAttribute('open', '');
    };
  }
  if (typeof proto.close !== 'function') {
    proto.close = function close(this: HTMLDialogElement) {
      if (!this.hasAttribute('open')) {
        return;
      }
      this.removeAttribute('open');
      this.dispatchEvent(new Event('close'));
    };
  }
}
```

In `client/test/setup.ts`, add `import { installDialogPolyfill } from './dialog';` next to the other local import, and call `installDialogPolyfill();` at the top level, straight after the imports.

- [ ] **Step 2: Write the failing tests**

Create `client/components/ui/Modal.test.tsx`:

```tsx
import { act, fireEvent, render, screen } from '@testing-library/react';
import { useState } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { ConfirmDialog } from './ConfirmDialog';
import { Modal } from './Modal';

function Harness({ onClose }: { onClose: () => void }) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <button onClick={() => setOpen(true)}>Open</button>
      <Modal
        open={open}
        onClose={() => {
          onClose();
          setOpen(false);
        }}
        title="Delete team?"
        actions={<button onClick={() => setOpen(false)}>Cancel</button>}
      >
        <p>Body text</p>
      </Modal>
    </>
  );
}

function openHarness() {
  const onClose = vi.fn();
  render(<Harness onClose={onClose} />);
  const opener = screen.getByRole('button', { name: 'Open' });
  opener.focus();
  fireEvent.click(opener);
  return { onClose, opener };
}

describe('Modal', () => {
  it('renders nothing inside the dialog while closed', () => {
    render(<Harness onClose={vi.fn()} />);
    expect(screen.queryByText('Body text')).toBeNull();
  });

  it('opens as a labelled dialog', () => {
    openHarness();
    const dialog = screen.getByRole('dialog', { name: 'Delete team?' });
    expect(dialog).toHaveAttribute('open');
    expect(screen.getByText('Body text')).toBeInTheDocument();
  });

  it('calls onClose once on Esc (the cancel event)', () => {
    const { onClose } = openHarness();
    fireEvent(screen.getByRole('dialog'), new Event('cancel', { cancelable: true }));
    expect(onClose).toHaveBeenCalledOnce();
    expect(screen.queryByText('Body text')).toBeNull();
  });

  it('closes on a backdrop click but not a click inside', () => {
    const { onClose } = openHarness();
    fireEvent.click(screen.getByText('Body text'));
    expect(onClose).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('dialog'));
    expect(onClose).toHaveBeenCalledOnce();
  });

  it('returns focus to the opener when closed by the parent', () => {
    const { opener, onClose } = openHarness();
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(screen.queryByText('Body text')).toBeNull();
    expect(document.activeElement).toBe(opener);
    expect(onClose).not.toHaveBeenCalled();
  });

  it('closes cleanly when unmounted while open', () => {
    const opener = document.createElement('button');
    document.body.append(opener);
    opener.focus();
    const { unmount } = render(
      <Modal open onClose={vi.fn()} title="Sign in">
        <p>Form</p>
      </Modal>,
    );
    expect(() => unmount()).not.toThrow();
    expect(document.activeElement).toBe(opener);
    opener.remove();
  });
});

describe('ConfirmDialog', () => {
  it('confirms and cancels', () => {
    const onConfirm = vi.fn();
    const onCancel = vi.fn();
    render(
      <ConfirmDialog
        open
        title="Sign out?"
        message="You will need to sign in again."
        confirmLabel="Sign out"
        onConfirm={onConfirm}
        onCancel={onCancel}
      />,
    );
    expect(screen.getByRole('dialog', { name: 'Sign out?' })).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Sign out' }));
    expect(onConfirm).toHaveBeenCalledOnce();
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(onCancel).toHaveBeenCalledOnce();
  });

  it('shows a pending state until onConfirm settles', async () => {
    let finish: () => void = () => {};
    const onConfirm = () =>
      new Promise<void>((resolve) => {
        finish = resolve;
      });
    render(
      <ConfirmDialog
        open
        title="Delete team?"
        message="This can't be undone."
        confirmLabel="Delete"
        tone="danger"
        onConfirm={onConfirm}
        onCancel={vi.fn()}
      />,
    );
    const confirm = screen.getByRole('button', { name: 'Delete' });
    fireEvent.click(confirm);
    expect(confirm).toHaveAttribute('aria-busy', 'true');
    expect(confirm).toBeDisabled();
    await act(async () => finish());
    expect(confirm).not.toHaveAttribute('aria-busy');
  });
});
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `yarn vitest run client/components/ui/Modal.test.tsx`
Expected: FAIL. `./ConfirmDialog` and `./Modal` can't be resolved.

- [ ] **Step 4: Implement**

`client/components/ui/Modal.tsx`:

```tsx
import { clsx } from 'clsx';
import { useEffect, useId, useRef, type ReactNode } from 'react';

export interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
  actions?: ReactNode;
  className?: string;
}

/**
 * A native <dialog> shown with showModal(), so the browser traps focus and
 * handles Esc. The parent owns `open`; Esc and backdrop clicks ask it to close.
 */
export function Modal({ open, onClose, title, children, actions, className }: ModalProps) {
  const ref = useRef<HTMLDialogElement>(null);
  const titleId = useId();
  // True while this component wants the dialog open; stops our own close() from
  // echoing back as a second onClose.
  const wantOpen = useRef(false);

  useEffect(() => {
    const dialog = ref.current;
    if (!open || !dialog) {
      return;
    }
    const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    wantOpen.current = true;
    if (!dialog.open) {
      dialog.showModal();
    }
    return () => {
      wantOpen.current = false;
      if (dialog.open) {
        dialog.close();
      }
      previous?.focus();
    };
  }, [open]);

  return (
    <dialog
      ref={ref}
      aria-labelledby={titleId}
      onCancel={(e) => {
        e.preventDefault();
        onClose();
      }}
      onClose={() => {
        if (wantOpen.current) {
          wantOpen.current = false;
          onClose();
        }
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) {
          onClose();
        }
      }}
      className={clsx(
        'm-auto w-[min(28rem,calc(100vw-2rem))] rounded-xl border border-line bg-bg p-0 text-ink shadow-2xl backdrop:bg-black/50',
        className,
      )}
    >
      {open && (
        <div>
          <div className="border-b border-line px-4 py-3">
            <h2
              id={titleId}
              className="font-display text-xl font-extrabold tracking-wide uppercase"
            >
              {title}
            </h2>
          </div>
          <div className="px-4 py-4">{children}</div>
          {actions && (
            <div className="flex justify-end gap-2 border-t border-line px-4 py-3">{actions}</div>
          )}
        </div>
      )}
    </dialog>
  );
}
```

`client/components/ui/ConfirmDialog.tsx`:

```tsx
import { useState, type ReactNode } from 'react';

import { Button } from './Button';
import { Modal } from './Modal';

export interface ConfirmDialogProps {
  open: boolean;
  title: string;
  message: ReactNode;
  confirmLabel: string;
  cancelLabel?: string;
  tone?: 'danger' | 'primary';
  /** Must handle its own errors; the dialog only shows the pending state. */
  onConfirm: () => void | Promise<void>;
  onCancel: () => void;
}

export function ConfirmDialog({
  open,
  title,
  message,
  confirmLabel,
  cancelLabel = 'Cancel',
  tone = 'primary',
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  const [busy, setBusy] = useState(false);

  async function confirm() {
    setBusy(true);
    try {
      await onConfirm();
    } finally {
      setBusy(false);
    }
  }

  return (
    <Modal
      open={open}
      onClose={() => {
        if (!busy) {
          onCancel();
        }
      }}
      title={title}
      actions={
        <>
          <Button variant="secondary" onClick={onCancel} disabled={busy}>
            {cancelLabel}
          </Button>
          <Button variant={tone === 'danger' ? 'danger' : 'primary'} loading={busy} onClick={confirm}>
            {confirmLabel}
          </Button>
        </>
      }
    >
      <div className="text-sm">{message}</div>
    </Modal>
  );
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `yarn vitest run client/components/ui/Modal.test.tsx`
Expected: PASS, 8 tests. If "calls onClose once on Esc" sees two calls, the `wantOpen` guard is being bypassed. Debug it; don't loosen the assertion.

- [ ] **Step 6: Commit**

```bash
yarn eslint --fix client/components/ui client/test
yarn typecheck
git add client/components/ui/Modal.tsx client/components/ui/ConfirmDialog.tsx client/components/ui/Modal.test.tsx client/test
git commit -q -m "Add native dialog Modal and ConfirmDialog" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 7: Disclosure, dismiss and Menu

**Files:**
- Create: `client/components/ui/useDisclosure.ts`, `client/components/ui/useDismiss.ts`, `client/components/ui/Menu.tsx`
- Test: `client/components/ui/Menu.test.tsx`

**Interfaces:**
- Consumes: `buttonClasses` (Task 2).
- Produces:
  - `useDisclosure(initial?: boolean): { open: boolean; setOpen(v: boolean): void; toggle(): void; close(): void }`
  - `useDismiss(ref: RefObject<HTMLElement | null>, open: boolean, onDismiss: () => void): void`, which dismisses on a mousedown outside `ref` or on Esc anywhere
  - `interface MenuItem { label: string; to?: string; href?: string; onSelect?: () => void }`
  - `Menu({ label: ReactNode; items: MenuItem[]; align?: 'start' | 'end'; triggerClassName?: string; triggerLabel?: string })`

- [ ] **Step 1: Write the failing tests**

Create `client/components/ui/Menu.test.tsx`:

```tsx
import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { describe, expect, it, vi } from 'vitest';

import { Menu } from './Menu';

function renderMenu(onSelect = vi.fn()) {
  render(
    <MemoryRouter>
      <Menu
        label="Jo Smith"
        items={[
          { label: 'Players', href: '/players' },
          { label: 'Design', to: '/design' },
          { label: 'Sign out', onSelect },
        ]}
      />
      <p>Outside</p>
    </MemoryRouter>,
  );
  return { trigger: screen.getByRole('button', { name: 'Jo Smith' }), onSelect };
}

describe('Menu', () => {
  it('starts closed', () => {
    const { trigger } = renderMenu();
    expect(trigger).toHaveAttribute('aria-haspopup', 'menu');
    expect(trigger).toHaveAttribute('aria-expanded', 'false');
    expect(screen.queryByRole('menu')).toBeNull();
  });

  it('opens and focuses the first item', () => {
    const { trigger } = renderMenu();
    fireEvent.click(trigger);
    expect(trigger).toHaveAttribute('aria-expanded', 'true');
    expect(screen.getByRole('menu')).toBeInTheDocument();
    expect(document.activeElement).toBe(screen.getByRole('menuitem', { name: 'Players' }));
  });

  it('renders legacy links, router links and buttons', () => {
    const { trigger } = renderMenu();
    fireEvent.click(trigger);
    expect(screen.getByRole('menuitem', { name: 'Players' })).toHaveAttribute('href', '/players');
    expect(screen.getByRole('menuitem', { name: 'Design' })).toHaveAttribute('href', '/design');
    expect(screen.getByRole('menuitem', { name: 'Sign out' }).tagName).toBe('BUTTON');
  });

  it('moves focus with the arrow keys and wraps', () => {
    const { trigger } = renderMenu();
    fireEvent.click(trigger);
    const menu = screen.getByRole('menu');
    fireEvent.keyDown(menu, { key: 'ArrowDown' });
    expect(document.activeElement).toBe(screen.getByRole('menuitem', { name: 'Design' }));
    fireEvent.keyDown(menu, { key: 'ArrowUp' });
    fireEvent.keyDown(menu, { key: 'ArrowUp' });
    expect(document.activeElement).toBe(screen.getByRole('menuitem', { name: 'Sign out' }));
  });

  it('closes on Esc and returns focus to the trigger', () => {
    const { trigger } = renderMenu();
    fireEvent.click(trigger);
    fireEvent.keyDown(screen.getByRole('menu'), { key: 'Escape' });
    expect(screen.queryByRole('menu')).toBeNull();
    expect(document.activeElement).toBe(trigger);
  });

  it('closes on an outside click', () => {
    const { trigger } = renderMenu();
    fireEvent.click(trigger);
    fireEvent.mouseDown(screen.getByText('Outside'));
    expect(screen.queryByRole('menu')).toBeNull();
  });

  it('runs onSelect and closes', () => {
    const { trigger, onSelect } = renderMenu();
    fireEvent.click(trigger);
    fireEvent.click(screen.getByRole('menuitem', { name: 'Sign out' }));
    expect(onSelect).toHaveBeenCalledOnce();
    expect(screen.queryByRole('menu')).toBeNull();
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/components/ui/Menu.test.tsx`
Expected: FAIL. `./Menu` can't be resolved.

- [ ] **Step 3: Implement**

`client/components/ui/useDisclosure.ts`:

```ts
import { useCallback, useState } from 'react';

export function useDisclosure(initial = false) {
  const [open, setOpen] = useState(initial);
  const toggle = useCallback(() => setOpen((o) => !o), []);
  const close = useCallback(() => setOpen(false), []);
  return { open, setOpen, toggle, close };
}
```

`client/components/ui/useDismiss.ts`:

```ts
import { useEffect, useRef, type RefObject } from 'react';

/** While open, calls onDismiss on a mousedown outside ref or on Esc anywhere. */
export function useDismiss(
  ref: RefObject<HTMLElement | null>,
  open: boolean,
  onDismiss: () => void,
): void {
  const callback = useRef(onDismiss);
  useEffect(() => {
    callback.current = onDismiss;
  });

  useEffect(() => {
    if (!open) {
      return;
    }
    function onPointer(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        callback.current();
      }
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') {
        callback.current();
      }
    }
    document.addEventListener('mousedown', onPointer);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onPointer);
      document.removeEventListener('keydown', onKey);
    };
  }, [open, ref]);
}
```

`client/components/ui/Menu.tsx`:

```tsx
import { clsx } from 'clsx';
import { useEffect, useId, useRef, type KeyboardEvent, type ReactNode } from 'react';
import { Link } from 'react-router';

import { buttonClasses } from './buttonStyles';
import { useDisclosure } from './useDisclosure';
import { useDismiss } from './useDismiss';

export interface MenuItem {
  label: string;
  /** An SPA route. */
  to?: string;
  /** A full-page URL (legacy page). */
  href?: string;
  onSelect?: () => void;
}

export interface MenuProps {
  label: ReactNode;
  items: MenuItem[];
  align?: 'start' | 'end';
  triggerClassName?: string;
  /** Accessible name when `label` is not plain text. */
  triggerLabel?: string;
}

const itemClass =
  'block w-full px-3 py-2 text-left text-sm text-ink hover:bg-surface focus-visible:bg-surface focus-visible:outline-none';

export function Menu({ label, items, align = 'end', triggerClassName, triggerLabel }: MenuProps) {
  const { open, setOpen, close } = useDisclosure();
  const rootRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const menuId = useId();

  useDismiss(rootRef, open, close);

  useEffect(() => {
    if (open) {
      listRef.current?.querySelector<HTMLElement>('[role="menuitem"]')?.focus();
    }
  }, [open]);

  function onKeyDown(e: KeyboardEvent<HTMLDivElement>) {
    const els = Array.from(listRef.current?.querySelectorAll<HTMLElement>('[role="menuitem"]') ?? []);
    const i = els.indexOf(document.activeElement as HTMLElement);
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      els[(i + 1) % els.length]?.focus();
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      els[i <= 0 ? els.length - 1 : i - 1]?.focus();
    } else if (e.key === 'Escape') {
      e.preventDefault();
      close();
      triggerRef.current?.focus();
    } else if (e.key === 'Tab') {
      close();
    }
  }

  return (
    <div ref={rootRef} className="relative">
      <button
        ref={triggerRef}
        type="button"
        aria-haspopup="menu"
        aria-expanded={open}
        aria-controls={open ? menuId : undefined}
        aria-label={triggerLabel}
        onClick={() => setOpen(!open)}
        className={triggerClassName ?? buttonClasses('secondary', 'sm')}
      >
        {label}
      </button>
      {open && (
        <div
          ref={listRef}
          id={menuId}
          role="menu"
          onKeyDown={onKeyDown}
          className={clsx(
            'absolute top-full z-40 mt-1 min-w-44 overflow-hidden rounded-lg border border-line bg-bg py-1 shadow-lg',
            align === 'end' ? 'right-0' : 'left-0',
          )}
        >
          {items.map((item) => {
            const onClick = () => {
              close();
              item.onSelect?.();
            };
            if (item.to !== undefined) {
              return (
                <Link
                  key={item.label}
                  role="menuitem"
                  tabIndex={-1}
                  to={item.to}
                  onClick={onClick}
                  className={itemClass}
                >
                  {item.label}
                </Link>
              );
            }
            if (item.href !== undefined) {
              return (
                <a
                  key={item.label}
                  role="menuitem"
                  tabIndex={-1}
                  href={item.href}
                  onClick={onClick}
                  className={itemClass}
                >
                  {item.label}
                </a>
              );
            }
            return (
              <button
                key={item.label}
                type="button"
                role="menuitem"
                tabIndex={-1}
                onClick={onClick}
                className={itemClass}
              >
                {item.label}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn vitest run client/components/ui/Menu.test.tsx`
Expected: PASS, 7 tests.

- [ ] **Step 5: Commit**

```bash
yarn eslint --fix client/components/ui
yarn typecheck
git add client/components/ui/useDisclosure.ts client/components/ui/useDismiss.ts client/components/ui/Menu.tsx client/components/ui/Menu.test.tsx
git commit -q -m "Add useDisclosure, useDismiss and accessible Menu" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 8: Toasts

**Files:**
- Create: `client/components/ui/toast/context.ts`, `client/components/ui/toast/ToastProvider.tsx`, `client/components/ui/toast/useToast.ts`
- Modify: `client/test/render.tsx`, `client/main.tsx`
- Test: `client/components/ui/toast/Toast.test.tsx`

**Interfaces:**
- Produces:
  - `type ToastTone = 'success' | 'error' | 'info'`
  - `interface ToastOptions { message: string; tone?: ToastTone }`
  - `useToast(): { show(options: ToastOptions): void }`
  - `<ToastProvider>`
  - `TOAST_DURATION_MS = 5000`
  - `MAX_TOASTS = 3`
  - `renderWithProviders` now wraps content in `ToastProvider` inside `AuthProvider`.

- [ ] **Step 1: Write the failing tests**

Create `client/components/ui/toast/Toast.test.tsx`:

```tsx
import { act, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { ToastProvider } from './ToastProvider';
import { useToast } from './useToast';

function Probe() {
  const toast = useToast();
  return (
    <>
      {['1', '2', '3', '4'].map((n) => (
        <button key={n} onClick={() => toast.show({ message: `Toast ${n}`, tone: 'success' })}>
          show {n}
        </button>
      ))}
    </>
  );
}

function renderProbe() {
  return render(
    <ToastProvider>
      <Probe />
    </ToastProvider>,
  );
}

describe('Toast', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('shows a toast in a polite live region', () => {
    const { container } = renderProbe();
    fireEvent.click(screen.getByRole('button', { name: 'show 1' }));
    const toast = screen.getByRole('button', { name: 'Toast 1' });
    expect(toast.closest('[aria-live="polite"]')).not.toBeNull();
    expect(container.ownerDocument.querySelector('[aria-live="polite"]')).toBeInTheDocument();
  });

  it('dismisses itself after 5 seconds', () => {
    renderProbe();
    fireEvent.click(screen.getByRole('button', { name: 'show 1' }));
    act(() => vi.advanceTimersByTime(4999));
    expect(screen.getByRole('button', { name: 'Toast 1' })).toBeInTheDocument();
    act(() => vi.advanceTimersByTime(1));
    expect(screen.queryByRole('button', { name: 'Toast 1' })).toBeNull();
  });

  it('dismisses on click', () => {
    renderProbe();
    fireEvent.click(screen.getByRole('button', { name: 'show 1' }));
    fireEvent.click(screen.getByRole('button', { name: 'Toast 1' }));
    expect(screen.queryByRole('button', { name: 'Toast 1' })).toBeNull();
  });

  it('keeps at most three, dropping the oldest', () => {
    renderProbe();
    for (const n of ['1', '2', '3', '4']) {
      fireEvent.click(screen.getByRole('button', { name: `show ${n}` }));
    }
    expect(screen.queryByRole('button', { name: 'Toast 1' })).toBeNull();
    for (const n of ['2', '3', '4']) {
      expect(screen.getByRole('button', { name: `Toast ${n}` })).toBeInTheDocument();
    }
  });

  it('throws a clear error outside the provider', () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
    expect(() => render(<Probe />)).toThrow('useToast must be used inside <ToastProvider>');
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/components/ui/toast`
Expected: FAIL. `./ToastProvider` can't be resolved.

- [ ] **Step 3: Implement**

`client/components/ui/toast/context.ts`:

```ts
import { createContext } from 'react';

export type ToastTone = 'success' | 'error' | 'info';

export interface ToastOptions {
  message: string;
  tone?: ToastTone;
}

export interface ToastApi {
  show: (options: ToastOptions) => void;
}

export const TOAST_DURATION_MS = 5000;
export const MAX_TOASTS = 3;

export const ToastContext = createContext<ToastApi | null>(null);
```

`client/components/ui/toast/useToast.ts`:

```ts
import { useContext } from 'react';

import { ToastContext, type ToastApi } from './context';

export function useToast(): ToastApi {
  const ctx = useContext(ToastContext);
  if (!ctx) {
    throw new Error('useToast must be used inside <ToastProvider>');
  }
  return ctx;
}
```

`client/components/ui/toast/ToastProvider.tsx`:

```tsx
import { clsx } from 'clsx';
import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from 'react';

import {
  MAX_TOASTS,
  TOAST_DURATION_MS,
  ToastContext,
  type ToastApi,
  type ToastOptions,
  type ToastTone,
} from './context';

interface Toast {
  id: number;
  message: string;
  tone: ToastTone;
}

const dots: Record<ToastTone, string> = {
  success: 'bg-success',
  error: 'bg-red',
  info: 'bg-blue',
};

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const nextId = useRef(1);
  const timers = useRef(new Map<number, ReturnType<typeof setTimeout>>());

  const dismiss = useCallback((id: number) => {
    const timer = timers.current.get(id);
    if (timer !== undefined) {
      clearTimeout(timer);
      timers.current.delete(id);
    }
    setToasts((ts) => ts.filter((t) => t.id !== id));
  }, []);

  const show = useCallback(
    ({ message, tone = 'info' }: ToastOptions) => {
      const id = nextId.current++;
      setToasts((ts) => [...ts, { id, message, tone }].slice(-MAX_TOASTS));
      timers.current.set(
        id,
        setTimeout(() => dismiss(id), TOAST_DURATION_MS),
      );
    },
    [dismiss],
  );

  useEffect(() => {
    const pending = timers.current;
    return () => pending.forEach((timer) => clearTimeout(timer));
  }, []);

  const api = useMemo<ToastApi>(() => ({ show }), [show]);

  return (
    <ToastContext.Provider value={api}>
      {children}
      <div
        aria-live="polite"
        className="pointer-events-none fixed right-4 bottom-4 z-50 flex w-80 max-w-[calc(100vw-2rem)] flex-col gap-2"
      >
        {toasts.map((t) => (
          <button
            key={t.id}
            type="button"
            title="Dismiss"
            onClick={() => dismiss(t.id)}
            className="pointer-events-auto flex items-center gap-2.5 rounded-lg bg-ink px-3.5 py-2.5 text-left text-sm text-bg shadow-lg"
          >
            <span aria-hidden="true" className={clsx('size-2 shrink-0 rounded-full', dots[t.tone])} />
            {t.message}
          </button>
        ))}
      </div>
    </ToastContext.Provider>
  );
}
```

`client/test/render.tsx`: import `ToastProvider` from `../components/ui/toast/ToastProvider` and wrap it as `<AuthProvider><ToastProvider>{ui}</ToastProvider></AuthProvider>`.

`client/main.tsx`: import `ToastProvider` the same way and wrap it as `<AuthProvider><ToastProvider><App /></ToastProvider></AuthProvider>`.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn vitest run client/components/ui/toast && yarn test:client`
Expected: PASS. 5 toast tests pass and the whole suite is green.

- [ ] **Step 5: Commit**

```bash
yarn eslint --fix client/components/ui/toast client/test client/main.tsx
yarn typecheck
git add client/components/ui/toast client/test/render.tsx client/main.tsx
git commit -q -m "Add toast provider and useToast" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 9: Sign-in dialog and account control

**Files:**
- Create: `client/api/auth.ts`, `client/lib/navigation.ts`, `client/components/layout/SignInDialog.tsx`, `client/components/layout/AccountControl.tsx`
- Modify: `client/api/types.ts`, `client/test/mockFetch.ts`
- Test: `client/components/layout/AccountControl.test.tsx`

**Interfaces:**
- Consumes:
  - `apiFetch` and `ApiError` (from `client/api/client.ts`);
  - `queryKeys.me` (from `client/api/queries.ts`);
  - `useAuth()` returning `{ user, isLoading, refresh }`;
  - `Modal`, `ConfirmDialog`, `Button`, `Field`, `Input`, `Checkbox`, `Alert`, `Menu`/`MenuItem`, `useDisclosure` and `useToast` (Tasks 3–8).
- Produces:
  - `LoginResponse`;
  - `login(input: LoginInput): Promise<LoginResponse>`;
  - `logout(): Promise<void>`;
  - `goTo(url: string): void`;
  - `SignInDialog({ open: boolean; onClose: () => void })`;
  - `AccountControl()`.

  `mockFetch` routes may now be functions returning `MockResponse | Promise<MockResponse>`.

- [ ] **Step 1: Extend the test fetch mock**

Replace `client/test/mockFetch.ts` with the version below. Routes can now be computed per call, including promises. Existing callers still work: a throwing function behaves as before.

```ts
import { vi } from 'vitest';

export interface MockResponse {
  status?: number;
  body?: unknown; // objects are JSON-encoded; strings are sent as-is
  contentType?: string;
}

export type MockRoute = MockResponse | (() => MockResponse | Promise<MockResponse>);

/**
 * Replaces global fetch with a stub keyed by path (e.g. '/api/v1/site').
 * A route may be a function, evaluated per request (it can throw or return a
 * promise). Unknown paths fail the test loudly with a 599.
 */
export function mockFetch(routes: Record<string, MockRoute>) {
  const fn = vi.fn(async (...args: [RequestInfo | URL, RequestInit?]) => {
    const [input] = args;
    const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url;
    const path = url.replace(/^https?:\/\/[^/]+/, '');
    const entry = routes[path];
    if (!entry) {
      return new Response(`no mock for ${path}`, { status: 599 });
    }
    const route = typeof entry === 'function' ? await entry() : entry;
    const status = route.status ?? 200;
    if (status === 204) {
      return new Response(null, { status });
    }
    const isText = typeof route.body === 'string';
    return new Response(
      route.body === undefined ? '' : isText ? (route.body as string) : JSON.stringify(route.body),
      {
        status,
        headers: {
          'Content-Type': route.contentType ?? (isText ? 'text/html' : 'application/json'),
        },
      },
    );
  });
  vi.stubGlobal('fetch', fn);
  return fn;
}
```

Run: `yarn test:client`
Expected: PASS. Existing tests are unaffected.

- [ ] **Step 2: Write the failing tests**

Create `client/components/layout/AccountControl.test.tsx`:

```tsx
import { act, fireEvent, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { CurrentUser } from '../../api/types';
import { goTo } from '../../lib/navigation';
import { mockFetch, type MockResponse } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import { AccountControl } from './AccountControl';

vi.mock('../../lib/navigation', () => ({ goTo: vi.fn() }));

const anonymous: MockResponse = {
  status: 401,
  body: { error: { code: 401, message: 'login required' } },
};

function user(name: string, role: string, perms: Partial<CurrentUser['permissions']> = {}) {
  return {
    id: 7,
    name,
    email: 'someone@example.test',
    role,
    permissions: { canEdit: false, canManageGallery: false, canManageUsers: false, ...perms },
  };
}

function openSignIn() {
  fireEvent.click(screen.getByRole('button', { name: 'Sign in' }));
  const dialog = screen.getByRole('dialog', { name: 'Sign in' });
  fireEvent.change(within(dialog).getByLabelText('Email'), {
    target: { value: 'jo@example.test' },
  });
  fireEvent.change(within(dialog).getByLabelText('Password'), {
    target: { value: 'hunter2' },
  });
  return dialog;
}

beforeEach(() => {
  vi.mocked(goTo).mockClear();
});

describe('AccountControl signed out', () => {
  it('offers Sign in and focuses Email when the dialog opens', async () => {
    mockFetch({ '/api/v1/auth/me': anonymous });
    renderWithProviders(<AccountControl />);
    fireEvent.click(await screen.findByRole('button', { name: 'Sign in' }));
    const dialog = screen.getByRole('dialog', { name: 'Sign in' });
    expect(document.activeElement).toBe(within(dialog).getByLabelText('Email'));
  });

  it('signs in, closes, refreshes and says who is signed in', async () => {
    let me: MockResponse = anonymous;
    const fetchMock = mockFetch({
      '/api/v1/auth/me': () => me,
      '/api/v1/auth/login': () => {
        me = { body: user('Jo Smith', 'Manager') };
        return { body: { user: user('Jo Smith', 'Manager'), resetRequired: false } };
      },
    });
    renderWithProviders(<AccountControl />);
    await screen.findByRole('button', { name: 'Sign in' });
    const dialog = openSignIn();
    fireEvent.click(within(dialog).getByLabelText('Remember me'));
    fireEvent.click(within(dialog).getByRole('button', { name: 'Sign in' }));

    expect(await screen.findByRole('button', { name: /Jo Smith/ })).toBeInTheDocument();
    expect(screen.queryByRole('dialog')).toBeNull();
    expect(screen.getByRole('button', { name: 'Signed in as Jo Smith' })).toBeInTheDocument();
    const loginCall = fetchMock.mock.calls.find(([url]) => String(url).endsWith('/auth/login'));
    expect(JSON.parse(String(loginCall?.[1]?.body))).toEqual({
      email: 'jo@example.test',
      password: 'hunter2',
      remember: true,
    });
  });

  it('sends reset-flagged accounts to the reset page', async () => {
    mockFetch({
      '/api/v1/auth/me': anonymous,
      '/api/v1/auth/login': { body: { resetRequired: true, resetUrl: '/reset/abc123' } },
    });
    renderWithProviders(<AccountControl />);
    await screen.findByRole('button', { name: 'Sign in' });
    const dialog = openSignIn();
    fireEvent.click(within(dialog).getByRole('button', { name: 'Sign in' }));
    await vi.waitFor(() => expect(goTo).toHaveBeenCalledWith('/reset/abc123'));
  });

  it('explains a wrong password, keeps the email and clears the password', async () => {
    mockFetch({
      '/api/v1/auth/me': anonymous,
      '/api/v1/auth/login': {
        status: 401,
        body: { error: { code: 401, message: 'invalid credentials' } },
      },
    });
    renderWithProviders(<AccountControl />);
    await screen.findByRole('button', { name: 'Sign in' });
    const dialog = openSignIn();
    fireEvent.click(within(dialog).getByRole('button', { name: 'Sign in' }));
    expect(await within(dialog).findByRole('alert')).toHaveTextContent(
      'Incorrect email or password.',
    );
    expect(within(dialog).getByLabelText('Email')).toHaveValue('jo@example.test');
    expect(within(dialog).getByLabelText('Password')).toHaveValue('');
  });

  it('shows the loading state while signing in', async () => {
    let answer: (r: MockResponse) => void = () => {};
    mockFetch({
      '/api/v1/auth/me': anonymous,
      '/api/v1/auth/login': () =>
        new Promise<MockResponse>((resolve) => {
          answer = resolve;
        }),
    });
    renderWithProviders(<AccountControl />);
    await screen.findByRole('button', { name: 'Sign in' });
    const dialog = openSignIn();
    const submit = within(dialog).getByRole('button', { name: 'Sign in' });
    fireEvent.click(submit);
    await vi.waitFor(() => expect(submit).toHaveAttribute('aria-busy', 'true'));
    await act(async () =>
      answer({ status: 401, body: { error: { code: 401, message: 'invalid credentials' } } }),
    );
    await vi.waitFor(() => expect(submit).not.toHaveAttribute('aria-busy'));
  });
});

describe('AccountControl signed in', () => {
  async function openMenuFor(body: ReturnType<typeof user>) {
    mockFetch({ '/api/v1/auth/me': { body } });
    renderWithProviders(<AccountControl />);
    fireEvent.click(await screen.findByRole('button', { name: new RegExp(body.name) }));
    return screen.getAllByRole('menuitem').map((el) => el.textContent);
  }

  it('gives a Manager Players, Account and Sign out', async () => {
    expect(await openMenuFor(user('Mo Manager', 'Manager'))).toEqual([
      'Players',
      'Account',
      'Sign out',
    ]);
  });

  it('adds Edit info for editors', async () => {
    const items = await openMenuFor(user('Ed Editor', 'Treasurer', { canEdit: true }));
    expect(items).toEqual(['Players', 'Account', 'Edit info', 'Sign out']);
    expect(screen.getByRole('menuitem', { name: 'Edit info' })).toHaveAttribute(
      'href',
      '/info/edit',
    );
  });

  it('adds Users for user admins', async () => {
    const items = await openMenuFor(
      user('Sam Sec', 'Club Secretary', { canEdit: true, canManageUsers: true }),
    );
    expect(items).toEqual(['Players', 'Account', 'Edit info', 'Users', 'Sign out']);
    expect(screen.getByRole('menuitem', { name: 'Users' })).toHaveAttribute('href', '/users');
  });

  it('signs out after confirmation', async () => {
    let me: MockResponse = { body: user('Jo Smith', 'Manager') };
    const fetchMock = mockFetch({
      '/api/v1/auth/me': () => me,
      '/api/v1/auth/logout': () => {
        me = anonymous;
        return { status: 204 };
      },
    });
    renderWithProviders(<AccountControl />);
    fireEvent.click(await screen.findByRole('button', { name: /Jo Smith/ }));
    fireEvent.click(screen.getByRole('menuitem', { name: 'Sign out' }));
    const confirm = screen.getByRole('dialog', { name: 'Sign out?' });
    fireEvent.click(within(confirm).getByRole('button', { name: 'Sign out' }));

    expect(await screen.findByRole('button', { name: 'Sign in' })).toBeInTheDocument();
    const logoutCall = fetchMock.mock.calls.find(([url]) => String(url).endsWith('/auth/logout'));
    expect(logoutCall?.[1]?.method).toBe('POST');
  });

  it('treats a 401 from logout as already signed out', async () => {
    let me: MockResponse = { body: user('Jo Smith', 'Manager') };
    mockFetch({
      '/api/v1/auth/me': () => me,
      '/api/v1/auth/logout': () => {
        me = anonymous;
        return { status: 401, body: { error: { code: 401, message: 'login required' } } };
      },
    });
    renderWithProviders(<AccountControl />);
    fireEvent.click(await screen.findByRole('button', { name: /Jo Smith/ }));
    fireEvent.click(screen.getByRole('menuitem', { name: 'Sign out' }));
    fireEvent.click(
      within(screen.getByRole('dialog', { name: 'Sign out?' })).getByRole('button', {
        name: 'Sign out',
      }),
    );
    expect(await screen.findByRole('button', { name: 'Sign in' })).toBeInTheDocument();
  });
});
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `yarn vitest run client/components/layout/AccountControl.test.tsx`
Expected: FAIL. `../../lib/navigation` and `./AccountControl` can't be resolved.

- [ ] **Step 4: Implement**

Append to `client/api/types.ts`:

```ts
/** auth.LoginResponse — POST /auth/login */
export interface LoginResponse {
  user?: CurrentUser;
  resetRequired: boolean;
  resetUrl?: string;
}
```

`client/api/auth.ts`:

```ts
import { apiFetch } from './client';
import type { LoginResponse } from './types';

export interface LoginInput {
  email: string;
  password: string;
  remember: boolean;
}

export function login(input: LoginInput): Promise<LoginResponse> {
  return apiFetch<LoginResponse>('/auth/login', { json: input });
}

export function logout(): Promise<void> {
  return apiFetch<void>('/auth/logout', { method: 'POST' });
}
```

`client/lib/navigation.ts`:

```ts
/** Full-page navigation (e.g. to a legacy page). A module so tests can mock it. */
export function goTo(url: string): void {
  window.location.assign(url);
}
```

`client/components/layout/SignInDialog.tsx`:

```tsx
import { useState, type FormEvent } from 'react';

import { login } from '../../api/auth';
import { ApiError } from '../../api/client';
import { useAuth } from '../../auth/useAuth';
import { goTo } from '../../lib/navigation';
import { Alert } from '../ui/Alert';
import { Button } from '../ui/Button';
import { Checkbox } from '../ui/Checkbox';
import { Input } from '../ui/controls';
import { Field } from '../ui/Field';
import { Modal } from '../ui/Modal';
import { useToast } from '../ui/toast/useToast';

export function SignInDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { refresh } = useAuth();
  const toast = useToast();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [remember, setRemember] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      const res = await login({ email, password, remember });
      if (res.resetRequired && res.resetUrl) {
        goTo(res.resetUrl);
        return;
      }
      setPassword('');
      onClose();
      if (res.user) {
        toast.show({ tone: 'success', message: `Signed in as ${res.user.name}` });
      }
      await refresh();
    } catch (err) {
      setPassword('');
      if (err instanceof ApiError && err.status === 401) {
        setError('Incorrect email or password.');
      } else {
        setError(err instanceof Error ? err.message : 'Sign-in failed. Please try again.');
      }
    } finally {
      setBusy(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Sign in">
      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        {error && <Alert tone="error">{error}</Alert>}
        <Field label="Email">
          <Input
            type="email"
            name="email"
            autoComplete="username"
            required
            autoFocus
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </Field>
        <Field label="Password">
          <Input
            type="password"
            name="password"
            autoComplete="current-password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </Field>
        <Checkbox
          label="Remember me"
          checked={remember}
          onChange={(e) => setRemember(e.target.checked)}
        />
        <div className="flex justify-end gap-2 border-t border-line pt-3">
          <Button variant="secondary" onClick={onClose} disabled={busy}>
            Cancel
          </Button>
          <Button type="submit" loading={busy}>
            Sign in
          </Button>
        </div>
      </form>
    </Modal>
  );
}
```

`client/components/layout/AccountControl.tsx`:

```tsx
import { useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';

import { logout } from '../../api/auth';
import { ApiError } from '../../api/client';
import { queryKeys } from '../../api/queries';
import { useAuth } from '../../auth/useAuth';
import { Button } from '../ui/Button';
import { buttonClasses } from '../ui/buttonStyles';
import { ConfirmDialog } from '../ui/ConfirmDialog';
import { Menu, type MenuItem } from '../ui/Menu';
import { useToast } from '../ui/toast/useToast';
import { useDisclosure } from '../ui/useDisclosure';
import { SignInDialog } from './SignInDialog';

export function AccountControl() {
  const { user, isLoading } = useAuth();
  const signIn = useDisclosure();
  const [confirmSignOut, setConfirmSignOut] = useState(false);
  const queryClient = useQueryClient();
  const toast = useToast();

  if (isLoading) {
    return <span aria-hidden="true" className="inline-block h-8 w-20" />;
  }

  if (!user) {
    return (
      <>
        <Button variant="secondary" size="sm" onClick={() => signIn.setOpen(true)}>
          Sign in
        </Button>
        <SignInDialog open={signIn.open} onClose={signIn.close} />
      </>
    );
  }

  async function signOut() {
    try {
      await logout();
    } catch (err) {
      // A 401 means the session had already expired: signed out either way.
      if (!(err instanceof ApiError && err.status === 401)) {
        setConfirmSignOut(false);
        toast.show({
          tone: 'error',
          message: `Couldn't sign out: ${err instanceof Error ? err.message : 'unknown error'}`,
        });
        return;
      }
    }
    setConfirmSignOut(false);
    // Not queryClient.clear(): that strands mounted observers on removed queries.
    queryClient.setQueryData(queryKeys.me, null);
    await queryClient.invalidateQueries();
  }

  const items: MenuItem[] = [
    { label: 'Players', href: '/players' },
    { label: 'Account', href: '/account' },
    ...(user.permissions.canEdit ? [{ label: 'Edit info', href: '/info/edit' }] : []),
    ...(user.permissions.canManageUsers ? [{ label: 'Users', href: '/users' }] : []),
    { label: 'Sign out', onSelect: () => setConfirmSignOut(true) },
  ];

  return (
    <>
      <Menu
        label={
          <>
            <span className="max-w-40 truncate">{user.name}</span>
            <span aria-hidden="true">▾</span>
          </>
        }
        triggerLabel={user.name}
        items={items}
        triggerClassName={buttonClasses('secondary', 'sm')}
      />
      <ConfirmDialog
        open={confirmSignOut}
        title="Sign out?"
        message="You'll need to sign in again to make changes."
        confirmLabel="Sign out"
        onConfirm={signOut}
        onCancel={() => setConfirmSignOut(false)}
      />
    </>
  );
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `yarn vitest run client/components/layout/AccountControl.test.tsx && yarn test:client`
Expected: PASS. 11 account tests pass and the suite is green. If "focuses Email" fails because React's `autoFocus` didn't run inside the dialog, focus the email input from a `useEffect` on `open` in `SignInDialog`, and record that as a ruling.

- [ ] **Step 6: Commit**

```bash
yarn eslint --fix client/api client/lib client/components/layout client/test
yarn typecheck
git add client/api client/lib client/components/layout client/test/mockFetch.ts
git commit -q -m "Add sign-in dialog and account menu with sign-out" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 10: Layout shell (masthead, nav, footer)

**Files:**
- Create in `client/components/layout/`: `navItems.ts`, `Masthead.tsx`, `NavBar.tsx`, `Footer.tsx`, `SkipLink.tsx`, `Layout.tsx`
- Delete: `client/components/Layout.tsx`
- Modify: `client/App.tsx` (the Layout import path)
- Test: `client/components/layout/NavBar.test.tsx`, `client/components/layout/Footer.test.tsx`, `client/App.test.tsx`

**Interfaces:**
- Consumes:
  - `ThemeToggle` (Task 2);
  - `AccountControl` (Task 9);
  - `Container` and `buttonClasses` (Tasks 2 and 3);
  - `useDisclosure` and `useDismiss` (Task 7);
  - `useSite` and `useAuth`.
- Produces:
  - `interface NavItem { label: string; legacyHref: string; to?: string }`
  - `navItems: NavItem[]`
  - `Masthead`, `NavBar`, `Footer` and `SkipLink`
  - a default-export `Layout` at `client/components/layout/Layout.tsx`

- [ ] **Step 1: Write the failing tests**

Create `client/components/layout/NavBar.test.tsx`:

```tsx
import { fireEvent, screen } from '@testing-library/react';
import { Link } from 'react-router';
import { beforeEach, describe, expect, it } from 'vitest';

import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import { NavBar } from './NavBar';
import { navItems } from './navItems';

beforeEach(() => {
  mockFetch({
    '/api/v1/auth/me': { status: 401, body: { error: { code: 401, message: 'login required' } } },
  });
});

function renderNav(route = '/') {
  renderWithProviders(
    <>
      <NavBar />
      <Link to="/elsewhere">elsewhere</Link>
    </>,
    { route },
  );
  return screen.getByRole('button', { name: 'Menu' });
}

describe('NavBar', () => {
  it('lists every nav item in order', () => {
    renderNav();
    expect(navItems.map((i) => i.label)).toEqual([
      'Home',
      'Teams',
      'News',
      "What's On",
      'Gallery',
      'Documents',
      'Programmes',
      'Sponsors',
      'Info',
      'Contact',
    ]);
    const nav = screen.getByRole('navigation', { name: 'Main' });
    for (const item of navItems) {
      expect(nav).toContainElement(screen.getByRole('link', { name: item.label }));
    }
  });

  it('uses a router link for ported pages and marks the current one', () => {
    renderNav('/');
    const home = screen.getByRole('link', { name: 'Home' });
    expect(home).toHaveAttribute('href', '/');
    expect(home).toHaveAttribute('aria-current', 'page');
  });

  it('uses a plain legacy link for pages not yet ported', () => {
    renderNav('/');
    const news = screen.getByRole('link', { name: 'News' });
    expect(news).toHaveAttribute('href', '/news');
    expect(news).not.toHaveAttribute('aria-current');
  });

  it('shows the theme toggle and the account control', async () => {
    renderNav();
    expect(screen.getByRole('button', { name: /^Theme:/ })).toBeInTheDocument();
    expect(await screen.findByRole('button', { name: 'Sign in' })).toBeInTheDocument();
  });

  it('opens and closes the phone menu', () => {
    const menu = renderNav();
    expect(menu).toHaveAttribute('aria-expanded', 'false');
    fireEvent.click(menu);
    expect(menu).toHaveAttribute('aria-expanded', 'true');
    expect(document.getElementById(menu.getAttribute('aria-controls') ?? '')).not.toBeNull();
    fireEvent.click(menu);
    expect(menu).toHaveAttribute('aria-expanded', 'false');
  });

  it('closes the phone menu on Esc', () => {
    const menu = renderNav();
    fireEvent.click(menu);
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(menu).toHaveAttribute('aria-expanded', 'false');
  });

  it('closes the phone menu on an outside click', () => {
    const menu = renderNav();
    fireEvent.click(menu);
    fireEvent.mouseDown(document.body);
    expect(menu).toHaveAttribute('aria-expanded', 'false');
  });

  it('closes the phone menu when the route changes', () => {
    const menu = renderNav();
    fireEvent.click(menu);
    fireEvent.click(screen.getByRole('link', { name: 'elsewhere' }));
    expect(menu).toHaveAttribute('aria-expanded', 'false');
  });
});
```

Create `client/components/layout/Footer.test.tsx`:

```tsx
import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import { Footer } from './Footer';

const site = { year: 2026, visitorCount: 12345, version: 'test', teams: [] };
const anonymous = {
  status: 401,
  body: { error: { code: 401, message: 'login required' } },
};
const signedIn = {
  body: {
    id: 1,
    name: 'Jo Smith',
    email: 'jo@example.test',
    role: 'Manager',
    permissions: { canEdit: false, canManageGallery: false, canManageUsers: false },
  },
};

describe('Footer', () => {
  it('shows copyright, credit and social links, but no visitor count, when signed out', async () => {
    mockFetch({ '/api/v1/site': { body: site }, '/api/v1/auth/me': anonymous });
    renderWithProviders(<Footer />);
    const footer = screen.getByRole('contentinfo');
    expect(await screen.findByText(/© 2020–2026 AFC Aldermaston/)).toBeInTheDocument();
    expect(footer).toHaveTextContent('Website provided by BSWDI');
    const credit = screen.getByRole('link', { name: 'BSWDI' });
    expect(credit).toHaveAttribute('href', 'https://bswdi.co.uk');
    expect(credit).toHaveAttribute('target', '_blank');
    expect(credit).toHaveAttribute('rel', 'noopener');
    expect(screen.getByRole('link', { name: 'AFC Aldermaston on Facebook' })).toHaveAttribute(
      'href',
      'https://www.facebook.com/AFC-Aldermaston-114651238068/',
    );
    expect(screen.getByRole('link', { name: 'AFC Aldermaston on X' })).toHaveAttribute(
      'href',
      'https://x.com/afcaldermaston',
    );
    expect(footer).not.toHaveTextContent('Visitor count');
  });

  it('shows the visitor count when signed in', async () => {
    mockFetch({ '/api/v1/site': { body: site }, '/api/v1/auth/me': signedIn });
    renderWithProviders(<Footer />);
    expect(await screen.findByText('Visitor count: 12,345')).toBeInTheDocument();
  });

  it('falls back to the current year before /site loads', () => {
    mockFetch({
      '/api/v1/site': () => new Promise(() => {}),
      '/api/v1/auth/me': anonymous,
    });
    renderWithProviders(<Footer />);
    expect(screen.getByRole('contentinfo')).toHaveTextContent(
      `© 2020–${new Date().getFullYear()} AFC Aldermaston`,
    );
  });
});
```

In `client/App.test.tsx`, replace the `'wraps pages in the layout'` test with:

```tsx
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
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/components/layout client/App.test.tsx`
Expected: FAIL. `./NavBar`, `./navItems` and `./Footer` can't be resolved, and the App layout test fails because there's no "Facta Non Verba" yet.

- [ ] **Step 3: Implement**

`client/components/layout/navItems.ts`:

```ts
export interface NavItem {
  label: string;
  /** The legacy page; used until the page is ported. */
  legacyHref: string;
  /** The SPA route (relative to /app) once sub-project 4 ports the page. */
  to?: string;
}

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

`client/components/layout/SkipLink.tsx`:

```tsx
export function SkipLink() {
  return (
    <a
      href="#content"
      className="sr-only focus:not-sr-only focus:fixed focus:top-3 focus:left-3 focus:z-50 focus:rounded-md focus:bg-bg focus:px-3 focus:py-2 focus:shadow-lg"
    >
      Skip to content
    </a>
  );
}
```

`client/components/layout/Masthead.tsx`:

```tsx
import { Link } from 'react-router';

import crest from '../../assets/crest.png';
import faLogo from '../../assets/fa-logo.jpeg';

export function Masthead() {
  return (
    <div className="grid grid-cols-[auto_1fr_auto] items-center gap-3 px-3.5 py-3 md:gap-4 md:px-7 md:py-4.5">
      <Link to="/" className="rounded-full">
        <img
          src={crest}
          alt="AFC Aldermaston home"
          className="size-13 rounded-full bg-white md:size-24"
        />
      </Link>
      <div className="text-center">
        <p className="font-display text-xl leading-none font-extrabold tracking-wider uppercase md:text-[40px]">
          AFC Aldermaston
        </p>
        <p className="mt-1.5 text-[10px] tracking-[0.08em] text-muted uppercase md:text-[13px] md:tracking-[0.14em]">
          Facta Non Verba · <span className="md:hidden">1952</span>
          <span className="hidden md:inline">Founded 1952</span>
        </p>
      </div>
      <img
        src={faLogo}
        alt="The FA Charter Standard"
        className="h-11.5 w-auto rounded-md bg-white p-1 md:h-21"
      />
    </div>
  );
}
```

`client/components/layout/NavBar.tsx`:

```tsx
import { clsx } from 'clsx';
import { useId, useRef, useState } from 'react';
import { matchPath, NavLink, useLocation } from 'react-router';

import { ThemeToggle } from '../../theme/ThemeToggle';
import { buttonClasses } from '../ui/buttonStyles';
import { useDisclosure } from '../ui/useDisclosure';
import { useDismiss } from '../ui/useDismiss';
import { AccountControl } from './AccountControl';
import { navItems, type NavItem } from './navItems';

function NavItemLink({ item, className }: { item: NavItem; className: string }) {
  if (item.to !== undefined) {
    return (
      <NavLink
        to={item.to}
        end
        className={({ isActive }) =>
          clsx(className, isActive && 'text-red shadow-[inset_0_-3px_0_var(--color-red)]')
        }
      >
        {item.label}
      </NavLink>
    );
  }
  return (
    <a href={item.legacyHref} className={className}>
      {item.label}
    </a>
  );
}

export function NavBar() {
  const { pathname } = useLocation();
  const menu = useDisclosure();
  const barRef = useRef<HTMLDivElement>(null);
  const panelId = useId();
  useDismiss(barRef, menu.open, menu.close);

  // Close the phone menu whenever the route changes (adjusting state during render).
  const [lastPath, setLastPath] = useState(pathname);
  if (pathname !== lastPath) {
    setLastPath(pathname);
    menu.close();
  }

  const current = navItems.find(
    (item) => item.to !== undefined && matchPath({ path: item.to, end: true }, pathname),
  );

  return (
    <nav aria-label="Main" className="border-t border-b-3 border-line border-b-red">
      <div
        ref={barRef}
        className="flex flex-wrap items-center justify-end gap-x-5 px-3.5 md:px-7"
      >
        <span className="mr-auto py-2 text-sm font-bold text-red md:hidden">{current?.label}</span>
        <ul
          id={panelId}
          className={clsx(
            'text-sm font-semibold md:flex md:items-center md:gap-5',
            menu.open
              ? 'max-md:order-last max-md:w-full max-md:border-t max-md:border-line max-md:py-1'
              : 'max-md:hidden',
          )}
        >
          {navItems.map((item) => (
            <li key={item.label}>
              <NavItemLink item={item} className="block py-2.5 hover:text-red md:py-3" />
            </li>
          ))}
        </ul>
        <div className="flex items-center gap-2 py-2">
          <ThemeToggle />
          <AccountControl />
          <button
            type="button"
            aria-expanded={menu.open}
            aria-controls={panelId}
            onClick={menu.toggle}
            className={buttonClasses('secondary', 'sm', 'md:hidden')}
          >
            Menu
          </button>
        </div>
      </div>
    </nav>
  );
}
```

`client/components/layout/Footer.tsx`:

```tsx
import { useSite } from '../../api/queries';
import { useAuth } from '../../auth/useAuth';

const socials = [
  {
    label: 'AFC Aldermaston on Facebook',
    href: 'https://www.facebook.com/AFC-Aldermaston-114651238068/',
    path: 'M14 8h3V4h-3c-2.8 0-5 2.2-5 5v2H7v4h2v9h4v-9h3l1-4h-4V9c0-.6.4-1 1-1z',
  },
  {
    label: 'AFC Aldermaston on X',
    href: 'https://x.com/afcaldermaston',
    path: 'M17.8 3h3.1l-6.8 7.8L22 21h-6.2l-4.9-6.4L5.3 21H2.2l7.3-8.3L2 3h6.4l4.4 5.8L17.8 3zm-1.1 16.2h1.7L7.4 4.7H5.6l11.1 14.5z',
  },
];

export function Footer() {
  const site = useSite();
  const { user } = useAuth();
  const year = site.data?.year ?? new Date().getFullYear();

  return (
    <footer className="mt-12 bg-footer text-white">
      <div className="mx-auto flex w-full max-w-6xl flex-wrap items-center justify-between gap-4 px-4 py-5 text-sm md:px-6">
        <div>
          <p>
            © 2020–{year} AFC Aldermaston · Website provided by{' '}
            <a
              href="https://bswdi.co.uk"
              target="_blank"
              rel="noopener"
              className="font-bold underline"
            >
              BSWDI
            </a>
          </p>
          {user && site.data && (
            <p className="mt-1 opacity-75">
              Visitor count: {site.data.visitorCount.toLocaleString('en-GB')}
            </p>
          )}
        </div>
        <ul className="flex gap-2.5">
          {socials.map((s) => (
            <li key={s.href}>
              <a
                href={s.href}
                target="_blank"
                rel="noopener"
                aria-label={s.label}
                className="grid size-8 place-items-center rounded-full bg-white/15 hover:bg-white/25"
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                  <path d={s.path} />
                </svg>
              </a>
            </li>
          ))}
        </ul>
      </div>
    </footer>
  );
}
```

`client/components/layout/Layout.tsx`:

```tsx
import { Outlet } from 'react-router';

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
          <Outlet />
        </Container>
      </main>
      <Footer />
    </div>
  );
}
```

Delete the old shell and update the import:

```bash
git rm -q client/components/Layout.tsx
```

In `client/App.tsx`, change `import Layout from './components/Layout';` to `import Layout from './components/layout/Layout';`.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn vitest run client/components/layout client/App.test.tsx && yarn test:client`
Expected: PASS. 8 NavBar, 3 Footer and 3 App tests pass, and the suite is green.

- [ ] **Step 5: Commit**

```bash
yarn eslint --fix client/components/layout client/App.tsx client/App.test.tsx
yarn typecheck
git add -A client/components client/App.tsx client/App.test.tsx
git commit -q -m "Add styled layout shell: masthead, nav with phone menu, footer, skip link" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 11: Showcase page, restyled pages, accessibility smoke tests and docs

**Files:**
- Create: `client/pages/DesignPage.tsx`, `client/test/axe.ts`, `client/a11y.test.tsx`
- Modify: `client/App.tsx`, `client/pages/HomePage.tsx`, `client/pages/NotFoundPage.tsx`, `client/App.test.tsx`, `README.md`
- Test: `client/pages/DesignPage.test.tsx`, `client/a11y.test.tsx`, plus the existing `client/pages/HomePage.test.tsx` (unchanged, and it must stay green)

**Interfaces:**
- Consumes: every component from Tasks 3–9; `axe-core`.
- Produces:
  - the `/design` route (`/app/design` in the browser);
  - `axeViolations(node: Element): Promise<string[]>`.

- [ ] **Step 1: Write the failing tests**

Create `client/test/axe.ts`:

```ts
import axe from 'axe-core';

/**
 * Runs axe over node and returns "rule: selectors" strings, so a failing
 * expect(...).toEqual([]) says exactly what broke. Colour contrast is off
 * because jsdom does not compute styles; the palette was checked by hand.
 */
export async function axeViolations(node: Element): Promise<string[]> {
  const results = await axe.run(node, { rules: { 'color-contrast': { enabled: false } } });
  return results.violations.map(
    (v) => `${v.id}: ${v.nodes.map((n) => n.target.join(' ')).join(', ')}`,
  );
}
```

Create `client/pages/DesignPage.test.tsx`:

```tsx
import { fireEvent, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';

import App from '../App';
import { mockFetch } from '../test/mockFetch';
import { renderWithProviders } from '../test/render';

beforeEach(() => {
  mockFetch({
    '/api/v1/site': { body: { year: 2026, visitorCount: 1, version: 'test', teams: [] } },
    '/api/v1/auth/me': { status: 401, body: { error: { code: 401, message: 'login required' } } },
  });
});

describe('DesignPage', () => {
  it('is served at /design inside the layout', () => {
    renderWithProviders(<App />, { route: '/design' });
    expect(screen.getByRole('heading', { level: 1, name: 'Design system' })).toBeInTheDocument();
    for (const section of ['Palette', 'Buttons', 'Badges and notices', 'Cards', 'Forms', 'Table', 'Dialogs, menus and toasts']) {
      expect(screen.getByRole('heading', { level: 2, name: section })).toBeInTheDocument();
    }
  });

  it('opens the sample modal and confirm dialog', () => {
    renderWithProviders(<App />, { route: '/design' });
    fireEvent.click(screen.getByRole('button', { name: 'Open modal' }));
    const modal = screen.getByRole('dialog', { name: 'Sample modal' });
    fireEvent.click(within(modal).getByRole('button', { name: 'Close' }));
    expect(screen.queryByRole('dialog', { name: 'Sample modal' })).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: 'Open confirm' }));
    expect(screen.getByRole('dialog', { name: 'Delete team?' })).toBeInTheDocument();
  });

  it('shows each toast tone', () => {
    renderWithProviders(<App />, { route: '/design' });
    fireEvent.click(screen.getByRole('button', { name: 'Success toast' }));
    expect(screen.getByRole('button', { name: 'Player saved' })).toBeInTheDocument();
  });

  it('shows the gradient for cards without an image', () => {
    const { container } = renderWithProviders(<App />, { route: '/design' });
    // Only the "No image" card: jsdom never fires the broken image's error event.
    expect(container.querySelectorAll('[data-fallback]').length).toBeGreaterThanOrEqual(1);
  });
});
```

Create `client/a11y.test.tsx`:

```tsx
import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import App from './App';
import { axeViolations } from './test/axe';
import { mockFetch, type MockResponse } from './test/mockFetch';
import { renderWithProviders } from './test/render';

const site = {
  year: 2026,
  visitorCount: 42,
  version: 'test',
  teams: [{ id: 1, name: 'First Team', isActive: true, isYouth: false, ages: 99 }],
};
const anonymous: MockResponse = {
  status: 401,
  body: { error: { code: 401, message: 'login required' } },
};
const signedIn: MockResponse = {
  body: {
    id: 1,
    name: 'Jo Smith',
    email: 'jo@example.test',
    role: 'Club Secretary',
    permissions: { canEdit: true, canManageGallery: true, canManageUsers: true },
  },
};

describe.each(['light', 'dark'] as const)('accessibility (%s theme)', (theme) => {
  it.each([
    ['home, signed out', '/', anonymous],
    ['home, signed in', '/', signedIn],
    ['design page', '/design', anonymous],
    ['not found', '/nope', anonymous],
  ])('%s has no axe violations', async (_name, route, me) => {
    localStorage.setItem('afc-theme', theme);
    mockFetch({ '/api/v1/site': { body: site }, '/api/v1/auth/me': me });
    const { container } = renderWithProviders(<App />, { route });
    await screen.findByRole('contentinfo');
    await screen.findByRole('button', { name: me === anonymous ? 'Sign in' : /Jo Smith/ });
    expect(document.documentElement.dataset.theme).toBe(theme);
    expect(await axeViolations(container)).toEqual([]);
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/DesignPage.test.tsx client/a11y.test.tsx`
Expected: FAIL. There's no "Design system" heading, because `/design` falls through to Not Found. The a11y tests may pass or fail at this point; either is fine. Record any violations they list, because Step 4 must end with none.

- [ ] **Step 3: Implement the pages**

`client/pages/DesignPage.tsx`:

```tsx
import { useState, type ReactNode } from 'react';

import { Alert } from '../components/ui/Alert';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { ButtonLink } from '../components/ui/ButtonLink';
import { Card, CardBody, CardMedia } from '../components/ui/Card';
import { Checkbox } from '../components/ui/Checkbox';
import { ConfirmDialog } from '../components/ui/ConfirmDialog';
import { FileInput, Input, Select, Textarea } from '../components/ui/controls';
import { EmptyState } from '../components/ui/EmptyState';
import { Field } from '../components/ui/Field';
import { Menu } from '../components/ui/Menu';
import { Modal } from '../components/ui/Modal';
import { PageHeader } from '../components/ui/PageHeader';
import { Skeleton } from '../components/ui/Skeleton';
import { Spinner } from '../components/ui/Spinner';
import { Table, TBody, Td, Th, THead, Tr } from '../components/ui/Table';
import { useToast } from '../components/ui/toast/useToast';

const swatches = [
  'red',
  'red-hover',
  'club-red',
  'blue',
  'footer',
  'ink',
  'muted',
  'line',
  'bg',
  'surface',
  'field',
  'success',
  'warning',
];

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="mb-10">
      <h2 className="mb-3 font-display text-2xl font-extrabold tracking-wide uppercase">{title}</h2>
      {children}
    </section>
  );
}

// Every shared component in every state: the reference for page ports and visual review.
export default function DesignPage() {
  const toast = useToast();
  const [modalOpen, setModalOpen] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);

  return (
    <>
      <PageHeader
        title="Design system"
        subtitle="Shared components and tokens for the AFC Aldermaston site."
      />

      <Section title="Palette">
        <ul className="grid grid-cols-2 gap-3 sm:grid-cols-4 lg:grid-cols-6">
          {swatches.map((name) => (
            <li key={name} className="overflow-hidden rounded-lg border border-line">
              <div className="h-12" style={{ background: `var(--color-${name})` }} />
              <p className="px-2 py-1 text-xs">{name}</p>
            </li>
          ))}
        </ul>
      </Section>

      <Section title="Buttons">
        <div className="flex flex-wrap items-center gap-2">
          <Button>Save</Button>
          <Button variant="secondary">Cancel</Button>
          <Button variant="ghost">View all</Button>
          <Button variant="danger">Delete</Button>
          <Button size="sm">Small</Button>
          <Button loading>Saving</Button>
          <Button disabled>Disabled</Button>
          <ButtonLink to="/" variant="secondary">
            Router link
          </ButtonLink>
          <Spinner />
        </div>
      </Section>

      <Section title="Badges and notices">
        <div className="mb-3 flex flex-wrap gap-2">
          <Badge tone="red">U12s</Badge>
          <Badge tone="blue">Webmaster</Badge>
          <Badge>Draft</Badge>
        </div>
        <div className="grid gap-2">
          <Alert tone="success">Team saved.</Alert>
          <Alert tone="error">Couldn&apos;t save: the email is already in use.</Alert>
          <Alert tone="warning">This team has no manager assigned.</Alert>
          <Alert tone="info">Fixtures are published on Fridays.</Alert>
        </div>
      </Section>

      <Section title="Cards">
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardMedia src="/app/favicon.png" alt="Club crest" />
            <CardBody>
              <Badge tone="red">News</Badge>
              <p className="mt-2 font-display text-xl font-extrabold uppercase">With an image</p>
            </CardBody>
          </Card>
          <Card>
            <CardMedia alt="" />
            <CardBody>
              <Badge tone="red">News</Badge>
              <p className="mt-2 font-display text-xl font-extrabold uppercase">No image</p>
              <p className="text-xs text-muted">28 Sep 2026</p>
            </CardBody>
          </Card>
          <Card>
            <CardMedia src="/app/does-not-exist.png" alt="" />
            <CardBody>
              <Badge tone="blue">What&apos;s on</Badge>
              <p className="mt-2 font-display text-xl font-extrabold uppercase">Broken image</p>
            </CardBody>
          </Card>
          <Card>
            <CardBody className="space-y-2">
              <Skeleton className="w-2/5" />
              <Skeleton />
              <Skeleton className="w-3/4" />
            </CardBody>
          </Card>
        </div>
        <Card className="mt-4">
          <EmptyState
            title="No news yet"
            message="Articles appear here once they are published."
            action={<Button variant="secondary">Refresh</Button>}
          />
        </Card>
      </Section>

      <Section title="Forms">
        <form className="grid max-w-2xl gap-4 sm:grid-cols-2" onSubmit={(e) => e.preventDefault()}>
          <Field label="Name" help="As shown on the team page.">
            <Input defaultValue="Jo Smith" />
          </Field>
          <Field label="Email" error="Enter a valid email address.">
            <Input type="email" defaultValue="jo@" />
          </Field>
          <Field label="Team">
            <Select defaultValue="u12">
              <option value="u12">Under 12s</option>
              <option value="first">First Team</option>
            </Select>
          </Field>
          <Field label="Photo">
            <FileInput accept="image/*" />
          </Field>
          <Field label="Notes" className="sm:col-span-2">
            <Textarea />
          </Field>
          <Checkbox label="Youth team" defaultChecked />
        </form>
      </Section>

      <Section title="Table">
        <Table>
          <THead>
            <Tr>
              <Th>Name</Th>
              <Th>Role</Th>
              <Th>Team</Th>
            </Tr>
          </THead>
          <TBody>
            <Tr>
              <Td>Jo Smith</Td>
              <Td>
                <Badge tone="blue">Manager</Badge>
              </Td>
              <Td>Under 12s</Td>
            </Tr>
            <Tr>
              <Td>Sam Patel</Td>
              <Td>
                <Badge tone="blue">Webmaster</Badge>
              </Td>
              <Td>—</Td>
            </Tr>
          </TBody>
        </Table>
      </Section>

      <Section title="Dialogs, menus and toasts">
        <div className="flex flex-wrap items-center gap-2">
          <Button variant="secondary" onClick={() => setModalOpen(true)}>
            Open modal
          </Button>
          <Button variant="danger" onClick={() => setConfirmOpen(true)}>
            Open confirm
          </Button>
          <Menu
            label="Sample menu ▾"
            triggerLabel="Sample menu"
            align="start"
            items={[
              { label: 'Players', href: '/players' },
              { label: 'Account', href: '/account' },
              { label: 'Say hello', onSelect: () => toast.show({ message: 'Hello' }) },
            ]}
          />
          <Button
            variant="secondary"
            onClick={() => toast.show({ tone: 'success', message: 'Player saved' })}
          >
            Success toast
          </Button>
          <Button
            variant="secondary"
            onClick={() => toast.show({ tone: 'error', message: "Couldn't save the player" })}
          >
            Error toast
          </Button>
          <Button
            variant="secondary"
            onClick={() => toast.show({ tone: 'info', message: 'Fixtures updated' })}
          >
            Info toast
          </Button>
        </div>
        <Modal
          open={modalOpen}
          onClose={() => setModalOpen(false)}
          title="Sample modal"
          actions={<Button onClick={() => setModalOpen(false)}>Close</Button>}
        >
          <p className="text-sm">Esc, the backdrop or Close dismiss this dialog.</p>
        </Modal>
        <ConfirmDialog
          open={confirmOpen}
          title="Delete team?"
          message="This removes Under 12s and unlinks its players. This can't be undone."
          confirmLabel="Delete"
          tone="danger"
          onConfirm={() => {
            setConfirmOpen(false);
            toast.show({ tone: 'success', message: 'Team deleted (not really)' });
          }}
          onCancel={() => setConfirmOpen(false)}
        />
      </Section>
    </>
  );
}
```

`client/App.tsx`: import `DesignPage` from `./pages/DesignPage`, and add a route before the catch-all:

```tsx
        <Route index element={<HomePage />} />
        <Route path="design" element={<DesignPage />} />
        <Route path="*" element={<NotFoundPage />} />
```

Replace `client/pages/HomePage.tsx`. The content and text are the same as before, now using the components:

```tsx
import { useSite } from '../api/queries';
import { useAuth } from '../auth/useAuth';
import { Alert } from '../components/ui/Alert';
import { Badge } from '../components/ui/Badge';
import { Card, CardBody } from '../components/ui/Card';
import { EmptyState } from '../components/ui/EmptyState';
import { PageHeader } from '../components/ui/PageHeader';
import { Skeleton } from '../components/ui/Skeleton';
import { Table, TBody, Td, Th, THead, Tr } from '../components/ui/Table';

// Proof that the client, API, session cookie and data layer work end to end.
// Real pages arrive in sub-project 4.
export default function HomePage() {
  const site = useSite();
  const { user, isLoading } = useAuth();

  let signedIn = 'Not signed in';
  if (isLoading) {
    signedIn = 'Checking sign-in…';
  } else if (user) {
    signedIn = `Signed in as ${user.name} (${user.role})`;
  }

  return (
    <>
      <PageHeader title="AFC Aldermaston" subtitle={signedIn} />
      {site.isPending && (
        <div className="space-y-2">
          <Skeleton className="w-1/3" />
          <Skeleton />
          <Skeleton className="w-2/3" />
        </div>
      )}
      {site.isError && (
        <Alert tone="error">Could not load site information: {site.error.message}</Alert>
      )}
      {site.data && (
        <Card>
          <CardBody className="space-y-4">
            <p className="text-muted">Visitors: {site.data.visitorCount}</p>
            <h2 className="font-display text-2xl font-extrabold tracking-wide uppercase">Teams</h2>
            {site.data.teams.length === 0 ? (
              <EmptyState title="No teams yet." />
            ) : (
              <Table>
                <THead>
                  <Tr>
                    <Th>Team</Th>
                    <Th>Type</Th>
                  </Tr>
                </THead>
                <TBody>
                  {site.data.teams.map((team) => (
                    <Tr key={team.id}>
                      <Td>{team.name}</Td>
                      <Td>
                        <Badge tone={team.isYouth ? 'red' : 'blue'}>
                          {team.isYouth ? 'Youth' : 'Adult'}
                        </Badge>
                      </Td>
                    </Tr>
                  ))}
                </TBody>
              </Table>
            )}
          </CardBody>
        </Card>
      )}
    </>
  );
}
```

Replace `client/pages/NotFoundPage.tsx`:

```tsx
import { ButtonLink } from '../components/ui/ButtonLink';
import { EmptyState } from '../components/ui/EmptyState';
import { PageHeader } from '../components/ui/PageHeader';

export default function NotFoundPage() {
  return (
    <>
      <PageHeader title="Page not found" />
      <EmptyState
        title="Nothing here"
        message="The page you were looking for doesn't exist or has moved."
        action={<ButtonLink to="/">Go to the start</ButtonLink>}
      />
    </>
  );
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn vitest run client/pages client/a11y.test.tsx client/App.test.tsx`
Expected: PASS. That's 4 DesignPage tests, 8 a11y tests (4 routes × 2 themes), 4 HomePage tests (unchanged) and 3 App tests.

If axe reports violations, fix the markup (labels, landmarks, headings) in the component responsible. Don't disable axe rules. The only rule switched off is `color-contrast`, in `client/test/axe.ts`.

- [ ] **Step 5: Document the design system**

In `README.md`, directly under the section that describes the client (the section with `yarn dev` and `/app`), add:

```markdown
### Design system

The client's look lives in `client/styles/app.css`. That file holds the Tailwind v4 theme: colours and fonts as CSS variables, with a `[data-theme='dark']` override. Shared components are in `client/components/ui/`, one file each, and the layout shell is in `client/components/layout/`. Run `yarn dev` and open `/app/design` to see every component in every state, in light and dark. The FA logo in the header is `client/assets/fa-logo.jpeg`; replace that file to update it.
```

Also fix the known README slip from sub-project 2: wherever it says `yarn test` runs `go test ./...`, change it to `go test ./server/...`.

- [ ] **Step 6: Full verification**

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
- all Vitest files pass;
- the build succeeds with no "can't be bundled" warning;
- the Go tests pass (untouched).

Manual check with `yarn build && ./build/afc` (needs a real `.env`), or `yarn dev`:
- `/app` and `/app/design` render in light and dark;
- the theme toggle cycles through the settings and survives a reload with no white flash;
- at 360px width the Menu button opens the link list, and a long account name truncates (Review Focus 5);
- sign-in and sign-out work against the real API.

If no `.env` is available, say so in the final report rather than claiming it was checked.

- [ ] **Step 7: Commit**

```bash
yarn eslint --fix client
git add -A client README.md
git commit -q -m "Add /design showcase, restyle home and not-found pages, add axe smoke tests" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```
