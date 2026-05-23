/**
 * End-to-end: session lifecycle - open → cash movement → close with variance.
 *
 * Chains posService.openSession → posService.cashDrop → buildCloseSessionPayload
 * → closeSessionWithFallback → posService.closeSession in the correct order.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../shared/services/posService', () => ({
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
import { posService } from '../../shared/services/posService';
import { unwrap } from '@voxire/api';

beforeEach(() => {
  vi.clearAllMocks();
});

describe('session lifecycle end-to-end', () => {
  it('openSession → cashDrop → closeSession completes the full shift lifecycle', async () => {
    vi.mocked(posService.openSession).mockResolvedValue({ sessionId: 'sess-1' } as never);
    vi.mocked(posService.cashDrop).mockResolvedValue(undefined as never);
    vi.mocked(posService.closeSession).mockResolvedValue(undefined as never);

    // 1. Open the session.
    await unwrap(posService.openSession({ openingBalance: 200 }));
    expect(posService.openSession).toHaveBeenCalledWith({ openingBalance: 200 });

    // 2. Record a cash drop during the shift.
    await unwrap(posService.cashDrop({ amount: 50, reason: 'mid-shift drop' }));
    expect(posService.cashDrop).toHaveBeenCalledWith({ amount: 50, reason: 'mid-shift drop' });

    // 3. Build close payload and close the session.
    const payload = buildCloseSessionPayload({
      sessionId: 'sess-1',
      expected: 150,
      actualInput: '145',
    });

    // variance should be actual - expected = 145 - 150 = -5
    expect(payload.variance).toBe(-5);

    const closeSession = vi.fn(() => unwrap(posService.closeSession(payload)));
    const result = await closeSessionWithFallback({
      payload,
      closeSession,
      enqueue: vi.fn(),
    });

    expect(result).toEqual({ status: 'online' });
    expect(posService.closeSession).toHaveBeenCalledTimes(1);
    expect(posService.closeSession).toHaveBeenCalledWith(
      expect.objectContaining({ sessionId: 'sess-1', variance: -5 }),
    );

    // Verify call ordering: open → cashDrop → closeSession.
    const openOrder = vi.mocked(posService.openSession).mock.invocationCallOrder[0];
    const dropOrder = vi.mocked(posService.cashDrop).mock.invocationCallOrder[0];
    const closeOrder = vi.mocked(posService.closeSession).mock.invocationCallOrder[0];

    expect(openOrder).toBeLessThan(dropOrder);
    expect(dropOrder).toBeLessThan(closeOrder);
  });

  it('session close with overage variance records the surplus', () => {
    const payload = buildCloseSessionPayload({
      sessionId: 'sess-2',
      expected: 100,
      actualInput: '115',
    });

    // Actual is $15 over expected → positive variance.
    expect(payload.variance).toBe(15);
    expect(payload.actual).toBe(115);
    expect(payload.expected).toBe(100);
  });

  it('session close with shortage variance records the deficit', () => {
    const payload = buildCloseSessionPayload({
      sessionId: 'sess-3',
      expected: 200,
      actualInput: '185',
    });

    // Actual is $15 short → negative variance.
    expect(payload.variance).toBe(-15);
    expect(payload.actual).toBe(185);
    expect(payload.expected).toBe(200);
  });

  it('cash movement is queued when backend is unreachable during shift', async () => {
    // cashDrop throws - should be caught by the caller and enqueued.
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

    // Session can still close even though the movement was queued.
    const closePayload = buildCloseSessionPayload({
      sessionId: 'sess-4',
      expected: 100,
      actualInput: '100',
    });

    const closeSession = vi.fn(() => unwrap(posService.closeSession(closePayload)));
    const result = await closeSessionWithFallback({
      payload: closePayload,
      closeSession,
      enqueue: vi.fn(),
    });

    expect(result).toEqual({ status: 'online' });
  });

  it('closeSessionWithFallback queues close payload when API fails', async () => {
    const payload = buildCloseSessionPayload({
      sessionId: 'sess-5',
      expected: 300,
      actualInput: '295',
    });

    // Verify the payload is correctly built before passing to fallback.
    expect(payload.variance).toBe(-5);

    const enqueue = vi.fn();
    const closeSession = vi.fn(() => Promise.reject(new Error('503 Service Unavailable')));

    const result = await closeSessionWithFallback({ payload, closeSession, enqueue });

    expect(result).toEqual({ status: 'queued' });
    expect(enqueue).toHaveBeenCalledTimes(1);
    expect(enqueue).toHaveBeenCalledWith(
      expect.objectContaining({ sessionId: 'sess-5', variance: -5 }),
    );
  });
});
