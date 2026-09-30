import { useAuth } from '../../auth/useAuth';
import { useSignIn } from '../../components/layout/useSignIn';
import { PageSkeleton } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Button } from '../../components/ui/Button';
import { EmptyState } from '../../components/ui/EmptyState';
import { PageHeader } from '../../components/ui/PageHeader';
import { DetailsCard } from './DetailsCard';
import { PasswordCard } from './PasswordCard';
import { PhotoCard } from './PhotoCard';

export default function AccountPage() {
  usePageTitle('Account');
  const { user, isLoading } = useAuth();
  const signIn = useSignIn();
  return (
    <>
      <PageHeader title="Your account" />
      {isLoading ? (
        <PageSkeleton />
      ) : !user ? (
        <EmptyState
          title="Sign in to see your account"
          message="Your session may have expired."
          action={<Button onClick={signIn.open}>Sign in</Button>}
        />
      ) : (
        <div className="grid gap-6 md:grid-cols-2">
          <PhotoCard user={user} />
          <DetailsCard user={user} />
          <div className="md:col-span-2">
            <PasswordCard />
          </div>
        </div>
      )}
    </>
  );
}
