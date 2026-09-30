import { fireEvent, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { affiliation, editor, home, manager, publicRoutes } from '../../test/fixtures';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import HomePage from './HomePage';

beforeEach(() => {
  Object.assign(URL, { createObjectURL: vi.fn(() => 'blob:x'), revokeObjectURL: vi.fn() });
});
afterEach(() => {
  delete (URL as unknown as Record<string, unknown>).createObjectURL;
  delete (URL as unknown as Record<string, unknown>).revokeObjectURL;
});

function renderHome(overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(<HomePage />);
  return fetchMock;
}

describe('Affiliations editing', () => {
  it('hides the controls from a Manager', async () => {
    renderHome({ '/api/v1/auth/me': manager });
    await screen.findByRole('region', { name: 'Affiliations' });
    expect(screen.queryByRole('button', { name: 'Add affiliation' })).toBeNull();
  });

  it('shows an empty Affiliations row to editors so they can add one', async () => {
    renderHome({ '/api/v1/home': { body: { ...home, affiliations: [] } } });
    const row = await screen.findByRole('region', { name: 'Affiliations' });
    expect(within(row).getByText('No affiliations yet')).toBeInTheDocument();
    expect(within(row).getByRole('button', { name: 'Add affiliation' })).toBeInTheDocument();
  });

  it('adds and deletes an affiliation', async () => {
    let n = 0;
    const fetchMock = renderHome({
      '/api/v1/affiliations': () => ({ status: 201, body: { id: 50, name: 'League' } }),
      [`/api/v1/affiliations/${affiliation.id}`]: () => (n++, { status: 204 }),
    });
    const row = await screen.findByRole('region', { name: 'Affiliations' });
    fireEvent.click(within(row).getByRole('button', { name: 'Add affiliation' }));
    const dialog = screen.getByRole('dialog', { name: 'Add affiliation' });
    fireEvent.change(within(dialog).getByLabelText('Name'), { target: { value: 'League' } });
    fireEvent.change(within(dialog).getByLabelText('Logo'), {
      target: { files: [new File(['x'], 'l.png', { type: 'image/png' })] },
    });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add affiliation' }));
    expect(await screen.findByRole('button', { name: 'Affiliation added' })).toBeInTheDocument();

    fireEvent.click(within(row).getByRole('button', { name: `Delete ${affiliation.name}` }));
    fireEvent.click(
      within(screen.getByRole('dialog', { name: `Delete ${affiliation.name}?` })).getByRole(
        'button',
        { name: 'Delete' },
      ),
    );
    expect(await screen.findByRole('button', { name: 'Affiliation deleted' })).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true);
    expect(n).toBe(1);
  });
});
