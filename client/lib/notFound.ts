import { ApiError } from '../api/client';

export function isNotFound(error: unknown): boolean {
  return error instanceof ApiError && error.status === 404;
}
