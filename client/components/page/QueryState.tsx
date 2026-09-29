import type { UseQueryResult } from '@tanstack/react-query';
import type { ReactNode } from 'react';

import { Alert } from '../ui/Alert';
import { Button } from '../ui/Button';
import { EmptyState } from '../ui/EmptyState';
import { Skeleton } from '../ui/Skeleton';

export function PageSkeleton() {
  return (
    <div className="space-y-3" aria-busy="true">
      <span className="sr-only">Loading</span>
      <Skeleton className="w-1/3" />
      <Skeleton />
      <Skeleton className="w-2/3" />
    </div>
  );
}

export interface QueryStateProps<T> {
  query: UseQueryResult<T>;
  children: (data: T) => ReactNode;
  isEmpty?: (data: T) => boolean;
  emptyTitle?: string;
  emptyMessage?: ReactNode;
}

/** One loading / error-with-retry / empty / content flow for every page. */
export function QueryState<T>({
  query,
  children,
  isEmpty,
  emptyTitle = 'Nothing here yet',
  emptyMessage,
}: QueryStateProps<T>) {
  if (query.isPending) {
    return <PageSkeleton />;
  }
  if (query.isError) {
    return (
      <Alert tone="error" className="flex flex-wrap items-center justify-between gap-3">
        <span>Couldn&apos;t load this: {query.error.message}</span>
        <Button size="sm" variant="secondary" onClick={() => void query.refetch()}>
          Retry
        </Button>
      </Alert>
    );
  }
  if (isEmpty?.(query.data)) {
    return <EmptyState title={emptyTitle} message={emptyMessage} />;
  }
  return <>{children(query.data)}</>;
}
