// The server's accepted image types (server/internal/upload).
export const IMAGE_TYPE_LIST = [
  'image/jpeg',
  'image/png',
  'image/gif',
  'image/webp',
  'image/avif',
  'image/apng',
  'image/svg+xml',
];
export const IMAGE_TYPES = IMAGE_TYPE_LIST.join(',');
export const IMAGE_TYPE_MESSAGE = 'Choose an image file (JPEG, PNG, GIF, WebP, AVIF, APNG or SVG).';

export function isAcceptedImage(file: File): boolean {
  return IMAGE_TYPE_LIST.includes(file.type);
}

/**
 * A local preview URL for a chosen image. Only a browser-issued blob: URL is
 * ever used as an <img> source; anything else is released and not shown.
 */
export function previewUrl(file: File): string | null {
  const url = URL.createObjectURL(file);
  if (url.startsWith('blob:')) {
    return url;
  }
  URL.revokeObjectURL(url);
  return null;
}

/** An image form field's value: a newly chosen file, and/or "remove the current one". */
export interface ImageValue {
  file: File | null;
  remove: boolean;
}

export const emptyImage: ImageValue = { file: null, remove: false };
