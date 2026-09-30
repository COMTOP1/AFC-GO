import { useAuth } from '../../auth/useAuth';
import { RequireSignIn } from '../../components/edit/RequireSignIn';
import { usePageTitle } from '../../components/page/usePageTitle';
import { PageHeader } from '../../components/ui/PageHeader';
import { DetailsCard } from './DetailsCard';
import { PasswordCard } from './PasswordCard';
import { PhotoCard } from './PhotoCard';

export default function AccountPage() {
  usePageTitle('Account');
  return (
    <>
      <PageHeader title="Your account" />
      <RequireSignIn title="Sign in to see your account">
        <AccountContent />
      </RequireSignIn>
    </>
  );
}

function AccountContent() {
  const { user } = useAuth();
  if (!user) {
    return null;
  }
  return (
    <div className="grid gap-6 md:grid-cols-2">
      <PhotoCard user={user} />
      <DetailsCard user={user} />
      <div className="md:col-span-2">
        <PasswordCard />
      </div>
    </div>
  );
}
