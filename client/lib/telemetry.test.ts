import { afterEach, describe, expect, it, vi } from 'vitest';

import { newTraceparent, reportEvent } from './telemetry';

function stubSendBeacon() {
  const fn = vi.fn<Navigator['sendBeacon']>(() => true);
  Object.defineProperty(navigator, 'sendBeacon', { value: fn, configurable: true });
  return fn;
}

afterEach(() => {
  Reflect.deleteProperty(navigator, 'sendBeacon');
});

describe('newTraceparent', () => {
  it('returns a valid W3C traceparent string', () => {
    expect(newTraceparent()).toMatch(/^00-[0-9a-f]{32}-[0-9a-f]{16}-01$/);
  });

  it('returns a different trace id each time', () => {
    expect(newTraceparent()).not.toBe(newTraceparent());
  });
});

describe('reportEvent', () => {
  it('sends the event via navigator.sendBeacon to /api/v1/telemetry', async () => {
    const beacon = stubSendBeacon();
    reportEvent({ type: 'api', name: 'GET /site', startTime: 1000, endTime: 1200 });

    expect(beacon).toHaveBeenCalledTimes(1);
    const [url, blob] = beacon.mock.calls[0] as [string, Blob];
    expect(url).toBe('/api/v1/telemetry');
    expect(blob.type).toBe('application/json');
    expect(JSON.parse(await blob.text())).toEqual({
      type: 'api',
      name: 'GET /site',
      startTime: 1000,
      endTime: 1200,
    });
  });

  it('does nothing when sendBeacon is unavailable (e.g. this test environment by default)', () => {
    expect(() => reportEvent({ type: 'error', name: 'unhandled', startTime: 1000 })).not.toThrow();
  });

  it('never throws when sendBeacon itself throws', () => {
    Object.defineProperty(navigator, 'sendBeacon', {
      value: () => {
        throw new Error('blocked');
      },
      configurable: true,
    });
    expect(() => reportEvent({ type: 'error', name: 'unhandled', startTime: 1000 })).not.toThrow();
  });
});
