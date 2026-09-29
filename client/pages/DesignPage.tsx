import { useState, type ReactNode } from 'react';

import { Alert } from '../components/ui/Alert';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { ButtonLink } from '../components/ui/ButtonLink';
import { Card, CardBody, CardMedia } from '../components/ui/Card';
import { Checkbox } from '../components/ui/Checkbox';
import { ConfirmDialog } from '../components/ui/ConfirmDialog';
import { FileInput, Input, Select, Textarea } from '../components/ui/controls';
import { EmptyState } from '../components/ui/EmptyState';
import { Field } from '../components/ui/Field';
import { Menu } from '../components/ui/Menu';
import { Modal } from '../components/ui/Modal';
import { PageHeader } from '../components/ui/PageHeader';
import { Skeleton } from '../components/ui/Skeleton';
import { Spinner } from '../components/ui/Spinner';
import { Table, TBody, Td, Th, THead, Tr } from '../components/ui/Table';
import { useToast } from '../components/ui/toast/useToast';

const swatches = [
  'red',
  'red-hover',
  'club-red',
  'blue',
  'footer',
  'ink',
  'muted',
  'line',
  'bg',
  'surface',
  'field',
  'success',
  'warning',
];

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="mb-10">
      <h2 className="mb-3 font-display text-2xl font-extrabold tracking-wide uppercase">{title}</h2>
      {children}
    </section>
  );
}

