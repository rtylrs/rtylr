import { describe, it, expect } from 'vitest';
import { digitsOnly, isValidPhoneLength, formatUsPhone, toE164, maskPhone } from './phoneUtils';

describe('digitsOnly', () => {
  it('strips dashes and spaces', () => {
    expect(digitsOnly('555-867-5309')).toBe('5558675309');
  });

  it('strips parentheses', () => {
    expect(digitsOnly('(555) 867-5309')).toBe('5558675309');
  });

  it('returns only digits from a mixed string', () => {
    expect(digitsOnly('+1 (800) 555-0100')).toBe('18005550100');
  });

  it('returns an empty string for a non-numeric input', () => {
    expect(digitsOnly('abc')).toBe('');
  });
});

describe('isValidPhoneLength', () => {
  it('returns true for a 10-digit US number', () => {
    expect(isValidPhoneLength('5558675309')).toBe(true);
  });

  it('returns true for a 7-digit number', () => {
    expect(isValidPhoneLength('5550100')).toBe(true);
  });

  it('returns false for a 5-digit number', () => {
    expect(isValidPhoneLength('12345')).toBe(false);
  });

  it('returns false for an empty string', () => {
    expect(isValidPhoneLength('')).toBe(false);
  });

  it('strips formatting before counting', () => {
    expect(isValidPhoneLength('(555) 867-5309')).toBe(true);
  });
});

describe('formatUsPhone', () => {
  it('formats a 10-digit string', () => {
    expect(formatUsPhone('5558675309')).toBe('(555) 867-5309');
  });

  it('formats a pre-formatted 10-digit string', () => {
    expect(formatUsPhone('(555) 867-5309')).toBe('(555) 867-5309');
  });

  it('returns the original string when digits are not 10', () => {
    expect(formatUsPhone('12345')).toBe('12345');
  });
});

describe('toE164', () => {
  it('prepends + and strips formatting', () => {
    expect(toE164('+1 (555) 867-5309')).toBe('+15558675309');
  });

  it('works with a plain digit string', () => {
    expect(toE164('15558675309')).toBe('+15558675309');
  });
});

describe('maskPhone', () => {
  it('shows only the last four digits', () => {
    expect(maskPhone('5558675309')).toBe('******5309');
  });

  it('returns **** for very short inputs', () => {
    expect(maskPhone('123')).toBe('****');
  });

  it('handles formatted phone strings', () => {
    const masked = maskPhone('(555) 867-5309');
    expect(masked.endsWith('5309')).toBe(true);
  });
});
