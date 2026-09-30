import { describe, expect, it } from 'vitest';

import { emptyImage } from '../lib/images';
import { mockFetch } from '../test/mockFetch';
import { setDisplayEmail } from './pages';
import { createPlayer, deletePlayer, updatePlayer } from './players';
import { createUser, deleteUser, resetUserPassword, updateUser } from './users';

function body(fetchMock: ReturnType<typeof mockFetch>, i = 0) {
  return fetchMock.mock.calls[i][1]?.body as FormData;
}

const playerInput = {
  name: 'Sam',
  teamId: 7,
  dateOfBirth: '1998-04-02',
  position: '',
  isCaptain: false,
  image: emptyImage,
};

describe('player writes', () => {
  it('creates and updates with every field, isCaptain explicit, no image fields when untouched', async () => {
    const fetchMock = mockFetch({
      '/api/v1/players': { status: 201, body: { id: 30 } },
      '/api/v1/players/30': { body: { id: 30 } },
    });
    await createPlayer(playerInput);
    await updatePlayer(30, playerInput);
    for (const i of [0, 1]) {
      const fd = body(fetchMock, i);
      expect([...fd.keys()].sort()).toEqual([
        'dateOfBirth',
        'isCaptain',
        'name',
        'position',
        'teamId',
      ]);
      expect(fd.get('isCaptain')).toBe('false');
      expect(fd.get('teamId')).toBe('7');
    }
    expect(fetchMock.mock.calls[1][1]?.method).toBe('PATCH');
  });

  it('sends removeImage only on update', async () => {
    const fetchMock = mockFetch({ '/api/v1/players/30': { body: { id: 30 } } });
    await updatePlayer(30, { ...playerInput, image: { file: null, remove: true } });
    expect(body(fetchMock).get('removeImage')).toBe('true');
  });

  it('deletes', async () => {
    const fetchMock = mockFetch({ '/api/v1/players/30': { status: 204 } });
    await deletePlayer(30);
    expect(fetchMock.mock.calls[0][1]?.method).toBe('DELETE');
  });
});

describe('user writes', () => {
  it('sends teamId only when given and role only when given', async () => {
    const fetchMock = mockFetch({
      '/api/v1/users': { status: 201, body: { user: { id: 3 }, emailSent: true } },
      '/api/v1/users/3': { body: { id: 3 } },
    });
    await createUser({
      name: 'N',
      email: 'n@x.test',
      phone: '',
      role: 'manager',
      teamId: 7,
      image: emptyImage,
    });
    expect(body(fetchMock).get('role')).toBe('manager');
    expect(body(fetchMock).get('teamId')).toBe('7');
    await updateUser(3, { name: 'N', email: 'n@x.test', phone: '', image: emptyImage });
    const fd = body(fetchMock, 1);
    expect([...fd.keys()].sort()).toEqual(['email', 'name', 'phone']);
    expect(fd.get('phone')).toBe('');
    expect(fetchMock.mock.calls[1][1]?.method).toBe('PATCH');
  });

  it('resets with POST and returns the result; deletes with DELETE', async () => {
    const fetchMock = mockFetch({
      '/api/v1/users/3/reset': { body: { emailSent: false, resetUrl: 'https://x/reset/t' } },
      '/api/v1/users/3': { status: 204 },
    });
    await expect(resetUserPassword(3)).resolves.toEqual({
      emailSent: false,
      resetUrl: 'https://x/reset/t',
    });
    expect(fetchMock.mock.calls[0][1]?.method).toBe('POST');
    await deleteUser(3);
    expect(fetchMock.mock.calls[1][1]?.method).toBe('DELETE');
  });
});

describe('display email', () => {
  it('PUTs JSON, including an empty email to clear it', async () => {
    const fetchMock = mockFetch({ '/api/v1/settings/display-email': { body: { email: '' } } });
    await setDisplayEmail('');
    expect(fetchMock.mock.calls[0][1]?.method).toBe('PUT');
    expect(JSON.parse(String(fetchMock.mock.calls[0][1]?.body))).toEqual({ email: '' });
  });
});
