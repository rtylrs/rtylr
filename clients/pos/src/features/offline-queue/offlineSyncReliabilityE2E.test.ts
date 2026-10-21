/**
 * End-to-end: offline queue reliability under network failure scenarios.
 *
 * Proves that orders queued while offline are correctly synced when
 * connectivity is restored, and that partial failures are handled
 * without losing other queued orders.
 *
 * The API clients are stubbed, not the replay: these tests drive the real
 * `replayQueuedOrder` so the charge half of the replay is covered. An earlier
 * version of this file injected a fake `syncOrder` and therefore only ever
 * proved that `createOrder` was called, which is exactly how a replay that
 * creates the order and never charges it passed for green.
 */
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { syncPendingOfflineOrders } from './offlineQueueSync';
import { useOfflineQueueStore } from '../../shared/stores/offlineQueueStore';
import { buildOfflineTender } from '../payment/checkoutWorkflow';
import { useCashDrawerStore } from '../../shared/stores/cashDrawerStore';
import type { CreateOrderRequest } from '@voxire/voxie/pos/orders/v1/types';
import type { CartItem } from '../../shared/stores/cartStore';
import type { OfflineTender, Order } from '../../shared/types/pos';

const api = vi.hoisted(() => ({
  createOrder: vi.fn(),
  processPayment: vi.fn(),
}));

vi.mock('@voxire/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@voxire/api')>();
  return {
    ...actual,
    createOneKitClient: () => ({ createOrder: api.createOrder, processPayment: api.processPayment }),
  };
});

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

/** oneKit ApiError shape: numeric statusCode plus a JSON body string. */
function apiError(statusCode: number, message: string): Error {
  return Object.assign(new Error(`Request failed with status ${statusCode}`), {
    statusCode,
    body: JSON.stringify({ message, status_code: statusCode }),
  });
}

/** Queue a cash sale through the real checkout workflow, offline. */
async function runOfflineCheckout(order: Order, items: CartItem[], notes: string): Promise<void> {
  const { buildCreateOrderPayload, processCheckout } = await import('../payment/checkoutWorkflow');
  const result = await processCheckout({
    order,
    payload: buildCreateOrderPayload({ items, method: 'cash', notes }),
    method: 'cash',
    createOrder: vi.fn(async () => {
      throw new Error('ECONNREFUSED - server unreachable');
    }),
    chargeOrder: vi.fn(),
    enqueueOffline: (p) => {
      // The same call features/payment/index.tsx makes: the tender goes in
      // with the entry, in one write, and the key is minted per entry.
      useOfflineQueueStore.getState().enqueue(p, {
        tender: buildOfflineTender('cash', order),
      });
    },
    recordCashSale: (amount, orderId) => useCashDrawerStore.getState().recordCashSale(amount, orderId),
  });
  expect(result).toEqual({ status: 'queued' });
}

function buildCheckoutOrderFor(total: number, id: string): Order {
  return {
    id,
    status: 'paid',
    items: [],
    subtotal: total,
    tax: 0,
    tip: 0,
    discount: 0,
    total,
    createdAt: '2026-05-13T10:00:00.000Z',
  };
}

async function queueOfflineCashSale(total: number, notes: string) {
  const { buildCheckoutOrder, buildCreateOrderPayload } = await import('../payment/checkoutWorkflow');
  const items: CartItem[] = [];
  const payload = buildCreateOrderPayload({ items, method: 'cash', notes });
  const order = buildCheckoutOrder({
    id: `local-${notes}`,
    items,
    subtotal: total,
    tax: 0,
    total,
    customerId: 1,
    createdAt: '2026-05-13T10:00:00.000Z',
  });

  await runOfflineCheckout(order, items, notes);
  return payload;
}

beforeEach(() => {
  useOfflineQueueStore.setState({ queue: [], online: true });
  useCashDrawerStore.setState({ movements: [], openingFloat: 0, opened: true });
  api.createOrder.mockReset();
  api.processPayment.mockReset();
  api.createOrder.mockResolvedValue({ orderId: '501' });
  api.processPayment.mockResolvedValue({ message: 'Payment processed' });
});

/** The header the generated client actually put on the wire for a call. */
const keyOf = (call: unknown[]): string | undefined =>
  ((call[1] as { headers?: Record<string, string> } | undefined)?.headers ?? {})[
    'Idempotency-Key'
  ];

