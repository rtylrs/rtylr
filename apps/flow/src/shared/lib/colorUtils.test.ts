import { describe, it, expect } from 'vitest';

import {
  isValidHex,
  normalizeHex,
  hexLuminance,
  textOnColor,
  blendHex,
  toPastel,
  paletteColor,
  PROJECT_COLORS,
} from './colorUtils';

describe('isValidHex', () => {
  it('accepts valid 6-digit hex with #', () => {
    expect(isValidHex('#3b82f6')).toBe(true);
  });

  it('accepts valid 6-digit hex without #', () => {
    expect(isValidHex('3b82f6')).toBe(true);
  });

  it('rejects short hex', () => {
    expect(isValidHex('#fff')).toBe(false);
  });

  it('rejects non-hex characters', () => {
    expect(isValidHex('#gggggg')).toBe(false);
  });
});

describe('normalizeHex', () => {
  it('adds # when missing', () => {
    expect(normalizeHex('3b82f6')).toBe('#3b82f6');
  });

  it('lowercases the hex', () => {
    expect(normalizeHex('#3B82F6')).toBe('#3b82f6');
  });
});

describe('hexLuminance', () => {
  it('returns 1 for white', () => {
    expect(hexLuminance('#ffffff')).toBeCloseTo(1, 1);
  });

  it('returns 0 for black', () => {
    expect(hexLuminance('#000000')).toBeCloseTo(0, 5);
  });

  it('returns value between 0 and 1 for mid-tone', () => {
    const lum = hexLuminance('#3b82f6');
    expect(lum).toBeGreaterThan(0);
    expect(lum).toBeLessThan(1);
  });
});

describe('textOnColor', () => {
  it('returns light for dark background', () => {
    expect(textOnColor('#000000')).toBe('light');
  });

  it('returns dark for light background', () => {
    expect(textOnColor('#ffffff')).toBe('dark');
  });
});

describe('blendHex', () => {
  it('returns first color at ratio 0', () => {
    expect(blendHex('#000000', '#ffffff', 0)).toBe('#000000');
  });

  it('returns second color at ratio 1', () => {
    expect(blendHex('#000000', '#ffffff', 1)).toBe('#ffffff');
  });

  it('returns gray at ratio 0.5', () => {
    expect(blendHex('#000000', '#ffffff', 0.5)).toBe('#808080');
  });
});

describe('toPastel', () => {
  it('returns a lighter version of the color', () => {
    const pastel = toPastel('#3b82f6');
    expect(hexLuminance(pastel)).toBeGreaterThan(hexLuminance('#3b82f6'));
  });
});

describe('paletteColor', () => {
  it('returns a color from the palette', () => {
    expect(PROJECT_COLORS).toContain(paletteColor(0));
  });

  it('wraps around the palette', () => {
    expect(paletteColor(0)).toBe(paletteColor(PROJECT_COLORS.length));
  });
});
