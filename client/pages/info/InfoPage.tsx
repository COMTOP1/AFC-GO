import { useInfo } from '../../api/pages';
import { useCanEdit } from '../../components/edit/useCanEdit';
import { QueryState } from '../../components/page/QueryState';
import { RichText } from '../../components/page/RichText';
import { usePageTitle } from '../../components/page/usePageTitle';
import { ButtonLink } from '../../components/ui/ButtonLink';
import { PageHeader } from '../../components/ui/PageHeader';
import { InfoFallback } from './fallback';

export default function InfoPage() {
  usePageTitle('Information');
  const info = useInfo();
  const { canEdit } = useCanEdit();
  return (
    <>
      <PageHeader
        title="Information"
        actions={
          canEdit && (
            <ButtonLink to="/info/edit" variant="secondary" size="sm">
              Edit
            </ButtonLink>
          )
        }
      />
      <QueryState query={info}>
        {(data) => (data.content.trim() ? <RichText html={data.content} /> : <InfoFallback />)}
      </QueryState>
    </>
  );
}
