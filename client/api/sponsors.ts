import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';
import { formData } from '../lib/editForm';

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

export function createSponsor(input: {
  name: string;
  website: string;
  purpose: string;
  /** '', 'A', 'O', 'Y' or a team id */
  team: string;
  image: File;
}): Promise<Sponsor> {
  return apiFetch<Sponsor>('/sponsors', { form: formData(input) });
}

export function deleteSponsor(id: number): Promise<void> {
  return apiFetch<void>(`/sponsors/${id}`, { method: 'DELETE' });
}
