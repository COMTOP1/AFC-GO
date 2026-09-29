// Keep in step with client/public/theme-init.js (same key, same rules).
export type ThemeSetting = 'system' | 'light' | 'dark';
export type ResolvedTheme = 'light' | 'dark';

export const THEME_STORAGE_KEY = 'afc-theme';
export const DARK_QUERY = '(prefers-color-scheme: dark)';

export const nextSetting: Record<ThemeSetting, ThemeSetting> = {
  system: 'light',
  light: 'dark',
  dark: 'system',
};

export function readSetting(): ThemeSetting {
  try {
    const saved = window.localStorage.getItem(THEME_STORAGE_KEY);
    if (saved === 'light' || saved === 'dark') {
      return saved;
    }
  } catch {
    // Storage blocked: follow the system.
  }
  return 'system';
}

export function writeSetting(setting: ThemeSetting): void {
  try {
    if (setting === 'system') {
      window.localStorage.removeItem(THEME_STORAGE_KEY);
    } else {
      window.localStorage.setItem(THEME_STORAGE_KEY, setting);
    }
  } catch {
    // Storage blocked: the choice lasts for this page only.
  }
}

export function systemPrefersDark(): boolean {
  return typeof window.matchMedia === 'function' && window.matchMedia(DARK_QUERY).matches;
}

export function resolveTheme(setting: ThemeSetting, prefersDark: boolean): ResolvedTheme {
  if (setting === 'system') {
    return prefersDark ? 'dark' : 'light';
  }
  return setting;
}
