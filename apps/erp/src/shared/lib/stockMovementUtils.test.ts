import { describe, it, expect } from 'vitest';
import {
  netQuantity,
  summarizeMovements,
  movementsForProduct,
  movementsOfType,
  movementsInDateRange,
  sortByDateDesc,
  recentMovements,
  runningBalance,
} from './stockMovementUtils';
import type { StockMovement } from './stockMovementUtils';

const makeMovement = (
  type: StockMovement['type'],
  quantity: number,
  overrides: Partial<StockMovement> = {}
): StockMovement => ({
  id: Math.random(),
  productId: 1,
  productName: 'Widget',
  sku: 'WGT-001',
  type,
  quantity,
  date: '2026-05-15',
  ...overrides,
});

describe('netQuantity', () => {
  it('returns positive quantity for "in" movements', () => {
    expect(netQuantity(makeMovement('in', 10))).toBe(10);
  });

  it('returns negative quantity for "out" movements', () => {
    expect(netQuantity(makeMovement('out', 5))).toBe(-5);
  });

  it('returns quantity for adjustments', () => {
    expect(netQuantity(makeMovement('adjustment', -3))).toBe(-3);
  });

  it('returns 0 for transfers', () => {
    expect(netQuantity(makeMovement('transfer', 10))).toBe(0);
  });
});

describe('summarizeMovements', () => {
  it('correctly totals each movement type', () => {
    const movements = [
      makeMovement('in', 100),
      makeMovement('out', 30),
      makeMovement('adjustment', -5),
      makeMovement('transfer', 20),
    ];
    const summary = summarizeMovements(movements);
    expect(summary.totalIn).toBe(100);
    expect(summary.totalOut).toBe(30);
    expect(summary.adjustments).toBe(-5);
    expect(summary.transfers).toBe(20);
    expect(summary.netChange).toBe(65); // 100 - 30 + (-5)
  });
});

describe('movementsForProduct', () => {
  it('filters by product id', () => {
    const movements = [
      makeMovement('in', 10, { productId: 1 }),
      makeMovement('in', 5, { productId: 2 }),
    ];
    const result = movementsForProduct(movements, 1);
    expect(result).toHaveLength(1);
    expect(result[0].productId).toBe(1);
  });
});

describe('movementsOfType', () => {
  it('filters by movement type', () => {
    const movements = [makeMovement('in', 10), makeMovement('out', 5), makeMovement('in', 3)];
    const ins = movementsOfType(movements, 'in');
    expect(ins).toHaveLength(2);
  });
});

describe('movementsInDateRange', () => {
  it('returns movements within date range', () => {
    const movements = [
      makeMovement('in', 5, { date: '2026-05-01' }),
      makeMovement('in', 5, { date: '2026-05-15' }),
      makeMovement('in', 5, { date: '2026-06-01' }),
    ];
    const result = movementsInDateRange(movements, '2026-05-10', '2026-05-20');
    expect(result).toHaveLength(1);
    expect(result[0].date).toBe('2026-05-15');
  });
});

describe('sortByDateDesc', () => {
  it('sorts movements by date descending', () => {
    const movements = [
      makeMovement('in', 5, { date: '2026-05-01' }),
      makeMovement('in', 5, { date: '2026-05-20' }),
      makeMovement('in', 5, { date: '2026-05-10' }),
    ];
    const sorted = sortByDateDesc(movements);
    expect(sorted[0].date).toBe('2026-05-20');
    expect(sorted[2].date).toBe('2026-05-01');
  });
});

describe('recentMovements', () => {
  it('returns top N most recent movements', () => {
    const movements = Array.from({ length: 15 }, (_, i) =>
      makeMovement('in', 1, { date: `2026-05-${String(i + 1).padStart(2, '0')}` })
    );
    const recent = recentMovements(movements, 5);
    expect(recent).toHaveLength(5);
    expect(recent[0].date).toBe('2026-05-15');
  });
});

describe('runningBalance', () => {
  it('computes correct running balance', () => {
    const movements = [
      makeMovement('in', 100, { date: '2026-05-01' }),
      makeMovement('out', 30, { date: '2026-05-05' }),
      makeMovement('adjustment', -5, { date: '2026-05-10' }),
    ];
    const balances = runningBalance(0, movements);
    expect(balances[0].balance).toBe(100);
    expect(balances[1].balance).toBe(70);
    expect(balances[2].balance).toBe(65);
  });
});
