import { useWhatsOnList, type WhatsOnPeriod } from '../../api/whatson';
import { useCanEdit } from '../../components/edit/useCanEdit';
import { CardGrid } from '../../components/page/CardGrid';
import { LinkCard } from '../../components/page/LinkCard';
import { QueryState } from '../../components/page/QueryState';
import { TabsNav } from '../../components/page/TabsNav';
import { usePageTitle } from '../../components/page/usePageTitle';
import { useTabParam } from '../../components/page/useTabParam';
import { ButtonLink } from '../../components/ui/ButtonLink';
import { PageHeader } from '../../components/ui/PageHeader';
import { formatDate } from '../../lib/format';

const tabs: { value: WhatsOnPeriod; label: string }[] = [
  { value: 'future', label: 'Upcoming' },
  { value: 'past', label: 'Past' },
  { value: 'all', label: 'All' },
];

const emptyTitles: Record<WhatsOnPeriod, string> = {
  future: 'No upcoming events',
  past: 'No past events',
  all: 'No events yet',
};

export default function WhatsOnPage() {
  usePageTitle("What's On");
  const period = useTabParam(
    'period',
    tabs.map((t) => t.value),
    'future',
  ) as WhatsOnPeriod;
  const events = useWhatsOnList(period);
  const { canEdit } = useCanEdit();
  return (
    <>
      <PageHeader
        title="What's On"
        actions={canEdit && <ButtonLink to="/whatson/new">Add event</ButtonLink>}
      />
      <TabsNav param="period" tabs={tabs} defaultValue="future" label="Event period" />
      <QueryState query={events}>
        {(list) => (
          <CardGrid
            items={list}
            getKey={(e) => e.id}
            resetKey={period}
            emptyTitle={emptyTitles[period]}
            render={(e) => (
              <LinkCard
                to={`/whatson/${e.id}`}
                imageUrl={e.imageUrl}
                title={e.title}
                meta={formatDate(e.dateOfEvent, 'dayDate')}
              />
            )}
          />
        )}
      </QueryState>
    </>
  );
}
