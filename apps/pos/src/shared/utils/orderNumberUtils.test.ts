import { describe, it, expect } from 'vitest';
import {
  padOrderNumber,
  formatOrderNumber,
  parseOrderNumber,
  isSameOrder,
  orderNumberFromTimestamp,
  isValidOrderNumberString,
} from './orderNumberUtils';

describe('padOrderNumber', () => {
  it('pads a number to the default width of 4', () => {
    expect(padOrderNumber(42)).toBe('0042');
  });

  it('pads to a custom width', () => {
    expect(padOrderNumber(5, 6)).toBe('000005');
  });

  it('does not truncate numbers exceeding the width', () => {
    expect(padOrderNumber(12345, 4)).toBe('12345');
  });

  it('returns all zeros for non-finite input', () => {
    expect(padOrderNumber(NaN)).toBe('0000');
    expect(padOrderNumber(-1)).toBe('0000');
  });
});

describe('formatOrderNumber', () => {
  it('formats with the default prefix and width', () => {
    expect(formatOrderNumber(42)).toBe('#0042');
  });

  it('uses a custom prefix', () => {
    expect(formatOrderNumber(7, 'ORD-')).toBe('ORD-0007');
  });
});

describe('parseOrderNumber', () => {
  it('parses a plain numeric string', () => {
    expect(parseOrderNumber('0042')).toBe(42);
  });

  it('parses a formatted string with prefix', () => {
    expect(parseOrderNumber('#0042')).toBe(42);
  });

  it('returns null for a non-numeric string', () => {
    expect(parseOrderNumber('abc')).toBeNull();
  });

  it('returns null for an empty string', () => {
    expect(parseOrderNumber('')).toBeNull();
  });
});

describe('isSameOrder', () => {
  it('returns true for two formatted strings representing the same order', () => {
    expect(isSameOrder('#0042', '0042')).toBe(true);
  });

  it('returns false for different order numbers', () => {
    expect(isSameOrder('#0042', '#0043')).toBe(false);
  });

  it('returns false when either string is invalid', () => {
    expect(isSameOrder('', '#0042')).toBe(false);
  });
});

describe('orderNumberFromTimestamp', () => {
  it('returns the timestamp modulo 10000 by default', () => {
    expect(orderNumberFromTimestamp(1717238400)).toBe(8400);
  });

  it('accepts a custom modulus', () => {
    expect(orderNumberFromTimestamp(1000, 100)).toBe(0);
  });

  it('returns 0 for non-finite input', () => {
    expect(orderNumberFromTimestamp(NaN)).toBe(0);
    expect(orderNumberFromTimestamp(-1)).toBe(0);
  });
});

describe('isValidOrderNumberString', () => {
  it('accepts a plain number string', () => {
    expect(isValidOrderNumberString('42')).toBe(true);
  });

  it('accepts a hash-prefixed number', () => {
    expect(isValidOrderNumberString('#0042')).toBe(true);
  });

  it('rejects non-numeric strings', () => {
    expect(isValidOrderNumberString('abc')).toBe(false);
  });

  it('rejects an empty string', () => {
    expect(isValidOrderNumberString('')).toBe(false);
  });

  it('rejects numbers longer than 6 digits', () => {
    expect(isValidOrderNumberString('1234567')).toBe(false);
  });
});
