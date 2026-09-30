import { fireEvent, screen, within } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import type { MockRoute } from '../../test/mockFetch';
import { editor, publicRoutes, team, teamDetail, teams, youthTeam } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import TeamPage from './TeamPage';
import TeamsPage from './TeamsPage';

function renderTeam(route: string, overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes(overrides));
  renderWithProviders(
    <Routes>
      <Route path="/team/:id" element={<TeamPage />} />
    </Routes>,
    { route },
  );
  return fetchMock;
}

describe('TeamsPage', () => {
  it('lists teams as cards with league and badge', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<TeamsPage />);
    const card = await screen.findByRole('link', { name: /First Team/ });
    expect(card).toHaveAttribute('href', `/team/${team.id}`);
    expect(within(card).getByText('Thames Valley Premier · Division 1')).toBeInTheDocument();
    expect(within(card).getByText('Adult')).toBeInTheDocument();
    expect(
      within(screen.getByRole('link', { name: /Under 12s/ })).getByText('Youth'),
    ).toBeInTheDocument();
    expect(document.title).toBe('Teams · AFC Aldermaston');
  });

  it('marks inactive teams (returned to signed-in users)', async () => {
    mockFetch(
      publicRoutes({
        '/api/v1/auth/me': editor,
        '/api/v1/teams': { body: [...teams, { ...team, id: 99, name: 'Vets', isActive: false }] },
      }),
    );
    renderWithProviders(<TeamsPage />);
    expect(
      within(await screen.findByRole('link', { name: /Vets/ })).getByText('Inactive'),
    ).toBeInTheDocument();
    expect(await screen.findByRole('link', { name: 'Add team' })).toHaveAttribute(
      'href',
      '/teams/new',
    );
  });

  it('shows the empty state', async () => {
    mockFetch(publicRoutes({ '/api/v1/teams': { body: [] } }));
    renderWithProviders(<TeamsPage />);
    expect(await screen.findByText('No teams yet')).toBeInTheDocument();
  });
});

describe('TeamPage', () => {
  it('shows details, links, managers, squad and sponsors', async () => {
    renderTeam(`/team/${team.id}`);
    expect(
      await screen.findByRole('heading', { level: 1, name: 'First Team' }),
    ).toBeInTheDocument();
    expect(screen.getByText('Our senior side.')).toBeInTheDocument();
    expect(screen.getByText('Sam Patel')).toBeInTheDocument();
    expect(screen.getByText('Alex Lee')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'jo@example.test' })).toHaveAttribute(
      'href',
      'mailto:jo@example.test',
    );
    expect(screen.getByRole('link', { name: 'League table ↗' })).toHaveAttribute(
      'href',
      teamDetail.team.leagueTableUrl,
    );
    expect(screen.getByRole('link', { name: 'Fixtures ↗' })).toHaveAttribute(
      'href',
      teamDetail.team.fixturesUrl,
    );
    const squad = screen.getByRole('region', { name: 'Squad' });
    expect(within(squad).getByText('Chris Captain')).toBeInTheDocument();
    expect(within(squad).getByText('Captain')).toBeInTheDocument();
    // No photo → the crest stands in.
    // Decorative photos (alt="") have no img role, so select the elements directly.
    expect(squad.querySelectorAll('img')[0]).toHaveAttribute(
      'src',
      expect.stringContaining('crest'),
    );
    expect(screen.getByRole('region', { name: 'Team sponsors' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Teams' })).toHaveAttribute('href', '/teams');
    expect(document.title).toBe('First Team · AFC Aldermaston');
  });

  it('hides the squad entirely for youth teams and omits missing links', async () => {
    renderTeam(`/team/${youthTeam.id}`);
    expect(await screen.findByRole('heading', { level: 1, name: 'Under 12s' })).toBeInTheDocument();
    expect(screen.queryByRole('region', { name: 'Squad' })).toBeNull();
    expect(screen.queryByText(/squad/i)).toBeNull();
    expect(screen.queryByRole('link', { name: 'League table ↗' })).toBeNull();
    expect(screen.queryByRole('link', { name: 'Fixtures ↗' })).toBeNull();
    expect(screen.queryByRole('region', { name: 'Team sponsors' })).toBeNull();
  });

  it('shows not found for an unknown team', async () => {
    renderTeam('/team/999', {
      '/api/v1/teams/999': { status: 404, body: { error: { code: 404, message: 'not found' } } },
    });
    expect(await screen.findByRole('heading', { name: 'Page not found' })).toBeInTheDocument();
  });

  it('shows not found for a malformed id without calling the API', async () => {
    const fetchMock = renderTeam('/team/abc');
    expect(await screen.findByRole('heading', { name: 'Page not found' })).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([u]) => String(u).startsWith('/api/v1/teams'))).toBe(false);
  });

  it('falls back to the crest when a player photo is broken', async () => {
    renderTeam(`/team/${team.id}`);
    const squad = await screen.findByRole('region', { name: 'Squad' });
    const photo = squad.querySelector('img[src="/p/10"]') as HTMLImageElement;
    fireEvent.error(photo);
    expect(squad.querySelector('img[src="/p/10"]')).toBeNull();
    expect(squad.querySelectorAll('img')[1]).toHaveAttribute(
      'src',
      expect.stringContaining('crest'),
    );
  });
});
