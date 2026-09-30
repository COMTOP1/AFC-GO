import { useState, type FormEvent } from 'react';

import { createDocument } from '../../api/documents';
import { queryKeys } from '../../api/queries';
import { FileField } from '../../components/edit/FileField';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Input } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';

export function AddDocumentDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const toast = useToast();
  const [name, setName] = useState('');
  const [file, setFile] = useState<File | null>(null);
  const [missing, setMissing] = useState<{ name?: string; file?: string }>({});
  const save = useSaveForm({
    submit: () => createDocument({ name, file: file as File }),
    invalidate: [queryKeys.documents],
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Document added' });
      setName('');
      setFile(null);
      onClose();
    },
  });

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const next = {
      name: name.trim() ? undefined : 'Enter a name',
      file: file ? undefined : 'Choose a file',
    };
    setMissing(next);
    if (!next.name && !next.file) {
      void save.run();
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Add document">
      <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
        {save.formError && <Alert tone="error">{save.formError}</Alert>}
        <Field label="Name" error={missing.name ?? save.fieldErrors.name}>
          <Input value={name} onChange={(e) => setName(e.target.value)} />
        </Field>
        <FileField
          label="File"
          value={file}
          onChange={setFile}
          error={missing.file ?? save.fieldErrors.file}
        />
        <div className="flex justify-end gap-2">
          <Button variant="secondary" onClick={onClose} disabled={save.busy}>
            Cancel
          </Button>
          <Button type="submit" loading={save.busy}>
            Add document
          </Button>
        </div>
      </form>
    </Modal>
  );
}
