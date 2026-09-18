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
  it('builds a generated CreateOrderRequest for submitting a local cart to Orders', () => {
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
      sessionId: '0',
      locationId: '2',
      tableId: '9',
      customerId: '5',
      customerName: '',
      type: 'takeaway',
      paymentMethod: '',
      notes: 'Pickup shelf',
      kitchenNotes: '',
      items: [{
        menuItemId: '7',
        name: 'Burger',
        sku: '',
        quantity: 2,
        unitPrice: 10,
        discountPct: 0,
        taxRate: 0,
        notes: 'No onions',
      }],
    });
  });

  it('omits undefined optional fields gracefully', () => {
    const result = buildOrderSubmitPayload({ items: [item()] });
    expect(result.locationId).toBe('0');
    expect(result.customerId).toBe('0');
    expect(result.items).toHaveLength(1);
  });

  it('extracts positive integer order IDs from create responses', () => {
    expect(extractCreatedOrderId({ orderId: '10' })).toBe(10);
    expect(extractCreatedOrderId({ orderId: '11' })).toBe(11);
  });

  it('rejects create responses without a valid server order ID', () => {
    expect(() => extractCreatedOrderId({ orderId: 'local-1' })).toThrow(ORDER_SUBMIT_ID_ERROR);
  });
});
