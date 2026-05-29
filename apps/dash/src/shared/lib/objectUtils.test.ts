import { describe, it, expect } from 'vitest';
import {
  pick,
  omit,
  merge,
  isEmpty,
  compactObject,
  mapValues,
  keys,
  deepEqual,
} from './objectUtils';

describe('pick', () => {
  it('returns only specified keys', () => {
    const result = pick({ a: 1, b: 2, c: 3 }, ['a', 'c']);
    expect(result).toEqual({ a: 1, c: 3 });
  });

  it('ignores missing keys', () => {
    const result = pick({ a: 1 }, ['a'] as ('a' | 'b')[]);
    expect(result).toEqual({ a: 1 });
  });
});

describe('omit', () => {
  it('removes specified keys', () => {
    const result = omit({ a: 1, b: 2, c: 3 }, ['b']);
    expect(result).toEqual({ a: 1, c: 3 });
  });

  it('does not mutate the original', () => {
    const obj = { a: 1, b: 2 };
    omit(obj, ['b']);
    expect(obj).toHaveProperty('b');
  });
});

describe('merge', () => {
  it('merges two objects (later wins)', () => {
    const result = merge({ a: 1, b: 2 }, { b: 3, c: 4 });
    expect(result).toEqual({ a: 1, b: 3, c: 4 });
  });

  it('returns empty object for no inputs', () => {
    expect(merge()).toEqual({});
  });
});

describe('isEmpty', () => {
  it('returns true for empty object', () => {
    expect(isEmpty({})).toBe(true);
  });

  it('returns false for object with keys', () => {
    expect(isEmpty({ a: 1 })).toBe(false);
  });
});

describe('compactObject', () => {
  it('removes null and undefined values', () => {
    const result = compactObject({ a: 1, b: null, c: undefined, d: 'hello' });
    expect(result).toEqual({ a: 1, d: 'hello' });
  });

  it('preserves falsy non-null values', () => {
    const result = compactObject({ a: 0, b: '', c: false });
    expect(result).toEqual({ a: 0, b: '', c: false });
  });
});

describe('mapValues', () => {
  it('transforms values', () => {
    const result = mapValues({ a: 1, b: 2 }, (v) => v * 2);
    expect(result).toEqual({ a: 2, b: 4 });
  });

  it('provides the key to the transform', () => {
    const result = mapValues({ foo: 1 }, (v, k) => `${String(k)}_${v}`);
    expect(result).toEqual({ foo: 'foo_1' });
  });
});

describe('keys', () => {
  it('returns keys of object', () => {
    expect(keys({ a: 1, b: 2 })).toEqual(['a', 'b']);
  });

  it('returns empty array for empty object', () => {
    expect(keys({})).toEqual([]);
  });
});

describe('deepEqual', () => {
  it('returns true for identical primitives', () => {
    expect(deepEqual(1, 1)).toBe(true);
  });

  it('returns true for deep equal objects', () => {
    expect(deepEqual({ a: { b: 1 } }, { a: { b: 1 } })).toBe(true);
  });

  it('returns false for differing objects', () => {
    expect(deepEqual({ a: 1 }, { a: 2 })).toBe(false);
  });

  it('returns false for different key counts', () => {
    expect(deepEqual({ a: 1 }, { a: 1, b: 2 })).toBe(false);
  });
});
