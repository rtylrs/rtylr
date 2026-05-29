import { describe, it, expect } from 'vitest';
import {
  toQueryString,
  fromQueryString,
  setParam,
  removeParam,
  getStringParam,
  getNumberParam,
  isEmpty,
} from './queryParamUtils';

describe('toQueryString', () => {
  it('serializes a simple param map', () => {
    expect(toQueryString({ page: 1, size: 25 })).toBe('page=1&size=25');
  });

  it('skips null and undefined values', () => {
    expect(toQueryString({ a: 'x', b: null, c: undefined })).toBe('a=x');
  });

  it('encodes special characters', () => {
    const qs = toQueryString({ q: 'hello world' });
    expect(qs).toBe('q=hello%20world');
  });

  it('returns empty string for empty params', () => {
    expect(toQueryString({})).toBe('');
  });
});

describe('fromQueryString', () => {
  it('parses a simple query string', () => {
    expect(fromQueryString('page=1&size=25')).toEqual({ page: '1', size: '25' });
  });

  it('strips leading question mark', () => {
    expect(fromQueryString('?a=1')).toEqual({ a: '1' });
  });

  it('decodes special characters', () => {
    expect(fromQueryString('q=hello%20world')).toEqual({ q: 'hello world' });
  });

  it('returns empty object for empty string', () => {
    expect(fromQueryString('')).toEqual({});
  });
});

describe('setParam', () => {
  it('adds a new param', () => {
    expect(setParam({ a: 1 }, 'b', 2)).toEqual({ a: 1, b: 2 });
  });

  it('overwrites an existing param', () => {
    expect(setParam({ page: 1 }, 'page', 2)).toEqual({ page: 2 });
  });
});

describe('removeParam', () => {
  it('removes a key from the map', () => {
    const result = removeParam({ a: 1, b: 2 }, 'a');
    expect(result).toEqual({ b: 2 });
  });

  it('does not mutate the original', () => {
    const original = { a: 1 };
    removeParam(original, 'a');
    expect(original.a).toBe(1);
  });
});

describe('getStringParam', () => {
  it('returns the string value', () => {
    expect(getStringParam({ q: 'hello' }, 'q')).toBe('hello');
  });

  it('converts numbers to strings', () => {
    expect(getStringParam({ page: 1 }, 'page')).toBe('1');
  });

  it('returns null for missing keys', () => {
    expect(getStringParam({}, 'missing')).toBeNull();
  });
});

describe('getNumberParam', () => {
  it('returns a number', () => {
    expect(getNumberParam({ page: 3 }, 'page')).toBe(3);
  });

  it('returns null for missing keys', () => {
    expect(getNumberParam({}, 'page')).toBeNull();
  });

  it('returns null for non-numeric values', () => {
    expect(getNumberParam({ q: 'abc' }, 'q')).toBeNull();
  });
});

describe('isEmpty', () => {
  it('returns true when all values are null', () => {
    expect(isEmpty({ a: null, b: null })).toBe(true);
  });

  it('returns false when any value is set', () => {
    expect(isEmpty({ a: 'x', b: null })).toBe(false);
  });
});
