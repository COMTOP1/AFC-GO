import { useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';

import { logout } from '../../api/auth';
import { ApiError } from '../../api/client';
import { queryKeys } from '../../api/queries';
import { useAuth } from '../../auth/useAuth';
import { Button } from '../ui/Button';
import { buttonClasses } from '../ui/buttonStyles';
import { ConfirmDialog } from '../ui/ConfirmDialog';
import { Menu, type MenuItem } from '../ui/Menu';
import { useToast } from '../ui/toast/useToast';
import { useDisclosure } from '../ui/useDisclosure';
import { SignInDialog } from './SignInDialog';

export function AccountControl() {
  const { user, isLoading } = useAuth();
  const signIn = useDisclosure();
  const [confirmSignOut, setConfirmSignOut] = useState(false);
  const queryClient = useQueryClient();
  const toast = useToast();

  if (isLoading) {
    return <span aria-hidden="true" className="inline-block h-8 w-20" />;
  }

  if (!user) {
    return (
      <>
        <Button variant="secondary" size="sm" onClick={() => signIn.setOpen(true)}>
          Sign in
        </Button>
        <SignInDialog open={signIn.open} onClose={signIn.close} />
      </>
    );
  }

  async function signOut() {
    try {
      await logout();
    } catch (err) {
      // A 401 means the session had already expired: signed out either way.
      if (!(err instanceof ApiError && err.status === 401)) {
        setConfirmSignOut(false);
        toast.show({
          tone: 'error',
          message: `Couldn't sign out: ${err instanceof Error ? err.message : 'unknown error'}`,
        });
        return;
      }
    }
    setConfirmSignOut(false);
    // Not queryClient.clear(): that strands mounted observers on removed queries.
    queryClient.setQueryData(queryKeys.me, null);
    await queryClient.invalidateQueries();
  }

  const items: MenuItem[] = [
    { label: 'Players', href: '/players' },
    { label: 'Account', href: '/account' },
    ...(user.permissions.canEdit ? [{ label: 'Edit info', href: '/info/edit' }] : []),
    ...(user.permissions.canManageUsers ? [{ label: 'Users', href: '/users' }] : []),
    { label: 'Sign out', onSelect: () => setConfirmSignOut(true) },
  ];

  return (
    <>
      <Menu
        label={
          <>
            <span className="max-w-40 truncate">{user.name}</span>
            <span aria-hidden="true">▾</span>
          </>
        }
        triggerLabel={user.name}
        items={items}
        triggerClassName={buttonClasses('secondary', 'sm')}
      />
      <ConfirmDialog
        open={confirmSignOut}
        title="Sign out?"
        message="You'll need to sign in again to make changes."
        confirmLabel="Sign out"
        onConfirm={signOut}
        onCancel={() => setConfirmSignOut(false)}
      />
    </>
  );
}
