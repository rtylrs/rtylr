import { describe, expect, it } from 'vitest';
import {
  ORDER_SUBMIT_ID_ERROR,
  buildOrderSubmitPayload,
  extractCreatedOrderId,
} from './orderSubmitWorkflow';
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

describe('orderSubmitWorkflow', () => {
  it('builds a CreateOrderPayload for submitting a local cart to Orders', () => {
    expect(
      buildOrderSubmitPayload({
        items: [item()],
        locationId: 2,
        tableId: 9,
        customerId: 5,
        orderType: 'takeaway',
        notes: 'Pickup shelf',
      }),
    ).toEqual({
      location_id: 2,
      table_id: 9,
      customer_id: 5,
      customer_name: undefined,
      type: 'takeaway',
      notes: 'Pickup shelf',
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
    const result = buildOrderSubmitPayload({ items: [item()] });
    expect(result.location_id).toBeUndefined();
    expect(result.customer_id).toBeUndefined();
    expect(result.items).toHaveLength(1);
  });

  it('extracts positive integer order IDs from create responses', () => {
    expect(extractCreatedOrderId({ id: 10 })).toBe(10);
    expect(extractCreatedOrderId({ order_id: '11' })).toBe(11);
  });

  it('rejects create responses without a valid server order ID', () => {
    expect(() => extractCreatedOrderId({ id: 'local-1' })).toThrow(ORDER_SUBMIT_ID_ERROR);
  });
});
