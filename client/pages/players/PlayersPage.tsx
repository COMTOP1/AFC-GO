import { useState } from 'react';
import { useSearchParams } from 'react-router';

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

function PlayersTable({ canEdit, onEdit }: { canEdit: boolean; onEdit: (p: Player) => void }) {
  const players = usePlayers();
  const teams = useTeams();
  const [params, setParams] = useSearchParams();
  const q = useSearchQuery();
  const teamFilter = params.get('team') ?? '';

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
              (p) => matchesQuery(p.name, q) && (!teamFilter || String(p.team?.id) === teamFilter),
            );
          if (shown.length === 0) {
            return <EmptyState title="No players match your filters" />;
          }
          return (
            <Table>
              <THead>
                <Tr>
                  <Th>
                    <span className="sr-only">Photo</span>
                  </Th>
                  <Th>Name</Th>
                  <Th>Position</Th>
                  <Th>Team</Th>
                  <Th>Date of birth</Th>
                  {canEdit && (
                    <Th>
                      <span className="sr-only">Actions</span>
                    </Th>
                  )}
                </Tr>
              </THead>
              <TBody>
                {shown.map((p) => (
                  <Tr key={p.id}>
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
                    <Td className="whitespace-nowrap">{born(p)}</Td>
                    {canEdit && (
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
          );
        }}
      </QueryState>
    </>
  );
}

export default function PlayersPage() {
  usePageTitle('Players');
  const { canEdit } = useCanEdit();
  // null = closed; {} = adding; { player } = editing.
  const [dialog, setDialog] = useState<{ player?: Player } | null>(null);
  return (
    <>
      <PageHeader
        title="Players"
        actions={canEdit && <Button onClick={() => setDialog({})}>Add player</Button>}
      />
      <RequireSignIn title="Sign in to see the players">
        <PlayersTable canEdit={canEdit} onEdit={(player) => setDialog({ player })} />
      </RequireSignIn>
      <PlayerDialog
        open={dialog !== null}
        player={dialog?.player}
        onClose={() => setDialog(null)}
      />
    </>
  );
}
