import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';

/** news.Article — GET /news, /news/:id */
export interface NewsArticle {
  id: number;
  title: string;
  content: string;
  date: string;
  imageUrl?: string;
}

export function useNewsList() {
  return useQuery({
    queryKey: queryKeys.news,
    queryFn: ({ signal }) => apiFetch<NewsArticle[]>('/news', { signal }),
  });
}

export function useNewsArticle(id: number | null) {
  return useQuery({
    queryKey: queryKeys.newsArticle(id ?? 0),
    queryFn: ({ signal }) => apiFetch<NewsArticle>(`/news/${id}`, { signal }),
    enabled: id !== null,
  });
}
