import { useSite } from '../api/queries';
import { useAuth } from '../auth/useAuth';

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
    <section>
      <h1>AFC Aldermaston</h1>
      <p>{signedIn}</p>
      {site.isPending && <p>Loading…</p>}
      {site.isError && <p role="alert">Could not load site information: {site.error.message}</p>}
      {site.data && (
        <>
          <p>Visitors: {site.data.visitorCount}</p>
          <h2>Teams</h2>
          {site.data.teams.length === 0 ? (
            <p>No teams yet.</p>
          ) : (
            <ul>
              {site.data.teams.map((team) => (
                <li key={team.id}>{team.name}</li>
              ))}
            </ul>
          )}
        </>
      )}
    </section>
  );
}
