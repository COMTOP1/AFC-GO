import { ApiError } from '../api/client';

/** True when the server rejected the request because the session has gone. */
export function isSessionExpired(error: unknown): boolean {
  return error instanceof ApiError && error.status === 401;
}
