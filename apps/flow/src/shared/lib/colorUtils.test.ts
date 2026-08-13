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
  accentTints,
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

describe('accentTints', () => {
  it('keeps the accent fill itself untouched', () => {
    expect(accentTints('#64c5b9', 'dark')['--accent']).toBe('#64c5b9');
    expect(accentTints('#64c5b9', 'light')['--accent']).toBe('#64c5b9');
  });

  it('washes the muted tint heavier in light mode, where .12 reads lighter', () => {
    expect(accentTints('#64c5b9', 'dark')['--accent-muted']).toBe('rgba(100, 197, 185, 0.12)');
    expect(accentTints('#64c5b9', 'light')['--accent-muted']).toBe('rgba(100, 197, 185, 0.18)');
  });

  it('strengthens the border and focus ring in light mode too', () => {
    const dark = accentTints('#64c5b9', 'dark');
    const light = accentTints('#64c5b9', 'light');

    expect(dark['--accent-border']).toBe('rgba(100, 197, 185, 0.3)');
    expect(light['--accent-border']).toBe('rgba(100, 197, 185, 0.42)');
    expect(dark['--accent-focus']).toBe('rgba(100, 197, 185, 0.5)');
    expect(light['--accent-focus']).toBe('rgba(100, 197, 185, 0.5)');
  });

  it('accepts the uppercase hex the workspace colour picker emits', () => {
    // FlowSettings' WORKSPACE_COLORS are uppercase, e.g. '#64C5B9'.
    expect(accentTints('#3B82F6', 'dark')['--accent-muted']).toBe('rgba(59, 130, 246, 0.12)');
    expect(accentTints('#3B82F6', 'dark')['--accent']).toBe('#3b82f6');
  });

  it('treats a three-digit hex as invalid, matching isValidHex', () => {
    expect(accentTints('#0f8', 'dark')['--accent']).toBe('#64c5b9');
  });

  it('falls back to the brand teal rather than emitting rgba(NaN)', () => {
    expect(accentTints('not-a-colour', 'dark')['--accent-muted']).toBe('rgba(100, 197, 185, 0.12)');
    expect(accentTints('', 'light')['--accent']).toBe('#64c5b9');
  });
});
