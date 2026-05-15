import { describe, expect, it } from 'vitest';
import { clamp } from './clamp';

describe('clamp', () => {
  it('returns the value when in range', () => {
    expect(clamp(5, 0, 10)).toBe(5);
  });
  it('clamps below the minimum', () => {
    expect(clamp(-3, 0, 10)).toBe(0);
  });
  it('clamps above the maximum', () => {
    expect(clamp(99, 0, 10)).toBe(10);
  });
  it('returns the minimum for NaN', () => {
    expect(clamp(NaN, 0, 10)).toBe(0);
  });
});
