import { describe, it, expect } from 'vitest';
import {
  truncate,
  capitalize,
  titleCase,
  camelToSnake,
  snakeToCamel,
  highlightMatch,
  initials,
} from './string';

describe('truncate', () => {
  it('returns the string unchanged when within limit', () => {
    expect(truncate('hello', 10)).toBe('hello');
  });

  it('truncates and appends ellipsis when over limit', () => {
    expect(truncate('hello world', 8)).toBe('hello...');
  });

  it('uses a custom ellipsis', () => {
    expect(truncate('hello world', 6, '!')).toBe('hello!');
  });

  it('returns the original string when it exactly fits', () => {
    expect(truncate('hello', 5)).toBe('hello');
  });
});

describe('capitalize', () => {
  it('capitalizes the first character', () => {
    expect(capitalize('hello')).toBe('Hello');
  });

  it('does not change already-capitalized strings', () => {
    expect(capitalize('Hello')).toBe('Hello');
  });

  it('returns an empty string unchanged', () => {
    expect(capitalize('')).toBe('');
  });

  it('does not change the rest of the string', () => {
    expect(capitalize('hELLO')).toBe('HELLO');
  });
});

describe('titleCase', () => {
  it('capitalizes each word', () => {
    expect(titleCase('total revenue growth')).toBe('Total Revenue Growth');
  });

  it('handles a single word', () => {
    expect(titleCase('revenue')).toBe('Revenue');
  });

  it('preserves inner capitalization', () => {
    expect(titleCase('mrr')).toBe('Mrr');
  });
});

describe('camelToSnake', () => {
  it('converts camelCase to snake_case', () => {
    expect(camelToSnake('totalRevenue')).toBe('total_revenue');
  });

  it('handles multiple uppercase letters', () => {
    expect(camelToSnake('activeCustomerCount')).toBe('active_customer_count');
  });

  it('returns lowercase strings unchanged', () => {
    expect(camelToSnake('revenue')).toBe('revenue');
  });
});

describe('snakeToCamel', () => {
  it('converts snake_case to camelCase', () => {
    expect(snakeToCamel('total_revenue')).toBe('totalRevenue');
  });

  it('handles multiple underscores', () => {
    expect(snakeToCamel('active_customer_count')).toBe('activeCustomerCount');
  });

  it('returns strings without underscores unchanged', () => {
    expect(snakeToCamel('revenue')).toBe('revenue');
  });
});

describe('highlightMatch', () => {
  it('wraps matching text in mark tags', () => {
    const result = highlightMatch('Total Revenue', 'revenue');
    expect(result).toContain('<mark>Revenue</mark>');
  });

  it('is case-insensitive', () => {
    const result = highlightMatch('total revenue', 'REVENUE');
    expect(result).toContain('<mark>revenue</mark>');
  });

  it('returns the original text for empty query', () => {
    expect(highlightMatch('hello', '')).toBe('hello');
  });

  it('handles multiple occurrences', () => {
    const result = highlightMatch('rev rev', 'rev');
    expect(result.split('<mark>').length - 1).toBe(2);
  });
});

describe('initials', () => {
  it('returns first and last initials for a full name', () => {
    expect(initials('Alice Smith')).toBe('AS');
  });

  it('returns a single initial for a one-word name', () => {
    expect(initials('Alice')).toBe('A');
  });

  it('uses first and last word for multi-word names', () => {
    expect(initials('Alice Jane Smith')).toBe('AS');
  });

  it('returns empty string for empty input', () => {
    expect(initials('')).toBe('');
  });

  it('handles extra whitespace', () => {
    expect(initials('  Alice   Smith  ')).toBe('AS');
  });
});
