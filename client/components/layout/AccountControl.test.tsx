import { act, fireEvent, screen, within } from '@testing-library/react';
import { useLocation } from 'react-router';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { CurrentUser } from '../../api/types';
import { goTo } from '../../lib/navigation';
import { mockFetch, type MockResponse } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import { AccountControl } from './AccountControl';

vi.mock('../../lib/navigation', () => ({ goTo: vi.fn() }));

function Location() {
  const l = useLocation();
  return <output data-testid="location">{l.pathname}</output>;
}

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
    await vi.waitFor(() => expect(goTo).toHaveBeenCalledWith('https://elsewhere.example/reset'));
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

  it('links Account to the in-app page', async () => {
    mockFetch({ '/api/v1/auth/me': { body: user('Mo Manager', 'Manager') } });
    renderWithProviders(
      <>
        <AccountControl />
        <Location />
      </>,
    );
    fireEvent.click(await screen.findByRole('button', { name: /Mo Manager/ }));
    expect(screen.getByRole('menuitem', { name: 'Players' })).toHaveAttribute('href', '/players');
    fireEvent.click(screen.getByRole('menuitem', { name: 'Account' }));
    expect(screen.getByTestId('location')).toHaveTextContent('/account');
  });

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

  it('returns focus to the account button when sign-out is cancelled', async () => {
    mockFetch({ '/api/v1/auth/me': { body: user('Jo Smith', 'Manager') } });
    renderWithProviders(<AccountControl />);
    const trigger = await screen.findByRole('button', { name: /Jo Smith/ });
    fireEvent.click(trigger);
    fireEvent.click(screen.getByRole('menuitem', { name: 'Sign out' }));
    const confirm = screen.getByRole('dialog', { name: 'Sign out?' });
    fireEvent.click(within(confirm).getByRole('button', { name: 'Cancel' }));
    expect(screen.queryByRole('dialog', { name: 'Sign out?' })).toBeNull();
    expect(document.activeElement).toBe(trigger);
  });
});
