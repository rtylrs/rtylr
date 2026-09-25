/**
 * End-to-end: checkout → loyalty points earn → redeem discount.
 *
 * Proves that the loyalty flow chains correctly with the payment workflow.
 */
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { buildCheckoutOrder, buildCreateOrderPayload, processCheckout } from '../payment/checkoutWorkflow';
import { earnPointsAfterPayment, redeemPointsForDiscount } from './loyaltyWorkflow';
import type { CartItem } from '../../shared/stores/cartStore';

const loyaltyClient = vi.hoisted(() => ({
  earnPoints: vi.fn(),
  redeemPoints: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => loyaltyClient,
}));

const item: CartItem = {
  lineId: '1::',
  id: '1::',
  productId: 1,
  name: 'Latte',
  unitPrice: 6.0,
  price: 6.0,
  quantity: 3,
  modifiers: [],
  discountPct: 0,
  discountAmount: 0,
  taxRate: 0,
  status: 'pending',
  notes: '',
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe('checkout → earn loyalty points', () => {
  it('earns points proportional to order total after successful online payment', async () => {
    vi.mocked(loyaltyClient.earnPoints).mockResolvedValue(undefined as never);

    const order = buildCheckoutOrder({
      id: 'local-1',
      items: [item],
      subtotal: 18,
      tax: 1.44,
      total: 19.44,
      createdAt: '2026-05-13T10:00:00.000Z',
    });
    const payload = buildCreateOrderPayload({ items: [item], method: 'cash' });

    const checkoutResult = await processCheckout({
      order,
      payload,
      method: 'cash',
      createOrder: vi.fn(() => Promise.resolve({ orderId: '1001', orderNumber: 'ORD-1001', message: 'ok' })),
      chargeOrder: vi.fn(() => Promise.resolve()),
      enqueueOffline: vi.fn(),
      recordCashSale: vi.fn(),
    });

    expect(checkoutResult).toEqual({ status: 'online', orderId: 1001 });

    // Customer ID 7 earns points after payment
    const earnResult = await earnPointsAfterPayment({
      customerId: 7,
      orderId: (checkoutResult as { orderId: number }).orderId,
      totalAmount: order.total,
    });

    expect(earnResult).toEqual({ earned: 19, newTotal: null }); // floor(19.44 * 1)
    expect(loyaltyClient.earnPoints).toHaveBeenCalledWith({
      customerId: '7',
      points: 19,
      orderId: '1001',
    });
  });

  it('skips earning when no customer is attached to the order', async () => {
    const order = buildCheckoutOrder({
      id: 'local-2',
      items: [item],
      subtotal: 18,
      tax: 1.44,
      total: 19.44,
      createdAt: '2026-05-13T10:00:00.000Z',
    });
    const payload = buildCreateOrderPayload({ items: [item], method: 'card' });

    await processCheckout({
      order,
      payload,
      method: 'card',
      createOrder: vi.fn(() => Promise.resolve({ orderId: '1002', orderNumber: 'ORD-1002', message: 'ok' })),
      chargeOrder: vi.fn(() => Promise.resolve()),
      enqueueOffline: vi.fn(),
      recordCashSale: vi.fn(),
    });

    const earnResult = await earnPointsAfterPayment({
      customerId: null,
      orderId: 1002,
      totalAmount: order.total,
    });

    expect(earnResult).toBeNull();
    expect(loyaltyClient.earnPoints).not.toHaveBeenCalled();
  });
});

describe('redeem points → discount applied before checkout', () => {
  it('redeems points and returns correct dollar discount', async () => {
    vi.mocked(loyaltyClient.redeemPoints).mockResolvedValue(undefined as never);

    // Customer has 500 points ($5.00 discount at $0.01/point)
    const redeemResult = await redeemPointsForDiscount({
      customerId: 7,
      orderId: 2001,
      pointsToRedeem: 500,
    });

    expect(redeemResult).toEqual({ pointsUsed: 500, discountAmount: 5.0 });
    expect(loyaltyClient.redeemPoints).toHaveBeenCalledWith({
      customerId: '7',
      points: 500,
    });
  });

  it('caps redemption so discount does not exceed order total', async () => {
    // Order total $8, customer has 2000 points ($20 value) - capped at 800 points
    const { maxRedeemablePoints } = await import('./loyaltyWorkflow');
    const capped = maxRedeemablePoints(2000, 8.0);
    expect(capped).toBe(800); // floor(8.0 / 0.01)
  });

  it('chains redeem → checkout with reduced total', async () => {
    vi.mocked(loyaltyClient.redeemPoints).mockResolvedValue(undefined as never);

    const redeemResult = await redeemPointsForDiscount({
      customerId: 7,
      orderId: 3001,
      pointsToRedeem: 200,
    });

    // $2 discount applied before building order
    const discountedTotal = 19.44 - (redeemResult?.discountAmount ?? 0);

    const order = buildCheckoutOrder({
      id: 'local-3',
      items: [item],
      subtotal: 18,
      tax: 1.44,
      total: discountedTotal,
      createdAt: '2026-05-13T10:00:00.000Z',
    });
    const payload = buildCreateOrderPayload({ items: [item], method: 'cash' });

    const checkoutResult = await processCheckout({
      order,
      payload,
      method: 'cash',
      createOrder: vi.fn(() => Promise.resolve({ orderId: '3001', orderNumber: 'ORD-3001', message: 'ok' })),
      chargeOrder: vi.fn(() => Promise.resolve()),
      enqueueOffline: vi.fn(),
      recordCashSale: vi.fn(),
    });

    expect(checkoutResult).toEqual({ status: 'online', orderId: 3001 });
    expect(order.total).toBeCloseTo(17.44, 2);
  });
});
