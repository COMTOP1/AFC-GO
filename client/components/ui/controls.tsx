import { clsx } from 'clsx';
import type { ComponentProps } from 'react';

import { controlClass, useFieldControl } from './fieldContext';

export function Input({ className, ...props }: ComponentProps<'input'>) {
  const field = useFieldControl();
  return <input {...field} {...props} className={clsx(controlClass, className)} />;
}

export function Textarea({ className, ...props }: ComponentProps<'textarea'>) {
  const field = useFieldControl();
  return <textarea rows={4} {...field} {...props} className={clsx(controlClass, className)} />;
}

export function Select({ className, ...props }: ComponentProps<'select'>) {
  const field = useFieldControl();
  return <select {...field} {...props} className={clsx(controlClass, className)} />;
}

export function FileInput({ className, ...props }: Omit<ComponentProps<'input'>, 'type'>) {
  const field = useFieldControl();
  return (
    <input
      type="file"
      {...field}
      {...props}
      className={clsx(
        controlClass,
        'file:mr-3 file:rounded file:border-0 file:bg-surface file:px-2 file:py-1 file:text-ink',
        className,
      )}
    />
  );
}
