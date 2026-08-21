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
    expect(closeSession).toHaveBeenCalledWith({ id: '42', closingFloat: 95 });
    expect(enqueue).not.toHaveBeenCalled();
  });

  it('durably queues the close payload with its session id when the API fails', async () => {
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
    expect(enqueue).toHaveBeenCalledWith({ ...payload, sessionId: 42 });
  });

  it('does not queue when no open session exists (permanent condition)', async () => {
    const closeSession = vi.fn(async () => undefined);
    const enqueue = vi.fn();

    await expect(
      closeSessionWithFallback({
        payload: buildCloseSessionPayload({ expected: 100, actualInput: '95' }),
        closeSession,
        enqueue,
        resolveSessionId: async () => {
          throw new Error('No open register session found.');
        },
      })
    ).rejects.toThrow('No open register session found.');

    expect(enqueue).not.toHaveBeenCalled();
  });

  it('rethrows permanent server rejections instead of queueing them', async () => {
    const rejection = Object.assign(new Error('unexpected status 422'), { statusCode: 422 });
    const enqueue = vi.fn();

    await expect(
      closeSessionWithFallback({
        payload: buildCloseSessionPayload({ expected: 100, actualInput: '95' }),
        closeSession: vi.fn(async () => {
          throw rejection;
        }),
        enqueue,
        resolveSessionId: mockResolveSessionId,
      })
    ).rejects.toThrow('unexpected status 422');

    expect(enqueue).not.toHaveBeenCalled();
  });
});

describe('buildCloseSessionPayload', () => {
  it('builds a finite close payload and variance from cash input', () => {
    expect(buildCloseSessionPayload({ expected: 100, actualInput: '95.50' })).toEqual({
      closingFloat: 95.5,
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
      closingFloat: 10,
      expected: 0,
      variance: 10,
    });
  });
});
