type FieldValue = string | number | boolean | File | null | undefined;

/**
 * A multipart body in the API's conventions: strings (including empty, which
 * clears a field on PATCH) and numbers as-is, booleans as "true"/"false",
 * files as files; null/undefined fields are left out entirely.
 */
export function formData(fields: Record<string, FieldValue>): FormData {
  const fd = new FormData();
  for (const [name, value] of Object.entries(fields)) {
    if (value === null || value === undefined) {
      continue;
    }
    if (value instanceof File) {
      fd.append(name, value);
    } else if (typeof value === 'boolean') {
      fd.append(name, value ? 'true' : 'false');
    } else {
      fd.append(name, String(value));
    }
  }
  return fd;
}

const dateParts = new Intl.DateTimeFormat('en-GB', {
  timeZone: 'Europe/London',
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
});

/** An API date as an <input type="date"> value (YYYY-MM-DD, UK time), or '' if invalid. */
export function toDateInput(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) {
    return '';
  }
  const p: Record<string, string> = {};
  for (const part of dateParts.formatToParts(d)) {
    p[part.type] = part.value;
  }
  return `${p.year}-${p.month}-${p.day}`;
}

export function formatSize(bytes: number): string {
  if (bytes < 1024) {
    return `${bytes} B`;
  }
  if (bytes < 1024 * 1024) {
    return `${Math.round(bytes / 1024)} KB`;
  }
  return `${(bytes / (1024 * 1024)).toFixed(1).replace(/\.0$/, '')} MB`;
}
