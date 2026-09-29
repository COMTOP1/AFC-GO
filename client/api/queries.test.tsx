import { describe, expect, it } from 'vitest';

import { mockFetch } from '../test/mockFetch';
import { ApiError } from './client';
import { fetchMe } from './queries';

const me = {
  id: 4,
  name: 'Trea Surer',
  email: 'treasurer@example.test',
  role: 'Treasurer',
  permissions: { canEdit: true, canManageGallery: true, canManageUsers: false },
};

describe('fetchMe', () => {
  it('returns the current user', async () => {
    mockFetch({ '/api/v1/auth/me': { body: me } });
    await expect(fetchMe()).resolves.toEqual(me);
  });

  it('treats 401 as anonymous', async () => {
    mockFetch({
      '/api/v1/auth/me': { status: 401, body: { error: { code: 401, message: 'login required' } } },
    });
    await expect(fetchMe()).resolves.toBeNull();
  });

  it('still throws other errors', async () => {
    mockFetch({
      '/api/v1/auth/me': {
        status: 500,
        body: { error: { code: 500, message: 'internal server error' } },
      },
    });
    await expect(fetchMe()).rejects.toBeInstanceOf(ApiError);
  });
});
