import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import App from './App';
import { axeViolations } from './test/axe';
import { anonymous, editor, event, newsArticle, publicRoutes, team } from './test/fixtures';
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
  '/nope',
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
    await screen.findByRole('button', { name: me === anonymous ? 'Sign in' : /Ed Editor/ });
    await screen.findByRole('heading', { level: 1 }, { timeout: 3000 });
    // Let the page's own query settle so we audit content, not the skeleton.
    await screen.findByRole('contentinfo');
    await new Promise((r) => setTimeout(r, 50));
    expect(document.documentElement.dataset.theme).toBe(theme);
    expect(await axeViolations(container)).toEqual([]);
  });
});
