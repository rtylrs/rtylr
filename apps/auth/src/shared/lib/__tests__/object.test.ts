import { describe, it, expect } from 'vitest';
import { omit, pick, isEmpty, merge, entries, fromEntries } from '../object';

describe('omit', () => {
  it('removes specified keys', () => {
    const result = omit({ a: 1, b: 2, c: 3 }, ['b']);
    expect(result).toEqual({ a: 1, c: 3 });
  });

  it('returns a copy when no keys match', () => {
    const obj = { a: 1 };
    const result = omit(obj, [] as never[]);
    expect(result).toEqual({ a: 1 });
  });
});

describe('pick', () => {
  it('returns only specified keys', () => {
    const result = pick({ a: 1, b: 2, c: 3 }, ['a', 'c']);
    expect(result).toEqual({ a: 1, c: 3 });
  });
});

describe('isEmpty', () => {
  it('returns true for an empty object', () => {
    expect(isEmpty({})).toBe(true);
  });

  it('returns false for an object with properties', () => {
    expect(isEmpty({ a: 1 })).toBe(false);
  });
});

describe('merge', () => {
  it('merges two objects with b overriding a', () => {
    expect(merge({ a: 1, b: 2 }, { b: 99 })).toEqual({ a: 1, b: 99 });
  });

  it('does not mutate the original', () => {
    const a = { x: 1 };
    merge(a, { x: 2 });
    expect(a.x).toBe(1);
  });
});

describe('entries', () => {
  it('returns key-value pairs', () => {
    expect(entries({ a: 1, b: 2 })).toEqual(
      expect.arrayContaining([
        ['a', 1],
        ['b', 2],
      ])
    );
  });
});

describe('fromEntries', () => {
  it('builds an object from pairs', () => {
    expect(
      fromEntries([
        ['a', 1],
        ['b', 2],
      ])
    ).toEqual({ a: 1, b: 2 });
  });
});
