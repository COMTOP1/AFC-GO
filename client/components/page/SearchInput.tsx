import { useSearchParams } from 'react-router';

import { Input } from '../ui/controls';
import { Field } from '../ui/Field';

/**
 * A search box bound to ?q=; edits replace the history entry so Back skips keystrokes.
 * It is uncontrolled on purpose: BrowserRouter commits address changes in a
 * transition, so a box driven by ?q= would snap back (and lose the caret) while
 * the address catches up. It starts from ?q= and remounts with each page.
 */
export function SearchInput({ label }: { label: string }) {
  const [params, setParams] = useSearchParams();
  return (
    <Field label={label} className="max-w-md">
      <Input
        type="search"
        defaultValue={params.get('q') ?? ''}
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
