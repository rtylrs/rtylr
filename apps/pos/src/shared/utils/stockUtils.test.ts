import { describe, it, expect } from 'vitest';
import {
  isLowStock,
  isOutOfStock,
  applyStockDelta,
  reorderQuantity,
  stockStatus,
  formatStockQty,
} from './stockUtils';

describe('isLowStock', () => {
  it('returns true when qty is below threshold', () => {
    expect(isLowStock(3, 5)).toBe(true);
  });

  it('returns true when qty equals threshold', () => {
    expect(isLowStock(5, 5)).toBe(true);
  });

  it('returns false when qty is above threshold', () => {
    expect(isLowStock(10, 5)).toBe(false);
  });

  it('returns false for non-finite inputs', () => {
    expect(isLowStock(NaN, 5)).toBe(false);
    expect(isLowStock(5, NaN)).toBe(false);
  });
});

describe('isOutOfStock', () => {
  it('returns true when qty is 0', () => {
    expect(isOutOfStock(0)).toBe(true);
  });

  it('returns true when qty is negative', () => {
    expect(isOutOfStock(-1)).toBe(true);
  });

  it('returns false when qty is positive', () => {
    expect(isOutOfStock(1)).toBe(false);
  });

  it('returns false for NaN', () => {
    expect(isOutOfStock(NaN)).toBe(false);
  });
});

describe('applyStockDelta', () => {
  it('adds a positive delta to current quantity', () => {
    expect(applyStockDelta(10, 5)).toBe(15);
  });

  it('subtracts a negative delta from current quantity', () => {
    expect(applyStockDelta(10, -3)).toBe(7);
  });

  it('floors at 0 when delta would make it negative', () => {
    expect(applyStockDelta(5, -10)).toBe(0);
  });

  it('handles a zero delta', () => {
    expect(applyStockDelta(5, 0)).toBe(5);
  });

  it('handles NaN gracefully', () => {
    expect(applyStockDelta(NaN, 5)).toBeGreaterThanOrEqual(0);
  });
});

describe('reorderQuantity', () => {
  it('returns the difference between target and current', () => {
    expect(reorderQuantity(3, 10)).toBe(7);
  });

  it('returns 0 when current meets the target', () => {
    expect(reorderQuantity(10, 10)).toBe(0);
  });

  it('returns 0 when current exceeds the target', () => {
    expect(reorderQuantity(15, 10)).toBe(0);
  });

  it('returns 0 for non-finite inputs', () => {
    expect(reorderQuantity(NaN, 10)).toBe(0);
    expect(reorderQuantity(5, NaN)).toBe(0);
  });
});

describe('stockStatus', () => {
  it('returns "out" when qty is 0', () => {
    expect(stockStatus(0, 5)).toBe('out');
  });

  it('returns "out" when qty is negative', () => {
    expect(stockStatus(-1, 5)).toBe('out');
  });

  it('returns "low" when qty is at or below threshold', () => {
    expect(stockStatus(3, 5)).toBe('low');
    expect(stockStatus(5, 5)).toBe('low');
  });

  it('returns "ok" when qty is above threshold', () => {
    expect(stockStatus(10, 5)).toBe('ok');
  });
});

describe('formatStockQty', () => {
  it('formats a positive quantity without a unit', () => {
    expect(formatStockQty(42)).toBe('42');
  });

  it('formats a positive quantity with a unit', () => {
    expect(formatStockQty(10, 'kg')).toBe('10 kg');
  });

  it('returns "Out" for zero quantity', () => {
    expect(formatStockQty(0)).toBe('Out');
  });

  it('returns "Out" for a negative quantity', () => {
    expect(formatStockQty(-5)).toBe('Out');
  });

  it('returns "Out" for non-finite quantity', () => {
    expect(formatStockQty(NaN)).toBe('Out');
  });
});
