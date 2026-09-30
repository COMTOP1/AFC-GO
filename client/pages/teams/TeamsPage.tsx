import { useTeams } from '../../api/teams';
import type { TeamSummary } from '../../api/types';
import { useCanEdit } from '../../components/edit/useCanEdit';
import { CardGrid } from '../../components/page/CardGrid';
import { LinkCard } from '../../components/page/LinkCard';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Badge } from '../../components/ui/Badge';
import { ButtonLink } from '../../components/ui/ButtonLink';
import { PageHeader } from '../../components/ui/PageHeader';

function leagueLine(team: TeamSummary): string | undefined {
  const parts = [team.league, team.division].filter(Boolean);
  return parts.length ? parts.join(' · ') : undefined;
}

export default function TeamsPage() {
  usePageTitle('Teams');
  const teams = useTeams();
  const { canEdit } = useCanEdit();
  return (
    <>
      <PageHeader
        title="Teams"
        actions={canEdit && <ButtonLink to="/teams/new">Add team</ButtonLink>}
      />
      <QueryState query={teams}>
        {(list) => (
          <CardGrid
            items={list}
            getKey={(t) => t.id}
            pageSize={Number.POSITIVE_INFINITY}
            emptyTitle="No teams yet"
            render={(t) => (
              <LinkCard
                to={`/team/${t.id}`}
                imageUrl={t.imageUrl}
                title={t.name}
                meta={leagueLine(t)}
              >
                <div className="flex flex-wrap gap-2 pt-1">
                  <Badge tone={t.isYouth ? 'red' : 'blue'}>{t.isYouth ? 'Youth' : 'Adult'}</Badge>
                  {!t.isActive && <Badge>Inactive</Badge>}
                </div>
              </LinkCard>
            )}
          />
        )}
      </QueryState>
    </>
  );
}
