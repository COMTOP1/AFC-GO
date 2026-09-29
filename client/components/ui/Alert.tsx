import { clsx } from 'clsx';
import type { ComponentProps } from 'react';

export type AlertTone = 'success' | 'error' | 'warning' | 'info';

const tones: Record<AlertTone, string> = {
  success: 'border-success',
  error: 'border-red',
  warning: 'border-warning',
  info: 'border-blue',
};

export function Alert({
  tone = 'info',
  className,
  ...props
}: ComponentProps<'div'> & { tone?: AlertTone }) {
  return (
    <div
      role={tone === 'error' ? 'alert' : 'status'}
      className={clsx(
        'rounded-lg border-l-4 bg-surface px-3 py-2.5 text-sm text-ink',
        tones[tone],
        className,
      )}
      {...props}
    />
  );
}
