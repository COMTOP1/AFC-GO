import { useEffect } from 'react';

const CLUB = 'AFC Aldermaston';

export function usePageTitle(title?: string): void {
  useEffect(() => {
    document.title = title ? `${title} · ${CLUB}` : CLUB;
  }, [title]);
}
