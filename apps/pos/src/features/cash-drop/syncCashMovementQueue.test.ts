import { describe, expect, it, vi } from 'vitest';
import type { QueuedCashMovement } from '../../shared/stores/cashMovementQueueStore';
import { syncQueuedCashMovements } from './syncCashMovementQueue';

const payload = {
  branch_id: 1,
  session_id: 0,
  employee_id: 1,
  amount: 25,
  reason: 'Safe deposit',
};

const entry = (overrides: Partial<QueuedCashMovement> = {}): QueuedCashMovement => ({
  id: 'cm_1',
  action: 'drop',
  payload,
  queuedAt: '2026-05-06T00:00:00.000Z',
  retries: 0,
  status: 'pending',
  ...overrides,
});

describe('syncQueuedCashMovements', () => {
  it('syncs queued cash movements and removes them', async () => {
    const syncMovement = vi.fn(async () => undefined);
    const retry = vi.fn();
    const markSynced = vi.fn();
    const markFailed = vi.fn();

    const result = await syncQueuedCashMovements({
      entries: [entry()],
      syncMovement,
      retry,
      markSynced,
      markFailed,
    });

    expect(result).toEqual({ attempted: 1, synced: 1, failed: 0 });
    expect(syncMovement).toHaveBeenCalledWith('drop', payload);
    expect(retry).toHaveBeenCalledWith('cm_1');
    expect(markSynced).toHaveBeenCalledWith('cm_1');
    expect(markFailed).not.toHaveBeenCalled();
  });

  it('marks failed cash movements with the backend error message', async () => {
    const syncMovement = vi.fn(async () => {
      throw new Error('offline');
    });
    const markFailed = vi.fn();

    const result = await syncQueuedCashMovements({
      entries: [entry()],
      syncMovement,
      retry: vi.fn(),
      markSynced: vi.fn(),
      markFailed,
    });

    expect(result).toEqual({ attempted: 1, synced: 0, failed: 1 });
    expect(markFailed).toHaveBeenCalledWith('cm_1', 'offline');
  });

  it('skips records that are already syncing', async () => {
    const syncMovement = vi.fn();

    const result = await syncQueuedCashMovements({
      entries: [entry({ status: 'syncing' })],
      syncMovement,
      retry: vi.fn(),
      markSynced: vi.fn(),
      markFailed: vi.fn(),
    });

    expect(result).toEqual({ attempted: 0, synced: 0, failed: 0 });
    expect(syncMovement).not.toHaveBeenCalled();
  });
});
