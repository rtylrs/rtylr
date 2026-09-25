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
import type { CreateOrderRequest } from '@voxire/voxie/pos/orders/v1/types';
import type { CartItem } from '../../shared/stores/cartStore';

function makePayload(notes: string): CreateOrderRequest {
  return {
    sessionId: '0',
    tableId: '0',
    locationId: '0',
    type: '',
    customerId: '0',
    customerName: '',
    paymentMethod: 'cash',
    notes,
    kitchenNotes: '',
    items: [],
  };
}

beforeEach(() => {
  useOfflineQueueStore.setState({ queue: [], online: true });
});

describe('offline queue sync under network failures', () => {
  it('syncs all pending orders when network recovers', async () => {
    const store = useOfflineQueueStore.getState();
    store.enqueue(makePayload('order-A'));
    store.enqueue(makePayload('order-B'));
    store.enqueue(makePayload('order-C'));

    const syncFn = vi.fn(async () => undefined);

    const result = await syncPendingOfflineOrders(syncFn);

    expect(result.attempted).toBe(3);
    expect(result.synced).toBe(3);
    expect(result.failed).toBe(0);
    expect(useOfflineQueueStore.getState().queue).toHaveLength(0);
  });

  it('continues syncing remaining orders after one fails (partial network failure)', async () => {
    const store = useOfflineQueueStore.getState();
    store.enqueue(makePayload('order-1'));
    store.enqueue(makePayload('order-2'));
    store.enqueue(makePayload('order-3'));

    let callCount = 0;
    const syncFn = vi.fn(async (order: CreateOrderRequest) => {
      callCount++;
      if (order.notes === 'order-2') {
        throw new Error('Service temporarily unavailable');
      }
    });

    const result = await syncPendingOfflineOrders(syncFn);

    expect(result.attempted).toBe(3);
    expect(result.synced).toBe(2);
    expect(result.failed).toBe(1);

    const queue = useOfflineQueueStore.getState().queue;
    const failedEntry = queue.find((e) => (e.order as CreateOrderRequest).notes === 'order-2');
    expect(failedEntry?.status).toBe('failed');

    expect(queue.find((e) => (e.order as CreateOrderRequest).notes === 'order-1')).toBeUndefined();
    expect(queue.find((e) => (e.order as CreateOrderRequest).notes === 'order-3')).toBeUndefined();
  });

  it('does not retry previously failed orders in the same sync pass', async () => {
    const store = useOfflineQueueStore.getState();
    const entry = store.enqueue(makePayload('order-fail'));
    store.markFailed(entry.id, 'Previous error');

    const syncFn = vi.fn(async () => undefined);
    const result = await syncPendingOfflineOrders(syncFn);

    expect(result.attempted).toBe(0);
    expect(syncFn).not.toHaveBeenCalled();

    const queue = useOfflineQueueStore.getState().queue;
    expect(queue.find((e) => (e.order as CreateOrderRequest).notes === 'order-fail')?.status).toBe('failed');
  });

  it('checkout falls back to offline queue when server is unreachable', async () => {
    const { buildCheckoutOrder, buildCreateOrderPayload, processCheckout } = await import('../payment/checkoutWorkflow');

    const items: CartItem[] = [];
    const payload = buildCreateOrderPayload({ items, method: 'cash' });
    const order = buildCheckoutOrder({
      id: 'local-offline-1',
      items,
      subtotal: 20,
      tax: 1.6,
      total: 21.6,
      customerId: 1,
      createdAt: '2026-05-13T10:00:00.000Z',
    });

    const enqueuedPayloads: CreateOrderRequest[] = [];
    const enqueueOffline = vi.fn((p: CreateOrderRequest) => {
      enqueuedPayloads.push(p);
      useOfflineQueueStore.getState().enqueue(p);
    });

    const result = await processCheckout({
      order,
      payload,
      method: 'cash',
      createOrder: vi.fn(async () => {
        throw new Error('ECONNREFUSED - server unreachable');
      }),
      chargeOrder: vi.fn(),
      enqueueOffline,
      recordCashSale: vi.fn(),
    });

    expect(result).toEqual({ status: 'queued' });
    expect(enqueueOffline).toHaveBeenCalledWith(payload);

    const queue = useOfflineQueueStore.getState().queue;
    expect(queue).toHaveLength(1);
  });

  it('queued order syncs successfully when connectivity is restored', async () => {
    const { buildCheckoutOrder, buildCreateOrderPayload, processCheckout } = await import('../payment/checkoutWorkflow');

    const items: CartItem[] = [];
    const payload = buildCreateOrderPayload({ items, method: 'cash', notes: 'local-offline-2' });
    const order = buildCheckoutOrder({
      id: 'local-offline-2',
      items,
      subtotal: 15,
      tax: 1.2,
      total: 16.2,
      customerId: 1,
      createdAt: '2026-05-13T10:00:00.000Z',
    });

    const store = useOfflineQueueStore.getState();
    const enqueueOffline = (p: CreateOrderRequest) => store.enqueue(p);

    await processCheckout({
      order,
      payload,
      method: 'cash',
      createOrder: vi.fn(async () => { throw new Error('offline'); }),
      chargeOrder: vi.fn(),
      enqueueOffline,
      recordCashSale: vi.fn(),
    });

    expect(useOfflineQueueStore.getState().queue).toHaveLength(1);

    const syncFn = vi.fn(async () => undefined);
    const result = await syncPendingOfflineOrders(syncFn);

    expect(result.synced).toBe(1);
    expect(syncFn).toHaveBeenCalledWith(
      expect.objectContaining({ notes: 'local-offline-2' })
    );
    expect(useOfflineQueueStore.getState().queue).toHaveLength(0);
  });
});
