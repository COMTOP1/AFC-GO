import { useId, useState, type FormEvent } from 'react';

import { changePassword } from '../../api/account';
import { ApiError } from '../../api/client';
import { PasswordFields } from '../../components/page/PasswordFields';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Card, CardBody } from '../../components/ui/Card';
import { Input } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { fieldError } from '../../components/ui/fieldError';
import { useToast } from '../../components/ui/toast/useToast';
import { useAuth } from '../../auth/useAuth';
import { isSessionExpired } from '../../lib/session';

type Key = 'oldPassword' | 'newPassword' | 'confirmationPassword';
const blank: Record<Key, string> = { oldPassword: '', newPassword: '', confirmationPassword: '' };
const emptyMessages: Record<Key, string> = {
  oldPassword: 'Enter your current password',
  newPassword: 'Enter a new password',
  confirmationPassword: 'Confirm your new password',
};

export function PasswordCard() {
  const headingId = useId();
  const toast = useToast();
  const { refresh } = useAuth();
  const [values, setValues] = useState(blank);
  const [errors, setErrors] = useState<Partial<Record<Key, string>>>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const set = (key: Key, value: string) => setValues((v) => ({ ...v, [key]: value }));

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const empty: Partial<Record<Key, string>> = {};
    for (const key of Object.keys(blank) as Key[]) {
      if (!values[key]) {
        empty[key] = emptyMessages[key];
      }
    }
    if (Object.keys(empty).length > 0) {
      setErrors(empty);
      return;
    }
    setBusy(true);
    setErrors({});
    setFormError(null);
    try {
      await changePassword(values);
      setValues(blank);
      toast.show({ tone: 'success', message: 'Password changed' });
    } catch (err) {
      if (isSessionExpired(err)) {
        // The page flips to its sign-in prompt once "me" is re-read.
        await refresh();
        return;
      }
      if (err instanceof ApiError && Object.keys(err.fields).length > 0) {
        setErrors({
          oldPassword: fieldError(err, 'oldPassword'),
          newPassword: fieldError(err, 'newPassword'),
          confirmationPassword: fieldError(err, 'confirmationPassword'),
        });
      } else {
        setFormError(
          err instanceof Error ? err.message : 'Something went wrong. Please try again.',
        );
      }
    } finally {
      setBusy(false);
    }
  }

  return (
    <Card>
      <CardBody>
        <section aria-labelledby={headingId}>
          <h2
            id={headingId}
            className="mb-4 font-display text-2xl font-extrabold tracking-wide uppercase"
          >
            Change password
          </h2>
          <form onSubmit={onSubmit} className="flex max-w-md flex-col gap-4" noValidate>
            <Field label="Current password" error={errors.oldPassword}>
              <Input
                type="password"
                autoComplete="current-password"
                value={values.oldPassword}
                onChange={(e) => set('oldPassword', e.target.value)}
              />
            </Field>
            <PasswordFields
              newPassword={values.newPassword}
              confirmationPassword={values.confirmationPassword}
              onChange={set}
              errors={errors}
            />
            {formError && <Alert tone="error">{formError}</Alert>}
            <div>
              <Button type="submit" loading={busy}>
                Change password
              </Button>
            </div>
          </form>
        </section>
      </CardBody>
    </Card>
  );
}
