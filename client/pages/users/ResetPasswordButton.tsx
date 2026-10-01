import { useState } from 'react';

import { resetUserPassword, type AdminUser } from '../../api/users';
import { useAuth } from '../../auth/useAuth';
import { Button } from '../../components/ui/Button';
import { ConfirmDialog } from '../../components/ui/ConfirmDialog';
import { useToast } from '../../components/ui/toast/useToast';
import { describeSaveError } from '../../lib/saveErrors';
import { isSessionExpired } from '../../lib/session';
import type { Secret } from './SecretDialog';

export function ResetPasswordButton({
  user,
  onSecret,
}: {
  user: AdminUser;
  onSecret: (secret: Secret) => void;
}) {
  const [open, setOpen] = useState(false);
  const toast = useToast();
  const { refresh } = useAuth();

  async function confirm() {
    try {
      const result = await resetUserPassword(user.id);
      setOpen(false);
      if (!result.emailSent && result.resetUrl) {
        onSecret({
          title: `Pass this link on to ${user.name}`,
          message: `We couldn't email ${user.name}. Send them this link to set a new password; it lasts 7 days.`,
          label: 'Reset link',
          value: result.resetUrl,
        });
        return;
      }
      toast.show({ tone: 'success', message: `Reset link emailed to ${user.email}` });
    } catch (err) {
      setOpen(false);
      if (isSessionExpired(err)) {
        await refresh();
        return;
      }
      toast.show({
        tone: 'error',
        message: `Couldn't reset the password: ${describeSaveError(err)}`,
      });
    }
  }

  return (
    <>
      <Button
        size="sm"
        variant="secondary"
        aria-label={`Reset password for ${user.name}`}
        onClick={() => setOpen(true)}
      >
        Reset password
      </Button>
      <ConfirmDialog
        open={open}
        title={`Reset ${user.name}'s password?`}
        message="They'll be emailed a link to set a new one (valid for 7 days) and must use it before they can sign in again."
        confirmLabel="Reset password"
        onConfirm={confirm}
        onCancel={() => setOpen(false)}
      />
    </>
  );
}
