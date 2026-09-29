import { describe, expect, it, vi } from 'vitest';

import { mockFetch } from '../test/mockFetch';
import { ApiError, apiFetch } from './client';
import { shouldRetry } from './queryClient';

describe('apiFetch', () => {
  it('returns parsed JSON and calls /api/v1 same-origin', async () => {
    const fetchMock = mockFetch({ '/api/v1/site': { body: { year: 2026 } } });
    await expect(apiFetch<{ year: number }>('/site')).resolves.toEqual({ year: 2026 });
    const [, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
    expect(init.credentials).toBe('same-origin');
    expect(init.method).toBe('GET');
  });

  it('sends JSON bodies with a content type and defaults to POST', async () => {
    const fetchMock = mockFetch({ '/api/v1/auth/login': { body: { ok: true } } });
    await apiFetch('/auth/login', { json: { email: 'a@b.test' } });
    const [, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
    expect(init.method).toBe('POST');
    expect((init.headers as Record<string, string>)['Content-Type']).toBe('application/json');
    expect(init.body).toBe('{"email":"a@b.test"}');
  });

  it('sends FormData without setting a content type', async () => {
    const fetchMock = mockFetch({ '/api/v1/news': { status: 201, body: { id: 1 } } });
    const form = new FormData();
    form.set('title', 'x');
    await apiFetch('/news', { form });
    const [, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
    expect(init.body).toBe(form);
    expect((init.headers as Record<string, string>)['Content-Type']).toBeUndefined();
  });

  it('returns undefined for 204', async () => {
    mockFetch({ '/api/v1/news/1': { status: 204 } });
    await expect(apiFetch('/news/1', { method: 'DELETE' })).resolves.toBeUndefined();
  });

  it('turns the error envelope into an ApiError with fields', async () => {
    mockFetch({
      '/api/v1/news': {
        status: 422,
        body: {
          error: {
            code: 422,
            message: 'validation failed',
            fields: { title: 'title is required' },
          },
        },
      },
    });
    const err = await apiFetch('/news', { json: {} }).catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect(err).toMatchObject({
      status: 422,
      message: 'validation failed',
      fields: { title: 'title is required' },
    });
  });

  it('non-JSON error body becomes an ApiError with the status text', async () => {
    mockFetch({ '/api/v1/site': { status: 502, body: '<html>Bad gateway</html>' } });
    const err = await apiFetch('/site').catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect((err as ApiError).status).toBe(502);
    expect((err as ApiError).message).not.toContain('<html>');
    expect((err as ApiError).fields).toEqual({});
  });

  it('empty 200 body is an error, not undefined data', async () => {
    mockFetch({ '/api/v1/site': { status: 200, body: '', contentType: 'application/json' } });
    const err = await apiFetch('/site').catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect((err as ApiError).message).toBe('The server sent an unexpected response.');
  });

  it('network failure is ApiError status 0', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => Promise.reject(new TypeError('Failed to fetch'))),
    );
    const err = await apiFetch('/site').catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect((err as ApiError).status).toBe(0);
  });

  it('lets aborts propagate unchanged', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => Promise.reject(new DOMException('aborted', 'AbortError'))),
    );
    const err = await apiFetch('/site').catch((e: unknown) => e);
    expect(err).toBeInstanceOf(DOMException);
  });
});

describe('shouldRetry', () => {
  it('never retries 4xx ApiErrors', () => {
    expect(shouldRetry(0, new ApiError(401, 'login required'))).toBe(false);
    expect(shouldRetry(0, new ApiError(404, 'not found'))).toBe(false);
  });
  it('retries other errors up to twice', () => {
    expect(shouldRetry(0, new ApiError(503, 'down'))).toBe(true);
    expect(shouldRetry(1, new ApiError(0, 'offline'))).toBe(true);
    expect(shouldRetry(2, new ApiError(500, 'boom'))).toBe(false);
    expect(shouldRetry(0, new Error('x'))).toBe(true);
  });
});
