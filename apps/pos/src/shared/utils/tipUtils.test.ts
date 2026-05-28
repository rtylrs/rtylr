import { describe, it, expect } from 'vitest';
import {
  calcTipAmount,
  calcTipPercent,
  isTipReasonable,
  suggestedTipAmounts,
  SUGGESTED_TIP_PCTS,
} from './tipUtils';

describe('calcTipAmount', () => {
  it('calculates 18% tip on a $50 bill', () => {
    expect(calcTipAmount(50, 18)).toBe(9);
  });

  it('calculates 20% tip on a $25 bill', () => {
    expect(calcTipAmount(25, 20)).toBe(5);
  });

  it('returns 0 for a zero subtotal', () => {
    expect(calcTipAmount(0, 18)).toBe(0);
  });

  it('returns 0 for a zero percentage', () => {
    expect(calcTipAmount(50, 0)).toBe(0);
  });

  it('returns 0 for a negative subtotal', () => {
    expect(calcTipAmount(-10, 15)).toBe(0);
  });

  it('returns 0 for a negative percentage', () => {
    expect(calcTipAmount(50, -5)).toBe(0);
  });

  it('returns 0 for NaN inputs', () => {
    expect(calcTipAmount(NaN, 15)).toBe(0);
    expect(calcTipAmount(50, NaN)).toBe(0);
  });

  it('rounds to two decimal places', () => {
    expect(calcTipAmount(33.33, 15)).toBeCloseTo(5.0, 2);
  });
});

describe('calcTipPercent', () => {
  it('calculates the tip percentage from amount and subtotal', () => {
    expect(calcTipPercent(9, 50)).toBe(18);
  });

  it('returns 0 for a zero tip amount', () => {
    expect(calcTipPercent(0, 50)).toBe(0);
  });

  it('returns 0 for a zero subtotal', () => {
    expect(calcTipPercent(5, 0)).toBe(0);
  });

  it('returns 0 for negative inputs', () => {
    expect(calcTipPercent(-5, 50)).toBe(0);
    expect(calcTipPercent(5, -50)).toBe(0);
  });
});

describe('isTipReasonable', () => {
  it('returns true for a reasonable tip', () => {
    expect(isTipReasonable(5, 50)).toBe(true);
  });

  it('returns true for a tip equal to the subtotal', () => {
    expect(isTipReasonable(50, 50)).toBe(true);
  });

  it('returns false for a tip exceeding the subtotal', () => {
    expect(isTipReasonable(60, 50)).toBe(false);
  });

  it('returns true for a zero tip', () => {
    expect(isTipReasonable(0, 50)).toBe(true);
  });

  it('returns false for a negative tip', () => {
    expect(isTipReasonable(-5, 50)).toBe(false);
  });

  it('returns false for NaN tip', () => {
    expect(isTipReasonable(NaN, 50)).toBe(false);
  });
});

describe('suggestedTipAmounts', () => {
  it('returns an entry for every suggested percentage', () => {
    const result = suggestedTipAmounts(100);
    for (const pct of SUGGESTED_TIP_PCTS) {
      expect(result[pct]).toBeDefined();
    }
  });

  it('calculates correct amounts for a $100 bill', () => {
    const result = suggestedTipAmounts(100);
    expect(result[10]).toBe(10);
    expect(result[15]).toBe(15);
    expect(result[20]).toBe(20);
    expect(result[25]).toBe(25);
  });

  it('returns all zeros for a zero subtotal', () => {
    const result = suggestedTipAmounts(0);
    for (const pct of SUGGESTED_TIP_PCTS) {
      expect(result[pct]).toBe(0);
    }
  });
});
