import type { ErrorEnvelope } from './types';

export const API_BASE = '/api/v1';

/** A failed API call. status 0 means the server could not be reached. */
export class ApiError extends Error {
  readonly status: number;
  readonly fields: Record<string, string>;

  constructor(status: number, message: string, fields: Record<string, string> = {}) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.fields = fields;
  }
}

export interface ApiRequest {
  method?: string;
  json?: unknown;
  form?: FormData;
  signal?: AbortSignal;
}

function isEnvelope(data: unknown): data is ErrorEnvelope {
  if (typeof data !== 'object' || data === null || !('error' in data)) {
    return false;
  }
  const error = (data as { error: unknown }).error;
  return (
    typeof error === 'object' &&
    error !== null &&
    typeof (error as { message?: unknown }).message === 'string'
  );
}

/**
 * Calls the JSON API. CSRF needs no handling: browsers send
 * Sec-Fetch-Site: same-origin, which the server accepts.
 */
export async function apiFetch<T>(path: string, req: ApiRequest = {}): Promise<T> {
  const headers: Record<string, string> = { Accept: 'application/json' };
  let body: BodyInit | undefined;
  if (req.json !== undefined) {
    headers['Content-Type'] = 'application/json';
    body = JSON.stringify(req.json);
  } else if (req.form) {
    body = req.form; // the browser sets the multipart boundary
  }

  let res: Response;
  try {
    res = await fetch(API_BASE + path, {
      method: req.method ?? (body === undefined ? 'GET' : 'POST'),
      headers,
      body,
      credentials: 'same-origin',
      signal: req.signal,
    });
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      throw err;
    }
    throw new ApiError(0, 'Could not reach the server. Check your connection and try again.');
  }

  if (res.status === 204) {
    return undefined as T;
  }

  const text = await res.text();
  let data: unknown;
  try {
    data = text === '' ? undefined : JSON.parse(text);
  } catch {
    data = undefined;
  }

  if (!res.ok) {
    if (isEnvelope(data)) {
      throw new ApiError(res.status, data.error.message, data.error.fields ?? {});
    }
    throw new ApiError(res.status, res.statusText || `Request failed with status ${res.status}`);
  }
  if (data === undefined) {
    throw new ApiError(res.status, 'The server sent an unexpected response.');
  }
  return data as T;
}
