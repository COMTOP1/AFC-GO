import { clsx } from 'clsx';
import type { ComponentProps } from 'react';

export type BadgeTone = 'red' | 'blue' | 'neutral';

const tones: Record<BadgeTone, string> = {
  red: 'bg-red/12 text-red',
  blue: 'bg-blue/14 text-blue',
  neutral: 'border border-line bg-surface text-muted',
};

export function Badge({
  tone = 'neutral',
  className,
  ...props
}: ComponentProps<'span'> & { tone?: BadgeTone }) {
  return (
    <span
      className={clsx(
        'inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-bold tracking-wider uppercase',
        tones[tone],
        className,
      )}
      {...props}
    />
  );
}
