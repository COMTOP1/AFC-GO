import { createContext } from 'react';

export type ToastTone = 'success' | 'error' | 'info';

export interface ToastOptions {
  message: string;
  tone?: ToastTone;
}

export interface ToastApi {
  show: (options: ToastOptions) => void;
}

export const TOAST_DURATION_MS = 5000;
export const MAX_TOASTS = 3;

export const ToastContext = createContext<ToastApi | null>(null);
