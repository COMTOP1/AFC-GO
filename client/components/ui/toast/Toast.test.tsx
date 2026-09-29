import { act, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { ToastProvider } from './ToastProvider';
import { useToast } from './useToast';

function Probe() {
  const toast = useToast();
  return (
    <>
      {['1', '2', '3', '4'].map((n) => (
        <button key={n} onClick={() => toast.show({ message: `Toast ${n}`, tone: 'success' })}>
          show {n}
        </button>
      ))}
    </>
  );
}

function renderProbe() {
  return render(
    <ToastProvider>
      <Probe />
    </ToastProvider>,
  );
}

describe('Toast', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('shows a toast in a polite live region', () => {
    const { container } = renderProbe();
    fireEvent.click(screen.getByRole('button', { name: 'show 1' }));
    const toast = screen.getByRole('button', { name: 'Toast 1' });
    expect(toast.closest('[aria-live="polite"]')).not.toBeNull();
    expect(container.ownerDocument.querySelector('[aria-live="polite"]')).toBeInTheDocument();
  });

  it('dismisses itself after 5 seconds', () => {
    renderProbe();
    fireEvent.click(screen.getByRole('button', { name: 'show 1' }));
    act(() => vi.advanceTimersByTime(4999));
    expect(screen.getByRole('button', { name: 'Toast 1' })).toBeInTheDocument();
    act(() => vi.advanceTimersByTime(1));
    expect(screen.queryByRole('button', { name: 'Toast 1' })).toBeNull();
  });

  it('dismisses on click', () => {
    renderProbe();
    fireEvent.click(screen.getByRole('button', { name: 'show 1' }));
    fireEvent.click(screen.getByRole('button', { name: 'Toast 1' }));
    expect(screen.queryByRole('button', { name: 'Toast 1' })).toBeNull();
  });

  it('keeps at most three, dropping the oldest', () => {
    renderProbe();
    for (const n of ['1', '2', '3', '4']) {
      fireEvent.click(screen.getByRole('button', { name: `show ${n}` }));
    }
    expect(screen.queryByRole('button', { name: 'Toast 1' })).toBeNull();
    for (const n of ['2', '3', '4']) {
      expect(screen.getByRole('button', { name: `Toast ${n}` })).toBeInTheDocument();
    }
  });

  it('throws a clear error outside the provider', () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
    expect(() => render(<Probe />)).toThrow('useToast must be used inside <ToastProvider>');
  });
});
