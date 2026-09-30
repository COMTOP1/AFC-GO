const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

const partsFormat = new Intl.DateTimeFormat('en-GB', {
  timeZone: 'Europe/London',
  weekday: 'short',
  day: 'numeric',
  month: 'numeric',
  year: 'numeric',
  hour: 'numeric',
  minute: '2-digit',
  hourCycle: 'h12',
});

/**
 * UK-style dates in the club's time zone: 'date' → "28 Sep 2026",
 * 'dayDate' → "Fri 16 Oct 2026",
 * 'dateTime' → "Fri 16 Oct 2026, 7pm" (minutes only when not :00).
 * Month names are fixed so ICU's "Sept" never appears.
 */
export function formatDate(iso: string, style: 'date' | 'dateTime' | 'dayDate' = 'date'): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) {
    return '';
  }
  const p: Record<string, string> = {};
  for (const part of partsFormat.formatToParts(d)) {
    p[part.type] = part.value;
  }
  const date = `${Number(p.day)} ${MONTHS[Number(p.month) - 1]} ${p.year}`;
  if (style === 'date') {
    return date;
  }
  if (style === 'dayDate') {
    return `${p.weekday} ${date}`;
  }
  const minutes = p.minute === '00' ? '' : `:${p.minute}`;
  return `${p.weekday} ${date}, ${p.hour}${minutes}${(p.dayPeriod ?? '').toLowerCase()}`;
}
