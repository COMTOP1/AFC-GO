import { useState, type FormEvent } from 'react';
import { useParams } from 'react-router';

import { resetPassword, useResetTokenCheck } from '../../api/account';
import { ApiError } from '../../api/client';
import { useSignIn } from '../../components/layout/useSignIn';
import { PasswordFields, type PasswordField } from '../../components/page/PasswordFields';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { ButtonLink } from '../../components/ui/ButtonLink';
import { EmptyState } from '../../components/ui/EmptyState';
import { fieldError } from '../../components/ui/fieldError';
import { PageHeader } from '../../components/ui/PageHeader';
import { isNotFound } from '../../lib/notFound';
import { RESET_TOKEN } from '../../lib/resetLink';

type Errors = Partial<Record<PasswordField, string>>;

function InvalidLink() {
  return (
    <EmptyState
      title="This reset link is invalid or has expired"
      message="Ask a club administrator for a new one, or sign in again if you were sent here after signing in."
      action={<ButtonLink to="/">Home</ButtonLink>}
    />
  );
}

export default function ResetPage() {
  usePageTitle('Reset password');
  const raw = useParams().token ?? '';
  const token = RESET_TOKEN.test(raw) ? raw : null;
  const check = useResetTokenCheck(token);
  const signIn = useSignIn();
  const [values, setValues] = useState({ newPassword: '', confirmationPassword: '' });
  const [errors, setErrors] = useState<Errors>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);
  const [expired, setExpired] = useState(false);

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (token === null) {
      return;
    }
    const empty: Errors = {};
    if (!values.newPassword) empty.newPassword = 'Enter a new password';
    if (!values.confirmationPassword) empty.confirmationPassword = 'Confirm your new password';
    if (Object.keys(empty).length > 0) {
      setErrors(empty);
      return;
    }
    setBusy(true);
    setErrors({});
    setFormError(null);
    try {
      await resetPassword(token, values);
      setDone(true);
    } catch (err) {
      if (isNotFound(err)) {
        setExpired(true);
      } else if (err instanceof ApiError && Object.keys(err.fields).length > 0) {
        setErrors({
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

  const header = <PageHeader title="Reset your password" />;

  if (token === null || expired || isNotFound(check.error)) {
    return (
      <>
        {header}
        <InvalidLink />
      </>
    );
  }

  if (done) {
    return (
      <>
        {header}
        <Alert tone="success" className="flex flex-wrap items-center justify-between gap-3">
          <span>Your password has been changed. You can now sign in with it.</span>
          <Button onClick={signIn.open}>Sign in</Button>
        </Alert>
      </>
    );
  }

  return (
    <>
      {header}
      <QueryState query={check}>
        {() => (
          <form onSubmit={onSubmit} className="flex max-w-md flex-col gap-4" noValidate>
            {formError && <Alert tone="error">{formError}</Alert>}
            <PasswordFields
              newPassword={values.newPassword}
              confirmationPassword={values.confirmationPassword}
              onChange={(field, value) => setValues((v) => ({ ...v, [field]: value }))}
              errors={errors}
            />
            <div>
              <Button type="submit" loading={busy}>
                Set new password
              </Button>
            </div>
          </form>
        )}
      </QueryState>
    </>
  );
}
