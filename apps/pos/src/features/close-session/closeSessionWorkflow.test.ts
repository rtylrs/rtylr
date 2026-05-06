import { describe, expect, it, vi } from 'vitest';
import { buildCloseSessionPayload, closeSessionWithFallback } from './closeSessionWorkflow';

const payload = {
  sessionId: 'current',
  expected: 100,
  actual: 95,
  variance: -5,
};

describe('closeSessionWithFallback', () => {
  it('closes online without queueing when the API succeeds', async () => {
    const closeSession = vi.fn(async () => undefined);
    const enqueue = vi.fn();

    const result = await closeSessionWithFallback({ payload, closeSession, enqueue });

    expect(result).toEqual({ status: 'online' });
    expect(closeSession).toHaveBeenCalledWith(payload);
    expect(enqueue).not.toHaveBeenCalled();
  });

  it('durably queues the close payload when the API fails', async () => {
    const closeSession = vi.fn(async () => {
      throw new Error('offline');
    });
    const enqueue = vi.fn();

    const result = await closeSessionWithFallback({ payload, closeSession, enqueue });

    expect(result).toEqual({ status: 'queued' });
    expect(enqueue).toHaveBeenCalledWith(payload);
  });
});

describe('buildCloseSessionPayload', () => {
  it('builds a finite close payload and variance from cash input', () => {
    expect(buildCloseSessionPayload({ sessionId: 'current', expected: 100, actualInput: '95.50' })).toEqual({
      sessionId: 'current',
      expected: 100,
      actual: 95.5,
      variance: -4.5,
    });
  });

  it('rejects malformed actual cash input', () => {
    expect(() =>
      buildCloseSessionPayload({ sessionId: 'current', expected: 100, actualInput: 'abc' })
    ).toThrow('Enter a valid counted cash amount.');
  });

  it('normalizes invalid expected cash to zero', () => {
    expect(buildCloseSessionPayload({ sessionId: 'current', expected: Number.NaN, actualInput: '10' })).toEqual({
      sessionId: 'current',
      expected: 0,
      actual: 10,
      variance: 10,
    });
  });
});
