import { Input } from '../ui/controls';
import { Field } from '../ui/Field';

export type PasswordField = 'newPassword' | 'confirmationPassword';

export const PASSWORD_HINT =
  'More than 8 characters, with a lower-case letter, an upper-case letter, a number and a special character.';

export interface PasswordFieldsProps {
  newPassword: string;
  confirmationPassword: string;
  onChange: (field: PasswordField, value: string) => void;
  errors: Partial<Record<PasswordField, string>>;
}

/** New + confirm password inputs; the server's messages replace the rules hint. */
export function PasswordFields({
  newPassword,
  confirmationPassword,
  onChange,
  errors,
}: PasswordFieldsProps) {
  return (
    <>
      <Field label="New password" help={PASSWORD_HINT} error={errors.newPassword}>
        <Input
          type="password"
          autoComplete="new-password"
          value={newPassword}
          onChange={(e) => onChange('newPassword', e.target.value)}
        />
      </Field>
      <Field label="Confirm new password" error={errors.confirmationPassword}>
        <Input
          type="password"
          autoComplete="new-password"
          value={confirmationPassword}
          onChange={(e) => onChange('confirmationPassword', e.target.value)}
        />
      </Field>
    </>
  );
}
