import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ImageWithFallback } from './ImageWithFallback';

describe('ImageWithFallback', () => {
  it('shows the image while it loads fine', () => {
    render(<ImageWithFallback src="/logo.png" alt="Acme" fallback={<span>Acme tile</span>} />);
    expect(screen.getByRole('img', { name: 'Acme' })).toHaveAttribute('src', '/logo.png');
    expect(screen.queryByText('Acme tile')).toBeNull();
  });

  it('swaps to the fallback node when the image fails', () => {
    render(<ImageWithFallback src="/gone.png" alt="Acme" fallback={<span>Acme tile</span>} />);
    fireEvent.error(screen.getByRole('img', { name: 'Acme' }));
    expect(screen.queryByRole('img')).toBeNull();
    expect(screen.getByText('Acme tile')).toBeInTheDocument();
  });

  it('swaps to a fallback image when given one', () => {
    const { container } = render(
      <ImageWithFallback src="/gone.png" fallbackSrc="/crest.png" alt="" />,
    );
    fireEvent.error(container.querySelector('img') as HTMLImageElement);
    expect(container.querySelector('img')).toHaveAttribute('src', '/crest.png');
  });

  it('uses the fallback straight away without a src', () => {
    render(<ImageWithFallback alt="Acme" fallback={<span>Acme tile</span>} />);
    expect(screen.getByText('Acme tile')).toBeInTheDocument();
  });
});
