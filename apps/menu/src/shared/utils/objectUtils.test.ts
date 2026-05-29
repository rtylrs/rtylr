import { describe, it, expect } from 'vitest';
import { isEmptyObject, omit, pick, compactObject, deepClone, getByPath } from './objectUtils';

describe('isEmptyObject', () => {
  it('returns true for an empty object', () => {
    expect(isEmptyObject({})).toBe(true);
  });

  it('returns false for a non-empty object', () => {
    expect(isEmptyObject({ a: 1 })).toBe(false);
  });
});

describe('omit', () => {
  it('removes the specified keys', () => {
    const result = omit({ a: 1, b: 2, c: 3 }, ['b', 'c']);
    expect(result).toEqual({ a: 1 });
  });

  it('does not mutate the original object', () => {
    const obj = { a: 1, b: 2 };
    omit(obj, ['b']);
    expect(obj).toEqual({ a: 1, b: 2 });
  });
});

describe('pick', () => {
  it('returns only the specified keys', () => {
    const result = pick({ a: 1, b: 2, c: 3 }, ['a', 'c']);
    expect(result).toEqual({ a: 1, c: 3 });
  });

  it('does not include keys that do not exist', () => {
    const result = pick({ a: 1 }, ['a', 'b' as keyof { a: number }]);
    expect(Object.keys(result)).toEqual(['a']);
  });
});

describe('compactObject', () => {
  it('removes falsy values', () => {
    const result = compactObject({ a: 1, b: 0, c: '', d: null, e: 'hello' });
    expect(result).toEqual({ a: 1, e: 'hello' });
  });

  it('returns an empty object when all values are falsy', () => {
    expect(compactObject({ a: 0, b: '' })).toEqual({});
  });
});

describe('deepClone', () => {
  it('returns a new object with the same values', () => {
    const original = { a: 1, b: { c: 2 } };
    const cloned = deepClone(original);
    expect(cloned).toEqual(original);
  });

  it('does not share references with the original', () => {
    const original = { a: { b: 1 } };
    const cloned = deepClone(original);
    cloned.a.b = 99;
    expect(original.a.b).toBe(1);
  });
});

describe('getByPath', () => {
  it('returns the value at a dot-notation path', () => {
    expect(getByPath({ a: { b: { c: 42 } } } as Record<string, unknown>, 'a.b.c')).toBe(42);
  });

  it('returns undefined for a non-existent path', () => {
    expect(getByPath({ a: 1 }, 'a.b.c')).toBeUndefined();
  });

  it('returns undefined for an empty path segment', () => {
    expect(getByPath({}, 'x')).toBeUndefined();
  });
});
