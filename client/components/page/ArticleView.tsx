import type { ReactNode } from 'react';
import { Link } from 'react-router';

import { ButtonLink } from '../ui/ButtonLink';
import { PageHeader } from '../ui/PageHeader';
import { FullImage } from './FullImage';
import { RichText } from './RichText';
import { usePageTitle } from './usePageTitle';

export interface ArticleViewProps {
  section: string;
  sectionHref: string;
  /** Text after the section in the breadcrumb, e.g. the date. */
  crumb: string;
  title: string;
  imageUrl?: string;
  html: string;
  subtitle?: ReactNode;
  backLabel: string;
  /** Editor controls (Edit/Delete) for the page header. */
  actions?: ReactNode;
}

/** A news article or event: image, breadcrumb, title, cleaned body, back link. */
export function ArticleView({
  section,
  sectionHref,
  crumb,
  title,
  imageUrl,
  html,
  subtitle,
  backLabel,
  actions,
}: ArticleViewProps) {
  usePageTitle(title);
  return (
    <article className="mx-auto max-w-3xl">
      {/* Shown in full at its own shape (list cards crop; the article page doesn't). */}
      <FullImage src={imageUrl} alt="" className="mb-6" />
      <nav aria-label="Breadcrumb" className="mb-2 text-sm text-muted">
        <Link to={sectionHref} className="font-semibold text-red">
          {section}
        </Link>{' '}
        / {crumb}
      </nav>
      <PageHeader title={title} subtitle={subtitle} actions={actions} />
      <RichText html={html} />
      <div className="mt-8">
        <ButtonLink to={sectionHref} variant="secondary">
          {backLabel}
        </ButtonLink>
      </div>
    </article>
  );
}
