import { describe, it, expect } from 'vitest';
import {
  isRootCategory,
  getRootCategories,
  getChildCategories,
  hasChildren,
  buildCategoryPath,
  sortCategoriesByName,
  sortCategoriesByProductCount,
  totalProductCount,
} from './categoryUtils';
import type { Category } from '../services/erpApi';

const electronics: Category = {
  id: 1,
  name: 'Electronics',
  parentId: null,
  productCount: 50,
  sortOrder: 1,
};
const phones: Category = { id: 2, name: 'Phones', parentId: 1, productCount: 30, sortOrder: 2 };
const accessories: Category = {
  id: 3,
  name: 'Accessories',
  parentId: 2,
  productCount: 10,
  sortOrder: 3,
};
const clothing: Category = {
  id: 4,
  name: 'Clothing',
  parentId: null,
  productCount: 20,
  sortOrder: 4,
};

const all = [electronics, phones, accessories, clothing];

describe('isRootCategory', () => {
  it('returns true for categories without a parent', () => {
    expect(isRootCategory(electronics)).toBe(true);
    expect(isRootCategory(clothing)).toBe(true);
  });

  it('returns false for categories with a parent', () => {
    expect(isRootCategory(phones)).toBe(false);
    expect(isRootCategory(accessories)).toBe(false);
  });
});

describe('getRootCategories', () => {
  it('returns only root categories', () => {
    const roots = getRootCategories(all);
    expect(roots).toHaveLength(2);
    expect(roots.every(isRootCategory)).toBe(true);
  });

  it('returns empty array when none are root', () => {
    expect(getRootCategories([phones, accessories])).toHaveLength(0);
  });
});

describe('getChildCategories', () => {
  it('returns children of a given parent', () => {
    expect(getChildCategories(all, 1)).toHaveLength(1);
    expect(getChildCategories(all, 1)[0].name).toBe('Phones');
  });

  it('returns empty array when no children', () => {
    expect(getChildCategories(all, 3)).toHaveLength(0);
  });
});

describe('hasChildren', () => {
  it('returns true when category has children', () => {
    expect(hasChildren(electronics, all)).toBe(true);
  });

  it('returns false when category has no children', () => {
    expect(hasChildren(accessories, all)).toBe(false);
  });
});

describe('buildCategoryPath', () => {
  it('returns just the name for root categories', () => {
    expect(buildCategoryPath(electronics, all)).toBe('Electronics');
  });

  it('builds a two-level path', () => {
    expect(buildCategoryPath(phones, all)).toBe('Electronics > Phones');
  });

  it('builds a three-level path', () => {
    expect(buildCategoryPath(accessories, all)).toBe('Electronics > Phones > Accessories');
  });
});

describe('sortCategoriesByName', () => {
  it('sorts alphabetically', () => {
    const sorted = sortCategoriesByName([clothing, electronics]);
    expect(sorted[0].name).toBe('Clothing');
    expect(sorted[1].name).toBe('Electronics');
  });

  it('does not mutate the original array', () => {
    const original = [clothing, electronics];
    sortCategoriesByName(original);
    expect(original[0].name).toBe('Clothing');
  });
});

describe('sortCategoriesByProductCount', () => {
  it('sorts highest product count first', () => {
    const sorted = sortCategoriesByProductCount(all);
    expect(sorted[0].productCount).toBeGreaterThanOrEqual(sorted[1].productCount);
  });
});

describe('totalProductCount', () => {
  it('returns sum of all product counts', () => {
    expect(totalProductCount(all)).toBe(110);
  });

  it('returns 0 for empty array', () => {
    expect(totalProductCount([])).toBe(0);
  });
});
