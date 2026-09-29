import { screen, waitFor } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { anonymous, editor, events, programmes, teams } from '../test/fixtures';
import { mockFetch } from '../test/mockFetch';
import { renderWithProviders } from '../test/render';
import { useNewsArticle } from './news';
import { useProgrammes } from './programmes';
import { queryKeys } from './queries';
import { useTeams } from './teams';
import { useWhatsOnList } from './whatson';

describe('section hooks', () => {
  it('useWhatsOnList asks the API for the period', async () => {
    const fetchMock = mockFetch({
      '/api/v1/auth/me': anonymous,
      '/api/v1/whatson?period=past': { body: events },
    });
    function Probe() {
      const q = useWhatsOnList('past');
      return <p>{q.data ? `${q.data.length} events` : 'loading'}</p>;
    }
    renderWithProviders(<Probe />);
    expect(await screen.findByText(`${events.length} events`)).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([u]) => String(u) === '/api/v1/whatson?period=past')).toBe(
      true,
    );
  });

  it('useProgrammes omits season for all, and sends it otherwise', async () => {
    const fetchMock = mockFetch({
      '/api/v1/auth/me': anonymous,
      '/api/v1/programmes': { body: programmes },
      '/api/v1/programmes?season=2': { body: programmes.slice(0, 1) },
    });
    function Probe({ season }: { season: number }) {
      const q = useProgrammes(season);
      return <p>{q.data ? `season ${season}: ${q.data.length}` : 'loading'}</p>;
    }
    renderWithProviders(
      <>
        <Probe season={0} />
        <Probe season={2} />
      </>,
    );
    expect(await screen.findByText(`season 0: ${programmes.length}`)).toBeInTheDocument();
    expect(await screen.findByText('season 2: 1')).toBeInTheDocument();
    const urls = fetchMock.mock.calls.map(([u]) => String(u));
    expect(urls).toContain('/api/v1/programmes');
    expect(urls).toContain('/api/v1/programmes?season=2');
  });

  it('useTeams caches signed-in and anonymous lists separately', async () => {
    mockFetch({ '/api/v1/auth/me': editor, '/api/v1/teams': { body: teams } });
    function Probe() {
      const q = useTeams();
      return <p>{q.data ? `${q.data.length} teams` : 'loading'}</p>;
    }
    const { queryClient } = renderWithProviders(<Probe />);
    expect(await screen.findByText(`${teams.length} teams`)).toBeInTheDocument();
    await waitFor(() => expect(queryClient.getQueryData(queryKeys.teams(true))).toEqual(teams));
  });

  it('detail hooks do not fetch without an id', async () => {
    const fetchMock = mockFetch({ '/api/v1/auth/me': anonymous });
    function Probe() {
      const q = useNewsArticle(null);
      return <p>{q.fetchStatus}</p>;
    }
    renderWithProviders(<Probe />);
    expect(await screen.findByText('idle')).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([u]) => String(u).startsWith('/api/v1/news'))).toBe(false);
  });
});
