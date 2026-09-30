import { useState, type FormEvent } from 'react';

import { createProgramme, useSeasons } from '../../api/programmes';
import { FileField } from '../../components/edit/FileField';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Input, Select } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';

type Missing = { name?: string; date?: string; file?: string };

function AddProgrammeDialogForm({ onClose }: { onClose: () => void }) {
  const toast = useToast();
  const seasons = useSeasons();
  const [name, setName] = useState('');
  const [date, setDate] = useState('');
  const [seasonId, setSeasonId] = useState('');
  const [file, setFile] = useState<File | null>(null);
  const [missing, setMissing] = useState<Missing>({});
  const save = useSaveForm({
    submit: () =>
      createProgramme({
        name,
        date,
        seasonId: seasonId ? Number(seasonId) : null,
        file: file as File,
      }),
    invalidate: [['programmes']],
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Programme added' });
      setName('');
      setDate('');
      setSeasonId('');
      setFile(null);
      onClose();
    },
  });

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const next: Missing = {
      name: name.trim() ? undefined : 'Enter a name',
      date: date ? undefined : 'Choose the date',
      file: file ? undefined : 'Choose a file',
    };
    setMissing(next);
    if (!next.name && !next.date && !next.file) {
      void save.run();
    }
  }

  return (
    <>
      <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
        {save.formError && <Alert tone="error">{save.formError}</Alert>}
        <Field label="Name" error={missing.name ?? save.fieldErrors.name}>
          <Input value={name} onChange={(e) => setName(e.target.value)} />
        </Field>
        <Field label="Date" error={missing.date ?? save.fieldErrors.date}>
          <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
        </Field>
        <Field label="Season" error={save.fieldErrors.seasonId}>
          <Select value={seasonId} onChange={(e) => setSeasonId(e.target.value)}>
            <option value="">No season</option>
            {seasons.data?.map((s) => (
              <option key={s.id} value={String(s.id)}>
                {s.name}
              </option>
            ))}
          </Select>
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
            Add programme
          </Button>
        </div>
      </form>
    </>
  );
}

/** Mounted only while open (Modal renders children only then), so every open starts empty. */
export function AddProgrammeDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  return (
    <Modal open={open} onClose={onClose} title="Add programme">
      <AddProgrammeDialogForm onClose={onClose} />
    </Modal>
  );
}
