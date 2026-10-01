import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { publicRoutes } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import { renderPdf } from './pdfRenderer';
import { ProgrammePreview } from './ProgrammePreview';

// test/setup.ts stands in for the real (canvas-drawing) renderer.
const render = vi.mocked(renderPdf);

beforeEach(() => {
  mockFetch(publicRoutes());
});

describe('ProgrammePreview', () => {
  it('shows a loading message, then the drawn pages', async () => {
    let finish: () => void = () => {};
    render.mockImplementation((_url, container) => {
      container.append(document.createElement('canvas'));
      return new Promise<void>((resolve) => {
        finish = resolve;
      });
    });
    const { container } = renderWithProviders(
      <ProgrammePreview url="/api/v1/files/programme/16" name="vs Downton" />,
    );
    expect(screen.getByText('Loading programme…')).toBeInTheDocument();
    await waitFor(() => expect(render).toHaveBeenCalled());
    expect(render.mock.calls[0][0]).toBe('/api/v1/files/programme/16');
    finish();
    await waitFor(() => expect(screen.queryByText('Loading programme…')).toBeNull());
    expect(container.querySelector('canvas')).not.toBeNull();
    expect(screen.getByRole('group', { name: 'Preview of vs Downton' })).toBeInTheDocument();
  });

  it('offers a download when the PDF cannot be drawn', async () => {
    render.mockRejectedValue(new Error('bad pdf'));
    renderWithProviders(<ProgrammePreview url="/api/v1/files/programme/16" name="vs Downton" />);
    expect(await screen.findByText(/We couldn't preview this programme\./)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Download it' })).toHaveAttribute(
      'href',
      '/api/v1/files/programme/16',
    );
    expect(screen.queryByText('Loading programme…')).toBeNull();
  });
});
