import { describe, it, expect } from 'vitest';
import { roundTo, clamp, formatCompact, safeDiv, formatOrdinal, lerp } from './number';

describe('roundTo', () => {
  it('rounds to the specified decimal places', () => {
    expect(roundTo(1.2345, 2)).toBeCloseTo(1.23);
  });

  it('rounds integers correctly', () => {
    expect(roundTo(5, 0)).toBe(5);
  });

  it('handles zero', () => {
    expect(roundTo(0, 3)).toBe(0);
  });

  it('rounds a value with more decimals than requested', () => {
    expect(roundTo(1.567, 1)).toBeCloseTo(1.6);
  });
});

describe('clamp', () => {
  it('returns the value when within range', () => {
    expect(clamp(5, 0, 10)).toBe(5);
  });

  it('returns min when below range', () => {
    expect(clamp(-5, 0, 10)).toBe(0);
  });

  it('returns max when above range', () => {
    expect(clamp(15, 0, 10)).toBe(10);
  });

  it('returns min when value equals min', () => {
    expect(clamp(0, 0, 10)).toBe(0);
  });

  it('returns max when value equals max', () => {
    expect(clamp(10, 0, 10)).toBe(10);
  });
});

describe('formatCompact', () => {
  it('formats values under 1000 as plain numbers', () => {
    expect(formatCompact(999)).toBe('999');
  });

  it('formats thousands with K suffix', () => {
    expect(formatCompact(1500)).toBe('1.5K');
  });

  it('formats millions with M suffix', () => {
    expect(formatCompact(2_500_000)).toBe('2.5M');
  });

  it('formats billions with B suffix', () => {
    expect(formatCompact(1_000_000_000)).toBe('1B');
  });

  it('handles negative numbers', () => {
    expect(formatCompact(-1500)).toBe('-1.5K');
  });

  it('formats zero as 0', () => {
    expect(formatCompact(0)).toBe('0');
  });
});

describe('safeDiv', () => {
  it('divides when denominator is nonzero', () => {
    expect(safeDiv(10, 4)).toBeCloseTo(2.5);
  });

  it('returns default fallback (0) when denominator is 0', () => {
    expect(safeDiv(10, 0)).toBe(0);
  });

  it('returns custom fallback when denominator is 0', () => {
    expect(safeDiv(10, 0, -1)).toBe(-1);
  });
});

describe('formatOrdinal', () => {
  it('returns 1st for 1', () => {
    expect(formatOrdinal(1)).toBe('1st');
  });

  it('returns 2nd for 2', () => {
    expect(formatOrdinal(2)).toBe('2nd');
  });

  it('returns 3rd for 3', () => {
    expect(formatOrdinal(3)).toBe('3rd');
  });

  it('returns 4th for 4', () => {
    expect(formatOrdinal(4)).toBe('4th');
  });

  it('returns 11th for 11 (special case)', () => {
    expect(formatOrdinal(11)).toBe('11th');
  });

  it('returns 21st for 21', () => {
    expect(formatOrdinal(21)).toBe('21st');
  });

  it('returns 12th for 12 (special case)', () => {
    expect(formatOrdinal(12)).toBe('12th');
  });
});

describe('lerp', () => {
  it('returns a when t is 0', () => {
    expect(lerp(0, 100, 0)).toBe(0);
  });

  it('returns b when t is 1', () => {
    expect(lerp(0, 100, 1)).toBe(100);
  });

  it('returns midpoint when t is 0.5', () => {
    expect(lerp(0, 100, 0.5)).toBe(50);
  });

  it('works with negative values', () => {
    expect(lerp(-100, 100, 0.5)).toBe(0);
  });
});
