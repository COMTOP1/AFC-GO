import { useSite } from '../../api/queries';
import { useSponsors } from '../../api/sponsors';
import { CardGrid } from '../../components/page/CardGrid';
import { ImageWithFallback } from '../../components/page/ImageWithFallback';
import { EditorLink } from '../../components/page/EditorLink';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Card, CardBody } from '../../components/ui/Card';
import { PageHeader } from '../../components/ui/PageHeader';
import { sponsorTeamLabel } from '../../lib/sponsorTeam';

export default function SponsorsPage() {
  usePageTitle('Sponsors');
  const sponsors = useSponsors();
  const site = useSite();
  return (
    <>
      <PageHeader title="Sponsors" actions={<EditorLink legacyHref="/sponsors" />} />
      <QueryState query={sponsors}>
        {(list) => (
          <CardGrid
            items={list}
            getKey={(s) => s.id}
            pageSize={Number.POSITIVE_INFINITY}
            emptyTitle="No sponsors yet"
            render={(s) => (
              <Card className="h-full">
                <div className="flex h-32 items-center justify-center border-b border-line bg-white p-4 text-black">
                  <ImageWithFallback
                    src={s.imageUrl}
                    alt=""
                    loading="lazy"
                    className="max-h-24 w-auto"
                    fallback={<span className="text-lg font-semibold">{s.name}</span>}
                  />
                </div>
                <CardBody className="space-y-1">
                  <h2 className="font-display text-xl font-extrabold uppercase">{s.name}</h2>
                  {s.purpose && <p className="text-sm">{s.purpose}</p>}
                  {(() => {
                    const label = sponsorTeamLabel(s.team, site.data?.teams ?? []);
                    return label && <p className="text-sm text-muted">Sponsor of {label}</p>;
                  })()}
                  {s.website && (
                    <a
                      href={s.website}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-block text-sm font-semibold text-red"
                    >
                      {s.name} website ↗
                    </a>
                  )}
                </CardBody>
              </Card>
            )}
          />
        )}
      </QueryState>
    </>
  );
}
