import { useAuth } from '../../auth/useAuth';

/** What the signed-in user may edit (all false when signed out). The server enforces the same rules. */
export function useCanEdit(): { canEdit: boolean; canManageGallery: boolean } {
  const { user } = useAuth();
  return {
    canEdit: user?.permissions.canEdit ?? false,
    canManageGallery: user?.permissions.canManageGallery ?? false,
  };
}
