import { describe, it, expect } from 'vitest';
import {
  applyPercent,
  percentDiscountAmount,
  applyFlat,
  effectiveDiscountPercent,
  isDiscountPermitted,
  isFlatDiscountPermitted,
  stackPercentDiscounts,
} from './discountUtils';

describe('applyPercent', () => {
  it('returns price after percent discount', () => {
    expect(applyPercent(100, 10)).toBeCloseTo(90);
  });

  it('returns 0 for 100% discount', () => {
    expect(applyPercent(100, 100)).toBe(0);
  });

  it('clamps negative discount to 0%', () => {
    expect(applyPercent(100, -10)).toBe(100);
  });

  it('clamps discount above 100%', () => {
    expect(applyPercent(100, 150)).toBe(0);
  });

  it('returns 0 for non-finite inputs', () => {
    expect(applyPercent(NaN, 10)).toBe(0);
    expect(applyPercent(100, NaN)).toBe(0);
  });
});

describe('percentDiscountAmount', () => {
  it('returns the discount amount for a percentage', () => {
    expect(percentDiscountAmount(100, 15)).toBeCloseTo(15);
  });

  it('returns 0 for 0% discount', () => {
    expect(percentDiscountAmount(100, 0)).toBe(0);
  });

  it('returns 0 for non-finite inputs', () => {
    expect(percentDiscountAmount(NaN, 10)).toBe(0);
  });
});

describe('applyFlat', () => {
  it('subtracts the flat discount from the amount', () => {
    expect(applyFlat(100, 20)).toBeCloseTo(80);
  });

  it('floors at 0 when discount exceeds amount', () => {
    expect(applyFlat(10, 20)).toBe(0);
  });

  it('returns amount unchanged for non-finite discount', () => {
    expect(applyFlat(100, NaN)).toBe(100);
  });
});

describe('effectiveDiscountPercent', () => {
  it('returns the correct percentage', () => {
    expect(effectiveDiscountPercent(100, 25)).toBeCloseTo(25);
  });

  it('returns 0 when original is 0', () => {
    expect(effectiveDiscountPercent(0, 10)).toBe(0);
  });

  it('caps at 100%', () => {
    expect(effectiveDiscountPercent(100, 120)).toBe(100);
  });
});

describe('isDiscountPermitted', () => {
  it('returns true for a valid discount percent', () => {
    expect(isDiscountPermitted(15)).toBe(true);
  });

  it('returns false for a negative discount', () => {
    expect(isDiscountPermitted(-5)).toBe(false);
  });

  it('returns false when exceeding the max', () => {
    expect(isDiscountPermitted(50, 30)).toBe(false);
  });

  it('returns false for NaN', () => {
    expect(isDiscountPermitted(NaN)).toBe(false);
  });
});

describe('isFlatDiscountPermitted', () => {
  it('returns true when flat discount is within the total', () => {
    expect(isFlatDiscountPermitted(10, 100)).toBe(true);
  });

  it('returns false when flat discount exceeds the total', () => {
    expect(isFlatDiscountPermitted(110, 100)).toBe(false);
  });

  it('returns false for negative discounts', () => {
    expect(isFlatDiscountPermitted(-5, 100)).toBe(false);
  });

  it('returns false for NaN', () => {
    expect(isFlatDiscountPermitted(NaN, 100)).toBe(false);
  });
});

describe('stackPercentDiscounts', () => {
  it('stacks two discounts multiplicatively', () => {
    // 100 - 10% = 90, 90 - 10% = 81
    expect(stackPercentDiscounts(100, 10, 10)).toBeCloseTo(81);
  });

  it('applying 100% twice still results in 0', () => {
    expect(stackPercentDiscounts(100, 100, 100)).toBe(0);
  });
});
