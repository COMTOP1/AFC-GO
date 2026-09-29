import { vi } from 'vitest';

export const DARK_QUERY = '(prefers-color-scheme: dark)';

export interface MediaController {
  /** Flips the OS preference and notifies every change listener. */
  setDark(dark: boolean): void;
}

/** Stubs window.matchMedia; only the dark-scheme query ever matches. */
export function installMatchMedia(initialDark = false): MediaController {
  let dark = initialDark;
  const listeners = new Set<(e: MediaQueryListEvent) => void>();
  vi.stubGlobal(
    'matchMedia',
    vi.fn((query: string) => ({
      get matches() {
        return query === DARK_QUERY && dark;
      },
      media: query,
      onchange: null,
      addEventListener: (_type: string, fn: (e: MediaQueryListEvent) => void) => listeners.add(fn),
      removeEventListener: (_type: string, fn: (e: MediaQueryListEvent) => void) =>
        listeners.delete(fn),
      addListener: (fn: (e: MediaQueryListEvent) => void) => listeners.add(fn),
      removeListener: (fn: (e: MediaQueryListEvent) => void) => listeners.delete(fn),
      dispatchEvent: () => false,
    })),
  );
  return {
    setDark(next: boolean) {
      dark = next;
      listeners.forEach((fn) => fn({ matches: next, media: DARK_QUERY } as MediaQueryListEvent));
    },
  };
}
