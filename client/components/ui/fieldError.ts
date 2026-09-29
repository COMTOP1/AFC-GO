import { ApiError } from '../../api/client';

/** The server's validation message for one form field, if the error carries one. */
export function fieldError(err: unknown, name: string): string | undefined {
  return err instanceof ApiError ? err.fields[name] : undefined;
}
