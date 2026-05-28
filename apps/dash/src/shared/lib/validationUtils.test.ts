import { describe, it, expect } from 'vitest';
import {
  isValidEmail,
  isValidUrl,
  isValidPhone,
  requireField,
  minLength,
  maxLength,
  numberRange,
  collectErrors,
  isValid,
} from './validationUtils';

describe('isValidEmail', () => {
  it('accepts a valid email', () => {
    expect(isValidEmail('user@example.com')).toBe(true);
  });

  it('accepts emails with subdomains', () => {
    expect(isValidEmail('user@mail.example.com')).toBe(true);
  });

  it('rejects missing @', () => {
    expect(isValidEmail('userexample.com')).toBe(false);
  });

  it('rejects missing domain', () => {
    expect(isValidEmail('user@')).toBe(false);
  });

  it('rejects empty string', () => {
    expect(isValidEmail('')).toBe(false);
  });

  it('rejects email with spaces', () => {
    expect(isValidEmail('user @example.com')).toBe(false);
  });
});

describe('isValidUrl', () => {
  it('accepts http URLs', () => {
    expect(isValidUrl('http://example.com')).toBe(true);
  });

  it('accepts https URLs', () => {
    expect(isValidUrl('https://example.com/path?q=1')).toBe(true);
  });

  it('rejects plain text', () => {
    expect(isValidUrl('not a url')).toBe(false);
  });

  it('rejects ftp URLs', () => {
    expect(isValidUrl('ftp://example.com')).toBe(false);
  });

  it('rejects empty string', () => {
    expect(isValidUrl('')).toBe(false);
  });
});

describe('isValidPhone', () => {
  it('accepts a standard phone number', () => {
    expect(isValidPhone('+1 (555) 123-4567')).toBe(true);
  });

  it('accepts a simple digit sequence', () => {
    expect(isValidPhone('5551234567')).toBe(true);
  });

  it('rejects a phone that is too short', () => {
    expect(isValidPhone('12345')).toBe(false);
  });

  it('rejects letters', () => {
    expect(isValidPhone('abc-def-ghij')).toBe(false);
  });
});

describe('requireField', () => {
  it('returns null for a non-empty string', () => {
    expect(requireField('hello')).toBeNull();
  });

  it('returns an error message for empty string', () => {
    expect(requireField('')).not.toBeNull();
  });

  it('returns an error message for whitespace-only string', () => {
    expect(requireField('   ')).not.toBeNull();
  });

  it('includes the field name in the error', () => {
    expect(requireField('', 'Email')).toContain('Email');
  });
});

describe('minLength', () => {
  it('returns null when length is sufficient', () => {
    expect(minLength('hello', 3)).toBeNull();
  });

  it('returns null when length equals minimum', () => {
    expect(minLength('hi', 2)).toBeNull();
  });

  it('returns an error when too short', () => {
    expect(minLength('hi', 5)).not.toBeNull();
  });
});

describe('maxLength', () => {
  it('returns null when within limit', () => {
    expect(maxLength('hello', 10)).toBeNull();
  });

  it('returns null when at the limit', () => {
    expect(maxLength('hello', 5)).toBeNull();
  });

  it('returns an error when too long', () => {
    expect(maxLength('hello world', 5)).not.toBeNull();
  });
});

describe('numberRange', () => {
  it('returns null when within range', () => {
    expect(numberRange(5, 0, 10)).toBeNull();
  });

  it('returns error when below minimum', () => {
    expect(numberRange(-1, 0, 10)).not.toBeNull();
  });

  it('returns error when above maximum', () => {
    expect(numberRange(11, 0, 10)).not.toBeNull();
  });

  it('returns null at the boundaries', () => {
    expect(numberRange(0, 0, 10)).toBeNull();
    expect(numberRange(10, 0, 10)).toBeNull();
  });
});

describe('collectErrors', () => {
  it('returns only non-null errors', () => {
    const errors = collectErrors(null, 'Error 1', null, 'Error 2');
    expect(errors).toEqual(['Error 1', 'Error 2']);
  });

  it('returns empty array when all validators pass', () => {
    expect(collectErrors(null, null, null)).toEqual([]);
  });
});

describe('isValid', () => {
  it('returns true when all validators pass', () => {
    expect(isValid(null, null, null)).toBe(true);
  });

  it('returns false when any validator fails', () => {
    expect(isValid(null, 'error', null)).toBe(false);
  });

  it('returns true for no validators', () => {
    expect(isValid()).toBe(true);
  });
});
