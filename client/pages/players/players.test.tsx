import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import { anonymous, editor, manager, players, publicRoutes, youthTeam } from '../../test/fixtures';
import { Location } from '../../test/Location';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import PlayersPage from './PlayersPage';

function renderPlayers(overrides: Record<string, MockRoute> = {}, route = '/players') {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(
    <>
      <Routes>
        <Route path="/players" element={<PlayersPage />} />
      </Routes>
      <Location />
    </>,
    { route },
  );
  return fetchMock;
}

const [sam, yan] = players;

describe('PlayersPage', () => {
  it('asks signed-out visitors to sign in', async () => {
    renderPlayers({ '/api/v1/auth/me': anonymous });
    expect(await screen.findByText('Sign in to see the players')).toBeInTheDocument();
  });

  it('lists players with captain badge, team and age; a Manager gets no controls', async () => {
    renderPlayers({ '/api/v1/auth/me': manager });
    const row = (await screen.findByText(sam.name)).closest('tr') as HTMLElement;
    expect(within(row).getByText('Captain')).toBeInTheDocument();
    expect(within(row).getByText('First Team')).toBeInTheDocument();
    expect(within(row).getByText('2 Apr 1998 (age 28)')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Add player' })).toBeNull();
    expect(screen.queryByRole('button', { name: `Edit ${sam.name}` })).toBeNull();
  });

  it('filters by search and team, kept in the address', async () => {
    renderPlayers();
    await screen.findByText(sam.name);
    fireEvent.change(screen.getByLabelText('Team'), { target: { value: String(youthTeam.id) } });
    expect(screen.queryByText(sam.name)).toBeNull();
    expect(screen.getByText(yan.name)).toBeInTheDocument();
    expect(screen.getByTestId('location')).toHaveTextContent(`team=${youthTeam.id}`);
    fireEvent.change(screen.getByLabelText('Search players'), { target: { value: 'zzz' } });
    expect(screen.getByText('No players match your filters')).toBeInTheDocument();
  });

  it('adds a player, validating first', async () => {
    const fetchMock = renderPlayers({
      '/api/v1/players': (init) =>
        init?.method === 'POST' ? { status: 201, body: sam } : { body: players },
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Add player' }));
    const dialog = screen.getByRole('dialog', { name: 'Add player' });
    expect(
      within(dialog).getByText('Photos are never shown for youth-team or under-18 players.'),
    ).toBeInTheDocument();
    fireEvent.click(within(dialog).getByRole('button', { name: 'Save player' }));
    expect(within(dialog).getByLabelText('Name')).toHaveAccessibleDescription('Enter a name');
    expect(within(dialog).getByLabelText('Team')).toHaveAccessibleDescription('Choose a team');
    expect(within(dialog).getByLabelText('Date of birth')).toHaveAccessibleDescription(
      'Enter the date of birth',
    );
    fireEvent.change(within(dialog).getByLabelText('Name'), { target: { value: 'New Kid' } });
    fireEvent.change(within(dialog).getByLabelText('Team'), {
      target: { value: String(youthTeam.id) },
    });
    fireEvent.change(within(dialog).getByLabelText('Date of birth'), {
      target: { value: '2015-01-20' },
    });
    fireEvent.click(within(dialog).getByRole('checkbox', { name: 'Captain' }));
    fireEvent.click(within(dialog).getByRole('button', { name: 'Save player' }));
    expect(await screen.findByRole('button', { name: 'Player saved' })).toBeInTheDocument();
    const fd = fetchMock.mock.calls.find(([, i]) => i?.method === 'POST')?.[1]?.body as FormData;
    expect(fd.get('name')).toBe('New Kid');
    expect(fd.get('teamId')).toBe(String(youthTeam.id));
    expect(fd.get('dateOfBirth')).toBe('2015-01-20');
    expect(fd.get('isCaptain')).toBe('true');
  });

  it('edits a player, sending an unticked Captain as false', async () => {
    const fetchMock = renderPlayers({ [`/api/v1/players/${sam.id}`]: { body: sam } });
    fireEvent.click(await screen.findByRole('button', { name: `Edit ${sam.name}` }));
    const dialog = screen.getByRole('dialog', { name: `Edit ${sam.name}` });
    expect(within(dialog).getByLabelText('Date of birth')).toHaveValue('1998-04-02');
    fireEvent.click(within(dialog).getByRole('checkbox', { name: 'Captain' }));
    fireEvent.click(within(dialog).getByRole('button', { name: 'Save player' }));
    await waitFor(() =>
      expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'PATCH')).toBe(true),
    );
    const fd = fetchMock.mock.calls.find(([, i]) => i?.method === 'PATCH')?.[1]?.body as FormData;
    expect(fd.get('isCaptain')).toBe('false');
    expect(fd.has('image')).toBe(false);
    expect(fd.has('removeImage')).toBe(false);
  });

  it('opens Add empty straight after closing Edit', async () => {
    renderPlayers();
    fireEvent.click(await screen.findByRole('button', { name: `Edit ${sam.name}` }));
    fireEvent.click(
      within(screen.getByRole('dialog', { name: `Edit ${sam.name}` })).getByRole('button', {
        name: 'Cancel',
      }),
    );
    fireEvent.click(screen.getByRole('button', { name: 'Add player' }));
    expect(
      within(screen.getByRole('dialog', { name: 'Add player' })).getByLabelText('Name'),
    ).toHaveValue('');
  });

  it('deletes a player after confirming', async () => {
    const fetchMock = renderPlayers({ [`/api/v1/players/${sam.id}`]: { status: 204 } });
    fireEvent.click(await screen.findByRole('button', { name: `Delete ${sam.name}` }));
    fireEvent.click(
      within(screen.getByRole('dialog', { name: `Delete ${sam.name}?` })).getByRole('button', {
        name: 'Delete',
      }),
    );
    expect(await screen.findByRole('button', { name: 'Player deleted' })).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true);
  });
});
