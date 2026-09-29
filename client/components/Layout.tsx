import { Outlet } from 'react-router';

// Bare page shell. Visual design is sub-project 3.
export default function Layout() {
  return (
    <>
      <header>
        <a href="/">AFC Aldermaston</a>
      </header>
      <main>
        <Outlet />
      </main>
      <footer>© {new Date().getFullYear()} AFC Aldermaston</footer>
    </>
  );
}
