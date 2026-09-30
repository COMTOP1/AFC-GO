import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';
import { formData } from '../lib/editForm';

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

export function createDocument(input: { name: string; file: File }): Promise<ClubDocument> {
  return apiFetch<ClubDocument>('/documents', { form: formData(input) });
}

export function deleteDocument(id: number): Promise<void> {
  return apiFetch<void>(`/documents/${id}`, { method: 'DELETE' });
}
