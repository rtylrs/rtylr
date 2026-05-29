import { describe, it, expect } from 'vitest';
import {
  calculateOrderQuantity,
  isAtReorderPoint,
  estimateOrderCost,
  stockToSalesRatio,
  inventoryTurnoverRate,
  fillRate,
} from './reorderMathExtensions';

describe('calculateOrderQuantity', () => {
  it('returns the difference between target and current stock', () => {
    expect(calculateOrderQuantity(30, 100)).toBe(70);
  });

  it('returns 0 when stock meets target', () => {
    expect(calculateOrderQuantity(100, 100)).toBe(0);
  });

  it('returns 0 when stock exceeds target', () => {
    expect(calculateOrderQuantity(120, 100)).toBe(0);
  });

  it('applies minimum order quantity', () => {
    expect(calculateOrderQuantity(95, 100, 10)).toBe(10);
  });

  it('returns needed amount when above minimum', () => {
    expect(calculateOrderQuantity(50, 100, 10)).toBe(50);
  });
});

describe('isAtReorderPoint', () => {
  it('returns true when stock equals reorder point', () => {
    expect(isAtReorderPoint(10, 10)).toBe(true);
  });

  it('returns true when stock is below reorder point', () => {
    expect(isAtReorderPoint(5, 10)).toBe(true);
  });

  it('returns false when stock is above reorder point', () => {
    expect(isAtReorderPoint(15, 10)).toBe(false);
  });
});

describe('estimateOrderCost', () => {
  it('returns quantity times unit cost', () => {
    expect(estimateOrderCost(50, 5)).toBe(250);
  });

  it('returns 0 for zero quantity', () => {
    expect(estimateOrderCost(0, 5)).toBe(0);
  });
});

describe('stockToSalesRatio', () => {
  it('returns days of inventory', () => {
    expect(stockToSalesRatio(100, 10)).toBe(10);
  });

  it('returns Infinity when no sales', () => {
    expect(stockToSalesRatio(100, 0)).toBe(Infinity);
  });
});

describe('inventoryTurnoverRate', () => {
  it('calculates turnover correctly', () => {
    expect(inventoryTurnoverRate(1200, 100)).toBe(12);
  });

  it('returns 0 when average stock is 0', () => {
    expect(inventoryTurnoverRate(1000, 0)).toBe(0);
  });
});

describe('fillRate', () => {
  it('returns 100 when available meets demand', () => {
    expect(fillRate(100, 100)).toBe(100);
  });

  it('returns partial fill rate', () => {
    expect(fillRate(50, 100)).toBe(50);
  });

  it('clamps to 100 when over-stocked', () => {
    expect(fillRate(150, 100)).toBe(100);
  });

  it('returns 100 when demand is 0', () => {
    expect(fillRate(50, 0)).toBe(100);
  });
});
