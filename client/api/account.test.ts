import { describe, expect, it } from 'vitest';

import { mockFetch } from '../test/mockFetch';
import {
  changePassword,
  checkResetToken,
  removeAccountImage,
  resetPassword,
  uploadAccountImage,
} from './account';

const me = {
  id: 1,
  name: 'Jo',
  email: 'jo@example.test',
  role: 'Manager',
  imageUrl: '/files/user/1',
  permissions: { canEdit: false, canManageGallery: false, canManageUsers: false },
};

describe('account API', () => {
  it('uploads the photo as multipart field "image" with PUT', async () => {
    const fetchMock = mockFetch({ '/api/v1/account/image': { body: me } });
    const file = new File(['x'], 'me.png', { type: 'image/png' });
    await expect(uploadAccountImage(file)).resolves.toEqual(me);
    const [, init] = fetchMock.mock.calls[0];
    expect(init?.method).toBe('PUT');
    expect(init?.body).toBeInstanceOf(FormData);
    expect((init?.body as FormData).get('image')).toBe(file);
  });

  it('removes the photo with DELETE', async () => {
    const fetchMock = mockFetch({ '/api/v1/account/image': { status: 204 } });
    await removeAccountImage();
    expect(fetchMock.mock.calls[0][1]?.method).toBe('DELETE');
  });

  it('changes the password with the exact JSON', async () => {
    const fetchMock = mockFetch({ '/api/v1/auth/password': { status: 204 } });
    await changePassword({ oldPassword: 'a', newPassword: 'b', confirmationPassword: 'b' });
    const [, init] = fetchMock.mock.calls[0];
    expect(init?.method).toBe('POST');
    expect(JSON.parse(String(init?.body))).toEqual({
      oldPassword: 'a',
      newPassword: 'b',
      confirmationPassword: 'b',
    });
  });

  it('checks and uses a reset token', async () => {
    const fetchMock = mockFetch({ '/api/v1/auth/reset/abc-123': { status: 204 } });
    await checkResetToken('abc-123');
    await resetPassword('abc-123', { newPassword: 'x', confirmationPassword: 'x' });
    expect(fetchMock.mock.calls[0][1]?.method).toBe('GET');
    expect(fetchMock.mock.calls[1][1]?.method).toBe('POST');
    expect(JSON.parse(String(fetchMock.mock.calls[1][1]?.body))).toEqual({
      newPassword: 'x',
      confirmationPassword: 'x',
    });
  });
});
