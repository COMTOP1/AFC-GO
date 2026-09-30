import { clsx } from 'clsx';
import { Link, useSearchParams } from 'react-router';

import { useTabParam } from './useTabParam';

export interface Tab {
  value: string;
  label: string;
}

export interface TabsNavProps {
  param: string;
  tabs: Tab[];
  defaultValue: string;
  /** Accessible name of the tab navigation. */
  label: string;
}

export function TabsNav({ param, tabs, defaultValue, label }: TabsNavProps) {
  const [params] = useSearchParams();
  const current = useTabParam(
    param,
    tabs.map((t) => t.value),
    defaultValue,
  );

  return (
    <nav aria-label={label} className="mb-5 flex gap-1 border-b border-line">
      {tabs.map((tab) => {
        const next = new URLSearchParams(params);
        next.set(param, tab.value);
        const active = tab.value === current;
        return (
          <Link
            key={tab.value}
            to={`?${next.toString()}`}
            aria-current={active ? 'page' : undefined}
            className={clsx(
              '-mb-px border-b-3 px-3 py-2 text-sm font-semibold',
              active ? 'border-red text-red' : 'border-transparent text-muted hover:text-ink',
            )}
          >
            {tab.label}
          </Link>
        );
      })}
    </nav>
  );
}
