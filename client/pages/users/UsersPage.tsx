import { useState } from 'react';
import { useSearchParams } from 'react-router';

import { queryKeys } from '../../api/queries';
import { useTeams } from '../../api/teams';
import { deleteUser, useUsers, type AdminUser } from '../../api/users';
import { useAuth } from '../../auth/useAuth';
import { DeleteButton } from '../../components/edit/DeleteButton';
import { RequireEditor } from '../../components/edit/RequireEditor';
import { QueryState } from '../../components/page/QueryState';
import { SearchInput } from '../../components/page/SearchInput';
import { Thumb } from '../../components/page/Thumb';
import { usePageTitle } from '../../components/page/usePageTitle';
import { useSearchQuery } from '../../components/page/useSearchQuery';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { Select } from '../../components/ui/controls';
import { EmptyState } from '../../components/ui/EmptyState';
import { Field } from '../../components/ui/Field';
import { PageHeader } from '../../components/ui/PageHeader';
import { Table, TBody, Td, Th, THead, Tr } from '../../components/ui/Table';
import { MANAGER_ROLE, ROLES } from '../../lib/roles';
import { matchesQuery } from '../../lib/text';
import { DisplayEmailCard } from './DisplayEmailCard';
import { ResetPasswordButton } from './ResetPasswordButton';
import { SecretDialog, type Secret } from './SecretDialog';
import { UserDialog } from './UserDialog';

interface UsersTableProps {
  currentUserId?: number;
  onEdit: (user: AdminUser) => void;
  onSecret: (secret: Secret) => void;
}

function UsersTable({ currentUserId, onEdit, onSecret }: UsersTableProps) {
  const users = useUsers();
  const teams = useTeams();
  const [params, setParams] = useSearchParams();
  const q = useSearchQuery();
  const roleFilter = params.get('role') ?? '';
  const teamName = (id?: number) => teams.data?.find((t) => t.id === id)?.name ?? '';

  return (
    <>
      <div className="mb-4 flex flex-wrap items-end gap-4">
        <SearchInput label="Search users" />
        <Field label="Role" className="w-56">
          <Select
            value={roleFilter}
            onChange={(e) => {
              const next = new URLSearchParams(params);
              if (e.target.value) {
                next.set('role', e.target.value);
              } else {
                next.delete('role');
              }
              setParams(next, { replace: true });
            }}
          >
            <option value="">All roles</option>
            {ROLES.map((r) => (
              <option key={r.code} value={r.code}>
                {r.label}
              </option>
            ))}
          </Select>
        </Field>
      </div>
      <QueryState query={users}>
        {(list) => {
          const shown = [...list]
            .sort((a, b) => a.name.localeCompare(b.name))
            .filter(
              (u) =>
                matchesQuery(`${u.name} ${u.email} ${u.role}`, q) &&
                (!roleFilter || u.roleCode === roleFilter),
            );
          if (shown.length === 0) {
            return <EmptyState title="No users match your filters" />;
          }
          return (
            <Table>
              <THead>
                <Tr>
                  <Th>
                    <span className="sr-only">Photo</span>
                  </Th>
                  <Th>Name</Th>
                  <Th>Email</Th>
                  <Th>Phone</Th>
                  <Th>Role</Th>
                  <Th>Team</Th>
                  <Th>
                    <span className="sr-only">Actions</span>
                  </Th>
                </Tr>
              </THead>
              <TBody>
                {shown.map((u) => {
                  const isSelf = u.id === currentUserId;
                  return (
                    <Tr key={u.id}>
                      <Td>
                        <Thumb src={u.imageUrl} />
                      </Td>
                      <Td>
                        <span className="font-medium">{u.name}</span>
                        {isSelf && (
                          <Badge tone="blue" className="ml-2">
                            You
                          </Badge>
                        )}
                      </Td>
                      <Td>
                        <a href={`mailto:${u.email}`} className="text-red underline">
                          {u.email}
                        </a>
                      </Td>
                      <Td className="whitespace-nowrap">{u.phone}</Td>
                      <Td>{u.role}</Td>
                      <Td>{u.roleCode === MANAGER_ROLE ? teamName(u.teamId) : ''}</Td>
                      <Td>
                        <div className="flex justify-end gap-2">
                          <Button
                            size="sm"
                            variant="secondary"
                            aria-label={`Edit ${u.name}`}
                            onClick={() => onEdit(u)}
                          >
                            Edit
                          </Button>
                          <ResetPasswordButton user={u} onSecret={onSecret} />
                          {!isSelf && (
                            <DeleteButton
                              ariaLabel={`Delete ${u.name}`}
                              confirmTitle={`Delete ${u.name}?`}
                              confirmMessage="They will no longer be able to sign in. This can't be undone."
                              onDelete={() => deleteUser(u.id)}
                              invalidate={[queryKeys.users, queryKeys.contact, ['team']]}
                              successMessage="User deleted"
                            />
                          )}
                        </div>
                      </Td>
                    </Tr>
                  );
                })}
              </TBody>
            </Table>
          );
        }}
      </QueryState>
    </>
  );
}

export default function UsersPage() {
  usePageTitle('Users');
  const { user } = useAuth();
  const canManageUsers = user?.permissions.canManageUsers ?? false;
  // null = closed; {} = adding; { user } = editing.
  const [dialog, setDialog] = useState<{ user?: AdminUser } | null>(null);
  const [secret, setSecret] = useState<Secret | null>(null);
  return (
    <>
      <PageHeader
        title="Users"
        actions={canManageUsers && <Button onClick={() => setDialog({})}>Add user</Button>}
      />
      <RequireEditor permission="canManageUsers">
        <DisplayEmailCard />
        <UsersTable
          currentUserId={user?.id}
          onEdit={(u) => setDialog({ user: u })}
          onSecret={setSecret}
        />
        <UserDialog
          open={dialog !== null}
          user={dialog?.user}
          isSelf={dialog?.user !== undefined && dialog.user.id === user?.id}
          onClose={() => setDialog(null)}
          onSecret={setSecret}
        />
        <SecretDialog secret={secret} onClose={() => setSecret(null)} />
      </RequireEditor>
    </>
  );
}
