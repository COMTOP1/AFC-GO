import { useNewsList } from '../../api/news';
import { CardGrid } from '../../components/page/CardGrid';
import { EditorLink } from '../../components/page/EditorLink';
import { LinkCard } from '../../components/page/LinkCard';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { PageHeader } from '../../components/ui/PageHeader';
import { formatDate } from '../../lib/format';

export default function NewsListPage() {
  usePageTitle('News');
  const news = useNewsList();
  return (
    <>
      <PageHeader title="News" actions={<EditorLink legacyHref="/news" />} />
      <QueryState query={news}>
        {(list) => (
          <CardGrid
            items={list}
            getKey={(a) => a.id}
            emptyTitle="No news yet"
            render={(a) => (
              <LinkCard
                to={`/news/${a.id}`}
                imageUrl={a.imageUrl}
                title={a.title}
                meta={formatDate(a.date)}
              />
            )}
          />
        )}
      </QueryState>
    </>
  );
}
