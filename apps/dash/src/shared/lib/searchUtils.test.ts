import { describe, it, expect } from 'vitest';
import {
  matchesText,
  filterByText,
  uniqueFieldValues,
  highlightMatches,
  parseSearchTerms,
  matchesAllTerms,
} from './searchUtils';

describe('matchesText', () => {
  it('returns true for case-insensitive match', () => {
    expect(matchesText('Hello World', 'hello')).toBe(true);
  });

  it('returns false when needle is absent', () => {
    expect(matchesText('Hello World', 'xyz')).toBe(false);
  });

  it('matches partial strings', () => {
    expect(matchesText('Dashboard', 'dash')).toBe(true);
  });
});

describe('filterByText', () => {
  const items = [
    { name: 'Alice', role: 'admin' },
    { name: 'Bob', role: 'staff' },
    { name: 'Carol', role: 'admin' },
  ];

  it('returns all items for empty query', () => {
    expect(filterByText(items, '', ['name'])).toHaveLength(3);
  });

  it('returns all items for whitespace query', () => {
    expect(filterByText(items, '   ', ['name'])).toHaveLength(3);
  });

  it('filters by single field', () => {
    const result = filterByText(items, 'admin', ['role']);
    expect(result).toHaveLength(2);
  });

  it('filters across multiple fields', () => {
    const result = filterByText(items, 'alice', ['name', 'role']);
    expect(result).toHaveLength(1);
    expect(result[0]!.name).toBe('Alice');
  });

  it('returns empty array when nothing matches', () => {
    expect(filterByText(items, 'xyz', ['name'])).toHaveLength(0);
  });
});

describe('uniqueFieldValues', () => {
  const items = [
    { name: 'Alice', dept: 'Engineering' },
    { name: 'Bob', dept: 'Sales' },
    { name: 'Carol', dept: 'Engineering' },
  ];

  it('returns unique values preserving order', () => {
    expect(uniqueFieldValues(items, 'dept')).toEqual(['Engineering', 'Sales']);
  });

  it('returns empty array for empty input', () => {
    expect(uniqueFieldValues([], 'dept')).toEqual([]);
  });
});

describe('highlightMatches', () => {
  it('wraps matched text', () => {
    const result = highlightMatches('Hello World', 'world');
    expect(result).toContain('<mark>World</mark>');
  });

  it('returns original text for empty query', () => {
    expect(highlightMatches('Hello World', '')).toBe('Hello World');
  });

  it('is case-insensitive', () => {
    const result = highlightMatches('HELLO', 'hello');
    expect(result).toContain('<mark>HELLO</mark>');
  });
});

describe('parseSearchTerms', () => {
  it('splits by whitespace', () => {
    expect(parseSearchTerms('foo bar baz')).toEqual(['foo', 'bar', 'baz']);
  });

  it('trims and filters empty strings', () => {
    expect(parseSearchTerms('  hello  world  ')).toEqual(['hello', 'world']);
  });

  it('returns empty array for empty string', () => {
    expect(parseSearchTerms('')).toEqual([]);
  });
});

describe('matchesAllTerms', () => {
  it('returns true when all terms match', () => {
    expect(matchesAllTerms('Alice Admin Manager', 'alice admin')).toBe(true);
  });

  it('returns false when any term is missing', () => {
    expect(matchesAllTerms('Alice Staff', 'alice admin')).toBe(false);
  });

  it('returns true for empty query', () => {
    expect(matchesAllTerms('anything', '')).toBe(true);
  });
});
