import { describe, it, expect } from 'vitest';
import {
  orderSubtotal,
  orderTotal,
  isTerminalStatus,
  isActiveStatus,
  statusLabel,
  sortOrdersByDate,
  filterByStatus,
} from './orderUtils';
import type { OrderStatus } from './orderUtils';

describe('orderSubtotal', () => {
  it('sums price times quantity', () => {
    const items = [
      { price: 10, quantity: 2 },
      { price: 5, quantity: 3 },
    ];
    expect(orderSubtotal(items)).toBe(35);
  });

  it('applies item-level discounts', () => {
    const items = [{ price: 100, quantity: 1, discount: 20 }];
    expect(orderSubtotal(items)).toBe(80);
  });

  it('returns 0 for empty order', () => {
    expect(orderSubtotal([])).toBe(0);
  });
});

describe('orderTotal', () => {
  it('adds tax to subtotal', () => {
    expect(orderTotal(100, 10)).toBeCloseTo(110);
  });

  it('handles 0% tax', () => {
    expect(orderTotal(100, 0)).toBeCloseTo(100);
  });

  it('handles fractional tax rates', () => {
    expect(orderTotal(200, 7.5)).toBeCloseTo(215);
  });
});

describe('isTerminalStatus', () => {
  it('returns true for completed', () => {
    expect(isTerminalStatus('completed')).toBe(true);
  });

  it('returns true for cancelled', () => {
    expect(isTerminalStatus('cancelled')).toBe(true);
  });

  it('returns true for refunded', () => {
    expect(isTerminalStatus('refunded')).toBe(true);
  });

  it('returns false for pending', () => {
    expect(isTerminalStatus('pending')).toBe(false);
  });

  it('returns false for processing', () => {
    expect(isTerminalStatus('processing')).toBe(false);
  });
});

describe('isActiveStatus', () => {
  it('returns true for pending', () => {
    expect(isActiveStatus('pending')).toBe(true);
  });

  it('returns true for processing', () => {
    expect(isActiveStatus('processing')).toBe(true);
  });

  it('returns false for completed', () => {
    expect(isActiveStatus('completed')).toBe(false);
  });
});

describe('statusLabel', () => {
  it('returns a human-readable label for each status', () => {
    expect(statusLabel('pending')).toBe('Pending');
    expect(statusLabel('processing')).toBe('Processing');
    expect(statusLabel('completed')).toBe('Completed');
    expect(statusLabel('cancelled')).toBe('Cancelled');
    expect(statusLabel('refunded')).toBe('Refunded');
  });
});

describe('sortOrdersByDate', () => {
  it('sorts newest first', () => {
    const orders = [
      { id: 1, createdAt: '2024-01-01', status: 'pending' as OrderStatus },
      { id: 2, createdAt: '2024-03-01', status: 'pending' as OrderStatus },
      { id: 3, createdAt: '2024-02-01', status: 'pending' as OrderStatus },
    ];
    const sorted = sortOrdersByDate(orders);
    expect(sorted[0].id).toBe(2);
    expect(sorted[2].id).toBe(1);
  });

  it('does not mutate the original array', () => {
    const orders = [{ id: 1, createdAt: '2024-01-01', status: 'pending' as OrderStatus }];
    sortOrdersByDate(orders);
    expect(orders[0].id).toBe(1);
  });
});

describe('filterByStatus', () => {
  const orders = [
    { id: 1, status: 'pending' as OrderStatus },
    { id: 2, status: 'completed' as OrderStatus },
    { id: 3, status: 'pending' as OrderStatus },
  ];

  it('filters by the given status', () => {
    const pending = filterByStatus(orders, 'pending');
    expect(pending).toHaveLength(2);
    expect(pending.every((o) => o.status === 'pending')).toBe(true);
  });

  it('returns empty array when no orders match', () => {
    expect(filterByStatus(orders, 'cancelled')).toHaveLength(0);
  });
});
