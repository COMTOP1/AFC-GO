import { clsx } from 'clsx';

import { ImageWithFallback } from './ImageWithFallback';

export interface FullImageProps {
  src?: string;
  alt: string;
  className?: string;
}

/**
 * An image shown in full at its own shape, never cropped. Very tall images are
 * capped and scaled down; a missing or broken one shows the club gradient.
 */
export function FullImage({ src, alt, className }: FullImageProps) {
  return (
    <div className={clsx('overflow-hidden rounded-lg border border-line bg-surface', className)}>
      <ImageWithFallback
        src={src}
        alt={alt}
        className="mx-auto block h-auto max-h-[70vh] w-auto max-w-full object-contain"
        fallback={
          <div
            aria-hidden="true"
            data-fallback=""
            className="w-full bg-linear-135 from-blue to-red"
            style={{ aspectRatio: '21 / 9' }}
          />
        }
      />
    </div>
  );
}
