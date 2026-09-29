import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ThemeProvider } from './ThemeProvider';
import { ThemeToggle } from './ThemeToggle';

describe('ThemeToggle', () => {
  it('cycles system → light → dark → system and remembers the choice', () => {
    render(
      <ThemeProvider>
        <ThemeToggle />
      </ThemeProvider>,
    );
    const button = () => screen.getByRole('button', { name: /^Theme:/ });
    expect(button()).toHaveAccessibleName('Theme: system (switch to light)');

    fireEvent.click(button());
    expect(button()).toHaveAccessibleName('Theme: light (switch to dark)');
    expect(localStorage.getItem('afc-theme')).toBe('light');

    fireEvent.click(button());
    expect(button()).toHaveAccessibleName('Theme: dark (switch to system)');
    expect(document.documentElement.dataset.theme).toBe('dark');

    fireEvent.click(button());
    expect(button()).toHaveAccessibleName('Theme: system (switch to light)');
    expect(localStorage.getItem('afc-theme')).toBeNull();
  });
});
