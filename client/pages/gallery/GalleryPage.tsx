import { useState } from 'react';

import { deletePhoto, useGallery } from '../../api/gallery';
import { queryKeys } from '../../api/queries';
import { DeleteButton } from '../../components/edit/DeleteButton';
import { useCanEdit } from '../../components/edit/useCanEdit';
import { ImageWithFallback } from '../../components/page/ImageWithFallback';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Button } from '../../components/ui/Button';
import { PageHeader } from '../../components/ui/PageHeader';
import { AddPhotoDialog } from './AddPhotoDialog';
import { Lightbox } from './Lightbox';

export default function GalleryPage() {
  usePageTitle('Gallery');
  const gallery = useGallery();
  const [openIndex, setOpenIndex] = useState<number | null>(null);
  const { canManageGallery } = useCanEdit();
  const [adding, setAdding] = useState(false);
  return (
    <>
      <PageHeader
        title="Gallery"
        actions={canManageGallery && <Button onClick={() => setAdding(true)}>Add photo</Button>}
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
                  {canManageGallery && (
                    <div className="mt-1 flex justify-end">
                      <DeleteButton
                        ariaLabel={`Delete photo ${img.caption || i + 1}`}
                        confirmTitle="Delete this photo?"
                        confirmMessage="This can't be undone."
                        onDelete={() => deletePhoto(img.id)}
                        invalidate={[queryKeys.gallery]}
                        successMessage="Photo deleted"
                      />
                    </div>
                  )}
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
      <AddPhotoDialog open={adding} onClose={() => setAdding(false)} />
    </>
  );
}
