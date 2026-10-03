/**
 * Best-effort client telemetry: API call outcomes and unhandled errors are
 * posted to the Go backend, which re-emits them as OpenTelemetry spans
 * alongside server-side traces (see server/internal/clienttelemetry).
 *
 * Never throws and never awaits the network: a telemetry failure must not
 * affect the page.
 */

const ENDPOINT = '/api/v1/telemetry';

export interface TelemetryEvent {
  /** "api" for a completed fetch call, "error" for an unhandled exception. */
  type: 'api' | 'error';
  /** What happened, e.g. "GET /site" or "unhandled". */
  name: string;
  /** Unix ms. */
  startTime: number;
  /** Unix ms; omitted for instantaneous events. */
  endTime?: number;
  /** The traceparent sent with the matching API request, if any. */
  traceparent?: string;
  attributes?: Record<string, string>;
  message?: string;
  stack?: string;
}

function randomHex(bytes: number): string {
  const arr = new Uint8Array(bytes);
  crypto.getRandomValues(arr);
  return Array.from(arr, (b) => b.toString(16).padStart(2, '0')).join('');
}

/** Builds a new W3C traceparent header value for one API call. */
export function newTraceparent(): string {
  return `00-${randomHex(16)}-${randomHex(8)}-01`;
}

/**
 * Reports a telemetry event. Fire-and-forget: never throws, never blocks,
 * and survives the page navigating away right after (e.g. a nav link
 * clicked straight after a failed save).
 *
 * Uses navigator.sendBeacon rather than fetch: it's the one transport built
 * for exactly this (queued by the browser, not cancelled on unload), and it
 * keeps these best-effort pings out of the app's own request/response flow
 * entirely - including in tests, which stub fetch but not sendBeacon.
 */
export function reportEvent(event: TelemetryEvent): void {
  if (typeof navigator === 'undefined' || typeof navigator.sendBeacon !== 'function') {
    return;
  }
  try {
    navigator.sendBeacon(ENDPOINT, new Blob([JSON.stringify(event)], { type: 'application/json' }));
  } catch {
    // best-effort; dropping the event is fine
  }
}
