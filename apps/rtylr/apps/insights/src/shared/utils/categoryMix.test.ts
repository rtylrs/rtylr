import { describe, it, expect } from 'vitest';
import {
  totalCategorySales,
  topCategories,
  categoryShare,
  categoryLabels,
  categoryValues,
} from './categoryMix';
import type { CategorySales } from '../hooks/queries/types';

function makeCat(id: number, name: string, revenue: number): CategorySales {
  return { categoryId: id, categoryName: name, unitsSold: 10, revenueMinorUnits: revenue };
}

describe('totalCategorySales', () => {
  it('sums revenue across all categories', () => {
    const cats = [makeCat(1, 'A', 1000), makeCat(2, 'B', 2000)];
    expect(totalCategorySales(cats)).toBe(3000);
  });

  it('returns 0 for empty array', () => {
    expect(totalCategorySales([])).toBe(0);
  });
});

describe('topCategories', () => {
  it('returns top n by revenue', () => {
    const cats = [makeCat(1, 'A', 100), makeCat(2, 'B', 500), makeCat(3, 'C', 200)];
    const top = topCategories(cats, 2);
    expect(top[0].categoryId).toBe(2);
    expect(top).toHaveLength(2);
  });

  it('does not mutate original array', () => {
    const cats = [makeCat(1, 'A', 100)];
    topCategories(cats, 1);
    expect(cats[0].categoryId).toBe(1);
  });
});

describe('categoryShare', () => {
  it('computes percentage for each category', () => {
    const cats = [makeCat(1, 'A', 1000), makeCat(2, 'B', 3000)];
    const shares = categoryShare(cats);
    expect(shares[0].computedPct).toBe(25);
    expect(shares[1].computedPct).toBe(75);
  });

  it('returns 0 percent when total is 0', () => {
    const cats = [makeCat(1, 'A', 0)];
    expect(categoryShare(cats)[0].computedPct).toBe(0);
  });
});

describe('categoryLabels', () => {
  it('extracts category names', () => {
    const cats = [makeCat(1, 'Food'), makeCat(2, 'Drinks')];
    expect(categoryLabels(cats)).toEqual(['Food', 'Drinks']);
  });
});

describe('categoryValues', () => {
  it('extracts revenue values', () => {
    const cats = [makeCat(1, 'A', 1000), makeCat(2, 'B', 2000)];
    expect(categoryValues(cats)).toEqual([1000, 2000]);
  });
});
