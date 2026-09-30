/**
 * A link address as typed by an editor, made safe: http(s) and mailto only;
 * a bare domain ("league.example/table") gets https://. Anything else is null.
 */
export function normalizeLink(input: string): string | null {
  const value = input.trim();
  if (!value) {
    return null;
  }
  const hasScheme = /^[a-z][a-z0-9+.-]*:/i.test(value);
  const candidate = hasScheme ? value : `https://${value}`;
  try {
    const url = new URL(candidate);
    if (url.protocol === 'mailto:') {
      return url.href;
    }
    if ((url.protocol === 'https:' || url.protocol === 'http:') && url.hostname.includes('.')) {
      return hasScheme ? url.href : candidate;
    }
  } catch {
    return null;
  }
  return null;
}
