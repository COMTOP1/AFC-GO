// Hand-written mirrors of the Go API's JSON (server/internal/*/types.go).

export interface ErrorEnvelope {
  error: {
    code: number;
    message: string;
    fields?: Record<string, string>;
  };
}

/** team.Public */
export interface TeamSummary {
  id: number;
  name: string;
  description?: string;
  league?: string;
  division?: string;
  leagueTableUrl?: string;
  fixturesUrl?: string;
  coach?: string;
  physio?: string;
  imageUrl?: string;
  isActive: boolean;
  isYouth: boolean;
  ages: number;
}

/** site.Info — GET /site */
export interface SiteInfo {
  year: number;
  visitorCount: number;
  displayEmail?: string;
  version: string;
  teams: TeamSummary[];
}

/** auth.Permissions */
export interface Permissions {
  canEdit: boolean;
  canManageGallery: boolean;
  canManageUsers: boolean;
}

/** auth.CurrentUser — GET /auth/me */
export interface CurrentUser {
  id: number;
  name: string;
  email: string;
  phone?: string;
  role: string;
  teamId?: number;
  imageUrl?: string;
  permissions: Permissions;
}
