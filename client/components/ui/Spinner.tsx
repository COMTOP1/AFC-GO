import { clsx } from 'clsx';

export function Spinner({ label = 'Loading', className }: { label?: string; className?: string }) {
  return (
    <span role="status" className={clsx('inline-flex items-center', className)}>
      <span
        aria-hidden="true"
        className="size-5 rounded-full border-2 border-line border-t-red motion-safe:animate-spin"
      />
      <span className="sr-only">{label}</span>
    </span>
  );
}
