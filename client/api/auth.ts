import { apiFetch } from './client';
import type { LoginResponse } from './types';

export interface LoginInput {
  email: string;
  password: string;
  remember: boolean;
}

export function login(input: LoginInput): Promise<LoginResponse> {
  return apiFetch<LoginResponse>('/auth/login', { json: input });
}

export function logout(): Promise<void> {
  return apiFetch<void>('/auth/logout', { method: 'POST' });
}
