/**
 * The TS-side equivalent of the Go tripwire
 * `services/onion/internal/platform/server/http/idempotency_pos_client_proof_test.go`.
 *
 * The Go test builds a raw request and hands it to the middleware, so it only
 * sees the headers a caller chose to put on that request; it cannot see this
 * app's client at all. These tests drive the real generated client through the
 * real `@voxire/api` transport and read the headers off the `RequestInit` that
 * reaches `fetch` - the same thing the Go handler reads off the wire.
 *
 * Drop the `Idempotency-Key` header at either charge call site and these fail.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createOneKitClient } from '@voxire/api';
import { OrderPaymentServiceClient } from '@voxire/voxie/pos/orders/payments/v1/client';
import {
  buildCheckoutOrder,
  buildCreateOrderPayload,
  processCheckout,
  type ChargeOrderRequest,
} from './checkoutWorkflow';
import { finalizeMultiPayment } from './multiPaymentWorkflow';
import { useOfflineQueueStore } from '../../shared/stores/offlineQueueStore';
import { syncPendingOfflineOrders } from '../offline-queue/offlineQueueSync';
import type { CartItem } from '../../shared/stores/cartStore';
import type { PaymentEntry } from '../../shared/types/pos';

const items: CartItem[] = [
  {
    lineId: '1::',
    id: '1::',
    productId: 1,
    name: 'Burger',
    unitPrice: 20,
    price: 20,
    quantity: 1,
    modifiers: [],
    discountPct: 0,
    discountAmount: 0,
    taxRate: 0.08,
    status: 'pending',
  },
];

let requests: { url: string; init: RequestInit }[] = [];
let originalFetch: typeof globalThis.fetch;

beforeEach(() => {
  requests = [];
  originalFetch = globalThis.fetch;
  globalThis.fetch = vi.fn(async (input, init) => {
    requests.push({ url: String(input), init: init ?? {} });
    // The generated clients decode the backend's snake_case wire shape, not the
    // camelCase objects the app sees.
    const body = String(input).endsWith('/payment')
      ? { message: 'Payment processed' }
      : { order_id: '42' };
    return new Response(JSON.stringify(body), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    });
  }) as unknown as typeof globalThis.fetch;
  useOfflineQueueStore.setState({ queue: [], online: true });
});

afterEach(() => {
  globalThis.fetch = originalFetch;
});

/** Exactly how features/payment/index.tsx builds and calls the payment client. */
const send = async (request: ChargeOrderRequest): Promise<void> => {
  const { idempotencyKey, ...payload } = request;
  await createOneKitClient(OrderPaymentServiceClient, 'pos').processPayment(payload, {
    headers: { 'Idempotency-Key': idempotencyKey },
  });
};

const headerOn = (n: number): string | null =>
  new Headers(requests[n]?.init.headers ?? {}).get('Idempotency-Key');

const order = (id: string) =>
  buildCheckoutOrder({
    id,
    items,
    subtotal: 20,
    tax: 1.6,
    total: 21.6,
    createdAt: '2026-05-06T00:00:00.000Z',
  });

describe('the POS payment POST carries Idempotency-Key on the wire', () => {
  it('sends the header the server requires, on the route that requires it', async () => {
    await send({ orderId: '42', paymentMethod: 'card', amountPaid: 21.6, idempotencyKey: 'wire-key-1' });

    expect(requests).toHaveLength(1);
    expect(requests[0].url).toContain('/pos/orders/payments/v1/orders/42/payment');
    // idempotency.go answers 428 Precondition Required for this route without it.
    expect(headerOn(0)).toBe('wire-key-1');
  });

  it('keys two separate sales differently', async () => {
    const seen: string[] = [];

    for (const [index, action] of [order('local-1'), order('local-2')].entries()) {
      await processCheckout({
        order: action,
        payload: buildCreateOrderPayload({ items, method: 'card' }),
        method: 'card',
        createOrder: vi.fn(() => Promise.resolve({ orderId: String(40 + index) })),
        chargeOrder: async (request) => {
          seen.push(request.idempotencyKey);
          await send(request);
        },
        enqueueOffline: vi.fn(),
        recordCashSale: vi.fn(),
      });
    }

    expect(requests).toHaveLength(2);
    expect(headerOn(0)).toBe(seen[0]);
    expect(headerOn(1)).toBe(seen[1]);
    expect(seen[0]).not.toBe(seen[1]);
  });

  it('keys a split-tender Finalize the same way', async () => {
    const payments: PaymentEntry[] = [
      { id: 'p1', method: 'cash', amount: 33, createdAt: '2026-05-06T00:00:00.000Z' },
      { id: 'p2', method: 'card', amount: 22, createdAt: '2026-05-06T00:00:00.000Z' },
    ];

    await finalizeMultiPayment({ orderId: '99', payments, chargeOrder: send });

    expect(requests[0].url).toContain('/pos/orders/payments/v1/orders/99/payment');
    expect(headerOn(0)).toEqual(expect.any(String));
  });

  it('replays a queued offline sale under the key stored on its entry', async () => {
    useOfflineQueueStore.getState().enqueue(
      buildCreateOrderPayload({ items, method: 'cash' }),
      { tender: { paymentMethod: 'cash', amountPaid: 21.6, localOrderId: 'local-off' } },
    );
    const stored = useOfflineQueueStore.getState().queue[0].idempotencyKey;

    const result = await syncPendingOfflineOrders();

    expect(result).toEqual({ attempted: 1, synced: 1, failed: 0 });
    expect(stored).toEqual(expect.any(String));
    const payment = requests.find((r) => r.url.endsWith('/payment'));
    expect(new Headers(payment?.init.headers ?? {}).get('Idempotency-Key')).toBe(stored);
  });
});
