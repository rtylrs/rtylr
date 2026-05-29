import { describe, it, expect } from 'vitest';
import {
  grossMarginPct,
  markupPct,
  saleFromMarkup,
  applyMultiplier,
  applyDiscount,
  profitPerUnit,
  roundPrice,
  isValidPrice,
} from './priceUtils';

describe('grossMarginPct', () => {
  it('calculates margin correctly', () => {
    expect(grossMarginPct(5, 10)).toBeCloseTo(50);
  });

  it('returns 0 when salePrice is 0', () => {
    expect(grossMarginPct(5, 0)).toBe(0);
  });

  it('returns 0 for break-even', () => {
    expect(grossMarginPct(10, 10)).toBe(0);
  });

  it('handles fractional prices', () => {
    expect(grossMarginPct(2.5, 5)).toBeCloseTo(50);
  });
});

describe('markupPct', () => {
  it('calculates markup correctly', () => {
    expect(markupPct(5, 10)).toBeCloseTo(100);
  });

  it('returns 0 when costPrice is 0', () => {
    expect(markupPct(0, 10)).toBe(0);
  });

  it('handles 50% markup', () => {
    expect(markupPct(10, 15)).toBeCloseTo(50);
  });
});

describe('saleFromMarkup', () => {
  it('calculates sale price from markup', () => {
    expect(saleFromMarkup(10, 50)).toBeCloseTo(15);
  });

  it('returns cost price for 0% markup', () => {
    expect(saleFromMarkup(10, 0)).toBe(10);
  });

  it('doubles cost price for 100% markup', () => {
    expect(saleFromMarkup(10, 100)).toBe(20);
  });
});

describe('applyMultiplier', () => {
  it('applies multiplier correctly', () => {
    expect(applyMultiplier(10, 1.2)).toBeCloseTo(12);
  });

  it('returns 0 for negative multiplier', () => {
    expect(applyMultiplier(10, -1)).toBe(0);
  });

  it('returns same price for multiplier of 1', () => {
    expect(applyMultiplier(10, 1)).toBe(10);
  });
});

describe('applyDiscount', () => {
  it('applies percentage discount', () => {
    expect(applyDiscount(100, 20)).toBeCloseTo(80);
  });

  it('returns original price for 0% discount', () => {
    expect(applyDiscount(100, 0)).toBe(100);
  });

  it('returns 0 for 100% discount', () => {
    expect(applyDiscount(100, 100)).toBe(0);
  });

  it('returns price for negative discount', () => {
    expect(applyDiscount(100, -10)).toBe(100);
  });
});

describe('profitPerUnit', () => {
  it('returns the difference between sale and cost', () => {
    expect(profitPerUnit(5, 10)).toBe(5);
  });

  it('returns negative for loss', () => {
    expect(profitPerUnit(10, 8)).toBe(-2);
  });
});

describe('roundPrice', () => {
  it('rounds to 2 decimal places', () => {
    expect(roundPrice(1.234)).toBe(1.23);
    expect(roundPrice(1.235)).toBeCloseTo(1.24, 1);
  });

  it('leaves whole numbers unchanged', () => {
    expect(roundPrice(10)).toBe(10);
  });
});

describe('isValidPrice', () => {
  it('returns true for positive prices', () => {
    expect(isValidPrice(10)).toBe(true);
    expect(isValidPrice(0)).toBe(true);
  });

  it('returns false for negative prices', () => {
    expect(isValidPrice(-1)).toBe(false);
  });

  it('returns false for NaN and Infinity', () => {
    expect(isValidPrice(NaN)).toBe(false);
    expect(isValidPrice(Infinity)).toBe(false);
  });
});
