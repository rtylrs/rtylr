import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useCashMovementQueueStore } from '../../shared/stores/cashMovementQueueStore';
import { retryCashMovement, syncAllCashMovements, syncPendingCashMovements } from './cashMovementQueueSync';

const { getState, setState } = useCashMovementQueueStore;

beforeEach(() => {
  setState({ queue: [] });
});

describe('cashMovementQueueSync', () => {
  it('syncs only pending cash movements during auto sync', async () => {
    const pending = getState().enqueue({
      action: 'drop',
      payload: { branch_id: 1, session_id: 0, employee_id: 1, amount: 10, reason: 'Safe' },
    });
    const failed = getState().enqueue({
      action: 'paid_in',
      payload: { branch_id: 1, session_id: 0, employee_id: 1, amount: 5, reason: 'Correction' },
    });
    getState().markFailed(failed.id, 'offline');
    const syncMovement = vi.fn(async () => undefined);

    const result = await syncPendingCashMovements(syncMovement);

    expect(result).toEqual({ attempted: 1, synced: 1, failed: 0 });
    expect(syncMovement).toHaveBeenCalledWith(pending.action, pending.payload);
    expect(getState().queue.map((item) => item.id)).toEqual([failed.id]);
  });

  it('syncs failed cash movements during manual sync all', async () => {
    const failed = getState().enqueue({
      action: 'paid_out',
      payload: { branch_id: 1, session_id: 0, employee_id: 1, amount: 5, reason: 'Petty cash' },
    });
    getState().markFailed(failed.id, 'offline');
    const syncMovement = vi.fn(async () => undefined);

    const result = await syncAllCashMovements(syncMovement);

    expect(result).toEqual({ attempted: 1, synced: 1, failed: 0 });
    expect(getState().queue).toEqual([]);
  });

  it('retries one cash movement by id', async () => {
    const queued = getState().enqueue({
      action: 'drop',
      payload: { branch_id: 1, session_id: 0, employee_id: 1, amount: 10, reason: 'Safe' },
    });

    const synced = await retryCashMovement(queued.id, vi.fn(async () => undefined));

    expect(synced).toBe(true);
    expect(getState().queue).toEqual([]);
  });
});
