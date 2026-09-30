# Players and Users Admin Pages Implementation Plan (Sub-project 4c-2)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move the Players and Users admin screens, including the public contact email and admin password resets, into the React client at `/app`.

**Architecture:**
- Two lazily loaded pages, `PlayersPage` and `UsersPage`. Each has a filterable design-system `Table`, and dialogs for adding and editing.
- Both are built from the 4c-1 editing toolkit (`useSaveForm`, `ImageField`, `DeleteButton`, `RequireEditor`) plus a few new pieces:
  - `RequireSignIn`, a `Thumb` component and `lib/roles.ts`;
  - `SecretDialog`, which shows a temporary password or reset link with Copy.
- New API modules `players.ts` and `users.ts`, and `setDisplayEmail` in `pages.ts`. There are no server changes.

**Tech Stack:** React 19, React Router 7, TanStack Query 5, TypeScript 6, Tailwind 4, Vitest 5, Testing Library and jsdom, axe-core.

**Spec:** `docs/superpowers/specs/2026-09-30-admin-pages-design.md`

## Global Constraints

- **Where to work:**
  - The worktree `/Users/liam/Code/Go/AFC-design-system`, branch `admin-pages`.
  - Never touch `/Users/liam/Code/Go/AFC`.
  - Never read, open, grep or copy `postgres_*.sql`.
  - Never `rm -rf` the current working directory.
- **Commits:** every commit message ends with a blank line and `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
- **Scope:** no Go changes and no new dependencies.
- **Multipart field names:**
  - players: `name`, `teamId`, `dateOfBirth` (YYYY-MM-DD), `position`, `isCaptain`, `image`, `removeImage`;
  - users: `name`, `email`, `phone`, `role` (code), `teamId`, `image`, `removeImage`;
  - display email: JSON `{ email }` sent with PUT to `/settings/display-email`.
- **Role codes, in this order:** `photographer`, `manager`, `programme_editor`, `league_secretary`, `treasurer`, `safeguarding_officer`, `club_secretary`, `chairperson`, `webmaster`. Their labels are Photographer, Manager, Programme Editor, League Secretary, Treasurer, Safeguarding Officer, Club Secretary, Chairperson, Webmaster.
- **Permissions:**
  - Players can be viewed by anyone signed in; Add, Edit and Delete need `canEdit`.
  - Users needs `canManageUsers`.
- **Your own row on Users:** no Delete button; the Role select is disabled; `role` is not sent when you save it.
- **Copy text:**
  - **Toasts:** "Player saved", "Player deleted", "User saved", "User added. They've been emailed a temporary password.", "User deleted", "Reset link emailed to <email>", "Contact email saved", "Copied", "Select and copy the text above", "Couldn't reset the password: <message>".
  - **Sign-in gate:** "Sign in to see the players".
  - **Your own row:** the Role help text "Ask another administrator to change your role."
  - **Duplicate email:** "That email address is already in use".
  - **Photo help text:** "Photos are never shown for youth-team or under-18 players."
- **Style:** token classes only, one component per `.tsx`, and no barrel files.
- **Before each task's tests:** run `yarn eslint --fix <files>`. Then commit.
- **Test helpers:**
  - `mockFetch` route functions receive the request's `RequestInit`, so a route can answer by method.
  - `fetchMock.mock.calls[i][1]` is the `RequestInit`.
  - A `<dialog>` stays in the DOM when closed, so assert with `queryByRole('dialog', { name })` returning null.
  - `Modal` renders its children only while open.

## Review Focus

1. **Editing another user and changing their role away from Manager:** no `teamId` is sent. Pinned in Task 4.
2. **Editing your own row:** no `role` is sent, even though the select shows your role. Pinned in Task 4.
3. **The browser has no clipboard API** (plain http, or permission denied): Copy selects the text and says so instead of failing silently. Pinned in Task 4.
4. **Unticking Captain on edit:** `isCaptain=false` is sent explicitly. Pinned in Task 2 and Task 3.
5. **Opening Add straight after closing Edit:** the dialog starts empty, not filled with the last user or player. Pinned in Task 3 and Task 4.

---

### Task 1: Foundations: roles, thumbnail, sign-in gate, permissions, fixtures and query keys

**Files:**
- Create: `client/lib/roles.ts`, `client/components/page/Thumb.tsx`, `client/components/edit/RequireSignIn.tsx`
- Modify:
  - `client/components/edit/RequireEditor.tsx` (adds the `canManageUsers` permission)
  - `client/pages/account/AccountPage.tsx` (uses `RequireSignIn`)
  - `client/api/queries.ts` (adds the `players` and `users` keys)
  - `client/test/fixtures.ts` (`userAdmin`, `players`, `adminUsers`, routes)
- Test: `client/components/edit/gates.test.tsx`

**Interfaces:**
- Produces:
  - `ROLES: { code: string; label: string }[]`, `MANAGER_ROLE = 'manager'`
  - `Thumb({ src?: string })`
  - `RequireSignIn({ title: string; children })`
  - `RequireEditor({ permission?: 'canEdit' | 'canManageGallery' | 'canManageUsers' })`
  - the query keys `queryKeys.players = ['players']` and `queryKeys.users = ['users']`
  - the fixtures `userAdmin` (id 1, "Una Admin", `canManageUsers` only), `players: Player[]` and `adminUsers: AdminUser[]`. Their types come from Task 2's modules, so Task 1 declares local interfaces first; see Step 3.

- [ ] **Step 1: Write the failing tests**

`client/components/edit/gates.test.tsx`:

```tsx
import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ROLES } from '../../lib/roles';
import { anonymous, editor, publicRoutes, userAdmin } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import { Thumb } from '../page/Thumb';
import { RequireEditor } from './RequireEditor';
import { RequireSignIn } from './RequireSignIn';

describe('RequireSignIn', () => {
  it('asks signed-out visitors to sign in', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': anonymous }));
    renderWithProviders(
      <RequireSignIn title="Sign in to see the players">
        <p>secret list</p>
      </RequireSignIn>,
    );
    expect(await screen.findByText('Sign in to see the players')).toBeInTheDocument();
    expect(screen.queryByText('secret list')).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Sign in' }));
    expect(screen.getByRole('dialog', { name: 'Sign in' })).toBeInTheDocument();
  });

  it('shows the children to anyone signed in', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': userAdmin }));
    renderWithProviders(
      <RequireSignIn title="Sign in to see the players">
        <p>secret list</p>
      </RequireSignIn>,
    );
    expect(await screen.findByText('secret list')).toBeInTheDocument();
  });
});

describe('RequireEditor canManageUsers', () => {
  it('lets user admins in and keeps editors out', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': editor }));
    renderWithProviders(
      <RequireEditor permission="canManageUsers">
        <p>users</p>
      </RequireEditor>,
    );
    expect(await screen.findByText("You don't have permission to edit this")).toBeInTheDocument();
  });

  it('renders for a user admin', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': userAdmin }));
    renderWithProviders(
      <RequireEditor permission="canManageUsers">
        <p>users</p>
      </RequireEditor>,
    );
    expect(await screen.findByText('users')).toBeInTheDocument();
  });
});

