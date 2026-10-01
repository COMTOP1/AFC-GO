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
import { ThemeProvider } from './theme/ThemeProvider';

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
