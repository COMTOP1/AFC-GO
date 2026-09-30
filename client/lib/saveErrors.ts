import { ApiError } from '../api/client';

/** A save/delete failure as a sentence for the person who pressed the button. */
export function describeSaveError(err: unknown): string {
  if (err instanceof ApiError && err.status === 413) {
    return 'That file is too large (15 MB maximum).';
  }
  return err instanceof Error ? err.message : 'Something went wrong. Please try again.';
}
