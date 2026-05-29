import { describe, it, expect } from 'vitest';
import {
  assertDefined,
  isPlainObject,
  isNonEmptyString,
  isFiniteNumber,
  isPositiveInteger,
  asString,
  asNumber,
} from './assertUtils';

describe('assertDefined', () => {
  it('returns the value when defined', () => {
    expect(assertDefined('hello')).toBe('hello');
    expect(assertDefined(0)).toBe(0);
  });

  it('throws for null', () => {
    expect(() => assertDefined(null)).toThrow();
  });

  it('throws for undefined', () => {
    expect(() => assertDefined(undefined)).toThrow();
  });

  it('includes the label in the error message', () => {
    expect(() => assertDefined(null, 'orderId')).toThrow('orderId');
  });
});

describe('isPlainObject', () => {
  it('returns true for a plain object', () => {
    expect(isPlainObject({ a: 1 })).toBe(true);
  });

  it('returns false for an array', () => {
    expect(isPlainObject([1, 2])).toBe(false);
  });

  it('returns false for null', () => {
    expect(isPlainObject(null)).toBe(false);
  });

  it('returns false for primitives', () => {
    expect(isPlainObject('string')).toBe(false);
    expect(isPlainObject(42)).toBe(false);
  });
});

describe('isNonEmptyString', () => {
  it('returns true for a non-empty string', () => {
    expect(isNonEmptyString('hello')).toBe(true);
  });

  it('returns false for an empty string', () => {
    expect(isNonEmptyString('')).toBe(false);
    expect(isNonEmptyString('   ')).toBe(false);
  });

  it('returns false for non-strings', () => {
    expect(isNonEmptyString(42)).toBe(false);
    expect(isNonEmptyString(null)).toBe(false);
  });
});

describe('isFiniteNumber', () => {
  it('returns true for a finite number', () => {
    expect(isFiniteNumber(42)).toBe(true);
    expect(isFiniteNumber(0)).toBe(true);
  });

  it('returns false for NaN', () => {
    expect(isFiniteNumber(NaN)).toBe(false);
  });

  it('returns false for Infinity', () => {
    expect(isFiniteNumber(Infinity)).toBe(false);
  });

  it('returns false for non-numbers', () => {
    expect(isFiniteNumber('42')).toBe(false);
  });
});

describe('isPositiveInteger', () => {
  it('returns true for a positive integer', () => {
    expect(isPositiveInteger(5)).toBe(true);
  });

  it('returns false for 0', () => {
    expect(isPositiveInteger(0)).toBe(false);
  });

  it('returns false for a float', () => {
    expect(isPositiveInteger(1.5)).toBe(false);
  });

  it('returns false for negative integers', () => {
    expect(isPositiveInteger(-1)).toBe(false);
  });
});

describe('asString', () => {
  it('returns the string when the value is a string', () => {
    expect(asString('hello')).toBe('hello');
  });

  it('returns null for non-strings', () => {
    expect(asString(42)).toBeNull();
    expect(asString(null)).toBeNull();
  });
});

describe('asNumber', () => {
  it('returns the number when the value is a finite number', () => {
    expect(asNumber(42)).toBe(42);
  });

  it('parses a numeric string', () => {
    expect(asNumber('3.14')).toBe(3.14);
  });

  it('returns null for non-numeric strings', () => {
    expect(asNumber('abc')).toBeNull();
  });

  it('returns null for NaN', () => {
    expect(asNumber(NaN)).toBeNull();
  });

  it('returns null for null', () => {
    expect(asNumber(null)).toBeNull();
  });
});
