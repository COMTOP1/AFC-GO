import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';

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
