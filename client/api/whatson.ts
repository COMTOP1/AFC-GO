import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';
import { formData } from '../lib/editForm';
import type { ImageValue } from '../lib/images';

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

export interface EventInput {
  title: string;
  content: string;
  /** YYYY-MM-DD */
  dateOfEvent: string;
  image: ImageValue;
}

function eventForm(input: EventInput, isUpdate: boolean): FormData {
  return formData({
    title: input.title,
    content: input.content,
    dateOfEvent: input.dateOfEvent,
    image: input.image.file,
    removeImage: isUpdate && input.image.remove ? true : undefined,
  });
}

export function createEvent(input: EventInput): Promise<WhatsOnEvent> {
  return apiFetch<WhatsOnEvent>('/whatson', { form: eventForm(input, false) });
}

export function updateEvent(id: number, input: EventInput): Promise<WhatsOnEvent> {
  return apiFetch<WhatsOnEvent>(`/whatson/${id}`, {
    method: 'PATCH',
    form: eventForm(input, true),
  });
}

export function deleteEvent(id: number): Promise<void> {
  return apiFetch<void>(`/whatson/${id}`, { method: 'DELETE' });
}
