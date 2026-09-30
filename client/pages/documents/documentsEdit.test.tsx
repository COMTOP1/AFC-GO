import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import { documents, editor, manager, publicRoutes } from '../../test/fixtures';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import DocumentsPage from './DocumentsPage';

function renderDocs(overrides: Record<string, MockRoute> = {}) {
  const fetchMock = mockFetch(publicRoutes({ '/api/v1/auth/me': editor, ...overrides }));
  renderWithProviders(
    <Routes>
      <Route path="/documents" element={<DocumentsPage />} />
    </Routes>,
    { route: '/documents' },
  );
  return fetchMock;
}

describe('Documents editing', () => {
  it('hides the controls from a Manager', async () => {
    renderDocs({ '/api/v1/auth/me': manager });
    await screen.findByText(documents[0].name);
    expect(screen.queryByRole('button', { name: 'Add document' })).toBeNull();
    expect(screen.queryByRole('button', { name: `Delete ${documents[0].name}` })).toBeNull();
  });

  it('adds a document from the dialog', async () => {
    const fetchMock = renderDocs({
      '/api/v1/documents': (init) =>
        init?.method === 'POST'
          ? { status: 201, body: { id: 99, name: 'Rules', fileUrl: '/f' } }
          : { body: documents },
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Add document' }));
    const dialog = screen.getByRole('dialog', { name: 'Add document' });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add document' }));
    expect(within(dialog).getByLabelText('Name')).toHaveAccessibleDescription('Enter a name');
    expect(within(dialog).getByLabelText('File')).toHaveAccessibleDescription('Choose a file');
    fireEvent.change(within(dialog).getByLabelText('Name'), { target: { value: 'Rules' } });
    const file = new File(['x'], 'rules.pdf', { type: 'application/pdf' });
    fireEvent.change(within(dialog).getByLabelText('File'), { target: { files: [file] } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add document' }));
    expect(await screen.findByRole('button', { name: 'Document added' })).toBeInTheDocument();
    expect(screen.queryByRole('dialog', { name: 'Add document' })).toBeNull();
    const fd = fetchMock.mock.calls.find(([, i]) => i?.method === 'POST')?.[1]?.body as FormData;
    expect(fd.get('name')).toBe('Rules');
    expect(fd.get('file')).toBe(file);
  });

  it('deletes a document after confirming', async () => {
    const fetchMock = renderDocs({ [`/api/v1/documents/${documents[0].id}`]: { status: 204 } });
    fireEvent.click(await screen.findByRole('button', { name: `Delete ${documents[0].name}` }));
    fireEvent.click(
      within(screen.getByRole('dialog', { name: `Delete ${documents[0].name}?` })).getByRole(
        'button',
        { name: 'Delete' },
      ),
    );
    expect(await screen.findByRole('button', { name: 'Document deleted' })).toBeInTheDocument();
    await waitFor(() =>
      expect(fetchMock.mock.calls.some(([, i]) => i?.method === 'DELETE')).toBe(true),
    );
  });
});
