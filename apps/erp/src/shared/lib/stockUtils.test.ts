import { describe, it, expect } from 'vitest';
import {
  isLowStock,
  isOutOfStock,
  stockStatusLabel,
  stockDeficit,
  stockSurplus,
  capacityUsedPct,
  inventoryValue,
} from './stockUtils';

describe('isLowStock', () => {
  it('returns true when quantity is below threshold', () => {
    expect(isLowStock(5, 10)).toBe(true);
  });

  it('returns false when quantity equals threshold', () => {
    expect(isLowStock(10, 10)).toBe(false);
  });

  it('returns false when quantity is above threshold', () => {
    expect(isLowStock(20, 10)).toBe(false);
  });

  it('returns true when quantity is 0 and threshold is positive', () => {
    expect(isLowStock(0, 1)).toBe(true);
  });

  it('returns false when both are 0', () => {
    expect(isLowStock(0, 0)).toBe(false);
  });
});

describe('isOutOfStock', () => {
  it('returns true when quantity is 0', () => {
    expect(isOutOfStock(0)).toBe(true);
  });

  it('returns true when quantity is negative', () => {
    expect(isOutOfStock(-1)).toBe(true);
  });

  it('returns false when quantity is positive', () => {
    expect(isOutOfStock(1)).toBe(false);
  });
});

describe('stockStatusLabel', () => {
  it('returns "Out of Stock" when quantity is 0', () => {
    expect(stockStatusLabel(0, 5)).toBe('Out of Stock');
  });

  it('returns "Low Stock" when quantity is below threshold but positive', () => {
    expect(stockStatusLabel(3, 10)).toBe('Low Stock');
  });

  it('returns "In Stock" when quantity meets or exceeds threshold', () => {
    expect(stockStatusLabel(10, 10)).toBe('In Stock');
    expect(stockStatusLabel(100, 10)).toBe('In Stock');
  });

  it('returns "Out of Stock" for negative quantity', () => {
    expect(stockStatusLabel(-5, 10)).toBe('Out of Stock');
  });
});

describe('stockDeficit', () => {
  it('returns the difference when current is below target', () => {
    expect(stockDeficit(3, 10)).toBe(7);
  });

  it('returns 0 when current equals target', () => {
    expect(stockDeficit(10, 10)).toBe(0);
  });

  it('returns 0 when current exceeds target', () => {
    expect(stockDeficit(20, 10)).toBe(0);
  });
});

describe('stockSurplus', () => {
  it('returns 0 when current is below target', () => {
    expect(stockSurplus(3, 10)).toBe(0);
  });

  it('returns 0 when current equals target', () => {
    expect(stockSurplus(10, 10)).toBe(0);
  });

  it('returns the excess when current exceeds target', () => {
    expect(stockSurplus(20, 10)).toBe(10);
  });
});

describe('capacityUsedPct', () => {
  it('returns 0 when capacity is 0', () => {
    expect(capacityUsedPct(10, 0)).toBe(0);
  });

  it('returns 0 when used is 0', () => {
    expect(capacityUsedPct(0, 100)).toBe(0);
  });

  it('returns 50 for half capacity', () => {
    expect(capacityUsedPct(50, 100)).toBe(50);
  });

  it('returns 100 when at capacity', () => {
    expect(capacityUsedPct(100, 100)).toBe(100);
  });

  it('clamps to 100 when over capacity', () => {
    expect(capacityUsedPct(150, 100)).toBe(100);
  });

  it('rounds to nearest integer', () => {
    expect(capacityUsedPct(1, 3)).toBe(33);
  });
});

describe('inventoryValue', () => {
  it('returns quantity multiplied by cost price', () => {
    expect(inventoryValue(10, 5)).toBe(50);
  });

  it('returns 0 when quantity is 0', () => {
    expect(inventoryValue(0, 5)).toBe(0);
  });

  it('returns 0 when cost price is 0', () => {
    expect(inventoryValue(10, 0)).toBe(0);
  });

  it('handles fractional cost prices', () => {
    expect(inventoryValue(3, 1.5)).toBeCloseTo(4.5);
  });
});
