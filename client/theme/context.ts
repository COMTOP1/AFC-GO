import { createContext } from 'react';

import type { ResolvedTheme, ThemeSetting } from './setting';

export interface ThemeState {
  setting: ThemeSetting;
  resolved: ResolvedTheme;
  setSetting: (setting: ThemeSetting) => void;
}

export const ThemeContext = createContext<ThemeState | null>(null);
