import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import type { NewsArticle } from './news';
import { queryKeys } from './queries';
import type { Sponsor } from './sponsors';
import type { WhatsOnEvent } from './whatson';
import { formData } from '../lib/editForm';

/** affiliation.Public */
export interface Affiliation {
  id: number;
  name: string;
  website?: string;
  imageUrl?: string;
}

/** site.Home — GET /home. Panels whose data failed to load are omitted. */
export interface HomeData {
  latestNews?: NewsArticle;
  nextEvent?: WhatsOnEvent;
  sponsors: Sponsor[];
  affiliations: Affiliation[];
}

export function useHome() {
  return useQuery({
    queryKey: queryKeys.home,
    queryFn: ({ signal }) => apiFetch<HomeData>('/home', { signal }),
  });
}

export function createAffiliation(input: {
  name: string;
  website: string;
  image: File;
}): Promise<Affiliation> {
  return apiFetch<Affiliation>('/affiliations', { form: formData(input) });
}

export function deleteAffiliation(id: number): Promise<void> {
  return apiFetch<void>(`/affiliations/${id}`, { method: 'DELETE' });
}
