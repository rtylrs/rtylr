import { describe, expect, it } from 'vitest';
import {
  applyFilter,
  applyFilters,
  paginateItems,
  searchByFields,
  totalPages,
} from './hrFilterUtils';

const employees = [
  { name: 'Alice', dept: 'Engineering', salary: 90000, active: true },
  { name: 'Bob', dept: 'HR', salary: 60000, active: false },
  { name: 'Charlie', dept: 'Engineering', salary: 75000, active: true },
  { name: 'Diana', dept: 'Sales', salary: 65000, active: true },
];

describe('applyFilter', () => {
  it('filters by equality', () => {
    const result = applyFilter(employees, {
      key: 'dept',
      operator: 'eq',
      value: 'Engineering',
    });
    expect(result).toHaveLength(2);
    expect(result.every((e) => e.dept === 'Engineering')).toBe(true);
  });

  it('filters by greater than', () => {
    const result = applyFilter(employees, {
      key: 'salary',
      operator: 'gt',
      value: 70000,
    });
    expect(result).toHaveLength(2);
  });

  it('filters by contains (case-insensitive)', () => {
    const result = applyFilter(employees, {
      key: 'name',
      operator: 'contains',
      value: 'ali',
    });
    expect(result).toHaveLength(1);
    expect(result[0].name).toBe('Alice');
  });

  it('filters by not equal', () => {
    const result = applyFilter(employees, {
      key: 'active',
      operator: 'neq',
      value: true,
    });
    expect(result).toHaveLength(1);
    expect(result[0].name).toBe('Bob');
  });
});

describe('applyFilters', () => {
  it('applies multiple filters', () => {
    const result = applyFilters(employees, [
      { key: 'dept', operator: 'eq', value: 'Engineering' },
      { key: 'salary', operator: 'gte', value: 80000 },
    ]);
    expect(result).toHaveLength(1);
    expect(result[0].name).toBe('Alice');
  });

  it('returns all items when no rules', () => {
    expect(applyFilters(employees, [])).toHaveLength(4);
  });
});

describe('searchByFields', () => {
  it('searches across multiple fields', () => {
    const result = searchByFields(employees, 'eng', ['dept', 'name']);
    expect(result).toHaveLength(2);
  });

  it('returns all items for empty query', () => {
    expect(searchByFields(employees, '', ['name'])).toHaveLength(4);
  });

  it('is case-insensitive', () => {
    const result = searchByFields(employees, 'ALICE', ['name']);
    expect(result).toHaveLength(1);
  });
});

describe('paginateItems', () => {
  const items = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];

  it('returns correct page', () => {
    expect(paginateItems(items, 1, 3)).toEqual([1, 2, 3]);
    expect(paginateItems(items, 2, 3)).toEqual([4, 5, 6]);
    expect(paginateItems(items, 4, 3)).toEqual([10]);
  });

  it('returns empty array for out-of-range page', () => {
    expect(paginateItems(items, 5, 3)).toEqual([]);
  });
});

describe('totalPages', () => {
  it('calculates total pages', () => {
    expect(totalPages(10, 3)).toBe(4);
    expect(totalPages(9, 3)).toBe(3);
    expect(totalPages(0, 10)).toBe(0);
  });

  it('returns 0 for pageSize 0', () => {
    expect(totalPages(10, 0)).toBe(0);
  });
});