describe('Thumb and roles', () => {
  it('falls back to the crest without a photo', () => {
    mockFetch(publicRoutes());
    const { container } = renderWithProviders(<Thumb />);
    expect(container.querySelector('img')?.getAttribute('src')).toMatch(/crest/);
  });

  it('lists the nine roles in order', () => {
    expect(ROLES.map((r) => r.code)).toEqual([
      'photographer',
      'manager',
      'programme_editor',
      'league_secretary',
      'treasurer',
      'safeguarding_officer',
      'club_secretary',
      'chairperson',
      'webmaster',
    ]);
    expect(ROLES[5].label).toBe('Safeguarding Officer');
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/components/edit/gates.test.tsx`
Expected: FAIL. `../../lib/roles` and `./RequireSignIn` can't be resolved, and `userAdmin` isn't exported.

- [ ] **Step 3: Implement**

`client/lib/roles.ts`:

```ts
/** The server's role codes (role.GetRole) with their display names, in the legacy order. */
export const ROLES: { code: string; label: string }[] = [
  { code: 'photographer', label: 'Photographer' },
  { code: 'manager', label: 'Manager' },
  { code: 'programme_editor', label: 'Programme Editor' },
  { code: 'league_secretary', label: 'League Secretary' },
  { code: 'treasurer', label: 'Treasurer' },
  { code: 'safeguarding_officer', label: 'Safeguarding Officer' },
  { code: 'club_secretary', label: 'Club Secretary' },
  { code: 'chairperson', label: 'Chairperson' },
  { code: 'webmaster', label: 'Webmaster' },
];

/** Only managers have a team. */
export const MANAGER_ROLE = 'manager';
```

`client/components/page/Thumb.tsx`:

```tsx
import crest from '../../assets/crest.png';
import { ImageWithFallback } from './ImageWithFallback';

/** A small round photo for table rows; the crest when there's none (or it's hidden). */
export function Thumb({ src }: { src?: string }) {
  return (
    <ImageWithFallback
      src={src}
      alt=""
      fallbackSrc={crest}
      loading="lazy"
      className="size-10 rounded-full border border-line bg-white object-cover"
    />
  );
}
```

`client/components/edit/RequireSignIn.tsx`:

```tsx
import type { ReactNode } from 'react';

import { useAuth } from '../../auth/useAuth';
import { useSignIn } from '../layout/useSignIn';
import { PageSkeleton } from '../page/QueryState';
import { Button } from '../ui/Button';
import { EmptyState } from '../ui/EmptyState';

/** Gate for member-only pages: loading placeholder, a sign-in prompt, or the page. */
export function RequireSignIn({ title, children }: { title: string; children: ReactNode }) {
  const { user, isLoading } = useAuth();
  const signIn = useSignIn();
  if (isLoading) {
    return <PageSkeleton />;
  }
  if (!user) {
    return (
      <EmptyState
        title={title}
        message="Your session may have expired."
        action={<Button onClick={signIn.open}>Sign in</Button>}
      />
    );
  }
  return <>{children}</>;
}
```

`client/components/edit/RequireEditor.tsx`: change the prop type to `permission?: 'canEdit' | 'canManageGallery' | 'canManageUsers';`.

`client/pages/account/AccountPage.tsx`: replace the body with the version below and drop the now-unused imports (`useSignIn`, `EmptyState`, `Button`, `PageSkeleton`) if lint reports them.

```tsx
export default function AccountPage() {
  usePageTitle('Account');
  return (
    <>
      <PageHeader title="Your account" />
      <RequireSignIn title="Sign in to see your account">
        <AccountContent />
      </RequireSignIn>
    </>
  );
}

function AccountContent() {
  const { user } = useAuth();
  if (!user) {
    return null;
  }
  return (
    <div className="grid gap-6 md:grid-cols-2">
      <PhotoCard user={user} />
      <DetailsCard user={user} />
      <div className="md:col-span-2">
        <PasswordCard />
      </div>
    </div>
  );
}
```

Add `import { RequireSignIn } from '../../components/edit/RequireSignIn';`.

`client/api/queries.ts`: add `players: ['players'] as const,` and `users: ['users'] as const,` to `queryKeys`.

`client/test/fixtures.ts`:

- Give `signedIn` a fifth parameter and pass it through:

```ts
function signedIn(
  name: string,
  role: string,
  canEdit: boolean,
  canManageGallery: boolean,
  canManageUsers = false,
) {
  return {
    body: {
      id: 1,
      name,
      email: 'someone@example.test',
      role,
      permissions: { canEdit, canManageGallery, canManageUsers },
    },
  };
}
```

- Add after `photographer`:

```ts
export const userAdmin: MockResponse = signedIn('Una Admin', 'Club Secretary', false, true, true);
```

- Before `publicRoutes`, add the fixtures below. Task 2 replaces the local interfaces with `import type { Player } from '../api/players'` and `import type { AdminUser } from '../api/users'`.

```ts
interface PlayerFixture {
  id: number;
  name: string;
  position?: string;
  isCaptain: boolean;
  dateOfBirth?: string;
  age?: number;
  team?: { id: number; name: string; isYouth: boolean };
  imageUrl?: string;
}

export const players: PlayerFixture[] = [
  {
    id: 30,
    name: 'Sam Striker',
    position: 'Forward',
    isCaptain: true,
    dateOfBirth: '1998-04-02T00:00:00Z',
    age: 28,
    team: { id: team.id, name: team.name, isYouth: false },
    imageUrl: '/api/v1/files/player/30',
  },
  {
    id: 31,
    name: 'Yan Youth',
    isCaptain: false,
    dateOfBirth: '2014-09-05T00:00:00Z',
    age: 12,
    team: { id: youthTeam.id, name: youthTeam.name, isYouth: true },
  },
];

interface AdminUserFixture {
  id: number;
  name: string;
  email: string;
  phone?: string;
  role: string;
  roleCode: string;
  teamId?: number;
  imageUrl?: string;
}

// id 1 is also the signed-in fixture users' id, so it's "you".
export const adminUsers: AdminUserFixture[] = [
  { id: 1, name: 'Una Admin', email: 'una@example.test', role: 'Club Secretary', roleCode: 'club_secretary' },
  {
    id: 2,
    name: 'Mo Manager',
    email: 'mo@example.test',
    phone: '07700 900000',
    role: 'Manager',
    roleCode: 'manager',
    teamId: team.id,
  },
];
```

- In `publicRoutes`, add `'/api/v1/players': { body: players },` and `'/api/v1/users': { body: adminUsers },`.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/components client/lib client/pages/account client/api client/test && yarn vitest run client/components/edit/gates.test.tsx client/pages/account && yarn test:client`
Expected: PASS. 6 new tests pass, and the account tests, including "Sign in to see your account", are still green.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add -A client
git commit -q -m "Add admin foundations: roles, thumbnails, sign-in gate, user-admin permission" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 2: Players and users API calls

**Files:**
- Create: `client/api/players.ts`, `client/api/users.ts`
- Modify: `client/api/pages.ts` (adds `setDisplayEmail`), `client/test/fixtures.ts` (uses the real types)
- Test: `client/api/admin.test.ts`

**Interfaces:**
- Consumes: `formData`, `ImageValue`, `emptyImage` (4c-1); `apiFetch`; `queryKeys`.
- Produces:
  - **Players:**
    - `Player`, `usePlayers()`
    - `PlayerInput { name: string; teamId: number; dateOfBirth: string; position: string; isCaptain: boolean; image: ImageValue }`
    - `createPlayer(i): Promise<Player>`, `updatePlayer(id, i): Promise<Player>`, `deletePlayer(id): Promise<void>`
  - **Users:**
    - `AdminUser`, `useUsers()`
    - `UserInput { name: string; email: string; phone: string; role?: string; teamId?: number; image: ImageValue }`
    - `CreatedUser { user: AdminUser; emailSent: boolean; tempPassword?: string }`, `ResetResult { emailSent: boolean; resetUrl?: string }`
    - `createUser(i): Promise<CreatedUser>`, `updateUser(id, i): Promise<AdminUser>`, `deleteUser(id): Promise<void>`, `resetUserPassword(id): Promise<ResetResult>`
  - **Contact email:** `setDisplayEmail(email: string): Promise<{ email: string }>`

- [ ] **Step 1: Write the failing tests**

`client/api/admin.test.ts`:

```ts
import { describe, expect, it } from 'vitest';

import { emptyImage } from '../lib/images';
import { mockFetch } from '../test/mockFetch';
import { setDisplayEmail } from './pages';
import { createPlayer, deletePlayer, updatePlayer } from './players';
import { createUser, deleteUser, resetUserPassword, updateUser } from './users';

function body(fetchMock: ReturnType<typeof mockFetch>, i = 0) {
  return fetchMock.mock.calls[i][1]?.body as FormData;
}

const playerInput = {
  name: 'Sam',
  teamId: 7,
  dateOfBirth: '1998-04-02',
  position: '',
  isCaptain: false,
  image: emptyImage,
};

describe('player writes', () => {
  it('creates and updates with every field, isCaptain explicit, no image fields when untouched', async () => {
    const fetchMock = mockFetch({
      '/api/v1/players': { status: 201, body: { id: 30 } },
      '/api/v1/players/30': { body: { id: 30 } },
    });
    await createPlayer(playerInput);
    await updatePlayer(30, playerInput);
    for (const i of [0, 1]) {
      const fd = body(fetchMock, i);
      expect([...fd.keys()].sort()).toEqual(['dateOfBirth', 'isCaptain', 'name', 'position', 'teamId']);
      expect(fd.get('isCaptain')).toBe('false');
      expect(fd.get('teamId')).toBe('7');
    }
    expect(fetchMock.mock.calls[1][1]?.method).toBe('PATCH');
  });

  it('sends removeImage only on update', async () => {
    const fetchMock = mockFetch({ '/api/v1/players/30': { body: { id: 30 } } });
    await updatePlayer(30, { ...playerInput, image: { file: null, remove: true } });
    expect(body(fetchMock).get('removeImage')).toBe('true');
  });

  it('deletes', async () => {
    const fetchMock = mockFetch({ '/api/v1/players/30': { status: 204 } });
    await deletePlayer(30);
    expect(fetchMock.mock.calls[0][1]?.method).toBe('DELETE');
  });
});

describe('user writes', () => {
  it('sends teamId only when given and role only when given', async () => {
    const fetchMock = mockFetch({
      '/api/v1/users': { status: 201, body: { user: { id: 3 }, emailSent: true } },
      '/api/v1/users/3': { body: { id: 3 } },
    });
    await createUser({ name: 'N', email: 'n@x.test', phone: '', role: 'manager', teamId: 7, image: emptyImage });
    expect(body(fetchMock).get('role')).toBe('manager');
    expect(body(fetchMock).get('teamId')).toBe('7');
    await updateUser(3, { name: 'N', email: 'n@x.test', phone: '', image: emptyImage });
    const fd = body(fetchMock, 1);
    expect([...fd.keys()].sort()).toEqual(['email', 'name', 'phone']);
    expect(fd.get('phone')).toBe('');
    expect(fetchMock.mock.calls[1][1]?.method).toBe('PATCH');
  });

  it('resets with POST and returns the result; deletes with DELETE', async () => {
    const fetchMock = mockFetch({
      '/api/v1/users/3/reset': { body: { emailSent: false, resetUrl: 'https://x/reset/t' } },
      '/api/v1/users/3': { status: 204 },
    });
    await expect(resetUserPassword(3)).resolves.toEqual({ emailSent: false, resetUrl: 'https://x/reset/t' });
    expect(fetchMock.mock.calls[0][1]?.method).toBe('POST');
    await deleteUser(3);
    expect(fetchMock.mock.calls[1][1]?.method).toBe('DELETE');
  });
});

describe('display email', () => {
  it('PUTs JSON, including an empty email to clear it', async () => {
    const fetchMock = mockFetch({ '/api/v1/settings/display-email': { body: { email: '' } } });
    await setDisplayEmail('');
    expect(fetchMock.mock.calls[0][1]?.method).toBe('PUT');
    expect(JSON.parse(String(fetchMock.mock.calls[0][1]?.body))).toEqual({ email: '' });
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/api/admin.test.ts`
Expected: FAIL. `./players` and `./users` can't be resolved.

- [ ] **Step 3: Implement**

`client/api/players.ts`:

```ts
import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';
import { formData } from '../lib/editForm';
import type { ImageValue } from '../lib/images';

/** player.Public — GET /players (signed-in users). imageUrl is absent when the photo is hidden. */
export interface Player {
  id: number;
  name: string;
  position?: string;
  isCaptain: boolean;
  dateOfBirth?: string;
  age?: number;
  team?: { id: number; name: string; isYouth: boolean };
  imageUrl?: string;
}

export function usePlayers() {
  return useQuery({
    queryKey: queryKeys.players,
    queryFn: ({ signal }) => apiFetch<Player[]>('/players', { signal }),
  });
}

export interface PlayerInput {
  name: string;
  teamId: number;
  /** YYYY-MM-DD */
  dateOfBirth: string;
  position: string;
  isCaptain: boolean;
  image: ImageValue;
}

function playerForm(input: PlayerInput, isUpdate: boolean): FormData {
  return formData({
    name: input.name,
    teamId: input.teamId,
    dateOfBirth: input.dateOfBirth,
    position: input.position,
    isCaptain: input.isCaptain,
    image: input.image.file,
    removeImage: isUpdate && input.image.remove ? true : undefined,
  });
}

export function createPlayer(input: PlayerInput): Promise<Player> {
  return apiFetch<Player>('/players', { form: playerForm(input, false) });
}

export function updatePlayer(id: number, input: PlayerInput): Promise<Player> {
  return apiFetch<Player>(`/players/${id}`, { method: 'PATCH', form: playerForm(input, true) });
}

export function deletePlayer(id: number): Promise<void> {
  return apiFetch<void>(`/players/${id}`, { method: 'DELETE' });
}
```

`client/api/users.ts`:

```ts
import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';
import { formData } from '../lib/editForm';
import type { ImageValue } from '../lib/images';

/** user.Admin — GET /users (user admins). role is the display name, roleCode the input code. */
export interface AdminUser {
  id: number;
  name: string;
  email: string;
  phone?: string;
  role: string;
  roleCode: string;
  teamId?: number;
  imageUrl?: string;
}

export function useUsers() {
  return useQuery({
    queryKey: queryKeys.users,
    queryFn: ({ signal }) => apiFetch<AdminUser[]>('/users', { signal }),
  });
}

export interface UserInput {
  name: string;
  email: string;
  /** Always sent; empty clears it. */
  phone: string;
  /** Role code; left out when editing your own account. */
  role?: string;
  /** Managers only. */
  teamId?: number;
  image: ImageValue;
}

/** user.Created — tempPassword is set only when the welcome email couldn't be sent. */
export interface CreatedUser {
  user: AdminUser;
  emailSent: boolean;
  tempPassword?: string;
}

/** user.ResetResult — resetUrl is set only when the reset email couldn't be sent. */
export interface ResetResult {
  emailSent: boolean;
  resetUrl?: string;
}

function userForm(input: UserInput, isUpdate: boolean): FormData {
  return formData({
    name: input.name,
    email: input.email,
    phone: input.phone,
    role: input.role,
    teamId: input.teamId,
    image: input.image.file,
    removeImage: isUpdate && input.image.remove ? true : undefined,
  });
}

export function createUser(input: UserInput): Promise<CreatedUser> {
  return apiFetch<CreatedUser>('/users', { form: userForm(input, false) });
}

export function updateUser(id: number, input: UserInput): Promise<AdminUser> {
  return apiFetch<AdminUser>(`/users/${id}`, { method: 'PATCH', form: userForm(input, true) });
}

export function deleteUser(id: number): Promise<void> {
  return apiFetch<void>(`/users/${id}`, { method: 'DELETE' });
}

export function resetUserPassword(id: number): Promise<ResetResult> {
  return apiFetch<ResetResult>(`/users/${id}/reset`, { method: 'POST' });
}
```

`client/api/pages.ts`: add

```ts
/** Sets (or, with an empty string, removes) the public contact email. */
export function setDisplayEmail(email: string): Promise<{ email: string }> {
  return apiFetch<{ email: string }>('/settings/display-email', { method: 'PUT', json: { email } });
}
```

`client/test/fixtures.ts`:
- delete the `PlayerFixture` and `AdminUserFixture` interfaces;
- add `import type { Player } from '../api/players';` and `import type { AdminUser } from '../api/users';`;
- type the arrays as `Player[]` and `AdminUser[]`.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/api client/test && yarn vitest run client/api/admin.test.ts && yarn test:client`
Expected: PASS. 6 new tests pass, and the suite is green.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add client/api client/test
git commit -q -m "Add players, users and contact-email API calls" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 3: Players page

**Files:**
- Create: `client/pages/players/PlayersPage.tsx`, `client/pages/players/PlayerDialog.tsx`
- Modify: `client/App.tsx` (route), `client/components/layout/AccountControl.tsx` (the Players link becomes `to`), `client/components/layout/AccountControl.test.tsx`
- Test: `client/pages/players/players.test.tsx`

**Interfaces:**
- Consumes:
  - `usePlayers`, `createPlayer`, `updatePlayer`, `deletePlayer` (Task 2);
  - `RequireSignIn`, `Thumb` (Task 1);
  - `useTeams`, `useCanEdit`, `useSaveForm`, `ImageField`, `DeleteButton`, `toDateInput`, `formatDate`, `matchesQuery`, `SearchInput`, `useSearchQuery`.
- Produces: the route `players`, and `PlayerDialog({ open, player?, onClose })`.

- [ ] **Step 1: Write the failing tests**

`client/pages/players/players.test.tsx`:

```tsx
import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import { anonymous, editor, manager, players, publicRoutes, youthTeam } from '../../test/fixtures';
import { Location } from '../../test/Location';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import PlayersPage from './PlayersPage';

function renderPlayers(overrides: Record<string, MockRoute> = {}, route = '/players') {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(
    <>
      <Routes>
        <Route path="/players" element={<PlayersPage />} />
      </Routes>
      <Location />
    </>,
    { route },
  );
  return fetchMock;
}

const [sam, yan] = players;

describe('PlayersPage', () => {
  it('asks signed-out visitors to sign in', async () => {
    renderPlayers({ '/api/v1/auth/me': anonymous });
    expect(await screen.findByText('Sign in to see the players')).toBeInTheDocument();
  });

  it('lists players with captain badge, team and age; a Manager gets no controls', async () => {
    renderPlayers({ '/api/v1/auth/me': manager });
    const row = (await screen.findByText(sam.name)).closest('tr') as HTMLElement;
    expect(within(row).getByText('Captain')).toBeInTheDocument();
    expect(within(row).getByText('First Team')).toBeInTheDocument();
    expect(within(row).getByText('2 Apr 1998 (age 28)')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Add player' })).toBeNull();
    expect(screen.queryByRole('button', { name: `Edit ${sam.name}` })).toBeNull();
  });

  it('filters by search and team, kept in the address', async () => {
    renderPlayers();
    await screen.findByText(sam.name);
    fireEvent.change(screen.getByLabelText('Team'), { target: { value: String(youthTeam.id) } });
    expect(screen.queryByText(sam.name)).toBeNull();
    expect(screen.getByText(yan.name)).toBeInTheDocument();
    expect(screen.getByTestId('location')).toHaveTextContent(`team=${youthTeam.id}`);
    fireEvent.change(screen.getByLabelText('Search players'), { target: { value: 'zzz' } });
    expect(screen.getByText('No players match your filters')).toBeInTheDocument();
  });

  it('adds a player, validating first', async () => {
    const fetchMock = renderPlayers({
      '/api/v1/players': (init) =>
        init?.method === 'POST' ? { status: 201, body: sam } : { body: players },
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Add player' }));
    const dialog = screen.getByRole('dialog', { name: 'Add player' });
    expect(within(dialog).getByText('Photos are never shown for youth-team or under-18 players.')).toBeInTheDocument();
    fireEvent.click(within(dialog).getByRole('button', { name: 'Save player' }));
    expect(within(dialog).getByLabelText('Name')).toHaveAccessibleDescription('Enter a name');
    expect(within(dialog).getByLabelText('Team')).toHaveAccessibleDescription('Choose a team');
    expect(within(dialog).getByLabelText('Date of birth')).toHaveAccessibleDescription('Enter the date of birth');
    fireEvent.change(within(dialog).getByLabelText('Name'), { target: { value: 'New Kid' } });
    fireEvent.change(within(dialog).getByLabelText('Team'), { target: { value: String(youthTeam.id) } });
    fireEvent.change(within(dialog).getByLabelText('Date of birth'), { target: { value: '2015-01-20' } });
    fireEvent.click(within(dialog).getByRole('checkbox', { name: 'Captain' }));
    fireEvent.click(within(dialog).getByRole('button', { name: 'Save player' }));
    expect(await screen.findByRole('button', { name: 'Player saved' })).toBeInTheDocument();
    const fd = fetchMock.mock.calls.find(([, i]) => i?.method === 'POST')?.[1]?.body as FormData;
    expect(fd.get('name')).toBe('New Kid');
    expect(fd.get('teamId')).toBe(String(youthTeam.id));
    expect(fd.get('dateOfBirth')).toBe('2015-01-20');
    expect(fd.get('isCaptain')).toBe('true');
  });

  it('edits a player, sending an unticked Captain as false', async () => {
    const fetchMock = renderPlayers({ [`/api/v1/players/${sam.id}`]: { body: sam } });
    fireEvent.click(await screen.findByRole('button', { name: `Edit ${sam.name}` }));
    const dialog = screen.getByRole('dialog', { name: `Edit ${sam.name}` });
    expect(within(dialog).getByLabelText('Date of birth')).toHaveValue('1998-04-02');
    fireEvent.click(within(dialog).getByRole('checkbox', { name: 'Captain' }));
    fireEvent.click(within(dialog).getByRole('button', { name: 'Save player' }));
    await waitFor(() => expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'PATCH')).toBe(true));
    const fd = fetchMock.mock.calls.find(([, i]) => i?.method === 'PATCH')?.[1]?.body as FormData;
    expect(fd.get('isCaptain')).toBe('false');
    expect(fd.has('image')).toBe(false);
    expect(fd.has('removeImage')).toBe(false);
  });

  it('opens Add empty straight after closing Edit', async () => {
    renderPlayers();
    fireEvent.click(await screen.findByRole('button', { name: `Edit ${sam.name}` }));
    fireEvent.click(within(screen.getByRole('dialog', { name: `Edit ${sam.name}` })).getByRole('button', { name: 'Cancel' }));
    fireEvent.click(screen.getByRole('button', { name: 'Add player' }));
    expect(within(screen.getByRole('dialog', { name: 'Add player' })).getByLabelText('Name')).toHaveValue('');
  });

  it('deletes a player after confirming', async () => {
    const fetchMock = renderPlayers({ [`/api/v1/players/${sam.id}`]: { status: 204 } });
    fireEvent.click(await screen.findByRole('button', { name: `Delete ${sam.name}` }));
    fireEvent.click(
      within(screen.getByRole('dialog', { name: `Delete ${sam.name}?` })).getByRole('button', { name: 'Delete' }),
    );
    expect(await screen.findByRole('button', { name: 'Player deleted' })).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true);
  });
});
```

In `client/components/layout/AccountControl.test.tsx`, add inside `describe('AccountControl signed in', …)`:

```tsx
  it('opens Players and Users in the app', async () => {
    mockFetch({
      '/api/v1/auth/me': {
        body: user('Sam Sec', 'Club Secretary', { canEdit: true, canManageUsers: true }),
      },
    });
    renderWithProviders(
      <>
        <AccountControl />
        <Location />
      </>,
    );
    fireEvent.click(await screen.findByRole('button', { name: /Sam Sec/ }));
    fireEvent.click(screen.getByRole('menuitem', { name: 'Players' }));
    expect(screen.getByTestId('location')).toHaveTextContent('/players');
    fireEvent.click(screen.getByRole('button', { name: /Sam Sec/ }));
    fireEvent.click(screen.getByRole('menuitem', { name: 'Users' }));
    expect(screen.getByTestId('location')).toHaveTextContent('/users');
  });
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/players client/components/layout`
Expected: FAIL. `./PlayersPage` can't be resolved, and the menu test stays on `/`.

- [ ] **Step 3: Implement**

`client/pages/players/PlayerDialog.tsx`:

```tsx
import { useState, type FormEvent } from 'react';

import { createPlayer, updatePlayer, type Player } from '../../api/players';
import { queryKeys } from '../../api/queries';
import { useTeams } from '../../api/teams';
import { ImageField } from '../../components/edit/ImageField';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Checkbox } from '../../components/ui/Checkbox';
import { Input, Select } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';
import { toDateInput } from '../../lib/editForm';
import { emptyImage, type ImageValue } from '../../lib/images';

type Missing = { name?: string; team?: string; dob?: string };

function PlayerForm({ player, onClose }: { player?: Player; onClose: () => void }) {
  const toast = useToast();
  const teams = useTeams();
  const [name, setName] = useState(player?.name ?? '');
  const [teamId, setTeamId] = useState(player?.team ? String(player.team.id) : '');
  const [dob, setDob] = useState(player?.dateOfBirth ? toDateInput(player.dateOfBirth) : '');
  const [position, setPosition] = useState(player?.position ?? '');
  const [isCaptain, setIsCaptain] = useState(player?.isCaptain ?? false);
  const [image, setImage] = useState<ImageValue>(emptyImage);
  const [missing, setMissing] = useState<Missing>({});

  const save = useSaveForm({
    submit: () => {
      const input = { name, teamId: Number(teamId), dateOfBirth: dob, position, isCaptain, image };
      return player ? updatePlayer(player.id, input) : createPlayer(input);
    },
    invalidate: [queryKeys.players, ['team']],
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Player saved' });
      onClose();
    },
  });

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const next: Missing = {
      name: name.trim() ? undefined : 'Enter a name',
      team: teamId ? undefined : 'Choose a team',
      dob: dob ? undefined : 'Enter the date of birth',
    };
    setMissing(next);
    if (!next.name && !next.team && !next.dob) {
      void save.run();
    }
  }

  const err = save.fieldErrors;
  return (
    <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
      {save.formError && <Alert tone="error">{save.formError}</Alert>}
      <Field label="Name" error={missing.name ?? err.name}>
        <Input value={name} onChange={(e) => setName(e.target.value)} />
      </Field>
      <Field label="Team" error={missing.team ?? err.teamId}>
        <Select value={teamId} onChange={(e) => setTeamId(e.target.value)}>
          <option value="">Choose…</option>
          {teams.data?.map((t) => (
            <option key={t.id} value={String(t.id)}>
              {t.name}
            </option>
          ))}
        </Select>
      </Field>
      <Field label="Date of birth" error={missing.dob ?? err.dateOfBirth}>
        <Input type="date" value={dob} onChange={(e) => setDob(e.target.value)} />
      </Field>
      <Field label="Position" error={err.position}>
        <Input value={position} onChange={(e) => setPosition(e.target.value)} />
      </Field>
      <Checkbox label="Captain" checked={isCaptain} onChange={(e) => setIsCaptain(e.target.checked)} />
      <ImageField
        label="Photo"
        currentUrl={player?.imageUrl}
        allowRemove={Boolean(player)}
        value={image}
        onChange={setImage}
        error={err.file ?? err.image}
      />
      <p className="text-sm text-muted">Photos are never shown for youth-team or under-18 players.</p>
      <div className="flex justify-end gap-2">
        <Button variant="secondary" onClick={onClose} disabled={save.busy}>
          Cancel
        </Button>
        <Button type="submit" loading={save.busy}>
          Save player
        </Button>
      </div>
    </form>
  );
}

/** Add or edit a player; the form mounts only while open, so each open starts fresh. */
export function PlayerDialog({
  open,
  player,
  onClose,
}: {
  open: boolean;
  player?: Player;
  onClose: () => void;
}) {
  return (
    <Modal open={open} onClose={onClose} title={player ? `Edit ${player.name}` : 'Add player'}>
      <PlayerForm player={player} onClose={onClose} />
    </Modal>
  );
}
```

`client/pages/players/PlayersPage.tsx`:

```tsx
import { useState } from 'react';
import { useSearchParams } from 'react-router';

import { deletePlayer, usePlayers, type Player } from '../../api/players';
import { queryKeys } from '../../api/queries';
import { useTeams } from '../../api/teams';
import { DeleteButton } from '../../components/edit/DeleteButton';
import { RequireSignIn } from '../../components/edit/RequireSignIn';
import { useCanEdit } from '../../components/edit/useCanEdit';
import { QueryState } from '../../components/page/QueryState';
import { SearchInput } from '../../components/page/SearchInput';
import { Thumb } from '../../components/page/Thumb';
import { usePageTitle } from '../../components/page/usePageTitle';
import { useSearchQuery } from '../../components/page/useSearchQuery';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { Select } from '../../components/ui/controls';
import { EmptyState } from '../../components/ui/EmptyState';
import { Field } from '../../components/ui/Field';
import { PageHeader } from '../../components/ui/PageHeader';
import { Table, TBody, Td, Th, THead, Tr } from '../../components/ui/Table';
import { formatDate } from '../../lib/format';
import { matchesQuery } from '../../lib/text';
import { PlayerDialog } from './PlayerDialog';

function byTeamThenName(a: Player, b: Player): number {
  return (
    (a.team?.name ?? '').localeCompare(b.team?.name ?? '') || a.name.localeCompare(b.name)
  );
}

function born(p: Player): string {
  if (!p.dateOfBirth) {
    return '—';
  }
  const age = p.age === undefined ? '' : ` (age ${p.age})`;
  return `${formatDate(p.dateOfBirth)}${age}`;
}

function PlayersTable({ canEdit, onEdit }: { canEdit: boolean; onEdit: (p: Player) => void }) {
  const players = usePlayers();
  const teams = useTeams();
  const [params, setParams] = useSearchParams();
  const q = useSearchQuery();
  const teamFilter = params.get('team') ?? '';

  return (
    <>
      <div className="mb-4 flex flex-wrap items-end gap-4">
        <SearchInput label="Search players" />
        <Field label="Team" className="w-56">
          <Select
            value={teamFilter}
            onChange={(e) => {
              const next = new URLSearchParams(params);
              if (e.target.value) {
                next.set('team', e.target.value);
              } else {
                next.delete('team');
              }
              setParams(next, { replace: true });
            }}
          >
            <option value="">All teams</option>
            {teams.data?.map((t) => (
              <option key={t.id} value={String(t.id)}>
                {t.name}
              </option>
            ))}
          </Select>
        </Field>
      </div>
      <QueryState query={players} isEmpty={(l) => l.length === 0} emptyTitle="No players yet">
        {(list) => {
          const shown = [...list]
            .sort(byTeamThenName)
            .filter(
              (p) =>
                matchesQuery(p.name, q) && (!teamFilter || String(p.team?.id) === teamFilter),
            );
          if (shown.length === 0) {
            return <EmptyState title="No players match your filters" />;
          }
          return (
            <Table>
              <THead>
                <Tr>
                  <Th>
                    <span className="sr-only">Photo</span>
                  </Th>
                  <Th>Name</Th>
                  <Th>Position</Th>
                  <Th>Team</Th>
                  <Th>Date of birth</Th>
                  {canEdit && (
                    <Th>
                      <span className="sr-only">Actions</span>
                    </Th>
                  )}
                </Tr>
              </THead>
              <TBody>
                {shown.map((p) => (
                  <Tr key={p.id}>
                    <Td>
                      <Thumb src={p.imageUrl} />
                    </Td>
                    <Td>
                      <span className="font-medium">{p.name}</span>
                      {p.isCaptain && (
                        <Badge tone="red" className="ml-2">
                          Captain
                        </Badge>
                      )}
                    </Td>
                    <Td>{p.position}</Td>
                    <Td>{p.team?.name ?? 'No team'}</Td>
                    <Td className="whitespace-nowrap">{born(p)}</Td>
                    {canEdit && (
                      <Td>
                        <div className="flex justify-end gap-2">
                          <Button
                            size="sm"
                            variant="secondary"
                            aria-label={`Edit ${p.name}`}
                            onClick={() => onEdit(p)}
                          >
                            Edit
                          </Button>
                          <DeleteButton
                            ariaLabel={`Delete ${p.name}`}
                            confirmTitle={`Delete ${p.name}?`}
                            confirmMessage="This can't be undone."
                            onDelete={() => deletePlayer(p.id)}
                            invalidate={[queryKeys.players, ['team']]}
                            successMessage="Player deleted"
                          />
                        </div>
                      </Td>
                    )}
                  </Tr>
                ))}
              </TBody>
            </Table>
          );
        }}
      </QueryState>
    </>
  );
}

export default function PlayersPage() {
  usePageTitle('Players');
  const { canEdit } = useCanEdit();
  // null = closed; {} = adding; { player } = editing.
  const [dialog, setDialog] = useState<{ player?: Player } | null>(null);
  return (
    <>
      <PageHeader
        title="Players"
        actions={canEdit && <Button onClick={() => setDialog({})}>Add player</Button>}
      />
      <RequireSignIn title="Sign in to see the players">
        <PlayersTable canEdit={canEdit} onEdit={(player) => setDialog({ player })} />
      </RequireSignIn>
      <PlayerDialog
        open={dialog !== null}
        player={dialog?.player}
        onClose={() => setDialog(null)}
      />
    </>
  );
}
```

`client/App.tsx`: add `const PlayersPage = lazy(() => import('./pages/players/PlayersPage'));` and `<Route path="players" element={<PlayersPage />} />`.

`client/components/layout/AccountControl.tsx`: change `{ label: 'Players', href: '/players' }` to `{ label: 'Players', to: '/players' }`, and `{ label: 'Users', href: '/users' }` to `{ label: 'Users', to: '/users' }`. The `/users` route arrives in Task 4; until then the link lands on the not-found page, which is fine mid-branch.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/pages/players client/App.tsx client/components/layout && yarn vitest run client/pages/players client/components/layout && yarn test:client`
Expected: PASS. 8 new tests pass, and the suite is green.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add -A client
git commit -q -m "Add the Players admin page" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 4: Users page: table, add and edit, own row, temporary password

**Files:**
- Create: `client/pages/users/UsersPage.tsx`, `client/pages/users/UserDialog.tsx`, `client/pages/users/SecretDialog.tsx`
- Modify: `client/App.tsx` (route)
- Test: `client/pages/users/users.test.tsx`

**Interfaces:**
- Consumes: `useUsers`, `createUser`, `updateUser`, `deleteUser`, `CreatedUser` (Task 2); `ROLES`, `MANAGER_ROLE`, `Thumb`, `RequireEditor` (Task 1); `useTeams`, `useAuth`, the toolkit.
- Produces:
  - `Secret { title: string; message: string; label: string; value: string }`
  - `SecretDialog({ secret: Secret | null; onClose })`
  - `UserDialog({ open, user?, isSelf, onClose, onSecret })`
  - `UsersPage`, which lays out the header, the display-email slot (filled in Task 5) and `UsersTable`, with a `ResetPasswordButton` slot (Task 5).

- [ ] **Step 1: Write the failing tests**

`client/pages/users/users.test.tsx`:

```tsx
import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it, vi } from 'vitest';

import { adminUsers, editor, publicRoutes, team, userAdmin } from '../../test/fixtures';
import { Location } from '../../test/Location';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import UsersPage from './UsersPage';

function renderUsers(overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': userAdmin, ...overrides }));
  renderWithProviders(
    <>
      <Routes>
        <Route path="/users" element={<UsersPage />} />
      </Routes>
      <Location />
    </>,
    { route: '/users' },
  );
  return fetchMock;
}

const [una, mo] = adminUsers;
const sent = (fetchMock: ReturnType<typeof mockFetch>, method: string) =>
  fetchMock.mock.calls.find(([, i]) => i?.method === method)?.[1]?.body as FormData;

describe('UsersPage access and table', () => {
  it('keeps out an editor who is not a user admin', async () => {
    renderUsers({ '/api/v1/auth/me': editor });
    expect(await screen.findByText("You don't have permission to edit this")).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Add user' })).toBeNull();
  });

  it('lists users; your own row says You and has no Delete', async () => {
    renderUsers();
    const own = (await screen.findByText(una.name)).closest('tr') as HTMLElement;
    expect(within(own).getByText('You')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: `Delete ${una.name}` })).toBeNull();
    const other = screen.getByText(mo.name).closest('tr') as HTMLElement;
    expect(within(other).getByRole('link', { name: mo.email })).toHaveAttribute('href', `mailto:${mo.email}`);
    expect(within(other).getByText(team.name)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: `Delete ${mo.name}` })).toBeInTheDocument();
  });

  it('filters by role and search, kept in the address', async () => {
    renderUsers();
    await screen.findByText(mo.name);
    fireEvent.change(screen.getByLabelText('Role'), { target: { value: 'manager' } });
    expect(screen.queryByText(una.name)).toBeNull();
    expect(screen.getByTestId('location')).toHaveTextContent('role=manager');
    fireEvent.change(screen.getByLabelText('Search users'), { target: { value: 'zzz' } });
    expect(screen.getByText('No users match your filters')).toBeInTheDocument();
  });
});

describe('UserDialog', () => {
  it('shows Team only for Manager and sends teamId only then', async () => {
    const fetchMock = renderUsers({
      '/api/v1/users': (init) =>
        init?.method === 'POST'
          ? { status: 201, body: { user: { ...mo, id: 9, name: 'New' }, emailSent: true } }
          : { body: adminUsers },
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Add user' }));
    const dialog = screen.getByRole('dialog', { name: 'Add user' });
    expect(within(dialog).queryByLabelText('Team')).toBeNull();
    fireEvent.change(within(dialog).getByLabelText('Role'), { target: { value: 'manager' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add user' }));
    expect(within(dialog).getByLabelText('Name')).toHaveAccessibleDescription('Enter a name');
    expect(within(dialog).getByLabelText('Email')).toHaveAccessibleDescription('Enter an email address');
    expect(within(dialog).getByLabelText('Team')).toHaveAccessibleDescription("Choose the manager's team");
    fireEvent.change(within(dialog).getByLabelText('Name'), { target: { value: 'New' } });
    fireEvent.change(within(dialog).getByLabelText('Email'), { target: { value: 'new@x.test' } });
    fireEvent.change(within(dialog).getByLabelText('Team'), { target: { value: String(team.id) } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add user' }));
    expect(
      await screen.findByRole('button', { name: "User added. They've been emailed a temporary password." }),
    ).toBeInTheDocument();
    const fd = sent(fetchMock, 'POST');
    expect(fd.get('role')).toBe('manager');
    expect(fd.get('teamId')).toBe(String(team.id));
  });

  it('drops teamId when a manager is changed to another role', async () => {
    const fetchMock = renderUsers({ [`/api/v1/users/${mo.id}`]: { body: mo } });
    fireEvent.click(await screen.findByRole('button', { name: `Edit ${mo.name}` }));
    const dialog = screen.getByRole('dialog', { name: `Edit ${mo.name}` });
    fireEvent.change(within(dialog).getByLabelText('Role'), { target: { value: 'treasurer' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Save user' }));
    await screen.findByRole('button', { name: 'User saved' });
    const fd = sent(fetchMock, 'PATCH');
    expect(fd.get('role')).toBe('treasurer');
    expect(fd.has('teamId')).toBe(false);
    expect(fd.get('phone')).toBe(mo.phone);
  });

  it('locks your own role and never sends it', async () => {
    const fetchMock = renderUsers({ [`/api/v1/users/${una.id}`]: { body: una } });
    fireEvent.click(await screen.findByRole('button', { name: `Edit ${una.name}` }));
    const dialog = screen.getByRole('dialog', { name: `Edit ${una.name}` });
    const role = within(dialog).getByLabelText('Role');
    expect(role).toBeDisabled();
    expect(role).toHaveAccessibleDescription('Ask another administrator to change your role.');
    fireEvent.change(within(dialog).getByLabelText('Phone'), { target: { value: '01234' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Save user' }));
    await screen.findByRole('button', { name: 'User saved' });
    const fd = sent(fetchMock, 'PATCH');
    expect(fd.has('role')).toBe(false);
    expect(fd.get('phone')).toBe('01234');
  });

  it('puts a duplicate email under the Email field', async () => {
    renderUsers({
      [`/api/v1/users/${mo.id}`]: {
        status: 409,
        body: { error: { code: 409, message: 'email address is already in use' } },
      },
    });
    fireEvent.click(await screen.findByRole('button', { name: `Edit ${mo.name}` }));
    const dialog = screen.getByRole('dialog', { name: `Edit ${mo.name}` });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Save user' }));
    await waitFor(() =>
      expect(within(dialog).getByLabelText('Email')).toHaveAccessibleDescription(
        'That email address is already in use',
      ),
    );
  });

  it('opens Add empty straight after closing Edit', async () => {
    renderUsers();
    fireEvent.click(await screen.findByRole('button', { name: `Edit ${mo.name}` }));
    fireEvent.click(within(screen.getByRole('dialog', { name: `Edit ${mo.name}` })).getByRole('button', { name: 'Cancel' }));
    fireEvent.click(screen.getByRole('button', { name: 'Add user' }));
    const dialog = screen.getByRole('dialog', { name: 'Add user' });
    expect(within(dialog).getByLabelText('Name')).toHaveValue('');
    expect(within(dialog).getByLabelText('Role')).toHaveValue('');
  });
});

describe('temporary password', () => {
  function addWithoutEmail() {
    const fetchMock = renderUsers({
      '/api/v1/users': (init) =>
        init?.method === 'POST'
          ? {
              status: 201,
              body: { user: { ...mo, id: 9, name: 'New' }, emailSent: false, tempPassword: 'Tmp-123!' },
            }
          : { body: adminUsers },
    });
    return fetchMock;
  }

  async function submitNewUser() {
    fireEvent.click(await screen.findByRole('button', { name: 'Add user' }));
    const dialog = screen.getByRole('dialog', { name: 'Add user' });
    fireEvent.change(within(dialog).getByLabelText('Name'), { target: { value: 'New' } });
    fireEvent.change(within(dialog).getByLabelText('Email'), { target: { value: 'new@x.test' } });
    fireEvent.change(within(dialog).getByLabelText('Role'), { target: { value: 'treasurer' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add user' }));
    return screen.findByRole('dialog', { name: 'Pass this on to New' });
  }

  it('shows the temporary password and copies it', async () => {
    addWithoutEmail();
    const writeText = vi.fn(async () => undefined);
    Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true });
    const secret = await submitNewUser();
    expect(within(secret).getByLabelText('Temporary password')).toHaveValue('Tmp-123!');
    expect(secret).toHaveTextContent("It won't be shown again.");
    fireEvent.click(within(secret).getByRole('button', { name: 'Copy' }));
    expect(await screen.findByRole('button', { name: 'Copied' })).toBeInTheDocument();
    expect(writeText).toHaveBeenCalledWith('Tmp-123!');
    fireEvent.click(within(secret).getByRole('button', { name: 'Done' }));
    expect(screen.queryByRole('dialog', { name: 'Pass this on to New' })).toBeNull();
    Object.defineProperty(navigator, 'clipboard', { value: undefined, configurable: true });
  });

  it('selects the text when there is no clipboard', async () => {
    addWithoutEmail();
    Object.defineProperty(navigator, 'clipboard', { value: undefined, configurable: true });
    const secret = await submitNewUser();
    const field = within(secret).getByLabelText('Temporary password') as HTMLInputElement;
    const select = vi.spyOn(field, 'select');
    fireEvent.click(within(secret).getByRole('button', { name: 'Copy' }));
    expect(await screen.findByRole('button', { name: 'Select and copy the text above' })).toBeInTheDocument();
    expect(select).toHaveBeenCalled();
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/users`
Expected: FAIL. `./UsersPage` can't be resolved.

- [ ] **Step 3: Implement**

`client/pages/users/SecretDialog.tsx`:

```tsx
import { useRef } from 'react';

import { Button } from '../../components/ui/Button';
import { Input } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';

/** A one-time value (temporary password, reset link) for the admin to pass on. */
export interface Secret {
  title: string;
  message: string;
  label: string;
  value: string;
}

function SecretBody({ secret, onClose }: { secret: Secret; onClose: () => void }) {
  const toast = useToast();
  const inputRef = useRef<HTMLInputElement>(null);

  async function copy() {
    try {
      // navigator.clipboard is missing on plain http and may be refused.
      if (!navigator.clipboard) {
        throw new Error('no clipboard');
      }
      await navigator.clipboard.writeText(secret.value);
      toast.show({ tone: 'success', message: 'Copied' });
    } catch {
      inputRef.current?.select();
      toast.show({ tone: 'info', message: 'Select and copy the text above' });
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <p className="text-sm">{secret.message}</p>
      <Field label={secret.label}>
        <Input
          ref={inputRef}
          readOnly
          value={secret.value}
          className="font-mono"
          onFocus={(e) => e.target.select()}
        />
      </Field>
      <div className="flex justify-end gap-2">
        <Button variant="secondary" onClick={() => void copy()}>
          Copy
        </Button>
        <Button onClick={onClose}>Done</Button>
      </div>
    </div>
  );
}

export function SecretDialog({ secret, onClose }: { secret: Secret | null; onClose: () => void }) {
  return (
    <Modal open={secret !== null} onClose={onClose} title={secret?.title ?? ''}>
      {secret && <SecretBody secret={secret} onClose={onClose} />}
    </Modal>
  );
}
```

`client/pages/users/UserDialog.tsx`:

```tsx
import { useState, type FormEvent } from 'react';

import { ApiError } from '../../api/client';
import { queryKeys } from '../../api/queries';
import { useTeams } from '../../api/teams';
import { createUser, updateUser, type AdminUser, type CreatedUser } from '../../api/users';
import { ImageField } from '../../components/edit/ImageField';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Input, Select } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';
import { emptyImage, type ImageValue } from '../../lib/images';
import { MANAGER_ROLE, ROLES } from '../../lib/roles';
import type { Secret } from './SecretDialog';

type Missing = { name?: string; email?: string; role?: string; team?: string };
type Result = { saved: AdminUser } | { created: CreatedUser };

export interface UserDialogProps {
  open: boolean;
  user?: AdminUser;
  /** Editing your own account: the role is locked and not sent. */
  isSelf: boolean;
  onClose: () => void;
  onSecret: (secret: Secret) => void;
}

function UserForm({ user, isSelf, onClose, onSecret }: Omit<UserDialogProps, 'open'>) {
  const toast = useToast();
  const teams = useTeams();
  const [name, setName] = useState(user?.name ?? '');
  const [email, setEmail] = useState(user?.email ?? '');
  const [phone, setPhone] = useState(user?.phone ?? '');
  const [role, setRole] = useState(user?.roleCode ?? '');
  const [teamId, setTeamId] = useState(user?.teamId ? String(user.teamId) : '');
  const [image, setImage] = useState<ImageValue>(emptyImage);
  const [missing, setMissing] = useState<Missing>({});
  const isManager = role === MANAGER_ROLE;

  const save = useSaveForm<Result>({
    submit: async () => {
      const input = {
        name,
        email,
        phone,
        role: isSelf ? undefined : role,
        teamId: isManager ? Number(teamId) : undefined,
        image,
      };
      try {
        return user
          ? { saved: await updateUser(user.id, input) }
          : { created: await createUser(input) };
      } catch (err) {
        // The server reports a duplicate email as a 409 with no field; show it on Email.
        if (err instanceof ApiError && err.status === 409) {
          throw new ApiError(409, err.message, { email: 'That email address is already in use' });
        }
        throw err;
      }
    },
    invalidate: [queryKeys.users, queryKeys.contact, ['team'], ...(isSelf ? [queryKeys.me] : [])],
    onSaved: (result) => {
      onClose();
      if ('saved' in result) {
        toast.show({ tone: 'success', message: 'User saved' });
        return;
      }
      const { created } = result;
      if (!created.emailSent && created.tempPassword) {
        onSecret({
          title: `Pass this on to ${created.user.name}`,
          message: `We couldn't email ${created.user.name}. Give them this temporary password; they'll choose a new one when they first sign in. It won't be shown again.`,
          label: 'Temporary password',
          value: created.tempPassword,
        });
        return;
      }
      toast.show({ tone: 'success', message: "User added. They've been emailed a temporary password." });
    },
  });

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const next: Missing = {
      name: name.trim() ? undefined : 'Enter a name',
      email: email.trim() ? undefined : 'Enter an email address',
      role: role ? undefined : 'Choose a role',
      team: !isManager || teamId ? undefined : "Choose the manager's team",
    };
    setMissing(next);
    if (!next.name && !next.email && !next.role && !next.team) {
      void save.run();
    }
  }

  const err = save.fieldErrors;
  return (
    <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
      {save.formError && <Alert tone="error">{save.formError}</Alert>}
      <Field label="Name" error={missing.name ?? err.name}>
        <Input value={name} onChange={(e) => setName(e.target.value)} />
      </Field>
      <Field label="Email" error={missing.email ?? err.email}>
        <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
      </Field>
      <Field label="Phone" error={err.phone}>
        <Input type="tel" value={phone} onChange={(e) => setPhone(e.target.value)} />
      </Field>
      <Field
        label="Role"
        error={missing.role ?? err.role}
        help={isSelf ? 'Ask another administrator to change your role.' : undefined}
      >
        <Select value={role} disabled={isSelf} onChange={(e) => setRole(e.target.value)}>
          <option value="">Choose…</option>
          {ROLES.map((r) => (
            <option key={r.code} value={r.code}>
              {r.label}
            </option>
          ))}
        </Select>
      </Field>
      {isManager && (
        <Field label="Team" error={missing.team ?? err.teamId}>
          <Select value={teamId} onChange={(e) => setTeamId(e.target.value)}>
            <option value="">Choose…</option>
            {teams.data?.map((t) => (
              <option key={t.id} value={String(t.id)}>
                {t.name}
              </option>
            ))}
          </Select>
        </Field>
      )}
      <ImageField
        label="Photo"
        currentUrl={user?.imageUrl}
        allowRemove={Boolean(user)}
        value={image}
        onChange={setImage}
        error={err.file ?? err.image}
      />
      <div className="flex justify-end gap-2">
        <Button variant="secondary" onClick={onClose} disabled={save.busy}>
          Cancel
        </Button>
        <Button type="submit" loading={save.busy}>
          {user ? 'Save user' : 'Add user'}
        </Button>
      </div>
    </form>
  );
}

