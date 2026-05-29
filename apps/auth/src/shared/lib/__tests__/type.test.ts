import { describe, it, expect } from 'vitest';
import {
  isObject,
  isString,
  isNumber,
  isBoolean,
  isNullish,
  isNonEmptyArray,
  isFunction,
  toObject,
  getString,
} from '../type';

describe('isObject', () => {
  it('returns true for plain objects', () => {
    expect(isObject({ a: 1 })).toBe(true);
  });

  it('returns false for arrays', () => {
    expect(isObject([1, 2, 3])).toBe(false);
  });

  it('returns false for null', () => {
    expect(isObject(null)).toBe(false);
  });

  it('returns false for primitives', () => {
    expect(isObject('string')).toBe(false);
  });
});

describe('isString', () => {
  it('returns true for strings', () => {
    expect(isString('hello')).toBe(true);
  });

  it('returns false for non-strings', () => {
    expect(isString(42)).toBe(false);
  });
});

describe('isNumber', () => {
  it('returns true for finite numbers', () => {
    expect(isNumber(3.14)).toBe(true);
  });

  it('returns false for NaN', () => {
    expect(isNumber(NaN)).toBe(false);
  });

  it('returns false for Infinity', () => {
    expect(isNumber(Infinity)).toBe(false);
  });
});

describe('isBoolean', () => {
  it('returns true for booleans', () => {
    expect(isBoolean(true)).toBe(true);
    expect(isBoolean(false)).toBe(true);
  });

  it('returns false for non-booleans', () => {
    expect(isBoolean(1)).toBe(false);
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
});

describe('isNonEmptyArray', () => {
  it('returns true for non-empty array', () => {
    expect(isNonEmptyArray([1])).toBe(true);
  });

  it('returns false for empty array', () => {
    expect(isNonEmptyArray([])).toBe(false);
  });

  it('returns false for non-arrays', () => {
    expect(isNonEmptyArray('hello')).toBe(false);
  });
});

describe('isFunction', () => {
  it('returns true for functions', () => {
    expect(isFunction(() => {})).toBe(true);
  });

  it('returns false for non-functions', () => {
    expect(isFunction(42)).toBe(false);
  });
});

describe('toObject', () => {
  it('returns the object unchanged', () => {
    const obj = { a: 1 };
    expect(toObject(obj)).toBe(obj);
  });

  it('returns empty object for non-objects', () => {
    expect(toObject('string')).toEqual({});
    expect(toObject(null)).toEqual({});
  });
});

describe('getString', () => {
  it('returns the string unchanged', () => {
    expect(getString('hello')).toBe('hello');
  });

  it('returns fallback for non-strings', () => {
    expect(getString(42, 'default')).toBe('default');
  });

  it('returns empty string by default', () => {
    expect(getString(null)).toBe('');
  });
});
