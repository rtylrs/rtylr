import { describe, expect, it } from 'vitest';
import {
  isDateRangeValid,
  isStartDateValid,
  isValidEmail,
  isValidHourlyRate,
  isValidIsoDate,
  isValidName,
  isValidPhone,
  isValidPin,
  isValidSalary,
  validateLeaveRequest,
} from './hrValidationUtils';

describe('isValidEmail', () => {
  it('returns true for valid emails', () => {
    expect(isValidEmail('user@example.com')).toBe(true);
    expect(isValidEmail('user.name+tag@company.org')).toBe(true);
  });

  it('returns false for invalid emails', () => {
    expect(isValidEmail('notanemail')).toBe(false);
    expect(isValidEmail('@nodomain')).toBe(false);
    expect(isValidEmail('no@')).toBe(false);
  });
});

describe('isValidPhone', () => {
  it('returns true for valid phone numbers', () => {
    expect(isValidPhone('+1 555-123-4567')).toBe(true);
    expect(isValidPhone('5551234567')).toBe(true);
  });

  it('returns false for invalid phones', () => {
    expect(isValidPhone('123')).toBe(false);
    expect(isValidPhone('abc')).toBe(false);
  });
});

describe('isValidPin', () => {
  it('validates 4-digit PIN by default', () => {
    expect(isValidPin('1234')).toBe(true);
    expect(isValidPin('0000')).toBe(true);
  });

  it('validates custom length PIN', () => {
    expect(isValidPin('123456', 6)).toBe(true);
    expect(isValidPin('12345', 6)).toBe(false);
  });

  it('returns false for non-numeric PINs', () => {
    expect(isValidPin('abcd')).toBe(false);
    expect(isValidPin('12 4')).toBe(false);
  });
});

describe('isValidName', () => {
  it('returns true for valid names', () => {
    expect(isValidName('John Doe')).toBe(true);
    expect(isValidName("O'Brien")).toBe(true);
    expect(isValidName('Mary-Jane')).toBe(true);
  });

  it('returns false for invalid names', () => {
    expect(isValidName('J')).toBe(false);
    expect(isValidName('John123')).toBe(false);
    expect(isValidName('')).toBe(false);
  });
});

describe('isValidIsoDate', () => {
  it('returns true for valid ISO dates', () => {
    expect(isValidIsoDate('2024-01-15')).toBe(true);
    expect(isValidIsoDate('2000-12-31')).toBe(true);
  });

  it('returns false for invalid dates', () => {
    expect(isValidIsoDate('2024-13-01')).toBe(false);
    expect(isValidIsoDate('not-a-date')).toBe(false);
    expect(isValidIsoDate('2024/01/15')).toBe(false);
  });
});

describe('isValidHourlyRate', () => {
  it('returns true for valid rates', () => {
    expect(isValidHourlyRate(15)).toBe(true);
    expect(isValidHourlyRate(100)).toBe(true);
  });

  it('returns false for invalid rates', () => {
    expect(isValidHourlyRate(0)).toBe(false);
    expect(isValidHourlyRate(-10)).toBe(false);
    expect(isValidHourlyRate(10001)).toBe(false);
  });
});

describe('isValidSalary', () => {
  it('returns true for valid salaries', () => {
    expect(isValidSalary(50000)).toBe(true);
    expect(isValidSalary(1)).toBe(true);
  });

  it('returns false for invalid salaries', () => {
    expect(isValidSalary(0)).toBe(false);
    expect(isValidSalary(-1000)).toBe(false);
    expect(isValidSalary(10_000_001)).toBe(false);
  });
});

describe('isDateRangeValid', () => {
  it('returns true when start is before or equal to end', () => {
    expect(isDateRangeValid('2024-01-01', '2024-01-15')).toBe(true);
    expect(isDateRangeValid('2024-01-15', '2024-01-15')).toBe(true);
  });

  it('returns false when start is after end', () => {
    expect(isDateRangeValid('2024-02-01', '2024-01-15')).toBe(false);
  });
});

describe('isStartDateValid', () => {
  it('returns true when start date is today or in the future', () => {
    expect(isStartDateValid('2030-01-01', '2024-01-01')).toBe(true);
    expect(isStartDateValid('2024-01-01', '2024-01-01')).toBe(true);
  });

  it('returns false when start date is in the past', () => {
    expect(isStartDateValid('2020-01-01', '2024-01-01')).toBe(false);
  });
});

describe('validateLeaveRequest', () => {
  it('returns null for valid leave request', () => {
    expect(validateLeaveRequest('2024-02-01', '2024-02-05', 10, 5)).toBeNull();
  });

  it('returns error when end is before start', () => {
    expect(validateLeaveRequest('2024-02-05', '2024-02-01', 10, 5)).not.toBeNull();
  });

  it('returns error when days exceed balance', () => {
    expect(validateLeaveRequest('2024-02-01', '2024-02-10', 3, 5)).not.toBeNull();
  });

  it('returns error when days is 0', () => {
    expect(validateLeaveRequest('2024-02-01', '2024-02-05', 10, 0)).not.toBeNull();
  });
});
