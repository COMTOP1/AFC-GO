import { useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';

import {
  createSeason,
  deleteSeason,
  renameSeason,
  useSeasons,
  type Season,
} from '../../api/programmes';
import { DeleteButton } from '../../components/edit/DeleteButton';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Input } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';

const refresh = [['seasons'], ['programmes']];

function SeasonRow({ season }: { season: Season }) {
  const toast = useToast();
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(season.name);
  const save = useSaveForm({
    submit: () => renameSeason(season.id, name),
    invalidate: refresh,
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Season renamed' });
      setEditing(false);
    },
  });
  if (editing) {
    return (
      <li className="flex flex-wrap items-end gap-2 py-2">
        <Field
          label={`New name for ${season.name}`}
          error={save.fieldErrors.name ?? save.formError ?? undefined}
          className="flex-1"
        >
          <Input value={name} onChange={(e) => setName(e.target.value)} />
        </Field>
        <Button size="sm" loading={save.busy} onClick={() => void save.run()}>
          Save name
        </Button>
        <Button size="sm" variant="secondary" onClick={() => setEditing(false)}>
          Cancel
        </Button>
      </li>
    );
  }
  return (
    <li className="flex items-center justify-between gap-2 py-2">
      <span>{season.name}</span>
      <span className="flex gap-2">
        <Button
          size="sm"
          variant="secondary"
          aria-label={`Rename ${season.name}`}
          onClick={() => setEditing(true)}
        >
          Rename
        </Button>
        <DeleteButton
          ariaLabel={`Delete ${season.name}`}
          confirmTitle={`Delete season ${season.name}?`}
          confirmMessage="Its programmes will stay, with no season."
          onDelete={() => deleteSeason(season.id)}
          invalidate={refresh}
          successMessage="Season deleted"
        />
      </span>
    </li>
  );
}

export function SeasonsDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const toast = useToast();
  const seasons = useSeasons();
  const queryClient = useQueryClient();
  const [name, setName] = useState('');
  const [missing, setMissing] = useState<string | undefined>();
  const add = useSaveForm({
    submit: () => createSeason(name),
    invalidate: refresh,
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Season added' });
      setName('');
    },
  });

  return (
    <Modal open={open} onClose={onClose} title="Seasons">
      <div className="flex flex-col gap-4">
        {seasons.isError && (
          <Alert tone="error">
            Couldn&apos;t load the seasons.{' '}
            <Button
              size="sm"
              variant="secondary"
              onClick={() => void queryClient.refetchQueries({ queryKey: ['seasons'] })}
            >
              Retry
            </Button>
          </Alert>
        )}
        <ul className="divide-y divide-line">
          {seasons.data?.map((s) => (
            <SeasonRow key={`${s.id}-${s.name}`} season={s} />
          ))}
        </ul>
        <form
          noValidate
          className="flex flex-wrap items-end gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            if (!name.trim()) {
              setMissing('Enter a season name');
              return;
            }
            setMissing(undefined);
            void add.run();
          }}
        >
          <Field
            label="New season"
            error={missing ?? add.fieldErrors.name ?? add.formError ?? undefined}
            className="flex-1"
          >
            <Input
              value={name}
              placeholder="e.g. 2027-28"
              onChange={(e) => setName(e.target.value)}
            />
          </Field>
          <Button type="submit" loading={add.busy}>
            Add season
          </Button>
        </form>
        <div className="flex justify-end">
          <Button variant="secondary" onClick={onClose}>
            Done
          </Button>
        </div>
      </div>
    </Modal>
  );
}
