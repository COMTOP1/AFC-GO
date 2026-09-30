import { useState } from 'react';

import { useGallery } from '../../api/gallery';
import { EditorLink } from '../../components/page/EditorLink';
import { ImageWithFallback } from '../../components/page/ImageWithFallback';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { PageHeader } from '../../components/ui/PageHeader';
import { Lightbox } from './Lightbox';

export default function GalleryPage() {
  usePageTitle('Gallery');
  const gallery = useGallery();
  const [openIndex, setOpenIndex] = useState<number | null>(null);
  return (
    <>
      <PageHeader
        title="Gallery"
        actions={<EditorLink legacyHref="/gallery" permission="canManageGallery" />}
      />
      <QueryState query={gallery} isEmpty={(l) => l.length === 0} emptyTitle="No photos yet">
        {(images) => (
          <>
            <ul className="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-4">
              {images.map((img, i) => (
                <li key={img.id}>
                  <button
                    type="button"
                    aria-label={img.caption || `Photo ${i + 1} of ${images.length}`}
                    onClick={() => setOpenIndex(i)}
                    className="block aspect-square w-full overflow-hidden rounded-lg border border-line"
                  >
                    <ImageWithFallback
                      src={img.imageUrl}
                      alt=""
                      loading="lazy"
                      className="size-full object-cover transition-transform hover:scale-105"
                      fallback={
                        <div
                          aria-hidden="true"
                          data-fallback=""
                          className="size-full bg-linear-135 from-blue to-red"
                        />
                      }
                    />
                  </button>
                </li>
              ))}
            </ul>
            <Lightbox
              images={images}
              index={openIndex}
              onIndexChange={setOpenIndex}
              onClose={() => setOpenIndex(null)}
            />
          </>
        )}
      </QueryState>
    </>
  );
}
