import { useQueryClient } from '@tanstack/react-query';
import { useEffect, useId, useRef, useState } from 'react';

import { removeAccountImage, uploadAccountImage } from '../../api/account';
import { queryKeys } from '../../api/queries';
import type { CurrentUser } from '../../api/types';
import crest from '../../assets/crest.png';
import { useAuth } from '../../auth/useAuth';
import { ImageWithFallback } from '../../components/page/ImageWithFallback';
import { Button } from '../../components/ui/Button';
import { Card, CardBody } from '../../components/ui/Card';
import { ConfirmDialog } from '../../components/ui/ConfirmDialog';
import { FileInput } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { fieldError } from '../../components/ui/fieldError';
import { useToast } from '../../components/ui/toast/useToast';

const IMAGE_TYPES = 'image/jpeg,image/png,image/gif,image/webp,image/avif,image/apng,image/svg+xml';

const photoClass = 'size-32 rounded-full border border-line bg-white object-cover';

export function PhotoCard({ user }: { user: CurrentUser }) {
  const headingId = useId();
  const queryClient = useQueryClient();
  const { refresh } = useAuth();
  const toast = useToast();
  const inputRef = useRef<HTMLInputElement>(null);
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const [error, setError] = useState<string | undefined>();
  const [saving, setSaving] = useState(false);
  const [confirmRemove, setConfirmRemove] = useState(false);

  // Release each preview URL when it is replaced, cleared, or the card unmounts.
  useEffect(() => {
    if (!preview) {
      return;
    }
    return () => URL.revokeObjectURL(preview);
  }, [preview]);

  function choose(next: File | null) {
    setError(undefined);
    setFile(next);
    setPreview(next ? URL.createObjectURL(next) : null);
    if (!next && inputRef.current) {
      inputRef.current.value = '';
    }
  }

  async function save() {
    if (!file) {
      return;
    }
    setSaving(true);
    setError(undefined);
    try {
      const updated = await uploadAccountImage(file);
      queryClient.setQueryData(queryKeys.me, updated);
      choose(null);
      toast.show({ tone: 'success', message: 'Photo updated' });
    } catch (err) {
      setError(
        fieldError(err, 'file') ??
          fieldError(err, 'image') ??
          (err instanceof Error ? err.message : 'Upload failed.'),
      );
    } finally {
      setSaving(false);
    }
  }

  async function remove() {
    try {
      await removeAccountImage();
      await refresh();
      toast.show({ tone: 'success', message: 'Photo removed' });
    } catch (err) {
      toast.show({
        tone: 'error',
        message: `Couldn't remove your photo: ${err instanceof Error ? err.message : 'unknown error'}`,
      });
    } finally {
      setConfirmRemove(false);
    }
  }

  return (
    <Card>
      <CardBody>
        <section aria-labelledby={headingId} className="space-y-4">
          <h2
            id={headingId}
            className="font-display text-2xl font-extrabold tracking-wide uppercase"
          >
            Photo
          </h2>
          {preview ? (
            <img src={preview} alt="Preview of your new photo" className={photoClass} />
          ) : (
            <ImageWithFallback
              src={user.imageUrl}
              fallbackSrc={crest}
              alt=""
              className={photoClass}
            />
          )}
          <Field label="Choose a new photo" error={error}>
            <FileInput
              ref={inputRef}
              accept={IMAGE_TYPES}
              onChange={(e) => choose(e.target.files?.[0] ?? null)}
            />
          </Field>
          <div className="flex flex-wrap gap-2">
            {file ? (
              <>
                <Button onClick={save} loading={saving}>
                  Save photo
                </Button>
                <Button variant="secondary" onClick={() => choose(null)} disabled={saving}>
                  Cancel
                </Button>
              </>
            ) : (
              user.imageUrl && (
                <Button variant="danger" onClick={() => setConfirmRemove(true)}>
                  Remove photo
                </Button>
              )
            )}
          </div>
        </section>
        <ConfirmDialog
          open={confirmRemove}
          title="Remove your photo?"
          message="Your photo will be replaced by the club crest."
          confirmLabel="Remove photo"
          tone="danger"
          onConfirm={remove}
          onCancel={() => setConfirmRemove(false)}
        />
      </CardBody>
    </Card>
  );
}
