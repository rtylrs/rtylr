import { describe, it, expect } from 'vitest';
import {
  validateName,
  validatePhone,
  validateEmail,
  validateCheckoutFields,
} from './validateCheckout';

describe('validateName', () => {
  it('returns valid for a non-empty name', () => {
    expect(validateName('John').valid).toBe(true);
  });

  it('returns invalid for an empty string', () => {
    const result = validateName('');
    expect(result.valid).toBe(false);
    expect(result.error).toBeTruthy();
  });

  it('returns invalid for a whitespace-only string', () => {
    expect(validateName('   ').valid).toBe(false);
  });
});

describe('validatePhone', () => {
  it('returns valid for a non-empty phone number', () => {
    expect(validatePhone('+1 234 567 8900').valid).toBe(true);
  });

  it('returns invalid for an empty string', () => {
    const result = validatePhone('');
    expect(result.valid).toBe(false);
    expect(result.error).toBeTruthy();
  });

  it('returns invalid for whitespace only', () => {
    expect(validatePhone('  ').valid).toBe(false);
  });
});

describe('validateEmail', () => {
  it('returns valid for a well-formed email', () => {
    expect(validateEmail('user@example.com').valid).toBe(true);
  });

  it('returns valid for an empty email (optional field)', () => {
    expect(validateEmail('').valid).toBe(true);
  });

  it('returns invalid for a string without @', () => {
    const result = validateEmail('notanemail');
    expect(result.valid).toBe(false);
    expect(result.error).toBeTruthy();
  });

  it('returns valid for whitespace-only email (treated as empty)', () => {
    expect(validateEmail('   ').valid).toBe(true);
  });
});

describe('validateCheckoutFields', () => {
  it('returns null when all required fields are valid', () => {
    const result = validateCheckoutFields({ name: 'Alice', phone: '555-0100' });
    expect(result).toBeNull();
  });

  it('returns a name error when name is missing', () => {
    const result = validateCheckoutFields({ name: '', phone: '555-0100' });
    expect(result).toBeTruthy();
    expect(result).toContain('name');
  });

  it('returns a phone error when phone is missing', () => {
    const result = validateCheckoutFields({ name: 'Alice', phone: '' });
    expect(result).toBeTruthy();
    expect(result).toContain('phone');
  });

  it('returns an email error when email is invalid', () => {
    const result = validateCheckoutFields({
      name: 'Alice',
      phone: '555-0100',
      email: 'bad-email',
    });
    expect(result).toBeTruthy();
    expect(result).toContain('email');
  });

  it('returns null when email is omitted', () => {
    const result = validateCheckoutFields({ name: 'Alice', phone: '555-0100' });
    expect(result).toBeNull();
  });

  it('returns null when email is empty string', () => {
    const result = validateCheckoutFields({ name: 'Alice', phone: '555-0100', email: '' });
    expect(result).toBeNull();
  });
});
