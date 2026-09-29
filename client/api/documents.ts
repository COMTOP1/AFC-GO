import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';

/** document.Public — GET /documents */
export interface ClubDocument {
  id: number;
  name: string;
  fileUrl: string;
}

export function useDocuments() {
  return useQuery({
    queryKey: queryKeys.documents,
    queryFn: ({ signal }) => apiFetch<ClubDocument[]>('/documents', { signal }),
  });
}
