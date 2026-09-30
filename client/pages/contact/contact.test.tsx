import { fireEvent, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { contact, publicRoutes } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import ContactPage from './ContactPage';

describe('ContactPage', () => {
  it('lists officials with their own email when no site email is set', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<ContactPage />);
    const sam = (await screen.findByRole('heading', { name: 'Sam Sec' })).closest(
      'li',
    ) as HTMLElement;
    expect(within(sam).getByText('Club Secretary')).toBeInTheDocument();
    expect(within(sam).getByRole('link', { name: 'sam@example.test' })).toHaveAttribute(
      'href',
      'mailto:sam@example.test',
    );
    // No photo → crest. Decorative (alt=""), so select the element directly.
    expect(sam.querySelector('img')).toHaveAttribute('src', expect.stringContaining('crest'));
    expect(document.title).toBe('Contact · AFC Aldermaston');
  });

  it('uses the site-wide email for everyone when set', async () => {
    mockFetch(
      publicRoutes({
        '/api/v1/contact': { body: { ...contact, displayEmail: 'club@example.test' } },
      }),
    );
    renderWithProviders(<ContactPage />);
    await screen.findByRole('heading', { name: 'Sam Sec' });
    expect(screen.getAllByRole('link', { name: 'club@example.test' })).toHaveLength(2);
    expect(screen.queryByRole('link', { name: 'sam@example.test' })).toBeNull();
  });

  it('shows the satnav notice and a titled, lazy map', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<ContactPage />);
    await screen.findByRole('heading', { name: 'Sam Sec' });
    expect(screen.getByText(/use the postcode/)).toHaveTextContent(
      "If you're using a satnav, use the postcode RG26 4QP — the postcode listed takes you some distance away.",
    );
    const map = screen.getByTitle('Map to Aldermaston Recreational Society');
    expect(map.tagName).toBe('IFRAME');
    expect(map).toHaveAttribute('loading', 'lazy');
  });

  it('falls back to the crest when a contact photo is broken', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<ContactPage />);
    const cara = (await screen.findByRole('heading', { name: 'Cara Chair' })).closest(
      'li',
    ) as HTMLElement;
    fireEvent.error(cara.querySelector('img') as HTMLImageElement);
    expect(cara.querySelector('img')).toHaveAttribute('src', expect.stringContaining('crest'));
  });
});
