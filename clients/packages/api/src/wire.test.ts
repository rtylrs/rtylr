import { describe, expect, it } from 'vitest';
import { safeRatio, toFiniteNumber, toOptionalFiniteNumber } from './wire';

describe('API wire number helpers', () => {
  it('coerces numeric strings and rejects non-finite values', () => {
    expect(toFiniteNumber('42.5')).toBe(42.5);
    expect(toFiniteNumber('invalid')).toBe(0);
    expect(toFiniteNumber(Infinity, 7)).toBe(7);
  });

  it('preserves missing optional values', () => {
    expect(toOptionalFiniteNumber(undefined)).toBeUndefined();
    expect(toOptionalFiniteNumber('12')).toBe(12);
    expect(toOptionalFiniteNumber('invalid')).toBeUndefined();
  });

  it('prevents invalid and zero-denominator ratios', () => {
    expect(safeRatio(5, 0)).toBe(0);
    expect(safeRatio('10', '4')).toBe(2.5);
    expect(safeRatio('invalid', 2)).toBe(0);
  });
});
