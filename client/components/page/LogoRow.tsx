import { useId } from 'react';

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

/** A titled row of sponsor/affiliation logos; renders nothing when empty. */
export function LogoRow({ title, items }: { title: string; items: LogoItem[] }) {
  const headingId = useId();
  if (items.length === 0) {
    return null;
  }
  return (
    <section aria-labelledby={headingId}>
      <h2
        id={headingId}
        className="mb-3 font-display text-2xl font-extrabold tracking-wide uppercase"
      >
        {title}
      </h2>
      <ul className="flex flex-wrap gap-3">
        {items.map((item) => (
          <li key={item.id}>
            <Logo item={item} />
          </li>
        ))}
      </ul>
    </section>
  );
}
