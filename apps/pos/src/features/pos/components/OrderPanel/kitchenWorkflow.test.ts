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
  it('builds a CreateOrderPayload for sending a local cart to kitchen', () => {
    expect(
      buildKitchenOrderPayload({
        items: [item()],
        locationId: 2,
        customerId: 5,
        kitchenNotes: 'Seat 3',
      }),
    ).toEqual({
      location_id: 2,
      table_id: undefined,
      customer_id: 5,
      customer_name: undefined,
      kitchen_notes: 'Seat 3',
      items: [{
        menu_item_id: 7,
        name: 'Burger',
        quantity: 2,
        unit_price: 10,
        discount_pct: 0,
        tax_rate: 0,
        notes: 'No onions',
      }],
    });
  });

  it('omits undefined optional fields gracefully', () => {
    const result = buildKitchenOrderPayload({ items: [item()] });
    expect(result.location_id).toBeUndefined();
    expect(result.customer_id).toBeUndefined();
    expect(result.items).toHaveLength(1);
  });

  it('extracts positive integer order IDs from create responses', () => {
    expect(extractCreatedOrderId({ id: 10 })).toBe(10);
    expect(extractCreatedOrderId({ order_id: '11' })).toBe(11);
  });

  it('rejects create responses without a valid server order ID', () => {
    expect(() => extractCreatedOrderId({ id: 'local-1' })).toThrow(KITCHEN_ORDER_ID_ERROR);
  });
});
