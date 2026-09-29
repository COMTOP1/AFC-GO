import { usePageTitle } from '../components/page/usePageTitle';
import { ButtonLink } from '../components/ui/ButtonLink';
import { EmptyState } from '../components/ui/EmptyState';
import { PageHeader } from '../components/ui/PageHeader';

export default function NotFoundPage() {
  usePageTitle('Page not found');
  return (
    <>
      <PageHeader title="Page not found" />
      <EmptyState
        title="Nothing here"
        message="The page you were looking for doesn't exist or has moved."
        action={<ButtonLink to="/">Go to the start</ButtonLink>}
      />
    </>
  );
}
