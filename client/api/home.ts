import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import type { NewsArticle } from './news';
import { queryKeys } from './queries';
import type { Sponsor } from './sponsors';
import type { WhatsOnEvent } from './whatson';

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