/** Add or edit a user; the form mounts only while open, so each open starts fresh. */
export function UserDialog({ open, user, isSelf, onClose, onSecret }: UserDialogProps) {
  return (
    <Modal open={open} onClose={onClose} title={user ? `Edit ${user.name}` : 'Add user'}>
      <UserForm user={user} isSelf={isSelf} onClose={onClose} onSecret={onSecret} />
    </Modal>
  );
}
```

`client/pages/users/UsersPage.tsx`:

```tsx
import { useState } from 'react';
import { useSearchParams } from 'react-router';

import { queryKeys } from '../../api/queries';
import { useTeams } from '../../api/teams';
import { deleteUser, useUsers, type AdminUser } from '../../api/users';
import { useAuth } from '../../auth/useAuth';
import { DeleteButton } from '../../components/edit/DeleteButton';
import { RequireEditor } from '../../components/edit/RequireEditor';
import { QueryState } from '../../components/page/QueryState';
import { SearchInput } from '../../components/page/SearchInput';
import { Thumb } from '../../components/page/Thumb';
import { usePageTitle } from '../../components/page/usePageTitle';
import { useSearchQuery } from '../../components/page/useSearchQuery';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { Select } from '../../components/ui/controls';
import { EmptyState } from '../../components/ui/EmptyState';
import { Field } from '../../components/ui/Field';
import { PageHeader } from '../../components/ui/PageHeader';
import { Table, TBody, Td, Th, THead, Tr } from '../../components/ui/Table';
import { MANAGER_ROLE, ROLES } from '../../lib/roles';
import { matchesQuery } from '../../lib/text';
import { SecretDialog, type Secret } from './SecretDialog';
import { UserDialog } from './UserDialog';

