import { beforeEach, describe, expect, it } from 'vitest';
import { useCashMovementQueueStore } from './cashMovementQueueStore';

const { getState, setState } = useCashMovementQueueStore;

beforeEach(() => {
  setState({ queue: [] });
});

describe('cashMovementQueueStore', () => {
  it('enqueues cash movement payloads as pending', () => {
    const queued = getState().enqueue({
      action: 'drop',
      payload: { amount: 25, reason: 'Safe deposit' },
    });

    expect(queued.id).toMatch(/^cm_/);
    expect(queued.action).toBe('drop');
    expect(queued.payload).toEqual({ amount: 25, reason: 'Safe deposit' });
    expect(queued.status).toBe('pending');
    expect(queued.retries).toBe(0);
    expect(getState().queue).toHaveLength(1);
  });

  it('marks a cash movement as syncing when retried', () => {
    const queued = getState().enqueue({ action: 'paid_out', payload: { amount: 10, reason: 'Petty cash' } });

    getState().retry(queued.id);

    expect(getState().queue[0]).toMatchObject({ retries: 1, status: 'syncing' });
  });

  it('removes synced cash movements', () => {
    const queued = getState().enqueue({ action: 'paid_in', payload: { amount: 10, reason: 'Correction' } });

    getState().markSynced(queued.id);

    expect(getState().queue).toEqual([]);
  });

  it('stores failure details for retry', () => {
    const queued = getState().enqueue({ action: 'drop', payload: { amount: 10, reason: 'Safe' } });

    getState().markFailed(queued.id, 'offline');

    expect(getState().queue[0]).toMatchObject({ status: 'failed', lastError: 'offline' });
  });
});
