import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ApiError } from '../../api/client';
import { Checkbox } from './Checkbox';
import { FileInput, Input, Select, Textarea } from './controls';
import { Field } from './Field';
import { fieldError } from './fieldError';

describe('Field', () => {
  it('labels the control and describes it with the help text', () => {
    render(
      <Field label="Email" help="We never share it.">
        <Input type="email" />
      </Field>,
    );
    const input = screen.getByLabelText('Email');
    expect(input).toHaveAccessibleDescription('We never share it.');
    expect(input).not.toHaveAttribute('aria-invalid');
  });

  it('shows the error instead of the help and marks the control invalid', () => {
    render(
      <Field label="Email" help="We never share it." error="Enter a valid email address.">
        <Input type="email" />
      </Field>,
    );
    const input = screen.getByLabelText('Email');
    expect(input).toHaveAccessibleDescription('Enter a valid email address.');
    expect(input).toHaveAttribute('aria-invalid', 'true');
    expect(screen.queryByText('We never share it.')).toBeNull();
  });

  it('keeps an explicit id', () => {
    render(
      <Field label="Name" id="player-name">
        <Input />
      </Field>,
    );
    expect(screen.getByLabelText('Name')).toHaveAttribute('id', 'player-name');
  });

  it('wires textarea, select and file inputs too', () => {
    render(
      <>
        <Field label="Bio">
          <Textarea />
        </Field>
        <Field label="Team">
          <Select>
            <option>Under 12s</option>
          </Select>
        </Field>
        <Field label="Photo" error="Too big">
          <FileInput />
        </Field>
      </>,
    );
    expect(screen.getByLabelText('Bio').tagName).toBe('TEXTAREA');
    expect(screen.getByLabelText('Team').tagName).toBe('SELECT');
    const file = screen.getByLabelText('Photo');
    expect(file).toHaveAttribute('type', 'file');
    expect(file).toHaveAccessibleDescription('Too big');
  });

  it('works outside a Field with plain props', () => {
    render(<Input aria-label="Search" />);
    expect(screen.getByRole('textbox', { name: 'Search' })).toBeInTheDocument();
  });
});

describe('Checkbox', () => {
  it('toggles from its label', () => {
    render(<Checkbox label="Youth team" />);
    const box = screen.getByRole('checkbox', { name: 'Youth team' });
    fireEvent.click(screen.getByText('Youth team'));
    expect(box).toBeChecked();
  });
});

describe('fieldError', () => {
  it('reads a field message from an ApiError', () => {
    const err = new ApiError(422, 'invalid', { email: 'already in use' });
    expect(fieldError(err, 'email')).toBe('already in use');
    expect(fieldError(err, 'name')).toBeUndefined();
  });

  it('ignores other errors', () => {
    expect(fieldError(new Error('boom'), 'email')).toBeUndefined();
    expect(fieldError(null, 'email')).toBeUndefined();
  });
});
