import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';
import { formData } from '../lib/editForm';

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

export function createProgramme(input: {
  name: string;
  /** YYYY-MM-DD */
  date: string;
  seasonId: number | null;
  file: File;
}): Promise<Programme> {
  return apiFetch<Programme>('/programmes', {
    form: formData({
      name: input.name,
      date: input.date,
      seasonId: input.seasonId,
      file: input.file,
    }),
  });
}

export function deleteProgramme(id: number): Promise<void> {
  return apiFetch<void>(`/programmes/${id}`, { method: 'DELETE' });
}

export function createSeason(name: string): Promise<Season> {
  return apiFetch<Season>('/seasons', { json: { name } });
}

export function renameSeason(id: number, name: string): Promise<Season> {
  return apiFetch<Season>(`/seasons/${id}`, { method: 'PATCH', json: { name } });
}

export function deleteSeason(id: number): Promise<void> {
  return apiFetch<void>(`/seasons/${id}`, { method: 'DELETE' });
}
