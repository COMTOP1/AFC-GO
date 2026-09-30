import { useId } from 'react';
import { Link, useParams } from 'react-router';

import { useTeam, type SquadMember, type TeamDetail } from '../../api/teams';
import crest from '../../assets/crest.png';
import { ImageWithFallback } from '../../components/page/ImageWithFallback';
import { LogoRow } from '../../components/page/LogoRow';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Badge } from '../../components/ui/Badge';
import { ButtonLink } from '../../components/ui/ButtonLink';
import { CardMedia } from '../../components/ui/Card';
import { PageHeader } from '../../components/ui/PageHeader';
import { parseId } from '../../lib/ids';
import { isNotFound } from '../../lib/notFound';
import NotFoundPage from '../NotFoundPage';

function Player({ player }: { player: SquadMember }) {
  return (
    <li className="text-center">
      <ImageWithFallback
        src={player.imageUrl}
        fallbackSrc={crest}
        alt=""
        loading="lazy"
        className="mx-auto mb-2 size-20 rounded-full border border-line bg-white object-cover"
      />
      <p className="font-semibold">{player.name}</p>
      {player.position && <p className="text-sm text-muted">{player.position}</p>}
      {player.isCaptain && <Badge tone="blue">Captain</Badge>}
    </li>
  );
}

function Squad({ players }: { players: SquadMember[] }) {
  const headingId = useId();
  return (
    <section aria-labelledby={headingId} className="mt-10">
      <h2
        id={headingId}
        className="mb-4 font-display text-2xl font-extrabold tracking-wide uppercase"
      >
        Squad
      </h2>
      <ul className="grid grid-cols-2 gap-6 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
        {players.map((p) => (
          <Player key={p.id} player={p} />
        ))}
      </ul>
    </section>
  );
}

function TeamContent({ detail }: { detail: TeamDetail }) {
  const { team, managers, sponsors, players } = detail;
  usePageTitle(team.name);
  const facts: [string, string | undefined][] = [
    ['League', team.league],
    ['Division', team.division],
    ['Coach', team.coach],
    ['Physio', team.physio],
  ];
  return (
    <>
      <nav aria-label="Breadcrumb" className="mb-2 text-sm text-muted">
        <Link to="/teams" className="font-semibold text-red">
          Teams
        </Link>{' '}
        / {team.name}
      </nav>
      <PageHeader title={team.name} subtitle={team.description} />
      <div className="grid gap-6 md:grid-cols-2">
        <div className="overflow-hidden rounded-lg border border-line">
          <CardMedia src={team.imageUrl} alt={`${team.name} team photo`} />
        </div>
        <div className="space-y-4">
          <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 text-sm">
            {facts
              .filter(([, value]) => value)
              .map(([label, value]) => (
                <div key={label} className="contents">
                  <dt className="text-muted">{label}</dt>
                  <dd>{value}</dd>
                </div>
              ))}
            {managers.length > 0 && (
              <div className="contents">
                <dt className="text-muted">{managers.length > 1 ? 'Managers' : 'Manager'}</dt>
                <dd className="space-y-1">
                  {managers.map((m) => (
                    <p key={m.email}>
                      {m.name} (
                      <a href={`mailto:${m.email}`} className="text-red underline">
                        {m.email}
                      </a>
                      )
                    </p>
                  ))}
                </dd>
              </div>
            )}
          </dl>
          <div className="flex flex-wrap gap-2">
            {team.leagueTableUrl && (
              <ButtonLink
                href={team.leagueTableUrl}
                variant="secondary"
                target="_blank"
                rel="noopener noreferrer"
              >
                League table ↗
              </ButtonLink>
            )}
            {team.fixturesUrl && (
              <ButtonLink
                href={team.fixturesUrl}
                variant="secondary"
                target="_blank"
                rel="noopener noreferrer"
              >
                Fixtures ↗
              </ButtonLink>
            )}
          </div>
        </div>
      </div>
      {/* The API returns no players for youth teams; the whole section is then omitted. */}
      {players.length > 0 && <Squad players={players} />}
      <div className="mt-10">
        <LogoRow title="Team sponsors" items={sponsors} />
      </div>
    </>
  );
}

export default function TeamPage() {
  const id = parseId(useParams().id);
  const team = useTeam(id);
  if (id === null || isNotFound(team.error)) {
    return <NotFoundPage />;
  }
  return <QueryState query={team}>{(detail) => <TeamContent detail={detail} />}</QueryState>;
}
