import { useNewsList } from '../../api/news';
import { useCanEdit } from '../../components/edit/useCanEdit';
import { CardGrid } from '../../components/page/CardGrid';
import { LinkCard } from '../../components/page/LinkCard';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { ButtonLink } from '../../components/ui/ButtonLink';
import { PageHeader } from '../../components/ui/PageHeader';
import { formatDate } from '../../lib/format';

export default function NewsListPage() {
  usePageTitle('News');
  const news = useNewsList();
  const { canEdit } = useCanEdit();
  return (
    <>
      <PageHeader
        title="News"
        actions={canEdit && <ButtonLink to="/news/new">Add article</ButtonLink>}
      />
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
