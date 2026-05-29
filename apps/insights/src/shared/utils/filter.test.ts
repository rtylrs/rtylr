import { describe, it, expect } from 'vitest';
import {
  addFilter,
  removeFilter,
  clearFilters,
  hasActiveFilters,
  activeFilterCount,
  applyFilters,
} from './filter';

describe('addFilter', () => {
  it('adds a key-value pair', () => {
    expect(addFilter({}, 'status', 'active')).toEqual({ status: 'active' });
  });

  it('overwrites existing key', () => {
    expect(addFilter({ status: 'old' }, 'status', 'new')).toEqual({ status: 'new' });
  });
});

describe('removeFilter', () => {
  it('removes the key', () => {
    const result = removeFilter({ status: 'active', type: 'a' }, 'status');
    expect(result).toEqual({ type: 'a' });
  });

  it('is a no-op for missing key', () => {
    expect(removeFilter({}, 'status')).toEqual({});
  });
});

describe('clearFilters', () => {
  it('returns an empty object', () => {
    expect(clearFilters()).toEqual({});
  });
});

describe('hasActiveFilters', () => {
  it('returns true when a filter has a value', () => {
    expect(hasActiveFilters({ status: 'active' })).toBe(true);
  });

  it('returns false for empty filters', () => {
    expect(hasActiveFilters({})).toBe(false);
  });

  it('returns false when all values are empty strings', () => {
    expect(hasActiveFilters({ status: '' })).toBe(false);
  });
});

describe('activeFilterCount', () => {
  it('counts non-empty values', () => {
    expect(activeFilterCount({ a: 'x', b: '', c: 'y' })).toBe(2);
  });
});

describe('applyFilters', () => {
  it('filters rows by active filters (case-insensitive)', () => {
    const rows = [
      { name: 'Alice', status: 'open' },
      { name: 'Bob', status: 'closed' },
    ];
    expect(applyFilters(rows, { status: 'open' })).toHaveLength(1);
  });

  it('returns all rows when no active filters', () => {
    const rows = [{ name: 'Alice' }, { name: 'Bob' }];
    expect(applyFilters(rows, {})).toHaveLength(2);
  });

  it('handles multiple filters with AND logic', () => {
    const rows = [
      { name: 'Alice', dept: 'eng' },
      { name: 'Bob', dept: 'eng' },
      { name: 'Carol', dept: 'hr' },
    ];
    expect(applyFilters(rows, { name: 'alice', dept: 'eng' })).toHaveLength(1);
  });
});
