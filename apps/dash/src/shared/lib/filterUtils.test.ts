import { describe, it, expect } from 'vitest';
import {
  hasNoFilters,
  activeFilterCount,
  clearFilter,
  clearAllFilters,
  applyFilters,
  matchesRule,
  applyRules,
} from './filterUtils';

describe('hasNoFilters', () => {
  it('returns true for all null values', () => {
    expect(hasNoFilters({ status: null, city: null })).toBe(true);
  });

  it('returns true for empty array values', () => {
    expect(hasNoFilters({ tags: [] })).toBe(true);
  });

  it('returns false when a filter is set', () => {
    expect(hasNoFilters({ status: 'active', city: null })).toBe(false);
  });
});

describe('activeFilterCount', () => {
  it('returns 0 for all null filters', () => {
    expect(activeFilterCount({ a: null, b: null })).toBe(0);
  });

  it('counts non-null filter values', () => {
    expect(activeFilterCount({ a: 'x', b: null, c: 42 })).toBe(2);
  });

  it('does not count empty arrays', () => {
    expect(activeFilterCount({ tags: [], status: 'active' })).toBe(1);
  });
});

describe('clearFilter', () => {
  it('sets the specified field to null', () => {
    const result = clearFilter({ status: 'active', city: 'London' }, 'status');
    expect(result.status).toBeNull();
    expect(result.city).toBe('London');
  });

  it('does not mutate the original object', () => {
    const original = { status: 'active' };
    clearFilter(original, 'status');
    expect(original.status).toBe('active');
  });
});

describe('clearAllFilters', () => {
  it('sets all fields to null', () => {
    const result = clearAllFilters({ a: 'x', b: 42, c: true });
    expect(result.a).toBeNull();
    expect(result.b).toBeNull();
    expect(result.c).toBeNull();
  });

  it('preserves all keys', () => {
    const result = clearAllFilters({ a: 'x', b: 'y' });
    expect(Object.keys(result)).toEqual(['a', 'b']);
  });
});

describe('applyFilters', () => {
  it('merges updates into current filters', () => {
    const result = applyFilters({ status: null, city: null }, { status: 'active' });
    expect(result.status).toBe('active');
    expect(result.city).toBeNull();
  });

  it('overwrites existing filter values', () => {
    const result = applyFilters({ status: 'active' }, { status: 'inactive' });
    expect(result.status).toBe('inactive');
  });
});

describe('matchesRule', () => {
  const record = { name: 'Alice', age: 30, city: 'London', tags: ['vip', 'member'] };

  it('matches eq operator', () => {
    expect(matchesRule(record, { field: 'age', operator: 'eq', value: 30 })).toBe(true);
    expect(matchesRule(record, { field: 'age', operator: 'eq', value: 25 })).toBe(false);
  });

  it('matches neq operator', () => {
    expect(matchesRule(record, { field: 'city', operator: 'neq', value: 'Paris' })).toBe(true);
    expect(matchesRule(record, { field: 'city', operator: 'neq', value: 'London' })).toBe(false);
  });

  it('matches gt operator', () => {
    expect(matchesRule(record, { field: 'age', operator: 'gt', value: 25 })).toBe(true);
    expect(matchesRule(record, { field: 'age', operator: 'gt', value: 30 })).toBe(false);
  });

  it('matches gte operator', () => {
    expect(matchesRule(record, { field: 'age', operator: 'gte', value: 30 })).toBe(true);
  });

  it('matches lt operator', () => {
    expect(matchesRule(record, { field: 'age', operator: 'lt', value: 35 })).toBe(true);
    expect(matchesRule(record, { field: 'age', operator: 'lt', value: 30 })).toBe(false);
  });

  it('matches lte operator', () => {
    expect(matchesRule(record, { field: 'age', operator: 'lte', value: 30 })).toBe(true);
  });

  it('matches contains operator case-insensitively', () => {
    expect(matchesRule(record, { field: 'name', operator: 'contains', value: 'alic' })).toBe(true);
    expect(matchesRule(record, { field: 'name', operator: 'contains', value: 'bob' })).toBe(false);
  });

  it('matches in operator', () => {
    expect(matchesRule(record, { field: 'city', operator: 'in', value: ['London', 'Paris'] })).toBe(
      true
    );
    expect(matchesRule(record, { field: 'city', operator: 'in', value: ['Berlin'] })).toBe(false);
  });
});

describe('applyRules', () => {
  const records = [
    { name: 'Alice', age: 30, city: 'London' },
    { name: 'Bob', age: 25, city: 'Paris' },
    { name: 'Carol', age: 35, city: 'London' },
  ];

  it('returns all records when no rules are applied', () => {
    expect(applyRules(records, [])).toHaveLength(3);
  });

  it('filters by a single rule', () => {
    const result = applyRules(records, [{ field: 'city', operator: 'eq', value: 'London' }]);
    expect(result).toHaveLength(2);
  });

  it('applies multiple rules (AND logic)', () => {
    const result = applyRules(records, [
      { field: 'city', operator: 'eq', value: 'London' },
      { field: 'age', operator: 'gt', value: 30 },
    ]);
    expect(result).toHaveLength(1);
    expect(result[0]!.name).toBe('Carol');
  });
});
