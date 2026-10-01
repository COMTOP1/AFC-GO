import { fireEvent, screen, within } from '@testing-library/react';
import { Route, Routes, useLocation } from 'react-router';
import { describe, expect, it } from 'vitest';

import { programmes, publicRoutes } from '../../test/fixtures';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import ProgrammesPage from './ProgrammesPage';

function Location() {
  const l = useLocation();
  return <output data-testid="location">{l.pathname + l.search}</output>;
}

function renderProgrammes(route = '/programmes', overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(
    publicRoutes({ '/api/v1/programmes?season=2': { body: [programmes[0]] }, ...overrides }),
  );
  renderWithProviders(
    <Routes>
      <Route
        path="/programmes"
        element={
          <>
            <ProgrammesPage />
            <Location />
          </>
        }
      />
    </Routes>,
    { route },
  );
  return fetchMock;
}

describe('ProgrammesPage', () => {
  it('groups programmes by season with "No season" last', async () => {
    renderProgrammes();
    const headings = await screen.findAllByRole('heading', { level: 2 });
    expect(headings.map((h) => h.textContent)).toEqual([
      'Latest programme',
      '2026-27',
      '2025-26',
      'No season',
    ]);
    const group = screen.getByRole('region', { name: '2026-27' });
    expect(within(group).getByText('vs Downton')).toBeInTheDocument();
    expect(within(group).getByText('5 Sep 2026')).toBeInTheDocument();
    expect(within(group).getByRole('link', { name: 'View vs Downton' })).toHaveAttribute(
      'href',
      '/api/v1/files/programme/16',
    );
    expect(document.title).toBe('Programmes · AFC Aldermaston');
  });

  it('filters by season through the address', async () => {
    const fetchMock = renderProgrammes();
    const select = await screen.findByLabelText('Season');
    await screen.findByRole('option', { name: '2026-27' });
    fireEvent.change(select, { target: { value: '2' } });
    expect(await screen.findByTestId('location')).toHaveTextContent('/programmes?season=2');
    const group = await screen.findByRole('region', { name: '2026-27' });
    expect(within(group).getByText('vs Downton')).toBeInTheDocument();
    expect(screen.queryByText('vs Marlow')).toBeNull();
    expect(fetchMock.mock.calls.some(([u]) => String(u) === '/api/v1/programmes?season=2')).toBe(
      true,
    );
    fireEvent.change(select, { target: { value: '0' } });
    expect(screen.getByTestId('location')).toHaveTextContent(/^\/programmes$/);
  });

  it('searches by name', async () => {
    renderProgrammes('/programmes?q=marlow');
    expect(await screen.findByText('vs Marlow')).toBeInTheDocument();
    // The latest-programme card ignores search; the list below doesn't.
    expect(screen.queryByRole('region', { name: '2026-27' })).toBeNull();
  });

  it('says when nothing matches, and when there are none', async () => {
    renderProgrammes('/programmes?q=zzz');
    expect(await screen.findByText("No programmes match 'zzz'")).toBeInTheDocument();
  });

  it('treats an unknown ?season= as all seasons instead of erroring', async () => {
    const fetchMock = renderProgrammes('/programmes?season=99', {
      '/api/v1/programmes?season=99': {
        status: 404,
        body: { error: { code: 404, message: 'season not found' } },
      },
    });
    expect(await screen.findByText('vs Marlow')).toBeInTheDocument();
    expect(screen.queryByRole('alert')).toBeNull();
    expect(screen.getByLabelText('Season')).toHaveValue('0');
    expect(fetchMock.mock.calls.some(([u]) => String(u).includes('season=99'))).toBe(false);
  });

  it('previews the latest programme', async () => {
    renderProgrammes();
    const latest = await screen.findByRole('region', { name: 'Latest programme' });
    expect(within(latest).getByText('vs Downton')).toBeInTheDocument();
    expect(within(latest).getByText('Season 2026-27 · 5 Sep 2026')).toBeInTheDocument();
    expect(within(latest).getByRole('link', { name: 'View vs Downton PDF' })).toHaveAttribute(
      'href',
      '/api/v1/files/programme/16',
    );
    expect(
      within(latest).getByRole('group', { name: 'Preview of vs Downton' }),
    ).toBeInTheDocument();
  });

  it('previews the latest programme in the chosen season', async () => {
    renderProgrammes('/programmes?season=1', {
      '/api/v1/programmes?season=1': { body: [programmes[1]] },
    });
    const latest = await screen.findByRole('region', { name: 'Latest programme' });
    expect(within(latest).getByText('vs Marlow')).toBeInTheDocument();
  });

  it('shows the empty state when there are no programmes', async () => {
    renderProgrammes('/programmes', { '/api/v1/programmes': { body: [] } });
    expect(await screen.findByText('No programmes yet')).toBeInTheDocument();
    expect(screen.queryByRole('region', { name: 'Latest programme' })).toBeNull();
  });
});
