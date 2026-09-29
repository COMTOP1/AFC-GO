import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react';

import { ThemeContext, type ThemeState } from './context';
import {
  DARK_QUERY,
  readSetting,
  resolveTheme,
  systemPrefersDark,
  writeSetting,
  type ThemeSetting,
} from './setting';

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [setting, setSettingState] = useState<ThemeSetting>(readSetting);
  const [prefersDark, setPrefersDark] = useState(systemPrefersDark);

  useEffect(() => {
    if (typeof window.matchMedia !== 'function') {
      return;
    }
    const mql = window.matchMedia(DARK_QUERY);
    const onChange = (e: MediaQueryListEvent) => setPrefersDark(e.matches);
    mql.addEventListener('change', onChange);
    return () => mql.removeEventListener('change', onChange);
  }, []);

  const resolved = resolveTheme(setting, prefersDark);

  useEffect(() => {
    document.documentElement.dataset.theme = resolved;
  }, [resolved]);

  const setSetting = useCallback((next: ThemeSetting) => {
    writeSetting(next);
    setSettingState(next);
  }, []);

  const value = useMemo<ThemeState>(
    () => ({ setting, resolved, setSetting }),
    [setting, resolved, setSetting],
  );

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}
