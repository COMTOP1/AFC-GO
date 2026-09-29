import { createContext } from 'react';

import type { CurrentUser } from '../api/types';

export interface AuthState {
  user: CurrentUser | null;
  isLoading: boolean;
  /** Re-checks who is signed in (after login/logout in sub-project 4). */
  refresh: () => Promise<void>;
}

export const AuthContext = createContext<AuthState | null>(null);
