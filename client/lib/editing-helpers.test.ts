import { describe, expect, it } from 'vitest';

import { ApiError } from '../api/client';
import { formatSize, formData, toDateInput } from './editForm';
import { IMAGE_TYPE_MESSAGE, isAcceptedImage } from './images';
import { describeSaveError } from './saveErrors';
import { SPONSOR_TEAM_CHOICES, sponsorTeamLabel } from './sponsorTeam';

describe('formData', () => {
  it('sends strings (even empty), numbers, booleans as true/false and files; skips null/undefined', () => {
    const file = new File(['x'], 'a.png', { type: 'image/png' });
    const fd = formData({
      title: 'Hi',
      content: '',
      ages: 12,
      isActive: false,
      isYouth: true,
      image: file,
      removeImage: undefined,
      seasonId: null,
    });
    expect(fd.get('title')).toBe('Hi');
    expect(fd.get('content')).toBe('');
    expect(fd.get('ages')).toBe('12');
    expect(fd.get('isActive')).toBe('false');
    expect(fd.get('isYouth')).toBe('true');
    expect(fd.get('image')).toBe(file);
    expect(fd.has('removeImage')).toBe(false);
    expect(fd.has('seasonId')).toBe(false);
  });
});

describe('dates and sizes', () => {
  it('turns an API date into a date-input value in UK time', () => {
    expect(toDateInput('2026-10-16T00:00:00Z')).toBe('2026-10-16');
    expect(toDateInput('2026-06-30T23:30:00Z')).toBe('2026-07-01');
    expect(toDateInput('nonsense')).toBe('');
  });

  it('formats file sizes', () => {
    expect(formatSize(512)).toBe('512 B');
    expect(formatSize(2048)).toBe('2 KB');
    expect(formatSize(1_572_864)).toBe('1.5 MB');
  });
});

describe('images', () => {
  it('accepts only the server image types', () => {
    expect(isAcceptedImage(new File([''], 'a.png', { type: 'image/png' }))).toBe(true);
    expect(isAcceptedImage(new File([''], 'a.html', { type: 'text/html' }))).toBe(false);
    expect(IMAGE_TYPE_MESSAGE).toBe(
      'Choose an image file (JPEG, PNG, GIF, WebP, AVIF, APNG or SVG).',
    );
  });
});

describe('sponsor team', () => {
  const teams = [{ id: 7, name: 'First Team' }];

  it('labels the codes and team ids', () => {
    expect(sponsorTeamLabel('A', teams)).toBe('All teams');
    expect(sponsorTeamLabel('O', teams)).toBe('Adult teams');
    expect(sponsorTeamLabel('Y', teams)).toBe('Youth teams');
    expect(sponsorTeamLabel('7', teams)).toBe('First Team');
    expect(sponsorTeamLabel('99', teams)).toBeUndefined();
    expect(sponsorTeamLabel('', teams)).toBeUndefined();
    expect(sponsorTeamLabel(undefined, teams)).toBeUndefined();
  });

  it('offers None, All, Adult and Youth as the fixed choices', () => {
    expect(SPONSOR_TEAM_CHOICES.map((c) => c.value)).toEqual(['', 'A', 'O', 'Y']);
  });
});

describe('describeSaveError', () => {
  it('explains too-large uploads and passes other messages through', () => {
    expect(describeSaveError(new ApiError(413, 'Request Entity Too Large'))).toBe(
      'That file is too large (15 MB maximum).',
    );
    expect(describeSaveError(new ApiError(500, 'boom'))).toBe('boom');
    expect(describeSaveError('?')).toBe('Something went wrong. Please try again.');
  });
});
