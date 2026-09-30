import { useMemo, useState, type ReactNode } from 'react';

import { SignInContext, type SignInApi } from './signInContext';
import { SignInDialog } from './SignInDialog';

/** Owns the one sign-in dialog; anything can open it with useSignIn().open(). */
export function SignInProvider({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false);
  const api = useMemo<SignInApi>(() => ({ open: () => setOpen(true) }), []);
  return (
    <SignInContext.Provider value={api}>
      {children}
      <SignInDialog open={open} onClose={() => setOpen(false)} />
    </SignInContext.Provider>
  );
}
