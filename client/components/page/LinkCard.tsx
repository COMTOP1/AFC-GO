import type { ReactNode } from 'react';
import { Link } from 'react-router';

import { Card, CardBody, CardMedia } from '../ui/Card';

export interface LinkCardProps {
  to: string;
  imageUrl?: string;
  title: string;
  meta?: ReactNode;
  children?: ReactNode;
}

/** An image card that links to a detail page; the gradient stands in for a missing image. */
export function LinkCard({ to, imageUrl, title, meta, children }: LinkCardProps) {
  return (
    <Card className="h-full transition-shadow hover:shadow-md">
      <Link to={to} className="block h-full">
        <CardMedia src={imageUrl} alt="" />
        <CardBody className="space-y-1">
          <h2 className="font-display text-xl leading-tight font-extrabold uppercase">{title}</h2>
          {meta && <div className="text-sm text-muted">{meta}</div>}
          {children}
        </CardBody>
      </Link>
    </Card>
  );
}
