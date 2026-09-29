import { screen } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import type { NewsArticle } from '../../api/news';
import { editor, newsArticle, newsNoImage, publicRoutes } from '../../test/fixtures';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import NewsArticlePage from './NewsArticlePage';
import NewsListPage from './NewsListPage';

function renderArticle(route: string, overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes(overrides));
  renderWithProviders(
    <Routes>
      <Route path="/news/:id" element={<NewsArticlePage />} />
    </Routes>,
    { route },
  );
  return fetchMock;
}

describe('NewsListPage', () => {
  it('lists articles newest first with dates, falling back to the gradient', async () => {
    mockFetch(publicRoutes());
    const { container } = renderWithProviders(<NewsListPage />);
    const first = await screen.findByRole('link', { name: new RegExp(newsArticle.title) });
    expect(first).toHaveAttribute('href', `/news/${newsArticle.id}`);
    expect(screen.getByText('28 Sep 2026')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: new RegExp(newsNoImage.title) })).toBeInTheDocument();
    expect(container.querySelectorAll('[data-fallback]')).toHaveLength(1);
    expect(document.title).toBe('News · AFC Aldermaston');
  });

  it('pages long lists with Show more', async () => {
    const many: NewsArticle[] = Array.from({ length: 14 }, (_, i) => ({
      ...newsNoImage,
      id: 100 + i,
      title: `Story ${i + 1}`,
    }));
    mockFetch(publicRoutes({ '/api/v1/news': { body: many } }));
    renderWithProviders(<NewsListPage />);
    expect(await screen.findByRole('button', { name: 'Show more (2 more)' })).toBeInTheDocument();
    expect(screen.getAllByRole('link', { name: /Story/ })).toHaveLength(12);
  });

  it('shows the empty state and the editor link', async () => {
    mockFetch(publicRoutes({ '/api/v1/news': { body: [] }, '/api/v1/auth/me': editor }));
    renderWithProviders(<NewsListPage />);
    expect(await screen.findByText('No news yet')).toBeInTheDocument();
    expect(
      await screen.findByRole('link', { name: 'Manage this on the classic site ↗' }),
    ).toHaveAttribute('href', '/news');
  });
});

describe('NewsArticlePage', () => {
  it('renders the article with cleaned content', async () => {
    renderArticle(`/news/${newsArticle.id}`);
    expect(
      await screen.findByRole('heading', { level: 1, name: newsArticle.title }),
    ).toBeInTheDocument();
    expect(screen.getByText('Aldermaston').tagName).toBe('B');
    expect(screen.getByRole('link', { name: 'News' })).toHaveAttribute('href', '/news');
    expect(screen.getByText(/28 Sep 2026/)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: '← All news' })).toHaveAttribute('href', '/news');
    expect(document.title).toBe(`${newsArticle.title} · AFC Aldermaston`);
  });

  it('shows not found for a missing article', async () => {
    renderArticle('/news/404', {
      '/api/v1/news/404': { status: 404, body: { error: { code: 404, message: 'not found' } } },
    });
    expect(await screen.findByRole('heading', { name: 'Page not found' })).toBeInTheDocument();
  });

  it('shows not found for a malformed id without calling the API', async () => {
    const fetchMock = renderArticle('/news/abc');
    expect(await screen.findByRole('heading', { name: 'Page not found' })).toBeInTheDocument();
    expect(fetchMock.mock.calls.some(([u]) => String(u).startsWith('/api/v1/news'))).toBe(false);
  });
});
