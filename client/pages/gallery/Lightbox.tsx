import { useEffect, useRef } from 'react';

import type { GalleryImage } from '../../api/gallery';
import { Button } from '../../components/ui/Button';
import { Modal } from '../../components/ui/Modal';

export interface LightboxProps {
  images: GalleryImage[];
  /** The open photo, or null when closed. */
  index: number | null;
  onIndexChange: (index: number) => void;
  onClose: () => void;
}

const SWIPE_PX = 50;

export function Lightbox({ images, index, onIndexChange, onClose }: LightboxProps) {
  const touchStartX = useRef<number | null>(null);
  const count = images.length;

  useEffect(() => {
    if (index === null) {
      return;
    }
    function onKey(e: KeyboardEvent) {
      if (index === null) {
        return;
      }
      if (e.key === 'ArrowRight') {
        e.preventDefault();
        onIndexChange((index + 1) % count);
      } else if (e.key === 'ArrowLeft') {
        e.preventDefault();
        onIndexChange((index - 1 + count) % count);
      }
    }
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [index, count, onIndexChange]);

  const image = index !== null ? images[index] : undefined;
  const go = (delta: number) => {
    if (index !== null) {
      onIndexChange((index + delta + count) % count);
    }
  };

  return (
    <Modal
      open={index !== null}
      onClose={onClose}
      title="Gallery"
      className="w-[min(64rem,calc(100vw-2rem))]"
    >
      {image && index !== null && (
        <div
          data-testid="lightbox-stage"
          onTouchStart={(e) => {
            touchStartX.current = e.touches[0]?.clientX ?? null;
          }}
          onTouchEnd={(e) => {
            const start = touchStartX.current;
            touchStartX.current = null;
            const end = e.changedTouches[0]?.clientX;
            if (start === null || end === undefined) {
              return;
            }
            const dx = end - start;
            if (Math.abs(dx) >= SWIPE_PX) {
              go(dx < 0 ? 1 : -1);
            }
          }}
        >
          <figure>
            <img
              src={image.imageUrl}
              alt={image.caption ?? ''}
              className="mx-auto max-h-[70vh] w-auto rounded-md object-contain"
            />
            {image.caption && (
              <figcaption className="mt-2 text-center text-sm">{image.caption}</figcaption>
            )}
          </figure>
          <div className="mt-4 flex items-center justify-between gap-2">
            <Button
              variant="secondary"
              size="sm"
              aria-label="Previous photo"
              onClick={() => go(-1)}
            >
              ← Previous
            </Button>
            <span aria-live="polite" className="text-sm text-muted">
              {index + 1} / {count}
            </span>
            <div className="flex gap-2">
              <Button variant="secondary" size="sm" aria-label="Next photo" onClick={() => go(1)}>
                Next →
              </Button>
              <Button variant="ghost" size="sm" onClick={onClose}>
                Close
              </Button>
            </div>
          </div>
        </div>
      )}
    </Modal>
  );
}
