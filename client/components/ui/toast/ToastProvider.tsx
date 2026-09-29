import { clsx } from 'clsx';
import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from 'react';

import {
  MAX_TOASTS,
  TOAST_DURATION_MS,
  ToastContext,
  type ToastApi,
  type ToastOptions,
  type ToastTone,
} from './context';

interface Toast {
  id: number;
  message: string;
  tone: ToastTone;
}

const dots: Record<ToastTone, string> = {
  success: 'bg-success',
  error: 'bg-red',
  info: 'bg-blue',
};

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const nextId = useRef(1);
  const timers = useRef(new Map<number, ReturnType<typeof setTimeout>>());

  const dismiss = useCallback((id: number) => {
    const timer = timers.current.get(id);
    if (timer !== undefined) {
      clearTimeout(timer);
      timers.current.delete(id);
    }
    setToasts((ts) => ts.filter((t) => t.id !== id));
  }, []);

  const show = useCallback(
    ({ message, tone = 'info' }: ToastOptions) => {
      const id = nextId.current++;
      setToasts((ts) => [...ts, { id, message, tone }].slice(-MAX_TOASTS));
      timers.current.set(
        id,
        setTimeout(() => dismiss(id), TOAST_DURATION_MS),
      );
    },
    [dismiss],
  );

  useEffect(() => {
    const pending = timers.current;
    return () => pending.forEach((timer) => clearTimeout(timer));
  }, []);

  const api = useMemo<ToastApi>(() => ({ show }), [show]);

  return (
    <ToastContext.Provider value={api}>
      {children}
      <div
        aria-live="polite"
        className="pointer-events-none fixed right-4 bottom-4 z-50 flex w-80 max-w-[calc(100vw-2rem)] flex-col gap-2"
      >
        {toasts.map((t) => (
          <button
            key={t.id}
            type="button"
            title="Dismiss"
            onClick={() => dismiss(t.id)}
            className="pointer-events-auto flex items-center gap-2.5 rounded-lg bg-ink px-3.5 py-2.5 text-left text-sm text-bg shadow-lg"
          >
            <span
              aria-hidden="true"
              className={clsx('size-2 shrink-0 rounded-full', dots[t.tone])}
            />
            {t.message}
          </button>
        ))}
      </div>
    </ToastContext.Provider>
  );
}
