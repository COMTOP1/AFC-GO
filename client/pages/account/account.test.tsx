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

function renderAccount(me: MockRoute, overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': me, ...overrides }));
  renderWithProviders(<AccountPage />, { route: '/account' });
  return fetchMock;
}

let revokeObjectURL: ReturnType<typeof vi.fn>;
beforeEach(() => {
  let n = 0;
  revokeObjectURL = vi.fn();
  Object.assign(URL, { createObjectURL: vi.fn(() => `blob:preview-${++n}`), revokeObjectURL });
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
    fireEvent.change(within(dialog).getByLabelText('Email'), {
      target: { value: 'jo@example.test' },
    });
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
    expect(
      within(details).getByText('To change these, ask a club administrator.'),
    ).toBeInTheDocument();
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
    const fetchMock = renderAccount(
      { body: baseUser },
      { '/api/v1/account/image': { body: updated } },
    );
    await screen.findByRole('region', { name: 'Photo' });
    const file = choose();
    expect(
      within(photoCard()).getByRole('img', { name: 'Preview of your new photo' }),
    ).toHaveAttribute('src', 'blob:preview-1');
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
    expect(
      within(photoCard()).getByRole('img', { name: 'Preview of your new photo' }),
    ).toHaveAttribute('src', 'blob:preview-2');
    fireEvent.click(within(photoCard()).getByRole('button', { name: 'Cancel' }));
    expect(revokeObjectURL).toHaveBeenCalledWith('blob:preview-2');
    expect(photoCard().querySelector('img')).toHaveAttribute('src', '/files/user/1');
  });

  it('shows an upload error under the file field', async () => {
    renderAccount(
      { body: baseUser },
      {
        '/api/v1/account/image': {
          status: 422,
          body: {
            error: { code: 422, message: 'invalid', fields: { file: 'file is not an image' } },
          },
        },
      },
    );
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
    expect(photoCard().querySelector('img')).toHaveAttribute(
      'src',
      expect.stringContaining('crest'),
    );
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
    const fetchMock = renderAccount(
      { body: baseUser },
      { '/api/v1/auth/password': { status: 204 } },
    );
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
    renderAccount(
      { body: baseUser },
      {
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
      },
    );
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
    expect(screen.getByLabelText('New password')).toHaveAccessibleDescription(
      'Enter a new password',
    );
    expect(fetchMock.mock.calls.some(([u]) => String(u).endsWith('/auth/password'))).toBe(false);
  });

  it('shows the button as busy while the request runs', async () => {
    let answer: (r: MockResponse) => void = () => {};
    renderAccount(
      { body: baseUser },
      {
        '/api/v1/auth/password': () =>
          new Promise<MockResponse>((resolve) => {
            answer = resolve;
          }),
      },
    );
    await screen.findByLabelText('Current password');
    fillPassword('a', 'Abcdefgh1!', 'Abcdefgh1!');
    const button = screen.getByRole('button', { name: 'Change password' });
    await waitFor(() => expect(button).toHaveAttribute('aria-busy', 'true'));
    answer({ status: 204 });
    await waitFor(() => expect(button).not.toHaveAttribute('aria-busy'));
  });

  describe('AccountPage when the session has expired', () => {
    const expired = { status: 401, body: { error: { code: 401, message: 'login required' } } };

    it('switches to the sign-in prompt when changing the password gets a 401', async () => {
      let me: MockResponse = { body: baseUser };
      renderAccount(() => me, {
        '/api/v1/auth/password': () => {
          me = anonymous;
          return expired;
        },
      });
      fireEvent.change(await screen.findByLabelText('Current password'), {
        target: { value: 'a' },
      });
      fireEvent.change(screen.getByLabelText('New password'), { target: { value: 'Abcdefgh1!' } });
      fireEvent.change(screen.getByLabelText('Confirm new password'), {
        target: { value: 'Abcdefgh1!' },
      });
      fireEvent.click(screen.getByRole('button', { name: 'Change password' }));
      expect(await screen.findByText('Sign in to see your account')).toBeInTheDocument();
      expect(screen.queryByText('login required')).toBeNull();
    });

    it('switches to the sign-in prompt when a photo upload gets a 401', async () => {
      let me: MockResponse = { body: baseUser };
      renderAccount(() => me, {
        '/api/v1/account/image': () => {
          me = anonymous;
          return expired;
        },
      });
      await screen.findByRole('region', { name: 'Photo' });
      choose();
      fireEvent.click(within(photoCard()).getByRole('button', { name: 'Save photo' }));
      expect(await screen.findByText('Sign in to see your account')).toBeInTheDocument();
    });

    it('explains a photo that is too large', async () => {
      renderAccount(
        { body: baseUser },
        {
          '/api/v1/account/image': {
            status: 413,
            body: { error: { code: 413, message: 'Request Entity Too Large' } },
          },
        },
      );
      await screen.findByRole('region', { name: 'Photo' });
      choose();
      fireEvent.click(within(photoCard()).getByRole('button', { name: 'Save photo' }));
      await waitFor(() =>
        expect(screen.getByLabelText('Choose a new photo')).toHaveAccessibleDescription(
          'That photo is too large (15 MB maximum).',
        ),
      );
    });
  });
});
