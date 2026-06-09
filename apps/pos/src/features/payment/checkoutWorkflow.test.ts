import { describe, expect, it, vi } from 'vitest';
import { buildCashTender, buildCheckoutOrder, processCheckout } from './checkoutWorkflow';
import type { CartItem } from '../../shared/stores/cartStore';

const items: CartItem[] = [
  {
    lineId: '1::',
    id: '1::',
    productId: 1,
    name: 'Burger',
    unitPrice: 10,
    price: 10,
    quantity: 2,
    modifiers: [],
    discountPct: 0,
    discountAmount: 0,
    taxRate: 0,
    status: 'pending',
  },
];

describe('buildCheckoutOrder', () => {
  it('builds a paid local order from the cart snapshot', () => {
    const order = buildCheckoutOrder({
      id: 'local-1',
      items,
      subtotal: 20,
      tax: 1.6,
      total: 21.6,
      employeeId: 7,
      createdAt: '2026-05-06T00:00:00.000Z',
    });

    expect(order).toMatchObject({
      id: 'local-1',
      status: 'paid',
      subtotal: 20,
      tax: 1.6,
      total: 21.6,
      customerId: '7',
    });
    expect(order.items).toEqual([
      { id: '1::', name: 'Burger', price: 10, quantity: 2, notes: '' },
    ]);
  });
});

describe('processCheckout', () => {
  it('creates and charges the order online', async () => {
    const createOrder = vi.fn(() => Promise.resolve({ id: 123 }));
    const chargeOrder = vi.fn(() => Promise.resolve());
    const enqueueOffline = vi.fn();
    const recordCashSale = vi.fn();

    const result = await processCheckout({
      order: buildCheckoutOrder({
        id: 'local-1',
        items,
        subtotal: 20,
        tax: 1.6,
        total: 21.6,
        employeeId: 7,
        createdAt: '2026-05-06T00:00:00.000Z',
      }),
      method: 'cash',
      createOrder,
      chargeOrder,
      enqueueOffline,
      recordCashSale,
    });

    expect(result).toEqual({ status: 'online', orderId: 123 });
    expect(chargeOrder).toHaveBeenCalledWith(123, { payment_method: 'cash', amount_paid: 21.6 });
    expect(enqueueOffline).not.toHaveBeenCalled();
    expect(recordCashSale).toHaveBeenCalledWith(21.6, 'local-1');
  });

  it('queues a cash order when order creation fails before reaching the server', async () => {
    const order = buildCheckoutOrder({
      id: 'local-1',
      items,
      subtotal: 20,
      tax: 1.6,
      total: 21.6,
      employeeId: 7,
      createdAt: '2026-05-06T00:00:00.000Z',
    });
    const enqueueOffline = vi.fn();
    const recordCashSale = vi.fn();

    const result = await processCheckout({
      order,
      method: 'cash',
      createOrder: vi.fn(() => Promise.reject(new Error('Network Error'))),
      chargeOrder: vi.fn(),
      enqueueOffline,
      recordCashSale,
    });

    expect(result).toEqual({ status: 'queued' });
    expect(enqueueOffline).toHaveBeenCalledWith(order);
    expect(recordCashSale).toHaveBeenCalledWith(21.6, 'local-1');
  });

  it('does not queue unverified card payments', async () => {
    await expect(
      processCheckout({
        order: buildCheckoutOrder({
          id: 'local-1',
          items,
          subtotal: 20,
          tax: 1.6,
          total: 21.6,
          employeeId: 7,
          createdAt: '2026-05-06T00:00:00.000Z',
        }),
        method: 'card',
        createOrder: vi.fn(() => Promise.reject(new Error('Network Error'))),
        chargeOrder: vi.fn(),
        enqueueOffline: vi.fn(),
        recordCashSale: vi.fn(),
      })
    ).rejects.toThrow('Payment failed. Please try again.');
  });

  it('fails before charging when the server does not return an order id', async () => {
    const chargeOrder = vi.fn(() => Promise.resolve());
    const recordCashSale = vi.fn();

    await expect(
      processCheckout({
        order: buildCheckoutOrder({
          id: 'local-1',
          items,
          subtotal: 20,
          tax: 1.6,
          total: 21.6,
          employeeId: 7,
          createdAt: '2026-05-06T00:00:00.000Z',
        }),
        method: 'card',
        createOrder: vi.fn(() => Promise.resolve({})),
        chargeOrder,
        enqueueOffline: vi.fn(),
        recordCashSale,
      })
    ).rejects.toThrow('Payment failed. Please try again.');

    expect(chargeOrder).not.toHaveBeenCalled();
    expect(recordCashSale).not.toHaveBeenCalled();
  });
});

describe('buildCashTender', () => {
  it('calculates cash tender change for valid received amount', () => {
    expect(buildCashTender('25.50', 21.6)).toEqual({
      received: 25.5,
      change: 3.9,
      canCharge: true,
    });
  });

  it('rejects malformed or insufficient cash tender', () => {
    expect(buildCashTender('abc', 21.6)).toEqual({
      received: 0,
      change: null,
      canCharge: false,
    });
    expect(buildCashTender('10', 21.6)).toEqual({
      received: 10,
      change: null,
      canCharge: false,
    });
  });
});
