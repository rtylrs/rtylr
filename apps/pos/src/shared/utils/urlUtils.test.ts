import { describe, it, expect } from 'vitest';
import {
  buildUrl,
  parseQuery,
  getQueryParam,
  serializeQuery,
  isAbsoluteUrl,
} from './urlUtils';

describe('buildUrl', () => {
  it('appends query parameters to a path', () => {
    const url = buildUrl('/orders', { page: 1, status: 'open' });
    expect(url).toBe('/orders?page=1&status=open');
  });

  it('returns just the path when params are empty', () => {
    expect(buildUrl('/orders', {})).toBe('/orders');
  });
});

describe('parseQuery', () => {
  it('parses a query string into a record', () => {
    expect(parseQuery('page=1&status=open')).toEqual({ page: '1', status: 'open' });
  });

  it('handles a leading "?" in the query string', () => {
    expect(parseQuery('?page=2')).toEqual({ page: '2' });
  });

  it('returns an empty object for an empty string', () => {
    expect(parseQuery('')).toEqual({});
  });
});

describe('getQueryParam', () => {
  it('returns the value for an existing key', () => {
    expect(getQueryParam('page=1&status=open', 'status')).toBe('open');
  });

  it('returns null for a missing key', () => {
    expect(getQueryParam('page=1', 'status')).toBeNull();
  });
});

describe('serializeQuery', () => {
  it('serializes a record into a query string', () => {
    const qs = serializeQuery({ a: 1, b: 'hello' });
    expect(qs).toContain('a=1');
    expect(qs).toContain('b=hello');
  });

  it('returns an empty string for empty params', () => {
    expect(serializeQuery({})).toBe('');
  });
});

describe('isAbsoluteUrl', () => {
  it('returns true for a valid http URL', () => {
    expect(isAbsoluteUrl('http://example.com')).toBe(true);
  });

  it('returns true for a valid https URL', () => {
    expect(isAbsoluteUrl('https://example.com/api')).toBe(true);
  });

  it('returns false for a relative path', () => {
    expect(isAbsoluteUrl('/orders')).toBe(false);
  });

  it('returns false for an empty string', () => {
    expect(isAbsoluteUrl('')).toBe(false);
  });
});
