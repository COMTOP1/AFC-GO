/** The server's role codes (role.GetRole) with their display names, in the legacy order. */
export const ROLES: { code: string; label: string }[] = [
  { code: 'photographer', label: 'Photographer' },
  { code: 'manager', label: 'Manager' },
  { code: 'programme_editor', label: 'Programme Editor' },
  { code: 'league_secretary', label: 'League Secretary' },
  { code: 'treasurer', label: 'Treasurer' },
  { code: 'safeguarding_officer', label: 'Safeguarding Officer' },
  { code: 'club_secretary', label: 'Club Secretary' },
  { code: 'chairperson', label: 'Chairperson' },
  { code: 'webmaster', label: 'Webmaster' },
];

/** Only managers have a team. */
export const MANAGER_ROLE = 'manager';
