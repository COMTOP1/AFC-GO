import { fireEvent, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { editor, galleryImages, manager, photographer, publicRoutes } from '../../test/fixtures';
import type { MockResponse, MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import GalleryPage from './GalleryPage';

beforeEach(() => {
  Object.assign(URL, { createObjectURL: vi.fn(() => 'blob:x'), revokeObjectURL: vi.fn() });
});
afterEach(() => {
  delete (URL as unknown as Record<string, unknown>).createObjectURL;
  delete (URL as unknown as Record<string, unknown>).revokeObjectURL;
});

function renderGallery(me: MockResponse, overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': me, ...overrides }));
  renderWithProviders(<GalleryPage />);
  return fetchMock;
}

describe('Gallery editing', () => {
  it('hides the controls from a Manager', async () => {
    renderGallery(manager);
    await screen.findByRole('button', { name: 'Cup final' });
    expect(screen.queryByRole('button', { name: 'Add photo' })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Delete photo Cup final' })).toBeNull();
  });

  it('lets a photographer add a photo with a caption', async () => {
    const fetchMock = renderGallery(photographer, {
      '/api/v1/gallery': (init) =>
        init?.method === 'POST' ? { status: 201, body: galleryImages[0] } : { body: galleryImages },
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Add photo' }));
    const dialog = screen.getByRole('dialog', { name: 'Add photo' });
    fireEvent.change(within(dialog).getByLabelText('Photo'), {
      target: { files: [new File(['x'], 'p.png', { type: 'image/png' })] },
    });
    fireEvent.change(within(dialog).getByLabelText('Caption'), { target: { value: 'Awards' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add photo' }));
    expect(await screen.findByRole('button', { name: 'Photo added' })).toBeInTheDocument();
    const fd = fetchMock.mock.calls.find(([, i]) => i?.method === 'POST')?.[1]?.body as FormData;
    expect(fd.get('caption')).toBe('Awards');
  });

  it('requires a photo', async () => {
    renderGallery(editor);
    fireEvent.click(await screen.findByRole('button', { name: 'Add photo' }));
    const dialog = screen.getByRole('dialog', { name: 'Add photo' });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add photo' }));
    expect(within(dialog).getByLabelText('Photo')).toHaveAccessibleDescription('Choose a photo');
  });

  it('deletes a photo after confirming, from a control beside the thumbnail', async () => {
    const fetchMock = renderGallery(editor, {
      [`/api/v1/gallery/${galleryImages[0].id}`]: { status: 204 },
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Delete photo Cup final' }));
    fireEvent.click(
      within(screen.getByRole('dialog', { name: 'Delete this photo?' })).getByRole('button', {
        name: 'Delete',
      }),
    );
    expect(await screen.findByRole('button', { name: 'Photo deleted' })).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true);
  });
});
