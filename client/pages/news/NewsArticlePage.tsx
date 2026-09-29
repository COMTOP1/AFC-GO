import { useParams } from 'react-router';

import { useNewsArticle } from '../../api/news';
import { ArticleView } from '../../components/page/ArticleView';
import { QueryState } from '../../components/page/QueryState';
import { formatDate } from '../../lib/format';
import { parseId } from '../../lib/ids';
import { isNotFound } from '../../lib/notFound';
import NotFoundPage from '../NotFoundPage';

export default function NewsArticlePage() {
  const id = parseId(useParams().id);
  const article = useNewsArticle(id);
  if (id === null || isNotFound(article.error)) {
    return <NotFoundPage />;
  }
  return (
    <QueryState query={article}>
      {(a) => (
        <ArticleView
          section="News"
          sectionHref="/news"
          crumb={formatDate(a.date)}
          title={a.title}
          imageUrl={a.imageUrl}
          html={a.content}
          backLabel="← All news"
          editorHref={`/news/${a.id}`}
        />
      )}
    </QueryState>
  );
}
