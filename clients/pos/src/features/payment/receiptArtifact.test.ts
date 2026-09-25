import { describe, expect, it } from 'vitest';
import { buildReceiptFromOrder } from './receiptArtifact';
import type { Order } from '../../shared/types/pos';

const order: Order = {
  id: 'local-1',
  status: 'paid',
  items: [{ id: 'line-1', name: 'Burger', price: 10, quantity: 2 }],
  subtotal: 20,
  tax: 1.6,
  tip: 0,
  discount: 0,
  total: 21.6,
  createdAt: '2026-05-06T00:00:00.000Z',
  paidAt: '2026-05-06T00:01:00.000Z',
};

describe('buildReceiptFromOrder', () => {
  it('creates a durable printable receipt for a paid checkout order', () => {
    expect(
      buildReceiptFromOrder({
        order,
        method: 'cash',
        printedAt: '2026-05-06T00:02:00.000Z',
      })
    ).toEqual({
      id: 'receipt-local-1',
      orderId: 'local-1',
      total: 21.6,
      printedAt: '2026-05-06T00:02:00.000Z',
      items: [{ name: 'Burger', qty: 2, price: 10 }],
      payments: [
        {
          id: 'payment-local-1',
          method: 'cash',
          amount: 21.6,
          createdAt: '2026-05-06T00:02:00.000Z',
        },
      ],
    });
  });
});
