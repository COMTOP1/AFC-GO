# Players and Users admin pages: design (sub-project 4c-2)

**Status:** approved in conversation on 2026-09-30.

**Context:** sub-projects 4a (public pages), 4b (account pages) and 4c-1 (content editing) are merged. 4c-2 moves the last classic-only screens, **Players** and **Users** (including the public contact email), into the React client at `/app`. After this, only the cutover (sub-project 5) still depends on the classic templates.

**Scope:** client only. There are no server changes; the `/api/v1` endpoints below already exist.

## Goals

- Signed-in members can browse players. Editors can add, edit and delete them.
- User admins (`permissions.canManageUsers`) can list, add, edit and delete users, reset a user's password, and set the public contact email.
- When the server can't send an email, the admin gets the temporary password or reset link on screen to pass on.
- An admin can't lock themselves out from the Users page.
- The user menu's **Players** and **Users** open the in-app pages.

## Non-goals

- Server changes, including server-side protection of one's own account (the client guards it; see Users).
- Bulk actions, CSV import or export, and pagination (club-sized lists).
- Changing who can see what: the server already applies the safeguarding rules for player photos.

## API used

| Call | Method and path | Body | Guard | Response |
|---|---|---|---|---|
| List players | `GET /players` | none | signed in | `Player[]` |
| Add player | `POST /players` | multipart: `name`*, `teamId`*, `dateOfBirth`* (YYYY-MM-DD), `position`, `isCaptain`, `image` | canEdit | 201 `Player` |
| Edit player | `PATCH /players/:id` | multipart, only the fields present: as above, plus `removeImage` | canEdit | 200 `Player` |
| Delete player | `DELETE /players/:id` | none | canEdit | 204 |
| List users | `GET /users` | none | canManageUsers | `AdminUser[]` |
| Add user | `POST /users` | multipart: `name`*, `email`*, `phone`, `role`* (code), `teamId` (managers), `image` | canManageUsers | 201 `{ user, emailSent, tempPassword? }` |
| Edit user | `PATCH /users/:id` | multipart, only the fields present: as above, plus `removeImage` | canManageUsers | 200 `AdminUser` |
| Delete user | `DELETE /users/:id` | none | canManageUsers | 204 |
| Reset password | `POST /users/:id/reset` | none | canManageUsers | 200 `{ emailSent, resetUrl? }` |
| Contact email | `PUT /settings/display-email` | JSON `{ email }` (empty clears it) | canManageUsers | 200 `{ email }` |
| Current contact email | `GET /contact` | none | public | `ContactData.displayEmail` |

**Types:**
- `Player`: `{ id, name, position?, isCaptain, dateOfBirth?, age?, team?: { id, name, isYouth }, imageUrl? }`. `imageUrl` is absent whenever the server hides the photo: youth team, under 18, or a nonsensical date of birth.
- `AdminUser`: `{ id, name, email, phone?, role, roleCode, teamId?, imageUrl? }`, where `role` is the display name and `roleCode` the input code.

**Role codes and labels** (`client/lib/roles.ts`, in this order):

| Code | Label |
|---|---|
| `photographer` | Photographer |
| `manager` | Manager |
| `programme_editor` | Programme Editor |
| `league_secretary` | League Secretary |
| `treasurer` | Treasurer |
| `safeguarding_officer` | Safeguarding Officer |
| `club_secretary` | Club Secretary |
| `chairperson` | Chairperson |
| `webmaster` | Webmaster |

**Server behaviour the client relies on:**
- **Users:**
  - 422 field errors can come on `name`, `email` ("email address is not valid"), `role` or `teamId` ("managers need an existing team").
  - A duplicate email is a **409** with the message "email address is already in use".
  - `teamId` is ignored unless the role is Manager.
  - A reset forces a new password at the next sign-in, and its link lasts 7 days.
- **Players:** 422 field errors come on the field names above.

## Shared behaviour

- **Forms** open in dialogs over the table, built from the 4c-1 toolkit:
  - `useSaveForm`: field errors, 401 re-reads the session, 413 gets the "file too large" message;
  - `ImageField`: the full-size preview and "Remove image" on edit;
  - `DeleteButton`, for confirm-then-delete.
- **Each dialog's form component mounts only while the dialog is open,** so every open starts empty. This is the 4c-1 pattern.
- **Filters are kept in the URL** (`?q=`, `?team=`, `?role=`), like Documents and Programmes. Search uses `matchesQuery`.
- **Thumbnails:** a new `components/page/Thumb.tsx`, a 40 px round photo that falls back to the club crest.
- **Tables** use the design-system `Table`, `THead`, `TBody`, `Tr`, `Th` and `Td`. On narrow screens the table scrolls sideways inside its own box; the page never does.

