import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { describe, expect, it, vi } from 'vitest';

import { Menu } from './Menu';

function renderMenu(onSelect = vi.fn()) {
  render(
    <MemoryRouter>
      <Menu
        label="Jo Smith"
        items={[
          { label: 'Players', href: '/players' },
          { label: 'Design', to: '/design' },
          { label: 'Sign out', onSelect },
        ]}
      />
      <p>Outside</p>
    </MemoryRouter>,
  );
  return { trigger: screen.getByRole('button', { name: 'Jo Smith' }), onSelect };
}

describe('Menu', () => {
  it('starts closed', () => {
    const { trigger } = renderMenu();
    expect(trigger).toHaveAttribute('aria-haspopup', 'menu');
    expect(trigger).toHaveAttribute('aria-expanded', 'false');
    expect(screen.queryByRole('menu')).toBeNull();
  });

  it('opens and focuses the first item', () => {
    const { trigger } = renderMenu();
    fireEvent.click(trigger);
    expect(trigger).toHaveAttribute('aria-expanded', 'true');
    expect(screen.getByRole('menu')).toBeInTheDocument();
    expect(document.activeElement).toBe(screen.getByRole('menuitem', { name: 'Players' }));
  });

  it('renders legacy links, router links and buttons', () => {
    const { trigger } = renderMenu();
    fireEvent.click(trigger);
    expect(screen.getByRole('menuitem', { name: 'Players' })).toHaveAttribute('href', '/players');
    expect(screen.getByRole('menuitem', { name: 'Design' })).toHaveAttribute('href', '/design');
    expect(screen.getByRole('menuitem', { name: 'Sign out' }).tagName).toBe('BUTTON');
  });

  it('moves focus with the arrow keys and wraps', () => {
    const { trigger } = renderMenu();
    fireEvent.click(trigger);
    const menu = screen.getByRole('menu');
    fireEvent.keyDown(menu, { key: 'ArrowDown' });
    expect(document.activeElement).toBe(screen.getByRole('menuitem', { name: 'Design' }));
    fireEvent.keyDown(menu, { key: 'ArrowUp' });
    fireEvent.keyDown(menu, { key: 'ArrowUp' });
    expect(document.activeElement).toBe(screen.getByRole('menuitem', { name: 'Sign out' }));
  });

  it('closes on Esc and returns focus to the trigger', () => {
    const { trigger } = renderMenu();
    fireEvent.click(trigger);
    fireEvent.keyDown(screen.getByRole('menu'), { key: 'Escape' });
    expect(screen.queryByRole('menu')).toBeNull();
    expect(document.activeElement).toBe(trigger);
  });

  it('closes on an outside click', () => {
    const { trigger } = renderMenu();
    fireEvent.click(trigger);
    fireEvent.mouseDown(screen.getByText('Outside'));
    expect(screen.queryByRole('menu')).toBeNull();
  });

  it('runs onSelect and closes', () => {
    const { trigger, onSelect } = renderMenu();
    fireEvent.click(trigger);
    fireEvent.click(screen.getByRole('menuitem', { name: 'Sign out' }));
    expect(onSelect).toHaveBeenCalledOnce();
    expect(screen.queryByRole('menu')).toBeNull();
  });

  it('keeps a visible focus outline on items', () => {
    const { trigger } = renderMenu();
    fireEvent.click(trigger);
    for (const item of screen.getAllByRole('menuitem')) {
      expect(item).not.toHaveClass('focus-visible:outline-none');
      expect(item).toHaveClass('focus-visible:-outline-offset-3');
    }
  });

  it('moves focus back to the trigger before running onSelect', () => {
    let focusedDuringSelect: Element | null = null;
    const { trigger } = renderMenu(
      vi.fn(() => {
        focusedDuringSelect = document.activeElement;
      }),
    );
    fireEvent.click(trigger);
    fireEvent.click(screen.getByRole('menuitem', { name: 'Sign out' }));
    expect(focusedDuringSelect).toBe(trigger);
  });
});