interface UsersTableProps {
  currentUserId?: number;
  onEdit: (user: AdminUser) => void;
  onSecret: (secret: Secret) => void;
}

function UsersTable({ currentUserId, onEdit }: UsersTableProps) {
  const users = useUsers();
  const teams = useTeams();
  const [params, setParams] = useSearchParams();
  const q = useSearchQuery();
  const roleFilter = params.get('role') ?? '';
  const teamName = (id?: number) => teams.data?.find((t) => t.id === id)?.name ?? '';

  return (
    <>
      <div className="mb-4 flex flex-wrap items-end gap-4">
        <SearchInput label="Search users" />
        <Field label="Role" className="w-56">
          <Select
            value={roleFilter}
            onChange={(e) => {
              const next = new URLSearchParams(params);
              if (e.target.value) {
                next.set('role', e.target.value);
              } else {
                next.delete('role');
              }
              setParams(next, { replace: true });
            }}
          >
            <option value="">All roles</option>
            {ROLES.map((r) => (
              <option key={r.code} value={r.code}>
                {r.label}
              </option>
            ))}
          </Select>
        </Field>
      </div>
      <QueryState query={users}>
        {(list) => {
          const shown = [...list]
            .sort((a, b) => a.name.localeCompare(b.name))
            .filter(
              (u) =>
                matchesQuery(`${u.name} ${u.email} ${u.role}`, q) &&
                (!roleFilter || u.roleCode === roleFilter),
            );
          if (shown.length === 0) {
            return <EmptyState title="No users match your filters" />;
          }
          return (
            <Table>
              <THead>
                <Tr>
                  <Th>
                    <span className="sr-only">Photo</span>
                  </Th>
                  <Th>Name</Th>
                  <Th>Email</Th>
                  <Th>Phone</Th>
                  <Th>Role</Th>
                  <Th>Team</Th>
                  <Th>
                    <span className="sr-only">Actions</span>
                  </Th>
                </Tr>
              </THead>
              <TBody>
                {shown.map((u) => {
                  const isSelf = u.id === currentUserId;
                  return (
                    <Tr key={u.id}>
                      <Td>
                        <Thumb src={u.imageUrl} />
                      </Td>
                      <Td>
                        <span className="font-medium">{u.name}</span>
                        {isSelf && (
                          <Badge tone="blue" className="ml-2">
                            You
                          </Badge>
                        )}
                      </Td>
                      <Td>
                        <a href={`mailto:${u.email}`} className="text-red underline">
                          {u.email}
                        </a>
                      </Td>
                      <Td className="whitespace-nowrap">{u.phone}</Td>
                      <Td>{u.role}</Td>
                      <Td>{u.roleCode === MANAGER_ROLE ? teamName(u.teamId) : ''}</Td>
                      <Td>
                        <div className="flex justify-end gap-2">
                          <Button
                            size="sm"
                            variant="secondary"
                            aria-label={`Edit ${u.name}`}
                            onClick={() => onEdit(u)}
                          >
                            Edit
                          </Button>
                          {!isSelf && (
                            <DeleteButton
                              ariaLabel={`Delete ${u.name}`}
                              confirmTitle={`Delete ${u.name}?`}
                              confirmMessage="They will no longer be able to sign in. This can't be undone."
                              onDelete={() => deleteUser(u.id)}
                              invalidate={[queryKeys.users, queryKeys.contact, ['team']]}
                              successMessage="User deleted"
                            />
                          )}
                        </div>
                      </Td>
                    </Tr>
                  );
                })}
              </TBody>
            </Table>
          );
        }}
      </QueryState>
    </>
  );
}

