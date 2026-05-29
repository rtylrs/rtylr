import { describe, it, expect } from 'vitest';
import {
  chartColor,
  assignColors,
  hexToRgb,
  relativeLuminance,
  contrastingTextColor,
} from './color';

describe('chartColor', () => {
  it('returns a color string for index 0', () => {
    expect(chartColor(0)).toMatch(/^#/);
  });

  it('wraps around past palette length', () => {
    expect(chartColor(0)).toBe(chartColor(8));
  });
});

describe('assignColors', () => {
  it('assigns a color to each key', () => {
    const result = assignColors(['a', 'b', 'c']);
    expect(Object.keys(result)).toHaveLength(3);
    expect(result['a']).toMatch(/^#/);
  });

  it('returns empty object for empty keys', () => {
    expect(assignColors([])).toEqual({});
  });
});

describe('hexToRgb', () => {
  it('parses a hex color with hash prefix', () => {
    expect(hexToRgb('#ffffff')).toEqual({ r: 255, g: 255, b: 255 });
  });

  it('parses a hex color without hash prefix', () => {
    expect(hexToRgb('000000')).toEqual({ r: 0, g: 0, b: 0 });
  });

  it('returns null for invalid hex', () => {
    expect(hexToRgb('#zzz')).toBeNull();
    expect(hexToRgb('abc')).toBeNull();
  });
});

describe('relativeLuminance', () => {
  it('returns 1 for white', () => {
    expect(relativeLuminance('#ffffff')).toBeCloseTo(1, 2);
  });

  it('returns 0 for black', () => {
    expect(relativeLuminance('#000000')).toBeCloseTo(0, 5);
  });

  it('returns -1 for invalid hex', () => {
    expect(relativeLuminance('nothex')).toBe(-1);
  });
});

describe('contrastingTextColor', () => {
  it('returns white for dark backgrounds', () => {
    expect(contrastingTextColor('#000000')).toBe('#ffffff');
  });

  it('returns black for light backgrounds', () => {
    expect(contrastingTextColor('#ffffff')).toBe('#000000');
  });
});
