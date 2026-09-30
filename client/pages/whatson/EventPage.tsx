import { useParams } from 'react-router';

import { useWhatsOnEvent } from '../../api/whatson';
import { ArticleView } from '../../components/page/ArticleView';
import { QueryState } from '../../components/page/QueryState';
import { formatDate } from '../../lib/format';
import { parseId } from '../../lib/ids';
import { isNotFound } from '../../lib/notFound';
import NotFoundPage from '../NotFoundPage';

export default function EventPage() {
  const id = parseId(useParams().id);
  const event = useWhatsOnEvent(id);
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
          editorHref={`/whatson/${e.id}`}
        />
      )}
    </QueryState>
  );
}
