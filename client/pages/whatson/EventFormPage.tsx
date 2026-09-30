import { useState, type FormEvent } from 'react';
import { useNavigate, useParams } from 'react-router';

import { createEvent, updateEvent, useWhatsOnEvent, type WhatsOnEvent } from '../../api/whatson';
import { ImageField } from '../../components/edit/ImageField';
import { RequireEditor } from '../../components/edit/RequireEditor';
import { RichTextEditor } from '../../components/edit/RichTextEditor';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Input } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { PageHeader } from '../../components/ui/PageHeader';
import { useToast } from '../../components/ui/toast/useToast';
import { toDateInput } from '../../lib/editForm';
import { parseId } from '../../lib/ids';
import { emptyImage, type ImageValue } from '../../lib/images';
import { isNotFound } from '../../lib/notFound';
import NotFoundPage from '../NotFoundPage';

function EventForm({ event }: { event?: WhatsOnEvent }) {
  const heading = event ? 'Edit event' : 'New event';
  usePageTitle(heading);
  const navigate = useNavigate();
  const toast = useToast();
  const [title, setTitle] = useState(event?.title ?? '');
  const [date, setDate] = useState(event ? toDateInput(event.dateOfEvent) : '');
  const [content, setContent] = useState(event?.content ?? '');
  const [image, setImage] = useState<ImageValue>(emptyImage);
  const [missing, setMissing] = useState<{ title?: string; date?: string }>({});

  const save = useSaveForm({
    submit: () => {
      const input = { title, content, dateOfEvent: date, image };
      return event ? updateEvent(event.id, input) : createEvent(input);
    },
    invalidate: [['whatson'], ['home']],
    onSaved: (e) => {
      toast.show({ tone: 'success', message: 'Event saved' });
      navigate(`/whatson/${e.id}`);
    },
  });

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const next = {
      title: title.trim() ? undefined : 'Enter a title',
      date: date ? undefined : 'Choose the date of the event',
    };
    setMissing(next);
    if (next.title || next.date) {
      return;
    }
    void save.run();
  }

  return (
    <>
      <PageHeader title={heading} />
      <form onSubmit={onSubmit} noValidate className="flex max-w-3xl flex-col gap-5">
        {save.formError && <Alert tone="error">{save.formError}</Alert>}
        <Field label="Title" error={missing.title ?? save.fieldErrors.title}>
          <Input value={title} onChange={(e) => setTitle(e.target.value)} />
        </Field>
        <Field label="Date of event" error={missing.date ?? save.fieldErrors.dateOfEvent}>
          <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
        </Field>
        <ImageField
          label="Image"
          currentUrl={event?.imageUrl}
          allowRemove={Boolean(event)}
          value={image}
          onChange={setImage}
          error={save.fieldErrors.file ?? save.fieldErrors.image}
        />
        <RichTextEditor
          label="Content"
          value={content}
          onChange={setContent}
          error={save.fieldErrors.content}
        />
        <div className="flex gap-2">
          <Button type="submit" loading={save.busy}>
            Save event
          </Button>
          <Button variant="secondary" onClick={() => navigate(-1)} disabled={save.busy}>
            Cancel
          </Button>
        </div>
      </form>
    </>
  );
}

function EditEvent({ id }: { id: number | null }) {
  const event = useWhatsOnEvent(id);
  if (id === null || isNotFound(event.error)) {
    return <NotFoundPage />;
  }
  return <QueryState query={event}>{(e) => <EventForm event={e} />}</QueryState>;
}

export default function EventFormPage() {
  const { id } = useParams();
  return (
    <RequireEditor>
      {id === undefined ? <EventForm /> : <EditEvent id={parseId(id)} />}
    </RequireEditor>
  );
}
