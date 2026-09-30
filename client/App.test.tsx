import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import App from './App';
import {
  editor,
  event,
  newsArticle,
  publicRoutes,
  resetToken,
  team,
  userAdmin,
} from './test/fixtures';
import { mockFetch } from './test/mockFetch';
import { renderWithProviders } from './test/render';

function renderAt(path: string) {
  mockFetch(publicRoutes());
  return renderWithProviders(<App />, { route: path });
}

const lazy = { timeout: 3000 };

describe('App routes', () => {
  it('renders the home page at /', () => {
    renderAt('/');
    expect(screen.getByRole('heading', { level: 1, name: 'AFC Aldermaston' })).toBeInTheDocument();
  });

  it.each([
    ['/teams', 'Teams'],
    [`/team/${team.id}`, team.name],
    ['/news', 'News'],
    [`/news/${newsArticle.id}`, newsArticle.title],
    ['/whatson', "What's On"],
    [`/whatson/${event.id}`, event.title],
    ['/gallery', 'Gallery'],
    ['/documents', 'Documents'],
    ['/programmes', 'Programmes'],
    ['/sponsors', 'Sponsors'],
    ['/info', 'Information'],
    ['/contact', 'Contact'],
    ['/design', 'Design system'],
    ['/account', 'Your account'],
    [`/reset/${resetToken}`, 'Reset your password'],
  ])('routes %s to its page', async (path, heading) => {
    renderAt(path);
    expect(
      await screen.findByRole('heading', { level: 1, name: heading }, lazy),
    ).toBeInTheDocument();
  });

  it.each([
    ['/news/new', 'New article'],
    [`/news/${newsArticle.id}/edit`, 'Edit article'],
    ['/whatson/new', 'New event'],
    [`/whatson/${event.id}/edit`, 'Edit event'],
    ['/info/edit', 'Edit information'],
    ['/teams/new', 'New team'],
    [`/team/${team.id}/edit`, 'Edit team'],
    ['/players', 'Players'],
  ])('routes %s to its edit page for editors', async (path, heading) => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': editor }));
    renderWithProviders(<App />, { route: path });
    expect(
      await screen.findByRole('heading', { level: 1, name: heading }, lazy),
    ).toBeInTheDocument();
  });

  it('routes /users to the Users page for user admins', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': userAdmin }));
    renderWithProviders(<App />, { route: '/users' });
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Users' }, lazy),
    ).toBeInTheDocument();
  });

  it('renders the not-found page for unknown routes', async () => {
    renderAt('/no/such/page');
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Page not found' }, lazy),
    ).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Go to the start' })).toHaveAttribute('href', '/');
  });

  it('wraps pages in the layout', () => {
    renderAt('/');
    const banner = screen.getByRole('banner');
    expect(banner).toHaveTextContent('AFC Aldermaston');
    expect(banner).toHaveTextContent('Facta Non Verba');
    expect(screen.getByRole('link', { name: 'AFC Aldermaston home' })).toHaveAttribute('href', '/');
    expect(screen.getByRole('img', { name: 'The FA Charter Standard' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Skip to content' })).toHaveAttribute(
      'href',
      '#content',
    );
    expect(screen.getByRole('main')).toHaveAttribute('id', 'content');
    expect(screen.getByRole('navigation', { name: 'Main' })).toBeInTheDocument();
    expect(screen.getByRole('contentinfo')).toHaveTextContent('AFC Aldermaston');
  });
});
