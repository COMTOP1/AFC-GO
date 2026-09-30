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
    const fetchMock = renderUsers({
      [`/api/v1/users/${mo.id}/reset`]: { body: { emailSent: true } },
    });
    await confirmReset();
    expect(
      await screen.findByRole('button', { name: `Reset link emailed to ${mo.email}` }),
    ).toBeInTheDocument();
    expect(
      fetchMock.mock.calls.some(([u, i]) => String(u).endsWith('/reset') && i?.method === 'POST'),
    ).toBe(true);
  });

  it('shows the link to pass on when the email failed', async () => {
    renderUsers({
      [`/api/v1/users/${mo.id}/reset`]: {
        body: { emailSent: false, resetUrl: 'https://afc.test/reset/abc' },
      },
    });
    await confirmReset();
    const secret = await screen.findByRole('dialog', { name: `Pass this link on to ${mo.name}` });
    expect(within(secret).getByLabelText('Reset link')).toHaveValue('https://afc.test/reset/abc');
  });

  it('toasts a failure', async () => {
    renderUsers({
      [`/api/v1/users/${mo.id}/reset`]: {
        status: 500,
        body: { error: { code: 500, message: 'boom' } },
      },
    });
    await confirmReset();
    expect(
      await screen.findByRole('button', { name: "Couldn't reset the password: boom" }),
    ).toBeInTheDocument();
  });
});

describe('Public contact email', () => {
  it('shows Not set, then saves a new email', async () => {
    const fetchMock = renderUsers({
      '/api/v1/settings/display-email': { body: { email: 'club@example.test' } },
    });
    const card = (await screen.findByRole('heading', { name: 'Public contact email' })).closest(
      'div',
    ) as HTMLElement;
    expect(await within(card).findByText('Not set')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Edit public contact email' }));
    const dialog = screen.getByRole('dialog', { name: 'Public contact email' });
    fireEvent.change(within(dialog).getByLabelText('Email'), {
      target: { value: 'club@example.test' },
    });
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
        body: {
          error: { code: 422, message: 'invalid', fields: { email: 'email address is not valid' } },
        },
      },
    });
    // Edit is enabled once the current email has loaded.
    await screen.findByText('Not set');
    fireEvent.click(screen.getByRole('button', { name: 'Edit public contact email' }));
    const dialog = screen.getByRole('dialog', { name: 'Public contact email' });
    fireEvent.change(within(dialog).getByLabelText('Email'), { target: { value: 'nope' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Save' }));
    expect(await within(dialog).findByText('email address is not valid')).toBeInTheDocument();
  });
});
