import { describe, it, expect } from 'vitest';
import {
  orderSubtotal,
  roundCents,
  calcTax,
  calcChange,
  canCharge,
  applyPercentDiscount,
  applyFlatDiscount,
  splitEvenly,
} from './orderCalc';

describe('orderSubtotal', () => {
  it('returns 0 for an empty item list', () => {
    expect(orderSubtotal([])).toBe(0);
  });

  it('sums price * quantity for a single item', () => {
    expect(orderSubtotal([{ id: '1', name: 'Burger', price: 12.5, quantity: 2 }])).toBe(25);
  });

  it('sums multiple line items', () => {
    const items = [
      { id: '1', name: 'Burger', price: 10, quantity: 2 },
      { id: '2', name: 'Fries', price: 4, quantity: 3 },
    ];
    expect(orderSubtotal(items)).toBe(32);
  });
});

describe('roundCents', () => {
  it('rounds to two decimal places', () => {
    expect(roundCents(1.234)).toBeCloseTo(1.23, 2);
  });

  it('leaves an already-rounded value unchanged', () => {
    expect(roundCents(5.25)).toBe(5.25);
  });

  it('rounds 0 to 0', () => {
    expect(roundCents(0)).toBe(0);
  });
});

describe('calcTax', () => {
  it('calculates 8% tax correctly', () => {
    expect(calcTax(100, 0.08)).toBe(8);
  });

  it('returns 0 for a zero subtotal', () => {
    expect(calcTax(0, 0.08)).toBe(0);
  });

  it('returns 0 for a negative subtotal', () => {
    expect(calcTax(-10, 0.08)).toBe(0);
  });

  it('returns 0 for a negative tax rate', () => {
    expect(calcTax(100, -0.08)).toBe(0);
  });

  it('returns 0 for non-finite tax rate', () => {
    expect(calcTax(100, NaN)).toBe(0);
  });

  it('rounds to two decimal places', () => {
    expect(calcTax(33.33, 0.1)).toBeCloseTo(3.33, 2);
  });
});

describe('calcChange', () => {
  it('returns correct change when tendering more than due', () => {
    expect(calcChange(20, 14.99)).toBeCloseTo(5.01, 2);
  });

  it('returns 0 when tendered equals amount due', () => {
    expect(calcChange(10, 10)).toBe(0);
  });

  it('returns 0 when tendered is less than amount due', () => {
    expect(calcChange(5, 10)).toBe(0);
  });

  it('returns 0 for non-finite tendered amount', () => {
    expect(calcChange(NaN, 10)).toBe(0);
  });

  it('returns 0 for a negative tendered amount', () => {
    expect(calcChange(-5, 10)).toBe(0);
  });
});

describe('canCharge', () => {
  it('returns true when tendered covers the full amount', () => {
    expect(canCharge(20, 19.99)).toBe(true);
  });

  it('returns true when tendered equals amount due', () => {
    expect(canCharge(10, 10)).toBe(true);
  });

  it('returns false when tendered is less than amount due', () => {
    expect(canCharge(5, 10)).toBe(false);
  });

  it('returns false when amount due is zero', () => {
    expect(canCharge(10, 0)).toBe(false);
  });

  it('returns false for non-finite inputs', () => {
    expect(canCharge(NaN, 10)).toBe(false);
    expect(canCharge(10, NaN)).toBe(false);
  });
});

describe('applyPercentDiscount', () => {
  it('applies a 10% discount to a subtotal', () => {
    expect(applyPercentDiscount(100, 10)).toBe(90);
  });

  it('applies a 100% discount reducing total to zero', () => {
    expect(applyPercentDiscount(100, 100)).toBe(0);
  });

  it('clamps discount above 100% to 100%', () => {
    expect(applyPercentDiscount(100, 150)).toBe(0);
  });

  it('applies a 0% discount leaving subtotal unchanged', () => {
    expect(applyPercentDiscount(50, 0)).toBe(50);
  });

  it('returns 0 for negative subtotal', () => {
    expect(applyPercentDiscount(-10, 10)).toBe(0);
  });
});

describe('applyFlatDiscount', () => {
  it('subtracts a flat amount from the subtotal', () => {
    expect(applyFlatDiscount(100, 15)).toBe(85);
  });

  it('floors at 0 when discount exceeds subtotal', () => {
    expect(applyFlatDiscount(10, 50)).toBe(0);
  });

  it('returns 0 for negative subtotal', () => {
    expect(applyFlatDiscount(-5, 3)).toBe(0);
  });

  it('leaves subtotal unchanged when flatAmount is 0', () => {
    expect(applyFlatDiscount(100, 0)).toBe(100);
  });

  it('returns subtotal when flatAmount is negative', () => {
    expect(applyFlatDiscount(100, -10)).toBe(100);
  });
});

describe('splitEvenly', () => {
  it('returns the total as-is for 1 seat', () => {
    expect(splitEvenly(30, 1)).toEqual([30]);
  });

  it('splits evenly across 2 seats', () => {
    expect(splitEvenly(20, 2)).toEqual([10, 10]);
  });

  it('distributes rounding remainder to first seat', () => {
    const parts = splitEvenly(10, 3);
    expect(parts).toHaveLength(3);
    const total = parts.reduce((a, b) => a + b, 0);
    expect(total).toBeCloseTo(10, 2);
  });

  it('returns an empty array for invalid total', () => {
    expect(splitEvenly(-1, 3)).toEqual([]);
  });
});
