import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { OrderServiceClient } from '@voxire/voxie/pos/orders/v1/client';
import { OrderPaymentServiceClient } from '@voxire/voxie/pos/orders/payments/v1/client';
import type { CreateOrderRequest } from '@voxire/voxie/pos/orders/v1/types';
import { useOfflineQueueStore } from '../../shared/stores/offlineQueueStore';
import type { OfflineQueueEntryMeta } from '../../shared/stores/offlineQueueStore';
import type { OfflineQueuedOrder } from '../../shared/types/pos';
import { syncQueuedOrders, type SyncOrder } from './syncOfflineQueue';
import { runExclusiveDrain } from './drainGuard';

const orderClient = createOneKitClient(OrderServiceClient, 'pos');
const paymentClient = createOneKitClient(OrderPaymentServiceClient, 'pos');

/**
 * ProcessPayment answers 409 "Order is already paid" for an order it has
 * already settled. Re-issuing the charge is the expected way to recover from a
 * crash between the charge succeeding and the queue entry being marked synced,
 * so this is a success, not a failure.
 */
function isAlreadyPaid(error: unknown): boolean {
  const statusCode = (error as { statusCode?: unknown } | null | undefined)?.statusCode;
  if (statusCode !== 409) return false;
  return getOneKitErrorMessage(error, '').toLowerCase().includes('already paid');
}

/**
 * Total the server will compute from the same line items, used for entries
 * persisted before the tender moved onto the queue entry.
 *
 * Mirrors ComputeTotals in services/onion/internal/modules/pos_orders/internal/
 * domain/order_totals.go operation for operation, because the two have to agree
 * to the cent: the server compares this figure against its own total with no
 * tolerance (ComputeChange rejects amountPaid < total), so a cent out in either
 * direction either strands the sale as "payment insufficient" or overstates the
 * change due. The algebraically equivalent `x * (1 - p)` / `x * (1 + t)` form
 * this replaces rounded differently from the server's
 * `subtotal - subtotal * p` / `subtotal + tax`, which is what let the two drift
 * across a half-cent boundary. The server's flat-discount term is left out
 * because CreateOrderItemInput carries no discount_amount, so that branch is
 * unreachable for this payload.
 *
 * The server persists the total as DECIMAL via fmt.Sprintf("%.2f"), so the
 * figure is converted to a 2-decimal number the same way rather than by
 * Math.round(x * 100), whose double rounding can land a cent away.
 */
function payloadOrderTotal(order: CreateOrderRequest): number {
  let subtotal = 0;
  let taxAmount = 0;
  for (const item of order.items ?? []) {
    const lineSubtotal = (item.unitPrice ?? 0) * (item.quantity ?? 0);
    const lineAfterPercentage = lineSubtotal - lineSubtotal * ((item.discountPct ?? 0) / 100);
    subtotal += lineAfterPercentage;
    taxAmount += lineAfterPercentage * ((item.taxRate ?? 0) / 100);
  }
  return Number((subtotal + taxAmount).toFixed(2));
}

/**
 * The money and the key this entry will be charged under, written back before
 * the first send. An entry persisted by an older build has neither, and minting
 * a fresh key on every drain would re-charge a sale the first attempt already
 * committed; persisting the replacement makes the next attempt reuse it.
 */
function replayStateFor(entry: OfflineQueuedOrder): OfflineQueueEntryMeta {
  const tender = entry.tender;
  const complete = { tender, idempotencyKey: entry.idempotencyKey };
  if (complete.idempotencyKey && tender && tender.amountPaid > 0 && tender.localOrderId) {
    return { tender, idempotencyKey: complete.idempotencyKey };
  }
  const repaired: OfflineQueueEntryMeta = {
    tender:
      tender && tender.amountPaid > 0 && tender.localOrderId
        ? tender
        : {
            // Nothing recorded what was tendered, so the best available figure
            // is the total the server itself will compute from the lines.
            paymentMethod: tender?.paymentMethod ?? entry.order.paymentMethod ?? 'cash',
            amountPaid: payloadOrderTotal(entry.order),
            localOrderId: entry.id,
          },
    idempotencyKey: complete.idempotencyKey || crypto.randomUUID(),
  };
  useOfflineQueueStore.getState().ensureReplayable(entry.id, repaired);
  return repaired;
}

/**
 * Replays one queued order: create it, then charge it.
 *
 * The charge is the load-bearing half. `CreateOrder` only writes a pending /
 * unpaid row, and every money surface (day summary, session totals, receipts,
 * dashboard) filters on `payment_status = 'paid'`; `deductSoldStock` only runs
 * inside ProcessPayment. Creating without charging therefore records a sale
 * that took cash but is worth nothing and consumed no stock.
 */
export const replayQueuedOrder: SyncOrder = async (entry) => {
  const live = useOfflineQueueStore.getState().queue.find((q) => q.id === entry.id) ?? entry;
  const { tender, idempotencyKey } = replayStateFor(live);

  let orderId = live.serverOrderId;
  if (!orderId) {
    const created = await orderClient.createOrder(entry.order);
    const createdId = String(created?.orderId ?? '');
    if (!/^[1-9][0-9]*$/.test(createdId)) {
      throw new Error('Create order did not return an order id');
    }
    // Persisted before the charge. A crash from here on replays this charge,
    // which MarkOrderPaid makes a no-op, instead of creating a second sale.
    useOfflineQueueStore.getState().recordServerOrderId(entry.id, createdId);
    orderId = createdId;
  }

  try {
    await paymentClient.processPayment(
      {
        orderId,
        paymentMethod: tender.paymentMethod,
        amountPaid: tender.amountPaid,
      },
      // The key the sale was queued under, hours after the tap if the till was
      // offline. Reusing it is what makes this the same action rather than a
      // second one; the server rejects the POST with 428 if it is missing.
      { headers: { 'Idempotency-Key': idempotencyKey } }
    );
  } catch (error) {
    if (!isAlreadyPaid(error)) throw error;
  }
};

export async function syncPendingOfflineOrders(
  syncOrder: SyncOrder = replayQueuedOrder
): Promise<{ attempted: number; synced: number; failed: number }> {
  return runExclusiveDrain('orders', () => {
    const store = useOfflineQueueStore.getState();
    // Recover entries stranded in 'syncing' by a crash before filtering.
    store.requeueStalled();
    const queue = useOfflineQueueStore.getState().queue;
    const pending = queue.filter((entry) => entry.status === 'pending');

    return syncQueuedOrders({
      entries: pending,
      syncOrder,
      retry: store.retry,
      markSynced: store.markSynced,
      markFailed: store.markFailed,
      readEntry: (id) => useOfflineQueueStore.getState().queue.find((q) => q.id === id),
    });
  });
}
