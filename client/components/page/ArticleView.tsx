import type { ReactNode } from 'react';
import { Link } from 'react-router';

import { ButtonLink } from '../ui/ButtonLink';
import { PageHeader } from '../ui/PageHeader';
import { EditorLink } from './EditorLink';
import { ImageWithFallback } from './ImageWithFallback';
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
  editorHref: string;
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
  editorHref,
}: ArticleViewProps) {
  usePageTitle(title);
  return (
    <article className="mx-auto max-w-3xl">
      {/* Shown in full at its own shape (list cards crop; the article page doesn't).
          Very tall images are capped and scaled down rather than cut off. */}
      <div className="mb-6 overflow-hidden rounded-lg border border-line bg-surface">
        <ImageWithFallback
          src={imageUrl}
          alt=""
          className="mx-auto block h-auto max-h-[70vh] w-auto max-w-full object-contain"
          fallback={
            <div
              aria-hidden="true"
              data-fallback=""
              className="w-full bg-linear-135 from-blue to-red"
              style={{ aspectRatio: '21 / 9' }}
            />
          }
        />
      </div>
      <nav aria-label="Breadcrumb" className="mb-2 text-sm text-muted">
        <Link to={sectionHref} className="font-semibold text-red">
          {section}
        </Link>{' '}
        / {crumb}
      </nav>
      <PageHeader
        title={title}
        subtitle={subtitle}
        actions={<EditorLink legacyHref={editorHref} />}
      />
      <RichText html={html} />
      <div className="mt-8">
        <ButtonLink to={sectionHref} variant="secondary">
          {backLabel}
        </ButtonLink>
      </div>
    </article>
  );
}
