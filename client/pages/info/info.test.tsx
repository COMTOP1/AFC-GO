import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { editor, publicRoutes } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import InfoPage from './InfoPage';

describe('InfoPage', () => {
  it('shows the stored info, cleaned', async () => {
    mockFetch(
      publicRoutes({
        '/api/v1/info': { body: { content: '<h2>Welcome</h2><p>Hi<script>x()</script></p>' } },
      }),
    );
    const { container } = renderWithProviders(<InfoPage />);
    expect(await screen.findByRole('heading', { level: 2, name: 'Welcome' })).toBeInTheDocument();
    expect(container.querySelector('script')).toBeNull();
    expect(document.title).toBe('Information · AFC Aldermaston');
  });

  it('falls back to the club history when nothing is stored', async () => {
    mockFetch(publicRoutes({ '/api/v1/info': { body: { content: '  ' } } }));
    renderWithProviders(<InfoPage />);
    expect(await screen.findByRole('heading', { name: 'Club history' })).toBeInTheDocument();
    expect(screen.getByText(/founded as AWRE Football Club in 1952/)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'secretary@afcaldermaston.co.uk' })).toHaveAttribute(
      'href',
      'mailto:secretary@afcaldermaston.co.uk',
    );
  });

  it('links editors to the classic info editor', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': editor }));
    renderWithProviders(<InfoPage />);
    expect(
      await screen.findByRole('link', { name: 'Manage this on the classic site ↗' }),
    ).toHaveAttribute('href', '/info/edit');
  });
});
