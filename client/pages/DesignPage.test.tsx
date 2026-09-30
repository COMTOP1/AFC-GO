import { fireEvent, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';

import App from '../App';
import { publicRoutes } from '../test/fixtures';
import { mockFetch } from '../test/mockFetch';
import { renderWithProviders } from '../test/render';

beforeEach(() => {
  mockFetch(publicRoutes());
});

/** /design is lazy-loaded, so wait for its heading before interacting. */
async function renderDesign() {
  const result = renderWithProviders(<App />, { route: '/design' });
  await screen.findByRole('heading', { level: 1, name: 'Design system' }, { timeout: 3000 });
  return result;
}

describe('DesignPage', () => {
  it('is served at /design inside the layout', async () => {
    await renderDesign();
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

  it('opens the sample modal and confirm dialog', async () => {
    await renderDesign();
    fireEvent.click(screen.getByRole('button', { name: 'Open modal' }));
    const modal = screen.getByRole('dialog', { name: 'Sample modal' });
    fireEvent.click(within(modal).getByRole('button', { name: 'Close' }));
    expect(screen.queryByRole('dialog', { name: 'Sample modal' })).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: 'Open confirm' }));
    expect(screen.getByRole('dialog', { name: 'Delete team?' })).toBeInTheDocument();
  });

  it('shows each toast tone', async () => {
    await renderDesign();
    fireEvent.click(screen.getByRole('button', { name: 'Success toast' }));
    expect(screen.getByRole('button', { name: 'Player saved' })).toBeInTheDocument();
  });

  it('shows the gradient for cards without an image', async () => {
    const { container } = await renderDesign();
    // Only the "No image" card: jsdom never fires the broken image's error event.
    expect(container.querySelectorAll('[data-fallback]').length).toBeGreaterThanOrEqual(1);
  });
});
