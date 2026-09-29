import { clsx } from 'clsx';
import { useEffect, useId, useRef, type KeyboardEvent, type ReactNode } from 'react';
import { Link } from 'react-router';

import { buttonClasses } from './buttonStyles';
import { useDisclosure } from './useDisclosure';
import { useDismiss } from './useDismiss';

export interface MenuItem {
  label: string;
  /** An SPA route. */
  to?: string;
  /** A full-page URL (legacy page). */
  href?: string;
  onSelect?: () => void;
}

export interface MenuProps {
  label: ReactNode;
  items: MenuItem[];
  align?: 'start' | 'end';
  triggerClassName?: string;
  /** Accessible name when `label` is not plain text. */
  triggerLabel?: string;
}

const itemClass =
  'block w-full px-3 py-2 text-left text-sm text-ink hover:bg-surface focus-visible:bg-surface focus-visible:outline-none';

export function Menu({ label, items, align = 'end', triggerClassName, triggerLabel }: MenuProps) {
  const { open, setOpen, close } = useDisclosure();
  const rootRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const menuId = useId();

  useDismiss(rootRef, open, close);

  useEffect(() => {
    if (open) {
      listRef.current?.querySelector<HTMLElement>('[role="menuitem"]')?.focus();
    }
  }, [open]);

  function onKeyDown(e: KeyboardEvent<HTMLDivElement>) {
    const els = Array.from(
      listRef.current?.querySelectorAll<HTMLElement>('[role="menuitem"]') ?? [],
    );
    const i = els.indexOf(document.activeElement as HTMLElement);
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      els[(i + 1) % els.length]?.focus();
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      els[i <= 0 ? els.length - 1 : i - 1]?.focus();
    } else if (e.key === 'Escape') {
      e.preventDefault();
      close();
      triggerRef.current?.focus();
    } else if (e.key === 'Tab') {
      close();
    }
  }

  return (
    <div ref={rootRef} className="relative">
      <button
        ref={triggerRef}
        type="button"
        aria-haspopup="menu"
        aria-expanded={open}
        aria-controls={open ? menuId : undefined}
        aria-label={triggerLabel}
        onClick={() => setOpen(!open)}
        className={triggerClassName ?? buttonClasses('secondary', 'sm')}
      >
        {label}
      </button>
      {open && (
        <div
          ref={listRef}
          id={menuId}
          role="menu"
          onKeyDown={onKeyDown}
          className={clsx(
            'absolute top-full z-40 mt-1 min-w-44 overflow-hidden rounded-lg border border-line bg-bg py-1 shadow-lg',
            align === 'end' ? 'right-0' : 'left-0',
          )}
        >
          {items.map((item) => {
            const onClick = () => {
              close();
              item.onSelect?.();
            };
            if (item.to !== undefined) {
              return (
                <Link
                  key={item.label}
                  role="menuitem"
                  tabIndex={-1}
                  to={item.to}
                  onClick={onClick}
                  className={itemClass}
                >
                  {item.label}
                </Link>
              );
            }
            if (item.href !== undefined) {
              return (
                <a
                  key={item.label}
                  role="menuitem"
                  tabIndex={-1}
                  href={item.href}
                  onClick={onClick}
                  className={itemClass}
                >
                  {item.label}
                </a>
              );
            }
            return (
              <button
                key={item.label}
                type="button"
                role="menuitem"
                tabIndex={-1}
                onClick={onClick}
                className={itemClass}
              >
                {item.label}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
