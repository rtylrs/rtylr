import { describe, it, expect } from 'vitest';
import { hexToRgba, isValidHex, lighten, darken, stringToColor, contrastColor } from './colorUtils';

describe('hexToRgba', () => {
  it('converts 6-digit hex to rgba', () => {
    expect(hexToRgba('#ffffff', 1)).toBe('rgba(255, 255, 255, 1)');
  });

  it('converts 6-digit hex without # to rgba', () => {
    expect(hexToRgba('000000', 0.5)).toBe('rgba(0, 0, 0, 0.5)');
  });

  it('handles alpha value', () => {
    const result = hexToRgba('#ff0000', 0.8);
    expect(result).toBe('rgba(255, 0, 0, 0.8)');
  });

  it('defaults to alpha 1', () => {
    expect(hexToRgba('#000000')).toBe('rgba(0, 0, 0, 1)');
  });
});

describe('isValidHex', () => {
  it('returns true for valid 6-digit hex', () => {
    expect(isValidHex('#03AED2')).toBe(true);
  });

  it('returns true for valid 3-digit hex', () => {
    expect(isValidHex('#fff')).toBe(true);
  });

  it('returns false for hex without #', () => {
    expect(isValidHex('ffffff')).toBe(false);
  });

  it('returns false for invalid hex', () => {
    expect(isValidHex('#xyz123')).toBe(false);
  });

  it('returns false for empty string', () => {
    expect(isValidHex('')).toBe(false);
  });
});

describe('lighten', () => {
  it('returns a lighter color', () => {
    const result = lighten('#000000', 50);
    expect(result).not.toBe('#000000');
    expect(result).toMatch(/^#[0-9a-f]{6}$/i);
  });

  it('returns near-white for 100%', () => {
    expect(lighten('#000000', 100)).toBe('#ffffff');
  });

  it('returns the same color for 0%', () => {
    expect(lighten('#ff0000', 0)).toBe('#ff0000');
  });
});

describe('darken', () => {
  it('returns a darker color', () => {
    const result = darken('#ffffff', 50);
    expect(result).not.toBe('#ffffff');
    expect(result).toMatch(/^#[0-9a-f]{6}$/i);
  });

  it('returns black for 100%', () => {
    expect(darken('#ffffff', 100)).toBe('#000000');
  });

  it('returns the same color for 0%', () => {
    expect(darken('#ff0000', 0)).toBe('#ff0000');
  });
});

describe('stringToColor', () => {
  it('returns an HSL color string', () => {
    expect(stringToColor('Alice')).toMatch(/^hsl\(/);
  });

  it('returns the same color for the same input', () => {
    expect(stringToColor('Alice')).toBe(stringToColor('Alice'));
  });

  it('returns different colors for different inputs', () => {
    expect(stringToColor('Alice')).not.toBe(stringToColor('Bob'));
  });
});

describe('contrastColor', () => {
  it('returns black for a light background', () => {
    expect(contrastColor('#ffffff')).toBe('#000000');
  });

  it('returns white for a dark background', () => {
    expect(contrastColor('#000000')).toBe('#ffffff');
  });
});
