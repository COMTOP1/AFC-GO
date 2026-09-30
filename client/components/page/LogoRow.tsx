import { useId, type ReactNode } from 'react';

import { ImageWithFallback } from './ImageWithFallback';

export interface LogoItem {
  id: number;
  name: string;
  website?: string;
  imageUrl?: string;
}

function Logo({ item }: { item: LogoItem }) {
  const inner = (
    <ImageWithFallback
      src={item.imageUrl}
      alt={item.name}
      loading="lazy"
      className="max-h-14 w-auto"
      fallback={<span className="text-center text-sm font-semibold">{item.name}</span>}
    />
  );
  const box =
    'flex h-20 w-36 items-center justify-center rounded-lg border border-line bg-white p-2 text-black';
  if (item.website) {
    return (
      <a href={item.website} target="_blank" rel="noopener noreferrer" className={box}>
        {inner}
      </a>
    );
  }
  return <div className={box}>{inner}</div>;
}

/** A titled row of sponsor/affiliation logos; renders nothing when empty unless asked to. */
export interface LogoRowProps {
  title: string;
  items: LogoItem[];
  /** Extra control under each logo (e.g. an editor's delete button). */
  itemAction?: (item: LogoItem) => ReactNode;
  /** Render the row even with no items (editors, so they can add one). */
  showWhenEmpty?: boolean;
  emptyText?: string;
  /** A control beside the row's heading (e.g. "Add affiliation"). */
  headerAction?: ReactNode;
}

export function LogoRow({
  title,
  items,
  itemAction,
  showWhenEmpty,
  emptyText,
  headerAction,
}: LogoRowProps) {
  const headingId = useId();
  if (items.length === 0 && !showWhenEmpty) {
    return null;
  }
  return (
    <section aria-labelledby={headingId}>
      <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
        <h2 id={headingId} className="font-display text-2xl font-extrabold tracking-wide uppercase">
          {title}
        </h2>
        {headerAction}
      </div>
      {items.length === 0 ? (
        <p className="text-sm text-muted">{emptyText}</p>
      ) : (
        <ul className="flex flex-wrap gap-3">
          {items.map((item) => (
            <li key={item.id} className="flex flex-col items-center gap-1">
              <Logo item={item} />
              {itemAction?.(item)}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
