import { describe, it, expect } from 'vitest';
import {
  matchesSearch,
  filterBySearch,
  paginate,
  uniqueStrings,
  uniqueValues,
} from './searchUtils';

describe('matchesSearch', () => {
  it('returns true when haystack contains needle (case-insensitive)', () => {
    expect(matchesSearch('Widget Alpha', 'widget')).toBe(true);
  });

  it('returns true for an exact match', () => {
    expect(matchesSearch('bolt', 'bolt')).toBe(true);
  });

  it('returns false when needle is not in haystack', () => {
    expect(matchesSearch('Widget Alpha', 'beta')).toBe(false);
  });

  it('is case-insensitive', () => {
    expect(matchesSearch('WIDGET', 'widget')).toBe(true);
    expect(matchesSearch('widget', 'WIDGET')).toBe(true);
  });

  it('matches partial strings', () => {
    expect(matchesSearch('electroplating', 'electro')).toBe(true);
  });
});

describe('filterBySearch', () => {
  const items = [
    { name: 'Widget A', sku: 'WA-001', category: 'Electronics' },
    { name: 'Gadget B', sku: 'GB-002', category: 'Tools' },
    { name: 'Widget C', sku: 'WC-003', category: 'Electronics' },
  ];

  it('returns all items when query is empty', () => {
    expect(filterBySearch(items, '', ['name'])).toHaveLength(3);
  });

  it('returns all items when query is only whitespace', () => {
    expect(filterBySearch(items, '   ', ['name'])).toHaveLength(3);
  });

  it('filters by a single field', () => {
    const result = filterBySearch(items, 'widget', ['name']);
    expect(result).toHaveLength(2);
  });

  it('filters across multiple fields', () => {
    const result = filterBySearch(items, 'WC-003', ['name', 'sku']);
    expect(result).toHaveLength(1);
    expect(result[0].name).toBe('Widget C');
  });

  it('returns empty array when nothing matches', () => {
    expect(filterBySearch(items, 'xyz', ['name', 'sku'])).toHaveLength(0);
  });
});

describe('paginate', () => {
  const items = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];

  it('returns the correct slice for page 1', () => {
    const { items: page } = paginate(items, 1, 3);
    expect(page).toEqual([1, 2, 3]);
  });

  it('returns the correct slice for page 2', () => {
    const { items: page } = paginate(items, 2, 3);
    expect(page).toEqual([4, 5, 6]);
  });

  it('returns the remaining items on the last page', () => {
    const { items: page } = paginate(items, 4, 3);
    expect(page).toEqual([10]);
  });

  it('calculates totalPages correctly', () => {
    const { totalPages } = paginate(items, 1, 3);
    expect(totalPages).toBe(4);
  });

  it('returns totalItems equal to input length', () => {
    const { totalItems } = paginate(items, 1, 5);
    expect(totalItems).toBe(10);
  });

  it('clamps page to 1 for invalid values', () => {
    const { page } = paginate(items, 0, 5);
    expect(page).toBe(1);
  });
});

describe('uniqueStrings', () => {
  it('removes duplicates while preserving order', () => {
    expect(uniqueStrings(['a', 'b', 'a', 'c'])).toEqual(['a', 'b', 'c']);
  });

  it('returns an empty array for empty input', () => {
    expect(uniqueStrings([])).toEqual([]);
  });

  it('returns the same array when all values are unique', () => {
    expect(uniqueStrings(['x', 'y', 'z'])).toEqual(['x', 'y', 'z']);
  });
});

describe('uniqueValues', () => {
  const data = [
    { id: 1, category: 'Electronics' },
    { id: 2, category: 'Tools' },
    { id: 3, category: 'Electronics' },
  ];

  it('returns unique values for a given key', () => {
    expect(uniqueValues(data, 'category')).toEqual(['Electronics', 'Tools']);
  });

  it('preserves insertion order', () => {
    const result = uniqueValues(data, 'category');
    expect(result[0]).toBe('Electronics');
    expect(result[1]).toBe('Tools');
  });

  it('returns all values when all are unique', () => {
    expect(uniqueValues(data, 'id')).toEqual([1, 2, 3]);
  });
});
