import { describe, expect, it } from 'vitest';

import {
  contrastRatio,
  darken,
  hexToRgb,
  isAccessibleAA,
  isAccessibleAAA,
  lighten,
  readableTextColor,
  relativeLuminance,
  rgbToHex,
  withAlphaOnWhite,
} from './colorUtils';

describe('hexToRgb', () => {
  it('parses a 6-digit hex color', () => {
    expect(hexToRgb('#ff0000')).toEqual({ r: 255, g: 0, b: 0 });
  });

  it('parses a 3-digit hex color', () => {
    expect(hexToRgb('#f00')).toEqual({ r: 255, g: 0, b: 0 });
  });

  it('parses without hash prefix', () => {
    expect(hexToRgb('00ff00')).toEqual({ r: 0, g: 255, b: 0 });
  });

  it('returns null for invalid input', () => {
    expect(hexToRgb('xyz')).toBeNull();
  });
});

describe('rgbToHex', () => {
  it('converts red to hex', () => {
    expect(rgbToHex(255, 0, 0)).toBe('#ff0000');
  });

  it('converts white to hex', () => {
    expect(rgbToHex(255, 255, 255)).toBe('#ffffff');
  });

  it('converts black to hex', () => {
    expect(rgbToHex(0, 0, 0)).toBe('#000000');
  });

  it('clamps values over 255', () => {
    expect(rgbToHex(300, 0, 0)).toBe('#ff0000');
  });
});

describe('relativeLuminance', () => {
  it('returns 1 for white', () => {
    expect(relativeLuminance(255, 255, 255)).toBeCloseTo(1, 2);
  });

  it('returns 0 for black', () => {
    expect(relativeLuminance(0, 0, 0)).toBe(0);
  });
});

describe('contrastRatio', () => {
  it('returns 21 for black on white', () => {
    expect(contrastRatio('#000000', '#ffffff')).toBeCloseTo(21, 0);
  });

  it('returns 1 for same color', () => {
    expect(contrastRatio('#ffffff', '#ffffff')).toBeCloseTo(1, 1);
  });

  it('returns 1 for invalid colors', () => {
    expect(contrastRatio('bad', 'also-bad')).toBe(1);
  });
});

describe('isAccessibleAA', () => {
  it('returns true for black on white', () => {
    expect(isAccessibleAA('#000000', '#ffffff')).toBe(true);
  });

  it('returns false for similar colors', () => {
    expect(isAccessibleAA('#cccccc', '#ffffff')).toBe(false);
  });
});

describe('isAccessibleAAA', () => {
  it('returns true for black on white', () => {
    expect(isAccessibleAAA('#000000', '#ffffff')).toBe(true);
  });

  it('returns false for moderate contrast', () => {
    expect(isAccessibleAAA('#767676', '#ffffff')).toBe(false);
  });
});

describe('readableTextColor', () => {
  it('returns black for a light background', () => {
    expect(readableTextColor('#ffffff')).toBe('#000000');
  });

  it('returns white for a dark background', () => {
    expect(readableTextColor('#000000')).toBe('#ffffff');
  });

  it('returns black for an invalid hex', () => {
    expect(readableTextColor('bad')).toBe('#000000');
  });
});

describe('lighten', () => {
  it('lightens a dark color', () => {
    const result = hexToRgb(lighten('#000000', 50));
    expect(result).not.toBeNull();
    expect(result!.r).toBeGreaterThan(0);
  });

  it('returns input for truly invalid hex', () => {
    expect(lighten('zzz!', 10)).toBe('zzz!');
  });
});

describe('darken', () => {
  it('darkens a light color', () => {
    const result = hexToRgb(darken('#ffffff', 50));
    expect(result).not.toBeNull();
    expect(result!.r).toBeLessThan(255);
  });

  it('returns input for truly invalid hex', () => {
    expect(darken('zzz!', 10)).toBe('zzz!');
  });
});

describe('withAlphaOnWhite', () => {
  it('returns white when alpha is 0', () => {
    expect(withAlphaOnWhite('#ff0000', 0)).toBe('#ffffff');
  });

  it('returns original color when alpha is 1', () => {
    expect(withAlphaOnWhite('#ff0000', 1)).toBe('#ff0000');
  });
});
