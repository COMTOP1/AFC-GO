import { fireEvent, screen, waitFor } from '@testing-library/react';
import { useState } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { ApiError } from '../../api/client';
import { anonymous, editor, manager, photographer, publicRoutes } from '../../test/fixtures';
import type { MockResponse } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import { emptyImage, type ImageValue } from '../../lib/images';
import { DeleteButton } from './DeleteButton';
import { FileField } from './FileField';
import { ImageField } from './ImageField';
import { RequireEditor } from './RequireEditor';
import { useCanEdit } from './useCanEdit';
import { useSaveForm } from './useSaveForm';

let revokeObjectURL: ReturnType<typeof vi.fn>;
beforeEach(() => {
  let n = 0;
  revokeObjectURL = vi.fn();
  Object.assign(URL, { createObjectURL: vi.fn(() => `blob:p-${++n}`), revokeObjectURL });
});
afterEach(() => {
  delete (URL as unknown as Record<string, unknown>).createObjectURL;
  delete (URL as unknown as Record<string, unknown>).revokeObjectURL;
});

function CanEditProbe() {
  const { canEdit, canManageGallery } = useCanEdit();
  return <p>{`edit:${canEdit} gallery:${canManageGallery}`}</p>;
}

describe('useCanEdit', () => {
  it.each([
    [anonymous, 'edit:false gallery:false'],
    [manager, 'edit:false gallery:false'],
    [photographer, 'edit:false gallery:true'],
    [editor, 'edit:true gallery:true'],
  ] as [MockResponse, string][])('reflects permissions', async (me, text) => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': me }));
    renderWithProviders(<CanEditProbe />);
    expect(await screen.findByText(text)).toBeInTheDocument();
  });
});

describe('RequireEditor', () => {
  it('shows the permission message to non-editors', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': manager }));
    renderWithProviders(
      <RequireEditor>
        <p>secret form</p>
      </RequireEditor>,
    );
    expect(await screen.findByText("You don't have permission to edit this")).toBeInTheDocument();
    expect(screen.queryByText('secret form')).toBeNull();
  });

  it('renders the children for editors, and gallery access for photographers', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': photographer }));
    renderWithProviders(
      <>
        <RequireEditor permission="canManageGallery">
          <p>photo form</p>
        </RequireEditor>
        <RequireEditor>
          <p>article form</p>
        </RequireEditor>
      </>,
    );
    expect(await screen.findByText('photo form')).toBeInTheDocument();
    expect(screen.queryByText('article form')).toBeNull();
  });
});

describe('useSaveForm', () => {
  function Form({ submit, onSaved }: { submit: () => Promise<unknown>; onSaved?: () => void }) {
    const save = useSaveForm({ submit, invalidate: [['news']], onSaved });
    return (
      <>
        <button onClick={() => void save.run()}>Save</button>
        <p>busy:{String(save.busy)}</p>
        <p>field:{save.fieldErrors.title ?? ''}</p>
        <p>form:{save.formError ?? ''}</p>
      </>
    );
  }

  it('invalidates and calls onSaved on success', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': editor }));
    const onSaved = vi.fn();
    const { queryClient } = renderWithProviders(
      <Form submit={async () => ({ id: 1 })} onSaved={onSaved} />,
    );
    const spy = vi.spyOn(queryClient, 'invalidateQueries');
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    await waitFor(() => expect(onSaved).toHaveBeenCalledWith({ id: 1 }));
    expect(spy).toHaveBeenCalledWith({ queryKey: ['news'] });
  });

  it('puts field errors on fields and explains 413s', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': editor }));
    let n = 0;
    renderWithProviders(
      <Form
        submit={async () => {
          n++;
          if (n === 1) throw new ApiError(422, 'invalid', { title: 'title is required' });
          throw new ApiError(413, 'Request Entity Too Large');
        }}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    expect(await screen.findByText('field:title is required')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    expect(
      await screen.findByText('form:That file is too large (15 MB maximum).'),
    ).toBeInTheDocument();
  });

  it('re-reads the session on a 401', async () => {
    let me: MockResponse = editor;
    const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': () => me }));
    renderWithProviders(
      <Form
        submit={async () => {
          me = anonymous;
          throw new ApiError(401, 'login required');
        }}
      />,
    );
    await screen.findByText('busy:false');
    const before = fetchMock.mock.calls.filter(([u]) => String(u).endsWith('/auth/me')).length;
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    await waitFor(() =>
      expect(
        fetchMock.mock.calls.filter(([u]) => String(u).endsWith('/auth/me')).length,
      ).toBeGreaterThan(before),
    );
    expect(screen.getByText('form:')).toBeInTheDocument();
  });
});

