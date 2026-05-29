import { describe, it, expect } from 'vitest';
import {
  isStockTakeComplete,
  isStockTakeActive,
  stockTakeProgressPct,
  hasVariance,
  hasShrinkage,
  stockTakeProgressLabel,
  sortStockTakesByDateDesc,
  filterStockTakesByStatus,
} from './stockTakeUtils';
import type { StockTake } from '../services/erpApi';

function makeTake(overrides: Partial<StockTake> = {}): StockTake {
  return {
    id: 1,
    name: 'Q1 Count',
    warehouseId: 1,
    warehouseName: 'Main',
    status: 'Draft',
    itemsCounted: 0,
    totalItems: 100,
    variance: 0,
    createdAt: '2024-01-01T00:00:00Z',
    ...overrides,
  };
}

describe('isStockTakeComplete', () => {
  it('returns true for Completed status', () => {
    expect(isStockTakeComplete(makeTake({ status: 'Completed' }))).toBe(true);
  });

  it('returns false for other statuses', () => {
    expect(isStockTakeComplete(makeTake({ status: 'Draft' }))).toBe(false);
    expect(isStockTakeComplete(makeTake({ status: 'In Progress' }))).toBe(false);
  });
});

describe('isStockTakeActive', () => {
  it('returns true for In Progress status', () => {
    expect(isStockTakeActive(makeTake({ status: 'In Progress' }))).toBe(true);
  });

  it('returns false for other statuses', () => {
    expect(isStockTakeActive(makeTake({ status: 'Draft' }))).toBe(false);
    expect(isStockTakeActive(makeTake({ status: 'Completed' }))).toBe(false);
  });
});

describe('stockTakeProgressPct', () => {
  it('returns correct percentage', () => {
    expect(stockTakeProgressPct(makeTake({ itemsCounted: 45, totalItems: 100 }))).toBe(45);
  });

  it('returns 0 when totalItems is 0', () => {
    expect(stockTakeProgressPct(makeTake({ itemsCounted: 0, totalItems: 0 }))).toBe(0);
  });

  it('clamps to 100', () => {
    expect(stockTakeProgressPct(makeTake({ itemsCounted: 110, totalItems: 100 }))).toBe(100);
  });

  it('returns 0 at start', () => {
    expect(stockTakeProgressPct(makeTake({ itemsCounted: 0, totalItems: 50 }))).toBe(0);
  });
});

describe('hasVariance', () => {
  it('returns true when variance is non-zero', () => {
    expect(hasVariance(makeTake({ variance: 5 }))).toBe(true);
    expect(hasVariance(makeTake({ variance: -3 }))).toBe(true);
  });

  it('returns false when variance is 0', () => {
    expect(hasVariance(makeTake({ variance: 0 }))).toBe(false);
  });
});

describe('hasShrinkage', () => {
  it('returns true when variance is negative', () => {
    expect(hasShrinkage(makeTake({ variance: -5 }))).toBe(true);
  });

  it('returns false when variance is 0 or positive', () => {
    expect(hasShrinkage(makeTake({ variance: 0 }))).toBe(false);
    expect(hasShrinkage(makeTake({ variance: 3 }))).toBe(false);
  });
});

describe('stockTakeProgressLabel', () => {
  it('returns formatted label', () => {
    expect(stockTakeProgressLabel(makeTake({ itemsCounted: 45, totalItems: 100 }))).toBe(
      '45 / 100 items'
    );
  });

  it('handles zero counts', () => {
    expect(stockTakeProgressLabel(makeTake({ itemsCounted: 0, totalItems: 50 }))).toBe(
      '0 / 50 items'
    );
  });
});

describe('sortStockTakesByDateDesc', () => {
  it('sorts newest first', () => {
    const takes = [
      makeTake({ id: 1, createdAt: '2024-01-01T00:00:00Z' }),
      makeTake({ id: 2, createdAt: '2024-03-01T00:00:00Z' }),
      makeTake({ id: 3, createdAt: '2024-02-01T00:00:00Z' }),
    ];
    const sorted = sortStockTakesByDateDesc(takes);
    expect(sorted[0].id).toBe(2);
    expect(sorted[2].id).toBe(1);
  });

  it('does not mutate original array', () => {
    const takes = [
      makeTake({ id: 1, createdAt: '2024-01-01T00:00:00Z' }),
      makeTake({ id: 2, createdAt: '2024-03-01T00:00:00Z' }),
    ];
    sortStockTakesByDateDesc(takes);
    expect(takes[0].id).toBe(1);
  });
});

describe('filterStockTakesByStatus', () => {
  const takes = [
    makeTake({ id: 1, status: 'Draft' }),
    makeTake({ id: 2, status: 'In Progress' }),
    makeTake({ id: 3, status: 'Completed' }),
    makeTake({ id: 4, status: 'Draft' }),
  ];

  it('filters by Draft status', () => {
    expect(filterStockTakesByStatus(takes, 'Draft')).toHaveLength(2);
  });

  it('filters by Completed status', () => {
    expect(filterStockTakesByStatus(takes, 'Completed')).toHaveLength(1);
  });

  it('returns empty array when no matches', () => {
    expect(filterStockTakesByStatus([], 'Draft')).toHaveLength(0);
  });
});
