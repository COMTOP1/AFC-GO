import { useLocation } from 'react-router';

/** Test helper: shows the router's current path so tests can assert navigation. */
export function Location() {
  const l = useLocation();
  return <output data-testid="location">{l.pathname + l.search}</output>;
}
