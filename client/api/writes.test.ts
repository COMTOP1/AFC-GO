import { describe, expect, it } from 'vitest';

import { emptyImage } from '../lib/images';
import { mockFetch } from '../test/mockFetch';
import { deleteDocument, createDocument } from './documents';
import { createPhoto } from './gallery';
import { createAffiliation } from './home';
import { createNews, deleteNews, updateNews } from './news';
import { setInfo } from './pages';
import { createProgramme, createSeason, deleteSeason, renameSeason } from './programmes';
import { createSponsor } from './sponsors';
import { createTeam, updateTeam } from './teams';
import { updateEvent } from './whatson';

const png = () => new File(['x'], 'a.png', { type: 'image/png' });
const pdf = () => new File(['x'], 'a.pdf', { type: 'application/pdf' });

function body(fetchMock: ReturnType<typeof mockFetch>, i = 0) {
  return fetchMock.mock.calls[i][1]?.body as FormData;
}

describe('write calls', () => {
  it('creates news with title/content/image and no removeImage', async () => {
    const fetchMock = mockFetch({ '/api/v1/news': { status: 201, body: { id: 5 } } });
    const file = png();
    await expect(
      createNews({ title: 'T', content: '<p>c</p>', image: { file, remove: false } }),
    ).resolves.toEqual({ id: 5 });
    const fd = body(fetchMock);
    expect(fetchMock.mock.calls[0][1]?.method).toBe('POST');
    expect([...fd.keys()].sort()).toEqual(['content', 'image', 'title']);
    expect(fd.get('image')).toBe(file);
  });

  it('updates news: no image fields when the image is untouched, removeImage when ticked', async () => {
    const fetchMock = mockFetch({ '/api/v1/news/5': { body: { id: 5 } } });
    await updateNews(5, { title: 'T', content: '', image: emptyImage });
    expect(fetchMock.mock.calls[0][1]?.method).toBe('PATCH');
    expect([...body(fetchMock).keys()].sort()).toEqual(['content', 'title']);
    await updateNews(5, { title: 'T', content: '', image: { file: null, remove: true } });
    expect(body(fetchMock, 1).get('removeImage')).toBe('true');
  });

  it('deletes with DELETE', async () => {
    const fetchMock = mockFetch({ '/api/v1/news/5': { status: 204 } });
    await deleteNews(5);
    expect(fetchMock.mock.calls[0][1]?.method).toBe('DELETE');
  });

  it('updates an event with its date', async () => {
    const fetchMock = mockFetch({ '/api/v1/whatson/3': { body: { id: 3 } } });
    await updateEvent(3, {
      title: 'Quiz',
      content: '',
      dateOfEvent: '2026-11-07',
      image: emptyImage,
    });
    expect(body(fetchMock).get('dateOfEvent')).toBe('2026-11-07');
  });

  it('saves info as JSON with PUT', async () => {
    const fetchMock = mockFetch({ '/api/v1/info': { status: 204 } });
    await setInfo('<p>x</p>');
    expect(fetchMock.mock.calls[0][1]?.method).toBe('PUT');
    expect(JSON.parse(String(fetchMock.mock.calls[0][1]?.body))).toEqual({ content: '<p>x</p>' });
  });

  it('always sends team booleans and every text field', async () => {
    const fetchMock = mockFetch({
      '/api/v1/teams': { status: 201, body: { id: 9 } },
      '/api/v1/teams/9': { body: { id: 9 } },
    });
    const input = {
      name: 'Vets',
      ages: 19,
      description: '',
      league: '',
      division: '',
      leagueTable: '',
      fixtures: '',
      coach: '',
      physio: '',
      isActive: false,
      isYouth: false,
      image: emptyImage,
    };
    await createTeam(input);
    await updateTeam(9, input);
    for (const i of [0, 1]) {
      const fd = body(fetchMock, i);
      expect(fd.get('isActive')).toBe('false');
      expect(fd.get('isYouth')).toBe('false');
      expect(fd.get('ages')).toBe('19');
      expect(fd.get('coach')).toBe('');
      expect(fd.has('image')).toBe(false);
    }
    expect(fetchMock.mock.calls[1][1]?.method).toBe('PATCH');
  });

  it('creates a document and a programme (seasonId only when chosen)', async () => {
    const fetchMock = mockFetch({
      '/api/v1/documents': { status: 201, body: { id: 1 } },
      '/api/v1/programmes': { status: 201, body: { id: 2 } },
      '/api/v1/documents/1': { status: 204 },
    });
    await createDocument({ name: 'Rules', file: pdf() });
    expect(body(fetchMock).get('name')).toBe('Rules');
    await createProgramme({ name: 'vs X', date: '2026-10-03', seasonId: null, file: pdf() });
    expect(body(fetchMock, 1).has('seasonId')).toBe(false);
    await createProgramme({ name: 'vs Y', date: '2026-10-10', seasonId: 2, file: pdf() });
    expect(body(fetchMock, 2).get('seasonId')).toBe('2');
    await deleteDocument(1);
    expect(fetchMock.mock.calls[3][1]?.method).toBe('DELETE');
  });

  it('manages seasons with JSON', async () => {
    const fetchMock = mockFetch({
      '/api/v1/seasons': { status: 201, body: { id: 3, name: '2027-28' } },
      '/api/v1/seasons/3': { body: { id: 3, name: '2027/28' } },
    });
    await createSeason('2027-28');
    await renameSeason(3, '2027/28');
    await deleteSeason(3);
    const calls = fetchMock.mock.calls.map(([, init]) => [init?.method, init?.body]);
    expect(calls).toEqual([
      ['POST', JSON.stringify({ name: '2027-28' })],
      ['PATCH', JSON.stringify({ name: '2027/28' })],
      ['DELETE', undefined],
    ]);
  });

  it('creates a sponsor, an affiliation and a photo', async () => {
    const fetchMock = mockFetch({
      '/api/v1/sponsors': { status: 201, body: { id: 1 } },
      '/api/v1/affiliations': { status: 201, body: { id: 1 } },
      '/api/v1/gallery': { status: 201, body: { id: 1 } },
    });
    await createSponsor({ name: 'Acme', website: '', purpose: 'Kit', team: 'Y', image: png() });
    expect(body(fetchMock).get('team')).toBe('Y');
    await createAffiliation({ name: 'FA', website: 'https://thefa.com', image: png() });
    expect(body(fetchMock, 1).get('website')).toBe('https://thefa.com');
    await createPhoto({ caption: '', image: png() });
    expect(body(fetchMock, 2).get('caption')).toBe('');
  });
});
