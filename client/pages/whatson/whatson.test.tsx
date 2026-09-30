import { fireEvent, screen } from '@testing-library/react';
import { Route, Routes, useNavigate } from 'react-router';
import { describe, expect, it } from 'vitest';

import type { WhatsOnEvent } from '../../api/whatson';
import { event, publicRoutes } from '../../test/fixtures';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import EventPage from './EventPage';
import WhatsOnPage from './WhatsOnPage';

function BackButton() {
  const navigate = useNavigate();
  return <button onClick={() => navigate(-1)}>Go back</button>;
}

function renderList(route = '/whatson', overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes(overrides));
  renderWithProviders(
    <Routes>
      <Route
        path="/whatson"
        element={
          <>
            <WhatsOnPage />
            <BackButton />
          </>
        }
      />
    </Routes>,
    { route },
  );
  return fetchMock;
}

describe('WhatsOnPage', () => {
  it('defaults to upcoming events', async () => {
    const fetchMock = renderList();
    expect(await screen.findByRole('link', { name: new RegExp(event.title) })).toHaveAttribute(
      'href',
      `/whatson/${event.id}`,
    );
    expect(screen.getByText('Fri 16 Oct 2026, 7pm')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Upcoming' })).toHaveAttribute('aria-current', 'page');
    expect(fetchMock.mock.calls.some(([u]) => String(u) === '/api/v1/whatson?period=future')).toBe(
      true,
    );
    expect(document.title).toBe("What's On · AFC Aldermaston");
  });

  it('switches tabs, requests that period and Back restores the tab', async () => {
    const fetchMock = renderList();
    await screen.findByRole('link', { name: new RegExp(event.title) });
    fireEvent.click(screen.getByRole('link', { name: 'Past' }));
    expect(await screen.findByText('No past events')).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([u]) => String(u) === '/api/v1/whatson?period=past')).toBe(
      true,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Go back' }));
    expect(await screen.findByRole('link', { name: new RegExp(event.title) })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Upcoming' })).toHaveAttribute('aria-current', 'page');
  });

  it('goes back to the first page of cards when the tab changes', async () => {
    const many: WhatsOnEvent[] = Array.from({ length: 20 }, (_, i) => ({
      ...event,
      id: 200 + i,
      title: `Event ${i + 1}`,
    }));
    renderList('/whatson?period=all', {
      '/api/v1/whatson?period=all': { body: many },
      '/api/v1/whatson?period=future': { body: many },
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Show more (8 more)' }));
    expect(screen.getAllByRole('link', { name: /^Event / })).toHaveLength(20);
    fireEvent.click(screen.getByRole('link', { name: 'Upcoming' }));
    expect(await screen.findByRole('button', { name: 'Show more (8 more)' })).toBeInTheDocument();
    expect(screen.getAllByRole('link', { name: /^Event / })).toHaveLength(12);
  });

  it('uses a period-specific empty title', async () => {
    renderList('/whatson?period=all', { '/api/v1/whatson?period=all': { body: [] } });
    expect(await screen.findByText('No events yet')).toBeInTheDocument();
  });
});

describe('EventPage', () => {
  function renderEvent(route: string, overrides: Record<string, MockRoute> = {}) {
    const fetchMock = mockFetch(publicRoutes(overrides));
    renderWithProviders(
      <Routes>
        <Route path="/whatson/:id" element={<EventPage />} />
      </Routes>,
      { route },
    );
    return fetchMock;
  }

  it('renders the event with its date and time', async () => {
    renderEvent(`/whatson/${event.id}`);
    expect(await screen.findByRole('heading', { level: 1, name: event.title })).toBeInTheDocument();
    expect(screen.getAllByText('Fri 16 Oct 2026, 7pm').length).toBeGreaterThan(0);
    expect(screen.getByText('Clubhouse, all welcome.')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: '← All events' })).toHaveAttribute('href', '/whatson');
  });

  it('shows the event image in full, at its own shape', async () => {
    renderEvent(`/whatson/${event.id}`, {
      [`/api/v1/whatson/${event.id}`]: { body: { ...event, imageUrl: '/api/v1/files/whatson/3' } },
    });
    await screen.findByRole('heading', { level: 1, name: event.title });
    const img = document.querySelector('article img') as HTMLImageElement;
    expect(img).toHaveAttribute('src', '/api/v1/files/whatson/3');
    expect(img.style.aspectRatio).toBe('');
    expect(img).toHaveClass('object-contain');
  });

  it('shows not found for a missing or malformed event', async () => {
    renderEvent('/whatson/404', {
      '/api/v1/whatson/404': { status: 404, body: { error: { code: 404, message: 'not found' } } },
    });
    expect(await screen.findByRole('heading', { name: 'Page not found' })).toBeInTheDocument();
  });
});
