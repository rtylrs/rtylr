import { describe, it, expect } from 'vitest';
import {
  totalPaid,
  totalTips,
  remainingDue,
  isFullyPaid,
  paymentsByMethod,
  paymentMethodCount,
  hasCashPayment,
  hasGiftCardPayment,
  largestPayment,
} from './paymentUtils';
import type { PaymentEntry } from '../types/pos';

const makePayment = (overrides: Partial<PaymentEntry> = {}): PaymentEntry => ({
  id: 'p1',
  method: 'cash',
  amount: 10,
  createdAt: '2024-06-01T10:00:00Z',
  ...overrides,
});

describe('totalPaid', () => {
  it('sums amounts across all payment entries', () => {
    const payments = [makePayment({ amount: 10 }), makePayment({ amount: 5.5 })];
    expect(totalPaid(payments)).toBeCloseTo(15.5);
  });

  it('returns 0 for an empty array', () => {
    expect(totalPaid([])).toBe(0);
  });

  it('ignores non-finite amounts', () => {
    const payments = [makePayment({ amount: NaN }), makePayment({ amount: 5 })];
    expect(totalPaid(payments)).toBe(5);
  });
});

describe('totalTips', () => {
  it('sums tip amounts across entries', () => {
    const payments = [
      makePayment({ tipAmount: 2 }),
      makePayment({ tipAmount: 3 }),
    ];
    expect(totalTips(payments)).toBe(5);
  });

  it('returns 0 when no tips are present', () => {
    expect(totalTips([makePayment()])).toBe(0);
  });
});

describe('remainingDue', () => {
  it('returns the difference when not fully paid', () => {
    expect(remainingDue(100, 60)).toBeCloseTo(40);
  });

  it('returns 0 when fully paid', () => {
    expect(remainingDue(100, 100)).toBe(0);
  });

  it('returns 0 when overpaid', () => {
    expect(remainingDue(100, 110)).toBe(0);
  });

  it('returns 0 for non-finite inputs', () => {
    expect(remainingDue(NaN, 50)).toBe(0);
  });
});

describe('isFullyPaid', () => {
  it('returns true when paid equals order total', () => {
    expect(isFullyPaid(100, 100)).toBe(true);
  });

  it('returns true when paid is within tolerance', () => {
    expect(isFullyPaid(100, 99.9995)).toBe(true);
  });

  it('returns false when underpaid', () => {
    expect(isFullyPaid(100, 99)).toBe(false);
  });

  it('returns false for non-finite inputs', () => {
    expect(isFullyPaid(NaN, 100)).toBe(false);
  });
});

describe('paymentsByMethod', () => {
  it('groups and sums amounts by payment method', () => {
    const payments = [
      makePayment({ method: 'cash', amount: 20 }),
      makePayment({ method: 'card', amount: 30 }),
      makePayment({ method: 'cash', amount: 10 }),
    ];
    const result = paymentsByMethod(payments);
    expect(result.cash).toBe(30);
    expect(result.card).toBe(30);
    expect(result.gift_card).toBe(0);
  });
});

describe('paymentMethodCount', () => {
  it('counts payments by method', () => {
    const payments = [
      makePayment({ method: 'cash' }),
      makePayment({ method: 'card' }),
      makePayment({ method: 'cash' }),
    ];
    const result = paymentMethodCount(payments);
    expect(result.cash).toBe(2);
    expect(result.card).toBe(1);
  });
});

describe('hasCashPayment', () => {
  it('returns true when at least one cash payment exists', () => {
    expect(hasCashPayment([makePayment({ method: 'cash' })])).toBe(true);
  });

  it('returns false when no cash payment exists', () => {
    expect(hasCashPayment([makePayment({ method: 'card' })])).toBe(false);
  });
});

describe('hasGiftCardPayment', () => {
  it('returns true when a gift card payment exists', () => {
    expect(hasGiftCardPayment([makePayment({ method: 'gift_card' })])).toBe(true);
  });

  it('returns false when no gift card payment exists', () => {
    expect(hasGiftCardPayment([makePayment({ method: 'cash' })])).toBe(false);
  });
});

describe('largestPayment', () => {
  it('returns the highest amount in the list', () => {
    const payments = [makePayment({ amount: 10 }), makePayment({ amount: 50 }), makePayment({ amount: 25 })];
    expect(largestPayment(payments)).toBe(50);
  });

  it('returns 0 for an empty array', () => {
    expect(largestPayment([])).toBe(0);
  });
});
