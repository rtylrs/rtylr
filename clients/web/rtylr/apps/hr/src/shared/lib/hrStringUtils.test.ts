import { describe, it, expect } from 'vitest';
import {
  truncate,
  toTitleCase,
  getInitials,
  formatPhone,
  maskString,
  normalizeSearchQuery,
  isBlank,
  camelToLabel,
  snakeToLabel,
} from './hrStringUtils';

describe('truncate', () => {
  it('returns string unchanged when within limit', () => {
    expect(truncate('Hello', 10)).toBe('Hello');
  });
  it('truncates with ellipsis when over limit', () => {
    expect(truncate('Hello World', 8)).toBe('Hello...');
  });
});

describe('toTitleCase', () => {
  it('capitalizes each word', () => {
    expect(toTitleCase('john doe smith')).toBe('John Doe Smith');
  });
  it('handles already-cased input', () => {
    expect(toTitleCase('JOHN DOE')).toBe('John Doe');
  });
});

describe('getInitials', () => {
  it('returns initials for two-part name', () => {
    expect(getInitials('John Doe')).toBe('JD');
  });
  it('returns single initial for one-part name', () => {
    expect(getInitials('Alice')).toBe('A');
  });
  it('uses only first two parts of long name', () => {
    expect(getInitials('John Middle Doe')).toBe('JM');
  });
});

describe('formatPhone', () => {
  it('formats 10-digit number', () => {
    expect(formatPhone('5551234567')).toBe('(555) 123-4567');
  });
  it('strips non-digit chars before formatting', () => {
    expect(formatPhone('555-123-4567')).toBe('(555) 123-4567');
  });
  it('returns original if not 10 digits', () => {
    expect(formatPhone('12345')).toBe('12345');
  });
});

describe('maskString', () => {
  it('masks all but last n chars', () => {
    expect(maskString('secret123', 3)).toBe('******123');
  });
  it('returns unchanged if shorter than visibleChars', () => {
    expect(maskString('ab', 5)).toBe('ab');
  });
});

describe('normalizeSearchQuery', () => {
  it('trims and lowercases', () => {
    expect(normalizeSearchQuery('  Hello World  ')).toBe('hello world');
  });
  it('collapses internal spaces', () => {
    expect(normalizeSearchQuery('foo   bar')).toBe('foo bar');
  });
});

describe('isBlank', () => {
  it('returns true for empty string', () => {
    expect(isBlank('')).toBe(true);
  });
  it('returns true for whitespace only', () => {
    expect(isBlank('   ')).toBe(true);
  });
  it('returns false for non-empty string', () => {
    expect(isBlank('hello')).toBe(false);
  });
});

describe('camelToLabel', () => {
  it('converts camelCase to label', () => {
    expect(camelToLabel('hireDate')).toBe('Hire Date');
  });
  it('handles single word', () => {
    expect(camelToLabel('salary')).toBe('Salary');
  });
});

describe('snakeToLabel', () => {
  it('converts snake_case to label', () => {
    expect(snakeToLabel('hire_date')).toBe('Hire Date');
  });
  it('capitalizes each part', () => {
    expect(snakeToLabel('full_time_employee')).toBe('Full Time Employee');
  });
});
