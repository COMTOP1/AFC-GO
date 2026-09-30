import { createContext } from 'react';

export interface SignInApi {
  /** Opens the app-wide sign-in dialog. */
  open: () => void;
}

export const SignInContext = createContext<SignInApi | null>(null);
