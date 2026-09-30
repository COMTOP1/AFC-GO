import { useDocuments } from '../../api/documents';
import { EditorLink } from '../../components/page/EditorLink';
import { QueryState } from '../../components/page/QueryState';
import { SearchInput } from '../../components/page/SearchInput';
import { usePageTitle } from '../../components/page/usePageTitle';
import { useSearchQuery } from '../../components/page/useSearchQuery';
import { ButtonLink } from '../../components/ui/ButtonLink';
import { EmptyState } from '../../components/ui/EmptyState';
import { PageHeader } from '../../components/ui/PageHeader';
import { matchesQuery } from '../../lib/text';

export default function DocumentsPage() {
  usePageTitle('Documents');
  const docs = useDocuments();
  const q = useSearchQuery();
  return (
    <>
      <PageHeader title="Documents" actions={<EditorLink legacyHref="/documents" />} />
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
                      <ButtonLink
                        href={d.fileUrl}
                        variant="secondary"
                        size="sm"
                        aria-label={`Download ${d.name}`}
                      >
                        Download
                      </ButtonLink>
                    </li>
                  ))}
                </ul>
              )}
            </>
          );
        }}
      </QueryState>
    </>
  );
}
