import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import * as telemetry from '../../lib/telemetry';
import { RouteErrorBoundary } from './RouteErrorBoundary';

function Boom(): never {
  throw new Error('kaboom');
}

describe('RouteErrorBoundary', () => {
  it('reports the error via telemetry', () => {
    const spy = vi.spyOn(telemetry, 'reportEvent').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});

    render(
      <RouteErrorBoundary>
        <Boom />
      </RouteErrorBoundary>,
    );

    expect(spy).toHaveBeenCalledWith(expect.objectContaining({ type: 'error', message: 'kaboom' }));
  });

  it('still shows the fallback UI', () => {
    vi.spyOn(telemetry, 'reportEvent').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});

    render(
      <RouteErrorBoundary>
        <Boom />
      </RouteErrorBoundary>,
    );

    expect(screen.getByText(/couldn't be loaded/i)).toBeInTheDocument();
  });
});
