import { describe, it, expect } from 'vitest';
import {
  formatPrice,
  priceDiff,
  discountPct,
  hasDiscount,
  unitPrice,
  perItemLabel,
  roundPrice,
} from './priceUtils';

describe('formatPrice', () => {
  it('formats a price as USD by default', () => {
    expect(formatPrice(9.99)).toBe('$9.99');
  });

  it('formats zero as $0.00', () => {
    expect(formatPrice(0)).toBe('$0.00');
  });

  it('formats large amounts with commas', () => {
    expect(formatPrice(1000)).toBe('$1,000.00');
  });
});

describe('priceDiff', () => {
  it('returns the difference between original and discounted', () => {
    expect(priceDiff(10, 8)).toBe(2);
  });

  it('returns 0 when discounted is higher than original', () => {
    expect(priceDiff(8, 10)).toBe(0);
  });

  it('returns 0 for equal prices', () => {
    expect(priceDiff(5, 5)).toBe(0);
  });
});

describe('discountPct', () => {
  it('calculates 20% discount correctly', () => {
    expect(discountPct(10, 8)).toBe(20);
  });

  it('returns 0 when original is 0', () => {
    expect(discountPct(0, 0)).toBe(0);
  });

  it('returns 100 for fully discounted item', () => {
    expect(discountPct(10, 0)).toBe(100);
  });
});

describe('hasDiscount', () => {
  it('returns true when discounted is less than original', () => {
    expect(hasDiscount(10, 8)).toBe(true);
  });

  it('returns false when prices are equal', () => {
    expect(hasDiscount(10, 10)).toBe(false);
  });

  it('returns false when discounted is higher', () => {
    expect(hasDiscount(8, 10)).toBe(false);
  });
});

describe('unitPrice', () => {
  it('calculates the per-unit price', () => {
    expect(unitPrice(30, 3)).toBe(10);
  });

  it('returns 0 for zero quantity', () => {
    expect(unitPrice(30, 0)).toBe(0);
  });
});

describe('perItemLabel', () => {
  it('returns a labeled price string', () => {
    expect(perItemLabel(5)).toBe('$5.00 / item');
  });
});

describe('roundPrice', () => {
  it('rounds to the nearest cent', () => {
    expect(roundPrice(1.005)).toBeCloseTo(1.01, 2);
  });

  it('leaves exact prices unchanged', () => {
    expect(roundPrice(9.99)).toBe(9.99);
  });
});
