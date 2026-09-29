import { clsx } from 'clsx';
import { useId, useRef, useState } from 'react';
import { matchPath, NavLink, useLocation } from 'react-router';

import { ThemeToggle } from '../../theme/ThemeToggle';
import { buttonClasses } from '../ui/buttonStyles';
import { useDisclosure } from '../ui/useDisclosure';
import { useDismiss } from '../ui/useDismiss';
import { AccountControl } from './AccountControl';
import { navItems, type NavItem } from './navItems';

function NavItemLink({ item, className }: { item: NavItem; className: string }) {
  if (item.to !== undefined) {
    return (
      <NavLink
        to={item.to}
        end
        className={({ isActive }) =>
          clsx(className, isActive && 'text-red shadow-[inset_0_-3px_0_var(--color-red)]')
        }
      >
        {item.label}
      </NavLink>
    );
  }
  return (
    <a href={item.legacyHref} className={className}>
      {item.label}
    </a>
  );
}

export function NavBar() {
  const { pathname } = useLocation();
  const menu = useDisclosure();
  const barRef = useRef<HTMLDivElement>(null);
  const panelId = useId();
  useDismiss(barRef, menu.open, menu.close);

  // Close the phone menu whenever the route changes (adjusting state during render).
  const [lastPath, setLastPath] = useState(pathname);
  if (pathname !== lastPath) {
    setLastPath(pathname);
    menu.close();
  }

  const current = navItems.find(
    (item) => item.to !== undefined && matchPath({ path: item.to, end: true }, pathname),
  );

  return (
    <nav aria-label="Main" className="border-t border-b-3 border-line border-b-red">
      <div ref={barRef} className="flex flex-wrap items-center justify-end gap-x-5 px-3.5 md:px-7">
        <span className="mr-auto py-2 text-sm font-bold text-red md:hidden">{current?.label}</span>
        <ul
          id={panelId}
          className={clsx(
            'text-sm font-semibold md:flex md:items-center md:gap-5',
            menu.open
              ? 'max-md:order-last max-md:w-full max-md:border-t max-md:border-line max-md:py-1'
              : 'max-md:hidden',
          )}
        >
          {navItems.map((item) => (
            <li key={item.label}>
              <NavItemLink item={item} className="block py-2.5 hover:text-red md:py-3" />
            </li>
          ))}
        </ul>
        <div className="flex items-center gap-2 py-2">
          <ThemeToggle />
          <AccountControl />
          <button
            type="button"
            aria-expanded={menu.open}
            aria-controls={panelId}
            onClick={menu.toggle}
            className={buttonClasses('secondary', 'sm', 'md:hidden')}
          >
            Menu
          </button>
        </div>
      </div>
    </nav>
  );
}
