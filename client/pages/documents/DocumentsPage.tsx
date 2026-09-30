import { useState } from 'react';

import { deleteDocument, useDocuments } from '../../api/documents';
import { queryKeys } from '../../api/queries';
import { DeleteButton } from '../../components/edit/DeleteButton';
import { useCanEdit } from '../../components/edit/useCanEdit';
import { QueryState } from '../../components/page/QueryState';
import { SearchInput } from '../../components/page/SearchInput';
import { usePageTitle } from '../../components/page/usePageTitle';
import { useSearchQuery } from '../../components/page/useSearchQuery';
import { Button } from '../../components/ui/Button';
import { ButtonLink } from '../../components/ui/ButtonLink';
import { EmptyState } from '../../components/ui/EmptyState';
import { PageHeader } from '../../components/ui/PageHeader';
import { matchesQuery } from '../../lib/text';
import { AddDocumentDialog } from './AddDocumentDialog';

export default function DocumentsPage() {
  usePageTitle('Documents');
  const docs = useDocuments();
  const q = useSearchQuery();
  const { canEdit } = useCanEdit();
  const [adding, setAdding] = useState(false);
  return (
    <>
      <PageHeader
        title="Documents"
        actions={canEdit && <Button onClick={() => setAdding(true)}>Add document</Button>}
      />
      <QueryState query={docs} isEmpty={(l) => l.length === 0} emptyTitle="No documents yet">
        {(list) => {
          const shown = list.filter((d) => matchesQuery(d.name, q));
          return (
            <>
              <SearchInput label="Search documents" />
              {shown.length === 0 ? (
                <EmptyState title={`No documents match '${q}'`} />
              ) : (
                <ul className="mt-4 divide-y divide-line rounded-lg border border-line">
                  {shown.map((d) => (
                    <li key={d.id} className="flex items-center justify-between gap-3 px-4 py-3">
                      <span className="font-medium">{d.name}</span>
                      <div className="flex gap-2">
                        <ButtonLink
                          href={d.fileUrl}
                          variant="secondary"
                          size="sm"
                          aria-label={`Download ${d.name}`}
                        >
                          Download
                        </ButtonLink>
                        {canEdit && (
                          <DeleteButton
                            ariaLabel={`Delete ${d.name}`}
                            confirmTitle={`Delete ${d.name}?`}
                            confirmMessage="This can't be undone."
                            onDelete={() => deleteDocument(d.id)}
                            invalidate={[queryKeys.documents]}
                            successMessage="Document deleted"
                          />
                        )}
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </>
          );
        }}
      </QueryState>
      <AddDocumentDialog open={adding} onClose={() => setAdding(false)} />
    </>
  );
}
