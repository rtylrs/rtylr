import { describe, it, expect } from 'vitest';
import {
  isEmailed,
  hasItems,
  receiptItemCount,
  isBalanced,
  paymentMethodLabels,
  receiptGreeting,
  receiptDateLabel,
} from './receiptUtils';
import type { Receipt, PaymentEntry } from '../types/pos';

const makePayment = (overrides: Partial<PaymentEntry> = {}): PaymentEntry => ({
  id: 'p1',
  method: 'cash',
  amount: 20,
  createdAt: '2024-06-01T10:00:00Z',
  ...overrides,
});

const makeReceipt = (overrides: Partial<Receipt> = {}): Receipt => ({
  id: 'r1',
  orderId: 'o1',
  total: 20,
  printedAt: '2024-06-01T10:00:00Z',
  items: [{ name: 'Burger', qty: 1, price: 10 }, { name: 'Fries', qty: 1, price: 10 }],
  payments: [makePayment({ amount: 20 })],
  ...overrides,
});

describe('isEmailed', () => {
  it('returns true when emailed is true', () => {
    expect(isEmailed(makeReceipt({ emailed: true }))).toBe(true);
  });

  it('returns false when emailed is false', () => {
    expect(isEmailed(makeReceipt({ emailed: false }))).toBe(false);
  });

  it('returns false when emailed is not set', () => {
    expect(isEmailed(makeReceipt())).toBe(false);
  });
});

describe('hasItems', () => {
  it('returns true when items are present', () => {
    expect(hasItems(makeReceipt())).toBe(true);
  });

  it('returns false for an empty items array', () => {
    expect(hasItems(makeReceipt({ items: [] }))).toBe(false);
  });
});

describe('receiptItemCount', () => {
  it('sums the qty across all items', () => {
    const receipt = makeReceipt({
      items: [{ name: 'A', qty: 2, price: 5 }, { name: 'B', qty: 3, price: 5 }],
    });
    expect(receiptItemCount(receipt)).toBe(5);
  });

  it('returns 0 for an empty items array', () => {
    expect(receiptItemCount(makeReceipt({ items: [] }))).toBe(0);
  });
});

describe('isBalanced', () => {
  it('returns true when payments exactly match the total', () => {
    expect(isBalanced(makeReceipt())).toBe(true);
  });

  it('returns true within tolerance', () => {
    const receipt = makeReceipt({ total: 20.005, payments: [makePayment({ amount: 20 })] });
    expect(isBalanced(receipt)).toBe(true);
  });

  it('returns false when payments differ significantly from total', () => {
    const receipt = makeReceipt({ total: 25, payments: [makePayment({ amount: 20 })] });
    expect(isBalanced(receipt)).toBe(false);
  });
});

describe('paymentMethodLabels', () => {
  it('returns unique payment method names in order of first appearance', () => {
    const receipt = makeReceipt({
      payments: [
        makePayment({ method: 'cash' }),
        makePayment({ method: 'card' }),
        makePayment({ method: 'cash' }),
      ],
    });
    expect(paymentMethodLabels(receipt)).toEqual(['cash', 'card']);
  });
});

describe('receiptGreeting', () => {
  it('returns a generic greeting when no name is provided', () => {
    expect(receiptGreeting()).toBe('Thank you!');
    expect(receiptGreeting('')).toBe('Thank you!');
  });

  it('includes the customer name when provided', () => {
    expect(receiptGreeting('Jane')).toBe('Thank you, Jane!');
  });

  it('trims whitespace from the name', () => {
    expect(receiptGreeting('  Jane  ')).toBe('Thank you, Jane!');
  });
});

describe('receiptDateLabel', () => {
  it('returns a formatted date string', () => {
    const label = receiptDateLabel(makeReceipt({ printedAt: '2024-06-01T10:00:00Z' }));
    expect(label).toBeTruthy();
    expect(typeof label).toBe('string');
  });

  it('returns an empty string for an invalid date', () => {
    expect(receiptDateLabel(makeReceipt({ printedAt: 'not-a-date' }))).toBe('');
  });
});
