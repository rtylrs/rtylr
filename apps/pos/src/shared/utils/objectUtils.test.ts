import { describe, it, expect } from 'vitest';
import {
  omit,
  pick,
  isEmpty,
  merge,
  shallowEqual,
  getOrDefault,
  compact,
} from './objectUtils';

describe('omit', () => {
  it('returns an object without the omitted keys', () => {
    const obj = { a: 1, b: 2, c: 3 };
    expect(omit(obj, ['b'])).toEqual({ a: 1, c: 3 });
  });

  it('returns the original object when no keys are omitted', () => {
    const obj = { a: 1, b: 2 };
    expect(omit(obj, [])).toEqual({ a: 1, b: 2 });
  });

  it('does not mutate the original object', () => {
    const obj = { a: 1, b: 2 };
    omit(obj, ['a']);
    expect(obj).toEqual({ a: 1, b: 2 });
  });
});

describe('pick', () => {
  it('returns only the specified keys', () => {
    const obj = { a: 1, b: 2, c: 3 };
    expect(pick(obj, ['a', 'c'])).toEqual({ a: 1, c: 3 });
  });

  it('returns an empty object when no keys are picked', () => {
    expect(pick({ a: 1 }, [])).toEqual({});
  });
});

describe('isEmpty', () => {
  it('returns true for an empty object', () => {
    expect(isEmpty({})).toBe(true);
  });

  it('returns false for a non-empty object', () => {
    expect(isEmpty({ a: 1 })).toBe(false);
  });
});

describe('merge', () => {
  it('merges two objects with overrides taking precedence', () => {
    expect(merge({ a: 1, b: 2 }, { b: 99, c: 3 } as any)).toEqual({ a: 1, b: 99, c: 3 });
  });

  it('does not mutate the base object', () => {
    const base = { a: 1, b: 2 };
    merge(base, { b: 99 });
    expect(base.b).toBe(2);
  });
});

describe('shallowEqual', () => {
  it('returns true for two equal objects', () => {
    expect(shallowEqual({ a: 1, b: 2 }, { a: 1, b: 2 })).toBe(true);
  });

  it('returns false when values differ', () => {
    expect(shallowEqual({ a: 1, b: 2 }, { a: 1, b: 3 })).toBe(false);
  });

  it('returns false when key counts differ', () => {
    expect(shallowEqual({ a: 1 } as any, { a: 1, b: 2 } as any)).toBe(false);
  });
});

describe('getOrDefault', () => {
  it('returns the value when defined', () => {
    expect(getOrDefault({ a: 42 }, 'a', 0)).toBe(42);
  });

  it('returns the fallback when value is undefined', () => {
    expect(getOrDefault({ a: undefined } as any, 'a', 99)).toBe(99);
  });
});

describe('compact', () => {
  it('removes falsy values from an object', () => {
    const obj = { a: 1, b: 0, c: '', d: null, e: 'hello' };
    expect(compact(obj as any)).toEqual({ a: 1, e: 'hello' });
  });

  it('returns an empty object for an all-falsy object', () => {
    expect(compact({ a: 0, b: '' } as any)).toEqual({});
  });
});
