import { describe, it, expect } from 'vitest';

import {
  isValidName,
  isValidTitle,
  isValidIsoDate,
  isValidDateRange,
  isValidEstimate,
  isValidHexColor,
  isValidEmail,
  isValidProgress,
  nameValidationError,
  titleValidationError,
} from './validationUtils';

describe('isValidName', () => {
  it('accepts normal names', () => {
    expect(isValidName('My Project')).toBe(true);
  });

  it('rejects empty name', () => {
    expect(isValidName('')).toBe(false);
    expect(isValidName('   ')).toBe(false);
  });

  it('rejects names over 100 chars', () => {
    expect(isValidName('a'.repeat(101))).toBe(false);
  });
});

describe('isValidTitle', () => {
  it('accepts valid title', () => {
    expect(isValidTitle('Fix the login bug')).toBe(true);
  });

  it('rejects empty title', () => {
    expect(isValidTitle('')).toBe(false);
  });

  it('rejects title over 255 chars', () => {
    expect(isValidTitle('x'.repeat(256))).toBe(false);
  });
});

describe('isValidIsoDate', () => {
  it('accepts valid ISO dates', () => {
    expect(isValidIsoDate('2026-01-01')).toBe(true);
    expect(isValidIsoDate('2026-12-31')).toBe(true);
  });

  it('rejects invalid dates', () => {
    expect(isValidIsoDate('not-a-date')).toBe(false);
    expect(isValidIsoDate('2026-13-01')).toBe(false);
  });
});

describe('isValidDateRange', () => {
  it('accepts valid range', () => {
    expect(isValidDateRange('2026-01-01', '2026-12-31')).toBe(true);
  });

  it('accepts equal dates', () => {
    expect(isValidDateRange('2026-06-01', '2026-06-01')).toBe(true);
  });

  it('rejects reversed range', () => {
    expect(isValidDateRange('2026-12-31', '2026-01-01')).toBe(false);
  });
});

describe('isValidEstimate', () => {
  it('accepts positive numbers', () => {
    expect(isValidEstimate(2.5)).toBe(true);
    expect(isValidEstimate(1)).toBe(true);
  });

  it('rejects zero and negatives', () => {
    expect(isValidEstimate(0)).toBe(false);
    expect(isValidEstimate(-1)).toBe(false);
  });

  it('rejects non-numbers', () => {
    expect(isValidEstimate('2')).toBe(false);
    expect(isValidEstimate(NaN)).toBe(false);
  });
});

describe('isValidHexColor', () => {
  it('accepts valid hex colors', () => {
    expect(isValidHexColor('#3b82f6')).toBe(true);
    expect(isValidHexColor('#FFFFFF')).toBe(true);
  });

  it('rejects invalid hex colors', () => {
    expect(isValidHexColor('3b82f6')).toBe(false);
    expect(isValidHexColor('#fff')).toBe(false);
  });
});

describe('isValidEmail', () => {
  it('accepts valid emails', () => {
    expect(isValidEmail('user@example.com')).toBe(true);
  });

  it('rejects invalid emails', () => {
    expect(isValidEmail('not-an-email')).toBe(false);
    expect(isValidEmail('@example.com')).toBe(false);
  });
});

describe('isValidProgress', () => {
  it('accepts 0 to 100', () => {
    expect(isValidProgress(0)).toBe(true);
    expect(isValidProgress(50)).toBe(true);
    expect(isValidProgress(100)).toBe(true);
  });

  it('rejects out of range', () => {
    expect(isValidProgress(-1)).toBe(false);
    expect(isValidProgress(101)).toBe(false);
  });

  it('rejects non-integers', () => {
    expect(isValidProgress(50.5)).toBe(false);
  });
});

describe('nameValidationError', () => {
  it('returns null for valid name', () => {
    expect(nameValidationError('My Project')).toBeNull();
  });

  it('returns error for empty name', () => {
    expect(nameValidationError('')).toBe('Name is required');
  });

  it('returns error for long name', () => {
    expect(nameValidationError('a'.repeat(101))).toBe('Name must be 100 characters or fewer');
  });
});

describe('titleValidationError', () => {
  it('returns null for valid title', () => {
    expect(titleValidationError('Fix bug')).toBeNull();
  });

  it('returns error for empty title', () => {
    expect(titleValidationError('')).toBe('Title is required');
  });
});
