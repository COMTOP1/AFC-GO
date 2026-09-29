import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { RichText } from './RichText';

describe('RichText', () => {
  it('renders cleaned HTML inside a prose wrapper', () => {
    const { container } = render(
      <RichText html={'<h2>Welcome</h2><p>Hello<script>bad()</script></p>'} />,
    );
    expect(screen.getByRole('heading', { level: 2, name: 'Welcome' })).toBeInTheDocument();
    expect(container.querySelector('script')).toBeNull();
    expect(container.firstElementChild).toHaveClass('prose');
  });
});
