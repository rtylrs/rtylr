import { describe, it, expect } from 'vitest';
import {
  hasActiveFilters,
  activeFilterCount,
  clearFilters,
  updateFilter,
  filtersEqual,
  filtersToParams,
} from './filterUtils';
import type { FilterState } from './filterUtils';

const empty: FilterState = { search: '', status: undefined, category: undefined, brand: undefined };
const active: FilterState = {
  search: 'widget',
  status: 'In Stock',
  category: undefined,
  brand: undefined,
};

describe('hasActiveFilters', () => {
  it('returns false for empty filters', () => {
    expect(hasActiveFilters(empty)).toBe(false);
  });

  it('returns true when any filter is set', () => {
    expect(hasActiveFilters(active)).toBe(true);
    expect(hasActiveFilters({ ...empty, category: 'Electronics' })).toBe(true);
  });
});

describe('activeFilterCount', () => {
  it('returns 0 for empty filters', () => {
    expect(activeFilterCount(empty)).toBe(0);
  });

  it('counts all active filters', () => {
    const f: FilterState = {
      search: 'w',
      status: 'In Stock',
      category: 'Electronics',
      brand: 'Acme',
    };
    expect(activeFilterCount(f)).toBe(4);
  });

  it('counts only set filters', () => {
    expect(activeFilterCount(active)).toBe(2);
  });
});

describe('clearFilters', () => {
  it('returns an empty filter state', () => {
    const cleared = clearFilters();
    expect(hasActiveFilters(cleared)).toBe(false);
  });
});

describe('updateFilter', () => {
  it('updates a single field', () => {
    const updated = updateFilter(empty, 'search', 'widget');
    expect(updated.search).toBe('widget');
    expect(updated.status).toBeUndefined();
  });

  it('does not mutate original', () => {
    updateFilter(empty, 'search', 'widget');
    expect(empty.search).toBe('');
  });
});

describe('filtersEqual', () => {
  it('returns true for identical filters', () => {
    expect(filtersEqual(empty, clearFilters())).toBe(true);
  });

  it('returns false for different filters', () => {
    expect(filtersEqual(empty, active)).toBe(false);
  });
});

describe('filtersToParams', () => {
  it('returns only non-empty params', () => {
    const params = filtersToParams(active);
    expect(params.search).toBe('widget');
    expect(params.status).toBe('In Stock');
    expect(params.category).toBeUndefined();
  });

  it('returns empty object for empty filters', () => {
    expect(Object.keys(filtersToParams(empty))).toHaveLength(0);
  });
});
