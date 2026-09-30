import { useState, type FormEvent } from 'react';

import { queryKeys, useSite } from '../../api/queries';
import { createSponsor } from '../../api/sponsors';
import { ImageField } from '../../components/edit/ImageField';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Input, Select } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';
import { emptyImage, type ImageValue } from '../../lib/images';
import { SPONSOR_TEAM_CHOICES } from '../../lib/sponsorTeam';

function AddSponsorDialogForm({ onClose }: { onClose: () => void }) {
  const toast = useToast();
  const site = useSite();
  const [name, setName] = useState('');
  const [website, setWebsite] = useState('');
  const [purpose, setPurpose] = useState('');
  const [team, setTeam] = useState('');
  const [image, setImage] = useState<ImageValue>(emptyImage);
  const [missing, setMissing] = useState<{ name?: string; image?: string }>({});
  const save = useSaveForm({
    submit: () => createSponsor({ name, website, purpose, team, image: image.file as File }),
    invalidate: [queryKeys.sponsors, queryKeys.home, ['team']],
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Sponsor added' });
      setName('');
      setWebsite('');
      setPurpose('');
      setTeam('');
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
    <>
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
        <Field label="Purpose" error={save.fieldErrors.purpose}>
          <Input
            value={purpose}
            placeholder="e.g. Kit sponsor"
            onChange={(e) => setPurpose(e.target.value)}
          />
        </Field>
        <Field label="Sponsors" error={save.fieldErrors.team}>
          <Select value={team} onChange={(e) => setTeam(e.target.value)}>
            {SPONSOR_TEAM_CHOICES.map((c) => (
              <option key={c.value} value={c.value}>
                {c.label}
              </option>
            ))}
            {site.data?.teams.map((t) => (
              <option key={t.id} value={String(t.id)}>
                {t.name}
              </option>
            ))}
          </Select>
        </Field>
        <div className="flex justify-end gap-2">
          <Button variant="secondary" onClick={onClose} disabled={save.busy}>
            Cancel
          </Button>
          <Button type="submit" loading={save.busy}>
            Add sponsor
          </Button>
        </div>
      </form>
    </>
  );
}

/** Mounted only while open (Modal renders children only then), so every open starts empty. */
export function AddSponsorDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  return (
    <Modal open={open} onClose={onClose} title="Add sponsor">
      <AddSponsorDialogForm onClose={onClose} />
    </Modal>
  );
}
