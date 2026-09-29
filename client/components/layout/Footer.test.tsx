import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import { Footer } from './Footer';

const site = { year: 2026, visitorCount: 12345, version: 'test', teams: [] };
const anonymous = {
  status: 401,
  body: { error: { code: 401, message: 'login required' } },
};
const signedIn = {
  body: {
    id: 1,
    name: 'Jo Smith',
    email: 'jo@example.test',
    role: 'Manager',
    permissions: { canEdit: false, canManageGallery: false, canManageUsers: false },
  },
};

describe('Footer', () => {
  it('shows copyright, credit and social links, but no visitor count, when signed out', async () => {
    mockFetch({ '/api/v1/site': { body: site }, '/api/v1/auth/me': anonymous });
    renderWithProviders(<Footer />);
    const footer = screen.getByRole('contentinfo');
    expect(await screen.findByText(/© 2020–2026 AFC Aldermaston/)).toBeInTheDocument();
    expect(footer).toHaveTextContent('Website provided by BSWDI');
    const credit = screen.getByRole('link', { name: 'BSWDI' });
    expect(credit).toHaveAttribute('href', 'https://bswdi.co.uk');
    expect(credit).toHaveAttribute('target', '_blank');
    expect(credit).toHaveAttribute('rel', 'noopener');
    expect(screen.getByRole('link', { name: 'AFC Aldermaston on Facebook' })).toHaveAttribute(
      'href',
      'https://www.facebook.com/AFC-Aldermaston-114651238068/',
    );
    expect(screen.getByRole('link', { name: 'AFC Aldermaston on X' })).toHaveAttribute(
      'href',
      'https://x.com/afcaldermaston',
    );
    expect(footer).not.toHaveTextContent('Visitor count');
  });

  it('shows the visitor count when signed in', async () => {
    mockFetch({ '/api/v1/site': { body: site }, '/api/v1/auth/me': signedIn });
    renderWithProviders(<Footer />);
    expect(await screen.findByText('Visitor count: 12,345')).toBeInTheDocument();
  });

  it('falls back to the current year before /site loads', () => {
    mockFetch({
      '/api/v1/site': () => new Promise(() => {}),
      '/api/v1/auth/me': anonymous,
    });
    renderWithProviders(<Footer />);
    expect(screen.getByRole('contentinfo')).toHaveTextContent(
      `© 2020–${new Date().getFullYear()} AFC Aldermaston`,
    );
  });
});
