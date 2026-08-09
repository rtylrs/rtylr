import { describe, it, expect } from 'vitest';
import {
  isValidEmail,
  isNonEmpty,
  isInRange,
  isDigitsOnly,
  isLengthBetween,
  isPositiveInt,
  isValidUrl,
  isValidQuantity,
} from './validationUtils';

describe('isValidEmail', () => {
  it('returns true for a valid email', () => {
    expect(isValidEmail('user@example.com')).toBe(true);
  });

  it('returns false for an email without @', () => {
    expect(isValidEmail('userexample.com')).toBe(false);
  });

  it('returns false for an empty string', () => {
    expect(isValidEmail('')).toBe(false);
  });

  it('returns false for an email without domain', () => {
    expect(isValidEmail('user@')).toBe(false);
  });
});

describe('isNonEmpty', () => {
  it('returns true for a non-empty string', () => {
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
  it('returns true when value is within range', () => {
    expect(isInRange(5, 1, 10)).toBe(true);
  });

  it('returns true when value equals min', () => {
    expect(isInRange(1, 1, 10)).toBe(true);
  });

  it('returns true when value equals max', () => {
    expect(isInRange(10, 1, 10)).toBe(true);
  });

  it('returns false when value is below min', () => {
    expect(isInRange(0, 1, 10)).toBe(false);
  });

  it('returns false when value is above max', () => {
    expect(isInRange(11, 1, 10)).toBe(false);
  });
});

describe('isDigitsOnly', () => {
  it('returns true for a digits-only string', () => {
    expect(isDigitsOnly('12345')).toBe(true);
  });

  it('returns false when the string contains letters', () => {
    expect(isDigitsOnly('123a5')).toBe(false);
  });

  it('returns false for an empty string', () => {
    expect(isDigitsOnly('')).toBe(false);
  });
});

describe('isLengthBetween', () => {
  it('returns true when length is within range', () => {
    expect(isLengthBetween('hello', 3, 10)).toBe(true);
  });

  it('returns false when string is too short', () => {
    expect(isLengthBetween('hi', 3, 10)).toBe(false);
  });

  it('returns false when string is too long', () => {
    expect(isLengthBetween('hello world!!', 3, 10)).toBe(false);
  });
});

describe('isPositiveInt', () => {
  it('returns true for a positive integer', () => {
    expect(isPositiveInt(5)).toBe(true);
  });

  it('returns false for zero', () => {
    expect(isPositiveInt(0)).toBe(false);
  });

  it('returns false for a negative integer', () => {
    expect(isPositiveInt(-1)).toBe(false);
  });

  it('returns false for a decimal', () => {
    expect(isPositiveInt(1.5)).toBe(false);
  });
});

describe('isValidUrl', () => {
  it('returns true for an https URL', () => {
    expect(isValidUrl('https://example.com')).toBe(true);
  });

  it('returns true for an http URL', () => {
    expect(isValidUrl('http://example.com')).toBe(true);
  });

  it('returns false for a plain string', () => {
    expect(isValidUrl('not a url')).toBe(false);
  });

  it('returns false for a javascript: URL', () => {
    expect(isValidUrl('javascript:alert(1)')).toBe(false);
  });
});

describe('isValidQuantity', () => {
  it('returns true for a positive integer', () => {
    expect(isValidQuantity(3)).toBe(true);
  });

  it('returns true for zero', () => {
    expect(isValidQuantity(0)).toBe(true);
  });

  it('returns false for a negative number', () => {
    expect(isValidQuantity(-1)).toBe(false);
  });

  it('returns false for a decimal', () => {
    expect(isValidQuantity(2.5)).toBe(false);
  });
});
