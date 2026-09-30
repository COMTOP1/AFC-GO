import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';

/** sponsor.Public — GET /sponsors */
export interface Sponsor {
  id: number;
  name: string;
  website?: string;
  purpose?: string;
  team?: string;
  imageUrl?: string;
}

export function useSponsors() {
  return useQuery({
    queryKey: queryKeys.sponsors,
    queryFn: ({ signal }) => apiFetch<Sponsor[]>('/sponsors', { signal }),
  });
}
