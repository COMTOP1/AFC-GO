import { QueryClient } from '@tanstack/react-query';

import { ApiError } from './client';

/** Client errors (4xx) never succeed on retry; anything else gets two more tries. */
export function shouldRetry(failureCount: number, error: unknown): boolean {
  if (error instanceof ApiError && error.status >= 400 && error.status < 500) {
    return false;
  }
  return failureCount < 2;
}

export function createQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: shouldRetry, staleTime: 30_000 },
    },
  });
}