## Players page (`/app/players`)

- **Access:**
  - Any signed-in user can view it.
  - Signed out, it shows the new **`RequireSignIn`** gate: "Sign in to see the players", the message "Your session may have expired.", and a Sign in button.
  - `RequireSignIn` is taken out of `AccountPage`, and the Account page uses it too.
- **Header:** "Players", plus **Add player** for editors (`canEdit`).
- **Filters:**
  - a search box labelled "Search players", matching on name;
  - a **Team** select with "All teams" and then every team from `useTeams()`, which includes inactive teams for signed-in users.
- **Table columns:**
  - **Photo:** the thumbnail.
  - **Name:** with a "Captain" badge when `isCaptain`.
  - **Position.**
  - **Team:** the team's name, or "No team" when missing.
  - **Date of birth:** `formatDate(dateOfBirth)` then " (age N)", or "—" when missing.
  - **Actions,** for editors only: **Edit**, and **Delete**, labelled "Delete <name>".
- **Sort order:** by team name, then player name.
- **Empty states:**
  - no players at all: "No players yet";
  - nothing matching: "No players match your filters".
- **Add or edit player dialog** (title "Add player" or "Edit <name>"):
  - **Name*:** "Enter a name" when empty.
  - **Team*:** a select, "Choose a team" when empty.
  - **Date of birth*:** a date input, "Enter the date of birth" when empty.
  - **Position.**
  - **Captain:** a checkbox.
  - **Photo:** `ImageField`, with Remove on edit. Its help text reads "Photos are never shown for youth-team or under-18 players."
  - **Buttons:** Cancel, and **Save player**.
  - **Edit sends** every text field, `isCaptain` as true or false, and the image fields only when touched, as in 4c-1.
  - **Toast:** "Player saved".
- **Delete:** the confirmation reads "Delete <name>?" / "This can't be undone.", then the toast "Player deleted".
- **Refreshes** after saving or deleting: `['players']` and `['team']` (team pages show squads).

## Users page (`/app/users`)

- **Access:** `RequireEditor permission="canManageUsers"`, which gains that option. Anyone else sees "You don't have permission to edit this".

### Public contact email card

A card above the table.
- **Heading:** "Public contact email".
- **Text:** the current `displayEmail` from `useContact()`, or "Not set", with the note "Shown on the Contact page."
- **Button:** **Edit**.
- **Dialog "Public contact email":**
  - one **Email** field (type email), with the help text "Leave empty to remove it from the Contact page.";
  - a Save button;
  - on success the toast "Contact email saved", and `['contact']` refreshes;
  - server errors show under the field.

### Users table

- **Header:** "Users" and **Add user**.
- **Filters:**
  - a search box labelled "Search users", matching on name, email and role;
  - a **Role** select with "All roles" and then the nine roles, filtering on `roleCode`.
- **Table columns:**
  - **Photo:** the thumbnail.
  - **Name:** with a "You" badge when `id === currentUser.id`.
  - **Email:** a `mailto:` link.
  - **Phone.**
  - **Role:** the display name.
  - **Team:** for managers, the team's name from `useTeams()`; otherwise empty.
  - **Actions:**
    - **Edit;**
    - **Reset password,** labelled "Reset password for <name>";
    - **Delete,** labelled "Delete <name>". Your own row has **no Delete**.
- **Sort order:** by name.
- **Empty states:** "No users match your filters".

### Add or edit user dialog

The dialog is titled "Add user" or "Edit <name>".
- **Fields:**
  - **Name*:** "Enter a name" when empty.
  - **Email*:** type email, "Enter an email address" when empty.
  - **Phone.**
  - **Role*:** a select, "Choose a role" when empty.
  - **Team:** only shown when the role is Manager, and required then ("Choose the manager's team"). Every team is offered.
  - **Photo:** `ImageField`, with Remove on edit.
- **Your own row:** the Role select is disabled, with the help text "Ask another administrator to change your role." `role` is **not sent** when saving your own row.
- **What's sent:**
  - `role` is the code;
  - `teamId` is sent only when the role is Manager;
  - `phone` is always sent on edit, so an empty value clears it;
  - image fields are sent only when touched.
