import { describe, it, expect } from 'vitest';
import {
  sortBrandsByName,
  sortBrandsByProductCount,
  totalBrandProductCount,
  hasProducts,
  brandInitials,
  filterBrandsWithProducts,
} from './brandUtils';
import type { Brand } from '../services/erpApi';

function makeBrand(overrides: Partial<Brand> = {}): Brand {
  return { id: 1, name: 'Acme', productCount: 10, ...overrides };
}

const brands: Brand[] = [
  { id: 1, name: 'Zebra Corp', productCount: 5 },
  { id: 2, name: 'Apple Inc', productCount: 50 },
  { id: 3, name: 'Mid Brand', productCount: 0 },
];

describe('sortBrandsByName', () => {
  it('sorts alphabetically', () => {
    const sorted = sortBrandsByName(brands);
    expect(sorted[0].name).toBe('Apple Inc');
    expect(sorted[2].name).toBe('Zebra Corp');
  });

  it('does not mutate the original array', () => {
    sortBrandsByName(brands);
    expect(brands[0].name).toBe('Zebra Corp');
  });
});

describe('sortBrandsByProductCount', () => {
  it('sorts highest count first', () => {
    const sorted = sortBrandsByProductCount(brands);
    expect(sorted[0].productCount).toBe(50);
  });
});

describe('totalBrandProductCount', () => {
  it('sums product counts', () => {
    expect(totalBrandProductCount(brands)).toBe(55);
  });

  it('returns 0 for empty array', () => {
    expect(totalBrandProductCount([])).toBe(0);
  });
});

describe('hasProducts', () => {
  it('returns true when productCount > 0', () => {
    expect(hasProducts(makeBrand({ productCount: 5 }))).toBe(true);
  });

  it('returns false when productCount is 0', () => {
    expect(hasProducts(makeBrand({ productCount: 0 }))).toBe(false);
  });
});

describe('brandInitials', () => {
  it('returns first 2 chars for single word', () => {
    expect(brandInitials('Acme')).toBe('AC');
  });

  it('returns first letter of first two words', () => {
    expect(brandInitials('Apple Inc')).toBe('AI');
  });

  it('handles extra whitespace', () => {
    expect(brandInitials('  Brand  Name  ')).toBe('BN');
  });
});

describe('filterBrandsWithProducts', () => {
  it('returns only brands with products', () => {
    const result = filterBrandsWithProducts(brands);
    expect(result).toHaveLength(2);
    expect(result.every((b) => b.productCount > 0)).toBe(true);
  });

  it('returns empty array when all brands have 0 products', () => {
    expect(filterBrandsWithProducts([makeBrand({ productCount: 0 })])).toHaveLength(0);
  });
});
