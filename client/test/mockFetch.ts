import { vi } from 'vitest';

export interface MockResponse {
  status?: number;
  body?: unknown; // objects are JSON-encoded; strings are sent as-is
  contentType?: string;
}

/**
 * Replaces global fetch with a stub keyed by path (e.g. '/api/v1/site').
 * Unknown paths fail the test loudly with a 599.
 */
export function mockFetch(routes: Record<string, MockResponse | (() => never)>) {
  const fn = vi.fn(async (input: RequestInfo | URL) => {
    const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url;
    const path = url.replace(/^https?:\/\/[^/]+/, '');
    const route = routes[path];
    if (!route) {
      return new Response(`no mock for ${path}`, { status: 599 });
    }
    if (typeof route === 'function') {
      return route();
    }
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
