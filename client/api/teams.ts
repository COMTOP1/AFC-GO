import { useQuery } from '@tanstack/react-query';

import { useAuth } from '../auth/useAuth';
import { apiFetch } from './client';
import { queryKeys } from './queries';
import type { Sponsor } from './sponsors';
import type { TeamSummary } from './types';

/** site.Manager */
export interface Manager {
  name: string;
  email: string;
}

/** player.Member — youth teams never return any. */
export interface SquadMember {
  id: number;
  name: string;
  position?: string;
  isCaptain: boolean;
  imageUrl?: string;
}

/** site.TeamDetail — GET /teams/:id */
export interface TeamDetail {
  team: TeamSummary;
  managers: Manager[];
  sponsors: Sponsor[];
  players: SquadMember[];
}

export function useTeams() {
  const { user } = useAuth();
  return useQuery({
    queryKey: queryKeys.teams(user !== null),
    queryFn: ({ signal }) => apiFetch<TeamSummary[]>('/teams', { signal }),
  });
}

export function useTeam(id: number | null) {
  return useQuery({
    queryKey: queryKeys.team(id ?? 0),
    queryFn: ({ signal }) => apiFetch<TeamDetail>(`/teams/${id}`, { signal }),
    enabled: id !== null,
  });
}
