import { describe, it, expect } from 'vitest';
import {
  isValidEmail,
  isValidHireDate,
  isProbationEndAfterHire,
  isValidLeaveRange,
  isValidHourlyRate,
  isValidMonthlySalary,
  isValidCurrencyCode,
  isValidPin,
  isValidName,
  validateLeaveRange,
} from './hrValidators';

describe('isValidEmail', () => {
  it('accepts a valid email', () => {
    expect(isValidEmail('alice@example.com')).toBe(true);
  });

  it('rejects email without @', () => {
    expect(isValidEmail('invalidemail.com')).toBe(false);
  });

  it('rejects empty string', () => {
    expect(isValidEmail('')).toBe(false);
  });
});

describe('isValidHireDate', () => {
  it('accepts a past date', () => {
    expect(isValidHireDate('2020-01-15')).toBe(true);
  });

  it('rejects a future date', () => {
    expect(isValidHireDate('2099-12-31')).toBe(false);
  });

  it('rejects an invalid date string', () => {
    expect(isValidHireDate('not-a-date')).toBe(false);
  });
});

describe('isProbationEndAfterHire', () => {
  it('returns true when probation end is after hire', () => {
    expect(isProbationEndAfterHire('2026-01-01', '2026-04-01')).toBe(true);
  });

  it('returns false when probation end equals hire date', () => {
    expect(isProbationEndAfterHire('2026-01-01', '2026-01-01')).toBe(false);
  });

  it('returns false when probation end is before hire', () => {
    expect(isProbationEndAfterHire('2026-03-01', '2026-01-01')).toBe(false);
  });
});

describe('isValidLeaveRange', () => {
  it('returns true when end equals start', () => {
    expect(isValidLeaveRange('2026-06-01', '2026-06-01')).toBe(true);
  });

  it('returns true when end is after start', () => {
    expect(isValidLeaveRange('2026-06-01', '2026-06-05')).toBe(true);
  });

  it('returns false when end is before start', () => {
    expect(isValidLeaveRange('2026-06-05', '2026-06-01')).toBe(false);
  });
});

describe('isValidHourlyRate', () => {
  it('accepts a positive rate', () => {
    expect(isValidHourlyRate(25)).toBe(true);
  });

  it('rejects zero', () => {
    expect(isValidHourlyRate(0)).toBe(false);
  });

  it('rejects negative', () => {
    expect(isValidHourlyRate(-10)).toBe(false);
  });

  it('rejects Infinity', () => {
    expect(isValidHourlyRate(Infinity)).toBe(false);
  });

  it('rejects NaN', () => {
    expect(isValidHourlyRate(NaN)).toBe(false);
  });
});

describe('isValidMonthlySalary', () => {
  it('accepts a valid salary', () => {
    expect(isValidMonthlySalary(5000)).toBe(true);
  });

  it('rejects zero', () => {
    expect(isValidMonthlySalary(0)).toBe(false);
  });

  it('rejects negative', () => {
    expect(isValidMonthlySalary(-500)).toBe(false);
  });
});

describe('isValidCurrencyCode', () => {
  it('accepts USD', () => {
    expect(isValidCurrencyCode('USD')).toBe(true);
  });

  it('rejects lowercase', () => {
    expect(isValidCurrencyCode('usd')).toBe(false);
  });

  it('rejects 2-letter code', () => {
    expect(isValidCurrencyCode('US')).toBe(false);
  });

  it('rejects empty string', () => {
    expect(isValidCurrencyCode('')).toBe(false);
  });
});

describe('isValidPin', () => {
  it('accepts a 4-digit PIN', () => {
    expect(isValidPin('1234')).toBe(true);
  });

  it('rejects 3-digit PIN', () => {
    expect(isValidPin('123')).toBe(false);
  });

  it('rejects PIN with letters', () => {
    expect(isValidPin('12ab')).toBe(false);
  });

  it('rejects empty string', () => {
    expect(isValidPin('')).toBe(false);
  });
});

describe('isValidName', () => {
  it('accepts a normal name', () => {
    expect(isValidName('Alice Smith')).toBe(true);
  });

  it('rejects empty string', () => {
    expect(isValidName('')).toBe(false);
  });

  it('rejects whitespace-only string', () => {
    expect(isValidName('   ')).toBe(false);
  });
});

describe('validateLeaveRange', () => {
  it('returns null for a valid range', () => {
    expect(validateLeaveRange('2026-06-01', '2026-06-05')).toBeNull();
  });

  it('returns error when dates are missing', () => {
    expect(validateLeaveRange('', '')).toMatch(/required/i);
  });

  it('returns error when end is before start', () => {
    expect(validateLeaveRange('2026-06-10', '2026-06-05')).toMatch(/end date/i);
  });
});
