import { useState, type FormEvent } from 'react';

import { createAffiliation } from '../../api/home';
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

export function AddAffiliationDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const toast = useToast();
  const [name, setName] = useState('');
  const [website, setWebsite] = useState('');
  const [image, setImage] = useState<ImageValue>(emptyImage);
  const [missing, setMissing] = useState<{ name?: string; image?: string }>({});
  const save = useSaveForm({
    submit: () => createAffiliation({ name, website, image: image.file as File }),
    invalidate: [queryKeys.home],
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Affiliation added' });
      setName('');
      setWebsite('');
      setImage(emptyImage);
      onClose();
    },
  });

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const next = {
      name: name.trim() ? undefined : 'Enter a name',
      image: image.file ? undefined : 'Choose a logo image',
    };
    setMissing(next);
    if (!next.name && !next.image) {
      void save.run();
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Add affiliation">
      <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
        {save.formError && <Alert tone="error">{save.formError}</Alert>}
        <Field label="Name" error={missing.name ?? save.fieldErrors.name}>
          <Input value={name} onChange={(e) => setName(e.target.value)} />
        </Field>
        <ImageField
          label="Logo"
          value={image}
          onChange={setImage}
          error={missing.image ?? save.fieldErrors.file ?? save.fieldErrors.image}
        />
        <Field label="Website" error={save.fieldErrors.website}>
          <Input type="url" value={website} onChange={(e) => setWebsite(e.target.value)} />
        </Field>
        <div className="flex justify-end gap-2">
          <Button variant="secondary" onClick={onClose} disabled={save.busy}>
            Cancel
          </Button>
          <Button type="submit" loading={save.busy}>
            Add affiliation
          </Button>
        </div>
      </form>
    </Modal>
  );
}
