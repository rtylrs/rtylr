import { describe, expect, it } from 'vitest';

import {
  isAlpha,
  isAlphanumeric,
  isLengthBetween,
  isMatch,
  isNumericInRange,
  isPostalCode,
  isRequired,
  isStrongPassword,
  isValidEmail,
  isValidPhone,
  isValidUrl,
  passwordErrors,
} from './validationUtils';

describe('isValidEmail', () => {
  it('accepts a standard email', () => {
    expect(isValidEmail('user@example.com')).toBe(true);
  });

  it('accepts email with subdomain', () => {
    expect(isValidEmail('user@mail.example.com')).toBe(true);
  });

  it('rejects email without at sign', () => {
    expect(isValidEmail('userexample.com')).toBe(false);
  });

  it('rejects email without domain', () => {
    expect(isValidEmail('user@')).toBe(false);
  });

  it('rejects empty string', () => {
    expect(isValidEmail('')).toBe(false);
  });
});

describe('isValidPhone', () => {
  it('accepts a 10-digit US number', () => {
    expect(isValidPhone('5551234567')).toBe(true);
  });

  it('accepts formatted number with dashes', () => {
    expect(isValidPhone('555-123-4567')).toBe(true);
  });

  it('rejects a 4-digit number', () => {
    expect(isValidPhone('1234')).toBe(false);
  });

  it('accepts international numbers', () => {
    expect(isValidPhone('+44 20 7946 0958')).toBe(true);
  });
});

describe('isValidUrl', () => {
  it('accepts https URL', () => {
    expect(isValidUrl('https://rtylr.com')).toBe(true);
  });

  it('accepts http URL', () => {
    expect(isValidUrl('http://example.com/path')).toBe(true);
  });

  it('rejects plain text', () => {
    expect(isValidUrl('not a url')).toBe(false);
  });

  it('rejects ftp protocol', () => {
    expect(isValidUrl('ftp://example.com')).toBe(false);
  });
});

describe('isRequired', () => {
  it('returns true for non-empty string', () => {
    expect(isRequired('hello')).toBe(true);
  });

  it('returns false for empty string', () => {
    expect(isRequired('')).toBe(false);
  });

  it('returns false for whitespace-only string', () => {
    expect(isRequired('   ')).toBe(false);
  });
});

describe('isLengthBetween', () => {
  it('returns true when length is in range', () => {
    expect(isLengthBetween('hello', 3, 10)).toBe(true);
  });

  it('returns false when too short', () => {
    expect(isLengthBetween('hi', 3, 10)).toBe(false);
  });

  it('returns false when too long', () => {
    expect(isLengthBetween('hello world!', 3, 10)).toBe(false);
  });
});

describe('isAlpha', () => {
  it('returns true for letters and spaces', () => {
    expect(isAlpha('Hello World')).toBe(true);
  });

  it('returns false when digits are present', () => {
    expect(isAlpha('Hello1')).toBe(false);
  });
});

describe('isAlphanumeric', () => {
  it('returns true for letters and digits', () => {
    expect(isAlphanumeric('abc123')).toBe(true);
  });

  it('returns false when special chars are present', () => {
    expect(isAlphanumeric('abc-123')).toBe(false);
  });
});

describe('isNumericInRange', () => {
  it('returns true for value in range', () => {
    expect(isNumericInRange(5, 1, 10)).toBe(true);
  });

  it('returns false for value below min', () => {
    expect(isNumericInRange(0, 1, 10)).toBe(false);
  });

  it('returns false for Infinity', () => {
    expect(isNumericInRange(Infinity)).toBe(false);
  });

  it('returns true with no bounds', () => {
    expect(isNumericInRange(999)).toBe(true);
  });
});

describe('isMatch', () => {
  it('returns true when strings are equal', () => {
    expect(isMatch('password', 'password')).toBe(true);
  });

  it('returns false when strings differ', () => {
    expect(isMatch('password', 'Password')).toBe(false);
  });
});

describe('isPostalCode', () => {
  it('accepts 5-digit US zip', () => {
    expect(isPostalCode('90210')).toBe(true);
  });

  it('accepts zip+4 format', () => {
    expect(isPostalCode('90210-1234')).toBe(true);
  });

  it('rejects non-zip string', () => {
    expect(isPostalCode('ABCDE')).toBe(false);
  });
});

describe('isStrongPassword', () => {
  it('returns true for a strong password', () => {
    expect(isStrongPassword('Str0ngPwd')).toBe(true);
  });

  it('returns false for a short password', () => {
    expect(isStrongPassword('Ab1')).toBe(false);
  });

  it('returns false for all lowercase', () => {
    expect(isStrongPassword('alllower1')).toBe(false);
  });
});

describe('passwordErrors', () => {
  it('returns no errors for strong password', () => {
    expect(passwordErrors('Str0ngPass')).toHaveLength(0);
  });

  it('reports missing uppercase', () => {
    expect(passwordErrors('weakpass1')).toContain('At least one uppercase letter required');
  });

  it('reports too short', () => {
    expect(passwordErrors('Ab1')).toContain('At least 8 characters required');
  });

  it('reports multiple errors', () => {
    const errors = passwordErrors('weak');
    expect(errors.length).toBeGreaterThan(1);
  });
});
