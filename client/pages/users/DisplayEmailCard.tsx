import { useState, type FormEvent } from 'react';

import { setDisplayEmail, useContact } from '../../api/pages';
import { queryKeys } from '../../api/queries';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Card, CardBody } from '../../components/ui/Card';
import { Input } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';

function DisplayEmailForm({ initial, onClose }: { initial: string; onClose: () => void }) {
  const toast = useToast();
  const [email, setEmail] = useState(initial);
  const save = useSaveForm({
    submit: () => setDisplayEmail(email),
    invalidate: [queryKeys.contact],
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Contact email saved' });
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
        label="Email"
        error={save.fieldErrors.email}
        help="Leave empty to remove it from the Contact page."
      >
        <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
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

/** The public contact email shown on the Contact page (display-email setting). */
export function DisplayEmailCard() {
  const contact = useContact();
  const [editing, setEditing] = useState(false);
  const email = contact.data?.displayEmail ?? '';
  return (
    <Card className="mb-6">
      <CardBody className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="font-display text-xl font-extrabold uppercase">Public contact email</h2>
          <p>{contact.isPending ? 'Loading…' : email || 'Not set'}</p>
          <p className="text-sm text-muted">Shown on the Contact page.</p>
        </div>
        <Button
          variant="secondary"
          aria-label="Edit public contact email"
          disabled={contact.isPending}
          onClick={() => setEditing(true)}
        >
          Edit
        </Button>
      </CardBody>
      <Modal open={editing} onClose={() => setEditing(false)} title="Public contact email">
        <DisplayEmailForm initial={email} onClose={() => setEditing(false)} />
      </Modal>
    </Card>
  );
}
