import { useId, useState } from 'react';
import { useSearchParams } from 'react-router';

import { deleteProgramme, useProgrammes, useSeasons, type Programme } from '../../api/programmes';
import { DeleteButton } from '../../components/edit/DeleteButton';
import { useCanEdit } from '../../components/edit/useCanEdit';
import { QueryState } from '../../components/page/QueryState';
import { SearchInput } from '../../components/page/SearchInput';
import { usePageTitle } from '../../components/page/usePageTitle';
import { useSearchQuery } from '../../components/page/useSearchQuery';
import { Button } from '../../components/ui/Button';
import { ButtonLink } from '../../components/ui/ButtonLink';
import { Select } from '../../components/ui/controls';
import { EmptyState } from '../../components/ui/EmptyState';
import { Field } from '../../components/ui/Field';
import { PageHeader } from '../../components/ui/PageHeader';
import { formatDate } from '../../lib/format';
import { parseId } from '../../lib/ids';
import { matchesQuery } from '../../lib/text';
import { AddProgrammeDialog } from './AddProgrammeDialog';
import { SeasonsDialog } from './SeasonsDialog';

const NO_SEASON = 'No season';

/** Groups in API order, with programmes that have no season last. */
function groupBySeason(list: Programme[]): [string, Programme[]][] {
  const groups = new Map<string, Programme[]>();
  for (const p of list) {
    const key = p.season?.name ?? NO_SEASON;
    groups.set(key, [...(groups.get(key) ?? []), p]);
  }
  const entries = [...groups.entries()];
  return [...entries.filter(([k]) => k !== NO_SEASON), ...entries.filter(([k]) => k === NO_SEASON)];
}

function SeasonGroup({
  name,
  items,
  canEdit,
}: {
  name: string;
  items: Programme[];
  canEdit: boolean;
}) {
  const headingId = useId();
  return (
    <section aria-labelledby={headingId}>
      <h2
        id={headingId}
        className="mb-2 font-display text-2xl font-extrabold tracking-wide uppercase"
      >
        {name}
      </h2>
      <ul className="divide-y divide-line rounded-lg border border-line">
        {items.map((p) => (
          <li key={p.id} className="flex items-center justify-between gap-3 px-4 py-3">
            <span>
              <span className="font-medium">{p.name}</span>{' '}
              <span className="text-sm text-muted">{formatDate(p.date)}</span>
            </span>
            <span className="flex gap-2">
              <ButtonLink
                href={p.fileUrl}
                variant="secondary"
                size="sm"
                target="_blank"
                rel="noopener noreferrer"
                aria-label={`View ${p.name}`}
              >
                View
              </ButtonLink>
              {canEdit && (
                <DeleteButton
                  ariaLabel={`Delete ${p.name}`}
                  confirmTitle={`Delete ${p.name}?`}
                  confirmMessage="This can't be undone."
                  onDelete={() => deleteProgramme(p.id)}
                  invalidate={[['programmes']]}
                  successMessage="Programme deleted"
                />
              )}
            </span>
          </li>
        ))}
      </ul>
    </section>
  );
}

export default function ProgrammesPage() {
  usePageTitle('Programmes');
  const [params, setParams] = useSearchParams();
  const requested = parseId(params.get('season') ?? undefined) ?? 0;
  const seasons = useSeasons();
  // A shared link can name a season that has since been deleted: once the list is
  // known, treat an unknown id as "All seasons" rather than showing a 404.
  const seasonId =
    requested !== 0 && seasons.isSuccess && !seasons.data.some((s) => s.id === requested)
      ? 0
      : requested;
  // Wait for the season list before asking for one season, so a stale id never hits the API.
  const programmes = useProgrammes(seasonId, {
    enabled: seasonId === 0 || !seasons.isPending,
  });
  const q = useSearchQuery();
  const { canEdit } = useCanEdit();
  const [adding, setAdding] = useState(false);
  const [managing, setManaging] = useState(false);

  return (
    <>
      <PageHeader
        title="Programmes"
        actions={
          canEdit && (
            <>
              <Button onClick={() => setAdding(true)}>Add programme</Button>
              <Button variant="secondary" onClick={() => setManaging(true)}>
                Manage seasons
              </Button>
            </>
          )
        }
      />
      <div className="mb-6 flex flex-wrap items-end gap-4">
        <Field label="Season" className="w-56">
          <Select
            value={String(seasonId)}
            onChange={(e) => {
              const next = new URLSearchParams(params);
              if (e.target.value === '0') {
                next.delete('season');
              } else {
                next.set('season', e.target.value);
              }
              setParams(next, { replace: true });
            }}
          >
            <option value="0">All seasons</option>
            {seasons.data?.map((s) => (
              <option key={s.id} value={String(s.id)}>
                {s.name}
              </option>
            ))}
          </Select>
        </Field>
        <SearchInput label="Search programmes" />
      </div>
      <QueryState query={programmes} isEmpty={(l) => l.length === 0} emptyTitle="No programmes yet">
        {(list) => {
          const shown = list.filter((p) => matchesQuery(p.name, q));
          if (shown.length === 0) {
            return <EmptyState title={`No programmes match '${q}'`} />;
          }
          return (
            <div className="space-y-8">
              {groupBySeason(shown).map(([name, items]) => (
                <SeasonGroup key={name} name={name} items={items} canEdit={canEdit} />
              ))}
            </div>
          );
        }}
      </QueryState>
      <AddProgrammeDialog open={adding} onClose={() => setAdding(false)} />
      <SeasonsDialog open={managing} onClose={() => setManaging(false)} />
    </>
  );
}
