import { describe, it, expect } from 'vitest';
import {
  isEmail,
  isStrongPassword,
  passwordsMatch,
  isPhoneNumber,
  isNotEmpty,
  getEmailError,
  getPasswordMatchError,
} from '../validation';

describe('isEmail', () => {
  it('accepts a standard email', () => {
    expect(isEmail('user@example.com')).toBe(true);
  });

  it('accepts email with subdomain', () => {
    expect(isEmail('user@mail.example.com')).toBe(true);
  });

  it('rejects empty string', () => {
    expect(isEmail('')).toBe(false);
  });

  it('rejects missing domain', () => {
    expect(isEmail('user@')).toBe(false);
  });

  it('rejects missing at-sign', () => {
    expect(isEmail('userexample.com')).toBe(false);
  });

  it('trims surrounding whitespace before checking', () => {
    expect(isEmail('  user@example.com  ')).toBe(true);
  });
});

describe('isStrongPassword', () => {
  it('accepts password with upper, lower and digit', () => {
    expect(isStrongPassword('Password1')).toBe(true);
  });

  it('rejects password shorter than 8 chars', () => {
    expect(isStrongPassword('Pass1')).toBe(false);
  });

  it('rejects password without uppercase', () => {
    expect(isStrongPassword('password1')).toBe(false);
  });

  it('rejects password without digit', () => {
    expect(isStrongPassword('Password')).toBe(false);
  });

  it('rejects empty string', () => {
    expect(isStrongPassword('')).toBe(false);
  });
});

describe('passwordsMatch', () => {
  it('returns true when both values are identical', () => {
    expect(passwordsMatch('secret', 'secret')).toBe(true);
  });

  it('returns false when values differ', () => {
    expect(passwordsMatch('secret', 'Secret')).toBe(false);
  });

  it('returns false when first arg is empty', () => {
    expect(passwordsMatch('', 'secret')).toBe(false);
  });
});

describe('isPhoneNumber', () => {
  it('accepts a standard phone number', () => {
    expect(isPhoneNumber('+1 555-123-4567')).toBe(true);
  });

  it('rejects empty string', () => {
    expect(isPhoneNumber('')).toBe(false);
  });

  it('rejects letters', () => {
    expect(isPhoneNumber('abcdef')).toBe(false);
  });
});

describe('isNotEmpty', () => {
  it('returns true for a non-empty string', () => {
    expect(isNotEmpty('hello')).toBe(true);
  });

  it('returns false for an empty string', () => {
    expect(isNotEmpty('')).toBe(false);
  });

  it('returns false for a whitespace-only string', () => {
    expect(isNotEmpty('   ')).toBe(false);
  });
});

describe('getEmailError', () => {
  it('returns undefined for empty value', () => {
    expect(getEmailError('')).toBeUndefined();
  });

  it('returns undefined for a valid email', () => {
    expect(getEmailError('user@example.com')).toBeUndefined();
  });

  it('returns an error string for an invalid email', () => {
    expect(getEmailError('not-an-email')).toBe('Please enter a valid email address');
  });
});

describe('getPasswordMatchError', () => {
  it('returns undefined when confirm is empty', () => {
    expect(getPasswordMatchError('secret', '')).toBeUndefined();
  });

  it('returns undefined when passwords match', () => {
    expect(getPasswordMatchError('secret', 'secret')).toBeUndefined();
  });

  it('returns an error string when passwords differ', () => {
    expect(getPasswordMatchError('secret', 'different')).toBe('Passwords do not match');
  });
});
