import { useQuery } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import { Route, Routes, useLocation, useNavigate } from 'react-router';
import { describe, expect, it, vi } from 'vitest';

import { ApiError } from '../../api/client';
import { isNotFound } from '../../lib/notFound';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import { CardGrid } from './CardGrid';
import { EditorLink } from './EditorLink';
import { QueryState } from './QueryState';
import { SearchInput } from './SearchInput';
import { TabsNav } from './TabsNav';
import { useSearchQuery } from './useSearchQuery';
import { useTabParam } from './useTabParam';

const anonymous = { status: 401, body: { error: { code: 401, message: 'login required' } } };

function Location() {
  const location = useLocation();
  return <output data-testid="location">{location.pathname + location.search}</output>;
}

describe('QueryState', () => {
  function Probe({ fn }: { fn: () => Promise<string[]> }) {
    const query = useQuery({ queryKey: ['probe'], queryFn: fn });
    return (
      <QueryState query={query} isEmpty={(d) => d.length === 0} emptyTitle="Nothing yet">
        {(data) => <p>{data.join(', ')}</p>}
      </QueryState>
    );
  }

  it('shows loading, then the data', async () => {
    mockFetch({ '/api/v1/auth/me': anonymous });
    renderWithProviders(<Probe fn={async () => ['a', 'b']} />);
    expect(screen.getByText('Loading')).toBeInTheDocument();
    expect(await screen.findByText('a, b')).toBeInTheDocument();
  });

  it('shows the empty state', async () => {
    mockFetch({ '/api/v1/auth/me': anonymous });
    renderWithProviders(<Probe fn={async () => []} />);
    expect(await screen.findByText('Nothing yet')).toBeInTheDocument();
  });

  it('shows the error and retries', async () => {
    mockFetch({ '/api/v1/auth/me': anonymous });
    const fn = vi
      .fn<() => Promise<string[]>>()
      .mockRejectedValueOnce(new ApiError(500, 'internal server error'))
      .mockResolvedValueOnce(['ok']);
    renderWithProviders(<Probe fn={fn} />);
    expect(await screen.findByRole('alert')).toHaveTextContent(
      "Couldn't load this: internal server error",
    );
    fireEvent.click(screen.getByRole('button', { name: 'Retry' }));
    expect(await screen.findByText('ok')).toBeInTheDocument();
  });
});

describe('CardGrid', () => {
  const items = Array.from({ length: 30 }, (_, i) => ({ id: i + 1 }));
  function Grid({ resetKey }: { resetKey?: string }) {
    return (
      <CardGrid
        items={items}
        getKey={(i) => i.id}
        render={(i) => <span>Item {i.id}</span>}
        emptyTitle="None"
        resetKey={resetKey}
      />
    );
  }

  it('shows 12, then 12 more, then the rest, then no button', () => {
    render(<Grid />);
    expect(screen.getAllByText(/^Item /)).toHaveLength(12);
    fireEvent.click(screen.getByRole('button', { name: 'Show more (18 more)' }));
    expect(screen.getAllByText(/^Item /)).toHaveLength(24);
    fireEvent.click(screen.getByRole('button', { name: 'Show more (6 more)' }));
    expect(screen.getAllByText(/^Item /)).toHaveLength(30);
    expect(screen.queryByRole('button', { name: /Show more/ })).toBeNull();
  });

  it('resets to the first page when resetKey changes', () => {
    const { rerender } = render(<Grid resetKey="future" />);
    fireEvent.click(screen.getByRole('button', { name: 'Show more (18 more)' }));
    expect(screen.getAllByText(/^Item /)).toHaveLength(24);
    rerender(<Grid resetKey="future" />);
    expect(screen.getAllByText(/^Item /)).toHaveLength(24);
    rerender(<Grid resetKey="past" />);
    expect(screen.getAllByText(/^Item /)).toHaveLength(12);
  });

  it('shows the empty state for no items', () => {
    render(<CardGrid items={[]} getKey={() => 1} render={() => null} emptyTitle="No news yet" />);
    expect(screen.getByText('No news yet')).toBeInTheDocument();
  });
});

