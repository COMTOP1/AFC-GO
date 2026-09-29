import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import App from './App';
import { axeViolations } from './test/axe';
import { mockFetch, type MockResponse } from './test/mockFetch';
import { renderWithProviders } from './test/render';

const site = {
  year: 2026,
  visitorCount: 42,
  version: 'test',
  teams: [{ id: 1, name: 'First Team', isActive: true, isYouth: false, ages: 99 }],
};
const anonymous: MockResponse = {
  status: 401,
  body: { error: { code: 401, message: 'login required' } },
};
const signedIn: MockResponse = {
  body: {
    id: 1,
    name: 'Jo Smith',
    email: 'jo@example.test',
    role: 'Club Secretary',
    permissions: { canEdit: true, canManageGallery: true, canManageUsers: true },
  },
};

describe.each(['light', 'dark'] as const)('accessibility (%s theme)', (theme) => {
  it.each([
    ['home, signed out', '/', anonymous],
    ['home, signed in', '/', signedIn],
    ['design page', '/design', anonymous],
    ['not found', '/nope', anonymous],
  ])('%s has no axe violations', async (_name, route, me) => {
    localStorage.setItem('afc-theme', theme);
    mockFetch({ '/api/v1/site': { body: site }, '/api/v1/auth/me': me });
    const { container } = renderWithProviders(<App />, { route });
    await screen.findByRole('contentinfo');
    await screen.findByRole('button', { name: me === anonymous ? 'Sign in' : /Jo Smith/ });
    expect(document.documentElement.dataset.theme).toBe(theme);
    expect(await axeViolations(container)).toEqual([]);
  });
});
