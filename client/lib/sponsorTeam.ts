/** The API's fixed sponsor "team" codes, plus "no team". Team ids follow in the select. */
export const SPONSOR_TEAM_CHOICES = [
  { value: '', label: 'None' },
  { value: 'A', label: 'All teams' },
  { value: 'O', label: 'Adult teams' },
  { value: 'Y', label: 'Youth teams' },
];

const CODES: Record<string, string> = { A: 'All teams', O: 'Adult teams', Y: 'Youth teams' };

/** What a sponsor's raw `team` value means, for display; undefined when there's nothing to show. */
export function sponsorTeamLabel(
  team: string | undefined,
  teams: { id: number; name: string }[],
): string | undefined {
  if (!team) {
    return undefined;
  }
  if (CODES[team]) {
    return CODES[team];
  }
  return teams.find((t) => String(t.id) === team)?.name;
}
