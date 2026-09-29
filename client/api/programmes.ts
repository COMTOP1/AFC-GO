import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';

/** programme.PublicSeason — GET /seasons */
export interface Season {
  id: number;
  name: string;
}

/** programme.Public — GET /programmes */
export interface Programme {
  id: number;
  name: string;
  date: string;
  fileUrl: string;
  season?: Season;
}

/** seasonId 0 means every season (the API's default when the param is absent). */
export function useProgrammes(seasonId: number, { enabled = true }: { enabled?: boolean } = {}) {
  return useQuery({
    queryKey: queryKeys.programmes(seasonId),
    queryFn: ({ signal }) =>
      apiFetch<Programme[]>(seasonId ? `/programmes?season=${seasonId}` : '/programmes', {
        signal,
      }),
    enabled,
  });
}

export function useSeasons() {
  return useQuery({
    queryKey: queryKeys.seasons,
    queryFn: ({ signal }) => apiFetch<Season[]>('/seasons', { signal }),
  });
}
