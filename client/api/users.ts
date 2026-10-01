import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';
import { formData } from '../lib/editForm';
import type { ImageValue } from '../lib/images';

/** user.Admin — GET /users (user admins). role is the display name, roleCode the input code. */
export interface AdminUser {
  id: number;
  name: string;
  email: string;
  phone?: string;
  role: string;
  roleCode: string;
  teamId?: number;
  imageUrl?: string;
}

export function useUsers() {
  return useQuery({
    queryKey: queryKeys.users,
    queryFn: ({ signal }) => apiFetch<AdminUser[]>('/users', { signal }),
  });
}

export interface UserInput {
  name: string;
  email: string;
  /** Always sent; empty clears it. */
  phone: string;
  /** Role code; left out when editing your own account. */
  role?: string;
  /** Managers only. */
  teamId?: number;
  image: ImageValue;
}

/** user.Created — tempPassword is set only when the welcome email couldn't be sent. */
export interface CreatedUser {
  user: AdminUser;
  emailSent: boolean;
  tempPassword?: string;
}

/** user.ResetResult — resetUrl is set only when the reset email couldn't be sent. */
export interface ResetResult {
  emailSent: boolean;
  resetUrl?: string;
}

function userForm(input: UserInput, isUpdate: boolean): FormData {
  return formData({
    name: input.name,
    email: input.email,
    phone: input.phone,
    role: input.role,
    teamId: input.teamId,
    image: input.image.file,
    removeImage: isUpdate && input.image.remove ? true : undefined,
  });
}

export function createUser(input: UserInput): Promise<CreatedUser> {
  return apiFetch<CreatedUser>('/users', { form: userForm(input, false) });
}

export function updateUser(id: number, input: UserInput): Promise<AdminUser> {
  return apiFetch<AdminUser>(`/users/${id}`, { method: 'PATCH', form: userForm(input, true) });
}

export function deleteUser(id: number): Promise<void> {
  return apiFetch<void>(`/users/${id}`, { method: 'DELETE' });
}

export function resetUserPassword(id: number): Promise<ResetResult> {
  return apiFetch<ResetResult>(`/users/${id}/reset`, { method: 'POST' });
}
