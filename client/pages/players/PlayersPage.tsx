import { useState } from 'react';
import { useSearchParams } from 'react-router';
import { useQueryClient } from '@tanstack/react-query';

import { deletePlayer, usePlayers, type Player } from '../../api/players';
import { queryKeys } from '../../api/queries';
import { useTeams } from '../../api/teams';
import { DeleteButton } from '../../components/edit/DeleteButton';
import { RequireSignIn } from '../../components/edit/RequireSignIn';
import { useCanEdit } from '../../components/edit/useCanEdit';
import { QueryState } from '../../components/page/QueryState';
import { SearchInput } from '../../components/page/SearchInput';
import { Thumb } from '../../components/page/Thumb';
import { usePageTitle } from '../../components/page/usePageTitle';
import { useSearchQuery } from '../../components/page/useSearchQuery';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { Checkbox } from '../../components/ui/Checkbox';
import { Select } from '../../components/ui/controls';
import { EmptyState } from '../../components/ui/EmptyState';
import { Field } from '../../components/ui/Field';
import { PageHeader } from '../../components/ui/PageHeader';
import { Table, TBody, Td, Th, THead, Tr } from '../../components/ui/Table';
import { formatDate } from '../../lib/format';
import { matchesQuery } from '../../lib/text';
import { PlayerDialog } from './PlayerDialog';

function byTeamThenName(a: Player, b: Player): number {
  return (a.team?.name ?? '').localeCompare(b.team?.name ?? '') || a.name.localeCompare(b.name);
}

function born(p: Player): string {
  if (!p.dateOfBirth) {
    return '—';
  }
  const age = p.age === undefined ? '' : ` (age ${p.age})`;
  return `${formatDate(p.dateOfBirth)}${age}`;
}

/** The team filter value for players whose team was removed. */
const NO_TEAM = 'none';

