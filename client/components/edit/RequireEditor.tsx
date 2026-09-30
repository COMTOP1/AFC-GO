import type { ReactNode } from 'react';
import { useNavigate } from 'react-router';

import { useAuth } from '../../auth/useAuth';
import { PageSkeleton } from '../page/QueryState';
import { Button } from '../ui/Button';
import { EmptyState } from '../ui/EmptyState';

export interface RequireEditorProps {
  permission?: 'canEdit' | 'canManageGallery' | 'canManageUsers';
  children: ReactNode;
}

/** Gate for edit routes: loading placeholder, a permission message, or the form. */
export function RequireEditor({ permission = 'canEdit', children }: RequireEditorProps) {
  const { user, isLoading } = useAuth();
  const navigate = useNavigate();
  if (isLoading) {
    return <PageSkeleton />;
  }
  if (!user?.permissions[permission]) {
    return (
      <EmptyState
        title="You don't have permission to edit this"
        message="Sign in with an editor account, or go back."
        action={
          <Button
            variant="secondary"
            onClick={() => (window.history.length > 1 ? navigate(-1) : navigate('/'))}
          >
            Back
          </Button>
        }
      />
    );
  }
  return <>{children}</>;
}