describe('TabsNav and useTabParam', () => {
  const tabs = [
    { value: 'future', label: 'Upcoming' },
    { value: 'past', label: 'Past' },
    { value: 'all', label: 'All' },
  ];

  function Page() {
    const period = useTabParam('period', ['future', 'past', 'all'], 'future');
    const navigate = useNavigate();
    return (
      <>
        <TabsNav param="period" tabs={tabs} defaultValue="future" label="Events" />
        <p>Current: {period}</p>
        <button onClick={() => navigate(-1)}>Back</button>
        <Location />
      </>
    );
  }

  function renderAt(route: string) {
    mockFetch({ '/api/v1/auth/me': anonymous });
    renderWithProviders(
      <Routes>
        <Route path="/whatson" element={<Page />} />
      </Routes>,
      { route },
    );
  }

  it('reads the active tab from the address and marks it', () => {
    renderAt('/whatson?period=past');
    expect(screen.getByText('Current: past')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Past' })).toHaveAttribute('aria-current', 'page');
    expect(screen.getByRole('link', { name: 'Upcoming' })).not.toHaveAttribute('aria-current');
  });

  it('falls back to the default for unknown values', () => {
    renderAt('/whatson?period=nonsense');
    expect(screen.getByText('Current: future')).toBeInTheDocument();
  });

  it('updates the address on click and Back restores it', () => {
    renderAt('/whatson?q=x');
    fireEvent.click(screen.getByRole('link', { name: 'All' }));
    expect(screen.getByText('Current: all')).toBeInTheDocument();
    expect(screen.getByTestId('location')).toHaveTextContent('/whatson?q=x&period=all');
    fireEvent.click(screen.getByRole('button', { name: 'Back' }));
    expect(screen.getByText('Current: future')).toBeInTheDocument();
  });
});

describe('SearchInput and useSearchQuery', () => {
  function Page() {
    const q = useSearchQuery();
    return (
      <>
        <SearchInput label="Search documents" />
        <p>Query: [{q}]</p>
        <Location />
      </>
    );
  }

  it('starts from ?q= and writes changes back to the address', () => {
    mockFetch({ '/api/v1/auth/me': anonymous });
    renderWithProviders(
      <Routes>
        <Route path="/documents" element={<Page />} />
      </Routes>,
      { route: '/documents?q=club' },
    );
    const box = screen.getByLabelText('Search documents');
    expect(box).toHaveValue('club');
    fireEvent.change(box, { target: { value: 'policy' } });
    expect(screen.getByText('Query: [policy]')).toBeInTheDocument();
    expect(screen.getByTestId('location')).toHaveTextContent('/documents?q=policy');
    fireEvent.change(box, { target: { value: '' } });
    expect(screen.getByTestId('location')).toHaveTextContent(/^\/documents$/);
  });
});

describe('EditorLink', () => {
  function user(perms: { canEdit: boolean; canManageGallery: boolean }) {
    return {
      body: {
        id: 1,
        name: 'Someone',
        email: 's@example.test',
        role: 'Manager',
        permissions: { ...perms, canManageUsers: false },
      },
    };
  }

  it('is hidden when signed out', async () => {
    mockFetch({ '/api/v1/auth/me': anonymous });
    renderWithProviders(<EditorLink legacyHref="/news" />);
    await new Promise((r) => setTimeout(r, 0));
    expect(screen.queryByRole('link')).toBeNull();
  });

  it('is hidden for a user who cannot edit', async () => {
    mockFetch({ '/api/v1/auth/me': user({ canEdit: false, canManageGallery: false }) });
    renderWithProviders(<EditorLink legacyHref="/news" />);
    await new Promise((r) => setTimeout(r, 0));
    expect(screen.queryByRole('link')).toBeNull();
  });

  it('links editors to the classic page', async () => {
    mockFetch({ '/api/v1/auth/me': user({ canEdit: true, canManageGallery: true }) });
    renderWithProviders(<EditorLink legacyHref="/news" />);
    expect(
      await screen.findByRole('link', { name: 'Manage this on the classic site ↗' }),
    ).toHaveAttribute('href', '/news');
  });

  it('follows canManageGallery for the gallery', async () => {
    mockFetch({ '/api/v1/auth/me': user({ canEdit: false, canManageGallery: true }) });
    renderWithProviders(<EditorLink legacyHref="/gallery" permission="canManageGallery" />);
    expect(await screen.findByRole('link')).toHaveAttribute('href', '/gallery');
  });
});

describe('isNotFound', () => {
  it('is true only for a 404 ApiError', () => {
    expect(isNotFound(new ApiError(404, 'not found'))).toBe(true);
    expect(isNotFound(new ApiError(500, 'boom'))).toBe(false);
    expect(isNotFound(new Error('x'))).toBe(false);
    expect(isNotFound(null)).toBe(false);
  });
});
