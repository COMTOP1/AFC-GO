import { fireEvent, screen } from '@testing-library/react';
import { Link, Route, Routes } from 'react-router';
import { describe, expect, it, vi } from 'vitest';

import { publicRoutes } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import Layout from './Layout';

function Boom(): never {
  throw new Error('chunk failed to load');
}

function renderLayout(route = '/one') {
  mockFetch(publicRoutes());
  renderWithProviders(
    <Routes>
      <Route element={<Layout />}>
        <Route
          path="one"
          element={
            <>
              <h1>One</h1>
              <Link to="/two">to two</Link>
              <Link to="?tab=x">same page, new tab</Link>
              <Link to="/broken">to broken</Link>
            </>
          }
        />
        <Route
          path="two"
          element={
            <>
              <h1>Two</h1>
              <Link to="/one">to one</Link>
            </>
          }
        />
        <Route path="broken" element={<Boom />} />
      </Route>
    </Routes>,
    { route },
  );
}

describe('Layout', () => {
  it('scrolls to the top and focuses the content on a new page, but not on a search change', () => {
    const scrollTo = vi.fn();
    vi.stubGlobal('scrollTo', scrollTo);
    renderLayout();
    expect(scrollTo).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('link', { name: 'same page, new tab' }));
    expect(scrollTo).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('link', { name: 'to two' }));
    expect(screen.getByRole('heading', { name: 'Two' })).toBeInTheDocument();
    expect(scrollTo).toHaveBeenCalledWith(0, 0);
    expect(document.activeElement).toBe(screen.getByRole('main'));
  });

  it('keeps the shell and explains when a page fails to render', () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
    vi.stubGlobal('scrollTo', vi.fn());
    renderLayout();
    fireEvent.click(screen.getByRole('link', { name: 'to broken' }));
    expect(screen.getByRole('alert')).toHaveTextContent("Sorry, this page couldn't be loaded");
    expect(screen.getByRole('button', { name: 'Reload' })).toBeInTheDocument();
    expect(screen.getByRole('navigation', { name: 'Main' })).toBeInTheDocument();
  });
});
