import { describe, expect, it, beforeEach, vi } from 'vitest';
import { syncPendingOfflineOrders } from './offlineQueueSync';
import { useOfflineQueueStore } from '../../shared/stores/offlineQueueStore';
import type { Order } from '../../shared/types/pos';

const makeOrder = (id: string): Order => ({
  id,
  status: 'paid',
  items: [],
  subtotal: 10,
  tax: 1,
  tip: 0,
  discount: 0,
  total: 11,
  createdAt: '2026-05-06T00:00:00.000Z',
});

beforeEach(() => {
  useOfflineQueueStore.setState({ queue: [], online: true });
});

describe('syncPendingOfflineOrders', () => {
  it('syncs pending orders and removes them from the queue', async () => {
    const queued = useOfflineQueueStore.getState().enqueue(makeOrder('order-1'));

    const result = await syncPendingOfflineOrders(vi.fn(() => Promise.resolve()));

    expect(result).toEqual({ attempted: 1, synced: 1, failed: 0 });
    expect(useOfflineQueueStore.getState().queue.find((q) => q.id === queued.id)).toBeUndefined();
  });

  it('does not automatically retry failed orders', async () => {
    const queued = useOfflineQueueStore.getState().enqueue(makeOrder('order-1'));
    useOfflineQueueStore.getState().markFailed(queued.id, 'Previous failure');
    const syncOrder = vi.fn(() => Promise.resolve());

    const result = await syncPendingOfflineOrders(syncOrder);

    expect(result).toEqual({ attempted: 0, synced: 0, failed: 0 });
    expect(syncOrder).not.toHaveBeenCalled();
    expect(useOfflineQueueStore.getState().queue[0].status).toBe('failed');
  });
});
