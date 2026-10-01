import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterEach, beforeEach, vi } from 'vitest';

import { installDialogPolyfill } from './dialog';
import { installMatchMedia } from './matchMedia';

installDialogPolyfill();

// jsdom can't run pdf.js (no canvas, no worker); tests use a stand-in renderer.
vi.mock('../pages/programmes/pdfRenderer', () => ({
  renderPdf: vi.fn(async () => undefined),
}));

// ProseMirror measures ranges when scrolling the selection into view; jsdom has no layout.
Range.prototype.getClientRects = () =>
  ({
    length: 0,
    item: () => null,
    [Symbol.iterator]: [][Symbol.iterator],
  }) as unknown as DOMRectList;
Range.prototype.getBoundingClientRect = () => new DOMRect();

beforeEach(() => {
  installMatchMedia(false);
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
  localStorage.clear();
  delete document.documentElement.dataset.theme;
});
