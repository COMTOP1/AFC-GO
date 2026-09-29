import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { usePageTitle } from './usePageTitle';

function Probe({ title }: { title?: string }) {
  usePageTitle(title);
  return null;
}

describe('usePageTitle', () => {
  it('sets a page title with the club suffix', () => {
    render(<Probe title="News" />);
    expect(document.title).toBe('News · AFC Aldermaston');
  });

  it('uses the bare club name without a title', () => {
    render(<Probe />);
    expect(document.title).toBe('AFC Aldermaston');
  });
});