describe('ImageField', () => {
  function Harness({ current }: { current?: string }) {
    const [value, setValue] = useState<ImageValue>(emptyImage);
    return (
      <>
        <ImageField
          label="Image"
          currentUrl={current}
          allowRemove
          value={value}
          onChange={setValue}
        />
        <p>
          file:{value.file?.name ?? ''} remove:{String(value.remove)}
        </p>
      </>
    );
  }

  function pick(file: File) {
    fireEvent.change(screen.getByLabelText('Image'), { target: { files: [file] } });
  }

  it('previews a chosen image and reports it', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<Harness current="/img/1" />);
    pick(new File(['x'], 'a.png', { type: 'image/png' }));
    expect(screen.getByRole('img', { name: 'Preview of the new image' })).toHaveAttribute(
      'src',
      'blob:p-1',
    );
    expect(screen.getByText('file:a.png remove:false')).toBeInTheDocument();
  });

  it('refuses a non-image without previewing it', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<Harness />);
    pick(new File(['<b>'], 'a.html', { type: 'text/html' }));
    expect(screen.getByLabelText('Image')).toHaveAccessibleDescription(
      'Choose an image file (JPEG, PNG, GIF, WebP, AVIF, APNG or SVG).',
    );
    expect(screen.queryByRole('img', { name: 'Preview of the new image' })).toBeNull();
    expect(URL.createObjectURL).not.toHaveBeenCalled();
  });

  it('offers Remove image for a current image, and a new file clears it', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<Harness current="/img/1" />);
    fireEvent.click(screen.getByRole('checkbox', { name: 'Remove image' }));
    expect(screen.getByText('file: remove:true')).toBeInTheDocument();
    pick(new File(['x'], 'b.png', { type: 'image/png' }));
    expect(screen.getByText('file:b.png remove:false')).toBeInTheDocument();
    pick(new File(['x'], 'c.png', { type: 'image/png' }));
    expect(revokeObjectURL).toHaveBeenCalledWith('blob:p-1');
  });

  it('has no Remove image without a current image', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(<Harness />);
    expect(screen.queryByRole('checkbox', { name: 'Remove image' })).toBeNull();
  });
});

describe('FileField', () => {
  it('shows the chosen file name and size', async () => {
    function Harness() {
      const [file, setFile] = useState<File | null>(null);
      return <FileField label="File" value={file} onChange={setFile} />;
    }
    mockFetch(publicRoutes());
    renderWithProviders(<Harness />);
    fireEvent.change(screen.getByLabelText('File'), {
      target: { files: [new File(['x'.repeat(2048)], 'rules.pdf', { type: 'application/pdf' })] },
    });
    expect(screen.getByLabelText('File')).toHaveAccessibleDescription('rules.pdf (2 KB)');
  });
});

describe('DeleteButton', () => {
  it('confirms, deletes, toasts and calls after', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': editor }));
    const onDelete = vi.fn(async () => undefined);
    const after = vi.fn();
    renderWithProviders(
      <DeleteButton
        ariaLabel="Delete Club rules"
        confirmTitle="Delete Club rules?"
        confirmMessage="This can't be undone."
        onDelete={onDelete}
        successMessage="Document deleted"
        after={after}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Delete Club rules' }));
    screen.getByRole('dialog', { name: 'Delete Club rules?' });
    fireEvent.click(screen.getAllByRole('button', { name: 'Delete' }).at(-1) as HTMLElement);
    expect(await screen.findByRole('button', { name: 'Document deleted' })).toBeInTheDocument();
    expect(onDelete).toHaveBeenCalledOnce();
    expect(after).toHaveBeenCalledOnce();
    expect(screen.queryByRole('dialog', { name: 'Delete Club rules?' })).toBeNull();
  });

  it('does nothing on Cancel, and toasts a failure', async () => {
    mockFetch(publicRoutes({ '/api/v1/auth/me': editor }));
    const onDelete = vi.fn(async () => {
      throw new ApiError(500, 'boom');
    });
    renderWithProviders(
      <DeleteButton
        confirmTitle="Delete it?"
        confirmMessage="Sure?"
        onDelete={onDelete}
        successMessage="Deleted"
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Delete' }));
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(onDelete).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: 'Delete' }));
    fireEvent.click(screen.getAllByRole('button', { name: 'Delete' }).at(-1) as HTMLElement);
    expect(
      await screen.findByRole('button', { name: "Couldn't delete: boom" }),
    ).toBeInTheDocument();
  });
});
