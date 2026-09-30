import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';
import { formData } from '../lib/editForm';

/** image.Public — GET /gallery */
export interface GalleryImage {
  id: number;
  caption?: string;
  imageUrl: string;
}

export function useGallery() {
  return useQuery({
    queryKey: queryKeys.gallery,
    queryFn: ({ signal }) => apiFetch<GalleryImage[]>('/gallery', { signal }),
  });
}

export function createPhoto(input: { caption: string; image: File }): Promise<GalleryImage> {
  return apiFetch<GalleryImage>('/gallery', { form: formData(input) });
}

export function deletePhoto(id: number): Promise<void> {
  return apiFetch<void>(`/gallery/${id}`, { method: 'DELETE' });
}
