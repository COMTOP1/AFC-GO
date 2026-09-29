import { screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { publicRoutes } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import SponsorsPage from './SponsorsPage';

describe('SponsorsPage', () => {
  it('shows every sponsor with purpose, team and website', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<SponsorsPage />);
    const acme = (await screen.findByRole('heading', { name: 'Acme Ltd' })).closest('li');
    expect(acme).not.toBeNull();
    const card = within(acme as HTMLElement);
    expect(card.getByText('Kit sponsor')).toBeInTheDocument();
    expect(card.getByText('Sponsor of First Team')).toBeInTheDocument();
    expect(card.getByRole('link', { name: 'Acme Ltd website ↗' })).toHaveAttribute(
      'href',
      'https://acme.example',
    );
    // No logo: the name tile stands in, and no website link.
    const plain = within(
      screen.getByRole('heading', { name: 'Corner Shop' }).closest('li') as HTMLElement,
    );
    expect(plain.queryByRole('link')).toBeNull();
    expect(screen.queryByRole('button', { name: /Show more/ })).toBeNull();
    expect(document.title).toBe('Sponsors · AFC Aldermaston');
  });

  it('shows the empty state', async () => {
    mockFetch(publicRoutes({ '/api/v1/sponsors': { body: [] } }));
    renderWithProviders(<SponsorsPage />);
    expect(await screen.findByText('No sponsors yet')).toBeInTheDocument();
  });
});
