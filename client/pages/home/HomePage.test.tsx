import { fireEvent, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { event, home, newsArticle, publicRoutes } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import HomePage from './HomePage';

describe('HomePage', () => {
  it('shows the news hero, next event, sponsors and affiliations', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<HomePage />);
    expect(screen.getByRole('heading', { level: 1, name: 'AFC Aldermaston' })).toBeInTheDocument();
    const heroLink = await screen.findByRole('link', { name: newsArticle.title });
    expect(heroLink).toHaveAttribute('href', `/news/${newsArticle.id}`);
    expect(
      screen.getByText('A late winner at Aldermaston in front of a record crowd.'),
    ).toBeInTheDocument();
    expect(screen.getByRole('link', { name: event.title })).toHaveAttribute(
      'href',
      `/whatson/${event.id}`,
    );
    expect(screen.getByText('Fri 16 Oct 2026, 7pm')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'All events →' })).toHaveAttribute('href', '/whatson');

    const sponsors = screen.getByRole('region', { name: 'Our sponsors' });
    expect(within(sponsors).getByRole('link', { name: 'Acme Ltd' })).toHaveAttribute(
      'href',
      'https://acme.example',
    );
    expect(within(sponsors).getByText('Corner Shop')).toBeInTheDocument();
    expect(screen.getByRole('region', { name: 'Affiliations' })).toBeInTheDocument();
    expect(document.title).toBe('AFC Aldermaston');
  });

  it('lets the event span the row when there is no news', async () => {
    mockFetch(publicRoutes({ '/api/v1/home': { body: { ...home, latestNews: undefined } } }));
    renderWithProviders(<HomePage />);
    expect(await screen.findByRole('link', { name: event.title })).toBeInTheDocument();
    expect(screen.queryByText('Latest news')).toBeNull();
  });

  it('skips the hero row and empty logo rows when there is nothing to show', async () => {
    mockFetch(publicRoutes({ '/api/v1/home': { body: { sponsors: [], affiliations: [] } } }));
    renderWithProviders(<HomePage />);
    await screen.findByRole('heading', { level: 1 });
    await new Promise((r) => setTimeout(r, 0));
    expect(screen.queryByText('Latest news')).toBeNull();
    expect(screen.queryByText('Next event')).toBeNull();
    expect(screen.queryByRole('region', { name: 'Our sponsors' })).toBeNull();
    expect(screen.queryByRole('region', { name: 'Affiliations' })).toBeNull();
  });

  it('shows a retryable error when /home fails', async () => {
    mockFetch(
      publicRoutes({
        '/api/v1/home': { status: 500, body: { error: { code: 500, message: 'boom' } } },
      }),
    );
    renderWithProviders(<HomePage />);
    expect(await screen.findByRole('alert')).toHaveTextContent("Couldn't load this: boom");
    expect(screen.getByRole('button', { name: 'Retry' })).toBeInTheDocument();
  });

  it('falls back to the name tile when a logo image is broken', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<HomePage />);
    const sponsors = await screen.findByRole('region', { name: 'Our sponsors' });
    fireEvent.error(within(sponsors).getByRole('img', { name: 'Acme Ltd' }));
    expect(within(sponsors).queryByRole('img')).toBeNull();
    expect(within(sponsors).getByRole('link', { name: 'Acme Ltd' })).toHaveTextContent('Acme Ltd');
  });
});
