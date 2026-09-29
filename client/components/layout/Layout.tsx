import { Suspense, useEffect, useRef } from 'react';
import { Outlet, useLocation, useNavigationType } from 'react-router';

import { PageSkeleton } from '../page/QueryState';
import { Container } from '../ui/Container';
import { Footer } from './Footer';
import { Masthead } from './Masthead';
import { NavBar } from './NavBar';
import { RouteErrorBoundary } from './RouteErrorBoundary';
import { SkipLink } from './SkipLink';

/** On a click-through to a new page, start at the top and move focus to the content. */
function useNewPageReset(pathname: string) {
  const navigationType = useNavigationType();
  const lastPath = useRef(pathname);
  useEffect(() => {
    if (pathname === lastPath.current) {
      return; // search-only changes (tabs, filters) keep their place
    }
    lastPath.current = pathname;
    if (navigationType === 'PUSH') {
      window.scrollTo(0, 0);
      document.getElementById('content')?.focus({ preventScroll: true });
    }
  }, [pathname, navigationType]);
}

export default function Layout() {
  const { pathname } = useLocation();
  useNewPageReset(pathname);
  return (
    <div className="flex min-h-screen flex-col bg-bg text-ink">
      <SkipLink />
      <header>
        <Masthead />
        <NavBar />
      </header>
      <main id="content" tabIndex={-1} className="flex-1 py-8 outline-none">
        <Container>
          <RouteErrorBoundary key={pathname}>
            <Suspense fallback={<PageSkeleton />}>
              <Outlet />
            </Suspense>
          </RouteErrorBoundary>
        </Container>
      </main>
      <Footer />
    </div>
  );
}
