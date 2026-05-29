import { describe, it, expect } from 'vitest';
import {
  sortByNameAsc,
  sortByNameDesc,
  sortByPriceAsc,
  sortByPriceDesc,
  filterActiveItems,
} from './sortMenuItems';

const items = [
  { name: 'Burger', salePrice: 9.99, isActive: true },
  { name: 'Salad', salePrice: 6.5, isActive: true },
  { name: 'Fries', salePrice: 3.0, isActive: false },
  { name: 'Shake', salePrice: 5.0, isActive: true },
];

describe('sortByNameAsc', () => {
  it('sorts items alphabetically A to Z', () => {
    const sorted = sortByNameAsc(items);
    expect(sorted.map((i) => i.name)).toEqual(['Burger', 'Fries', 'Salad', 'Shake']);
  });

  it('does not mutate the original array', () => {
    const original = [...items];
    sortByNameAsc(items);
    expect(items).toEqual(original);
  });

  it('handles an empty array', () => {
    expect(sortByNameAsc([])).toEqual([]);
  });
});

describe('sortByNameDesc', () => {
  it('sorts items alphabetically Z to A', () => {
    const sorted = sortByNameDesc(items);
    expect(sorted.map((i) => i.name)).toEqual(['Shake', 'Salad', 'Fries', 'Burger']);
  });

  it('does not mutate the original array', () => {
    const original = [...items];
    sortByNameDesc(items);
    expect(items).toEqual(original);
  });
});

describe('sortByPriceAsc', () => {
  it('sorts items from lowest to highest price', () => {
    const sorted = sortByPriceAsc(items);
    expect(sorted.map((i) => i.salePrice)).toEqual([3.0, 5.0, 6.5, 9.99]);
  });

  it('places items without salePrice last', () => {
    const withUndefined = [{ name: 'Mystery' }, { name: 'Fries', salePrice: 3.0 }];
    const sorted = sortByPriceAsc(withUndefined);
    expect(sorted[0].name).toBe('Fries');
    expect(sorted[1].name).toBe('Mystery');
  });

  it('does not mutate the original array', () => {
    const original = [...items];
    sortByPriceAsc(items);
    expect(items).toEqual(original);
  });
});

describe('sortByPriceDesc', () => {
  it('sorts items from highest to lowest price', () => {
    const sorted = sortByPriceDesc(items);
    expect(sorted.map((i) => i.salePrice)).toEqual([9.99, 6.5, 5.0, 3.0]);
  });

  it('places items without salePrice last', () => {
    const withUndefined = [{ name: 'Mystery' }, { name: 'Fries', salePrice: 3.0 }];
    const sorted = sortByPriceDesc(withUndefined);
    expect(sorted[0].name).toBe('Fries');
    expect(sorted[1].name).toBe('Mystery');
  });
});

describe('filterActiveItems', () => {
  it('removes items where isActive is false', () => {
    const filtered = filterActiveItems(items);
    expect(filtered.every((i) => i.isActive !== false)).toBe(true);
  });

  it('keeps items where isActive is true', () => {
    const filtered = filterActiveItems(items);
    expect(filtered).toHaveLength(3);
  });

  it('keeps items where isActive is undefined', () => {
    const withUndefined = [{ name: 'Burger' }];
    expect(filterActiveItems(withUndefined)).toHaveLength(1);
  });

  it('returns an empty array when all items are inactive', () => {
    const allInactive = items.map((i) => ({ ...i, isActive: false }));
    expect(filterActiveItems(allInactive)).toHaveLength(0);
  });
});
