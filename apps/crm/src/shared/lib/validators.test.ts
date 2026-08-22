import { describe, it, expect } from 'vitest';
import {
  isValidEmail,
  isValidPhone,
  isNonEmpty,
  isInRange,
  isIsoDate,
  isValidPoints,
  isValidUrl,
  isOptionalPositiveInteger,
} from './validators';

describe('isValidEmail', () => {
  it('accepts a standard email address', () => {
    expect(isValidEmail('user@example.com')).toBe(true);
  });
  it('rejects an email missing the @ sign', () => {
    expect(isValidEmail('userexample.com')).toBe(false);
  });
  it('rejects an empty string', () => {
    expect(isValidEmail('')).toBe(false);
  });
  it('trims surrounding whitespace before checking', () => {
    expect(isValidEmail('  user@example.com  ')).toBe(true);
  });
});

describe('isValidPhone', () => {
  it('accepts a US phone number with dashes', () => {
    expect(isValidPhone('555-123-4567')).toBe(true);
  });
  it('accepts an international number with leading plus', () => {
    expect(isValidPhone('+1 (555) 123-4567')).toBe(true);
  });
  it('rejects a string that is too short', () => {
    expect(isValidPhone('123')).toBe(false);
  });
  it('rejects an empty string', () => {
    expect(isValidPhone('')).toBe(false);
  });
  it('rejects punctuation without enough digits', () => {
    expect(isValidPhone('---.( )')).toBe(false);
    expect(isValidPhone('+1 (555) 12')).toBe(false);
  });
  it('accepts a seven-digit local number', () => {
    expect(isValidPhone('555-1234')).toBe(true);
  });
});

describe('isNonEmpty', () => {
  it('returns true for a non-blank string', () => {
    expect(isNonEmpty('hello')).toBe(true);
  });
  it('returns false for an empty string', () => {
    expect(isNonEmpty('')).toBe(false);
  });
  it('returns false for a whitespace-only string', () => {
    expect(isNonEmpty('   ')).toBe(false);
  });
});

describe('isInRange', () => {
  it('returns true for a value within range', () => {
    expect(isInRange(5, 1, 10)).toBe(true);
  });
  it('returns false for a value below min', () => {
    expect(isInRange(0, 1, 10)).toBe(false);
  });
  it('returns false for NaN', () => {
    expect(isInRange(NaN, 0, 10)).toBe(false);
  });
});

describe('isIsoDate', () => {
  it('accepts a valid ISO date', () => {
    expect(isIsoDate('2024-06-15')).toBe(true);
  });
  it('rejects a date with slashes', () => {
    expect(isIsoDate('2024/06/15')).toBe(false);
  });
  it('rejects an empty string', () => {
    expect(isIsoDate('')).toBe(false);
  });
  it('rejects an impossible calendar date', () => {
    expect(isIsoDate('2024-02-30')).toBe(false);
  });
});

describe('isValidPoints', () => {
  it('accepts zero points', () => {
    expect(isValidPoints(0)).toBe(true);
  });
  it('rejects a negative number', () => {
    expect(isValidPoints(-1)).toBe(false);
  });
  it('rejects a decimal number', () => {
    expect(isValidPoints(1.5)).toBe(false);
  });
});

describe('isOptionalPositiveInteger', () => {
  it('allows an empty optional value', () => {
    expect(isOptionalPositiveInteger('')).toBe(true);
  });
  it('accepts a positive whole number', () => {
    expect(isOptionalPositiveInteger('42')).toBe(true);
  });
  it('rejects zero, decimals, and non-numeric values', () => {
    expect(isOptionalPositiveInteger('0')).toBe(false);
    expect(isOptionalPositiveInteger('1.5')).toBe(false);
    expect(isOptionalPositiveInteger('user-1')).toBe(false);
  });
});

describe('isValidUrl', () => {
  it('accepts an https URL', () => {
    expect(isValidUrl('https://example.com')).toBe(true);
  });
  it('rejects a plain string', () => {
    expect(isValidUrl('not-a-url')).toBe(false);
  });
  it('rejects an empty string', () => {
    expect(isValidUrl('')).toBe(false);
  });
});