const tender = (overrides: Partial<OfflineTender> = {}): OfflineTender => ({
  paymentMethod: 'cash',
  amountPaid: 16.2,
  localOrderId: 'local-tender',
  ...overrides,
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

    const syncFn = vi.fn(async (entry) => {
      if (entry.order.notes === 'order-2') {
        throw new Error('Service temporarily unavailable');
      }
    });

    const result = await syncPendingOfflineOrders(syncFn);

    expect(result.attempted).toBe(3);
    expect(result.synced).toBe(2);
    expect(result.failed).toBe(1);

    const queue = useOfflineQueueStore.getState().queue;
    const failedEntry = queue.find((e) => e.order.notes === 'order-2');
    expect(failedEntry?.status).toBe('failed');

    expect(queue.find((e) => e.order.notes === 'order-1')).toBeUndefined();
    expect(queue.find((e) => e.order.notes === 'order-3')).toBeUndefined();
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
    expect(queue.find((e) => e.order.notes === 'order-fail')?.status).toBe('failed');
  });

  it('checkout falls back to offline queue when server is unreachable', async () => {
    const { buildCheckoutOrder, buildCreateOrderPayload, processCheckout } = await import(
      '../payment/checkoutWorkflow'
    );

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
    const chargeOrder = vi.fn();
    const enqueueOffline = vi.fn((p: CreateOrderRequest) => {
      enqueuedPayloads.push(p);
      useOfflineQueueStore.getState().enqueue(p, {
        tender: buildOfflineTender('cash', order),
      });
    });

    const result = await processCheckout({
      order,
      payload,
      method: 'cash',
      createOrder: vi.fn(async () => {
        throw new Error('ECONNREFUSED - server unreachable');
      }),
      chargeOrder,
      enqueueOffline,
      recordCashSale: vi.fn(),
    });

    expect(result).toEqual({ status: 'queued' });
    expect(enqueueOffline).toHaveBeenCalledWith(payload);
    // Charging is the drain's job, never checkout's: the order does not exist
    // yet, so there is nothing to charge.
    expect(chargeOrder).not.toHaveBeenCalled();

    const queue = useOfflineQueueStore.getState().queue;
    expect(queue).toHaveLength(1);
  });

  it('records the cash tender against the drawer movement it came from', async () => {
    await queueOfflineCashSale(16.2, 'local-tender');

    const sale = useCashDrawerStore
      .getState()
      .movements.find((m) => m.type === 'sale' && m.amount > 0);
    const [entry] = useOfflineQueueStore.getState().queue;
    const tender = entry.tender;

    expect(sale).toBeDefined();
    expect(tender).toBeDefined();
    expect(tender?.paymentMethod).toBe('cash');
    expect(tender?.amountPaid).toBe(sale?.amount);
  });

  it('charges the queued cash sale when connectivity is restored', async () => {
    const payload = await queueOfflineCashSale(16.2, 'local-offline-2');

    expect(useOfflineQueueStore.getState().queue).toHaveLength(1);
    // The queued payload alone carries no amount, so nothing was replayable
    // until the tender was captured at checkout.
    expect(payload.paymentMethod).toBe('cash');

    const result = await syncPendingOfflineOrders();

    expect(result).toEqual({ attempted: 1, synced: 1, failed: 0 });
    expect(api.createOrder).toHaveBeenCalledWith(expect.objectContaining({ notes: 'local-offline-2' }));
    expect(api.processPayment).toHaveBeenCalledWith(
      {
        orderId: '501',
        paymentMethod: 'cash',
        amountPaid: 16.2,
      },
      { headers: { 'Idempotency-Key': expect.any(String) } },
    );
    expect(useOfflineQueueStore.getState().queue).toHaveLength(0);
  });

  it('keeps the entry queued for a payment retry when the charge fails', async () => {
    await queueOfflineCashSale(16.2, 'local-charge-fails');
    api.processPayment.mockRejectedValueOnce(new Error('gateway timeout'));

    const result = await syncPendingOfflineOrders();

    expect(result).toEqual({ attempted: 1, synced: 0, failed: 1 });
    const entry = useOfflineQueueStore.getState().queue[0];
    expect(entry.status).toBe('failed');
    expect(entry.lastError).toBe('gateway timeout');
  });

  it('retries the charge on the same order, under the same key, not a second sale', async () => {
    await queueOfflineCashSale(16.2, 'local-crash');
    api.processPayment.mockRejectedValueOnce(new Error('gateway timeout'));

    await syncPendingOfflineOrders();

    // The assigned order id survived the failed charge, so the retry must not
    // re-create: a second create would be a second sale for one tender.
    expect(useOfflineQueueStore.getState().queue[0].serverOrderId).toBe('501');

    useOfflineQueueStore.setState((s) => ({
      queue: s.queue.map((q) => ({ ...q, status: 'pending' as const })),
    }));
    api.processPayment.mockResolvedValue({ message: 'Payment processed' });

    const result = await syncPendingOfflineOrders();

    expect(result).toEqual({ attempted: 1, synced: 1, failed: 0 });
    expect(api.createOrder).toHaveBeenCalledTimes(1);
    expect(api.processPayment).toHaveBeenCalledTimes(2);
    expect(api.processPayment).toHaveBeenLastCalledWith(
      {
        orderId: '501',
        paymentMethod: 'cash',
        amountPaid: 16.2,
      },
      { headers: { 'Idempotency-Key': expect.any(String) } },
    );
    expect(useOfflineQueueStore.getState().queue).toHaveLength(0);
  });

  it('replays under the key the sale was queued with, not a fresh one', async () => {
    // A crash between sending the charge and reading the response is exactly
    // the case a per-request key cannot survive: the first attempt may already
    // have charged, so the retry has to be recognisable as the same action.
    await queueOfflineCashSale(16.2, 'local-same-key');
    const queuedKey = useOfflineQueueStore.getState().queue[0].idempotencyKey;

    api.processPayment.mockRejectedValueOnce(new Error('socket hang up'));
    await syncPendingOfflineOrders();
    api.processPayment.mockResolvedValue({ message: 'Payment processed' });
    useOfflineQueueStore.setState((s) => ({
      queue: s.queue.map((q) => ({ ...q, status: 'pending' as const })),
    }));
    await syncPendingOfflineOrders();

    const keys = api.processPayment.mock.calls.map(keyOf);
    expect(keys).toHaveLength(2);
    expect(keys[0]).toBe(queuedKey);
    expect(keys[1]).toBe(queuedKey);
  });

  it('gives two queued sales two different keys', async () => {
    await queueOfflineCashSale(16.2, 'local-key-a');
    await queueOfflineCashSale(9.5, 'local-key-b');

    const [first, second] = useOfflineQueueStore.getState().queue;
    expect(first.idempotencyKey).not.toBe(second.idempotencyKey);

    api.createOrder
      .mockResolvedValueOnce({ orderId: '501' })
      .mockResolvedValueOnce({ orderId: '502' });
    await syncPendingOfflineOrders();

    const keys = api.processPayment.mock.calls.map(keyOf);
    expect(new Set(keys).size).toBe(2);
  });

  it('treats an already-paid order on replay as a settled sale, not a failure', async () => {
    await queueOfflineCashSale(16.2, 'local-already-paid');
    const queued = useOfflineQueueStore.getState().queue[0];
    const key = queued.idempotencyKey;

    await syncPendingOfflineOrders();
    expect(api.createOrder).toHaveBeenCalledTimes(1);

    // The process then died after the charge committed but before markSynced
    // persisted. What survived the crash is the entry plus the order id written
    // before the charge and the key it was queued under.
    api.createOrder.mockClear();
    api.processPayment.mockRejectedValueOnce(apiError(409, 'Order is already paid'));
    useOfflineQueueStore.setState({
      queue: [{ ...queued, serverOrderId: '501', retries: 1, status: 'syncing' }],
    });

    const result = await syncPendingOfflineOrders();

    expect(result).toEqual({ attempted: 1, synced: 1, failed: 0 });
    // Re-charging the same order, not creating and charging a second sale.
    expect(api.createOrder).not.toHaveBeenCalled();
    expect(api.processPayment).toHaveBeenCalledWith(
      {
        orderId: '501',
        paymentMethod: 'cash',
        amountPaid: 16.2,
      },
      { headers: { 'Idempotency-Key': key } },
    );
    expect(useOfflineQueueStore.getState().queue).toHaveLength(0);
  });

  it('does not double-charge when a drain is triggered more than once', async () => {
    await queueOfflineCashSale(16.2, 'local-double-drain');

    await Promise.all([
      syncPendingOfflineOrders(),
      syncPendingOfflineOrders(),
      syncPendingOfflineOrders(),
    ]);
    await syncPendingOfflineOrders();

    expect(api.createOrder).toHaveBeenCalledTimes(1);
    expect(api.processPayment).toHaveBeenCalledTimes(1);
    expect(useOfflineQueueStore.getState().queue).toHaveLength(0);
  });

  it('keeps a failed entry re-chargeable after draining other orders', async () => {
    await queueOfflineCashSale(16.2, 'local-failed');
    api.createOrder.mockResolvedValueOnce({ orderId: '501' });
    api.processPayment.mockRejectedValueOnce(new Error('gateway timeout'));

    await syncPendingOfflineOrders();
    expect(useOfflineQueueStore.getState().queue[0].status).toBe('failed');

    // A second offline sale drains while the first is still failed. Nothing may
    // discard the failed entry's recorded order id, and its retry would then
    // create a second order.
    await queueOfflineCashSale(9.5, 'local-other');
    api.createOrder.mockResolvedValueOnce({ orderId: '502' });
    api.processPayment.mockResolvedValue({ message: 'Payment processed' });

    await syncPendingOfflineOrders();
    expect(useOfflineQueueStore.getState().queue).toHaveLength(1);
    expect(useOfflineQueueStore.getState().queue[0].status).toBe('failed');

    // Operator presses Retry on the failed entry.
    useOfflineQueueStore.setState((s) => ({
      queue: s.queue.map((q) => ({ ...q, status: 'pending' as const })),
    }));
    await syncPendingOfflineOrders();

    expect(api.createOrder).toHaveBeenCalledTimes(2);
    expect(api.processPayment).toHaveBeenLastCalledWith(
      {
        orderId: '501',
        paymentMethod: 'cash',
        amountPaid: 16.2,
      },
      { headers: { 'Idempotency-Key': expect.any(String) } },
    );
    expect(useOfflineQueueStore.getState().queue).toHaveLength(0);
  });

  it('gives two identical offline sales their own order, tender and key', async () => {
    const first = buildCheckoutOrderFor(16.2, 'local-twin');
    const second = buildCheckoutOrderFor(16.2, 'local-twin');

    for (const order of [first, second]) {
      await runOfflineCheckout(order, [], 'local-twin');
    }
    expect(useOfflineQueueStore.getState().queue).toHaveLength(2);

    api.createOrder
      .mockResolvedValueOnce({ orderId: '501' })
      .mockResolvedValueOnce({ orderId: '502' });

    const result = await syncPendingOfflineOrders();

    expect(result).toEqual({ attempted: 2, synced: 2, failed: 0 });
    // Byte-identical payloads: only the entry carrying each tender and key
    // keeps them from sharing an order or a charge.
    expect(api.createOrder).toHaveBeenCalledTimes(2);
    expect(api.processPayment).toHaveBeenNthCalledWith(
      1,
      { orderId: '501', paymentMethod: 'cash', amountPaid: 16.2 },
      { headers: { 'Idempotency-Key': expect.any(String) } },
    );
    expect(api.processPayment).toHaveBeenNthCalledWith(
      2,
      { orderId: '502', paymentMethod: 'cash', amountPaid: 16.2 },
      { headers: { 'Idempotency-Key': expect.any(String) } },
    );
    const keys = api.processPayment.mock.calls.map(keyOf);
    expect(keys[0]).not.toBe(keys[1]);
  });

  it('completes a pre-upgrade entry once and reuses the key on its retry', async () => {
    // An entry persisted before the tender and key lived on the entry itself.
    const payload: CreateOrderRequest = {
      ...makePayload('local-legacy'),
      items: [{ menuItemId: '7', name: 'Latte', quantity: 1, unitPrice: 10, discountPct: 0, taxRate: 0 }],
    };
    useOfflineQueueStore.setState((s) => ({
      queue: [
        {
          id: 'legacy',
          order: payload,
          queuedAt: '2026-05-13T10:00:00.000Z',
          retries: 0,
          status: 'syncing',
        } as unknown as (typeof s.queue)[number],
      ],
    }));

    api.processPayment.mockRejectedValueOnce(new Error('gateway timeout'));
    await syncPendingOfflineOrders();

    const [entry] = useOfflineQueueStore.getState().queue;
    // Derived from the line items, because nothing recorded what was tendered.
    expect(entry.tender?.amountPaid).toBe(10);
    expect(entry.idempotencyKey).toEqual(expect.any(String));

    api.processPayment.mockResolvedValue({ message: 'Payment processed' });
    useOfflineQueueStore.setState((s) => ({
      queue: s.queue.map((q) => ({ ...q, status: 'pending' as const })),
    }));
    await syncPendingOfflineOrders();

    const keys = api.processPayment.mock.calls.map(keyOf);
    expect(keys[0]).toBe(entry.idempotencyKey);
    expect(keys[1]).toBe(entry.idempotencyKey);
    expect(useOfflineQueueStore.getState().queue).toHaveLength(0);
  });


  it('replays a non-cash queued order through the same create-then-charge path', async () => {
    const payload: CreateOrderRequest = {
      sessionId: '0',
      tableId: '0',
      locationId: '0',
      type: '',
      customerId: '0',
      customerName: '',
      paymentMethod: 'qr',
      notes: 'local-qr',
      kitchenNotes: '',
      items: [],
    };
    useOfflineQueueStore.getState().enqueue(payload, {
      tender: tender({ paymentMethod: 'qr', amountPaid: 31.5 }),
      idempotencyKey: 'k-qr',
    });

    const result = await syncPendingOfflineOrders();

    expect(result).toEqual({ attempted: 1, synced: 1, failed: 0 });
    expect(api.processPayment).toHaveBeenCalledWith(
      { orderId: '501', paymentMethod: 'qr', amountPaid: 31.5 },
      { headers: { 'Idempotency-Key': 'k-qr' } },
    );
  });
});
