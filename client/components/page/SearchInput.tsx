import { useSearchParams } from 'react-router';

import { Input } from '../ui/controls';
import { Field } from '../ui/Field';

/** A search box bound to ?q=; edits replace the history entry so Back skips keystrokes. */
export function SearchInput({ label }: { label: string }) {
  const [params, setParams] = useSearchParams();
  return (
    <Field label={label} className="max-w-md">
      <Input
        type="search"
        value={params.get('q') ?? ''}
        onChange={(e) => {
          const next = new URLSearchParams(params);
          if (e.target.value) {
            next.set('q', e.target.value);
          } else {
            next.delete('q');
          }
          setParams(next, { replace: true });
        }}
      />
    </Field>
  );
}
