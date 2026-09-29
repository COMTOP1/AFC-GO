import { useSite } from '../api/queries';
import { useAuth } from '../auth/useAuth';
import { Alert } from '../components/ui/Alert';
import { Badge } from '../components/ui/Badge';
import { Card, CardBody } from '../components/ui/Card';
import { EmptyState } from '../components/ui/EmptyState';
import { PageHeader } from '../components/ui/PageHeader';
import { Skeleton } from '../components/ui/Skeleton';
import { Table, TBody, Td, Th, THead, Tr } from '../components/ui/Table';

// Proof that the client, API, session cookie and data layer work end to end.
// Real pages arrive in sub-project 4.
export default function HomePage() {
  const site = useSite();
  const { user, isLoading } = useAuth();

  let signedIn = 'Not signed in';
  if (isLoading) {
    signedIn = 'Checking sign-in…';
  } else if (user) {
    signedIn = `Signed in as ${user.name} (${user.role})`;
  }

  return (
    <>
      <PageHeader title="AFC Aldermaston" subtitle={signedIn} />
      {site.isPending && (
        <div className="space-y-2">
          <Skeleton className="w-1/3" />
          <Skeleton />
          <Skeleton className="w-2/3" />
        </div>
      )}
      {site.isError && (
        <Alert tone="error">Could not load site information: {site.error.message}</Alert>
      )}
      {site.data && (
        <Card>
          <CardBody className="space-y-4">
            <p className="text-muted">Visitors: {site.data.visitorCount}</p>
            <h2 className="font-display text-2xl font-extrabold tracking-wide uppercase">Teams</h2>
            {site.data.teams.length === 0 ? (
              <EmptyState title="No teams yet." />
            ) : (
              <Table>
                <THead>
                  <Tr>
                    <Th>Team</Th>
                    <Th>Type</Th>
                  </Tr>
                </THead>
                <TBody>
                  {site.data.teams.map((team) => (
                    <Tr key={team.id}>
                      <Td>{team.name}</Td>
                      <Td>
                        <Badge tone={team.isYouth ? 'red' : 'blue'}>
                          {team.isYouth ? 'Youth' : 'Adult'}
                        </Badge>
                      </Td>
                    </Tr>
                  ))}
                </TBody>
              </Table>
            )}
          </CardBody>
        </Card>
      )}
    </>
  );
}
