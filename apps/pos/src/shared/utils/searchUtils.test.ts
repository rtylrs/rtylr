import { describe, it, expect } from 'vitest';
import {
  normalizeQuery,
  matchesQuery,
  filterByQuery,
  sortByField,
  dedupeBy,
} from './searchUtils';

describe('normalizeQuery', () => {
  it('trims whitespace', () => {
    expect(normalizeQuery('  burger  ')).toBe('burger');
  });

  it('lowercases the query', () => {
    expect(normalizeQuery('BURGER')).toBe('burger');
  });

  it('handles an already-normalized query', () => {
    expect(normalizeQuery('fries')).toBe('fries');
  });

  it('returns an empty string for blank input', () => {
    expect(normalizeQuery('   ')).toBe('');
  });
});

describe('matchesQuery', () => {
  it('returns true when the text contains the query', () => {
    expect(matchesQuery('Cheeseburger', 'cheese')).toBe(true);
  });

  it('is case-insensitive', () => {
    expect(matchesQuery('Cheeseburger', 'CHEESE')).toBe(true);
  });

  it('returns false when the text does not contain the query', () => {
    expect(matchesQuery('Fries', 'burger')).toBe(false);
  });

  it('returns true for an empty query (show all)', () => {
    expect(matchesQuery('Anything', '')).toBe(true);
  });
});

describe('filterByQuery', () => {
  const items = [
    { id: '1', name: 'Burger', category: 'Mains' },
    { id: '2', name: 'Fries', category: 'Sides' },
    { id: '3', name: 'Cola', category: 'Drinks' },
    { id: '4', name: 'Cheese Burger', category: 'Mains' },
  ];

  it('returns all items for an empty query', () => {
    expect(filterByQuery(items, '', ['name'])).toHaveLength(4);
  });

  it('filters by a single field', () => {
    const result = filterByQuery(items, 'burger', ['name']);
    expect(result).toHaveLength(2);
    expect(result.map((i) => i.id)).toEqual(['1', '4']);
  });

  it('filters across multiple fields', () => {
    const result = filterByQuery(items, 'mains', ['name', 'category']);
    expect(result).toHaveLength(2);
  });

  it('returns an empty array when nothing matches', () => {
    expect(filterByQuery(items, 'pizza', ['name', 'category'])).toHaveLength(0);
  });

  it('is case-insensitive', () => {
    expect(filterByQuery(items, 'FRIES', ['name'])).toHaveLength(1);
  });
});

describe('sortByField', () => {
  const items = [
    { name: 'Cola', price: '3' },
    { name: 'Burger', price: '12' },
    { name: 'Fries', price: '4' },
  ];

  it('sorts ascending by default', () => {
    const result = sortByField(items, 'name');
    expect(result.map((i) => i.name)).toEqual(['Burger', 'Cola', 'Fries']);
  });

  it('sorts descending when specified', () => {
    const result = sortByField(items, 'name', 'desc');
    expect(result.map((i) => i.name)).toEqual(['Fries', 'Cola', 'Burger']);
  });

  it('does not mutate the original array', () => {
    const original = [...items];
    sortByField(items, 'name');
    expect(items).toEqual(original);
  });
});

describe('dedupeBy', () => {
  it('removes duplicates by key function', () => {
    const items = [
      { id: '1', name: 'Burger' },
      { id: '2', name: 'Fries' },
      { id: '1', name: 'Burger Duplicate' },
    ];
    const result = dedupeBy(items, (i) => i.id);
    expect(result).toHaveLength(2);
    expect(result[0].name).toBe('Burger');
  });

  it('returns all items when no duplicates exist', () => {
    const items = [{ id: '1' }, { id: '2' }, { id: '3' }];
    expect(dedupeBy(items, (i) => i.id)).toHaveLength(3);
  });

  it('handles an empty array', () => {
    expect(dedupeBy([], (i: { id: string }) => i.id)).toEqual([]);
  });
});
