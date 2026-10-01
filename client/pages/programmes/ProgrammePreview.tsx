import { useEffect, useRef, useState } from 'react';

type State = 'loading' | 'ready' | 'error';

/**
 * Every page of a programme PDF drawn in a scrolling box, like the old site's
 * programme page. pdf.js is only downloaded when a preview is shown. Give it a key of
 * the url so a different programme starts fresh.
 */
export function ProgrammePreview({ url, name }: { url: string; name: string }) {
  const pagesRef = useRef<HTMLDivElement>(null);
  const [state, setState] = useState<State>('loading');

  useEffect(() => {
    const pages = pagesRef.current;
    if (!pages) {
      return;
    }
    let cancelled = false;
    import('./pdfRenderer')
      .then(({ renderPdf }) => renderPdf(url, pages, () => cancelled))
      .then(
        () => !cancelled && setState('ready'),
        () => !cancelled && setState('error'),
      );
    return () => {
      cancelled = true;
      pages.replaceChildren();
    };
  }, [url]);

  return (
    <div
      role="group"
      aria-label={`Preview of ${name}`}
      className="max-h-[80vh] overflow-y-auto rounded-lg border border-line bg-surface p-2.5"
    >
      {state === 'loading' && (
        <p role="status" className="p-5 text-center text-muted">
          Loading programme…
        </p>
      )}
      {state === 'error' && (
        <p className="p-5 text-center">
          We couldn&apos;t preview this programme.{' '}
          <a href={url} download className="font-semibold text-red underline">
            Download it
          </a>
        </p>
      )}
      {/* pdf.js draws the pages here; React never renders into this box. */}
      <div
        ref={pagesRef}
        className="space-y-2.5 [&>canvas]:mx-auto [&>canvas]:block [&>canvas]:h-auto [&>canvas]:max-w-full [&>canvas]:shadow-md"
      />
    </div>
  );
}
