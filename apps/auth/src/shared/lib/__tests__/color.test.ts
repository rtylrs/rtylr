import { describe, it, expect } from 'vitest';
import { hexToRgb, rgbToHex, relativeLuminance, contrastRatio, textOnBackground } from '../color';

describe('hexToRgb', () => {
  it('parses a 6-digit hex', () => {
    expect(hexToRgb('#ff0000')).toEqual({ r: 255, g: 0, b: 0 });
  });

  it('parses a 3-digit hex', () => {
    expect(hexToRgb('#f00')).toEqual({ r: 255, g: 0, b: 0 });
  });

  it('parses without leading #', () => {
    expect(hexToRgb('00ff00')).toEqual({ r: 0, g: 255, b: 0 });
  });

  it('returns null for invalid hex', () => {
    expect(hexToRgb('xyz')).toBeNull();
  });
});

describe('rgbToHex', () => {
  it('converts red to ff0000', () => {
    expect(rgbToHex(255, 0, 0)).toBe('ff0000');
  });

  it('pads single digit values', () => {
    expect(rgbToHex(0, 0, 0)).toBe('000000');
  });
});

describe('relativeLuminance', () => {
  it('returns 1 for white', () => {
    expect(relativeLuminance('#ffffff')).toBeCloseTo(1, 2);
  });

  it('returns 0 for black', () => {
    expect(relativeLuminance('#000000')).toBeCloseTo(0, 2);
  });
});

describe('contrastRatio', () => {
  it('returns 21 for black on white', () => {
    expect(contrastRatio('#ffffff', '#000000')).toBeCloseTo(21, 0);
  });

  it('returns 1 for same colour', () => {
    expect(contrastRatio('#ff0000', '#ff0000')).toBeCloseTo(1, 2);
  });
});

describe('textOnBackground', () => {
  it('returns dark for a light background', () => {
    expect(textOnBackground('#ffffff')).toBe('dark');
  });

  it('returns light for a dark background', () => {
    expect(textOnBackground('#000000')).toBe('light');
  });
});
