import { describe, expect, it } from 'vitest';

import { formatDate } from './format';
import { parseId } from './ids';
import { matchesQuery } from './text';

describe('formatDate', () => {
  it('formats a date in UK style', () => {
    expect(formatDate('2026-09-28T10:00:00Z')).toBe('28 Sep 2026');
  });

  it('formats a date and time in UK summer time, dropping :00', () => {
    expect(formatDate('2026-10-16T18:00:00Z', 'dateTime')).toBe('Fri 16 Oct 2026, 7pm');
  });

  it('keeps minutes and uses GMT in winter', () => {
    expect(formatDate('2026-12-19T19:30:00Z', 'dateTime')).toBe('Sat 19 Dec 2026, 7:30pm');
  });

  it('shows midnight as 12am', () => {
    expect(formatDate('2026-12-19T00:00:00Z', 'dateTime')).toBe('Sat 19 Dec 2026, 12am');
  });

  it('does not pad single-digit days', () => {
    expect(formatDate('2026-09-05T13:00:00Z')).toBe('5 Sep 2026');
    expect(formatDate('2026-10-02T18:00:00Z', 'dateTime')).toBe('Fri 2 Oct 2026, 7pm');
  });

  it('returns an empty string for an invalid date', () => {
    expect(formatDate('not a date')).toBe('');
  });
});

describe('matchesQuery', () => {
  it('matches every word, ignoring case and spacing', () => {
    expect(matchesQuery('Club constitution', '  CLUB  const')).toBe(true);
    expect(matchesQuery('Club constitution', 'club policy')).toBe(false);
  });

  it('matches everything for an empty query', () => {
    expect(matchesQuery('Anything', '')).toBe(true);
    expect(matchesQuery('Anything', '   ')).toBe(true);
  });
});

describe('parseId', () => {
  it('accepts positive integers', () => {
    expect(parseId('42')).toBe(42);
  });

  it('rejects everything else', () => {
    for (const raw of [undefined, '', 'abc', '0', '-1', '1.5', '12abc', '99999999999999999999']) {
      expect(parseId(raw)).toBeNull();
    }
  });
});
