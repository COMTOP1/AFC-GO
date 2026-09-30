import { useState, type FormEvent } from 'react';

import { ApiError } from '../../api/client';
import { queryKeys } from '../../api/queries';
import { useTeams } from '../../api/teams';
import { createUser, updateUser, type AdminUser, type CreatedUser } from '../../api/users';
import { ImageField } from '../../components/edit/ImageField';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Input, Select } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';
import { emptyImage, type ImageValue } from '../../lib/images';
import { MANAGER_ROLE, ROLES } from '../../lib/roles';
import type { Secret } from './SecretDialog';

type Missing = { name?: string; email?: string; role?: string; team?: string };
type Result = { saved: AdminUser } | { created: CreatedUser };

export interface UserDialogProps {
  open: boolean;
  user?: AdminUser;
  /** Editing your own account: the role is locked and not sent. */
  isSelf: boolean;
  onClose: () => void;
  onSecret: (secret: Secret) => void;
}

function UserForm({ user, isSelf, onClose, onSecret }: Omit<UserDialogProps, 'open'>) {
  const toast = useToast();
  const teams = useTeams();
  const [name, setName] = useState(user?.name ?? '');
  const [email, setEmail] = useState(user?.email ?? '');
  const [phone, setPhone] = useState(user?.phone ?? '');
  const [role, setRole] = useState(user?.roleCode ?? '');
  const [teamId, setTeamId] = useState(user?.teamId ? String(user.teamId) : '');
  const [image, setImage] = useState<ImageValue>(emptyImage);
  const [missing, setMissing] = useState<Missing>({});
  const isManager = role === MANAGER_ROLE;

  const save = useSaveForm<Result>({
    submit: async () => {
      const input = {
        name,
        email,
        phone,
        role: isSelf ? undefined : role,
        teamId: isManager ? Number(teamId) : undefined,
        image,
      };
      try {
        return user
          ? { saved: await updateUser(user.id, input) }
          : { created: await createUser(input) };
      } catch (err) {
        // The server reports a duplicate email as a 409 with no field; show it on Email.
        if (err instanceof ApiError && err.status === 409) {
          throw new ApiError(409, err.message, { email: 'That email address is already in use' });
        }
        throw err;
      }
    },
    invalidate: [queryKeys.users, queryKeys.contact, ['team'], ...(isSelf ? [queryKeys.me] : [])],
    onSaved: (result) => {
      onClose();
      if ('saved' in result) {
        toast.show({ tone: 'success', message: 'User saved' });
        return;
      }
      const { created } = result;
      if (!created.emailSent && created.tempPassword) {
        onSecret({
          title: `Pass this on to ${created.user.name}`,
          message: `We couldn't email ${created.user.name}. Give them this temporary password; they'll choose a new one when they first sign in. It won't be shown again.`,
          label: 'Temporary password',
          value: created.tempPassword,
        });
        return;
      }
      toast.show({
        tone: 'success',
        message: "User added. They've been emailed a temporary password.",
      });
    },
  });

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const next: Missing = {
      name: name.trim() ? undefined : 'Enter a name',
      email: email.trim() ? undefined : 'Enter an email address',
      role: role ? undefined : 'Choose a role',
      team: !isManager || teamId ? undefined : "Choose the manager's team",
    };
    setMissing(next);
    if (!next.name && !next.email && !next.role && !next.team) {
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
      <Field label="Email" error={missing.email ?? err.email}>
        <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
      </Field>
      <Field label="Phone" error={err.phone}>
        <Input type="tel" value={phone} onChange={(e) => setPhone(e.target.value)} />
      </Field>
      <Field
        label="Role"
        error={missing.role ?? err.role}
        help={isSelf ? 'Ask another administrator to change your role.' : undefined}
      >
        <Select value={role} disabled={isSelf} onChange={(e) => setRole(e.target.value)}>
          <option value="">Choose…</option>
          {ROLES.map((r) => (
            <option key={r.code} value={r.code}>
              {r.label}
            </option>
          ))}
        </Select>
      </Field>
      {isManager && (
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
      )}
      <ImageField
        label="Photo"
        currentUrl={user?.imageUrl}
        allowRemove={Boolean(user)}
        value={image}
        onChange={setImage}
        error={err.file ?? err.image}
      />
      <div className="flex justify-end gap-2">
        <Button variant="secondary" onClick={onClose} disabled={save.busy}>
          Cancel
        </Button>
        <Button type="submit" loading={save.busy}>
          {user ? 'Save user' : 'Add user'}
        </Button>
      </div>
    </form>
  );
}

/** Add or edit a user; the form mounts only while open, so each open starts fresh. */
export function UserDialog({ open, user, isSelf, onClose, onSecret }: UserDialogProps) {
  return (
    <Modal open={open} onClose={onClose} title={user ? `Edit ${user.name}` : 'Add user'}>
      <UserForm user={user} isSelf={isSelf} onClose={onClose} onSecret={onSecret} />
    </Modal>
  );
}
