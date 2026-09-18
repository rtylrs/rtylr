import { describe, it, expect } from 'vitest';
import {
  expectedCash,
  cashVariance,
  varianceLabel,
  isVarianceAcceptable,
} from './sessionUtils';

describe('expectedCash', () => {
  it('calculates expected cash from a simple scenario', () => {
    const result = expectedCash({
      openingFloat: 100,
      sales: 200,
      refunds: 20,
      paidIn: 0,
      paidOut: 30,
      drops: 50,
    });
    // 100 + 200 - 20 - 30 - 50 = 200
    expect(result).toBe(200);
  });

  it('returns the opening float when there is no activity', () => {
    expect(expectedCash({ openingFloat: 100, sales: 0, refunds: 0, paidIn: 0, paidOut: 0, drops: 0 })).toBe(100);
  });

  it('floors at 0 when expenses exceed income', () => {
    expect(expectedCash({ openingFloat: 0, sales: 0, refunds: 100, paidIn: 0, paidOut: 0, drops: 0 })).toBe(0);
  });

  it('ignores non-finite values', () => {
    const result = expectedCash({ openingFloat: NaN, sales: 50, refunds: 0, paidIn: 0, paidOut: 0, drops: 0 });
    expect(result).toBe(50);
  });

  it('adds paid_in to expected cash', () => {
    const result = expectedCash({ openingFloat: 100, sales: 0, refunds: 0, paidIn: 50, paidOut: 0, drops: 0 });
    expect(result).toBe(150);
  });
});

describe('cashVariance', () => {
  it('returns a positive variance when actual exceeds expected', () => {
    expect(cashVariance(110, 100)).toBe(10);
  });

  it('returns a negative variance when actual is below expected', () => {
    expect(cashVariance(90, 100)).toBe(-10);
  });

  it('returns 0 when actual equals expected', () => {
    expect(cashVariance(100, 100)).toBe(0);
  });

  it('handles non-finite inputs gracefully', () => {
    expect(cashVariance(NaN, 100)).toBe(-100);
    expect(cashVariance(100, NaN)).toBe(100);
  });
});

describe('varianceLabel', () => {
  it('returns "over" for a positive variance', () => {
    expect(varianceLabel(5)).toBe('over');
  });

  it('returns "under" for a negative variance', () => {
    expect(varianceLabel(-5)).toBe('under');
  });

  it('returns "balanced" for zero variance', () => {
    expect(varianceLabel(0)).toBe('balanced');
  });

  it('returns "balanced" for NaN', () => {
    expect(varianceLabel(NaN)).toBe('balanced');
  });
});

describe('isVarianceAcceptable', () => {
  it('returns true when variance is within the default tolerance', () => {
    expect(isVarianceAcceptable(0.5)).toBe(true);
  });

  it('returns true when variance exactly equals the tolerance', () => {
    expect(isVarianceAcceptable(1)).toBe(true);
  });

  it('returns false when variance exceeds the tolerance', () => {
    expect(isVarianceAcceptable(1.01)).toBe(false);
  });

  it('accepts a custom tolerance', () => {
    expect(isVarianceAcceptable(5, 10)).toBe(true);
    expect(isVarianceAcceptable(15, 10)).toBe(false);
  });

  it('returns false for NaN variance', () => {
    expect(isVarianceAcceptable(NaN)).toBe(false);
  });
});
