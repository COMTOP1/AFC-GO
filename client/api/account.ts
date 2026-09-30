import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import type { CurrentUser } from './types';

/** auth.PasswordInput — POST /auth/password */
export interface PasswordChange {
  oldPassword: string;
  newPassword: string;
  confirmationPassword: string;
}

/** auth.ResetInput — POST /auth/reset/:token */
export interface PasswordReset {
  newPassword: string;
  confirmationPassword: string;
}

export function uploadAccountImage(file: File): Promise<CurrentUser> {
  const form = new FormData();
  form.append('image', file);
  return apiFetch<CurrentUser>('/account/image', { method: 'PUT', form });
}

export function removeAccountImage(): Promise<void> {
  return apiFetch<void>('/account/image', { method: 'DELETE' });
}

export function changePassword(input: PasswordChange): Promise<void> {
  return apiFetch<void>('/auth/password', { json: input });
}

export function checkResetToken(token: string, signal?: AbortSignal): Promise<void> {
  return apiFetch<void>(`/auth/reset/${encodeURIComponent(token)}`, { signal });
}

export function resetPassword(token: string, input: PasswordReset): Promise<void> {
  return apiFetch<void>(`/auth/reset/${encodeURIComponent(token)}`, { json: input });
}

/** Whether a reset link is still usable; disabled (never fetched) for a null token. */
export function useResetTokenCheck(token: string | null) {
  return useQuery({
    queryKey: ['auth', 'reset', token ?? ''],
    queryFn: async ({ signal }) => {
      await checkResetToken(token as string, signal);
      return true; // query data can't be undefined
    },
    enabled: token !== null,
    retry: false,
    gcTime: 0,
  });
}
