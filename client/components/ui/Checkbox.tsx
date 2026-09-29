import { clsx } from 'clsx';
import type { ComponentProps, ReactNode } from 'react';

export interface CheckboxProps extends Omit<ComponentProps<'input'>, 'type'> {
  label: ReactNode;
}

export function Checkbox({ label, className, ...props }: CheckboxProps) {
  return (
    <label className={clsx('inline-flex items-center gap-2 text-sm', className)}>
      <input type="checkbox" className="size-4 accent-red" {...props} />
      {label}
    </label>
  );
}