- **A 409** puts "That email address is already in use" under Email.
- **Buttons:** Cancel, and **Save user** (edit) or **Add user** (add).
- **After editing:** the toast "User saved". Refreshes `['users']`, plus `['auth','me']` when you edited yourself, so the header updates, and `['contact']` and `['team']`, because those pages show managers and contacts.
- **After adding:**
  - If `emailSent`: the toast "User added. They've been emailed a temporary password."
  - Otherwise the add dialog closes and a **SecretDialog** opens:
    - title "Pass this on to <name>";
    - the text "We couldn't email <name>. Give them this temporary password; they'll choose a new one when they first sign in. It won't be shown again.";
    - the password in a read-only monospace field, a **Copy** button and **Done**.

### Reset password

- **Confirmation:** "Reset <name>'s password?". The message reads "They'll be emailed a link to set a new one (valid for 7 days) and must use it before they can sign in again.", and the confirm button is labelled **Reset password**.
- **Email sent:** the toast "Reset link emailed to <email>".
- **Email not sent:** a SecretDialog titled "Pass this link on to <name>", with the text "We couldn't email <name>. Send them this link to set a new password; it lasts 7 days.", the `resetUrl`, **Copy** and **Done**.
- **Errors:** the toast "Couldn't reset the password: <message>".
- **Your own row:** allowed.

### Delete user

- **Confirmation:** "Delete <name>?" / "They will no longer be able to sign in. This can't be undone."
- **Then:** the toast "User deleted". Refreshes `['users']`, `['contact']` and `['team']`.

### SecretDialog (`pages/users/SecretDialog.tsx`)

- Takes `{ open, title, message, value, onClose }`.
- **Copy** uses `navigator.clipboard.writeText` and toasts "Copied".
- If the clipboard isn't available or fails, it selects the text in the field and toasts "Select and copy the text above".

## Navigation

In `AccountControl`, **Players** (`to: '/players'`) and **Users** (`to: '/users'`) become in-app links. Their visibility rules don't change. The routes `players` and `users` are lazily loaded in `App.tsx`.

## Files

- **API:**
  - `client/api/players.ts`: `Player`, `usePlayers`, `PlayerInput`, `createPlayer`, `updatePlayer`, `deletePlayer`.
  - `client/api/users.ts`: `AdminUser`, `useUsers`, `UserInput`, `createUser`, `updateUser`, `deleteUser`, `resetUserPassword`, `CreatedUser`, `ResetResult`.
  - `client/api/pages.ts`: adds `setDisplayEmail(email)`.
  - `client/api/queries.ts`: adds the keys `players` and `users`.
- **Shared code:**
  - `client/lib/roles.ts`: `ROLES`.
  - `client/components/page/Thumb.tsx`.
  - `client/components/edit/RequireSignIn.tsx`.
  - `client/components/edit/RequireEditor.tsx`: adds the `canManageUsers` permission.
- **Players:** `client/pages/players/PlayersPage.tsx` and `PlayerDialog.tsx`.
- **Users:** `client/pages/users/UsersPage.tsx`, `UserDialog.tsx`, `ResetPasswordButton.tsx`, `DisplayEmailCard.tsx` and `SecretDialog.tsx`.
- **Test fixtures:** `players`, `adminUsers`, and a `userAdmin` signed-in fixture (`canManageUsers` without `canEdit`).

## Testing

- **API payloads:**
  - player create and edit multipart, with `isCaptain=false` sent explicitly;
  - user create and edit multipart, with `teamId` only for Manager and no `role` on your own row;
  - the reset POST;
  - the display-email PUT JSON.
- **Players page:**
  - signed out shows the sign-in gate;
  - a Manager can view but gets no controls;
  - an editor can add, edit and delete;
  - search and the team filter are kept in the URL;
  - the captain badge and the age;
  - the photo help text;
  - validation messages.
- **Users page:**
  - an editor who isn't a user admin sees the permission message;
  - a user admin who isn't an editor gets full access;
  - your own row has no Delete and a disabled Role;
  - the Team field appears only for Manager;
  - a 409 shows under Email;
  - add with `emailSent` true and false, including the SecretDialog and Copy with and without the clipboard;
  - reset with `emailSent` true and false;
  - delete;
  - display email set and cleared;
  - the search and role filters.
- **AccountControl:** Players and Users navigate within the app.
- **Accessibility:** axe on `/players` and `/users`, signed in and signed out, in both themes, plus an open user dialog.
- **README:** all admin now happens in the app; the classic pages remain only until the cutover.
