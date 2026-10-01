import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import App from './App';
import { axeViolations } from './test/axe';
import {
  anonymous,
  editor,
  event,
  expiredResetToken,
  newsArticle,
  publicRoutes,
  resetToken,
  team,
  userAdmin,
} from './test/fixtures';
import type { MockResponse } from './test/mockFetch';
import { mockFetch } from './test/mockFetch';
import { renderWithProviders } from './test/render';

const routes = [
  '/',
  '/teams',
  `/team/${team.id}`,
  '/news',
  `/news/${newsArticle.id}`,
  '/whatson',
  `/whatson/${event.id}`,
  '/gallery',
  '/documents',
  '/programmes',
  '/sponsors',
  '/info',
  '/contact',
  '/design',
  '/account',
  `/reset/${resetToken}`,
  `/reset/${expiredResetToken}`,
  '/nope',
  '/news/new',
  `/news/${newsArticle.id}/edit`,
  '/whatson/new',
  `/whatson/${event.id}/edit`,
  '/info/edit',
  '/teams/new',
  `/team/${team.id}/edit`,
  '/players',
  '/users',
];

const cases: [string, string, MockResponse][] = routes.flatMap((route) => [
  [`${route}, signed out`, route, anonymous] as [string, string, MockResponse],
  [`${route}, signed in`, route, editor] as [string, string, MockResponse],
]);

describe.each(['light', 'dark'] as const)('accessibility (%s theme)', (theme) => {
  it.each(cases)('%s has no axe violations', async (_name, route, me) => {
    localStorage.setItem('afc-theme', theme);
    mockFetch(publicRoutes({ '/api/v1/auth/me': me }));
    const { container } = renderWithProviders(<App />, { route });
    await screen.findAllByRole('button', { name: me === anonymous ? 'Sign in' : /Ed Editor/ });
    if (route.endsWith('/new') || route.endsWith('/edit')) {
      await screen.findAllByText(
        me === anonymous ? "You don't have permission to edit this" : /Save|Edit information/,
        undefined,
        { timeout: 3000 },
      );
      // Audit the loaded rich-text editor, not its placeholder.
      if (me !== anonymous && !route.includes('team')) {
        await screen.findByRole('textbox', { name: /Content|Club information/ }, { timeout: 3000 });
      }
    } else {
      await screen.findByRole('heading', { level: 1 }, { timeout: 3000 });
    }
    // Let the page's own query settle so we audit content, not the skeleton.
    await screen.findByRole('contentinfo');
    await new Promise((r) => setTimeout(r, 50));
    expect(document.documentElement.dataset.theme).toBe(theme);
    expect(await axeViolations(container)).toEqual([]);
  });
});

it('has no axe violations with the Add document dialog open', async () => {
  mockFetch(publicRoutes({ '/api/v1/auth/me': editor }));
  const { container } = renderWithProviders(<App />, { route: '/documents' });
  fireEvent.click(await screen.findByRole('button', { name: 'Add document' }, { timeout: 3000 }));
  await screen.findByRole('dialog', { name: 'Add document' });
  expect(await axeViolations(container)).toEqual([]);
});

it.each(['light', 'dark'] as const)(
  'has no axe violations on Users for a user admin (%s)',
  async (theme) => {
    localStorage.setItem('afc-theme', theme);
    mockFetch(publicRoutes({ '/api/v1/auth/me': userAdmin }));
    const { container } = renderWithProviders(<App />, { route: '/users' });
    await screen.findByText('Una Admin', undefined, { timeout: 3000 });
    expect(await axeViolations(container)).toEqual([]);
  },
);

it('has no axe violations with the Edit user dialog open', async () => {
  mockFetch(publicRoutes({ '/api/v1/auth/me': userAdmin }));
  const { container } = renderWithProviders(<App />, { route: '/users' });
  fireEvent.click(
    await screen.findByRole('button', { name: 'Edit Mo Manager' }, { timeout: 3000 }),
  );
  await screen.findByRole('dialog', { name: 'Edit Mo Manager' });
  expect(await axeViolations(container)).toEqual([]);
});
