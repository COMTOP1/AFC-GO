import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import { typeInEditor } from '../../test/editor';
import { editor, event, manager, publicRoutes } from '../../test/fixtures';
import { Location } from '../../test/Location';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import EventFormPage from './EventFormPage';
import EventPage from './EventPage';
import WhatsOnPage from './WhatsOnPage';

function renderEvents(route: string, overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(
    <>
      <Routes>
        <Route path="/whatson" element={<WhatsOnPage />} />
        <Route path="/whatson/new" element={<EventFormPage />} />
        <Route path="/whatson/:id" element={<EventPage />} />
        <Route path="/whatson/:id/edit" element={<EventFormPage />} />
      </Routes>
      <Location />
    </>,
    { route },
  );
  return fetchMock;
}

describe("What's On editing", () => {
  it('shows Add event to editors only', async () => {
    renderEvents('/whatson');
    expect(await screen.findByRole('link', { name: 'Add event' })).toHaveAttribute(
      'href',
      '/whatson/new',
    );
  });

  it('hides the controls from a Manager', async () => {
    renderEvents(`/whatson/${event.id}`, { '/api/v1/auth/me': manager });
    await screen.findByRole('heading', { level: 1, name: event.title });
    expect(screen.queryByRole('link', { name: 'Edit' })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Delete' })).toBeNull();
  });

  it('requires a title and a date', async () => {
    const fetchMock = renderEvents('/whatson/new');
    fireEvent.click(await screen.findByRole('button', { name: 'Save event' }));
    expect(screen.getByLabelText('Title')).toHaveAccessibleDescription('Enter a title');
    expect(screen.getByLabelText('Date of event')).toHaveAccessibleDescription(
      'Choose the date of the event',
    );
    expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(false);
  });

  it('creates an event with its date and opens it', async () => {
    const fetchMock = renderEvents('/whatson/new', {
      '/api/v1/whatson': { status: 201, body: { ...event, id: 77 } },
    });
    fireEvent.change(await screen.findByLabelText('Title'), { target: { value: 'Quiz night' } });
    fireEvent.change(screen.getByLabelText('Date of event'), { target: { value: '2026-11-07' } });
    await typeInEditor('Content', '<p>Teams of four</p>');
    fireEvent.click(screen.getByRole('button', { name: 'Save event' }));
    await waitFor(() => expect(screen.getByTestId('location')).toHaveTextContent('/whatson/77'));
    const fd = fetchMock.mock.calls.find(([, i]) => i?.method === 'POST')?.[1]?.body as FormData;
    expect(fd.get('dateOfEvent')).toBe('2026-11-07');
    expect(fd.get('title')).toBe('Quiz night');
  });

  it('loads the existing date when editing', async () => {
    renderEvents(`/whatson/${event.id}/edit`);
    expect(await screen.findByLabelText('Date of event')).toHaveValue('2026-10-16');
  });

  it('deletes an event after confirming', async () => {
    const fetchMock = renderEvents(`/whatson/${event.id}`, {
      [`/api/v1/whatson/${event.id}`]: () => ({ body: event }),
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Delete' }));
    fireEvent.click(
      within(screen.getByRole('dialog', { name: 'Delete this event?' })).getByRole('button', {
        name: 'Delete',
      }),
    );
    await waitFor(() => expect(screen.getByTestId('location')).toHaveTextContent(/^\/whatson$/));
    expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true);
  });
});
