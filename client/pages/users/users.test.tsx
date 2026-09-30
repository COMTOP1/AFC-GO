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
    expect(within(other).getByRole('link', { name: mo.email })).toHaveAttribute(
      'href',
      `mailto:${mo.email}`,
    );
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
    expect(within(dialog).getByLabelText('Email')).toHaveAccessibleDescription(
      'Enter an email address',
    );
    expect(within(dialog).getByLabelText('Team')).toHaveAccessibleDescription(
      "Choose the manager's team",
    );
    fireEvent.change(within(dialog).getByLabelText('Name'), { target: { value: 'New' } });
    fireEvent.change(within(dialog).getByLabelText('Email'), { target: { value: 'new@x.test' } });
    fireEvent.change(within(dialog).getByLabelText('Team'), { target: { value: String(team.id) } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add user' }));
    expect(
      await screen.findByRole('button', {
        name: "User added. They've been emailed a temporary password.",
      }),
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
    fireEvent.click(
      within(screen.getByRole('dialog', { name: `Edit ${mo.name}` })).getByRole('button', {
        name: 'Cancel',
      }),
    );
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
              body: {
                user: { ...mo, id: 9, name: 'New' },
                emailSent: false,
                tempPassword: 'Tmp-123!',
              },
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
    expect(
      await screen.findByRole('button', { name: 'Select and copy the text above' }),
    ).toBeInTheDocument();
    expect(select).toHaveBeenCalled();
  });
});
