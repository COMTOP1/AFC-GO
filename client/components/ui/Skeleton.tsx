import { clsx } from 'clsx';

export function Skeleton({ className }: { className?: string }) {
  return (
    <div
      aria-hidden="true"
      className={clsx('h-3 rounded-md bg-line motion-safe:animate-pulse', className)}
    />
  );
}
