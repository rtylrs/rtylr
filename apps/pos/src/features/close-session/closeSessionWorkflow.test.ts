import { describe, expect, it, vi } from 'vitest';
import { buildCloseSessionPayload, closeSessionWithFallback } from './closeSessionWorkflow';

const mockResolveSessionId = async (): Promise<number> => 42;

describe('closeSessionWithFallback', () => {
  it('closes online without queueing when the API succeeds', async () => {
    const closeSession = vi.fn(async () => undefined);
    const enqueue = vi.fn();

    const payload = buildCloseSessionPayload({ expected: 100, actualInput: '95' });
    const result = await closeSessionWithFallback({
      payload,
      closeSession,
      enqueue,
      resolveSessionId: mockResolveSessionId,
    });

    expect(result).toEqual({ status: 'online' });
    expect(closeSession).toHaveBeenCalledWith(42, { closing_float: 95 });
    expect(enqueue).not.toHaveBeenCalled();
  });

  it('durably queues the close payload when the API fails', async () => {
    const payload = buildCloseSessionPayload({ expected: 100, actualInput: '95' });
    const closeSession = vi.fn(async () => {
      throw new Error('offline');
    });
    const enqueue = vi.fn();

    const result = await closeSessionWithFallback({
      payload,
      closeSession,
      enqueue,
      resolveSessionId: mockResolveSessionId,
    });

    expect(result).toEqual({ status: 'queued' });
    expect(enqueue).toHaveBeenCalledWith(payload);
  });
});

describe('buildCloseSessionPayload', () => {
  it('builds a finite close payload and variance from cash input', () => {
    expect(buildCloseSessionPayload({ expected: 100, actualInput: '95.50' })).toEqual({
      closing_float: 95.5,
      expected: 100,
      variance: -4.5,
    });
  });

  it('rejects malformed actual cash input', () => {
    expect(() =>
      buildCloseSessionPayload({ expected: 100, actualInput: 'abc' })
    ).toThrow('Enter a valid counted cash amount.');
  });

  it('normalizes invalid expected cash to zero', () => {
    expect(buildCloseSessionPayload({ expected: Number.NaN, actualInput: '10' })).toEqual({
      closing_float: 10,
      expected: 0,
      variance: 10,
    });
  });
});
