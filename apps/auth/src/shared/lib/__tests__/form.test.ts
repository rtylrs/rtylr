import { describe, it, expect } from 'vitest';
import {
  isFormValid,
  fieldError,
  hasFieldError,
  clearFieldError,
  invalidFields,
  mergeErrors,
} from '../form';

const errors = { email: 'Required', password: '', username: 'Too short' };

describe('isFormValid', () => {
  it('returns false when there are errors', () => {
    expect(isFormValid(errors)).toBe(false);
  });

  it('returns true when all fields are valid', () => {
    expect(isFormValid({ email: '', password: '' })).toBe(true);
  });

  it('returns true for empty errors object', () => {
    expect(isFormValid({})).toBe(true);
  });
});

describe('fieldError', () => {
  it('returns the error message for a field', () => {
    expect(fieldError(errors, 'email')).toBe('Required');
  });

  it('returns empty string for field with no error', () => {
    expect(fieldError(errors, 'password')).toBe('');
  });

  it('returns empty string for missing field', () => {
    expect(fieldError(errors, 'nonexistent')).toBe('');
  });
});

describe('hasFieldError', () => {
  it('returns true for field with error', () => {
    expect(hasFieldError(errors, 'email')).toBe(true);
  });

  it('returns false for field without error', () => {
    expect(hasFieldError(errors, 'password')).toBe(false);
  });
});

describe('clearFieldError', () => {
  it('removes the specified field error', () => {
    const cleared = clearFieldError(errors, 'email');
    expect(cleared.email).toBeUndefined();
    expect(cleared.username).toBe('Too short');
  });

  it('does not mutate original', () => {
    clearFieldError(errors, 'email');
    expect(errors.email).toBe('Required');
  });
});

describe('invalidFields', () => {
  it('returns only fields with errors', () => {
    expect(invalidFields(errors).sort()).toEqual(['email', 'username']);
  });
});

describe('mergeErrors', () => {
  it('merges two error objects', () => {
    const a = { email: 'Required' };
    const b = { password: 'Too short' };
    expect(mergeErrors(a, b)).toEqual({ email: 'Required', password: 'Too short' });
  });

  it('second takes priority on conflict', () => {
    const merged = mergeErrors({ email: 'old' }, { email: 'new' });
    expect(merged.email).toBe('new');
  });
});
