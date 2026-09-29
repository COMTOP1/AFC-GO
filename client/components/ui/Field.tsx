import { clsx } from 'clsx';
import { useId, type ReactNode } from 'react';

import { FieldContext, type FieldControlProps } from './fieldContext';

export interface FieldProps {
  label: string;
  help?: string;
  error?: string;
  id?: string;
  children: ReactNode;
  className?: string;
}

export function Field({ label, help, error, id, children, className }: FieldProps) {
  const autoId = useId();
  const controlId = id ?? autoId;
  const messageId = `${controlId}-message`;
  const message = error ?? help;
  const control: FieldControlProps = {
    id: controlId,
    'aria-describedby': message ? messageId : undefined,
    'aria-invalid': error ? true : undefined,
  };

  return (
    <div className={clsx('flex flex-col gap-1.5', className)}>
      <label htmlFor={controlId} className="text-sm font-semibold">
        {label}
      </label>
      <FieldContext.Provider value={control}>{children}</FieldContext.Provider>
      {message && (
        <p id={messageId} className={error ? 'text-sm text-red' : 'text-xs text-muted'}>
          {message}
        </p>
      )}
    </div>
  );
}
