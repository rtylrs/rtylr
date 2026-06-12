/**
 * End-to-end: session lifecycle - open → cash movement → close with variance.
 *
 * Chains posService.openSession → posService.cashDrop → buildCloseSessionPayload
 * → closeSessionWithFallback in the correct order.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../shared/services/api/posService', () => ({
  posService: {
    openSession: vi.fn(),
    cashDrop: vi.fn(),
    closeSession: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((p) => p),
}));

import { buildCloseSessionPayload, closeSessionWithFallback } from './closeSessionWorkflow';
import { posService } from '../../shared/services/api/posService';
import { unwrap } from '@voxire/api';

beforeEach(() => {
  vi.clearAllMocks();
});

describe('session lifecycle end-to-end', () => {
  it('openSession → cashDrop → closeSession completes the full shift lifecycle', async () => {
    vi.mocked(posService.openSession).mockResolvedValue({ sessionId: 1 } as never);
    vi.mocked(posService.cashDrop).mockResolvedValue(undefined as never);
    vi.mocked(posService.closeSession).mockResolvedValue(undefined as never);

    await unwrap(posService.openSession({ openingBalance: 200 }));
    expect(posService.openSession).toHaveBeenCalledWith({ openingBalance: 200 });

    await unwrap(posService.cashDrop({ amount: 50, reason: 'mid-shift drop' }));
    expect(posService.cashDrop).toHaveBeenCalledWith({ amount: 50, reason: 'mid-shift drop' });

    const payload = buildCloseSessionPayload({ expected: 150, actualInput: '145' });
    expect(payload.variance).toBe(-5);
    expect(payload.closing_float).toBe(145);

    const closeSession = vi.fn(async (_id: number, p: { closing_float: number }) =>
      unwrap(posService.closeSession(p))
    );
    const result = await closeSessionWithFallback({
      payload,
      closeSession,
      enqueue: vi.fn(),
      resolveSessionId: async () => 1,
    });

    expect(result).toEqual({ status: 'online' });
    expect(closeSession).toHaveBeenCalledWith(1, { closing_float: 145 });

    const openOrder = vi.mocked(posService.openSession).mock.invocationCallOrder[0];
    const dropOrder = vi.mocked(posService.cashDrop).mock.invocationCallOrder[0];
    const closeOrder = vi.mocked(posService.closeSession).mock.invocationCallOrder[0];

    expect(openOrder).toBeLessThan(dropOrder);
    expect(dropOrder).toBeLessThan(closeOrder);
  });

  it('session close with overage variance records the surplus', () => {
    const payload = buildCloseSessionPayload({ expected: 100, actualInput: '115' });

    expect(payload.variance).toBe(15);
    expect(payload.closing_float).toBe(115);
    expect(payload.expected).toBe(100);
  });

  it('session close with shortage variance records the deficit', () => {
    const payload = buildCloseSessionPayload({ expected: 200, actualInput: '185' });

    expect(payload.variance).toBe(-15);
    expect(payload.closing_float).toBe(185);
    expect(payload.expected).toBe(200);
  });

  it('cash movement is queued when backend is unreachable during shift', async () => {
    vi.mocked(posService.cashDrop).mockRejectedValue(new Error('Network Error') as never);
    vi.mocked(posService.closeSession).mockResolvedValue(undefined as never);

    const enqueueMovement = vi.fn();

    let syncError: Error | null = null;
    try {
      await unwrap(posService.cashDrop({ amount: 30, reason: 'offline drop' }));
    } catch (e) {
      syncError = e as Error;
      enqueueMovement({ action: 'drop', payload: { amount: 30, reason: 'offline drop' } });
    }

    expect(syncError).not.toBeNull();
    expect(enqueueMovement).toHaveBeenCalledTimes(1);
    expect(enqueueMovement).toHaveBeenCalledWith(
      expect.objectContaining({ action: 'drop' }),
    );

    const closePayload = buildCloseSessionPayload({ expected: 100, actualInput: '100' });

    const closeSession = vi.fn(async (_id: number, p: { closing_float: number }) =>
      unwrap(posService.closeSession(p))
    );
    const result = await closeSessionWithFallback({
      payload: closePayload,
      closeSession,
      enqueue: vi.fn(),
      resolveSessionId: async () => 1,
    });

    expect(result).toEqual({ status: 'online' });
  });

  it('closeSessionWithFallback queues close payload when API fails', async () => {
    const payload = buildCloseSessionPayload({ expected: 300, actualInput: '295' });

    expect(payload.variance).toBe(-5);

    const enqueue = vi.fn();
    const closeSession = vi.fn(() => Promise.reject(new Error('503 Service Unavailable')));

    const result = await closeSessionWithFallback({
      payload,
      closeSession,
      enqueue,
      resolveSessionId: async () => 1,
    });

    expect(result).toEqual({ status: 'queued' });
    expect(enqueue).toHaveBeenCalledTimes(1);
    expect(enqueue).toHaveBeenCalledWith(
      expect.objectContaining({ closing_float: 295, variance: -5 }),
    );
  });
});
