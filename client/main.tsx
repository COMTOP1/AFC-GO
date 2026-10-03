import { QueryClientProvider } from '@tanstack/react-query';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router';

import './styles/app.css';

import { createQueryClient } from './api/queryClient';
import App from './App';
import { AuthProvider } from './auth/AuthProvider';
import { SignInProvider } from './components/layout/SignInProvider';
import { ToastProvider } from './components/ui/toast/ToastProvider';
import { reportEvent } from './lib/telemetry';
import { ThemeProvider } from './theme/ThemeProvider';

// Errors outside the React tree (async code, event handlers) never reach
// RouteErrorBoundary, so they're reported here instead.
window.addEventListener('error', (e) => {
  reportEvent({
    type: 'error',
    name: 'window.onerror',
    startTime: Date.now(),
    message: e.error instanceof Error ? e.error.message : e.message,
    stack: e.error instanceof Error ? e.error.stack : undefined,
  });
});
window.addEventListener('unhandledrejection', (e) => {
  reportEvent({
    type: 'error',
    name: 'unhandledrejection',
    startTime: Date.now(),
    message: e.reason instanceof Error ? e.reason.message : String(e.reason),
    stack: e.reason instanceof Error ? e.reason.stack : undefined,
  });
});

const root = document.getElementById('root');
if (!root) {
  throw new Error('missing #root element');
}

const queryClient = createQueryClient();

createRoot(root).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <ThemeProvider>
          <AuthProvider>
            <ToastProvider>
              <SignInProvider>
                <App />
              </SignInProvider>
            </ToastProvider>
          </AuthProvider>
        </ThemeProvider>
      </BrowserRouter>
    </QueryClientProvider>
  </StrictMode>,
);
