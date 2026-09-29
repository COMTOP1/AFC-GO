import { createContext, useContext } from 'react';

/** Props a Field hands to its control so label, help and error are wired up. */
export interface FieldControlProps {
  id: string;
  'aria-describedby'?: string;
  'aria-invalid'?: true;
}

export const FieldContext = createContext<FieldControlProps | null>(null);

export function useFieldControl(): FieldControlProps | null {
  return useContext(FieldContext);
}

export const controlClass =
  'w-full rounded-md border border-line bg-field px-2.5 py-2 text-sm text-ink placeholder:text-muted focus-visible:border-blue disabled:opacity-60 aria-[invalid=true]:border-red';
