# Account Pages Design (Sub-project 4b)

Part of sub-project 4 in moving AFC-GO to the MV-Controller layout.

| # | Sub-project | Status |
|---|---|---|
| 1–3 | Server/API, client scaffold, design system | Done (#11, #12, #13) |
| 4a | Public pages (read-only) | Done (#14) |
| **4b** | **Account pages: account, change password, reset link** (this spec) | Designing |
| 4c | Editing on every page, plus the Users, Players and Edit info pages | Later |
| 5 | Cutover (SPA takes over `/`; delete templates) | Later |

## Context

`main` serves the React client at `/app`. Every public page is ported, and the shell has a sign-in dialog and an account menu. The account menu's **Account** item and the sign-in dialog's reset redirect still point at legacy pages:
- the legacy account page is `/account`;
- the sign-in dialog calls `goTo(res.resetUrl)`, where the server returns `/reset/<token>`.

The legacy account page shows name, email, phone and role read-only. It has a popup for uploading or removing the user's photo (with a preview) and a popup for changing the password. The legacy reset page (`/reset/:url`) asks for a new password and confirmation.

The API already covers all of this, so no server changes are needed:

| Endpoint | Body | Result |
|---|---|---|
| `GET /account` (or the cached `/auth/me`) | — | `CurrentUser` |
| `PUT /account/image` | multipart form, field `image` | `CurrentUser` (updated) |
| `DELETE /account/image` | — | `204` |
| `POST /auth/password` | JSON `{ oldPassword, newPassword, confirmationPassword }` | `204`, or `422` with fields `oldPassword` ("old password is not correct"), `newPassword` ("password needs …"), `confirmationPassword` ("passwords do not match") |
| `GET /auth/reset/:token` | — | `204`, or `404` "reset link is invalid or has expired" |
| `POST /auth/reset/:token` | JSON `{ newPassword, confirmationPassword }` | `204`, `404` as above, or `422` with fields as above |

**Password rules** (`auth.PasswordProblems`): more than 8 characters, with at least one lower-case letter, one upper-case letter, one number and one special character from `@$!%*?&|^£;:/.,<>()_=+~§±#{}-`.

**Accepted image types** (`upload`): JPEG, PNG, GIF, WebP, AVIF, APNG and SVG.

**Reset links:** sign-in returns `/reset/<token>` for reset-flagged accounts, and the admin "password reset" email contains `https://<domain>/reset/<token>`.

## Decisions

- **Reset links stay as they are; the redirect happens in the client.** The SPA gets a `reset/:token` route, which is `/app/reset/<token>` until the cutover. The sign-in dialog maps the server's `/reset/<token>` to that route and navigates in-app. Emails keep pointing at the legacy page until sub-project 5, when `/reset/<token>` becomes the SPA route without any change to the links. No server changes.
- **One Account page with three cards:** Your details (read-only), Photo, and Change password. The forms are inline, with no popups.
- **The server is the only authority on passwords.** The client shows the rules as a hint and places the server's field errors on the right inputs. It only checks that required fields aren't empty.
- **One sign-in dialog, openable from anywhere.** A `SignInProvider` owns the dialog, and `useSignIn().open()` opens it. The header button, the Account page's signed-out prompt and the reset page's success state all use it.

## Pages and flows

### `/account` (`client/pages/account/AccountPage.tsx`)

- **Title:** "Your account". **Page title:** "Account · AFC Aldermaston".
- **While the sign-in check is loading:** the page loading placeholder.
- **Signed out:** an `EmptyState` titled "Sign in to see your account", with the message "Your session may have expired." and a primary **Sign in** button that calls `useSignIn().open()`. Once sign-in succeeds, the page renders the signed-in view (it reads `useAuth()`).
- **Signed in:** three `Card`s, stacked on phones and in a two-column grid on wider screens (Photo and Details side by side, Password full width underneath).

**Your details** (`DetailsCard`):
- A definition list of Name, Email, Phone (only when set), Role, and Team (only when `teamId` is set; it shows the team's name from `useSite().data.teams`, or nothing if it can't be found).
- Then the muted note "To change these, ask a club administrator."

**Photo** (`PhotoCard.tsx`):
- **Current photo:** `ImageWithFallback` with `src={user.imageUrl}` and `fallbackSrc={crest}`, round, 128px, with `alt=""`.
- **No file chosen:**
  - a `FileInput` labelled "Choose a new photo", with `accept` set to the MIME list above;
  - a **Remove photo** danger button, shown only when `user.imageUrl` is set.
- **After choosing a file:**
  - a preview of the chosen file (`URL.createObjectURL`, released with `URL.revokeObjectURL` when the file changes, on Cancel, after saving, and on unmount), shown in place of the current photo;
  - **Save photo** (primary, loading while uploading) and **Cancel** (secondary).
- **Save** sends `PUT /account/image` with form field `image`:
  - on success it sets the query cache for `me` to the returned `CurrentUser`, clears the chosen file and preview, and shows the toast "Photo updated";
  - on an error, the `ApiError` message appears as the file field's error. `fieldError(err, 'file')` or `fieldError(err, 'image')` is used when present; the server's upload errors currently use `file`.
- **Remove** opens a danger `ConfirmDialog`:
  - text: "Remove your photo?" / "Your photo will be replaced by the club crest." / "Remove photo";
  - confirming sends `DELETE /account/image`, then refetches `me` (`useAuth().refresh()`) and shows the toast "Photo removed";
  - an error shows as an error toast, "Couldn't remove your photo: <message>".

**Change password** (`PasswordCard.tsx`):
- A form with **Current password** (`autocomplete=current-password`) and `PasswordFields` (below), plus a **Change password** submit button with a loading state.
- Submitting sends `POST /auth/password`. On success, all three fields clear, a toast says "Password changed", and the user stays signed in. On a `422`, each field shows its server message. Any other error shows in an `Alert` above the button.
- **Empty fields:** if any field is empty, the form isn't sent. Instead each empty field shows "Enter your current password", "Enter a new password" or "Confirm your new password".

### `/reset/:token` (`client/pages/reset/ResetPage.tsx`)

- **Title:** "Reset your password". **Page title:** "Reset password · AFC Aldermaston".
- **Checking the link:** a token that doesn't match `/^[A-Za-z0-9-]{1,100}$/` is treated as invalid without calling the API. Otherwise the page calls `GET /auth/reset/:token` (a query that isn't retried) and shows the loading placeholder meanwhile.
- **Invalid or expired link (`404`):** an `EmptyState` titled "This reset link is invalid or has expired", with the message "Ask a club administrator for a new one, or sign in again if you were sent here after signing in." and a **Home** `ButtonLink` to `/`.
- **Other errors while checking:** `QueryState`'s error-with-Retry.
- **Valid link:** a form with `PasswordFields` and a **Set new password** button (loading state). Submitting sends `POST /auth/reset/:token`.
  - `422`: field errors.
  - `404` (the link expired in the meantime): the invalid-link state.
  - Other errors: an `Alert`.
- **Success:** an `Alert` (success tone) reading "Your password has been changed. You can now sign in with it.", with a primary **Sign in** button that calls `useSignIn().open()`. The form is gone.
- **Signed in:** the page works the same; the reset applies to the token's account.

### Shared: `PasswordFields` (`client/components/page/PasswordFields.tsx`)

- **Props:** `{ newPassword, confirmationPassword, onChange(field, value), errors: { newPassword?, confirmationPassword? } }`.
- It renders two `Field`/`Input` pairs, **New password** and **Confirm new password** (`type=password`, `autocomplete=new-password`).
- The New password field's help text is the rules hint: "More than 8 characters, with a lower-case letter, an upper-case letter, a number and a special character." A field error replaces the help text, as `Field` already does.

### Sign-in dialog and provider

- **`client/components/layout/SignInProvider.tsx`:** holds `open` state and renders the single `SignInDialog`. Placement: `main.tsx` and `renderWithProviders` put it inside `ToastProvider` and wrap `App`/`ui` with it.
- **`useSignIn()`:** returns `{ open(): void }` (context in `signInContext.ts`, hook in `useSignIn.ts`, following the split used for theme and toast).
- **`AccountControl`:** signed out, its **Sign in** button calls `useSignIn().open()`. It no longer renders a `SignInDialog` itself.
- **`SignInDialog` reset redirect:** when `res.resetRequired && res.resetUrl`, it closes the dialog.
  - If `resetUrl` matches `^/reset/([A-Za-z0-9-]+)$`, it calls React Router's `navigate('/reset/<token>')`, which is in-app, so `/app/reset/<token>` in the browser.
  - Otherwise it falls back to `goTo(resetUrl)`, as today.
  - The mapping lives in `client/lib/resetLink.ts` as `appResetPath(url: string): string | null`.
- **Account menu:** the **Account** item becomes `{ label: 'Account', to: '/account' }`, an in-app link. Players, Edit info and Users stay as legacy `href`s until 4c.

### Data (`client/api/account.ts`)

```ts
export interface PasswordChange { oldPassword: string; newPassword: string; confirmationPassword: string }
export interface PasswordReset { newPassword: string; confirmationPassword: string }

uploadAccountImage(file: File): Promise<CurrentUser>        // PUT /account/image, FormData field "image"
removeAccountImage(): Promise<void>                          // DELETE /account/image
changePassword(input: PasswordChange): Promise<void>         // POST /auth/password
checkResetToken(token: string): Promise<void>                // GET /auth/reset/:token
resetPassword(token: string, input: PasswordReset): Promise<void>  // POST /auth/reset/:token
useResetTokenCheck(token: string | null)                     // query key ['auth','reset',token]; disabled for null
```

`apiFetch` already supports `form` bodies and `method`, and it maps `422` field errors into `ApiError.fields`. Tokens are URL-encoded in paths.

### Routes

`App.tsx` gains two lazy routes, `account` → `AccountPage` and `reset/:token` → `ResetPage`.

## Testing

Vitest, Testing Library and jsdom, using `mockFetch`, `publicRoutes` and the other fixtures, and `renderWithProviders` (now including `SignInProvider`). jsdom lacks `URL.createObjectURL` and `URL.revokeObjectURL`, so tests stub them with `vi.stubGlobal` on `URL`.

- **Account page:**
  - **Signed out:** the prompt shows, and its Sign in button opens the "Sign in" dialog.
  - **Signed in:** the details show; Phone and Team appear only when set; the team name comes from `/site`.
  - **Photo:**
    - choosing a file shows the preview (blob URL) with Save and Cancel;
    - Save sends a `PUT /api/v1/account/image` request whose body is `FormData` with an `image` file, then the photo updates from the response and the toast "Photo updated" appears;
    - Cancel restores the current photo and revokes the preview URL;
    - an upload error shows under the file field;
    - Remove asks for confirmation, sends `DELETE`, refetches `me`, and shows the "Photo removed" toast;
    - Remove isn't shown without a photo;
    - a broken photo URL falls back to the crest.
  - **Change password:**
    - it sends the exact JSON;
    - each `422` field message appears on its input;
    - empty fields show the "Enter…" messages and nothing is sent;
    - success clears the fields and shows the toast;
    - the button is busy while the request runs.
- **Reset page:**
  - a malformed token shows the invalid state without calling the API;
  - a `404` check shows the invalid state with the Home link;
  - a valid check shows the form with the rules hint;
  - field errors show;
  - success shows the confirmation, and Sign in opens the dialog;
  - a `404` on submit switches to the invalid state.
- **Sign-in dialog:** a `resetRequired` response with `/reset/abc-123` lands on the reset route in-app (the page heading appears and `goTo` isn't called), and the dialog closes. An unrecognised `resetUrl` still calls `goTo`.
- **`appResetPath`:** maps `/reset/<uuid>`; returns `null` for other paths, an absolute URL, or an empty token.
- **Account menu:** the Account item's `href` is `/account`, and clicking it stays in-app.
- **Accessibility:** `client/a11y.test.tsx` adds `/account` (signed in and signed out) and `/reset/<token>` (valid and invalid), in both themes.
- **The usual checks:** lint, typecheck, test and build stay green, and the Go tests are unchanged.

## Out of scope

- Editing name, email or phone (the API can't).
- The Players, Users and Edit info pages, and all content editing (4c).
- Changing the server's reset-link URLs or email templates (they become SPA routes at the cutover in sub-project 5).
- Removing legacy routes.
