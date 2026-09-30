import { useState, type FormEvent } from 'react';

import { createPlayer, updatePlayer, type Player } from '../../api/players';
import { queryKeys } from '../../api/queries';
import { useTeams } from '../../api/teams';
import { ImageField } from '../../components/edit/ImageField';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Checkbox } from '../../components/ui/Checkbox';
import { Input, Select } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';
import { toDateInput } from '../../lib/editForm';
import { emptyImage, type ImageValue } from '../../lib/images';

type Missing = { name?: string; team?: string; dob?: string };

function PlayerForm({ player, onClose }: { player?: Player; onClose: () => void }) {
  const toast = useToast();
  const teams = useTeams();
  const [name, setName] = useState(player?.name ?? '');
  const [teamId, setTeamId] = useState(player?.team ? String(player.team.id) : '');
  const [dob, setDob] = useState(player?.dateOfBirth ? toDateInput(player.dateOfBirth) : '');
  const [position, setPosition] = useState(player?.position ?? '');
  const [isCaptain, setIsCaptain] = useState(player?.isCaptain ?? false);
  const [image, setImage] = useState<ImageValue>(emptyImage);
  const [missing, setMissing] = useState<Missing>({});

  const save = useSaveForm({
    submit: () => {
      const input = { name, teamId: Number(teamId), dateOfBirth: dob, position, isCaptain, image };
      return player ? updatePlayer(player.id, input) : createPlayer(input);
    },
    invalidate: [queryKeys.players, ['team']],
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Player saved' });
      onClose();
    },
  });

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const next: Missing = {
      name: name.trim() ? undefined : 'Enter a name',
      team: teamId ? undefined : 'Choose a team',
      dob: dob ? undefined : 'Enter the date of birth',
    };
    setMissing(next);
    if (!next.name && !next.team && !next.dob) {
      void save.run();
    }
  }

  const err = save.fieldErrors;
  return (
    <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
      {save.formError && <Alert tone="error">{save.formError}</Alert>}
      <Field label="Name" error={missing.name ?? err.name}>
        <Input value={name} onChange={(e) => setName(e.target.value)} />
      </Field>
      <Field label="Team" error={missing.team ?? err.teamId}>
        <Select value={teamId} onChange={(e) => setTeamId(e.target.value)}>
          <option value="">Choose…</option>
          {teams.data?.map((t) => (
            <option key={t.id} value={String(t.id)}>
              {t.name}
            </option>
          ))}
        </Select>
      </Field>
      <Field label="Date of birth" error={missing.dob ?? err.dateOfBirth}>
        <Input type="date" value={dob} onChange={(e) => setDob(e.target.value)} />
      </Field>
      <Field label="Position" error={err.position}>
        <Input value={position} onChange={(e) => setPosition(e.target.value)} />
      </Field>
      <Checkbox
        label="Captain"
        checked={isCaptain}
        onChange={(e) => setIsCaptain(e.target.checked)}
      />
      <ImageField
        label="Photo"
        currentUrl={player?.imageUrl}
        allowRemove={Boolean(player)}
        value={image}
        onChange={setImage}
        error={err.file ?? err.image}
      />
      <p className="text-sm text-muted">
        Photos are never shown for youth-team or under-18 players.
      </p>
      <div className="flex justify-end gap-2">
        <Button variant="secondary" onClick={onClose} disabled={save.busy}>
          Cancel
        </Button>
        <Button type="submit" loading={save.busy}>
          Save player
        </Button>
      </div>
    </form>
  );
}

/** Add or edit a player; the form mounts only while open, so each open starts fresh. */
export function PlayerDialog({
  open,
  player,
  onClose,
}: {
  open: boolean;
  player?: Player;
  onClose: () => void;
}) {
  return (
    <Modal open={open} onClose={onClose} title={player ? `Edit ${player.name}` : 'Add player'}>
      <PlayerForm player={player} onClose={onClose} />
    </Modal>
  );
}
