# Account Pages Implementation Plan (Sub-project 4b)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the Account page (details, photo upload/remove, change password) and the password reset-link page to the React client. The sign-in dialog routes reset-flagged accounts into the SPA, and a single app-wide sign-in dialog replaces the one inside `AccountControl`.

**Architecture:**
- A small `client/api/account.ts` module with API calls, plus a `useResetTokenCheck` query.
- A `SignInProvider` owns the one `SignInDialog` and exposes `useSignIn().open()`.
- The pages live in `client/pages/account/` and `client/pages/reset/`, and both are lazily routed.
- `PasswordFields` is shared by the reset page and the change-password card.
- The server stays the authority on password rules. Its `422` field errors appear on the matching inputs.

**Tech Stack:** React 19, React Router 7, TanStack Query 5, TypeScript 6, Tailwind 4, Vitest 5, Testing Library and jsdom, axe-core.

**Spec:** `docs/superpowers/specs/2026-09-30-account-pages-design.md`

## Global Constraints

- **Where to work:** the worktree `/Users/liam/Code/Go/AFC-design-system`, branch `account-pages`. Never touch `/Users/liam/Code/Go/AFC`, never read `postgres_*.sql`, and never `rm -rf` the current working directory.
- **Commits:** every commit message ends with a blank line and `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
- **Scope:** no server (Go) changes and no new dependencies.
- **API calls:**
  - `PUT /account/image` (multipart field `image`, returns `CurrentUser`);
  - `DELETE /account/image` (204);
  - `POST /auth/password` with `{ oldPassword, newPassword, confirmationPassword }` (204, or 422 with fields of the same names);
  - `GET` and `POST /auth/reset/:token`, the POST with `{ newPassword, confirmationPassword }` (204, 404, or 422).
- **Reset token pattern:** `/^[A-Za-z0-9-]{1,100}$/`. The server's reset URL has the form `/reset/<token>`.
- **Copy, verbatim:**
  - Titles: "Your account"; "Sign in to see your account"; "Your session may have expired."; "To change these, ask a club administrator."
  - Photo: "Choose a new photo"; "Save photo"; "Cancel"; "Remove photo"; "Remove your photo?"; "Your photo will be replaced by the club crest."; "Photo updated"; "Photo removed"; "Couldn't remove your photo: <message>".
  - Password: "Current password"; "New password"; "Confirm new password"; "Change password"; "Password changed"; "Enter your current password"; "Enter a new password"; "Confirm your new password".
  - Password hint: "More than 8 characters, with a lower-case letter, an upper-case letter, a number and a special character."
  - Reset: "Reset your password"; "This reset link is invalid or has expired"; "Ask a club administrator for a new one, or sign in again if you were sent here after signing in."; "Set new password"; "Your password has been changed. You can now sign in with it."
- **Page titles:** "Account · AFC Aldermaston" and "Reset password · AFC Aldermaston".
- **Accepted image types:** `image/jpeg,image/png,image/gif,image/webp,image/avif,image/apng,image/svg+xml`.
- **Style and lint:** use token classes only, one component per `.tsx`, no barrel files. Run `yarn eslint --fix <files>` **before** running a task's tests, then commit.
- **jsdom gaps:** jsdom has no `URL.createObjectURL`/`revokeObjectURL`. Tests assign mocks onto `URL` (don't stub the `URL` global) and delete them afterwards.

## Review Focus

1. **Choosing a second photo before saving, or cancelling and then choosing again:** the preview shows the latest file, and each replaced preview URL is revoked. Pinned in Task 4.
2. **A reset link that expires between opening the page and submitting:** the submit gets a 404 and the page switches to the invalid-link state. It doesn't show a raw error. Pinned in Task 3.
3. **Submitting the password forms with empty fields:** nothing is sent, and each empty field says what it needs. Pinned in Tasks 3 and 4.
4. **A signed-in user following an admin reset link:** the reset page still works, and the header stays signed in. Pinned in Task 3.
5. **Signing in on `/account` from the signed-out prompt:** the page switches to the signed-in view without a reload. Pinned in Task 4.

---

### Task 1: Account API calls and the reset-link mapping

**Files:**
- Create: `client/api/account.ts`, `client/lib/resetLink.ts`
- Modify: `client/test/fixtures.ts` (reset fixtures)
- Test: `client/api/account.test.ts`, `client/lib/resetLink.test.ts`

**Interfaces:**
- Produces:
  - `PasswordChange`, `PasswordReset`
  - `uploadAccountImage(file: File): Promise<CurrentUser>`
  - `removeAccountImage(): Promise<void>`
  - `changePassword(input: PasswordChange): Promise<void>`
  - `checkResetToken(token: string, signal?: AbortSignal): Promise<void>`
  - `resetPassword(token: string, input: PasswordReset): Promise<void>`
  - `useResetTokenCheck(token: string | null)`, whose query data is `true`
  - `RESET_TOKEN: RegExp`, `appResetPath(url: string): string | null`
  - fixtures `resetToken`, `expiredResetToken`, and matching `publicRoutes` entries

- [ ] **Step 1: Write the failing tests**

`client/lib/resetLink.test.ts`:

```ts
import { describe, expect, it } from 'vitest';

import { appResetPath, RESET_TOKEN } from './resetLink';

describe('appResetPath', () => {
  it('maps the server reset URL to the in-app route', () => {
    expect(appResetPath('/reset/5f0b6c1e-1a2b-4c3d-8e9f-000000000001')).toBe(
      '/reset/5f0b6c1e-1a2b-4c3d-8e9f-000000000001',
    );
  });

  it('rejects anything else', () => {
    for (const url of ['/reset/', '/reset/a/b', 'https://x.example/reset/abc', '/other/abc', '/reset/<x>']) {
      expect(appResetPath(url)).toBeNull();
    }
  });

  it('exposes the token pattern', () => {
    expect(RESET_TOKEN.test('abc-123')).toBe(true);
    expect(RESET_TOKEN.test('abc 123')).toBe(false);
  });
});
```

`client/api/account.test.ts`:

```ts
import { describe, expect, it } from 'vitest';

import { mockFetch } from '../test/mockFetch';
import {
  changePassword,
  checkResetToken,
  removeAccountImage,
  resetPassword,
  uploadAccountImage,
} from './account';

const me = {
  id: 1,
  name: 'Jo',
  email: 'jo@example.test',
  role: 'Manager',
  imageUrl: '/files/user/1',
  permissions: { canEdit: false, canManageGallery: false, canManageUsers: false },
};

