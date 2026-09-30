import { useInfo } from '../../api/pages';
import { EditorLink } from '../../components/page/EditorLink';
import { QueryState } from '../../components/page/QueryState';
import { RichText } from '../../components/page/RichText';
import { usePageTitle } from '../../components/page/usePageTitle';
import { PageHeader } from '../../components/ui/PageHeader';
import { InfoFallback } from './fallback';

export default function InfoPage() {
  usePageTitle('Information');
  const info = useInfo();
  return (
    <>
      <PageHeader title="Information" actions={<EditorLink legacyHref="/info/edit" />} />
      <QueryState query={info}>
        {(data) => (data.content.trim() ? <RichText html={data.content} /> : <InfoFallback />)}
      </QueryState>
    </>
  );
}
