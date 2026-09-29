import { afterEach, describe, expect, it, vi } from 'vitest';

import src from '../public/theme-init.js?raw';
import { installMatchMedia } from '../test/matchMedia';

function runInit() {
  new Function(src)();
}

describe('theme-init.js', () => {
  afterEach(() => {
    localStorage.clear();
    delete document.documentElement.dataset.theme;
  });

  it('uses a stored dark setting', () => {
    installMatchMedia(false);
    localStorage.setItem('afc-theme', 'dark');
    runInit();
    expect(document.documentElement.dataset.theme).toBe('dark');
  });

  it('uses a stored light setting even when the system is dark', () => {
    installMatchMedia(true);
    localStorage.setItem('afc-theme', 'light');
    runInit();
    expect(document.documentElement.dataset.theme).toBe('light');
  });

  it('follows the system when nothing is stored', () => {
    installMatchMedia(true);
    runInit();
    expect(document.documentElement.dataset.theme).toBe('dark');
  });

  it('ignores junk in storage', () => {
    installMatchMedia(false);
    localStorage.setItem('afc-theme', 'purple');
    runInit();
    expect(document.documentElement.dataset.theme).toBe('light');
  });

  it('falls back to the system when storage throws', () => {
    installMatchMedia(true);
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('blocked');
    });
    runInit();
    expect(document.documentElement.dataset.theme).toBe('dark');
  });

  it('chooses light when matchMedia is missing', () => {
    vi.stubGlobal('matchMedia', undefined);
    runInit();
    expect(document.documentElement.dataset.theme).toBe('light');
  });
});
