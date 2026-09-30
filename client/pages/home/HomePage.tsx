import { clsx } from 'clsx';
import { Link } from 'react-router';

import type { HomeData } from '../../api/home';
import { useHome } from '../../api/home';
import type { NewsArticle } from '../../api/news';
import type { WhatsOnEvent } from '../../api/whatson';
import { LogoRow } from '../../components/page/LogoRow';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Card, CardBody, CardMedia } from '../../components/ui/Card';
import { formatDate } from '../../lib/format';
import { plainText } from '../../lib/sanitize';

const kicker = 'text-xs font-bold tracking-widest uppercase';

function NewsHero({ article, className }: { article: NewsArticle; className?: string }) {
  const href = `/news/${article.id}`;
  return (
    <Card className={className}>
      <Link to={href} className="relative block">
        <CardMedia src={article.imageUrl} alt="" />
        <div className="absolute inset-x-0 bottom-0 bg-linear-to-t from-black/75 to-transparent p-4 text-white">
          {/* Hidden from the link's name, which is just the headline. */}
          <p className={kicker} aria-hidden="true">
            Latest news
          </p>
          <h2 className="font-display text-2xl leading-none font-extrabold uppercase md:text-3xl">
            {article.title}
          </h2>
        </div>
      </Link>
      <CardBody>
        <p className="text-muted">{plainText(article.content, 180)}</p>
        <Link to={href} className="mt-2 inline-block font-semibold text-red" tabIndex={-1}>
          Read more →
        </Link>
      </CardBody>
    </Card>
  );
}

function NextEvent({ event }: { event: WhatsOnEvent }) {
  return (
    <Card>
      <CardBody className="flex h-full flex-col gap-2">
        <p className={clsx(kicker, 'text-red')}>Next event</p>
        <h2 className="font-display text-2xl leading-none font-extrabold uppercase">
          <Link to={`/whatson/${event.id}`} className="hover:text-red">
            {event.title}
          </Link>
        </h2>
        <p className="text-sm font-semibold">{formatDate(event.dateOfEvent, 'dateTime')}</p>
        <p className="text-sm text-muted">{plainText(event.content, 120)}</p>
        <Link to="/whatson" className="mt-auto font-semibold text-red">
          All events →
        </Link>
      </CardBody>
    </Card>
  );
}

function HomeContent({ data }: { data: HomeData }) {
  const { latestNews, nextEvent, sponsors, affiliations } = data;
  return (
    <div className="space-y-10">
      {(latestNews || nextEvent) && (
        <div className={clsx('grid gap-4', latestNews && nextEvent && 'md:grid-cols-3')}>
          {latestNews && (
            <NewsHero article={latestNews} className={nextEvent ? 'md:col-span-2' : undefined} />
          )}
          {nextEvent && <NextEvent event={nextEvent} />}
        </div>
      )}
      <LogoRow title="Our sponsors" items={sponsors} />
      <LogoRow title="Affiliations" items={affiliations} />
    </div>
  );
}

export default function HomePage() {
  usePageTitle();
  const home = useHome();
  return (
    <>
      {/* The masthead shows the club name; the page still needs its own h1. */}
      <h1 className="sr-only">AFC Aldermaston</h1>
      <QueryState query={home}>{(data) => <HomeContent data={data} />}</QueryState>
    </>
  );
}