describe('account API', () => {
  it('uploads the photo as multipart field "image" with PUT', async () => {
    const fetchMock = mockFetch({ '/api/v1/account/image': { body: me } });
    const file = new File(['x'], 'me.png', { type: 'image/png' });
    await expect(uploadAccountImage(file)).resolves.toEqual(me);
    const [, init] = fetchMock.mock.calls[0];
    expect(init?.method).toBe('PUT');
    expect(init?.body).toBeInstanceOf(FormData);
    expect((init?.body as FormData).get('image')).toBe(file);
  });

  it('removes the photo with DELETE', async () => {
    const fetchMock = mockFetch({ '/api/v1/account/image': { status: 204 } });
    await removeAccountImage();
    expect(fetchMock.mock.calls[0][1]?.method).toBe('DELETE');
  });

  it('changes the password with the exact JSON', async () => {
    const fetchMock = mockFetch({ '/api/v1/auth/password': { status: 204 } });
    await changePassword({ oldPassword: 'a', newPassword: 'b', confirmationPassword: 'b' });
    const [, init] = fetchMock.mock.calls[0];
    expect(init?.method).toBe('POST');
    expect(JSON.parse(String(init?.body))).toEqual({
      oldPassword: 'a',
      newPassword: 'b',
      confirmationPassword: 'b',
    });
  });

  it('checks and uses a reset token', async () => {
    const fetchMock = mockFetch({ '/api/v1/auth/reset/abc-123': { status: 204 } });
    await checkResetToken('abc-123');
    await resetPassword('abc-123', { newPassword: 'x', confirmationPassword: 'x' });
    expect(fetchMock.mock.calls[0][1]?.method).toBe('GET');
    expect(fetchMock.mock.calls[1][1]?.method).toBe('POST');
    expect(JSON.parse(String(fetchMock.mock.calls[1][1]?.body))).toEqual({
      newPassword: 'x',
      confirmationPassword: 'x',
    });
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/lib/resetLink.test.ts client/api/account.test.ts`
Expected: FAIL. `./resetLink` and `./account` can't be resolved.

- [ ] **Step 3: Implement**

`client/lib/resetLink.ts`:

```ts
/** The shape of a password-reset token (the server issues UUIDs). */
export const RESET_TOKEN = /^[A-Za-z0-9-]{1,100}$/;

/**
 * The server's reset URL ("/reset/<token>") as an in-app route, or null when
 * it isn't one. Until the cutover the SPA lives under /app, so the router adds
 * the basename; afterwards the same path is the SPA route.
 */
export function appResetPath(url: string): string | null {
  const match = /^\/reset\/([^/]+)$/.exec(url);
  return match && RESET_TOKEN.test(match[1]) ? `/reset/${match[1]}` : null;
}
```

`client/api/account.ts`:

```ts
import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import type { CurrentUser } from './types';

/** auth.PasswordInput — POST /auth/password */
export interface PasswordChange {
  oldPassword: string;
  newPassword: string;
  confirmationPassword: string;
}

/** auth.ResetInput — POST /auth/reset/:token */
export interface PasswordReset {
  newPassword: string;
  confirmationPassword: string;
}

export function uploadAccountImage(file: File): Promise<CurrentUser> {
  const form = new FormData();
  form.append('image', file);
  return apiFetch<CurrentUser>('/account/image', { method: 'PUT', form });
}

export function removeAccountImage(): Promise<void> {
  return apiFetch<void>('/account/image', { method: 'DELETE' });
}

export function changePassword(input: PasswordChange): Promise<void> {
  return apiFetch<void>('/auth/password', { json: input });
}

export function checkResetToken(token: string, signal?: AbortSignal): Promise<void> {
  return apiFetch<void>(`/auth/reset/${encodeURIComponent(token)}`, { signal });
}

export function resetPassword(token: string, input: PasswordReset): Promise<void> {
  return apiFetch<void>(`/auth/reset/${encodeURIComponent(token)}`, { json: input });
}

/** Whether a reset link is still usable; disabled (never fetched) for a null token. */
export function useResetTokenCheck(token: string | null) {
  return useQuery({
    queryKey: ['auth', 'reset', token ?? ''],
    queryFn: async ({ signal }) => {
      await checkResetToken(token as string, signal);
      return true; // query data can't be undefined
    },
    enabled: token !== null,
    retry: false,
    gcTime: 0,
  });
}
```

Append to `client/test/fixtures.ts`, before `publicRoutes`:

```ts
export const resetToken = '5f0b6c1e-1a2b-4c3d-8e9f-000000000001';
export const expiredResetToken = '5f0b6c1e-1a2b-4c3d-8e9f-00000000dead';
```

Then add these entries to the object returned by `publicRoutes`, before `...overrides`:

```ts
    [`/api/v1/auth/reset/${resetToken}`]: { status: 204 },
    [`/api/v1/auth/reset/${expiredResetToken}`]: {
      status: 404,
      body: { error: { code: 404, message: 'reset link is invalid or has expired' } },
    },
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/api client/lib client/test && yarn vitest run client/lib/resetLink.test.ts client/api/account.test.ts`
Expected: PASS, 7 tests.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add client/api/account.ts client/api/account.test.ts client/lib/resetLink.ts client/lib/resetLink.test.ts client/test/fixtures.ts
git commit -q -m "Add account API calls and the in-app reset link mapping" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 2: One app-wide sign-in dialog, with in-app reset redirects

**Files:**
- Create: `client/components/layout/signInContext.ts`, `client/components/layout/useSignIn.ts`, `client/components/layout/SignInProvider.tsx`
- Modify: `client/components/layout/SignInDialog.tsx`, `client/components/layout/AccountControl.tsx`, `client/main.tsx`, `client/test/render.tsx`, `client/components/layout/AccountControl.test.tsx`

**Interfaces:**
- Consumes: `appResetPath` (Task 1).
- Produces:
  - `SignInProvider({ children })`
  - `useSignIn(): { open(): void }`
  - `renderWithProviders` now includes `SignInProvider` inside `ToastProvider`

- [ ] **Step 1: Update the tests first**

In `client/components/layout/AccountControl.test.tsx`:

(a) Replace the test `'sends reset-flagged accounts to the reset page'` with:

```tsx
  it('sends reset-flagged accounts to the in-app reset page', async () => {
    mockFetch({
      '/api/v1/auth/me': anonymous,
      '/api/v1/auth/login': { body: { resetRequired: true, resetUrl: '/reset/abc-123' } },
    });
    renderWithProviders(
      <>
        <AccountControl />
        <Location />
      </>,
    );
    await screen.findByRole('button', { name: 'Sign in' });
    const dialog = openSignIn();
    fireEvent.click(within(dialog).getByRole('button', { name: 'Sign in' }));
    await vi.waitFor(() =>
      expect(screen.getByTestId('location')).toHaveTextContent('/reset/abc-123'),
    );
    expect(goTo).not.toHaveBeenCalled();
    expect(screen.queryByRole('dialog', { name: 'Sign in' })).toBeNull();
  });

  it('still does a full-page load for an unrecognised reset URL', async () => {
    mockFetch({
      '/api/v1/auth/me': anonymous,
      '/api/v1/auth/login': {
        body: { resetRequired: true, resetUrl: 'https://elsewhere.example/reset' },
      },
    });
    renderWithProviders(<AccountControl />);
    await screen.findByRole('button', { name: 'Sign in' });
    const dialog = openSignIn();
    fireEvent.click(within(dialog).getByRole('button', { name: 'Sign in' }));
    await vi.waitFor(() =>
      expect(goTo).toHaveBeenCalledWith('https://elsewhere.example/reset'),
    );
  });
```

(b) Add a `Location` helper after the imports, and add `import { useLocation } from 'react-router';` to the imports:

```tsx
function Location() {
  const l = useLocation();
  return <output data-testid="location">{l.pathname}</output>;
}
```

(c) In the `'AccountControl signed in'` describe, add:

```tsx
  it('links Account to the in-app page', async () => {
    await openMenuFor(user('Mo Manager', 'Manager'));
    expect(screen.getByRole('menuitem', { name: 'Account' })).toHaveAttribute('href', '/account');
    expect(screen.getByRole('menuitem', { name: 'Players' })).toHaveAttribute('href', '/players');
  });
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/components/layout/AccountControl.test.tsx`
Expected: FAIL. In "in-app reset page", `goTo` was called instead and the location stays `/`.

- [ ] **Step 3: Implement**

`client/components/layout/signInContext.ts`:

```ts
import { createContext } from 'react';

export interface SignInApi {
  /** Opens the app-wide sign-in dialog. */
  open: () => void;
}

export const SignInContext = createContext<SignInApi | null>(null);
```

`client/components/layout/useSignIn.ts`:

```ts
import { useContext } from 'react';

import { SignInContext, type SignInApi } from './signInContext';

export function useSignIn(): SignInApi {
  const ctx = useContext(SignInContext);
  if (!ctx) {
    throw new Error('useSignIn must be used inside <SignInProvider>');
  }
  return ctx;
}
```

`client/components/layout/SignInProvider.tsx`:

```tsx
import { useMemo, useState, type ReactNode } from 'react';

import { SignInContext, type SignInApi } from './signInContext';
import { SignInDialog } from './SignInDialog';

/** Owns the one sign-in dialog; anything can open it with useSignIn().open(). */
export function SignInProvider({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false);
  const api = useMemo<SignInApi>(() => ({ open: () => setOpen(true) }), []);
  return (
    <SignInContext.Provider value={api}>
      {children}
      <SignInDialog open={open} onClose={() => setOpen(false)} />
    </SignInContext.Provider>
  );
}
```

`client/components/layout/SignInDialog.tsx`:
- Add the imports `import { useNavigate } from 'react-router';` and `import { appResetPath } from '../../lib/resetLink';`.
- Add `const navigate = useNavigate();` after `const toast = useToast();`.
- Replace the reset branch with:

```tsx
      if (res.resetRequired && res.resetUrl) {
        setPassword('');
        onClose();
        const path = appResetPath(res.resetUrl);
        if (path) {
          navigate(path);
        } else {
          goTo(res.resetUrl);
        }
        return;
      }
```

`client/components/layout/AccountControl.tsx`:
- Remove the imports of `SignInDialog` and `useDisclosure`, and the `const signIn = useDisclosure();` line.
- Add `import { useSignIn } from './useSignIn';` and `const signIn = useSignIn();`.
- Replace the signed-out return with:

```tsx
  if (!user) {
    return (
      <Button variant="secondary" size="sm" onClick={signIn.open}>
        Sign in
      </Button>
    );
  }
```

- Change the Account menu item from `{ label: 'Account', href: '/account' }` to `{ label: 'Account', to: '/account' }`.

`client/test/render.tsx`: add `import { SignInProvider } from '../components/layout/SignInProvider';` and change `<ToastProvider>{node}</ToastProvider>` to:

```tsx
            <ToastProvider>
              <SignInProvider>{node}</SignInProvider>
            </ToastProvider>
```

`client/main.tsx`: add `import { SignInProvider } from './components/layout/SignInProvider';` and wrap `<App />` as `<SignInProvider><App /></SignInProvider>` inside `ToastProvider`.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/components/layout client/test client/main.tsx && yarn vitest run client/components/layout && yarn test:client`
Expected: PASS. The AccountControl tests (including the 3 new ones) pass, and the whole suite is green.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add client/components/layout client/test/render.tsx client/main.tsx
git commit -q -m "Share one sign-in dialog app-wide and route reset-flagged accounts in-app" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 3: Password fields and the reset page

**Files:**
- Create: `client/components/page/PasswordFields.tsx`, `client/pages/reset/ResetPage.tsx`
- Test: `client/pages/reset/reset.test.tsx`

**Interfaces:**
- Consumes: `resetPassword`, `useResetTokenCheck`, `RESET_TOKEN` (Task 1); `useSignIn` (Task 2); `QueryState`, `usePageTitle`, `isNotFound`, `fieldError`, `PageHeader`, `EmptyState`, `ButtonLink`, `Button`, `Alert`, `Field`, `Input`.
- Produces:
  - `type PasswordField = 'newPassword' | 'confirmationPassword'`
  - `PASSWORD_HINT`
  - `PasswordFields({ newPassword, confirmationPassword, onChange(field, value), errors })`
  - the default export `ResetPage`

- [ ] **Step 1: Write the failing tests**

`client/pages/reset/reset.test.tsx`:

```tsx
import { fireEvent, screen, within } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import { editor, expiredResetToken, publicRoutes, resetToken } from '../../test/fixtures';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import ResetPage from './ResetPage';

const invalidTitle = 'This reset link is invalid or has expired';

function renderReset(token: string, overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes(overrides));
  renderWithProviders(
    <Routes>
      <Route path="/reset/:token" element={<ResetPage />} />
    </Routes>,
    { route: `/reset/${token}` },
  );
  return fetchMock;
}

function fill(newPassword: string, confirmation: string) {
  fireEvent.change(screen.getByLabelText('New password'), { target: { value: newPassword } });
  fireEvent.change(screen.getByLabelText('Confirm new password'), {
    target: { value: confirmation },
  });
  fireEvent.click(screen.getByRole('button', { name: 'Set new password' }));
}

describe('ResetPage', () => {
  it('rejects a malformed token without calling the API', async () => {
    const fetchMock = renderReset('not%20a%20token');
    expect(await screen.findByText(invalidTitle)).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([u]) => String(u).includes('/auth/reset/'))).toBe(false);
  });

  it('explains an expired link and offers Home', async () => {
    renderReset(expiredResetToken);
    expect(await screen.findByText(invalidTitle)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Home' })).toHaveAttribute('href', '/');
    expect(screen.queryByLabelText('New password')).toBeNull();
    expect(document.title).toBe('Reset password · AFC Aldermaston');
  });

  it('shows the form with the rules hint for a valid link', async () => {
    renderReset(resetToken);
    const field = await screen.findByLabelText('New password');
    expect(field).toHaveAccessibleDescription(
      'More than 8 characters, with a lower-case letter, an upper-case letter, a number and a special character.',
    );
    expect(screen.getByRole('heading', { level: 1, name: 'Reset your password' })).toBeInTheDocument();
  });

  it('asks for empty fields without sending anything', async () => {
    const fetchMock = renderReset(resetToken);
    await screen.findByLabelText('New password');
    fireEvent.click(screen.getByRole('button', { name: 'Set new password' }));
    expect(screen.getByLabelText('New password')).toHaveAccessibleDescription('Enter a new password');
    expect(screen.getByLabelText('Confirm new password')).toHaveAccessibleDescription(
      'Confirm your new password',
    );
    expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(false);
  });

  it('shows the server field errors', async () => {
    let calls = 0;
    renderReset(resetToken, {
      [`/api/v1/auth/reset/${resetToken}`]: () =>
        calls++ === 0
          ? { status: 204 }
          : {
              status: 422,
              body: {
                error: {
                  code: 422,
                  message: 'invalid',
                  fields: {
                    newPassword: 'password needs at least 1 number',
                    confirmationPassword: 'passwords do not match',
                  },
                },
              },
            },
    });
    await screen.findByLabelText('New password');
    fill('abcdefghij', 'different');
    expect(await screen.findByText('password needs at least 1 number')).toBeInTheDocument();
    expect(screen.getByLabelText('Confirm new password')).toHaveAccessibleDescription(
      'passwords do not match',
    );
  });

  it('confirms success and offers sign-in', async () => {
    renderReset(resetToken);
    await screen.findByLabelText('New password');
    fill('Abcdefgh1!', 'Abcdefgh1!');
    const done = await screen.findByText('Your password has been changed. You can now sign in with it.');
    expect(screen.queryByLabelText('New password')).toBeNull();
    fireEvent.click(within(done.closest('[role="status"]') as HTMLElement).getByRole('button', { name: 'Sign in' }));
    expect(screen.getByRole('dialog', { name: 'Sign in' })).toBeInTheDocument();
  });

  it('switches to the invalid state if the link expires before submitting', async () => {
    let checks = 0;
    renderReset(resetToken, {
      [`/api/v1/auth/reset/${resetToken}`]: () =>
        checks++ === 0
          ? { status: 204 }
          : { status: 404, body: { error: { code: 404, message: 'expired' } } },
    });
    await screen.findByLabelText('New password');
    fill('Abcdefgh1!', 'Abcdefgh1!');
    expect(await screen.findByText(invalidTitle)).toBeInTheDocument();
  });

  it('works for a signed-in user too', async () => {
    renderReset(resetToken, { '/api/v1/auth/me': editor });
    expect(await screen.findByLabelText('New password')).toBeInTheDocument();
  });
});
```


- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/reset`
Expected: FAIL. `./ResetPage` can't be resolved.

- [ ] **Step 3: Implement**

`client/components/page/PasswordFields.tsx`:

```tsx
import { Input } from '../ui/controls';
import { Field } from '../ui/Field';

export type PasswordField = 'newPassword' | 'confirmationPassword';

export const PASSWORD_HINT =
  'More than 8 characters, with a lower-case letter, an upper-case letter, a number and a special character.';

export interface PasswordFieldsProps {
  newPassword: string;
  confirmationPassword: string;
  onChange: (field: PasswordField, value: string) => void;
  errors: Partial<Record<PasswordField, string>>;
}

/** New + confirm password inputs; the server's messages replace the rules hint. */
export function PasswordFields({
  newPassword,
  confirmationPassword,
  onChange,
  errors,
}: PasswordFieldsProps) {
  return (
    <>
      <Field label="New password" help={PASSWORD_HINT} error={errors.newPassword}>
        <Input
          type="password"
          autoComplete="new-password"
          value={newPassword}
          onChange={(e) => onChange('newPassword', e.target.value)}
        />
      </Field>
      <Field label="Confirm new password" error={errors.confirmationPassword}>
        <Input
          type="password"
          autoComplete="new-password"
          value={confirmationPassword}
          onChange={(e) => onChange('confirmationPassword', e.target.value)}
        />
      </Field>
    </>
  );
}
```

`client/pages/reset/ResetPage.tsx`:

```tsx
import { useState, type FormEvent } from 'react';
import { useParams } from 'react-router';

import { resetPassword, useResetTokenCheck } from '../../api/account';
import { ApiError } from '../../api/client';
import { useSignIn } from '../../components/layout/useSignIn';
import { PasswordFields, type PasswordField } from '../../components/page/PasswordFields';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { ButtonLink } from '../../components/ui/ButtonLink';
import { EmptyState } from '../../components/ui/EmptyState';
import { fieldError } from '../../components/ui/fieldError';
import { PageHeader } from '../../components/ui/PageHeader';
import { isNotFound } from '../../lib/notFound';
import { RESET_TOKEN } from '../../lib/resetLink';

type Errors = Partial<Record<PasswordField, string>>;

function InvalidLink() {
  return (
    <EmptyState
      title="This reset link is invalid or has expired"
      message="Ask a club administrator for a new one, or sign in again if you were sent here after signing in."
      action={<ButtonLink to="/">Home</ButtonLink>}
    />
  );
}

export default function ResetPage() {
  usePageTitle('Reset password');
  const raw = useParams().token ?? '';
  const token = RESET_TOKEN.test(raw) ? raw : null;
  const check = useResetTokenCheck(token);
  const signIn = useSignIn();
  const [values, setValues] = useState({ newPassword: '', confirmationPassword: '' });
  const [errors, setErrors] = useState<Errors>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);
  const [expired, setExpired] = useState(false);

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (token === null) {
      return;
    }
    const empty: Errors = {};
    if (!values.newPassword) empty.newPassword = 'Enter a new password';
    if (!values.confirmationPassword) empty.confirmationPassword = 'Confirm your new password';
    if (Object.keys(empty).length > 0) {
      setErrors(empty);
      return;
    }
    setBusy(true);
    setErrors({});
    setFormError(null);
    try {
      await resetPassword(token, values);
      setDone(true);
    } catch (err) {
      if (isNotFound(err)) {
        setExpired(true);
      } else if (err instanceof ApiError && Object.keys(err.fields).length > 0) {
        setErrors({
          newPassword: fieldError(err, 'newPassword'),
          confirmationPassword: fieldError(err, 'confirmationPassword'),
        });
      } else {
        setFormError(err instanceof Error ? err.message : 'Something went wrong. Please try again.');
      }
    } finally {
      setBusy(false);
    }
  }

  const header = <PageHeader title="Reset your password" />;

  if (token === null || expired || isNotFound(check.error)) {
    return (
      <>
        {header}
        <InvalidLink />
      </>
    );
  }

  if (done) {
    return (
      <>
        {header}
        <Alert tone="success" className="flex flex-wrap items-center justify-between gap-3">
          <span>Your password has been changed. You can now sign in with it.</span>
          <Button onClick={signIn.open}>Sign in</Button>
        </Alert>
      </>
    );
  }

  return (
    <>
      {header}
      <QueryState query={check}>
        {() => (
          <form onSubmit={onSubmit} className="flex max-w-md flex-col gap-4" noValidate>
            {formError && <Alert tone="error">{formError}</Alert>}
            <PasswordFields
              newPassword={values.newPassword}
              confirmationPassword={values.confirmationPassword}
              onChange={(field, value) => setValues((v) => ({ ...v, [field]: value }))}
              errors={errors}
            />
            <div>
              <Button type="submit" loading={busy}>
                Set new password
              </Button>
            </div>
          </form>
        )}
      </QueryState>
    </>
  );
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/components/page client/pages/reset && yarn vitest run client/pages/reset`
Expected: PASS, 8 tests.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add client/components/page/PasswordFields.tsx client/pages/reset
git commit -q -m "Add the password reset-link page" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 4: The Account page

**Files:**
- Create in `client/pages/account/`: `AccountPage.tsx`, `DetailsCard.tsx`, `PhotoCard.tsx`, `PasswordCard.tsx`
- Test: `client/pages/account/account.test.tsx`

**Interfaces:**
- Consumes:
  - `uploadAccountImage`, `removeAccountImage`, `changePassword` (Task 1);
  - `useSignIn` (Task 2);
  - `PasswordFields`, `PasswordField` (Task 3);
  - `useAuth`, `useSite`, `queryKeys`, `ImageWithFallback`, `ConfirmDialog`, `useToast`, `fieldError`, `PageSkeleton`, `Card`, `CardBody`, `Field`, `FileInput`, `Input`, `Button`, `Alert`, `EmptyState`, `PageHeader`, `usePageTitle`, `crest`.
- Produces: the default export `AccountPage`.

- [ ] **Step 1: Write the failing tests**

`client/pages/account/account.test.tsx`:

```tsx
import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { anonymous, publicRoutes, team } from '../../test/fixtures';
import type { MockResponse, MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import AccountPage from './AccountPage';

const baseUser = {
  id: 1,
  name: 'Jo Smith',
  email: 'jo@example.test',
  role: 'Manager',
  permissions: { canEdit: false, canManageGallery: false, canManageUsers: false },
};

function renderAccount(me: MockResponse | (() => MockResponse), overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': me, ...overrides }));
  renderWithProviders(<AccountPage />, { route: '/account' });
  return fetchMock;
}

let createObjectURL: ReturnType<typeof vi.fn>;
let revokeObjectURL: ReturnType<typeof vi.fn>;
beforeEach(() => {
  let n = 0;
  createObjectURL = vi.fn(() => `blob:preview-${++n}`);
  revokeObjectURL = vi.fn();
  Object.assign(URL, { createObjectURL, revokeObjectURL });
});
afterEach(() => {
  delete (URL as unknown as Record<string, unknown>).createObjectURL;
  delete (URL as unknown as Record<string, unknown>).revokeObjectURL;
});

function photoCard() {
  return screen.getByRole('region', { name: 'Photo' });
}

function choose(name = 'me.png') {
  const file = new File(['x'], name, { type: 'image/png' });
  fireEvent.change(screen.getByLabelText('Choose a new photo'), { target: { files: [file] } });
  return file;
}

describe('AccountPage signed out', () => {
  it('prompts to sign in and shows the account after signing in', async () => {
    let me: MockResponse = anonymous;
    renderAccount(() => me, {
      '/api/v1/auth/login': () => {
        me = { body: baseUser };
        return { body: { user: baseUser, resetRequired: false } };
      },
    });
    expect(await screen.findByText('Sign in to see your account')).toBeInTheDocument();
    fireEvent.click(screen.getAllByRole('button', { name: 'Sign in' })[0]);
    const dialog = screen.getByRole('dialog', { name: 'Sign in' });
    fireEvent.change(within(dialog).getByLabelText('Email'), { target: { value: 'jo@example.test' } });
    fireEvent.change(within(dialog).getByLabelText('Password'), { target: { value: 'pw' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Sign in' }));
    expect(await screen.findByText('jo@example.test')).toBeInTheDocument();
    expect(document.title).toBe('Account · AFC Aldermaston');
  });
});

describe('AccountPage details', () => {
  it('shows details, and phone/team only when set', async () => {
    renderAccount({ body: { ...baseUser, phone: '07123 456789', teamId: team.id } });
    const details = await screen.findByRole('region', { name: 'Your details' });
    expect(within(details).getByText('Jo Smith')).toBeInTheDocument();
    expect(within(details).getByText('07123 456789')).toBeInTheDocument();
    expect(await within(details).findByText(team.name)).toBeInTheDocument();
    expect(within(details).getByText('To change these, ask a club administrator.')).toBeInTheDocument();
  });

  it('leaves out phone and team when not set', async () => {
    renderAccount({ body: baseUser });
    const details = await screen.findByRole('region', { name: 'Your details' });
    expect(within(details).queryByText('Phone')).toBeNull();
    expect(within(details).queryByText('Team')).toBeNull();
  });
});

describe('AccountPage photo', () => {
  it('previews, saves as multipart "image", updates and toasts', async () => {
    const updated = { ...baseUser, imageUrl: '/files/user/1?v=2' };
    const fetchMock = renderAccount({ body: baseUser }, {
      '/api/v1/account/image': { body: updated },
    });
    await screen.findByRole('region', { name: 'Photo' });
    const file = choose();
    expect(within(photoCard()).getByRole('img', { name: 'Preview of your new photo' })).toHaveAttribute(
      'src',
      'blob:preview-1',
    );
    fireEvent.click(within(photoCard()).getByRole('button', { name: 'Save photo' }));
    expect(await screen.findByRole('button', { name: 'Photo updated' })).toBeInTheDocument();
    const put = fetchMock.mock.calls.find(([, init]) => init?.method === 'PUT');
    expect((put?.[1]?.body as FormData).get('image')).toBe(file);
    expect(photoCard().querySelector('img')).toHaveAttribute('src', '/files/user/1?v=2');
    expect(revokeObjectURL).toHaveBeenCalledWith('blob:preview-1');
  });

  it('replaces the preview when choosing again, and Cancel discards it', async () => {
    renderAccount({ body: { ...baseUser, imageUrl: '/files/user/1' } });
    await screen.findByRole('region', { name: 'Photo' });
    choose('a.png');
    choose('b.png');
    expect(revokeObjectURL).toHaveBeenCalledWith('blob:preview-1');
    expect(within(photoCard()).getByRole('img', { name: 'Preview of your new photo' })).toHaveAttribute(
      'src',
      'blob:preview-2',
    );
    fireEvent.click(within(photoCard()).getByRole('button', { name: 'Cancel' }));
    expect(revokeObjectURL).toHaveBeenCalledWith('blob:preview-2');
    expect(photoCard().querySelector('img')).toHaveAttribute('src', '/files/user/1');
  });

  it('shows an upload error under the file field', async () => {
    renderAccount({ body: baseUser }, {
      '/api/v1/account/image': {
        status: 422,
        body: { error: { code: 422, message: 'invalid', fields: { file: 'file is not an image' } } },
      },
    });
    await screen.findByRole('region', { name: 'Photo' });
    choose();
    fireEvent.click(within(photoCard()).getByRole('button', { name: 'Save photo' }));
    await waitFor(() =>
      expect(screen.getByLabelText('Choose a new photo')).toHaveAccessibleDescription(
        'file is not an image',
      ),
    );
  });

  it('removes the photo after confirmation', async () => {
    let me: MockResponse = { body: { ...baseUser, imageUrl: '/files/user/1' } };
    const fetchMock = renderAccount(() => me, {
      '/api/v1/account/image': () => {
        me = { body: baseUser };
        return { status: 204 };
      },
    });
    await screen.findByRole('region', { name: 'Photo' });
    fireEvent.click(within(photoCard()).getByRole('button', { name: 'Remove photo' }));
    const confirm = screen.getByRole('dialog', { name: 'Remove your photo?' });
    fireEvent.click(within(confirm).getByRole('button', { name: 'Remove photo' }));
    expect(await screen.findByRole('button', { name: 'Photo removed' })).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'DELETE')).toBe(true);
    await waitFor(() =>
      expect(within(photoCard()).queryByRole('button', { name: 'Remove photo' })).toBeNull(),
    );
  });

  it('falls back to the crest when the photo is broken', async () => {
    renderAccount({ body: { ...baseUser, imageUrl: '/files/user/gone' } });
    await screen.findByRole('region', { name: 'Photo' });
    fireEvent.error(photoCard().querySelector('img') as HTMLImageElement);
    expect(photoCard().querySelector('img')).toHaveAttribute('src', expect.stringContaining('crest'));
  });

  it('hides Remove when there is no photo', async () => {
    renderAccount({ body: baseUser });
    await screen.findByRole('region', { name: 'Photo' });
    expect(within(photoCard()).queryByRole('button', { name: 'Remove photo' })).toBeNull();
  });
});

describe('AccountPage change password', () => {
  function fillPassword(old: string, next: string, confirm: string) {
    fireEvent.change(screen.getByLabelText('Current password'), { target: { value: old } });
    fireEvent.change(screen.getByLabelText('New password'), { target: { value: next } });
    fireEvent.change(screen.getByLabelText('Confirm new password'), { target: { value: confirm } });
    fireEvent.click(screen.getByRole('button', { name: 'Change password' }));
  }

  it('sends the exact JSON, clears the form and toasts', async () => {
    const fetchMock = renderAccount({ body: baseUser }, { '/api/v1/auth/password': { status: 204 } });
    await screen.findByLabelText('Current password');
    fillPassword('old-pass', 'Abcdefgh1!', 'Abcdefgh1!');
    expect(await screen.findByRole('button', { name: 'Password changed' })).toBeInTheDocument();
    const post = fetchMock.mock.calls.find(([u]) => String(u).endsWith('/auth/password'));
    expect(JSON.parse(String(post?.[1]?.body))).toEqual({
      oldPassword: 'old-pass',
      newPassword: 'Abcdefgh1!',
      confirmationPassword: 'Abcdefgh1!',
    });
    expect(screen.getByLabelText('Current password')).toHaveValue('');
    expect(screen.getByLabelText('New password')).toHaveValue('');
  });

  it('puts each server message on its field', async () => {
    renderAccount({ body: baseUser }, {
      '/api/v1/auth/password': {
        status: 422,
        body: {
          error: {
            code: 422,
            message: 'invalid',
            fields: {
              oldPassword: 'old password is not correct',
              newPassword: 'password needs at least 1 number',
              confirmationPassword: 'passwords do not match',
            },
          },
        },
      },
    });
    await screen.findByLabelText('Current password');
    fillPassword('x', 'y', 'z');
    await waitFor(() =>
      expect(screen.getByLabelText('Current password')).toHaveAccessibleDescription(
        'old password is not correct',
      ),
    );
    expect(screen.getByLabelText('New password')).toHaveAccessibleDescription(
      'password needs at least 1 number',
    );
    expect(screen.getByLabelText('Confirm new password')).toHaveAccessibleDescription(
      'passwords do not match',
    );
  });

  it('asks for empty fields without sending anything', async () => {
    const fetchMock = renderAccount({ body: baseUser });
    await screen.findByLabelText('Current password');
    fireEvent.click(screen.getByRole('button', { name: 'Change password' }));
    expect(screen.getByLabelText('Current password')).toHaveAccessibleDescription(
      'Enter your current password',
    );
    expect(screen.getByLabelText('New password')).toHaveAccessibleDescription('Enter a new password');
    expect(fetchMock.mock.calls.some(([u]) => String(u).endsWith('/auth/password'))).toBe(false);
  });

  it('shows the button as busy while the request runs', async () => {
    let answer: (r: MockResponse) => void = () => {};
    renderAccount({ body: baseUser }, {
      '/api/v1/auth/password': () =>
        new Promise<MockResponse>((resolve) => {
          answer = resolve;
        }),
    });
    await screen.findByLabelText('Current password');
    fillPassword('a', 'Abcdefgh1!', 'Abcdefgh1!');
    const button = screen.getByRole('button', { name: 'Change password' });
    await waitFor(() => expect(button).toHaveAttribute('aria-busy', 'true'));
    answer({ status: 204 });
    await waitFor(() => expect(button).not.toHaveAttribute('aria-busy'));
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/account`
Expected: FAIL. `./AccountPage` can't be resolved.

- [ ] **Step 3: Implement**

`client/pages/account/DetailsCard.tsx`:

```tsx
import { useId } from 'react';

import { useSite } from '../../api/queries';
import type { CurrentUser } from '../../api/types';
import { Card, CardBody } from '../../components/ui/Card';

export function DetailsCard({ user }: { user: CurrentUser }) {
  const headingId = useId();
  const site = useSite();
  const teamName = user.teamId
    ? site.data?.teams.find((t) => t.id === user.teamId)?.name
    : undefined;
  const rows: [string, string | undefined][] = [
    ['Name', user.name],
    ['Email', user.email],
    ['Phone', user.phone],
    ['Role', user.role],
    ['Team', teamName],
  ];
  return (
    <Card>
      <CardBody>
        <section aria-labelledby={headingId} className="space-y-4">
          <h2 id={headingId} className="font-display text-2xl font-extrabold tracking-wide uppercase">
            Your details
          </h2>
          <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 text-sm">
            {rows
              .filter(([, value]) => value)
              .map(([label, value]) => (
                <div key={label} className="contents">
                  <dt className="text-muted">{label}</dt>
                  <dd>{value}</dd>
                </div>
              ))}
          </dl>
          <p className="text-sm text-muted">To change these, ask a club administrator.</p>
        </section>
      </CardBody>
    </Card>
  );
}
```

`client/pages/account/PhotoCard.tsx`:

```tsx
import { useQueryClient } from '@tanstack/react-query';
import { useEffect, useId, useRef, useState } from 'react';

import { removeAccountImage, uploadAccountImage } from '../../api/account';
import { queryKeys } from '../../api/queries';
import type { CurrentUser } from '../../api/types';
import crest from '../../assets/crest.png';
import { useAuth } from '../../auth/useAuth';
import { ImageWithFallback } from '../../components/page/ImageWithFallback';
import { Button } from '../../components/ui/Button';
import { Card, CardBody } from '../../components/ui/Card';
import { ConfirmDialog } from '../../components/ui/ConfirmDialog';
import { FileInput } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { fieldError } from '../../components/ui/fieldError';
import { useToast } from '../../components/ui/toast/useToast';

const IMAGE_TYPES =
  'image/jpeg,image/png,image/gif,image/webp,image/avif,image/apng,image/svg+xml';

const photoClass = 'size-32 rounded-full border border-line bg-white object-cover';

export function PhotoCard({ user }: { user: CurrentUser }) {
  const headingId = useId();
  const queryClient = useQueryClient();
  const { refresh } = useAuth();
  const toast = useToast();
  const inputRef = useRef<HTMLInputElement>(null);
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const [error, setError] = useState<string | undefined>();
  const [saving, setSaving] = useState(false);
  const [confirmRemove, setConfirmRemove] = useState(false);

  // Release each preview URL when it is replaced, cleared, or the card unmounts.
  useEffect(() => {
    if (!preview) {
      return;
    }
    return () => URL.revokeObjectURL(preview);
  }, [preview]);

  function choose(next: File | null) {
    setError(undefined);
    setFile(next);
    setPreview(next ? URL.createObjectURL(next) : null);
    if (!next && inputRef.current) {
      inputRef.current.value = '';
    }
  }

  async function save() {
    if (!file) {
      return;
    }
    setSaving(true);
    setError(undefined);
    try {
      const updated = await uploadAccountImage(file);
      queryClient.setQueryData(queryKeys.me, updated);
      choose(null);
      toast.show({ tone: 'success', message: 'Photo updated' });
    } catch (err) {
      setError(
        fieldError(err, 'file') ??
          fieldError(err, 'image') ??
          (err instanceof Error ? err.message : 'Upload failed.'),
      );
    } finally {
      setSaving(false);
    }
  }

  async function remove() {
    try {
      await removeAccountImage();
      await refresh();
      toast.show({ tone: 'success', message: 'Photo removed' });
    } catch (err) {
      toast.show({
        tone: 'error',
        message: `Couldn't remove your photo: ${err instanceof Error ? err.message : 'unknown error'}`,
      });
    } finally {
      setConfirmRemove(false);
    }
  }

  return (
    <Card>
      <CardBody>
        <section aria-labelledby={headingId} className="space-y-4">
          <h2 id={headingId} className="font-display text-2xl font-extrabold tracking-wide uppercase">
            Photo
          </h2>
          {preview ? (
            <img src={preview} alt="Preview of your new photo" className={photoClass} />
          ) : (
            <ImageWithFallback src={user.imageUrl} fallbackSrc={crest} alt="" className={photoClass} />
          )}
          <Field label="Choose a new photo" error={error}>
            <FileInput
              ref={inputRef}
              accept={IMAGE_TYPES}
              onChange={(e) => choose(e.target.files?.[0] ?? null)}
            />
          </Field>
          <div className="flex flex-wrap gap-2">
            {file ? (
              <>
                <Button onClick={save} loading={saving}>
                  Save photo
                </Button>
                <Button variant="secondary" onClick={() => choose(null)} disabled={saving}>
                  Cancel
                </Button>
              </>
            ) : (
              user.imageUrl && (
                <Button variant="danger" onClick={() => setConfirmRemove(true)}>
                  Remove photo
                </Button>
              )
            )}
          </div>
        </section>
        <ConfirmDialog
          open={confirmRemove}
          title="Remove your photo?"
          message="Your photo will be replaced by the club crest."
          confirmLabel="Remove photo"
          tone="danger"
          onConfirm={remove}
          onCancel={() => setConfirmRemove(false)}
        />
      </CardBody>
    </Card>
  );
}
```

`client/pages/account/PasswordCard.tsx`:

```tsx
import { useId, useState, type FormEvent } from 'react';

import { changePassword } from '../../api/account';
import { ApiError } from '../../api/client';
import { PasswordFields } from '../../components/page/PasswordFields';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Card, CardBody } from '../../components/ui/Card';
import { Input } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { fieldError } from '../../components/ui/fieldError';
import { useToast } from '../../components/ui/toast/useToast';

type Key = 'oldPassword' | 'newPassword' | 'confirmationPassword';
const blank: Record<Key, string> = { oldPassword: '', newPassword: '', confirmationPassword: '' };
const emptyMessages: Record<Key, string> = {
  oldPassword: 'Enter your current password',
  newPassword: 'Enter a new password',
  confirmationPassword: 'Confirm your new password',
};

export function PasswordCard() {
  const headingId = useId();
  const toast = useToast();
  const [values, setValues] = useState(blank);
  const [errors, setErrors] = useState<Partial<Record<Key, string>>>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const set = (key: Key, value: string) => setValues((v) => ({ ...v, [key]: value }));

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const empty: Partial<Record<Key, string>> = {};
    for (const key of Object.keys(blank) as Key[]) {
      if (!values[key]) {
        empty[key] = emptyMessages[key];
      }
    }
    if (Object.keys(empty).length > 0) {
      setErrors(empty);
      return;
    }
    setBusy(true);
    setErrors({});
    setFormError(null);
    try {
      await changePassword(values);
      setValues(blank);
      toast.show({ tone: 'success', message: 'Password changed' });
    } catch (err) {
      if (err instanceof ApiError && Object.keys(err.fields).length > 0) {
        setErrors({
          oldPassword: fieldError(err, 'oldPassword'),
          newPassword: fieldError(err, 'newPassword'),
          confirmationPassword: fieldError(err, 'confirmationPassword'),
        });
      } else {
        setFormError(err instanceof Error ? err.message : 'Something went wrong. Please try again.');
      }
    } finally {
      setBusy(false);
    }
  }

  return (
    <Card>
      <CardBody>
        <section aria-labelledby={headingId}>
          <h2
            id={headingId}
            className="mb-4 font-display text-2xl font-extrabold tracking-wide uppercase"
          >
            Change password
          </h2>
          <form onSubmit={onSubmit} className="flex max-w-md flex-col gap-4" noValidate>
            <Field label="Current password" error={errors.oldPassword}>
              <Input
                type="password"
                autoComplete="current-password"
                value={values.oldPassword}
                onChange={(e) => set('oldPassword', e.target.value)}
              />
            </Field>
            <PasswordFields
              newPassword={values.newPassword}
              confirmationPassword={values.confirmationPassword}
              onChange={set}
              errors={errors}
            />
            {formError && <Alert tone="error">{formError}</Alert>}
            <div>
              <Button type="submit" loading={busy}>
                Change password
              </Button>
            </div>
          </form>
        </section>
      </CardBody>
    </Card>
  );
}
```

`client/pages/account/AccountPage.tsx`:

```tsx
import { useAuth } from '../../auth/useAuth';
import { useSignIn } from '../../components/layout/useSignIn';
import { PageSkeleton } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Button } from '../../components/ui/Button';
import { EmptyState } from '../../components/ui/EmptyState';
import { PageHeader } from '../../components/ui/PageHeader';
import { DetailsCard } from './DetailsCard';
import { PasswordCard } from './PasswordCard';
import { PhotoCard } from './PhotoCard';

export default function AccountPage() {
  usePageTitle('Account');
  const { user, isLoading } = useAuth();
  const signIn = useSignIn();
  return (
    <>
      <PageHeader title="Your account" />
      {isLoading ? (
        <PageSkeleton />
      ) : !user ? (
        <EmptyState
          title="Sign in to see your account"
          message="Your session may have expired."
          action={<Button onClick={signIn.open}>Sign in</Button>}
        />
      ) : (
        <div className="grid gap-6 md:grid-cols-2">
          <PhotoCard user={user} />
          <DetailsCard user={user} />
          <div className="md:col-span-2">
            <PasswordCard />
          </div>
        </div>
      )}
    </>
  );
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/pages/account && yarn vitest run client/pages/account`
Expected: PASS, 14 tests. The signed-out test clicks the **first** "Sign in" button in the DOM; in this isolated render (no header), that's the page's own button.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add client/pages/account
git commit -q -m "Add the account page: details, photo upload/remove, change password" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 5: Routes, accessibility checks and docs

**Files:**
- Modify: `client/App.tsx`, `client/App.test.tsx`, `client/a11y.test.tsx`, `README.md`

**Interfaces:**
- Consumes: `AccountPage` and `ResetPage` default exports, and the fixtures `resetToken`, `expiredResetToken`, `editor`.

- [ ] **Step 1: Write the failing tests**

In `client/App.test.tsx`:
- import `resetToken` from `./test/fixtures`;
- add these two cases to the `it.each` route table:

```tsx
    ['/account', 'Your account'],
    [`/reset/${resetToken}`, 'Reset your password'],
```

In `client/a11y.test.tsx`:
- import `resetToken` and `expiredResetToken`;
- add `'/account'`, `` `/reset/${resetToken}` `` and `` `/reset/${expiredResetToken}` `` to `routes`;
- because signed-out `/account` has two "Sign in" buttons (header and page), change the account-control wait to:

```tsx
    await screen.findAllByRole('button', { name: me === anonymous ? 'Sign in' : /Ed Editor/ });
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/App.test.tsx`
Expected: FAIL. `/account` renders "Page not found", and the reset route does too.

- [ ] **Step 3: Implement**

In `client/App.tsx`, add the lazy imports:

```tsx
const AccountPage = lazy(() => import('./pages/account/AccountPage'));
const ResetPage = lazy(() => import('./pages/reset/ResetPage'));
```

and the routes, before `design`:

```tsx
        <Route path="account" element={<AccountPage />} />
        <Route path="reset/:token" element={<ResetPage />} />
```

In `README.md`, under "Public pages", add:

```markdown
### Account pages

`/app/account` shows the signed-in user's details and lets them change their photo and password. `/app/reset/<token>` is the password reset page. Sign-in sends reset-flagged accounts there; the admin reset email still links to the classic `/reset/<token>` page until the cutover, when that path becomes the SPA's.
```

- [ ] **Step 4: Run the tests to verify they pass, then run the full verification**

Run: `yarn eslint --fix client README.md 2>/dev/null; yarn vitest run client/App.test.tsx client/a11y.test.tsx`
Expected: PASS. 18 App tests; 72 a11y tests (18 routes × signed in/out × 2 themes) with no axe violations. If axe reports violations, fix the markup; don't disable rules.

Then run each of these and read the output:

```bash
yarn lint
yarn typecheck
yarn test:client
yarn build:client
yarn test:server
```

Expected: all exit 0.

- [ ] **Step 5: Commit**

```bash
git add -A client README.md
git commit -q -m "Route the account and reset pages, and extend axe checks" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```
