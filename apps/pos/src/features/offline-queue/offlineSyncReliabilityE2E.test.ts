/**
 * End-to-end: offline queue reliability under network failure scenarios.
 *
 * Proves that orders queued while offline are correctly synced when
 * connectivity is restored, and that partial failures are handled
 * without losing other queued orders.
 */
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { syncPendingOfflineOrders } from './offlineQueueSync';
import { useOfflineQueueStore } from '../../shared/stores/offlineQueueStore';
import type { Order } from '../../shared/types/pos';

function makeOrder(id: string, total = 20): Order {
  return {
    id,
    status: 'paid',
    items: [],
    subtotal: total,
    tax: total * 0.08,
    tip: 0,
    discount: 0,
    total: total + total * 0.08,
    createdAt: '2026-05-13T10:00:00.000Z',
  };
}

beforeEach(() => {
  useOfflineQueueStore.setState({ queue: [], online: true });
});

describe('offline queue sync under network failures', () => {
  it('syncs all pending orders when network recovers', async () => {
    const store = useOfflineQueueStore.getState();
    store.enqueue(makeOrder('order-A'));
    store.enqueue(makeOrder('order-B'));
    store.enqueue(makeOrder('order-C'));

    const syncFn = vi.fn(async () => undefined);

    const result = await syncPendingOfflineOrders(syncFn);

    expect(result.attempted).toBe(3);
    expect(result.synced).toBe(3);
    expect(result.failed).toBe(0);
    expect(useOfflineQueueStore.getState().queue).toHaveLength(0);
  });

  it('continues syncing remaining orders after one fails (partial network failure)', async () => {
    const store = useOfflineQueueStore.getState();
    store.enqueue(makeOrder('order-1'));
    store.enqueue(makeOrder('order-2'));
    store.enqueue(makeOrder('order-3'));

    let callCount = 0;
    const syncFn = vi.fn(async (order: Order) => {
      callCount++;
      // order-2 fails (e.g. intermittent 503)
      if (order.id === 'order-2') {
        throw new Error('Service temporarily unavailable');
      }
    });

    const result = await syncPendingOfflineOrders(syncFn);

    expect(result.attempted).toBe(3);
    expect(result.synced).toBe(2);   // orders 1 and 3 succeed
    expect(result.failed).toBe(1);   // order 2 failed

    // Failed order stays in queue with 'failed' status
    const queue = useOfflineQueueStore.getState().queue;
    const failedEntry = queue.find((e) => e.order.id === 'order-2');
    expect(failedEntry?.status).toBe('failed');

    // Successfully synced orders are removed
    expect(queue.find((e) => e.order.id === 'order-1')).toBeUndefined();
    expect(queue.find((e) => e.order.id === 'order-3')).toBeUndefined();
  });

  it('does not retry previously failed orders in the same sync pass', async () => {
    const store = useOfflineQueueStore.getState();
    const entry = store.enqueue(makeOrder('order-fail'));
    store.markFailed(entry.id, 'Previous error');

    const syncFn = vi.fn(async () => undefined);
    const result = await syncPendingOfflineOrders(syncFn);

    expect(result.attempted).toBe(0);
    expect(syncFn).not.toHaveBeenCalled();

    // Entry still in queue as 'failed' - needs manual retry or re-queue
    const queue = useOfflineQueueStore.getState().queue;
    expect(queue.find((e) => e.order.id === 'order-fail')?.status).toBe('failed');
  });

  it('checkout falls back to offline queue when server is unreachable', async () => {
    const { buildCheckoutOrder, processCheckout } = await import('../payment/checkoutWorkflow');

    const order = buildCheckoutOrder({
      id: 'local-offline-1',
      items: [],
      subtotal: 20,
      tax: 1.6,
      total: 21.6,
      employeeId: 1,
      createdAt: '2026-05-13T10:00:00.000Z',
    });

    const enqueueOffline = vi.fn((o: Order) => {
      useOfflineQueueStore.getState().enqueue(o);
    });

    const result = await processCheckout({
      order,
      method: 'cash',
      createOrder: vi.fn(async () => {
        throw new Error('ECONNREFUSED - server unreachable');
      }),
      chargeOrder: vi.fn(),
      enqueueOffline,
      recordCashSale: vi.fn(),
    });

    expect(result).toEqual({ status: 'queued' });
    expect(enqueueOffline).toHaveBeenCalledWith(order);

    // Order is now in the offline queue
    const queue = useOfflineQueueStore.getState().queue;
    expect(queue.some((e) => e.order.id === 'local-offline-1')).toBe(true);
  });

  it('queued order syncs successfully when connectivity is restored', async () => {
    // Simulate: offline checkout → queue → network recovers → sync
    const { buildCheckoutOrder, processCheckout } = await import('../payment/checkoutWorkflow');

    const order = buildCheckoutOrder({
      id: 'local-offline-2',
      items: [],
      subtotal: 15,
      tax: 1.2,
      total: 16.2,
      employeeId: 1,
      createdAt: '2026-05-13T10:00:00.000Z',
    });

    const store = useOfflineQueueStore.getState();
    const enqueueOffline = (o: Order) => store.enqueue(o);

    await processCheckout({
      order,
      method: 'cash',
      createOrder: vi.fn(async () => { throw new Error('offline'); }),
      chargeOrder: vi.fn(),
      enqueueOffline,
      recordCashSale: vi.fn(),
    });

    expect(useOfflineQueueStore.getState().queue).toHaveLength(1);

    // Network recovers - sync
    const syncFn = vi.fn(async () => undefined);
    const result = await syncPendingOfflineOrders(syncFn);

    expect(result.synced).toBe(1);
    expect(syncFn).toHaveBeenCalledWith(expect.objectContaining({ id: 'local-offline-2' }));
    expect(useOfflineQueueStore.getState().queue).toHaveLength(0);
  });
});
