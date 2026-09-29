import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { installMatchMedia } from '../test/matchMedia';
import { ThemeProvider } from './ThemeProvider';
import { useTheme } from './useTheme';

function Probe() {
  const { setting, resolved, setSetting } = useTheme();
  return (
    <>
      <p>setting:{setting}</p>
      <p>resolved:{resolved}</p>
      <button onClick={() => setSetting('dark')}>dark</button>
      <button onClick={() => setSetting('system')}>system</button>
    </>
  );
}

function renderProbe() {
  return render(
    <ThemeProvider>
      <Probe />
    </ThemeProvider>,
  );
}

describe('ThemeProvider', () => {
  it('defaults to system and resolves light on a light OS', () => {
    installMatchMedia(false);
    renderProbe();
    expect(screen.getByText('setting:system')).toBeInTheDocument();
    expect(document.documentElement.dataset.theme).toBe('light');
  });

  it('resolves dark on a dark OS', () => {
    installMatchMedia(true);
    renderProbe();
    expect(screen.getByText('resolved:dark')).toBeInTheDocument();
    expect(document.documentElement.dataset.theme).toBe('dark');
  });

  it('applies a stored setting', () => {
    installMatchMedia(false);
    localStorage.setItem('afc-theme', 'dark');
    renderProbe();
    expect(screen.getByText('setting:dark')).toBeInTheDocument();
    expect(document.documentElement.dataset.theme).toBe('dark');
  });

  it('follows OS changes live while on system', () => {
    const media = installMatchMedia(false);
    renderProbe();
    act(() => media.setDark(true));
    expect(document.documentElement.dataset.theme).toBe('dark');
  });

  it('persists a chosen setting and forgets it again for system', () => {
    installMatchMedia(false);
    renderProbe();
    fireEvent.click(screen.getByRole('button', { name: 'dark' }));
    expect(localStorage.getItem('afc-theme')).toBe('dark');
    expect(document.documentElement.dataset.theme).toBe('dark');
    fireEvent.click(screen.getByRole('button', { name: 'system' }));
    expect(localStorage.getItem('afc-theme')).toBeNull();
  });

  it('still renders when storage throws', () => {
    installMatchMedia(false);
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('blocked');
    });
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('blocked');
    });
    renderProbe();
    expect(screen.getByText('setting:system')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'dark' }));
    expect(screen.getByText('setting:dark')).toBeInTheDocument();
  });

  it('resolves light when matchMedia is missing', () => {
    vi.stubGlobal('matchMedia', undefined);
    renderProbe();
    expect(screen.getByText('resolved:light')).toBeInTheDocument();
  });
});
