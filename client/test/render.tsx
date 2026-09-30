import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render } from '@testing-library/react';
import type { ReactElement } from 'react';
import { MemoryRouter } from 'react-router';

import { AuthProvider } from '../auth/AuthProvider';
import { SignInProvider } from '../components/layout/SignInProvider';
import { ToastProvider } from '../components/ui/toast/ToastProvider';
import { ThemeProvider } from '../theme/ThemeProvider';

export interface RenderOptions {
  route?: string;
  queryClient?: QueryClient;
}

/** Renders ui inside the same providers as main.tsx, with retries off. */
export function renderWithProviders(
  ui: ReactElement,
  {
    route = '/',
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false, gcTime: Infinity } },
    }),
  }: RenderOptions = {},
) {
  const wrap = (node: ReactElement) => (
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[route]}>
        <ThemeProvider>
          <AuthProvider>
            <ToastProvider>
              <SignInProvider>{node}</SignInProvider>
            </ToastProvider>
          </AuthProvider>
        </ThemeProvider>
      </MemoryRouter>
    </QueryClientProvider>
  );
  const result = render(wrap(ui));
  return { ...result, queryClient, rerender: (next: ReactElement) => result.rerender(wrap(next)) };
}
