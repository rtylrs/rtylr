import { describe, expect, it } from 'vitest';
import { safeGet } from './safeAccess';

describe('safeGet', () => {
  it('walks a nested path', () => {
    expect(safeGet({ a: { b: { c: 1 } } }, 'a.b.c')).toBe(1);
  });
  it('returns undefined for missing paths', () => {
    expect(safeGet({ a: 1 }, 'b.c')).toBeUndefined();
  });
  it('returns the object for empty path', () => {
    expect(safeGet({ a: 1 }, '')).toEqual({ a: 1 });
  });
});
