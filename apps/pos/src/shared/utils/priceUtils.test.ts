import { describe, it, expect } from 'vitest';
import {
  modifierTotal,
  effectiveUnitPrice,
  lineGross,
  itemDiscountAmount,
  lineNet,
  lineTax,
} from './priceUtils';
import type { CartItem } from '../stores/cartStore';

const makeItem = (overrides: Partial<CartItem> = {}): CartItem => ({
  lineId: '1::',
  id: '1::',
  productId: 1,
  name: 'Burger',
  unitPrice: 10,
  price: 10,
  quantity: 2,
  modifiers: [],
  discountPct: 0,
  discountAmount: 0,
  taxRate: 0.1,
  status: 'pending',
  ...overrides,
});

describe('modifierTotal', () => {
  it('returns 0 when there are no modifiers', () => {
    expect(modifierTotal({ modifiers: [] })).toBe(0);
  });

  it('sums positive modifier price deltas', () => {
    const item = makeItem({ modifiers: [
      { id: 1, name: 'Cheese', priceDelta: 1.5 },
      { id: 2, name: 'Bacon', priceDelta: 2 },
    ]});
    expect(modifierTotal(item)).toBe(3.5);
  });

  it('ignores negative price deltas', () => {
    const item = makeItem({ modifiers: [{ id: 1, name: 'Minus', priceDelta: -1 }] });
    expect(modifierTotal(item)).toBe(0);
  });
});

describe('effectiveUnitPrice', () => {
  it('returns the base unit price when there are no modifiers', () => {
    expect(effectiveUnitPrice({ unitPrice: 10, modifiers: [] })).toBe(10);
  });

  it('adds modifier totals to the unit price', () => {
    const item = makeItem({ unitPrice: 10, modifiers: [{ id: 1, name: 'Cheese', priceDelta: 1.5 }] });
    expect(effectiveUnitPrice(item)).toBe(11.5);
  });

  it('returns 0 for a non-finite unit price', () => {
    expect(effectiveUnitPrice({ unitPrice: NaN, modifiers: [] })).toBe(0);
  });
});

describe('lineGross', () => {
  it('calculates gross = unitPrice * quantity', () => {
    const item = makeItem({ unitPrice: 12.5, quantity: 2 });
    expect(lineGross(item)).toBe(25);
  });

  it('includes modifier cost in the gross total', () => {
    const item = makeItem({
      unitPrice: 10,
      quantity: 2,
      modifiers: [{ id: 1, name: 'Cheese', priceDelta: 1.5 }],
    });
    expect(lineGross(item)).toBe(23);
  });

  it('returns 0 for a zero quantity', () => {
    expect(lineGross(makeItem({ quantity: 0 }))).toBe(0);
  });
});

describe('itemDiscountAmount', () => {
  it('returns 0 when there is no discount', () => {
    expect(itemDiscountAmount(makeItem())).toBe(0);
  });

  it('calculates a percentage discount from the gross total', () => {
    const item = makeItem({ unitPrice: 10, quantity: 2, discountPct: 10 });
    expect(itemDiscountAmount(item)).toBe(2);
  });

  it('applies a flat discount directly', () => {
    const item = makeItem({ unitPrice: 10, quantity: 2, discountAmount: 5 });
    expect(itemDiscountAmount(item)).toBe(5);
  });

  it('caps a flat discount at the gross total', () => {
    const item = makeItem({ unitPrice: 5, quantity: 1, discountAmount: 100 });
    expect(itemDiscountAmount(item)).toBe(5);
  });
});

describe('lineNet', () => {
  it('returns the gross when there is no discount', () => {
    const item = makeItem({ unitPrice: 10, quantity: 2 });
    expect(lineNet(item)).toBe(20);
  });

  it('subtracts the percentage discount from the gross', () => {
    const item = makeItem({ unitPrice: 10, quantity: 2, discountPct: 10 });
    expect(lineNet(item)).toBe(18);
  });

  it('floors at 0', () => {
    const item = makeItem({ unitPrice: 5, quantity: 1, discountPct: 100 });
    expect(lineNet(item)).toBe(0);
  });
});

describe('lineTax', () => {
  it('calculates tax on the net line total', () => {
    const item = makeItem({ unitPrice: 10, quantity: 2, taxRate: 0.1 });
    expect(lineTax(item)).toBe(2);
  });

  it('returns 0 for a zero tax rate', () => {
    const item = makeItem({ unitPrice: 10, quantity: 2, taxRate: 0 });
    expect(lineTax(item)).toBe(0);
  });

  it('applies tax after the per-item discount', () => {
    const item = makeItem({ unitPrice: 10, quantity: 2, discountPct: 10, taxRate: 0.1 });
    expect(lineTax(item)).toBeCloseTo(1.8, 2);
  });

  it('returns 0 for a non-finite tax rate', () => {
    const item = makeItem({ taxRate: NaN });
    expect(lineTax(item)).toBe(0);
  });
});
