import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import { typeInEditor } from '../../test/editor';
import { anonymous, editor, manager, newsArticle, publicRoutes } from '../../test/fixtures';
import { Location } from '../../test/Location';
import type { MockResponse, MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import NewsArticlePage from './NewsArticlePage';
import NewsFormPage from './NewsFormPage';
import NewsListPage from './NewsListPage';

function renderNews(route: string, overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(
    <>
      <Routes>
        <Route path="/news" element={<NewsListPage />} />
        <Route path="/news/new" element={<NewsFormPage />} />
        <Route path="/news/:id" element={<NewsArticlePage />} />
        <Route path="/news/:id/edit" element={<NewsFormPage />} />
      </Routes>
      <Location />
    </>,
    { route },
  );
  return fetchMock;
}

const saved = { ...newsArticle, id: 42, title: 'New title' };

describe('News editing controls', () => {
  it.each([[anonymous], [manager]] as [MockResponse][])(
    'are hidden from non-editors',
    async (me) => {
      renderNews('/news', { '/api/v1/auth/me': me });
      await screen.findByRole('link', { name: new RegExp(newsArticle.title) });
      expect(screen.queryByRole('link', { name: 'Add article' })).toBeNull();
    },
  );

  it('show Add on the list and Edit/Delete on an article for editors', async () => {
    renderNews(`/news/${newsArticle.id}`);
    expect(await screen.findByRole('link', { name: 'Edit' })).toHaveAttribute(
      'href',
      `/news/${newsArticle.id}/edit`,
    );
    expect(screen.getByRole('button', { name: 'Delete' })).toBeInTheDocument();
    expect(screen.queryByText(/classic site/)).toBeNull();
  });

  it('delete confirms, calls DELETE and returns to the list', async () => {
    const fetchMock = renderNews(`/news/${newsArticle.id}`, {
      [`/api/v1/news/${newsArticle.id}`]: () => ({ body: newsArticle }),
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Delete' }));
    const dialog = screen.getByRole('dialog', { name: 'Delete this article?' });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Delete' }));
    await waitFor(() => expect(screen.getByTestId('location')).toHaveTextContent(/^\/news$/));
    expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'DELETE')).toBe(true);
    expect(screen.getByRole('button', { name: 'Article deleted' })).toBeInTheDocument();
  });
});

describe('NewsFormPage', () => {
  it('shows the permission message to non-editors', async () => {
    renderNews('/news/new', { '/api/v1/auth/me': manager });
    expect(await screen.findByText("You don't have permission to edit this")).toBeInTheDocument();
  });

  it('asks for a title without sending anything', async () => {
    const fetchMock = renderNews('/news/new');
    fireEvent.click(await screen.findByRole('button', { name: 'Save article' }));
    expect(screen.getByLabelText('Title')).toHaveAccessibleDescription('Enter a title');
    expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(false);
  });

  it('creates an article and opens it', async () => {
    const fetchMock = renderNews('/news/new', { '/api/v1/news': { status: 201, body: saved } });
    fireEvent.change(await screen.findByLabelText('Title'), { target: { value: 'New title' } });
    await typeInEditor('Content', '<p>Body</p>');
    fireEvent.click(screen.getByRole('button', { name: 'Save article' }));
    await waitFor(() => expect(screen.getByTestId('location')).toHaveTextContent('/news/42'));
    const post = fetchMock.mock.calls.find(([, init]) => init?.method === 'POST');
    const fd = post?.[1]?.body as FormData;
    expect(fd.get('title')).toBe('New title');
    expect(fd.get('content')).toBe('<p>Body</p>');
    expect(await screen.findByRole('button', { name: 'Article saved' })).toBeInTheDocument();
  });

  it('loads an article for editing and can remove its image', async () => {
    const fetchMock = renderNews(`/news/${newsArticle.id}/edit`, {
      [`/api/v1/news/${newsArticle.id}`]: () => ({ body: newsArticle }),
    });
    expect(await screen.findByLabelText('Title')).toHaveValue(newsArticle.title);
    fireEvent.click(screen.getByRole('checkbox', { name: 'Remove image' }));
    fireEvent.click(screen.getByRole('button', { name: 'Save article' }));
    await waitFor(() =>
      expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'PATCH')).toBe(true),
    );
    const patch = fetchMock.mock.calls.find(([, init]) => init?.method === 'PATCH');
    const fd = patch?.[1]?.body as FormData;
    expect(fd.get('removeImage')).toBe('true');
    expect(fd.get('content')).toBe(newsArticle.content);
  });

  it('shows server field errors', async () => {
    renderNews('/news/new', {
      '/api/v1/news': {
        status: 422,
        body: { error: { code: 422, message: 'invalid', fields: { title: 'title is too long' } } },
      },
    });
    fireEvent.change(await screen.findByLabelText('Title'), { target: { value: 'x' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save article' }));
    await waitFor(() =>
      expect(screen.getByLabelText('Title')).toHaveAccessibleDescription('title is too long'),
    );
  });
});
