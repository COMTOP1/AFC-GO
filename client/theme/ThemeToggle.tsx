import type { ReactNode } from 'react';

import { buttonClasses } from '../components/ui/buttonStyles';
import { nextSetting, type ThemeSetting } from './setting';
import { useTheme } from './useTheme';

const iconProps = {
  width: 18,
  height: 18,
  viewBox: '0 0 24 24',
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 2,
  strokeLinecap: 'round' as const,
  strokeLinejoin: 'round' as const,
  'aria-hidden': true,
};

const icons: Record<ThemeSetting, ReactNode> = {
  system: (
    <svg {...iconProps}>
      <rect x="3" y="4" width="18" height="12" rx="2" />
      <path d="M8 20h8M12 16v4" />
    </svg>
  ),
  light: (
    <svg {...iconProps}>
      <circle cx="12" cy="12" r="4" />
      <path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" />
    </svg>
  ),
  dark: (
    <svg {...iconProps}>
      <path d="M20 14.5A8 8 0 0 1 9.5 4a8 8 0 1 0 10.5 10.5z" />
    </svg>
  ),
};

export function ThemeToggle() {
  const { setting, setSetting } = useTheme();
  const next = nextSetting[setting];
  return (
    <button
      type="button"
      onClick={() => setSetting(next)}
      aria-label={`Theme: ${setting} (switch to ${next})`}
      title={`Theme: ${setting}`}
      className={buttonClasses('secondary', 'sm', 'px-2')}
    >
      {icons[setting]}
    </button>
  );
}
