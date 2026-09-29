import { fireEvent, screen } from '@testing-library/react';
import { Link } from 'react-router';
import { beforeEach, describe, expect, it } from 'vitest';

import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import { NavBar } from './NavBar';
import { navItems } from './navItems';

beforeEach(() => {
  mockFetch({
    '/api/v1/auth/me': { status: 401, body: { error: { code: 401, message: 'login required' } } },
  });
});

function renderNav(route = '/') {
  renderWithProviders(
    <>
      <NavBar />
      <Link to="/elsewhere">elsewhere</Link>
    </>,
    { route },
  );
  return screen.getByRole('button', { name: 'Menu' });
}

describe('NavBar', () => {
  it('lists every nav item in order', () => {
    renderNav();
    expect(navItems.map((i) => i.label)).toEqual([
      'Home',
      'Teams',
      'News',
      "What's On",
      'Gallery',
      'Documents',
      'Programmes',
      'Sponsors',
      'Info',
      'Contact',
    ]);
    const nav = screen.getByRole('navigation', { name: 'Main' });
    for (const item of navItems) {
      expect(nav).toContainElement(screen.getByRole('link', { name: item.label }));
    }
  });

  it('uses a router link for ported pages and marks the current one', () => {
    renderNav('/');
    const home = screen.getByRole('link', { name: 'Home' });
    expect(home).toHaveAttribute('href', '/');
    expect(home).toHaveAttribute('aria-current', 'page');
  });

  it('uses a plain legacy link for pages not yet ported', () => {
    renderNav('/');
    const news = screen.getByRole('link', { name: 'News' });
    expect(news).toHaveAttribute('href', '/news');
    expect(news).not.toHaveAttribute('aria-current');
  });

  it('shows the theme toggle and the account control', async () => {
    renderNav();
    expect(screen.getByRole('button', { name: /^Theme:/ })).toBeInTheDocument();
    expect(await screen.findByRole('button', { name: 'Sign in' })).toBeInTheDocument();
  });

  it('opens and closes the phone menu', () => {
    const menu = renderNav();
    expect(menu).toHaveAttribute('aria-expanded', 'false');
    fireEvent.click(menu);
    expect(menu).toHaveAttribute('aria-expanded', 'true');
    expect(document.getElementById(menu.getAttribute('aria-controls') ?? '')).not.toBeNull();
    fireEvent.click(menu);
    expect(menu).toHaveAttribute('aria-expanded', 'false');
  });

  it('closes the phone menu on Esc', () => {
    const menu = renderNav();
    fireEvent.click(menu);
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(menu).toHaveAttribute('aria-expanded', 'false');
  });

  it('closes the phone menu on an outside click', () => {
    const menu = renderNav();
    fireEvent.click(menu);
    fireEvent.mouseDown(document.body);
    expect(menu).toHaveAttribute('aria-expanded', 'false');
  });

  it('closes the phone menu when the route changes', () => {
    const menu = renderNav();
    fireEvent.click(menu);
    fireEvent.click(screen.getByRole('link', { name: 'elsewhere' }));
    expect(menu).toHaveAttribute('aria-expanded', 'false');
  });
});
