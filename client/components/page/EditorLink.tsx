import { useAuth } from '../../auth/useAuth';

export interface EditorLinkProps {
  /** The legacy page whose edit controls cover this content. */
  legacyHref: string;
  permission?: 'canEdit' | 'canManageGallery';
}

/** Until sub-project 4c, editors manage content on the legacy pages. */
export function EditorLink({ legacyHref, permission = 'canEdit' }: EditorLinkProps) {
  const { user } = useAuth();
  if (!user?.permissions[permission]) {
    return null;
  }
  return (
    <a href={legacyHref} className="text-sm font-semibold text-red hover:underline">
      Manage this on the classic site ↗
    </a>
  );
}
