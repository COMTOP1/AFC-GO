import { useRef } from 'react';

import { Button } from '../../components/ui/Button';
import { Input } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { Modal } from '../../components/ui/Modal';
import { useToast } from '../../components/ui/toast/useToast';

/** A one-time value (temporary password, reset link) for the admin to pass on. */
export interface Secret {
  title: string;
  message: string;
  label: string;
  value: string;
}

function SecretBody({ secret, onClose }: { secret: Secret; onClose: () => void }) {
  const toast = useToast();
  const inputRef = useRef<HTMLInputElement>(null);

  async function copy() {
    try {
      // navigator.clipboard is missing on plain http and may be refused.
      if (!navigator.clipboard) {
        throw new Error('no clipboard');
      }
      await navigator.clipboard.writeText(secret.value);
      toast.show({ tone: 'success', message: 'Copied' });
    } catch {
      inputRef.current?.select();
      toast.show({ tone: 'info', message: 'Select and copy the text above' });
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <p className="text-sm">{secret.message}</p>
      <Field label={secret.label}>
        <Input
          ref={inputRef}
          readOnly
          value={secret.value}
          className="font-mono"
          onFocus={(e) => e.target.select()}
        />
      </Field>
      <div className="flex justify-end gap-2">
        <Button variant="secondary" onClick={() => void copy()}>
          Copy
        </Button>
        <Button onClick={onClose}>Done</Button>
      </div>
    </div>
  );
}

export function SecretDialog({ secret, onClose }: { secret: Secret | null; onClose: () => void }) {
  return (
    <Modal open={secret !== null} onClose={onClose} title={secret?.title ?? ''}>
      {secret && <SecretBody secret={secret} onClose={onClose} />}
    </Modal>
  );
}
