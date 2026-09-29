import { clsx } from 'clsx';
import type { ComponentProps } from 'react';

export function Container({ className, ...props }: ComponentProps<'div'>) {
  return <div className={clsx('mx-auto w-full max-w-6xl px-4 md:px-6', className)} {...props} />;
}
