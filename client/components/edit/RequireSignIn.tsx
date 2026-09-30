import type { ReactNode } from 'react';

import { useAuth } from '../../auth/useAuth';
import { useSignIn } from '../layout/useSignIn';
import { PageSkeleton } from '../page/QueryState';
import { Button } from '../ui/Button';
import { EmptyState } from '../ui/EmptyState';

/** Gate for member-only pages: loading placeholder, a sign-in prompt, or the page. */
export function RequireSignIn({ title, children }: { title: string; children: ReactNode }) {
  const { user, isLoading } = useAuth();
  const signIn = useSignIn();
  if (isLoading) {
    return <PageSkeleton />;
  }
  if (!user) {
    return (
      <EmptyState
        title={title}
        message="Your session may have expired."
        action={<Button onClick={signIn.open}>Sign in</Button>}
      />
    );
  }
  return <>{children}</>;
}
