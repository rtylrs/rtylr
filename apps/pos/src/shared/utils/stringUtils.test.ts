import { describe, it, expect } from 'vitest';
import {
  capitalize,
  toTitleCase,
  initials,
  maskExceptLast,
  isEmail,
  isNumericString,
  digitsOnly,
  normalizeWhitespace,
} from './stringUtils';

describe('capitalize', () => {
  it('capitalizes the first letter', () => {
    expect(capitalize('burger')).toBe('Burger');
  });

  it('leaves other characters unchanged', () => {
    expect(capitalize('hello WORLD')).toBe('Hello WORLD');
  });

  it('returns an empty string for empty input', () => {
    expect(capitalize('')).toBe('');
  });
});

describe('toTitleCase', () => {
  it('converts snake_case to title case', () => {
    expect(toTitleCase('order_status')).toBe('Order Status');
  });

  it('converts kebab-case to title case', () => {
    expect(toTitleCase('payment-method')).toBe('Payment Method');
  });

  it('handles a single word', () => {
    expect(toTitleCase('burger')).toBe('Burger');
  });

  it('lowercases the non-first letters of each word', () => {
    expect(toTitleCase('CASH_DROP')).toBe('Cash Drop');
  });
});

describe('initials', () => {
  it('returns initials from a full name', () => {
    expect(initials('John Doe')).toBe('JD');
  });

  it('returns a single initial for a single name', () => {
    expect(initials('Alice')).toBe('A');
  });

  it('uses the first and last name for multiple words', () => {
    expect(initials('John Michael Doe')).toBe('JD');
  });

  it('returns an empty string for an empty name', () => {
    expect(initials('')).toBe('');
  });

  it('handles extra whitespace', () => {
    expect(initials('  Jane   Smith  ')).toBe('JS');
  });
});

describe('maskExceptLast', () => {
  it('masks all but the last 4 characters', () => {
    expect(maskExceptLast('1234567890')).toBe('******7890');
  });

  it('respects a custom visible count', () => {
    expect(maskExceptLast('ABCDE', 2)).toBe('***DE');
  });

  it('returns the full string when it is shorter than visibleCount', () => {
    expect(maskExceptLast('1234', 6)).toBe('1234');
  });

  it('uses the provided mask character', () => {
    expect(maskExceptLast('ABCDE', 2, 'X')).toBe('XXXDE');
  });
});

describe('isEmail', () => {
  it('accepts a valid email', () => {
    expect(isEmail('user@example.com')).toBe(true);
  });

  it('accepts an email with a subdomain', () => {
    expect(isEmail('user@mail.example.com')).toBe(true);
  });

  it('rejects an email without an at-sign', () => {
    expect(isEmail('userexample.com')).toBe(false);
  });

  it('rejects an empty string', () => {
    expect(isEmail('')).toBe(false);
  });

  it('rejects an email with spaces', () => {
    expect(isEmail('user @example.com')).toBe(false);
  });
});

describe('isNumericString', () => {
  it('returns true for a string of digits', () => {
    expect(isNumericString('12345')).toBe(true);
  });

  it('returns false for a string with letters', () => {
    expect(isNumericString('123a5')).toBe(false);
  });

  it('returns false for a decimal number', () => {
    expect(isNumericString('12.5')).toBe(false);
  });

  it('returns false for an empty string', () => {
    expect(isNumericString('')).toBe(false);
  });
});

describe('digitsOnly', () => {
  it('removes non-digit characters', () => {
    expect(digitsOnly('+1 (555) 123-4567')).toBe('15551234567');
  });

  it('leaves an all-digit string unchanged', () => {
    expect(digitsOnly('12345')).toBe('12345');
  });

  it('returns an empty string for an all-non-digit string', () => {
    expect(digitsOnly('abc')).toBe('');
  });
});

describe('normalizeWhitespace', () => {
  it('trims leading and trailing whitespace', () => {
    expect(normalizeWhitespace('  hello  ')).toBe('hello');
  });

  it('collapses multiple internal spaces', () => {
    expect(normalizeWhitespace('hello   world')).toBe('hello world');
  });

  it('handles a string with tabs and newlines', () => {
    expect(normalizeWhitespace('hello\t\nworld')).toBe('hello world');
  });

  it('returns an empty string for blank input', () => {
    expect(normalizeWhitespace('   ')).toBe('');
  });
});
