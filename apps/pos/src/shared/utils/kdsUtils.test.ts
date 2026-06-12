import { describe, it, expect } from 'vitest';
import {
  isKdsOrderActive,
  isKdsOrderComplete,
  ordersByStation,
  nextKdsStatus,
  totalKdsItemCount,
  sortKdsOrdersOldestFirst,
  uniqueStations,
} from './kdsUtils';
import type { KDSOrderRaw } from './schemas/kdsOrders';

const makeOrder = (overrides: Partial<KDSOrderRaw> = {}): KDSOrderRaw => ({
  id: '1',
  station: 'Kitchen',
  items: [{ name: 'Burger', qty: 2 }],
  status: 'new',
  createdAt: '2024-06-01T10:00:00Z',
  ...overrides,
});

describe('isKdsOrderActive', () => {
  it('returns true for "new" status', () => {
    expect(isKdsOrderActive(makeOrder({ status: 'new' }))).toBe(true);
  });

  it('returns true for "cooking" status', () => {
    expect(isKdsOrderActive(makeOrder({ status: 'cooking' }))).toBe(true);
  });

  it('returns false for "ready" status', () => {
    expect(isKdsOrderActive(makeOrder({ status: 'ready' }))).toBe(false);
  });

  it('returns false for "served" status', () => {
    expect(isKdsOrderActive(makeOrder({ status: 'served' }))).toBe(false);
  });
});

describe('isKdsOrderComplete', () => {
  it('returns true for "ready" status', () => {
    expect(isKdsOrderComplete(makeOrder({ status: 'ready' }))).toBe(true);
  });

  it('returns true for "served" status', () => {
    expect(isKdsOrderComplete(makeOrder({ status: 'served' }))).toBe(true);
  });

  it('returns false for "new" status', () => {
    expect(isKdsOrderComplete(makeOrder({ status: 'new' }))).toBe(false);
  });
});

describe('ordersByStation', () => {
  const orders = [
    makeOrder({ id: '1', station: 'Kitchen' }),
    makeOrder({ id: '2', station: 'Bar' }),
    makeOrder({ id: '3', station: 'kitchen' }),
  ];

  it('filters orders by station name (case-insensitive)', () => {
    const result = ordersByStation(orders, 'Kitchen');
    expect(result).toHaveLength(2);
  });

  it('returns an empty array when no orders match', () => {
    expect(ordersByStation(orders, 'Grill')).toHaveLength(0);
  });
});

describe('nextKdsStatus', () => {
  it('advances new -> cooking', () => {
    expect(nextKdsStatus('new')).toBe('cooking');
  });

  it('advances cooking -> ready', () => {
    expect(nextKdsStatus('cooking')).toBe('ready');
  });

  it('advances ready -> served', () => {
    expect(nextKdsStatus('ready')).toBe('served');
  });

  it('stays at served when already served', () => {
    expect(nextKdsStatus('served')).toBe('served');
  });
});

describe('totalKdsItemCount', () => {
  it('sums item quantities across all orders', () => {
    const orders = [
      makeOrder({ items: [{ name: 'Burger', qty: 2 }] }),
      makeOrder({ items: [{ name: 'Fries', qty: 3 }, { name: 'Cola', qty: 1 }] }),
    ];
    expect(totalKdsItemCount(orders)).toBe(6);
  });

  it('returns 0 for an empty list', () => {
    expect(totalKdsItemCount([])).toBe(0);
  });
});

describe('sortKdsOrdersOldestFirst', () => {
  it('sorts orders from oldest to newest', () => {
    const orders = [
      makeOrder({ id: '2', createdAt: '2024-06-01T10:05:00Z' }),
      makeOrder({ id: '1', createdAt: '2024-06-01T10:00:00Z' }),
    ];
    const sorted = sortKdsOrdersOldestFirst(orders);
    expect(sorted[0].id).toBe('1');
    expect(sorted[1].id).toBe('2');
  });

  it('does not mutate the original array', () => {
    const orders = [
      makeOrder({ id: '2', createdAt: '2024-06-01T10:05:00Z' }),
      makeOrder({ id: '1', createdAt: '2024-06-01T10:00:00Z' }),
    ];
    sortKdsOrdersOldestFirst(orders);
    expect(orders[0].id).toBe('2');
  });
});

describe('uniqueStations', () => {
  it('returns unique station names sorted alphabetically', () => {
    const orders = [
      makeOrder({ station: 'Bar' }),
      makeOrder({ station: 'Kitchen' }),
      makeOrder({ station: 'Bar' }),
    ];
    expect(uniqueStations(orders)).toEqual(['Bar', 'Kitchen']);
  });

  it('returns an empty array for an empty list', () => {
    expect(uniqueStations([])).toEqual([]);
  });
});
