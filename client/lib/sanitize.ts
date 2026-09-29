import DOMPurify from 'dompurify';

// Off-site links in stored content open in a new tab without leaking the opener.
DOMPurify.addHook('afterSanitizeAttributes', (node) => {
  if (node.tagName !== 'A' || !node.hasAttribute('href')) {
    return;
  }
  try {
    const url = new URL(node.getAttribute('href') ?? '', window.location.href);
    if (url.origin !== window.location.origin) {
      node.setAttribute('target', '_blank');
      node.setAttribute('rel', 'noopener noreferrer');
    }
  } catch {
    // Unparseable href: leave it as DOMPurify cleaned it.
  }
});

/**
 * Cleans stored article/info HTML before it is rendered. The server only
 * sanitises on save, so older rows may contain anything.
 */
export function cleanHtml(html: string): string {
  return DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
    FORBID_TAGS: ['style', 'form', 'iframe'],
  });
}

const BLOCKS = 'p,div,br,li,h1,h2,h3,h4,h5,h6,blockquote,tr';

/** Cleaned HTML as plain text, cut at a word boundary to at most maxChars (plus "…"). */
export function plainText(html: string, maxChars: number): string {
  const doc = new DOMParser().parseFromString(cleanHtml(html), 'text/html');
  doc.body.querySelectorAll(BLOCKS).forEach((el) => el.after(' '));
  const text = (doc.body.textContent ?? '').replace(/\s+/g, ' ').trim();
  if (text.length <= maxChars) {
    return text;
  }
  const cut = text.slice(0, maxChars);
  const lastSpace = cut.lastIndexOf(' ');
  return `${(lastSpace > 0 ? cut.slice(0, lastSpace) : cut).replace(/[\s,.;:]+$/, '')}…`;
}
