import { useState } from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import { useNavigate } from 'react-router';

import { setInfo, useInfo } from '../../api/pages';
import { queryKeys } from '../../api/queries';
import { RequireEditor } from '../../components/edit/RequireEditor';
import { RichTextEditor } from '../../components/edit/RichTextEditor';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { PageHeader } from '../../components/ui/PageHeader';
import { useToast } from '../../components/ui/toast/useToast';
import { InfoFallback } from './fallback';

function InfoForm({ initial }: { initial: string }) {
  const navigate = useNavigate();
  const toast = useToast();
  const [content, setContent] = useState(initial);
  const save = useSaveForm({
    submit: () => setInfo(content),
    invalidate: [queryKeys.info],
    onSaved: () => {
      toast.show({ tone: 'success', message: 'Club information saved' });
      navigate('/info');
    },
  });
  return (
    <form
      noValidate
      className="flex max-w-3xl flex-col gap-5"
      onSubmit={(e) => {
        e.preventDefault();
        void save.run();
      }}
    >
      {save.formError && <Alert tone="error">{save.formError}</Alert>}
      <RichTextEditor
        label="Club information"
        value={initial}
        onChange={setContent}
        error={save.fieldErrors.content}
      />
      <div className="flex gap-2">
        <Button type="submit" loading={save.busy}>
          Save
        </Button>
        <Button variant="secondary" onClick={() => navigate('/info')} disabled={save.busy}>
          Cancel
        </Button>
      </div>
    </form>
  );
}

export default function InfoEditPage() {
  usePageTitle('Edit information');
  const info = useInfo();
  return (
    <RequireEditor>
      <PageHeader title="Edit information" />
      <QueryState query={info}>
        {(data) => (
          <InfoForm
            // With nothing saved yet, start from the wording the page already shows.
            initial={data.content.trim() ? data.content : renderToStaticMarkup(<InfoFallback />)}
          />
        )}
      </QueryState>
    </RequireEditor>
  );
}
