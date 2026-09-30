import { fireEvent, screen, waitFor } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import { editorFor, typeInEditor } from '../../test/editor';
import { editor, info, manager, publicRoutes } from '../../test/fixtures';
import { Location } from '../../test/Location';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import InfoEditPage from './InfoEditPage';
import InfoPage from './InfoPage';

function renderInfo(route: string, overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(
    <>
      <Routes>
        <Route path="/info" element={<InfoPage />} />
        <Route path="/info/edit" element={<InfoEditPage />} />
      </Routes>
      <Location />
    </>,
    { route },
  );
  return fetchMock;
}

describe('Info editing', () => {
  it('links editors to the in-app editor', async () => {
    renderInfo('/info');
    expect(await screen.findByRole('link', { name: 'Edit' })).toHaveAttribute('href', '/info/edit');
  });

  it('hides Edit from a Manager', async () => {
    renderInfo('/info', { '/api/v1/auth/me': manager });
    await screen.findByRole('heading', { level: 2, name: 'Welcome' });
    expect(screen.queryByRole('link', { name: 'Edit' })).toBeNull();
  });

  it('starts from the stored content and saves it with PUT', async () => {
    const fetchMock = renderInfo('/info/edit', { '/api/v1/info': () => ({ body: info }) });
    expect((await editorFor('Club information')).getHTML()).toContain('Welcome');
    await typeInEditor('Club information', '<p>Updated</p>');
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    await waitFor(() => expect(screen.getByTestId('location')).toHaveTextContent(/^\/info$/));
    const put = fetchMock.mock.calls.find(([, i]) => i?.method === 'PUT');
    expect(JSON.parse(String(put?.[1]?.body))).toEqual({ content: '<p>Updated</p>' });
    expect(screen.getByRole('button', { name: 'Club information saved' })).toBeInTheDocument();
  });

  it('starts from the club history when nothing is stored', async () => {
    renderInfo('/info/edit', { '/api/v1/info': { body: { content: '' } } });
    const html = (await editorFor('Club information')).getHTML();
    expect(html).toContain('Club history');
    expect(html).toContain('AWRE Football Club');
  });
});
