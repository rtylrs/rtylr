import { describe, it, expect } from 'vitest';
import { cn } from './cn';

describe('cn', () => {
  it('returns empty string for no args', () => {
    expect(cn()).toBe('');
  });

  it('joins plain string values', () => {
    expect(cn('foo', 'bar')).toBe('foo bar');
  });

  it('ignores falsy values', () => {
    expect(cn('a', '', false, null, undefined, 'b')).toBe('a b');
  });

  it('flattens nested arrays', () => {
    expect(cn(['foo', 'bar'], 'baz')).toBe('foo bar baz');
  });

  it('includes object keys whose values are truthy', () => {
    expect(cn({ active: true, disabled: false, visible: true })).toBe('active visible');
  });

  it('handles mixed inputs: string, array, object', () => {
    expect(cn('base', ['extra', false && 'nope'], { highlight: true })).toBe(
      'base extra highlight'
    );
  });

  it('coerces numbers to strings', () => {
    expect(cn(0, 1)).toBe('1');
  });

  it('deduplication is NOT performed (caller responsibility)', () => {
    expect(cn('a', 'a')).toBe('a a');
  });
});
