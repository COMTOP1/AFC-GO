import { useId } from 'react';

import { useSite } from '../../api/queries';
import type { CurrentUser } from '../../api/types';
import { Card, CardBody } from '../../components/ui/Card';

export function DetailsCard({ user }: { user: CurrentUser }) {
  const headingId = useId();
  const site = useSite();
  const teamName = user.teamId
    ? site.data?.teams.find((t) => t.id === user.teamId)?.name
    : undefined;
  const rows: [string, string | undefined][] = [
    ['Name', user.name],
    ['Email', user.email],
    ['Phone', user.phone],
    ['Role', user.role],
    ['Team', teamName],
  ];
  return (
    <Card>
      <CardBody>
        <section aria-labelledby={headingId} className="space-y-4">
          <h2
            id={headingId}
            className="font-display text-2xl font-extrabold tracking-wide uppercase"
          >
            Your details
          </h2>
          <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 text-sm">
            {rows
              .filter(([, value]) => value)
              .map(([label, value]) => (
                <div key={label} className="contents">
                  <dt className="text-muted">{label}</dt>
                  <dd>{value}</dd>
                </div>
              ))}
          </dl>
          <p className="text-sm text-muted">To change these, ask a club administrator.</p>
        </section>
      </CardBody>
    </Card>
  );
}
