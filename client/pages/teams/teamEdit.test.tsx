import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import { editor, manager, publicRoutes, team, teamDetail } from '../../test/fixtures';
import { Location } from '../../test/Location';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import TeamFormPage from './TeamFormPage';
import TeamPage from './TeamPage';
import TeamsPage from './TeamsPage';

function renderTeams(route: string, overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(
    <>
      <Routes>
        <Route path="/teams" element={<TeamsPage />} />
        <Route path="/teams/new" element={<TeamFormPage />} />
        <Route path="/team/:id" element={<TeamPage />} />
        <Route path="/team/:id/edit" element={<TeamFormPage />} />
      </Routes>
      <Location />
    </>,
    { route },
  );
  return fetchMock;
}

describe('Team editing', () => {
  it('shows Add team to editors and hides it from a Manager', async () => {
    renderTeams('/teams');
    expect(await screen.findByRole('link', { name: 'Add team' })).toHaveAttribute(
      'href',
      '/teams/new',
    );
  });

  it('hides Edit/Delete on a team from a Manager', async () => {
    renderTeams(`/team/${team.id}`, { '/api/v1/auth/me': manager });
    await screen.findByRole('heading', { level: 1, name: team.name });
    expect(screen.queryByRole('link', { name: 'Edit' })).toBeNull();
  });

  it('asks for a name and age group; new teams default to active', async () => {
    const fetchMock = renderTeams('/teams/new');
    fireEvent.click(await screen.findByRole('button', { name: 'Save team' }));
    expect(screen.getByLabelText('Name')).toHaveAccessibleDescription('Enter a name');
    expect(screen.getByLabelText('Age group')).toHaveAccessibleDescription('Choose the age group');
    expect(screen.getByRole('checkbox', { name: 'Active team' })).toBeChecked();
    expect(screen.getByRole('checkbox', { name: 'Youth team' })).not.toBeChecked();
    expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'POST')).toBe(false);
  });

  it('creates a team with every field and opens it', async () => {
    const fetchMock = renderTeams('/teams/new', {
      '/api/v1/teams': { status: 201, body: { ...team, id: 30, name: 'Under 10s' } },
    });
    fireEvent.change(await screen.findByLabelText('Name'), { target: { value: 'Under 10s' } });
    fireEvent.change(screen.getByLabelText('Age group'), { target: { value: '10' } });
    fireEvent.change(screen.getByLabelText('Coach'), { target: { value: 'Sam' } });
    fireEvent.click(screen.getByRole('checkbox', { name: 'Youth team' }));
    fireEvent.click(screen.getByRole('button', { name: 'Save team' }));
    await waitFor(() => expect(screen.getByTestId('location')).toHaveTextContent('/team/30'));
    const fd = fetchMock.mock.calls.find(([, i]) => i?.method === 'POST')?.[1]?.body as FormData;
    expect(fd.get('name')).toBe('Under 10s');
    expect(fd.get('ages')).toBe('10');
    expect(fd.get('coach')).toBe('Sam');
    expect(fd.get('isYouth')).toBe('true');
    expect(fd.get('isActive')).toBe('true');
    expect(fd.get('physio')).toBe('');
  });

  it('loads a team for editing and sends an unticked Active as false', async () => {
    const fetchMock = renderTeams(`/team/${team.id}/edit`, {
      [`/api/v1/teams/${team.id}`]: () => ({ body: teamDetail }),
    });
    expect(await screen.findByLabelText('Name')).toHaveValue(team.name);
    expect(screen.getByLabelText('League table URL')).toHaveValue(teamDetail.team.leagueTableUrl);
    fireEvent.click(screen.getByRole('checkbox', { name: 'Active team' }));
    fireEvent.click(screen.getByRole('button', { name: 'Save team' }));
    await waitFor(() =>
      expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'PATCH')).toBe(true),
    );
    const fd = fetchMock.mock.calls.find(([, i]) => i?.method === 'PATCH')?.[1]?.body as FormData;
    expect(fd.get('isActive')).toBe('false');
    expect(fd.get('leagueTable')).toBe(teamDetail.team.leagueTableUrl);
    expect(fd.has('image')).toBe(false);
  });

  it('warns that deleting unlinks players, sponsors and managers', async () => {
    const fetchMock = renderTeams(`/team/${team.id}`, {
      [`/api/v1/teams/${team.id}`]: () => ({ body: teamDetail }),
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Delete' }));
    const dialog = screen.getByRole('dialog', { name: 'Delete this team?' });
    expect(dialog).toHaveTextContent(
      'Its players, sponsors and managers will be unlinked from it.',
    );
    fireEvent.click(within(dialog).getByRole('button', { name: 'Delete' }));
    await waitFor(() => expect(screen.getByTestId('location')).toHaveTextContent(/^\/teams$/));
    expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true);
  });
});
