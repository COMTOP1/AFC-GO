/** True when every word of q appears in text, ignoring case and extra spaces. */
export function matchesQuery(text: string, q: string): boolean {
  const haystack = text.toLowerCase();
  return q
    .toLowerCase()
    .split(/\s+/)
    .filter(Boolean)
    .every((word) => haystack.includes(word));
}
