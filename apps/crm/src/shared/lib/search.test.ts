import { describe, it, expect } from 'vitest';
import { matchesSearch, filterByQuery, filterByValue, filterByTag, collectTags } from './search';

// ── matchesSearch ─────────────────────────────────────────────────────────────

describe('matchesSearch', () => {
  it('returns true for an exact match', () => {
    expect(matchesSearch('Alice', 'Alice')).toBe(true);
  });

  it('is case-insensitive', () => {
    expect(matchesSearch('Alice Johnson', 'alice')).toBe(true);
    expect(matchesSearch('alice', 'ALICE')).toBe(true);
  });

  it('returns true for a partial match', () => {
    expect(matchesSearch('John Smith', 'mith')).toBe(true);
  });

  it('returns false when needle is not found', () => {
    expect(matchesSearch('Alice', 'Bob')).toBe(false);
  });

  it('trims the needle before matching', () => {
    expect(matchesSearch('Alice', '  Alice  ')).toBe(true);
  });
});

// ── filterByQuery ─────────────────────────────────────────────────────────────

describe('filterByQuery', () => {
  interface Customer {
    name: string;
    email: string;
  }

  const customers: Customer[] = [
    { name: 'Alice Johnson', email: 'alice@example.com' },
    { name: 'Bob Smith', email: 'bob@test.com' },
    { name: 'Charlie Brown', email: 'charlie@example.com' },
  ];

  it('returns all items for an empty query', () => {
    expect(filterByQuery(customers, '', ['name', 'email'])).toHaveLength(3);
  });

  it('returns all items for a whitespace-only query', () => {
    expect(filterByQuery(customers, '   ', ['name', 'email'])).toHaveLength(3);
  });

  it('filters by name field', () => {
    const result = filterByQuery(customers, 'alice', ['name', 'email']);
    expect(result).toHaveLength(1);
    expect(result[0].name).toBe('Alice Johnson');
  });

  it('filters by email field', () => {
    const result = filterByQuery(customers, 'example.com', ['name', 'email']);
    expect(result).toHaveLength(2);
  });

  it('returns an empty array when no items match', () => {
    expect(filterByQuery(customers, 'nobody', ['name', 'email'])).toHaveLength(0);
  });
});

// ── filterByValue ─────────────────────────────────────────────────────────────

describe('filterByValue', () => {
  interface Lead {
    id: number;
    stage: string;
  }

  const leads: Lead[] = [
    { id: 1, stage: 'cold' },
    { id: 2, stage: 'contacted' },
    { id: 3, stage: 'cold' },
  ];

  it('returns all items when value is undefined', () => {
    expect(filterByValue(leads, 'stage', undefined)).toHaveLength(3);
  });

  it('filters items by the given value', () => {
    const result = filterByValue(leads, 'stage', 'cold');
    expect(result).toHaveLength(2);
    expect(result.every((l) => l.stage === 'cold')).toBe(true);
  });

  it('returns empty array when no items match', () => {
    expect(filterByValue(leads, 'stage', 'lost')).toHaveLength(0);
  });
});

// ── filterByTag ───────────────────────────────────────────────────────────────

describe('filterByTag', () => {
  interface Customer {
    name: string;
    tags: string[];
  }

  const customers: Customer[] = [
    { name: 'Alice', tags: ['vip', 'regular'] },
    { name: 'Bob', tags: ['regular'] },
    { name: 'Charlie', tags: ['vip'] },
  ];

  it('returns all items for an empty tag', () => {
    expect(filterByTag(customers, 'tags', '')).toHaveLength(3);
  });

  it('filters to items containing the tag', () => {
    const result = filterByTag(customers, 'tags', 'vip');
    expect(result).toHaveLength(2);
    expect(result.map((c) => c.name)).toEqual(['Alice', 'Charlie']);
  });

  it('returns empty array when no items have the tag', () => {
    expect(filterByTag(customers, 'tags', 'premium')).toHaveLength(0);
  });
});

// ── collectTags ───────────────────────────────────────────────────────────────

describe('collectTags', () => {
  interface Customer {
    name: string;
    tags: string[];
  }

  const customers: Customer[] = [
    { name: 'Alice', tags: ['vip', 'regular'] },
    { name: 'Bob', tags: ['regular', 'at-risk'] },
    { name: 'Charlie', tags: ['vip'] },
  ];

  it('returns all unique tags sorted alphabetically', () => {
    expect(collectTags(customers, 'tags')).toEqual(['at-risk', 'regular', 'vip']);
  });

  it('returns an empty array when no items have tags', () => {
    expect(collectTags([], 'tags')).toEqual([]);
  });

  it('handles items with empty tag arrays', () => {
    const items = [{ name: 'A', tags: [] }];
    expect(collectTags(items, 'tags')).toEqual([]);
  });
});
