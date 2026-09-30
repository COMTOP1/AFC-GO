import { lazy, Suspense } from 'react';

import { Skeleton } from '../ui/Skeleton';
import type { RichTextEditorProps } from './richText';

// Tiptap/ProseMirror are only downloaded when an edit form opens.
const RichTextEditorImpl = lazy(() => import('./RichTextEditorImpl'));

export function RichTextEditor(props: RichTextEditorProps) {
  return (
    <Suspense fallback={<Skeleton className="h-48" />}>
      <RichTextEditorImpl {...props} />
    </Suspense>
  );
}
