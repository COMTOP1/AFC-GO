import { useState, type FormEvent } from 'react';
import { useNavigate, useParams } from 'react-router';

import { createNews, updateNews, useNewsArticle, type NewsArticle } from '../../api/news';
import { queryKeys } from '../../api/queries';
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
import { parseId } from '../../lib/ids';
import { emptyImage, type ImageValue } from '../../lib/images';
import { isNotFound } from '../../lib/notFound';
import NotFoundPage from '../NotFoundPage';

function NewsForm({ article }: { article?: NewsArticle }) {
  const heading = article ? 'Edit article' : 'New article';
  usePageTitle(heading);
  const navigate = useNavigate();
  const toast = useToast();
  const [title, setTitle] = useState(article?.title ?? '');
  const [content, setContent] = useState(article?.content ?? '');
  const [image, setImage] = useState<ImageValue>(emptyImage);
  const [missing, setMissing] = useState<string | undefined>();

  const save = useSaveForm({
    submit: () =>
      article
        ? updateNews(article.id, { title, content, image })
        : createNews({ title, content, image }),
    invalidate: [queryKeys.news, queryKeys.home],
    onSaved: (a) => {
      toast.show({ tone: 'success', message: 'Article saved' });
      navigate(`/news/${a.id}`);
    },
  });

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!title.trim()) {
      setMissing('Enter a title');
      return;
    }
    setMissing(undefined);
    void save.run();
  }

  return (
    <>
      <PageHeader title={heading} />
      <form onSubmit={onSubmit} noValidate className="flex max-w-3xl flex-col gap-5">
        {save.formError && <Alert tone="error">{save.formError}</Alert>}
        <Field label="Title" error={missing ?? save.fieldErrors.title}>
          <Input value={title} onChange={(e) => setTitle(e.target.value)} />
        </Field>
        <ImageField
          label="Image"
          currentUrl={article?.imageUrl}
          allowRemove={Boolean(article)}
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
            Save article
          </Button>
          <Button variant="secondary" onClick={() => navigate(-1)} disabled={save.busy}>
            Cancel
          </Button>
        </div>
      </form>
    </>
  );
}

function EditNews({ id }: { id: number | null }) {
  const article = useNewsArticle(id);
  if (id === null || isNotFound(article.error)) {
    return <NotFoundPage />;
  }
  return <QueryState query={article}>{(a) => <NewsForm article={a} />}</QueryState>;
}

export default function NewsFormPage() {
  const { id } = useParams();
  return (
    <RequireEditor>{id === undefined ? <NewsForm /> : <EditNews id={parseId(id)} />}</RequireEditor>
  );
}
