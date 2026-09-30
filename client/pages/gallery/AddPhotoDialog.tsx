import { useState, type FormEvent } from 'react';

import { createPhoto } from '../../api/gallery';
import { queryKeys } from '../../api/queries';
import { ImageField } from '../../components/edit/ImageField';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Input } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';
import { emptyImage, type ImageValue } from '../../lib/images';

export function AddPhotoDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const toast = useToast();
  const [caption, setCaption] = useState('');
  const [image, setImage] = useState<ImageValue>(emptyImage);
  const [missing, setMissing] = useState<string | undefined>();
  const save = useSaveForm({
    submit: () => createPhoto({ caption, image: image.file as File }),
    invalidate: [queryKeys.gallery],
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Photo added' });
      setCaption('');
      setImage(emptyImage);
      onClose();
    },
  });

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!image.file) {
      setMissing('Choose a photo');
      return;
    }
    setMissing(undefined);
    void save.run();
  }

  return (
    <Modal open={open} onClose={onClose} title="Add photo">
      <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
        {save.formError && <Alert tone="error">{save.formError}</Alert>}
        <ImageField
          label="Photo"
          value={image}
          onChange={setImage}
          error={missing ?? save.fieldErrors.file ?? save.fieldErrors.image}
        />
        <Field label="Caption" error={save.fieldErrors.caption}>
          <Input value={caption} onChange={(e) => setCaption(e.target.value)} />
        </Field>
        <div className="flex justify-end gap-2">
          <Button variant="secondary" onClick={onClose} disabled={save.busy}>
            Cancel
          </Button>
          <Button type="submit" loading={save.busy}>
            Add photo
          </Button>
        </div>
      </form>
    </Modal>
  );
}
