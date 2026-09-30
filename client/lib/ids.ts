/** A route :id as a positive integer, or null when it isn't one. */
export function parseId(raw: string | undefined): number | null {
  if (!raw || !/^[1-9]\d*$/.test(raw)) {
    return null;
  }
  const id = Number(raw);
  return Number.isSafeInteger(id) ? id : null;
}
