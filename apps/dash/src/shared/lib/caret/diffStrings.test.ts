import { describe, expect, it } from 'vitest';
import { commonPrefixLength } from './diffStrings';

describe('commonPrefixLength', () => {
  it('returns the length of the common prefix', () => {
    expect(commonPrefixLength('hello world', 'hello there')).toBe(6);
  });
  it('returns 0 for no common prefix', () => {
    expect(commonPrefixLength('abc', 'xyz')).toBe(0);
  });
  it('returns length of the shorter when one is a prefix', () => {
    expect(commonPrefixLength('foo', 'foobar')).toBe(3);
  });
});
