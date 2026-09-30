import { useQuery } from '@tanstack/react-query';

import { useAuth } from '../auth/useAuth';
import { apiFetch } from './client';
import { queryKeys } from './queries';
import type { Sponsor } from './sponsors';
import type { TeamSummary } from './types';
import { formData } from '../lib/editForm';
import type { ImageValue } from '../lib/images';

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
  const { user, isLoading } = useAuth();
  return useQuery({
    queryKey: queryKeys.teams(user !== null),
    queryFn: ({ signal }) => apiFetch<TeamSummary[]>('/teams', { signal }),
    // Wait for the sign-in check: fetching first as anonymous and then again as
    // signed-in would fetch twice and flash the list back to its loading state.
    enabled: !isLoading,
  });
}

export function useTeam(id: number | null) {
  return useQuery({
    queryKey: queryKeys.team(id ?? 0),
    queryFn: ({ signal }) => apiFetch<TeamDetail>(`/teams/${id}`, { signal }),
    enabled: id !== null,
  });
}

export interface TeamInput {
  name: string;
  /** 6–18 for "Under N", 19 for "Over 18". */
  ages: number;
  description: string;
  league: string;
  division: string;
  leagueTable: string;
  fixtures: string;
  coach: string;
  physio: string;
  isActive: boolean;
  isYouth: boolean;
  image: ImageValue;
}

function teamForm(input: TeamInput, isUpdate: boolean): FormData {
  return formData({
    name: input.name,
    ages: input.ages,
    description: input.description,
    league: input.league,
    division: input.division,
    leagueTable: input.leagueTable,
    fixtures: input.fixtures,
    coach: input.coach,
    physio: input.physio,
    isActive: input.isActive,
    isYouth: input.isYouth,
    image: input.image.file,
    removeImage: isUpdate && input.image.remove ? true : undefined,
  });
}

export function createTeam(input: TeamInput): Promise<TeamSummary> {
  return apiFetch<TeamSummary>('/teams', { form: teamForm(input, false) });
}

export function updateTeam(id: number, input: TeamInput): Promise<TeamSummary> {
  return apiFetch<TeamSummary>(`/teams/${id}`, { method: 'PATCH', form: teamForm(input, true) });
}

export function deleteTeam(id: number): Promise<void> {
  return apiFetch<void>(`/teams/${id}`, { method: 'DELETE' });
}
