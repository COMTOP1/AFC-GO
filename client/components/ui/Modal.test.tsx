import { act, fireEvent, render, screen } from '@testing-library/react';
import { useState } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { ConfirmDialog } from './ConfirmDialog';
import { Modal } from './Modal';

function Harness({ onClose }: { onClose: () => void }) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <button onClick={() => setOpen(true)}>Open</button>
      <Modal
        open={open}
        onClose={() => {
          onClose();
          setOpen(false);
        }}
        title="Delete team?"
        actions={<button onClick={() => setOpen(false)}>Cancel</button>}
      >
        <p>Body text</p>
      </Modal>
    </>
  );
}

function openHarness() {
  const onClose = vi.fn();
  render(<Harness onClose={onClose} />);
  const opener = screen.getByRole('button', { name: 'Open' });
  opener.focus();
  fireEvent.click(opener);
  return { onClose, opener };
}

describe('Modal', () => {
  it('renders nothing inside the dialog while closed', () => {
    render(<Harness onClose={vi.fn()} />);
    expect(screen.queryByText('Body text')).toBeNull();
  });

  it('opens as a labelled dialog', () => {
    openHarness();
    const dialog = screen.getByRole('dialog', { name: 'Delete team?' });
    expect(dialog).toHaveAttribute('open');
    expect(screen.getByText('Body text')).toBeInTheDocument();
  });

  it('calls onClose once on Esc (the cancel event)', () => {
    const { onClose } = openHarness();
    fireEvent(screen.getByRole('dialog'), new Event('cancel', { cancelable: true }));
    expect(onClose).toHaveBeenCalledOnce();
    expect(screen.queryByText('Body text')).toBeNull();
  });

  it('closes on a backdrop click but not a click inside', () => {
    const { onClose } = openHarness();
    fireEvent.click(screen.getByText('Body text'));
    expect(onClose).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('dialog'));
    expect(onClose).toHaveBeenCalledOnce();
  });

  it('returns focus to the opener when closed by the parent', () => {
    const { opener, onClose } = openHarness();
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(screen.queryByText('Body text')).toBeNull();
    expect(document.activeElement).toBe(opener);
    expect(onClose).not.toHaveBeenCalled();
  });

  it('closes cleanly when unmounted while open', () => {
    const opener = document.createElement('button');
    document.body.append(opener);
    opener.focus();
    const { unmount } = render(
      <Modal open onClose={vi.fn()} title="Sign in">
        <p>Form</p>
      </Modal>,
    );
    expect(() => unmount()).not.toThrow();
    expect(document.activeElement).toBe(opener);
    opener.remove();
  });
});

describe('ConfirmDialog', () => {
  it('confirms and cancels', async () => {
    const onConfirm = vi.fn();
    const onCancel = vi.fn();
    render(
      <ConfirmDialog
        open
        title="Sign out?"
        message="You will need to sign in again."
        confirmLabel="Sign out"
        onConfirm={onConfirm}
        onCancel={onCancel}
      />,
    );
    expect(screen.getByRole('dialog', { name: 'Sign out?' })).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Sign out' }));
    expect(onConfirm).toHaveBeenCalledOnce();
    // Cancel is disabled while onConfirm is pending; let it settle first.
    await act(async () => {});
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(onCancel).toHaveBeenCalledOnce();
  });

  it('shows a pending state until onConfirm settles', async () => {
    let finish: () => void = () => {};
    const onConfirm = () =>
      new Promise<void>((resolve) => {
        finish = resolve;
      });
    render(
      <ConfirmDialog
        open
        title="Delete team?"
        message="This can't be undone."
        confirmLabel="Delete"
        tone="danger"
        onConfirm={onConfirm}
        onCancel={vi.fn()}
      />,
    );
    const confirm = screen.getByRole('button', { name: 'Delete' });
    fireEvent.click(confirm);
    expect(confirm).toHaveAttribute('aria-busy', 'true');
    expect(confirm).toBeDisabled();
    await act(async () => finish());
    expect(confirm).not.toHaveAttribute('aria-busy');
  });
});
