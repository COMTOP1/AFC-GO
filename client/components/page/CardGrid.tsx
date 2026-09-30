import { useState, type ReactNode } from 'react';

import { Button } from '../ui/Button';
import { EmptyState } from '../ui/EmptyState';

export interface CardGridProps<T> {
  items: T[];
  render: (item: T) => ReactNode;
  getKey: (item: T) => string | number;
  /** Cards per "Show more" step; Infinity shows everything. */
  pageSize?: number;
  /** Changing this (e.g. a tab value) goes back to the first page. */
  resetKey?: string;
  emptyTitle: string;
  emptyMessage?: ReactNode;
}

export function CardGrid<T>({
  items,
  render,
  getKey,
  pageSize = 12,
  resetKey,
  emptyTitle,
  emptyMessage,
}: CardGridProps<T>) {
  const [shown, setShown] = useState(pageSize);
  const [lastKey, setLastKey] = useState(resetKey);
  if (resetKey !== lastKey) {
    setLastKey(resetKey);
    setShown(pageSize);
  }

  if (items.length === 0) {
    return <EmptyState title={emptyTitle} message={emptyMessage} />;
  }
  const remaining = items.length - shown;
  return (
    <>
      <ul className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {items.slice(0, shown).map((item) => (
          <li key={getKey(item)}>{render(item)}</li>
        ))}
      </ul>
      {remaining > 0 && (
        <div className="mt-6 text-center">
          <Button variant="secondary" onClick={() => setShown((s) => s + pageSize)}>
            Show more ({remaining} more)
          </Button>
        </div>
      )}
    </>
  );
}
