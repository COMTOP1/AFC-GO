import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';

/** setting.InfoContent — GET /info */
export interface InfoContent {
  content: string;
}

/** site.ContactPerson */
export interface ContactPerson {
  id: number;
  name: string;
  email: string;
  role: string;
  imageUrl?: string;
}

/** site.Contact — GET /contact */
export interface ContactData {
  displayEmail?: string;
  people: ContactPerson[];
}

export function useInfo() {
  return useQuery({
    queryKey: queryKeys.info,
    queryFn: ({ signal }) => apiFetch<InfoContent>('/info', { signal }),
  });
}

export function useContact() {
  return useQuery({
    queryKey: queryKeys.contact,
    queryFn: ({ signal }) => apiFetch<ContactData>('/contact', { signal }),
  });
}

export function setInfo(content: string): Promise<void> {
  return apiFetch<void>('/info', { method: 'PUT', json: { content } });
}