// Every shared component in every state: the reference for page ports and visual review.
export default function DesignPage() {
  const toast = useToast();
  const [modalOpen, setModalOpen] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);

  return (
    <>
      <PageHeader
        title="Design system"
        subtitle="Shared components and tokens for the AFC Aldermaston site."
      />

      <Section title="Palette">
        <ul className="grid grid-cols-2 gap-3 sm:grid-cols-4 lg:grid-cols-6">
          {swatches.map((name) => (
            <li key={name} className="overflow-hidden rounded-lg border border-line">
              <div className="h-12" style={{ background: `var(--color-${name})` }} />
              <p className="px-2 py-1 text-xs">{name}</p>
            </li>
          ))}
        </ul>
      </Section>

      <Section title="Buttons">
        <div className="flex flex-wrap items-center gap-2">
          <Button>Save</Button>
          <Button variant="secondary">Cancel</Button>
          <Button variant="ghost">View all</Button>
          <Button variant="danger">Delete</Button>
          <Button size="sm">Small</Button>
          <Button loading>Saving</Button>
          <Button disabled>Disabled</Button>
          <ButtonLink to="/" variant="secondary">
            Router link
          </ButtonLink>
          <Spinner />
        </div>
      </Section>

      <Section title="Badges and notices">
        <div className="mb-3 flex flex-wrap gap-2">
          <Badge tone="red">U12s</Badge>
          <Badge tone="blue">Webmaster</Badge>
          <Badge>Draft</Badge>
        </div>
        <div className="grid gap-2">
          <Alert tone="success">Team saved.</Alert>
          <Alert tone="error">Couldn&apos;t save: the email is already in use.</Alert>
          <Alert tone="warning">This team has no manager assigned.</Alert>
          <Alert tone="info">Fixtures are published on Fridays.</Alert>
        </div>
      </Section>

      <Section title="Cards">
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardMedia src="/app/favicon.png" alt="Club crest" />
            <CardBody>
              <Badge tone="red">News</Badge>
              <p className="mt-2 font-display text-xl font-extrabold uppercase">With an image</p>
            </CardBody>
          </Card>
          <Card>
            <CardMedia alt="" />
            <CardBody>
              <Badge tone="red">News</Badge>
              <p className="mt-2 font-display text-xl font-extrabold uppercase">No image</p>
              <p className="text-xs text-muted">28 Sep 2026</p>
            </CardBody>
          </Card>
          <Card>
            <CardMedia src="/app/does-not-exist.png" alt="" />
            <CardBody>
              <Badge tone="blue">What&apos;s on</Badge>
              <p className="mt-2 font-display text-xl font-extrabold uppercase">Broken image</p>
            </CardBody>
          </Card>
          <Card>
            <CardBody className="space-y-2">
              <Skeleton className="w-2/5" />
              <Skeleton />
              <Skeleton className="w-3/4" />
            </CardBody>
          </Card>
        </div>
        <Card className="mt-4">
          <EmptyState
            title="No news yet"
            message="Articles appear here once they are published."
            action={<Button variant="secondary">Refresh</Button>}
          />
        </Card>
      </Section>

      <Section title="Forms">
        <form className="grid max-w-2xl gap-4 sm:grid-cols-2" onSubmit={(e) => e.preventDefault()}>
          <Field label="Name" help="As shown on the team page.">
            <Input defaultValue="Jo Smith" />
          </Field>
          <Field label="Email" error="Enter a valid email address.">
            <Input type="email" defaultValue="jo@" />
          </Field>
          <Field label="Team">
            <Select defaultValue="u12">
              <option value="u12">Under 12s</option>
              <option value="first">First Team</option>
            </Select>
          </Field>
          <Field label="Photo">
            <FileInput accept="image/*" />
          </Field>
          <Field label="Notes" className="sm:col-span-2">
            <Textarea />
          </Field>
          <Checkbox label="Youth team" defaultChecked />
        </form>
      </Section>

      <Section title="Table">
        <Table>
          <THead>
            <Tr>
              <Th>Name</Th>
              <Th>Role</Th>
              <Th>Team</Th>
            </Tr>
          </THead>
          <TBody>
            <Tr>
              <Td>Jo Smith</Td>
              <Td>
                <Badge tone="blue">Manager</Badge>
              </Td>
              <Td>Under 12s</Td>
            </Tr>
            <Tr>
              <Td>Sam Patel</Td>
              <Td>
                <Badge tone="blue">Webmaster</Badge>
              </Td>
              <Td>—</Td>
            </Tr>
          </TBody>
        </Table>
      </Section>

      <Section title="Dialogs, menus and toasts">
        <div className="flex flex-wrap items-center gap-2">
          <Button variant="secondary" onClick={() => setModalOpen(true)}>
            Open modal
          </Button>
          <Button variant="danger" onClick={() => setConfirmOpen(true)}>
            Open confirm
          </Button>
          <Menu
            label="Sample menu ▾"
            triggerLabel="Sample menu"
            align="start"
            items={[
              { label: 'Players', href: '/players' },
              { label: 'Account', href: '/account' },
              { label: 'Say hello', onSelect: () => toast.show({ message: 'Hello' }) },
            ]}
          />
          <Button
            variant="secondary"
            onClick={() => toast.show({ tone: 'success', message: 'Player saved' })}
          >
            Success toast
          </Button>
          <Button
            variant="secondary"
            onClick={() => toast.show({ tone: 'error', message: "Couldn't save the player" })}
          >
            Error toast
          </Button>
          <Button
            variant="secondary"
            onClick={() => toast.show({ tone: 'info', message: 'Fixtures updated' })}
          >
            Info toast
          </Button>
        </div>
        <Modal
          open={modalOpen}
          onClose={() => setModalOpen(false)}
          title="Sample modal"
          actions={<Button onClick={() => setModalOpen(false)}>Close</Button>}
        >
          <p className="text-sm">Esc, the backdrop or Close dismiss this dialog.</p>
        </Modal>
        <ConfirmDialog
          open={confirmOpen}
          title="Delete team?"
          message="This removes Under 12s and unlinks its players. This can't be undone."
          confirmLabel="Delete"
          tone="danger"
          onConfirm={() => {
            setConfirmOpen(false);
            toast.show({ tone: 'success', message: 'Team deleted (not really)' });
          }}
          onCancel={() => setConfirmOpen(false)}
        />
      </Section>
    </>
  );
}
