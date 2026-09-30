import { act, fireEvent, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { editor, galleryImages, manager, publicRoutes } from '../../test/fixtures';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import GalleryPage from './GalleryPage';

async function renderGallery(overrides: Record<string, MockRoute> = {}) {
  mockFetch(publicRoutes(overrides));
  renderWithProviders(<GalleryPage />);
  return screen.findByRole('button', { name: 'Cup final' });
}

function swipe(el: Element, fromX: number, toX: number) {
  const start = new Event('touchstart', { bubbles: true });
  Object.defineProperty(start, 'touches', { value: [{ clientX: fromX }] });
  const end = new Event('touchend', { bubbles: true });
  Object.defineProperty(end, 'changedTouches', { value: [{ clientX: toX }] });
  act(() => {
    el.dispatchEvent(start);
    el.dispatchEvent(end);
  });
}

describe('GalleryPage', () => {
  it('shows thumbnails named by caption, or by position', async () => {
    await renderGallery();
    expect(screen.getByRole('button', { name: 'Photo 2 of 3' })).toBeInTheDocument();
    expect(document.title).toBe('Gallery · AFC Aldermaston');
  });

  it('opens the viewer at the clicked photo with a counter and caption', async () => {
    const thumb = await renderGallery();
    fireEvent.click(thumb);
    const dialog = screen.getByRole('dialog');
    expect(within(dialog).getByText('1 / 3')).toBeInTheDocument();
    expect(within(dialog).getByRole('img', { name: 'Cup final' })).toHaveAttribute(
      'src',
      galleryImages[0].imageUrl,
    );
  });

  it('moves with the buttons and wraps around', async () => {
    fireEvent.click(await renderGallery());
    const dialog = screen.getByRole('dialog');
    fireEvent.click(within(dialog).getByRole('button', { name: 'Previous photo' }));
    expect(within(dialog).getByText('3 / 3')).toBeInTheDocument();
    fireEvent.click(within(dialog).getByRole('button', { name: 'Next photo' }));
    expect(within(dialog).getByText('1 / 3')).toBeInTheDocument();
  });

  it('moves with the arrow keys', async () => {
    fireEvent.click(await renderGallery());
    fireEvent.keyDown(document, { key: 'ArrowRight' });
    expect(screen.getByText('2 / 3')).toBeInTheDocument();
    fireEvent.keyDown(document, { key: 'ArrowLeft' });
    expect(screen.getByText('1 / 3')).toBeInTheDocument();
  });

  it('moves with a swipe of 50px or more, but not a small drag', async () => {
    fireEvent.click(await renderGallery());
    const stage = screen.getByTestId('lightbox-stage');
    swipe(stage, 300, 280);
    expect(screen.getByText('1 / 3')).toBeInTheDocument();
    swipe(stage, 300, 200);
    expect(screen.getByText('2 / 3')).toBeInTheDocument();
    swipe(stage, 100, 200);
    expect(screen.getByText('1 / 3')).toBeInTheDocument();
  });

  it('closes on Esc and returns focus to the thumbnail', async () => {
    const thumb = await renderGallery();
    thumb.focus();
    fireEvent.click(thumb);
    fireEvent(screen.getByRole('dialog'), new Event('cancel', { cancelable: true }));
    expect(screen.queryByText('1 / 3')).toBeNull();
    expect(document.activeElement).toBe(thumb);
  });

  it('closes with the Close button', async () => {
    fireEvent.click(await renderGallery());
    fireEvent.click(screen.getByRole('button', { name: 'Close' }));
    expect(screen.queryByText('1 / 3')).toBeNull();
  });

  it('shows the empty state', async () => {
    mockFetch(publicRoutes({ '/api/v1/gallery': { body: [] } }));
    renderWithProviders(<GalleryPage />);
    expect(await screen.findByText('No photos yet')).toBeInTheDocument();
  });

  it('shows Add photo to people who can manage the gallery', async () => {
    await renderGallery({ '/api/v1/auth/me': editor });
    expect(await screen.findByRole('button', { name: 'Add photo' })).toBeInTheDocument();
  });

  it('hides Add photo from managers', async () => {
    await renderGallery({ '/api/v1/auth/me': manager });
    await new Promise((r) => setTimeout(r, 0));
    expect(screen.queryByRole('button', { name: 'Add photo' })).toBeNull();
  });

  it('shows a placeholder for a broken thumbnail and in the viewer', async () => {
    const thumb = await renderGallery();
    fireEvent.error(thumb.querySelector('img') as HTMLImageElement);
    expect(thumb.querySelector('img')).toBeNull();
    expect(thumb.querySelector('[data-fallback]')).not.toBeNull();
    fireEvent.click(thumb);
    const dialog = screen.getByRole('dialog');
    fireEvent.error(within(dialog).getByRole('img', { name: 'Cup final' }));
    expect(within(dialog).getByText("This photo couldn't be loaded.")).toBeInTheDocument();
  });
});
