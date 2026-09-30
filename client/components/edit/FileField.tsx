import { formatSize } from '../../lib/editForm';
import { FileInput } from '../ui/controls';
import { Field } from '../ui/Field';

export interface FileFieldProps {
  label: string;
  value: File | null;
  onChange: (file: File | null) => void;
  error?: string;
  accept?: string;
}

export function FileField({ label, value, onChange, error, accept }: FileFieldProps) {
  return (
    <Field
      label={label}
      error={error}
      help={value ? `${value.name} (${formatSize(value.size)})` : undefined}
    >
      <FileInput accept={accept} onChange={(e) => onChange(e.target.files?.[0] ?? null)} />
    </Field>
  );
}
