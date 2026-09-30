import { useState, type FormEvent } from 'react';
import { useNavigate, useParams } from 'react-router';

import { createTeam, updateTeam, useTeam } from '../../api/teams';
import type { TeamSummary } from '../../api/types';
import { ImageField } from '../../components/edit/ImageField';
import { RequireEditor } from '../../components/edit/RequireEditor';
import { useSaveForm } from '../../components/edit/useSaveForm';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Checkbox } from '../../components/ui/Checkbox';
import { Input, Select, Textarea } from '../../components/ui/controls';
import { Field } from '../../components/ui/Field';
import { PageHeader } from '../../components/ui/PageHeader';
import { useToast } from '../../components/ui/toast/useToast';
import { AGE_GROUPS } from '../../lib/ageGroups';
import { parseId } from '../../lib/ids';
import { emptyImage, type ImageValue } from '../../lib/images';
import { isNotFound } from '../../lib/notFound';
import NotFoundPage from '../NotFoundPage';

type TextKey =
  'name' | 'description' | 'league' | 'division' | 'leagueTable' | 'fixtures' | 'coach' | 'physio';

function TeamForm({ team }: { team?: TeamSummary }) {
  const heading = team ? 'Edit team' : 'New team';
  usePageTitle(heading);
  const navigate = useNavigate();
  const toast = useToast();
  const [text, setText] = useState<Record<TextKey, string>>({
    name: team?.name ?? '',
    description: team?.description ?? '',
    league: team?.league ?? '',
    division: team?.division ?? '',
    leagueTable: team?.leagueTableUrl ?? '',
    fixtures: team?.fixturesUrl ?? '',
    coach: team?.coach ?? '',
    physio: team?.physio ?? '',
  });
  const [ages, setAges] = useState(team ? String(team.ages) : '');
  const [isActive, setIsActive] = useState(team?.isActive ?? true);
  const [isYouth, setIsYouth] = useState(team?.isYouth ?? false);
  const [image, setImage] = useState<ImageValue>(emptyImage);
  const [missing, setMissing] = useState<{ name?: string; ages?: string }>({});

  const save = useSaveForm({
    submit: () => {
      const input = { ...text, ages: Number(ages), isActive, isYouth, image };
      return team ? updateTeam(team.id, input) : createTeam(input);
    },
    invalidate: [['teams'], ['team'], ['site']],
    onSaved: (t) => {
      toast.show({ tone: 'success', message: 'Team saved' });
      navigate(`/team/${t.id}`);
    },
  });

  const set = (key: TextKey) => (value: string) => setText((t) => ({ ...t, [key]: value }));
  const err = (key: string) => save.fieldErrors[key];

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const next = {
      name: text.name.trim() ? undefined : 'Enter a name',
      ages: ages ? undefined : 'Choose the age group',
    };
    setMissing(next);
    if (next.name || next.ages) {
      return;
    }
    void save.run();
  }

  const textField = (key: TextKey, label: string, type = 'text') => (
    <Field label={label} error={err(key)}>
      <Input type={type} value={text[key]} onChange={(e) => set(key)(e.target.value)} />
    </Field>
  );

  return (
    <>
      <PageHeader title={heading} />
      <form onSubmit={onSubmit} noValidate className="grid max-w-3xl gap-5 md:grid-cols-2">
        {save.formError && (
          <Alert tone="error" className="md:col-span-2">
            {save.formError}
          </Alert>
        )}
        <Field label="Name" error={missing.name ?? err('name')}>
          <Input value={text.name} onChange={(e) => set('name')(e.target.value)} />
        </Field>
        <Field label="Age group" error={missing.ages ?? err('ages')}>
          <Select value={ages} onChange={(e) => setAges(e.target.value)}>
            <option value="">Choose…</option>
            {AGE_GROUPS.map((g) => (
              <option key={g.value} value={String(g.value)}>
                {g.label}
              </option>
            ))}
          </Select>
        </Field>
        <Field label="Description" error={err('description')} className="md:col-span-2">
          <Textarea value={text.description} onChange={(e) => set('description')(e.target.value)} />
        </Field>
        {textField('league', 'League')}
        {textField('division', 'Division')}
        {textField('leagueTable', 'League table URL', 'url')}
        {textField('fixtures', 'Fixtures URL', 'url')}
        {textField('coach', 'Coach')}
        {textField('physio', 'Physio')}
        <div className="md:col-span-2">
          <ImageField
            label="Team photo"
            currentUrl={team?.imageUrl}
            allowRemove={Boolean(team)}
            value={image}
            onChange={setImage}
            error={err('file') ?? err('image')}
          />
        </div>
        <Checkbox
          label="Active team"
          checked={isActive}
          onChange={(e) => setIsActive(e.target.checked)}
        />
        <Checkbox
          label="Youth team"
          checked={isYouth}
          onChange={(e) => setIsYouth(e.target.checked)}
        />
        <div className="flex gap-2 md:col-span-2">
          <Button type="submit" loading={save.busy}>
            Save team
          </Button>
          <Button variant="secondary" onClick={() => navigate(-1)} disabled={save.busy}>
            Cancel
          </Button>
        </div>
      </form>
    </>
  );
}

function EditTeam({ id }: { id: number | null }) {
  const detail = useTeam(id);
  if (id === null || isNotFound(detail.error)) {
    return <NotFoundPage />;
  }
  return <QueryState query={detail}>{(d) => <TeamForm team={d.team} />}</QueryState>;
}

export default function TeamFormPage() {
  const { id } = useParams();
  return (
    <RequireEditor>{id === undefined ? <TeamForm /> : <EditTeam id={parseId(id)} />}</RequireEditor>
  );
}
