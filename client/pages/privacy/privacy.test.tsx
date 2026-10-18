import { fireEvent, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';

import { OPT_OUT_COOKIE, isVisitCountOptedOut } from '../../lib/visitCount';
import { publicRoutes, site } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import PrivacyPage from './PrivacyPage';

function clearOptOut() {
  document.cookie = `${OPT_OUT_COOKIE}=; Max-Age=0; Path=/`;
}

describe('PrivacyPage', () => {
  afterEach(clearOptOut);

  it('lists every cookie the site sets', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<PrivacyPage />);
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Privacy notice' }),
    ).toBeInTheDocument();
    for (const name of ['_csrf', 'afc_aldermaston_visited', OPT_OUT_COOKIE]) {
      expect(screen.getByRole('cell', { name })).toBeInTheDocument();
    }
    expect(
      screen.getByRole('cell', { name: '24 hours, or 31 days if you tick "Remember me"' }),
    ).toBeInTheDocument();
    expect(document.title).toBe('Privacy notice · AFC Aldermaston');
  });

  it('points requests at the site-wide email when set', async () => {
    mockFetch(
      publicRoutes({ '/api/v1/site': { body: { ...site, displayEmail: 'club@example.test' } } }),
    );
    renderWithProviders(<PrivacyPage />);
    expect((await screen.findAllByRole('link', { name: 'club@example.test' }))[0]).toHaveAttribute(
      'href',
      'mailto:club@example.test',
    );
  });

  it('falls back to the contact page without a site-wide email', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<PrivacyPage />);
    await screen.findByRole('heading', { level: 1, name: 'Privacy notice' });
    expect(screen.getAllByRole('link', { name: 'contact page' })[0]).toHaveAttribute(
      'href',
      '/contact',
    );
  });

  it('opts out of and back into the visitor count', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<PrivacyPage />);
    const box = await screen.findByRole('checkbox', { name: "Don't count my visits" });
    expect(box).not.toBeChecked();

    fireEvent.click(box);
    expect(box).toBeChecked();
    expect(isVisitCountOptedOut()).toBe(true);

    fireEvent.click(box);
    expect(box).not.toBeChecked();
    expect(isVisitCountOptedOut()).toBe(false);
  });
});
