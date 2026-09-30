import { useNavigate, useParams } from 'react-router';

import { deleteNews, useNewsArticle } from '../../api/news';
import { queryKeys } from '../../api/queries';
import { DeleteButton } from '../../components/edit/DeleteButton';
import { useCanEdit } from '../../components/edit/useCanEdit';
import { ArticleView } from '../../components/page/ArticleView';
import { QueryState } from '../../components/page/QueryState';
import { ButtonLink } from '../../components/ui/ButtonLink';
import { formatDate } from '../../lib/format';
import { parseId } from '../../lib/ids';
import { isNotFound } from '../../lib/notFound';
import NotFoundPage from '../NotFoundPage';

export default function NewsArticlePage() {
  const id = parseId(useParams().id);
  const article = useNewsArticle(id);
  const { canEdit } = useCanEdit();
  const navigate = useNavigate();
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
          actions={
            canEdit && (
              <>
                <ButtonLink to={`/news/${a.id}/edit`} variant="secondary" size="sm">
                  Edit
                </ButtonLink>
                <DeleteButton
                  confirmTitle="Delete this article?"
                  confirmMessage="This can't be undone."
                  onDelete={() => deleteNews(a.id)}
                  invalidate={[queryKeys.news, queryKeys.home]}
                  successMessage="Article deleted"
                  after={() => navigate('/news')}
                />
              </>
            )
          }
        />
      )}
    </QueryState>
  );
}