export default function UsersPage() {
  usePageTitle('Users');
  const { user } = useAuth();
  const canManageUsers = user?.permissions.canManageUsers ?? false;
  // null = closed; {} = adding; { user } = editing.
  const [dialog, setDialog] = useState<{ user?: AdminUser } | null>(null);
  const [secret, setSecret] = useState<Secret | null>(null);
  return (
    <>
      <PageHeader
        title="Users"
        actions={canManageUsers && <Button onClick={() => setDialog({})}>Add user</Button>}
      />
      <RequireEditor permission="canManageUsers">
        <UsersTable
          currentUserId={user?.id}
          onEdit={(u) => setDialog({ user: u })}
          onSecret={setSecret}
        />
        <UserDialog
          open={dialog !== null}
          user={dialog?.user}
          isSelf={dialog?.user !== undefined && dialog.user.id === user?.id}
          onClose={() => setDialog(null)}
          onSecret={setSecret}
        />
        <SecretDialog secret={secret} onClose={() => setSecret(null)} />
      </RequireEditor>
    </>
  );
}
```

`client/App.tsx`: add `const UsersPage = lazy(() => import('./pages/users/UsersPage'));` and `<Route path="users" element={<UsersPage />} />`.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/pages/users client/App.tsx && yarn vitest run client/pages/users && yarn test:client`
Expected: PASS. 10 new tests pass, and the suite is green.

