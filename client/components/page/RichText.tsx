import { clsx } from 'clsx';
import { useMemo } from 'react';

import { cleanHtml } from '../../lib/sanitise';

export function RichText({ html, className }: { html: string; className?: string }) {
  const clean = useMemo(() => cleanHtml(html), [html]);
  return <div className={clsx('prose', className)} dangerouslySetInnerHTML={{ __html: clean }} />;
}
