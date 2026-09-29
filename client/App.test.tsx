import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import App from './App';
import { mockFetch } from './test/mockFetch';
import { renderWithProviders } from './test/render';

function renderAt(path: string) {
  mockFetch({
    '/api/v1/site': { body: { year: 2026, visitorCount: 1, version: 'test', teams: [] } },
    '/api/v1/auth/me': { status: 401, body: { error: { code: 401, message: 'login required' } } },
  });
  return renderWithProviders(<App />, { route: path });
}

describe('App routes', () => {
  it('renders the home page at /', () => {
    renderAt('/');
    expect(screen.getByRole('heading', { level: 1, name: 'AFC Aldermaston' })).toBeInTheDocument();
  });

  it('renders the not-found page for unknown routes', () => {
    renderAt('/no/such/page');
    expect(screen.getByRole('heading', { level: 1, name: 'Page not found' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Go to the start' })).toHaveAttribute('href', '/');
  });

  it('wraps pages in the layout', () => {
    renderAt('/');
    expect(screen.getByRole('banner')).toBeInTheDocument();
    expect(screen.getByRole('main')).toBeInTheDocument();
    expect(screen.getByRole('contentinfo')).toHaveTextContent('AFC Aldermaston');
  });
});
