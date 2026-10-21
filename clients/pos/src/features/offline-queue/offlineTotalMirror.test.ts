/**
 * The tender a pre-upgrade queue entry never recorded is reconstructed from the
 * order payload, and the server then compares that figure against its own total
 * with no tolerance: ComputeChange rejects `amountPaid < total`. A cent low
 * strands the sale as "payment insufficient"; a cent high overstates the change
 * due. So the reconstruction has to agree with
 * services/onion/internal/modules/pos_orders/internal/domain.ComputeTotals to
 * the cent.
 *
 * The expected totals below were produced by running the Go implementation
 * (ComputeTotals, then fmt.Sprintf("%.2f") as FloatToDecimal persists it), not by
 * running this arithmetic. They are the contract: if the server's totals change,
 * these fail rather than letting the two drift apart silently.
 */
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { syncPendingOfflineOrders } from './offlineQueueSync';
import { useOfflineQueueStore } from '../../shared/stores/offlineQueueStore';
import type { CreateOrderRequest } from '@voxire/voxie/pos/orders/v1/types';

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

type Item = { unitPrice: number; quantity: number; discountPct?: number; taxRate?: number };

/** Totals generated from the Go ComputeTotals, keyed by fixture name. */
const SERVER_TOTAL: Record<string, number> = {
  'single cash line': 16.2,
  'percentage discount and tax': 220.15,
  'half-cent boundary': 0.14,
  'repeating decimal price': 92.3,
  'many sub-cent lines': 2.16,
  'odd fractional quantity': 47.17,
  'wide price and rate spread': 1276.49,
};

const FIXTURES: Array<{ name: string; items: Item[] }> = [
  { name: 'single cash line', items: [{ unitPrice: 16.2, quantity: 1 }] },
  {
    name: 'percentage discount and tax',
    items: [{ unitPrice: 100, quantity: 2, discountPct: 7.5, taxRate: 19 }],
  },
  { name: 'half-cent boundary', items: [{ unitPrice: 0.145, quantity: 1 }] },
  {
    name: 'repeating decimal price',
    items: [{ unitPrice: 33.33, quantity: 3, discountPct: 12.5, taxRate: 5.5 }],
  },
  {
    name: 'many sub-cent lines',
    items: [
      { unitPrice: 0.1, quantity: 3, taxRate: 20 },
      { unitPrice: 0.2, quantity: 3, taxRate: 20 },
      { unitPrice: 0.3, quantity: 3, taxRate: 20 },
    ],
  },
  {
    name: 'odd fractional quantity',
    items: [
      { unitPrice: 12.345, quantity: 3, discountPct: 3.33, taxRate: 8.25 },
      { unitPrice: 7.77, quantity: 1, taxRate: 8.25 },
    ],
  },
  {
    name: 'wide price and rate spread',
    items: [
      { unitPrice: 199.99, quantity: 7, discountPct: 15.5, taxRate: 7.25 },
      { unitPrice: 0.99, quantity: 11, discountPct: 33.3, taxRate: 7.25 },
    ],
  },
];

function payloadWith(items: Item[]): CreateOrderRequest {
  return {
    sessionId: '0',
    tableId: '0',
    locationId: '0',
    type: '',
    customerId: '0',
    customerName: '',
    paymentMethod: 'cash',
    notes: 'legacy-entry',
    kitchenNotes: '',
    items: items.map((item) => ({
      unitPrice: item.unitPrice,
      quantity: item.quantity,
      discountPct: item.discountPct ?? 0,
      taxRate: item.taxRate ?? 0,
    })),
  };
}

beforeEach(() => {
  useOfflineQueueStore.setState({ queue: [], online: true });
  api.createOrder.mockReset();
  api.processPayment.mockReset();
  api.createOrder.mockResolvedValue({ orderId: '501' });
  api.processPayment.mockResolvedValue({ message: 'Payment processed' });
});

describe('reconstructed tender for a pre-upgrade queue entry', () => {
  it.each(FIXTURES)('matches the server total for $name', async ({ name, items }) => {
    // No meta: the placeholder tender is incomplete, which is the state an
    // entry persisted before the tender moved onto the queue entry replays from.
    useOfflineQueueStore.getState().enqueue(payloadWith(items));

    const result = await syncPendingOfflineOrders();

    expect(result).toEqual({ attempted: 1, synced: 1, failed: 0 });
    expect(api.processPayment).toHaveBeenCalledWith(
      expect.objectContaining({ amountPaid: SERVER_TOTAL[name] }),
      expect.anything()
    );
  });

  it('never charges less than the server total, so the charge is not rejected', async () => {
    for (const { name, items } of FIXTURES) {
      useOfflineQueueStore.setState({ queue: [], online: true });
      api.processPayment.mockClear();
      useOfflineQueueStore.getState().enqueue(payloadWith(items));

      await syncPendingOfflineOrders();

      const charged = api.processPayment.mock.calls[0]?.[0] as { amountPaid: number };
      expect(charged.amountPaid).toBeGreaterThanOrEqual(SERVER_TOTAL[name]);
    }
  });
});
