import { Outlet } from 'react-router';

import { Container } from '../ui/Container';
import { Footer } from './Footer';
import { Masthead } from './Masthead';
import { NavBar } from './NavBar';
import { SkipLink } from './SkipLink';

export default function Layout() {
  return (
    <div className="flex min-h-screen flex-col bg-bg text-ink">
      <SkipLink />
      <header>
        <Masthead />
        <NavBar />
      </header>
      <main id="content" tabIndex={-1} className="flex-1 py-8 outline-none">
        <Container>
          <Outlet />
        </Container>
      </main>
      <Footer />
    </div>
  );
}
