import { useQueryClient, type QueryKey } from '@tanstack/react-query';
import { useState } from 'react';

import { ApiError } from '../../api/client';
import { useAuth } from '../../auth/useAuth';
import { describeSaveError } from '../../lib/saveErrors';
import { isSessionExpired } from '../../lib/session';

export interface SaveFormOptions<T> {
  submit: () => Promise<T>;
  /** Query keys (or prefixes) to refresh after a successful save. */
  invalidate?: QueryKey[];
  onSaved?: (result: T) => void;
}

/** Busy state, server field errors, and the after-save refresh for an edit form. */
export function useSaveForm<T>({ submit, invalidate = [], onSaved }: SaveFormOptions<T>) {
  const queryClient = useQueryClient();
  const { refresh } = useAuth();
  const [busy, setBusy] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState<string | null>(null);

  async function run() {
    setBusy(true);
    setFieldErrors({});
    setFormError(null);
    try {
      const result = await submit();
      for (const queryKey of invalidate) {
        void queryClient.invalidateQueries({ queryKey });
      }
      onSaved?.(result);
    } catch (err) {
      if (isSessionExpired(err)) {
        await refresh();
      } else if (err instanceof ApiError && Object.keys(err.fields).length > 0) {
        setFieldErrors(err.fields);
      } else {
        setFormError(describeSaveError(err));
      }
    } finally {
      setBusy(false);
    }
  }

  return { run, busy, fieldErrors, formError };
}
