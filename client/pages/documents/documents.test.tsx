import { fireEvent, screen } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';

import { publicRoutes } from '../../test/fixtures';
import type { MockRoute } from '../../test/mockFetch';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import DocumentsPage from './DocumentsPage';

function renderDocs(route = '/documents', overrides: Record<string, MockRoute> = {}) {
  mockFetch(publicRoutes(overrides));
  renderWithProviders(
    <Routes>
      <Route path="/documents" element={<DocumentsPage />} />
    </Routes>,
    { route },
  );
}

describe('DocumentsPage', () => {
  it('lists documents with labelled download links', async () => {
    renderDocs();
    expect(await screen.findByRole('link', { name: 'Download Club constitution' })).toHaveAttribute(
      'href',
      '/api/v1/files/document/14',
    );
    expect(screen.getByRole('link', { name: 'Download Safeguarding policy' })).toBeInTheDocument();
    expect(document.title).toBe('Documents · AFC Aldermaston');
  });

  it('filters as you type, ignoring case and spacing', async () => {
    renderDocs();
    const box = await screen.findByLabelText('Search documents');
    fireEvent.change(box, { target: { value: '  CLUB  const' } });
    expect(screen.getByText('Club constitution')).toBeInTheDocument();
    expect(screen.queryByText('Safeguarding policy')).toBeNull();
    fireEvent.change(box, { target: { value: '' } });
    expect(screen.getByText('Safeguarding policy')).toBeInTheDocument();
  });

  it('starts filtered from ?q= and says when nothing matches', async () => {
    renderDocs('/documents?q=zzz');
    expect(await screen.findByText("No documents match 'zzz'")).toBeInTheDocument();
  });

  it('shows the empty state', async () => {
    renderDocs('/documents', { '/api/v1/documents': { body: [] } });
    expect(await screen.findByText('No documents yet')).toBeInTheDocument();
  });
});
