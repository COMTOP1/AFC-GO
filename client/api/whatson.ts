import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';

/** whatson.Event — GET /whatson, /whatson/:id */
export interface WhatsOnEvent {
  id: number;
  title: string;
  content: string;
  date: string;
  dateOfEvent: string;
  imageUrl?: string;
}

export type WhatsOnPeriod = 'future' | 'past' | 'all';

export function useWhatsOnList(period: WhatsOnPeriod) {
  return useQuery({
    queryKey: queryKeys.whatson(period),
    queryFn: ({ signal }) => apiFetch<WhatsOnEvent[]>(`/whatson?period=${period}`, { signal }),
  });
}

export function useWhatsOnEvent(id: number | null) {
  return useQuery({
    queryKey: queryKeys.whatsonEvent(id ?? 0),
    queryFn: ({ signal }) => apiFetch<WhatsOnEvent>(`/whatson/${id}`, { signal }),
    enabled: id !== null,
  });
}
