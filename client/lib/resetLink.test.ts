import { describe, expect, it } from 'vitest';

import { appResetPath, RESET_TOKEN } from './resetLink';

describe('appResetPath', () => {
  it('maps the server reset URL to the in-app route', () => {
    expect(appResetPath('/reset/5f0b6c1e-1a2b-4c3d-8e9f-000000000001')).toBe(
      '/reset/5f0b6c1e-1a2b-4c3d-8e9f-000000000001',
    );
  });

  it('rejects anything else', () => {
    for (const url of [
      '/reset/',
      '/reset/a/b',
      'https://x.example/reset/abc',
      '/other/abc',
      '/reset/<x>',
    ]) {
      expect(appResetPath(url)).toBeNull();
    }
  });

  it('exposes the token pattern', () => {
    expect(RESET_TOKEN.test('abc-123')).toBe(true);
    expect(RESET_TOKEN.test('abc 123')).toBe(false);
  });
});