If `vi.spyOn(field, 'select')` can't spy on the jsdom method, spy on `HTMLInputElement.prototype.select` instead. Don't drop the assertion.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add -A client
git commit -q -m "Add the Users admin page with add/edit, own-row protection and temporary passwords" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 5: Reset password and the public contact email

**Files:**
- Create: `client/pages/users/ResetPasswordButton.tsx`, `client/pages/users/DisplayEmailCard.tsx`
- Modify: `client/pages/users/UsersPage.tsx`
- Test: `client/pages/users/usersReset.test.tsx`

**Interfaces:**
- Consumes: `resetUserPassword`, `setDisplayEmail`, `useContact` (Task 2 and 4a); `Secret` (Task 4); `ConfirmDialog`, `isSessionExpired`, `describeSaveError`, `useSaveForm`.
- Produces: `ResetPasswordButton({ user: AdminUser; onSecret })` and `DisplayEmailCard()`.

- [ ] **Step 1: Write the failing tests**

`client/pages/users/usersReset.test.tsx`:

```tsx
import { fireEvent, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { adminUsers, contact, publicRoutes, userAdmin } from '../../test/fixtures';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import UsersPage from './UsersPage';

function renderUsers(overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': userAdmin, ...overrides }));
  renderWithProviders(<UsersPage />, { route: '/users' });
  return fetchMock;
}

const mo = adminUsers[1];

async function confirmReset() {
  fireEvent.click(await screen.findByRole('button', { name: `Reset password for ${mo.name}` }));
  const confirm = screen.getByRole('dialog', { name: `Reset ${mo.name}'s password?` });
  expect(confirm).toHaveTextContent('valid for 7 days');
  fireEvent.click(within(confirm).getByRole('button', { name: 'Reset password' }));
}

