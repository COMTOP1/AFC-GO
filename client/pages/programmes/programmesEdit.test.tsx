import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import { editor, programmes, publicRoutes, seasons } from '../../test/fixtures';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import ProgrammesPage from './ProgrammesPage';

function renderProgrammes(overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(
    <Routes>
      <Route path="/programmes" element={<ProgrammesPage />} />
    </Routes>,
    { route: '/programmes' },
  );
  return fetchMock;
}

const pdf = () => new File(['x'], 'p.pdf', { type: 'application/pdf' });

describe('Programmes editing', () => {
  it('adds a programme without a season (no seasonId sent)', async () => {
    const fetchMock = renderProgrammes({
      '/api/v1/programmes': (init) =>
        init?.method === 'POST' ? { status: 201, body: programmes[0] } : { body: programmes },
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Add programme' }));
    const dialog = screen.getByRole('dialog', { name: 'Add programme' });
    fireEvent.change(within(dialog).getByLabelText('Name'), { target: { value: 'vs Z' } });
    fireEvent.change(within(dialog).getByLabelText('Date'), { target: { value: '2026-10-03' } });
    fireEvent.change(within(dialog).getByLabelText('File'), { target: { files: [pdf()] } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add programme' }));
    expect(await screen.findByRole('button', { name: 'Programme added' })).toBeInTheDocument();
    const post = fetchMock.mock.calls.find(
      ([u, i]) => i?.method === 'POST' && String(u).endsWith('/programmes'),
    );
    const fd = post?.[1]?.body as FormData;
    expect(fd.get('date')).toBe('2026-10-03');
    expect(fd.has('seasonId')).toBe(false);
  });

  it('adds a programme to a chosen season', async () => {
    const fetchMock = renderProgrammes({
      '/api/v1/programmes': (init) =>
        init?.method === 'POST' ? { status: 201, body: programmes[0] } : { body: programmes },
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Add programme' }));
    const dialog = screen.getByRole('dialog', { name: 'Add programme' });
    await within(dialog).findByRole('option', { name: seasons[1].name });
    fireEvent.change(within(dialog).getByLabelText('Name'), { target: { value: 'vs Z' } });
    fireEvent.change(within(dialog).getByLabelText('Date'), { target: { value: '2026-10-03' } });
    fireEvent.change(within(dialog).getByLabelText('Season'), {
      target: { value: String(seasons[1].id) },
    });
    fireEvent.change(within(dialog).getByLabelText('File'), { target: { files: [pdf()] } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add programme' }));
    await screen.findByRole('button', { name: 'Programme added' });
    const post = fetchMock.mock.calls.find(
      ([u, i]) => i?.method === 'POST' && String(u).endsWith('/programmes'),
    );
    expect((post?.[1]?.body as FormData).get('seasonId')).toBe(String(seasons[1].id));
  });

  it('manages seasons: add, rename, and delete with the unlink warning', async () => {
    const fetchMock = renderProgrammes({
      '/api/v1/seasons': (init) =>
        init?.method === 'POST'
          ? { status: 201, body: { id: 3, name: '2027-28' } }
          : { body: seasons },
      [`/api/v1/seasons/${seasons[0].id}`]: { body: { ...seasons[0], name: '2025/26' } },
      [`/api/v1/seasons/${seasons[1].id}`]: { status: 204 },
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Manage seasons' }));
    const dialog = screen.getByRole('dialog', { name: 'Seasons' });
    fireEvent.change(await within(dialog).findByLabelText('New season'), {
      target: { value: '2027-28' },
    });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add season' }));
    expect(await screen.findByRole('button', { name: 'Season added' })).toBeInTheDocument();
    expect(
      JSON.parse(
        String(
          fetchMock.mock.calls.find(
            ([u, i]) => i?.method === 'POST' && String(u).endsWith('/seasons'),
          )?.[1]?.body,
        ),
      ),
    ).toEqual({ name: '2027-28' });

    fireEvent.click(within(dialog).getByRole('button', { name: `Rename ${seasons[0].name}` }));
    const input = within(dialog).getByLabelText(`New name for ${seasons[0].name}`);
    fireEvent.change(input, { target: { value: '2025/26' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Save name' }));
    expect(await screen.findByRole('button', { name: 'Season renamed' })).toBeInTheDocument();

    fireEvent.click(within(dialog).getByRole('button', { name: `Delete ${seasons[1].name}` }));
    const confirm = screen.getByRole('dialog', { name: `Delete season ${seasons[1].name}?` });
    expect(confirm).toHaveTextContent('Its programmes will stay, with no season.');
    fireEvent.click(within(confirm).getByRole('button', { name: 'Delete' }));
    expect(await screen.findByRole('button', { name: 'Season deleted' })).toBeInTheDocument();
    await waitFor(() =>
      expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true),
    );
  });

  it('deletes a programme from its row', async () => {
    const fetchMock = renderProgrammes({
      [`/api/v1/programmes/${programmes[0].id}`]: { status: 204 },
    });
    fireEvent.click(await screen.findByRole('button', { name: `Delete ${programmes[0].name}` }));
    fireEvent.click(
      within(screen.getByRole('dialog', { name: `Delete ${programmes[0].name}?` })).getByRole(
        'button',
        { name: 'Delete' },
      ),
    );
    expect(await screen.findByRole('button', { name: 'Programme deleted' })).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true);
  });
});
