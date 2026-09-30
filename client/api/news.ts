import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';
import { formData } from '../lib/editForm';
import type { ImageValue } from '../lib/images';

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

export interface NewsInput {
  title: string;
  content: string;
  image: ImageValue;
}

function newsForm(input: NewsInput, isUpdate: boolean): FormData {
  return formData({
    title: input.title,
    content: input.content,
    image: input.image.file,
    removeImage: isUpdate && input.image.remove ? true : undefined,
  });
}

export function createNews(input: NewsInput): Promise<NewsArticle> {
  return apiFetch<NewsArticle>('/news', { form: newsForm(input, false) });
}

export function updateNews(id: number, input: NewsInput): Promise<NewsArticle> {
  return apiFetch<NewsArticle>(`/news/${id}`, { method: 'PATCH', form: newsForm(input, true) });
}

export function deleteNews(id: number): Promise<void> {
  return apiFetch<void>(`/news/${id}`, { method: 'DELETE' });
}
