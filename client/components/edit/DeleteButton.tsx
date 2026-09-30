import { useQueryClient, type QueryKey } from '@tanstack/react-query';
import { useState, type ReactNode } from 'react';

import { useAuth } from '../../auth/useAuth';
import { describeSaveError } from '../../lib/saveErrors';
import { isSessionExpired } from '../../lib/session';
import { Button } from '../ui/Button';
import type { ButtonSize } from '../ui/buttonStyles';
import { ConfirmDialog } from '../ui/ConfirmDialog';
import { useToast } from '../ui/toast/useToast';

export interface DeleteButtonProps {
  confirmTitle: string;
  confirmMessage: ReactNode;
  onDelete: () => Promise<unknown>;
  invalidate?: QueryKey[];
  successMessage: string;
  after?: () => void;
  /** Visible text; defaults to "Delete". */
  children?: ReactNode;
  /** Accessible name when the text alone is ambiguous (e.g. "Delete Club rules"). */
  ariaLabel?: string;
  size?: ButtonSize;
}

export function DeleteButton({
  confirmTitle,
  confirmMessage,
  onDelete,
  invalidate = [],
  successMessage,
  after,
  children = 'Delete',
  ariaLabel,
  size = 'sm',
}: DeleteButtonProps) {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();
  const toast = useToast();
  const { refresh } = useAuth();

  async function confirm() {
    try {
      await onDelete();
      for (const queryKey of invalidate) {
        void queryClient.invalidateQueries({ queryKey });
      }
      setOpen(false);
      toast.show({ tone: 'success', message: successMessage });
      after?.();
    } catch (err) {
      setOpen(false);
      if (isSessionExpired(err)) {
        await refresh();
        return;
      }
      toast.show({ tone: 'error', message: `Couldn't delete: ${describeSaveError(err)}` });
    }
  }

  return (
    <>
      <Button variant="danger" size={size} aria-label={ariaLabel} onClick={() => setOpen(true)}>
        {children}
      </Button>
      <ConfirmDialog
        open={open}
        title={confirmTitle}
        message={confirmMessage}
        confirmLabel="Delete"
        tone="danger"
        onConfirm={confirm}
        onCancel={() => setOpen(false)}
      />
    </>
  );
}