function PlayersTable({ canManage, onEdit }: { canManage: boolean; onEdit: (p: Player) => void }) {
  const players = usePlayers();
  const teams = useTeams();
  const [params, setParams] = useSearchParams();
  const q = useSearchQuery();
  const teamFilter = params.get('team') ?? '';
  const [selected, setSelected] = useState<ReadonlySet<number>>(new Set());
  const queryClient = useQueryClient();

  function toggle(id: number, on: boolean) {
    const next = new Set(selected);
    if (on) {
      next.add(id);
    } else {
      next.delete(id);
    }
    setSelected(next);
  }

  // Deletes one at a time so a failure stops the rest. The list is refreshed
  // either way, so whatever was deleted before a failure disappears.
  async function deleteSelected() {
    try {
      for (const id of selected) {
        await deletePlayer(id);
      }
    } finally {
      void queryClient.invalidateQueries({ queryKey: queryKeys.players });
      void queryClient.invalidateQueries({ queryKey: ['team'] });
    }
  }

  return (
    <>
      <div className="mb-4 flex flex-wrap items-end gap-4">
        <SearchInput label="Search players" />
        <Field label="Team" className="w-56">
          <Select
            value={teamFilter}
            onChange={(e) => {
              const next = new URLSearchParams(params);
              if (e.target.value) {
                next.set('team', e.target.value);
              } else {
                next.delete('team');
              }
              setParams(next, { replace: true });
            }}
          >
            <option value="">All teams</option>
            <option value={NO_TEAM}>No team</option>
            {teams.data?.map((t) => (
              <option key={t.id} value={String(t.id)}>
                {t.name}
              </option>
            ))}
          </Select>
        </Field>
      </div>
      <QueryState query={players} isEmpty={(l) => l.length === 0} emptyTitle="No players yet">
        {(list) => {
          const shown = [...list]
            .sort(byTeamThenName)
            .filter(
              (p) =>
                matchesQuery(p.name, q) &&
                (!teamFilter ||
                  (teamFilter === NO_TEAM ? !p.team : String(p.team?.id) === teamFilter)),
            );
          if (shown.length === 0) {
            return <EmptyState title="No players match your filters" />;
          }
          // The server leaves dates of birth out for viewers who don't need
          // them (e.g. Programme Editors); drop the column when none came.
          const showBorn = list.some((p) => p.dateOfBirth);
          const shownSelected = shown.filter((p) => selected.has(p.id));
          return (
            <>
              {canManage && selected.size > 0 && (
                <div className="mb-3 flex items-center gap-3">
                  <DeleteButton
                    confirmTitle={`Delete ${selected.size} ${selected.size === 1 ? 'player' : 'players'}?`}
                    confirmMessage="Their details and photos will be removed. This can't be undone."
                    onDelete={deleteSelected}
                    successMessage={`${selected.size === 1 ? 'Player' : 'Players'} deleted`}
                    after={() => setSelected(new Set())}
                  >
                    {`Delete selected (${selected.size})`}
                  </DeleteButton>
                  <Button size="sm" variant="secondary" onClick={() => setSelected(new Set())}>
                    Clear selection
                  </Button>
                </div>
              )}
              <Table>
                <THead>
                  <Tr>
                    {canManage && (
                      <Th>
                        <Checkbox
                          label={<span className="sr-only">Select all shown</span>}
                          checked={shownSelected.length === shown.length}
                          onChange={(e) => {
                            const next = new Set(selected);
                            for (const p of shown) {
                              if (e.target.checked) {
                                next.add(p.id);
                              } else {
                                next.delete(p.id);
                              }
                            }
                            setSelected(next);
                          }}
                        />
                      </Th>
                    )}
                    <Th>
                      <span className="sr-only">Photo</span>
                    </Th>
                    <Th>Name</Th>
                    <Th>Position</Th>
                    <Th>Team</Th>
                    {showBorn && <Th>Date of birth</Th>}
                    {canManage && (
                      <Th>
                        <span className="sr-only">Actions</span>
                      </Th>
                    )}
                  </Tr>
                </THead>
                <TBody>
                  {shown.map((p) => (
                    <Tr key={p.id}>
                      {canManage && (
                        <Td>
                          <Checkbox
                            label={<span className="sr-only">{`Select ${p.name}`}</span>}
                            checked={selected.has(p.id)}
                            onChange={(e) => toggle(p.id, e.target.checked)}
                          />
                        </Td>
                      )}
                      <Td>
                        <Thumb src={p.imageUrl} />
                      </Td>
                      <Td>
                        <span className="font-medium">{p.name}</span>
                        {p.isCaptain && (
                          <Badge tone="red" className="ml-2">
                            Captain
                          </Badge>
                        )}
                      </Td>
                      <Td>{p.position}</Td>
                      <Td>{p.team?.name ?? 'No team'}</Td>
                      {showBorn && <Td className="whitespace-nowrap">{born(p)}</Td>}
                      {canManage && (
                        <Td>
                          <div className="flex justify-end gap-2">
                            <Button
                              size="sm"
                              variant="secondary"
                              aria-label={`Edit ${p.name}`}
                              onClick={() => onEdit(p)}
                            >
                              Edit
                            </Button>
                            <DeleteButton
                              ariaLabel={`Delete ${p.name}`}
                              confirmTitle={`Delete ${p.name}?`}
                              confirmMessage="This can't be undone."
                              onDelete={() => deletePlayer(p.id)}
                              invalidate={[queryKeys.players, ['team']]}
                              successMessage="Player deleted"
                            />
                          </div>
                        </Td>
                      )}
                    </Tr>
                  ))}
                </TBody>
              </Table>
            </>
          );
        }}
      </QueryState>
    </>
  );
}

export default function PlayersPage() {
  usePageTitle('Players');
  const { canManagePlayers } = useCanEdit();
  // null = closed; {} = adding; { player } = editing.
  const [dialog, setDialog] = useState<{ player?: Player } | null>(null);
  return (
    <>
      <PageHeader
        title="Players"
        actions={canManagePlayers && <Button onClick={() => setDialog({})}>Add player</Button>}
      />
      <RequireSignIn title="Sign in to see the players">
        <PlayersTable canManage={canManagePlayers} onEdit={(player) => setDialog({ player })} />
      </RequireSignIn>
      <PlayerDialog
        open={dialog !== null}
        player={dialog?.player}
        onClose={() => setDialog(null)}
      />
    </>
  );
}
