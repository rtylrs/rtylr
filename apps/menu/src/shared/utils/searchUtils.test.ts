import { describe, it, expect } from 'vitest';
import {
  containsQuery,
  tokenise,
  matchesAllTokens,
  highlight,
  filterByField,
  sortByRelevance,
} from './searchUtils';

describe('containsQuery', () => {
  it('returns true when the query is found', () => {
    expect(containsQuery('Spicy Burger', 'burger')).toBe(true);
  });

  it('is case-insensitive', () => {
    expect(containsQuery('Caesar Salad', 'SALAD')).toBe(true);
  });

  it('returns true for an empty query', () => {
    expect(containsQuery('anything', '')).toBe(true);
  });

  it('returns false when query is not found', () => {
    expect(containsQuery('Pizza', 'pasta')).toBe(false);
  });
});

describe('tokenise', () => {
  it('splits on whitespace', () => {
    expect(tokenise('spicy  chicken')).toEqual(['spicy', 'chicken']);
  });

  it('lowercases all tokens', () => {
    expect(tokenise('PIZZA Margherita')).toEqual(['pizza', 'margherita']);
  });

  it('returns empty array for blank input', () => {
    expect(tokenise('   ')).toHaveLength(0);
  });
});

describe('matchesAllTokens', () => {
  it('returns true when all tokens match', () => {
    expect(matchesAllTokens('Spicy Chicken Wings', 'chicken wings')).toBe(true);
  });

  it('returns false when any token is missing', () => {
    expect(matchesAllTokens('Chicken Wings', 'chicken wings pasta')).toBe(false);
  });

  it('returns true for empty query', () => {
    expect(matchesAllTokens('anything', '')).toBe(true);
  });
});

describe('highlight', () => {
  it('wraps matching text with default markers', () => {
    expect(highlight('Spicy Burger', 'Burger')).toBe('Spicy <mark>Burger</mark>');
  });

  it('is case-insensitive', () => {
    expect(highlight('BURGER', 'burger')).toContain('<mark>BURGER</mark>');
  });

  it('returns original text when no match', () => {
    expect(highlight('Pizza', 'pasta')).toBe('Pizza');
  });

  it('returns original text for empty query', () => {
    expect(highlight('Pizza', '')).toBe('Pizza');
  });

  it('accepts custom markers', () => {
    expect(highlight('Hello World', 'World', '[', ']')).toBe('Hello [World]');
  });
});

describe('filterByField', () => {
  const items = [{ name: 'Burger' }, { name: 'Salad' }, { name: 'Chicken Burger' }];

  it('filters items that match the query', () => {
    const result = filterByField(items, (i) => i.name, 'burger');
    expect(result).toHaveLength(2);
  });

  it('returns all items for empty query', () => {
    expect(filterByField(items, (i) => i.name, '')).toHaveLength(3);
  });
});

describe('sortByRelevance', () => {
  const items = [{ name: 'Spicy Burger' }, { name: 'Burger Combo' }, { name: 'Salad' }];

  it('puts items starting with query first', () => {
    const result = sortByRelevance(items, (i) => i.name, 'burger');
    expect(result[0].name).toBe('Burger Combo');
  });

  it('returns items unchanged for empty query', () => {
    const result = sortByRelevance(items, (i) => i.name, '');
    expect(result).toEqual(items);
  });
});
