import { describe, it, expect } from 'vitest';
import { calculateReorderPoint, calculateEOQ, calculateDaysCoverage } from './reorderMath';

// ---------------------------------------------------------------------------
// calculateReorderPoint
// ---------------------------------------------------------------------------

describe('calculateReorderPoint', () => {
  it('returns 0 when daily sales is 0', () => {
    expect(calculateReorderPoint(0, 7)).toBe(0);
  });

  it('returns 0 when lead time is 0', () => {
    expect(calculateReorderPoint(10, 0)).toBe(0);
  });

  it('computes the correct point with default safety factor (0.5)', () => {
    // avgDailySales=10, leadTimeDays=5 → demand=50, safetyStock=ceil(50*0.5)=25
    // result = ceil(50 + 25) = 75
    expect(calculateReorderPoint(10, 5)).toBe(75);
  });

  it('computes correctly with a custom safety factor of 0', () => {
    // No safety stock → just the demand during lead time.
    expect(calculateReorderPoint(10, 5, 0)).toBe(50);
  });

  it('computes correctly with a custom safety factor of 1', () => {
    // demand=20, safetyStock=ceil(20*1)=20, result=ceil(40)=40
    expect(calculateReorderPoint(4, 5, 1)).toBe(40);
  });

  it('always returns a non-negative integer', () => {
    const result = calculateReorderPoint(-5, 3);
    expect(result).toBeGreaterThanOrEqual(0);
    expect(Number.isInteger(result)).toBe(true);
  });

  it('returns a whole number (ceil applied)', () => {
    // avgDailySales=3, leadTimeDays=1 → demand=3, safetyStock=ceil(1.5)=2, result=ceil(5)=5
    expect(calculateReorderPoint(3, 1)).toBe(5);
  });
});

// ---------------------------------------------------------------------------
// calculateEOQ
// ---------------------------------------------------------------------------

describe('calculateEOQ', () => {
  it('returns 0 when annualDemand is 0', () => {
    expect(calculateEOQ(0, 50, 10)).toBe(0);
  });

  it('returns 0 when orderCost is 0', () => {
    expect(calculateEOQ(1000, 0, 10)).toBe(0);
  });

  it('returns 0 when holdingCost is 0', () => {
    expect(calculateEOQ(1000, 50, 0)).toBe(0);
  });

  it('returns 0 for negative inputs', () => {
    expect(calculateEOQ(-100, 50, 10)).toBe(0);
    expect(calculateEOQ(100, -50, 10)).toBe(0);
    expect(calculateEOQ(100, 50, -10)).toBe(0);
  });

  it('computes the classic textbook example', () => {
    // D=1000, S=50, H=2 → EOQ = sqrt(2*1000*50/2) = sqrt(50000) ≈ 223.6 → ceil=224
    expect(calculateEOQ(1000, 50, 2)).toBe(224);
  });

  it('returns a positive integer', () => {
    const result = calculateEOQ(500, 40, 5);
    expect(result).toBeGreaterThan(0);
    expect(Number.isInteger(result)).toBe(true);
  });
});

// ---------------------------------------------------------------------------
// calculateDaysCoverage
// ---------------------------------------------------------------------------

describe('calculateDaysCoverage', () => {
  it('returns Infinity when avgDailySales is 0', () => {
    expect(calculateDaysCoverage(100, 0)).toBe(Infinity);
  });

  it('returns Infinity when avgDailySales is negative', () => {
    expect(calculateDaysCoverage(100, -5)).toBe(Infinity);
  });

  it('returns 0 when currentStock is 0', () => {
    expect(calculateDaysCoverage(0, 10)).toBe(0);
  });

  it('computes days coverage correctly (floored)', () => {
    // 25 / 4 = 6.25 → floor = 6
    expect(calculateDaysCoverage(25, 4)).toBe(6);
  });

  it('returns an exact integer when stock divides evenly', () => {
    expect(calculateDaysCoverage(30, 5)).toBe(6);
  });
});
