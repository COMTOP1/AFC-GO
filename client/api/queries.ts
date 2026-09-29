import { useQuery } from '@tanstack/react-query';

import { ApiError, apiFetch } from './client';
import type { CurrentUser, SiteInfo } from './types';

export const queryKeys = {
  site: ['site'] as const,
  me: ['auth', 'me'] as const,
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
