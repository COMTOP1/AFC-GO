import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router';
import { describe, expect, it } from 'vitest';

import { SearchInput } from './SearchInput';

describe('SearchInput under BrowserRouter', () => {
  it('keeps what was typed (and the caret) while the address catches up', async () => {
    window.history.replaceState(null, '', '/documents?q=abcd');
    render(
      <QueryClientProvider client={new QueryClient()}>
        <BrowserRouter>
          <SearchInput label="Search documents" />
        </BrowserRouter>
      </QueryClientProvider>,
    );
    const box = screen.getByLabelText('Search documents') as HTMLInputElement;
    expect(box).toHaveValue('abcd');
    box.setSelectionRange(2, 2);
    fireEvent.change(box, { target: { value: 'abXcd', selectionStart: 3, selectionEnd: 3 } });
    // BrowserRouter commits location updates in a transition; the box must not
    // snap back to the old value in the meantime.
    expect(box).toHaveValue('abXcd');
    await waitFor(() => expect(window.location.search).toBe('?q=abXcd'));
    expect(box).toHaveValue('abXcd');
    expect(box.selectionStart).toBe(3);
    window.history.replaceState(null, '', '/');
  });
});