describe('Reset password', () => {
  it('toasts when the email was sent', async () => {
    const fetchMock = renderUsers({ [`/api/v1/users/${mo.id}/reset`]: { body: { emailSent: true } } });
    await confirmReset();
    expect(await screen.findByRole('button', { name: `Reset link emailed to ${mo.email}` })).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([u, i]) => String(u).endsWith('/reset') && i?.method === 'POST')).toBe(true);
  });

  it('shows the link to pass on when the email failed', async () => {
    renderUsers({
      [`/api/v1/users/${mo.id}/reset`]: { body: { emailSent: false, resetUrl: 'https://afc.test/reset/abc' } },
    });
    await confirmReset();
    const secret = await screen.findByRole('dialog', { name: `Pass this link on to ${mo.name}` });
    expect(within(secret).getByLabelText('Reset link')).toHaveValue('https://afc.test/reset/abc');
  });

  it('toasts a failure', async () => {
    renderUsers({
      [`/api/v1/users/${mo.id}/reset`]: { status: 500, body: { error: { code: 500, message: 'boom' } } },
    });
    await confirmReset();
    expect(await screen.findByRole('button', { name: "Couldn't reset the password: boom" })).toBeInTheDocument();
  });
});

describe('Public contact email', () => {
  it('shows Not set, then saves a new email', async () => {
    const fetchMock = renderUsers({
      '/api/v1/settings/display-email': { body: { email: 'club@example.test' } },
    });
    const card = (await screen.findByRole('heading', { name: 'Public contact email' })).closest('div') as HTMLElement;
    expect(await within(card).findByText('Not set')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Edit public contact email' }));
    const dialog = screen.getByRole('dialog', { name: 'Public contact email' });
    fireEvent.change(within(dialog).getByLabelText('Email'), { target: { value: 'club@example.test' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Save' }));
    expect(await screen.findByRole('button', { name: 'Contact email saved' })).toBeInTheDocument();
    const put = fetchMock.mock.calls.find(([, i]) => i?.method === 'PUT');
    expect(JSON.parse(String(put?.[1]?.body))).toEqual({ email: 'club@example.test' });
  });

  it('starts from the current email and can clear it', async () => {
    const fetchMock = renderUsers({
      '/api/v1/contact': { body: { ...contact, displayEmail: 'sec@example.test' } },
      '/api/v1/settings/display-email': { body: { email: '' } },
    });
    expect(await screen.findByText('sec@example.test')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Edit public contact email' }));
    const dialog = screen.getByRole('dialog', { name: 'Public contact email' });
    const field = within(dialog).getByLabelText('Email');
    expect(field).toHaveValue('sec@example.test');
    expect(field).toHaveAccessibleDescription('Leave empty to remove it from the Contact page.');
    fireEvent.change(field, { target: { value: '' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Save' }));
    await screen.findByRole('button', { name: 'Contact email saved' });
    const put = fetchMock.mock.calls.find(([, i]) => i?.method === 'PUT');
    expect(JSON.parse(String(put?.[1]?.body))).toEqual({ email: '' });
  });

  it('shows a server error under the field', async () => {
    renderUsers({
      '/api/v1/settings/display-email': {
        status: 422,
        body: { error: { code: 422, message: 'invalid', fields: { email: 'email address is not valid' } } },
      },
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Edit public contact email' }));
    const dialog = screen.getByRole('dialog', { name: 'Public contact email' });
    fireEvent.change(within(dialog).getByLabelText('Email'), { target: { value: 'nope' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Save' }));
    expect(await within(dialog).findByText('email address is not valid')).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `yarn vitest run client/pages/users/usersReset.test.tsx`
Expected: FAIL. There's no "Reset password for …" button and no contact-email card.

- [ ] **Step 3: Implement**

`client/pages/users/ResetPasswordButton.tsx`:

```tsx
import { useState } from 'react';

import { resetUserPassword, type AdminUser } from '../../api/users';
import { useAuth } from '../../auth/useAuth';
import { Button } from '../../components/ui/Button';
import { ConfirmDialog } from '../../components/ui/ConfirmDialog';
import { useToast } from '../../components/ui/toast/useToast';
import { describeSaveError } from '../../lib/saveErrors';
import { isSessionExpired } from '../../lib/session';
import type { Secret } from './SecretDialog';

export function ResetPasswordButton({
  user,
  onSecret,
}: {
  user: AdminUser;
  onSecret: (secret: Secret) => void;
}) {
  const [open, setOpen] = useState(false);
  const toast = useToast();
  const { refresh } = useAuth();

  async function confirm() {
    try {
      const result = await resetUserPassword(user.id);
      setOpen(false);
      if (!result.emailSent && result.resetUrl) {
        onSecret({
          title: `Pass this link on to ${user.name}`,
          message: `We couldn't email ${user.name}. Send them this link to set a new password; it lasts 7 days.`,
          label: 'Reset link',
          value: result.resetUrl,
        });
        return;
      }
      toast.show({ tone: 'success', message: `Reset link emailed to ${user.email}` });
    } catch (err) {
      setOpen(false);
      if (isSessionExpired(err)) {
        await refresh();
        return;
      }
      toast.show({ tone: 'error', message: `Couldn't reset the password: ${describeSaveError(err)}` });
    }
  }

  return (
    <>
      <Button
        size="sm"
        variant="secondary"
        aria-label={`Reset password for ${user.name}`}
        onClick={() => setOpen(true)}
      >
        Reset password
      </Button>
      <ConfirmDialog
        open={open}
        title={`Reset ${user.name}'s password?`}
        message="They'll be emailed a link to set a new one (valid for 7 days) and must use it before they can sign in again."
        confirmLabel="Reset password"
        onConfirm={confirm}
        onCancel={() => setOpen(false)}
      />
    </>
  );
}
```

`client/pages/users/DisplayEmailCard.tsx`:

```tsx
import { useState, type FormEvent } from 'react';

import { setDisplayEmail, useContact } from '../../api/pages';
import { queryKeys } from '../../api/queries';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Card, CardBody } from '../../components/ui/Card';
import { Input } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';

function DisplayEmailForm({ initial, onClose }: { initial: string; onClose: () => void }) {
  const toast = useToast();
  const [email, setEmail] = useState(initial);
  const save = useSaveForm({
    submit: () => setDisplayEmail(email),
    invalidate: [queryKeys.contact],
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Contact email saved' });
      onClose();
    },
  });
  return (
    <form
      noValidate
      className="flex flex-col gap-4"
      onSubmit={(e: FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        void save.run();
      }}
    >
      {save.formError && <Alert tone="error">{save.formError}</Alert>}
      <Field
        label="Email"
        error={save.fieldErrors.email}
        help="Leave empty to remove it from the Contact page."
      >
        <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
      </Field>
      <div className="flex justify-end gap-2">
        <Button variant="secondary" onClick={onClose} disabled={save.busy}>
          Cancel
        </Button>
        <Button type="submit" loading={save.busy}>
          Save
        </Button>
      </div>
    </form>
  );
}

/** The public contact email shown on the Contact page (display-email setting). */
export function DisplayEmailCard() {
  const contact = useContact();
  const [editing, setEditing] = useState(false);
  const email = contact.data?.displayEmail ?? '';
  return (
    <Card className="mb-6">
      <CardBody className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="font-display text-xl font-extrabold uppercase">Public contact email</h2>
          <p>{contact.isPending ? 'Loading…' : email || 'Not set'}</p>
          <p className="text-sm text-muted">Shown on the Contact page.</p>
        </div>
        <Button
          variant="secondary"
          aria-label="Edit public contact email"
          disabled={contact.isPending}
          onClick={() => setEditing(true)}
        >
          Edit
        </Button>
      </CardBody>
      <Modal open={editing} onClose={() => setEditing(false)} title="Public contact email">
        <DisplayEmailForm initial={email} onClose={() => setEditing(false)} />
      </Modal>
    </Card>
  );
}
```

`client/pages/users/UsersPage.tsx`:
- import `DisplayEmailCard` and `ResetPasswordButton`;
- in `UsersTable`, destructure `onSecret` too: `function UsersTable({ currentUserId, onEdit, onSecret }: UsersTableProps)`;
- between the Edit button and the Delete button, add `<ResetPasswordButton user={u} onSecret={onSecret} />`;
- in `UsersPage`, render `<DisplayEmailCard />` as the first child of `RequireEditor`.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `yarn eslint --fix client/pages/users && yarn vitest run client/pages/users && yarn test:client`
Expected: PASS. 6 new tests pass, and the suite is green.

- [ ] **Step 5: Commit**

```bash
yarn typecheck
git add -A client
git commit -q -m "Add password resets and the public contact email to the Users page" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 6: Accessibility, routes, README and verification

**Files:**
- Modify: `client/a11y.test.tsx`, `client/App.test.tsx`, `README.md`

- [ ] **Step 1: Extend the route and accessibility tests**

In `client/App.test.tsx`, add to the route list of the `it.each` that renders as editor (`'routes %s to its edit page for editors'`) the entry `['/players', 'Players']`. Then add:

```tsx
  it('routes /users to the Users page for user admins', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': userAdmin }));
    renderWithProviders(<App />, { route: '/users' });
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Users' }, lazy),
    ).toBeInTheDocument();
  });
```

(import `userAdmin`).

In `client/a11y.test.tsx`:
- add `'/players'` and `'/users'` to `routes`. Both render their `h1` whether or not you have access, so the existing heading wait works;
- after the existing dialog test, add:

```tsx
it.each(['light', 'dark'] as const)('has no axe violations on Users for a user admin (%s)', async (theme) => {
  localStorage.setItem('afc-theme', theme);
  mockFetch(publicRoutes({ '/api/v1/auth/me': userAdmin }));
  const { container } = renderWithProviders(<App />, { route: '/users' });
  await screen.findByText('Una Admin', undefined, { timeout: 3000 });
  expect(await axeViolations(container)).toEqual([]);
});

it('has no axe violations with the Edit user dialog open', async () => {
  mockFetch(publicRoutes({ '/api/v1/auth/me': userAdmin }));
  const { container } = renderWithProviders(<App />, { route: '/users' });
  fireEvent.click(await screen.findByRole('button', { name: 'Edit Mo Manager' }, { timeout: 3000 }));
  await screen.findByRole('dialog', { name: 'Edit Mo Manager' });
  expect(await axeViolations(container)).toEqual([]);
});
```

(import `userAdmin`).

- [ ] **Step 2: Run the tests**

Run: `yarn eslint --fix client/App.test.tsx client/a11y.test.tsx && yarn vitest run client/App.test.tsx client/a11y.test.tsx`
Expected: PASS. If axe reports a violation, fix the markup; don't disable rules.

- [ ] **Step 3: README**

In `README.md`, replace the sentence "The Players and Users admin pages are still on the classic site." with:

```markdown
Signed-in members can browse players at `/app/players` (editors add, edit and delete them), and user admins manage accounts, password resets and the public contact email at `/app/users`. Every page now has an in-app version; the classic templates remain only until the cutover.
```

- [ ] **Step 4: Full verification**

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
git commit -q -m "Extend route and accessibility checks to the admin pages; document them" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```
