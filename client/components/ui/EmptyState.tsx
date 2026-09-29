import type { ReactNode } from 'react';

export interface EmptyStateProps {
  title: string;
  message?: ReactNode;
  action?: ReactNode;
}

export function EmptyState({ title, message, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center gap-2 py-10 text-center text-muted">
      <p className="font-display text-xl font-extrabold tracking-wide text-ink uppercase">
        {title}
      </p>
      {message && <div className="text-sm">{message}</div>}
      {action && <div className="mt-2">{action}</div>}
    </div>
  );
}
