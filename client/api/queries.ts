import { useQuery } from '@tanstack/react-query';

import { ApiError, apiFetch } from './client';
import type { CurrentUser, SiteInfo } from './types';

export const queryKeys = {
  site: ['site'] as const,
  me: ['auth', 'me'] as const,
  home: ['home'] as const,
  // The server includes inactive teams for signed-in users.
  teams: (signedIn: boolean) => ['teams', { signedIn }] as const,
  team: (id: number) => ['team', id] as const,
  news: ['news'] as const,
  newsArticle: (id: number) => ['news', id] as const,
  whatson: (period: string) => ['whatson', period] as const,
  whatsonEvent: (id: number) => ['whatson', 'event', id] as const,
  gallery: ['gallery'] as const,
  documents: ['documents'] as const,
  programmes: (seasonId: number) => ['programmes', seasonId] as const,
  seasons: ['seasons'] as const,
  sponsors: ['sponsors'] as const,
  info: ['info'] as const,
  contact: ['contact'] as const,
  players: ['players'] as const,
  users: ['users'] as const,
};

export function useSite() {
  return useQuery({
    queryKey: queryKeys.site,
    queryFn: ({ signal }) => apiFetch<SiteInfo>('/site', { signal }),
  });
}

/** The signed-in user, or null when nobody is (a 401 is not an error here). */
export async function fetchMe(signal?: AbortSignal): Promise<CurrentUser | null> {
  try {
    return await apiFetch<CurrentUser>('/auth/me', { signal });
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      return null;
    }
    throw err;
  }
}

export function useMe() {
  return useQuery({
    queryKey: queryKeys.me,
    queryFn: ({ signal }) => fetchMe(signal),
  });
}
