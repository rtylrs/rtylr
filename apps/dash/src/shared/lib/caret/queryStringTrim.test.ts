import { describe, expect, it } from 'vitest';
import { trimQueryParams } from './queryStringTrim';

describe('trimQueryParams', () => {
  it('drops null/undefined/empty values', () => {
    expect(trimQueryParams({ a: 'x', b: null, c: undefined, d: '' })).toEqual({ a: 'x' });
  });
  it('stringifies numbers', () => {
    expect(trimQueryParams({ n: 5 })).toEqual({ n: '5' });
  });
});
