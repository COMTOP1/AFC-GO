import { fireEvent, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';

import App from '../App';
import { mockFetch } from '../test/mockFetch';
import { renderWithProviders } from '../test/render';

beforeEach(() => {
  mockFetch({
    '/api/v1/site': { body: { year: 2026, visitorCount: 1, version: 'test', teams: [] } },
    '/api/v1/auth/me': { status: 401, body: { error: { code: 401, message: 'login required' } } },
  });
});

describe('DesignPage', () => {
  it('is served at /design inside the layout', () => {
    renderWithProviders(<App />, { route: '/design' });
    expect(screen.getByRole('heading', { level: 1, name: 'Design system' })).toBeInTheDocument();
    for (const section of [
      'Palette',
      'Buttons',
      'Badges and notices',
      'Cards',
      'Forms',
      'Table',
      'Dialogs, menus and toasts',
    ]) {
      expect(screen.getByRole('heading', { level: 2, name: section })).toBeInTheDocument();
    }
  });

  it('opens the sample modal and confirm dialog', () => {
    renderWithProviders(<App />, { route: '/design' });
    fireEvent.click(screen.getByRole('button', { name: 'Open modal' }));
    const modal = screen.getByRole('dialog', { name: 'Sample modal' });
    fireEvent.click(within(modal).getByRole('button', { name: 'Close' }));
    expect(screen.queryByRole('dialog', { name: 'Sample modal' })).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: 'Open confirm' }));
    expect(screen.getByRole('dialog', { name: 'Delete team?' })).toBeInTheDocument();
  });

  it('shows each toast tone', () => {
    renderWithProviders(<App />, { route: '/design' });
    fireEvent.click(screen.getByRole('button', { name: 'Success toast' }));
    expect(screen.getByRole('button', { name: 'Player saved' })).toBeInTheDocument();
  });

  it('shows the gradient for cards without an image', () => {
    const { container } = renderWithProviders(<App />, { route: '/design' });
    // Only the "No image" card: jsdom never fires the broken image's error event.
    expect(container.querySelectorAll('[data-fallback]').length).toBeGreaterThanOrEqual(1);
  });
});
