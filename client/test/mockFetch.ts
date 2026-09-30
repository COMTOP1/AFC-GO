import { vi } from 'vitest';

export interface MockResponse {
  status?: number;
  body?: unknown; // objects are JSON-encoded; strings are sent as-is
  contentType?: string;
}

export type MockRoute =
  MockResponse | ((init?: RequestInit) => MockResponse | Promise<MockResponse>);

/**
 * Replaces global fetch with a stub keyed by path (e.g. '/api/v1/site').
 * A route may be a function, evaluated per request with the request's init (it
 * can throw or return a promise). Unknown paths fail the test loudly with a 599.
 */
export function mockFetch(routes: Record<string, MockRoute>) {
  const fn = vi.fn(async (...args: [RequestInfo | URL, RequestInit?]) => {
    const [input, init] = args;
    const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url;
    const path = url.replace(/^https?:\/\/[^/]+/, '');
    const entry = routes[path];
    if (!entry) {
      return new Response(`no mock for ${path}`, { status: 599 });
    }
    const route = typeof entry === 'function' ? await entry(init) : entry;
    const status = route.status ?? 200;
    if (status === 204) {
      return new Response(null, { status });
    }
    const isText = typeof route.body === 'string';
    return new Response(
      route.body === undefined ? '' : isText ? (route.body as string) : JSON.stringify(route.body),
      {
        status,
        headers: {
          'Content-Type': route.contentType ?? (isText ? 'text/html' : 'application/json'),
        },
      },
    );
  });
  vi.stubGlobal('fetch', fn);
  return fn;
}
