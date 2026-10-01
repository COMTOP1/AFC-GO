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
