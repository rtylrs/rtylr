import { describe, it, expect } from 'vitest';
import {
  isObject,
  isString,
  isNumber,
  isBoolean,
  isArray,
  isNullish,
  isFunction,
  isNonEmptyString,
} from './typeGuards';

describe('isObject', () => {
  it('returns true for a plain object', () => {
    expect(isObject({ a: 1 })).toBe(true);
  });

  it('returns false for null', () => {
    expect(isObject(null)).toBe(false);
  });

  it('returns false for an array', () => {
    expect(isObject([1, 2])).toBe(false);
  });

  it('returns false for a string', () => {
    expect(isObject('hello')).toBe(false);
  });
});

describe('isString', () => {
  it('returns true for a string', () => {
    expect(isString('hello')).toBe(true);
  });

  it('returns false for a number', () => {
    expect(isString(42)).toBe(false);
  });
});

describe('isNumber', () => {
  it('returns true for a number', () => {
    expect(isNumber(42)).toBe(true);
  });

  it('returns false for NaN', () => {
    expect(isNumber(NaN)).toBe(false);
  });

  it('returns false for a string', () => {
    expect(isNumber('42')).toBe(false);
  });
});

describe('isBoolean', () => {
  it('returns true for true', () => {
    expect(isBoolean(true)).toBe(true);
  });

  it('returns true for false', () => {
    expect(isBoolean(false)).toBe(true);
  });

  it('returns false for 1', () => {
    expect(isBoolean(1)).toBe(false);
  });
});

describe('isArray', () => {
  it('returns true for an array', () => {
    expect(isArray([1, 2, 3])).toBe(true);
  });

  it('returns false for an object', () => {
    expect(isArray({ length: 3 })).toBe(false);
  });
});

describe('isNullish', () => {
  it('returns true for null', () => {
    expect(isNullish(null)).toBe(true);
  });

  it('returns true for undefined', () => {
    expect(isNullish(undefined)).toBe(true);
  });

  it('returns false for 0', () => {
    expect(isNullish(0)).toBe(false);
  });

  it('returns false for an empty string', () => {
    expect(isNullish('')).toBe(false);
  });
});

describe('isFunction', () => {
  it('returns true for a function', () => {
    expect(isFunction(() => {})).toBe(true);
  });

  it('returns false for an object', () => {
    expect(isFunction({})).toBe(false);
  });
});

describe('isNonEmptyString', () => {
  it('returns true for a non-empty string', () => {
    expect(isNonEmptyString('hello')).toBe(true);
  });

  it('returns false for an empty string', () => {
    expect(isNonEmptyString('')).toBe(false);
  });

  it('returns false for a whitespace-only string', () => {
    expect(isNonEmptyString('   ')).toBe(false);
  });

  it('returns false for a number', () => {
    expect(isNonEmptyString(42)).toBe(false);
  });
});
