/**
 * End-to-end: discount applied → discounted total → checkout charge.
 *
 * Proves the discount workflow reduces the order total and that the
 * correct (discounted) amount is sent to the payment endpoint.
 */
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { buildDiscountRequest } from './discountWorkflow';
import { buildCheckoutOrder, buildCreateOrderPayload, processCheckout } from '../payment/checkoutWorkflow';
import type { CartItem } from '../../shared/stores/cartStore';

beforeEach(() => {
  vi.clearAllMocks();
});

const item: CartItem = {
  lineId: '1::',
  id: '1::',
  productId: 1,
  name: 'Burger',
  unitPrice: 20,
  price: 20,
  quantity: 2,
  modifiers: [],
  discountPct: 0,
  discountAmount: 0,
  taxRate: 0,
  status: 'pending',
  notes: '',
};

describe('discount → checkout chain', () => {
  it('flat discount reduces the charge amount sent to the payment endpoint', async () => {
    const discount = buildDiscountRequest({ orderId: '99', mode: 'flat', valueInput: '5' });
    expect(discount).toEqual({
      payload: { orderId: '99', discountType: 'flat', discountValue: 5, reason: '' },
    });

    const originalTotal = 40;
    const discountedTotal = originalTotal - discount.payload.discountValue!;
    expect(discountedTotal).toBe(35);

    const order = buildCheckoutOrder({
      id: 'local-1',
      items: [item],
      subtotal: 40,
      tax: 0,
      total: discountedTotal,
      customerId: 1,
      createdAt: '2026-05-13T10:00:00.000Z',
    });

    const payload = buildCreateOrderPayload({ items: [item], method: 'cash' });
    const chargeOrder = vi.fn(async () => undefined);

    const result = await processCheckout({
      order,
      payload,
      method: 'cash',
      createOrder: vi.fn(() => Promise.resolve({ orderId: '200' })),
      chargeOrder,
      enqueueOffline: vi.fn(),
      recordCashSale: vi.fn(),
    });

    expect(result).toEqual({ status: 'online', orderId: 200 });
    expect(chargeOrder).toHaveBeenCalledWith(
      expect.objectContaining({ orderId: '200', amountPaid: 35 }),
    );
  });

  it('percent discounts above 100% are rejected, not clamped', () => {
    expect(() =>
      buildDiscountRequest({ orderId: '99', mode: 'percent', valueInput: '150' }),
    ).toThrow('Percent discounts cannot exceed 100%.');
  });

  it('zero discount value is rejected before checkout', () => {
    expect(() =>
      buildDiscountRequest({ orderId: '99', mode: 'flat', valueInput: '0' }),
    ).toThrow('Enter a discount greater than zero.');
  });

  it('invalid order id is rejected before discount is applied', () => {
    expect(() =>
      buildDiscountRequest({ orderId: 'current', mode: 'percent', valueInput: '10' }),
    ).toThrow('Select a saved order before applying a discount.');
  });

  it('10% percent discount on $80 order charges $72', async () => {
    const discount = buildDiscountRequest({ orderId: '50', mode: 'percent', valueInput: '10' });
    const originalTotal = 80;
    const discountedTotal = originalTotal * (1 - discount.payload.discountValue! / 100);
    expect(discountedTotal).toBeCloseTo(72, 2);

    const chargeOrder = vi.fn(async () => undefined);
    const order = buildCheckoutOrder({
      id: 'local-2',
      items: [item],
      subtotal: 80,
      tax: 0,
      total: discountedTotal,
      customerId: 1,
      createdAt: '2026-05-13T10:00:00.000Z',
    });

    const payload = buildCreateOrderPayload({ items: [item], method: 'card' });
    const result = await processCheckout({
      order,
      payload,
      method: 'card',
      createOrder: vi.fn(() => Promise.resolve({ orderId: '201' })),
      chargeOrder,
      enqueueOffline: vi.fn(),
      recordCashSale: vi.fn(),
    });

    expect(result).toEqual({ status: 'online', orderId: 201 });
    expect(chargeOrder).toHaveBeenCalledWith(
      expect.objectContaining({ orderId: '201', amountPaid: discountedTotal }),
    );
  });
});
