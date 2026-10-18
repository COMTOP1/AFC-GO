import { useState, type FormEvent } from 'react';

import { setTeamEmailMode, useContact, type TeamEmailMode } from '../../api/pages';
import { queryKeys } from '../../api/queries';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Card, CardBody } from '../../components/ui/Card';
import { Select } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';

const modes: { value: TeamEmailMode; label: string; help: string }[] = [
  {
    value: 'manager',
    label: "Each manager's own email",
    help: "Team pages show every manager's personal email address.",
  },
  {
    value: 'team',
    label: "The team's contact email",
    help: "Team pages show the contact email set on each team; managers' names only.",
  },
  {
    value: 'site',
    label: 'The public contact email',
    help: "Team pages show the site-wide public contact email; managers' names only.",
  },
];

function labelOf(mode: TeamEmailMode): string {
  return modes.find((m) => m.value === mode)?.label ?? mode;
}

function TeamEmailForm({ initial, onClose }: { initial: TeamEmailMode; onClose: () => void }) {
  const toast = useToast();
  const [mode, setMode] = useState<TeamEmailMode>(initial);
  const save = useSaveForm({
    submit: () => setTeamEmailMode(mode),
    invalidate: [queryKeys.contact, ['team']],
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Team page email saved' });
      onClose();
    },
  });
  return (
    <form
      noValidate
      className="flex flex-col gap-4"
      onSubmit={(e: FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        void save.run();
      }}
    >
      {save.formError && <Alert tone="error">{save.formError}</Alert>}
      <Field
        label="Team pages show"
        error={save.fieldErrors.mode}
        help={modes.find((m) => m.value === mode)?.help}
      >
        <Select value={mode} onChange={(e) => setMode(e.target.value as TeamEmailMode)}>
          {modes.map((m) => (
            <option key={m.value} value={m.value}>
              {m.label}
            </option>
          ))}
        </Select>
      </Field>
      <div className="flex justify-end gap-2">
        <Button variant="secondary" onClick={onClose} disabled={save.busy}>
          Cancel
        </Button>
        <Button type="submit" loading={save.busy}>
          Save
        </Button>
      </div>
    </form>
  );
}

/** Which email address team pages show (team-email-mode setting). */
export function TeamEmailCard() {
  const contact = useContact();
  const [editing, setEditing] = useState(false);
  const mode = contact.data?.teamEmailMode ?? 'manager';
  return (
    <Card className="mb-6">
      <CardBody className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="font-display text-xl font-extrabold uppercase">Team page email</h2>
          <p>
            {contact.isPending
              ? 'Loading…'
              : contact.isError
                ? "Couldn't load the current setting."
                : labelOf(mode)}
          </p>
          <p className="text-sm text-muted">
            How to contact a team. If the chosen address isn&apos;t set, no email is shown.
          </p>
        </div>
        <Button
          variant="secondary"
          aria-label="Edit team page email"
          disabled={!contact.isSuccess}
          onClick={() => setEditing(true)}
        >
          Edit
        </Button>
      </CardBody>
      <Modal open={editing} onClose={() => setEditing(false)} title="Team page email">
        <TeamEmailForm initial={mode} onClose={() => setEditing(false)} />
      </Modal>
    </Card>
  );
}
