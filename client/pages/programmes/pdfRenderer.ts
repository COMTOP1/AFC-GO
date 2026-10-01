import { getDocument, GlobalWorkerOptions } from 'pdfjs-dist';
import workerUrl from 'pdfjs-dist/build/pdf.worker.min.mjs?url';

GlobalWorkerOptions.workerSrc = workerUrl;

/**
 * Draws every page of the PDF at url into container, one canvas per page, at
 * the container's width (sharp on high-density screens). Stops early once
 * isCancelled() is true, e.g. when the preview unmounts.
 */
export async function renderPdf(
  url: string,
  container: HTMLElement,
  isCancelled: () => boolean,
): Promise<void> {
  const task = getDocument({ url });
  try {
    const pdf = await task.promise;
    const ratio = window.devicePixelRatio || 1;
    for (let n = 1; n <= pdf.numPages; n++) {
      if (isCancelled()) {
        return;
      }
      const page = await pdf.getPage(n);
      const natural = page.getViewport({ scale: 1 });
      const width = container.clientWidth || natural.width;
      const viewport = page.getViewport({ scale: width / natural.width });
      const canvas = document.createElement('canvas');
      canvas.width = Math.floor(viewport.width * ratio);
      canvas.height = Math.floor(viewport.height * ratio);
      canvas.style.width = `${Math.floor(viewport.width)}px`;
      canvas.setAttribute('role', 'img');
      canvas.setAttribute('aria-label', `Page ${n} of ${pdf.numPages}`);
      container.append(canvas);
      await page.render({
        canvas,
        viewport,
        transform: ratio === 1 ? undefined : [ratio, 0, 0, ratio, 0, 0],
      }).promise;
    }
  } finally {
    void task.destroy();
  }
}
