import { describe, it, expect } from 'vitest';
import { assert, assertDefined, assertNonEmpty } from '../assert';

describe('assert', () => {
  it('does not throw when condition is truthy', () => {
    expect(() => assert(true, 'fail')).not.toThrow();
    expect(() => assert(1, 'fail')).not.toThrow();
  });

  it('throws with the given message when condition is falsy', () => {
    expect(() => assert(false, 'condition failed')).toThrow('condition failed');
    expect(() => assert(0, 'zero is falsy')).toThrow('zero is falsy');
    expect(() => assert(null, 'null is falsy')).toThrow('null is falsy');
  });
});

describe('assertDefined', () => {
  it('returns the value when it is defined', () => {
    expect(assertDefined('hello', 'msg')).toBe('hello');
    expect(assertDefined(0, 'msg')).toBe(0);
  });

  it('throws for null', () => {
    expect(() => assertDefined(null, 'was null')).toThrow('was null');
  });

  it('throws for undefined', () => {
    expect(() => assertDefined(undefined, 'was undefined')).toThrow('was undefined');
  });
});

describe('assertNonEmpty', () => {
  it('does not throw for a non-empty string', () => {
    expect(() => assertNonEmpty('hello', 'fail')).not.toThrow();
  });

  it('throws for an empty string', () => {
    expect(() => assertNonEmpty('', 'empty string')).toThrow('empty string');
  });

  it('throws for a whitespace-only string', () => {
    expect(() => assertNonEmpty('   ', 'blank string')).toThrow('blank string');
  });
});
