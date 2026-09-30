import { useQuery } from '@tanstack/react-query';

import { apiFetch } from './client';
import { queryKeys } from './queries';
import { formData } from '../lib/editForm';
import type { ImageValue } from '../lib/images';

/** player.Public — GET /players (signed-in users). imageUrl is absent when the photo is hidden. */
export interface Player {
  id: number;
  name: string;
  position?: string;
  isCaptain: boolean;
  dateOfBirth?: string;
  age?: number;
  team?: { id: number; name: string; isYouth: boolean };
  imageUrl?: string;
}

export function usePlayers() {
  return useQuery({
    queryKey: queryKeys.players,
    queryFn: ({ signal }) => apiFetch<Player[]>('/players', { signal }),
  });
}

export interface PlayerInput {
  name: string;
  teamId: number;
  /** YYYY-MM-DD */
  dateOfBirth: string;
  position: string;
  isCaptain: boolean;
  image: ImageValue;
}

function playerForm(input: PlayerInput, isUpdate: boolean): FormData {
  return formData({
    name: input.name,
    teamId: input.teamId,
    dateOfBirth: input.dateOfBirth,
    position: input.position,
    isCaptain: input.isCaptain,
    image: input.image.file,
    removeImage: isUpdate && input.image.remove ? true : undefined,
  });
}

export function createPlayer(input: PlayerInput): Promise<Player> {
  return apiFetch<Player>('/players', { form: playerForm(input, false) });
}

export function updatePlayer(id: number, input: PlayerInput): Promise<Player> {
  return apiFetch<Player>(`/players/${id}`, { method: 'PATCH', form: playerForm(input, true) });
}

export function deletePlayer(id: number): Promise<void> {
  return apiFetch<void>(`/players/${id}`, { method: 'DELETE' });
}
