import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { describe, expect, it, vi } from 'vitest';

import { Alert } from './Alert';
import { Badge } from './Badge';
import { Button } from './Button';
import { ButtonLink } from './ButtonLink';
import { Card, CardBody, CardMedia } from './Card';
import { EmptyState } from './EmptyState';
import { PageHeader } from './PageHeader';
import { Spinner } from './Spinner';

describe('Button', () => {
  it('defaults to type=button and fires onClick', () => {
    const onClick = vi.fn();
    render(<Button onClick={onClick}>Save</Button>);
    const button = screen.getByRole('button', { name: 'Save' });
    expect(button).toHaveAttribute('type', 'button');
    fireEvent.click(button);
    expect(onClick).toHaveBeenCalledOnce();
  });

  it('is busy and disabled while loading', () => {
    const onClick = vi.fn();
    render(
      <Button loading onClick={onClick}>
        Save
      </Button>,
    );
    const button = screen.getByRole('button', { name: 'Save' });
    expect(button).toBeDisabled();
    expect(button).toHaveAttribute('aria-busy', 'true');
    fireEvent.click(button);
    expect(onClick).not.toHaveBeenCalled();
  });
});

describe('ButtonLink', () => {
  it('renders a router link for `to` and a plain link for `href`', () => {
    render(
      <MemoryRouter>
        <ButtonLink to="/design">Design</ButtonLink>
        <ButtonLink href="/news">News</ButtonLink>
      </MemoryRouter>,
    );
    expect(screen.getByRole('link', { name: 'Design' })).toHaveAttribute('href', '/design');
    expect(screen.getByRole('link', { name: 'News' })).toHaveAttribute('href', '/news');
  });
});

describe('Spinner', () => {
  it('announces loading', () => {
    render(<Spinner />);
    expect(screen.getByRole('status')).toHaveTextContent('Loading');
  });
});

describe('Badge and Alert', () => {
  it('renders badge text', () => {
    render(<Badge tone="red">U12s</Badge>);
    expect(screen.getByText('U12s')).toBeInTheDocument();
  });

  it('uses the darker red for red badge text so it passes AA on the tint', () => {
    render(<Badge tone="red">Youth</Badge>);
    expect(screen.getByText('Youth')).toHaveClass('text-red-hover');
  });

  it('uses role=alert for errors and role=status otherwise', () => {
    render(
      <>
        <Alert tone="error">Broken</Alert>
        <Alert tone="success">Saved</Alert>
      </>,
    );
    expect(screen.getByRole('alert')).toHaveTextContent('Broken');
    expect(screen.getByRole('status')).toHaveTextContent('Saved');
  });
});

describe('CardMedia', () => {
  it('shows the uploaded image', () => {
    render(<CardMedia src="/files/news/1" alt="Cup final" />);
    expect(screen.getByRole('img', { name: 'Cup final' })).toHaveAttribute('src', '/files/news/1');
  });

  it('shows the gradient when there is no image', () => {
    const { container } = render(<CardMedia alt="Cup final" />);
    expect(screen.queryByRole('img')).toBeNull();
    expect(container.querySelector('[data-fallback]')).toBeInTheDocument();
  });

  it('treats an empty src as no image', () => {
    const { container } = render(<CardMedia src="" alt="Cup final" />);
    expect(container.querySelector('[data-fallback]')).toBeInTheDocument();
  });

  it('falls back to the gradient when the image fails to load', () => {
    const { container } = render(<CardMedia src="/files/news/404" alt="Cup final" />);
    fireEvent.error(screen.getByRole('img', { name: 'Cup final' }));
    expect(screen.queryByRole('img')).toBeNull();
    expect(container.querySelector('[data-fallback]')).toBeInTheDocument();
  });

  it('composes inside a card', () => {
    render(
      <Card>
        <CardMedia alt="" />
        <CardBody>First team win the cup</CardBody>
      </Card>,
    );
    expect(screen.getByText('First team win the cup')).toBeInTheDocument();
  });
});

describe('EmptyState and PageHeader', () => {
  it('renders the empty state parts', () => {
    render(
      <EmptyState title="No news yet" message="Check back soon." action={<a href="/">Home</a>} />,
    );
    expect(screen.getByText('No news yet')).toBeInTheDocument();
    expect(screen.getByText('Check back soon.')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Home' })).toBeInTheDocument();
  });

  it('renders the page title as the only h1', () => {
    render(<PageHeader title="Teams" subtitle="All club teams" actions={<button>Add</button>} />);
    expect(screen.getByRole('heading', { level: 1, name: 'Teams' })).toBeInTheDocument();
    expect(screen.getByText('All club teams')).toBeInTheDocument();
  });
});
