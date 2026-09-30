import { useContext } from 'react';

import { SignInContext, type SignInApi } from './signInContext';

export function useSignIn(): SignInApi {
  const ctx = useContext(SignInContext);
  if (!ctx) {
    throw new Error('useSignIn must be used inside <SignInProvider>');
  }
  return ctx;
}
