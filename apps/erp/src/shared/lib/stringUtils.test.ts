import { describe, it, expect } from 'vitest';
import {
  capitalize,
  toTitleCase,
  camelToLabel,
  snakeToLabel,
  nameInitials,
  containsIgnoreCase,
  slugify,
  pluralize,
} from './stringUtils';

describe('capitalize', () => {
  it('capitalizes first letter and lowercases rest', () => {
    expect(capitalize('hello WORLD')).toBe('Hello world');
  });

  it('handles empty string', () => {
    expect(capitalize('')).toBe('');
  });

  it('handles single character', () => {
    expect(capitalize('a')).toBe('A');
  });
});

describe('toTitleCase', () => {
  it('capitalizes each word', () => {
    expect(toTitleCase('hello world')).toBe('Hello World');
  });

  it('lowercases properly', () => {
    expect(toTitleCase('HELLO WORLD')).toBe('Hello World');
  });
});

describe('camelToLabel', () => {
  it('splits camelCase into words', () => {
    expect(camelToLabel('stockCount')).toBe('Stock Count');
  });

  it('handles PascalCase', () => {
    expect(camelToLabel('TotalValue')).toBe('Total Value');
  });
});

describe('snakeToLabel', () => {
  it('converts snake_case to readable label', () => {
    expect(snakeToLabel('purchase_order')).toBe('Purchase Order');
  });

  it('handles single word', () => {
    expect(snakeToLabel('name')).toBe('Name');
  });
});

describe('nameInitials', () => {
  it('returns first 2 chars for single word', () => {
    expect(nameInitials('Alice')).toBe('AL');
  });

  it('returns first letters of first and last name', () => {
    expect(nameInitials('Alice Smith')).toBe('AS');
  });

  it('returns first letters of first and last for multiple names', () => {
    expect(nameInitials('Alice Bob Smith')).toBe('AS');
  });

  it('handles empty string', () => {
    expect(nameInitials('')).toBe('');
  });
});

describe('containsIgnoreCase', () => {
  it('returns true when substring found (case-insensitive)', () => {
    expect(containsIgnoreCase('Hello World', 'world')).toBe(true);
    expect(containsIgnoreCase('HELLO', 'hello')).toBe(true);
  });

  it('returns false when not found', () => {
    expect(containsIgnoreCase('Hello', 'xyz')).toBe(false);
  });
});

describe('slugify', () => {
  it('converts to lowercase with hyphens', () => {
    expect(slugify('Hello World')).toBe('hello-world');
  });

  it('removes special characters', () => {
    expect(slugify('Hello! World?')).toBe('hello-world');
  });

  it('handles multiple spaces', () => {
    expect(slugify('Hello   World')).toBe('hello-world');
  });
});

describe('pluralize', () => {
  it('returns singular for count 1', () => {
    expect(pluralize('item', 1)).toBe('item');
  });

  it('returns plural form with s for other counts', () => {
    expect(pluralize('item', 2)).toBe('items');
    expect(pluralize('item', 0)).toBe('items');
  });

  it('uses custom plural when provided', () => {
    expect(pluralize('person', 2, 'people')).toBe('people');
  });
});
