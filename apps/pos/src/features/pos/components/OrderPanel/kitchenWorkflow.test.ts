import { describe, expect, it } from 'vitest';
import {
  KITCHEN_ORDER_ID_ERROR,
  buildKitchenOrderPayload,
  extractCreatedOrderId,
} from './kitchenWorkflow';
import type { CartItem } from '../../../../shared/stores/cartStore';

const item = (overrides: Partial<CartItem> = {}): CartItem => ({
  id: '7::',
  lineId: '7::',
  productId: 7,
  name: 'Burger',
  unitPrice: 10,
  price: 10,
  quantity: 2,
  notes: 'No onions',
  modifiers: [],
  discountPct: 0,
  discountAmount: 0,
  taxRate: 0,
  status: 'pending',
  ...overrides,
});

describe('kitchenWorkflow', () => {
  it('builds a create-order payload for sending a local cart to kitchen', () => {
    expect(
      buildKitchenOrderPayload({
        items: [item()],
        subtotal: 20,
        tax: 2,
        total: 22,
        employeeId: 5,
        now: () => '2026-01-01T10:00:00.000Z',
        localId: () => 'local-1',
      }),
    ).toEqual({
      id: 'local-1',
      status: 'open',
      items: [{ id: '7::', name: 'Burger', price: 10, quantity: 2, notes: 'No onions' }],
      subtotal: 20,
      tax: 2,
      tip: 0,
      discount: 0,
      total: 22,
      createdAt: '2026-01-01T10:00:00.000Z',
      customerId: '5',
    });
  });

  it('extracts positive integer order IDs from create responses', () => {
    expect(extractCreatedOrderId({ id: 10 })).toBe(10);
    expect(extractCreatedOrderId({ order_id: '11' })).toBe(11);
  });

  it('rejects create responses without a valid server order ID', () => {
    expect(() => extractCreatedOrderId({ id: 'local-1' })).toThrow(KITCHEN_ORDER_ID_ERROR);
  });
});
