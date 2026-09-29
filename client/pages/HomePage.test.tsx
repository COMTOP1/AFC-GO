import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { mockFetch } from '../test/mockFetch';
import { renderWithProviders } from '../test/render';
import HomePage from './HomePage';

const site = {
  year: 2026,
  visitorCount: 1234,
  version: 'test',
  teams: [
    { id: 1, name: 'First Team', isActive: true, isYouth: false, ages: 99 },
    { id: 2, name: 'Under 12s', isActive: true, isYouth: true, ages: 12 },
  ],
};

const anonymous = {
  status: 401,
  body: { error: { code: 401, message: 'login required' } },
};

describe('HomePage', () => {
  it('shows site info for an anonymous visitor', async () => {
    mockFetch({ '/api/v1/site': { body: site }, '/api/v1/auth/me': anonymous });
    renderWithProviders(<HomePage />);
    expect(await screen.findByText('Not signed in')).toBeInTheDocument();
    expect(await screen.findByText('Visitors: 1234')).toBeInTheDocument();
    expect(screen.getByText('First Team')).toBeInTheDocument();
    expect(screen.getByText('Under 12s')).toBeInTheDocument();
  });

  it('shows who is signed in', async () => {
    mockFetch({
      '/api/v1/site': { body: site },
      '/api/v1/auth/me': {
        body: {
          id: 1,
          name: 'Web Master',
          email: 'webmaster@example.test',
          role: 'Webmaster',
          permissions: { canEdit: true, canManageGallery: true, canManageUsers: true },
        },
      },
    });
    renderWithProviders(<HomePage />);
    expect(await screen.findByText('Signed in as Web Master (Webmaster)')).toBeInTheDocument();
  });

  it('shows a readable error when the API fails', async () => {
    mockFetch({
      '/api/v1/site': {
        status: 500,
        body: { error: { code: 500, message: 'internal server error' } },
      },
      '/api/v1/auth/me': anonymous,
    });
    renderWithProviders(<HomePage />);
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Could not load site information: internal server error',
    );
  });

  it('says so when there are no teams', async () => {
    mockFetch({ '/api/v1/site': { body: { ...site, teams: [] } }, '/api/v1/auth/me': anonymous });
    renderWithProviders(<HomePage />);
    expect(await screen.findByText('No teams yet.')).toBeInTheDocument();
  });
});
