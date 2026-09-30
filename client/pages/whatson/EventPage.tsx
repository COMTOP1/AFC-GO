import { useNavigate, useParams } from 'react-router';

import { deleteEvent, useWhatsOnEvent } from '../../api/whatson';
import { DeleteButton } from '../../components/edit/DeleteButton';
import { useCanEdit } from '../../components/edit/useCanEdit';
import { ArticleView } from '../../components/page/ArticleView';
import { QueryState } from '../../components/page/QueryState';
import { ButtonLink } from '../../components/ui/ButtonLink';
import { formatDate } from '../../lib/format';
import { parseId } from '../../lib/ids';
import { isNotFound } from '../../lib/notFound';
import NotFoundPage from '../NotFoundPage';

export default function EventPage() {
  const id = parseId(useParams().id);
  const event = useWhatsOnEvent(id);
  const { canEdit } = useCanEdit();
  const navigate = useNavigate();
  if (id === null || isNotFound(event.error)) {
    return <NotFoundPage />;
  }
  return (
    <QueryState query={event}>
      {(e) => (
        <ArticleView
          section="What's On"
          sectionHref="/whatson"
          crumb={formatDate(e.dateOfEvent)}
          title={e.title}
          subtitle={formatDate(e.dateOfEvent, 'dayDate')}
          imageUrl={e.imageUrl}
          html={e.content}
          backLabel="← All events"
          actions={
            canEdit && (
              <>
                <ButtonLink to={`/whatson/${e.id}/edit`} variant="secondary" size="sm">
                  Edit
                </ButtonLink>
                <DeleteButton
                  confirmTitle="Delete this event?"
                  confirmMessage="This can't be undone."
                  onDelete={() => deleteEvent(e.id)}
                  invalidate={[['whatson'], ['home']]}
                  successMessage="Event deleted"
                  after={() => navigate('/whatson')}
                />
              </>
            )
          }
        />
      )}
    </QueryState>
  );
}
