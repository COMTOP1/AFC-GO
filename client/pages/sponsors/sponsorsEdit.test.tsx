import { fireEvent, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { editor, manager, publicRoutes, sponsor, team } from '../../test/fixtures';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import SponsorsPage from './SponsorsPage';

beforeEach(() => {
  Object.assign(URL, { createObjectURL: vi.fn(() => 'blob:x'), revokeObjectURL: vi.fn() });
});
afterEach(() => {
  delete (URL as unknown as Record<string, unknown>).createObjectURL;
  delete (URL as unknown as Record<string, unknown>).revokeObjectURL;
});

function renderSponsors(overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(<SponsorsPage />);
  return fetchMock;
}

describe('Sponsors editing', () => {
  it('hides the controls from a Manager', async () => {
    renderSponsors({ '/api/v1/auth/me': manager });
    await screen.findByRole('heading', { name: sponsor.name });
    expect(screen.queryByRole('button', { name: 'Add sponsor' })).toBeNull();
  });

  it('offers the team choices and sends the chosen team id', async () => {
    const fetchMock = renderSponsors({
      '/api/v1/sponsors': (init) =>
        init?.method === 'POST' ? { status: 201, body: sponsor } : { body: [sponsor] },
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Add sponsor' }));
    const dialog = screen.getByRole('dialog', { name: 'Add sponsor' });
    const select = within(dialog).getByLabelText('Sponsors');
    await within(dialog).findByRole('option', { name: team.name });
    expect(
      within(select)
        .getAllByRole('option')
        .map((o) => (o as HTMLOptionElement).value)
        .slice(0, 4),
    ).toEqual(['', 'A', 'O', 'Y']);
    fireEvent.change(within(dialog).getByLabelText('Name'), { target: { value: 'Corner Shop' } });
    fireEvent.change(within(dialog).getByLabelText('Logo'), {
      target: { files: [new File(['x'], 'l.png', { type: 'image/png' })] },
    });
    fireEvent.change(select, { target: { value: String(team.id) } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add sponsor' }));
    expect(await screen.findByRole('button', { name: 'Sponsor added' })).toBeInTheDocument();
    const fd = fetchMock.mock.calls.find(([, i]) => i?.method === 'POST')?.[1]?.body as FormData;
    expect(fd.get('team')).toBe(String(team.id));
    expect(fd.get('name')).toBe('Corner Shop');
  });

  it('requires a name and a logo', async () => {
    renderSponsors();
    fireEvent.click(await screen.findByRole('button', { name: 'Add sponsor' }));
    const dialog = screen.getByRole('dialog', { name: 'Add sponsor' });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add sponsor' }));
    expect(within(dialog).getByLabelText('Name')).toHaveAccessibleDescription('Enter a name');
    expect(within(dialog).getByLabelText('Logo')).toHaveAccessibleDescription(
      'Choose a logo image',
    );
  });

  it('deletes a sponsor after confirming', async () => {
    const fetchMock = renderSponsors({ [`/api/v1/sponsors/${sponsor.id}`]: { status: 204 } });
    fireEvent.click(await screen.findByRole('button', { name: `Delete ${sponsor.name}` }));
    fireEvent.click(
      within(screen.getByRole('dialog', { name: `Delete ${sponsor.name}?` })).getByRole('button', {
        name: 'Delete',
      }),
    );
    expect(await screen.findByRole('button', { name: 'Sponsor deleted' })).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true);
  });
});
