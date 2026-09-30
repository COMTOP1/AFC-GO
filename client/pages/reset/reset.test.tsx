import { act, fireEvent, screen, within } from '@testing-library/react';
import { Link, Route, Routes } from 'react-router';
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
    expect(
      screen.getByRole('heading', { level: 1, name: 'Reset your password' }),
    ).toBeInTheDocument();
  });

  it('asks for empty fields without sending anything', async () => {
    const fetchMock = renderReset(resetToken);
    await screen.findByLabelText('New password');
    fireEvent.click(screen.getByRole('button', { name: 'Set new password' }));
    expect(screen.getByLabelText('New password')).toHaveAccessibleDescription(
      'Enter a new password',
    );
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
    const done = await screen.findByText(
      'Your password has been changed. You can now sign in with it.',
    );
    expect(screen.queryByLabelText('New password')).toBeNull();
    fireEvent.click(
      within(done.closest('[role="status"]') as HTMLElement).getByRole('button', {
        name: 'Sign in',
      }),
    );
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

  it('keeps the success message when the (now used) link is re-checked', async () => {
    let calls = 0;
    const fetchMock = mockFetch(
      publicRoutes({
        [`/api/v1/auth/reset/${resetToken}`]: () =>
          calls++ < 2
            ? { status: 204 }
            : { status: 404, body: { error: { code: 404, message: 'expired' } } },
      }),
    );
    const { queryClient } = renderWithProviders(
      <Routes>
        <Route path="/reset/:token" element={<ResetPage />} />
      </Routes>,
      { route: `/reset/${resetToken}` },
    );
    await screen.findByLabelText('New password');
    fill('Abcdefgh1!', 'Abcdefgh1!');
    await screen.findByText('Your password has been changed. You can now sign in with it.');
    const before = fetchMock.mock.calls.length;
    await act(async () => {
      await queryClient.invalidateQueries();
      await new Promise((r) => setTimeout(r, 20));
    });
    // Whether or not the used link is re-checked, the success message must stay.
    expect(fetchMock.mock.calls.length).toBeGreaterThanOrEqual(before);
    expect(
      screen.getByText('Your password has been changed. You can now sign in with it.'),
    ).toBeInTheDocument();
    expect(screen.queryByText(invalidTitle)).toBeNull();
  });

  it('starts fresh when sent to a different reset link', async () => {
    const oldToken = 'aaaaaaaa-1111-4111-8111-000000000001';
    let calls = 0;
    mockFetch(
      publicRoutes({
        [`/api/v1/auth/reset/${oldToken}`]: () =>
          calls++ === 0
            ? { status: 204 }
            : { status: 404, body: { error: { code: 404, message: 'expired' } } },
      }),
    );
    renderWithProviders(
      <Routes>
        <Route
          path="/reset/:token"
          element={
            <>
              <ResetPage />
              <Link to={`/reset/${resetToken}`}>new link</Link>
            </>
          }
        />
      </Routes>,
      { route: `/reset/${oldToken}` },
    );
    await screen.findByLabelText('New password');
    fill('Abcdefgh1!', 'Abcdefgh1!');
    expect(await screen.findByText(invalidTitle)).toBeInTheDocument();
    fireEvent.click(screen.getByRole('link', { name: 'new link' }));
    expect(await screen.findByLabelText('New password')).toBeInTheDocument();
    expect(screen.queryByText(invalidTitle)).toBeNull();
  });
});
