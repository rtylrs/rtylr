import { describe, it, expect } from 'vitest';
import {
  parseNonNegativeNumber,
  parsePositiveNumber,
  parsePositiveInteger,
  clampInput,
  decimalPlaces,
  keypadDisplay,
} from './numericInput';

describe('parseNonNegativeNumber', () => {
  it('parses a positive decimal string', () => {
    expect(parseNonNegativeNumber('12.50')).toBe(12.5);
  });

  it('parses zero', () => {
    expect(parseNonNegativeNumber('0')).toBe(0);
  });

  it('returns null for a negative value', () => {
    expect(parseNonNegativeNumber('-5')).toBeNull();
  });

  it('returns null for a non-numeric string', () => {
    expect(parseNonNegativeNumber('abc')).toBeNull();
  });

  it('returns null for an empty string', () => {
    expect(parseNonNegativeNumber('')).toBeNull();
  });

  it('trims whitespace before parsing', () => {
    expect(parseNonNegativeNumber('  10  ')).toBe(10);
  });

  it('returns null for Infinity', () => {
    expect(parseNonNegativeNumber('Infinity')).toBeNull();
  });
});

describe('parsePositiveNumber', () => {
  it('parses a positive decimal string', () => {
    expect(parsePositiveNumber('5.99')).toBe(5.99);
  });

  it('returns null for zero', () => {
    expect(parsePositiveNumber('0')).toBeNull();
  });

  it('returns null for a negative value', () => {
    expect(parsePositiveNumber('-1')).toBeNull();
  });

  it('returns null for a non-numeric string', () => {
    expect(parsePositiveNumber('abc')).toBeNull();
  });
});

describe('parsePositiveInteger', () => {
  it('parses a positive integer string', () => {
    expect(parsePositiveInteger('42')).toBe(42);
  });

  it('floors a decimal string', () => {
    expect(parsePositiveInteger('9.7')).toBe(9);
  });

  it('returns null for zero', () => {
    expect(parsePositiveInteger('0')).toBeNull();
  });

  it('returns null for a negative value', () => {
    expect(parsePositiveInteger('-5')).toBeNull();
  });

  it('returns null for a non-numeric string', () => {
    expect(parsePositiveInteger('xyz')).toBeNull();
  });

  it('returns null for an empty string', () => {
    expect(parsePositiveInteger('')).toBeNull();
  });
});

describe('clampInput', () => {
  it('returns the value when within range', () => {
    expect(clampInput('5', 0, 10)).toBe(5);
  });

  it('clamps to the minimum', () => {
    expect(clampInput('-1', 0, 10)).toBe(0);
  });

  it('clamps to the maximum', () => {
    expect(clampInput('15', 0, 10)).toBe(10);
  });

  it('returns null for a non-numeric string', () => {
    expect(clampInput('abc', 0, 100)).toBeNull();
  });

  it('handles the boundary values', () => {
    expect(clampInput('0', 0, 100)).toBe(0);
    expect(clampInput('100', 0, 100)).toBe(100);
  });
});

describe('decimalPlaces', () => {
  it('returns 0 for an integer string', () => {
    expect(decimalPlaces('42')).toBe(0);
  });

  it('returns 2 for a 2-decimal string', () => {
    expect(decimalPlaces('12.50')).toBe(2);
  });

  it('returns 1 for a 1-decimal string', () => {
    expect(decimalPlaces('9.5')).toBe(1);
  });

  it('returns 0 for an empty string', () => {
    expect(decimalPlaces('')).toBe(0);
  });
});

describe('keypadDisplay', () => {
  it('displays an integer without decimal point', () => {
    expect(keypadDisplay(5)).toBe('5');
  });

  it('displays a decimal number', () => {
    expect(keypadDisplay(5.5)).toBe('5.5');
  });

  it('displays 0 for non-finite input', () => {
    expect(keypadDisplay(NaN)).toBe('0');
    expect(keypadDisplay(Infinity)).toBe('0');
  });

  it('rounds to 2 decimal places', () => {
    expect(keypadDisplay(5.999)).toBe('6');
  });
});
