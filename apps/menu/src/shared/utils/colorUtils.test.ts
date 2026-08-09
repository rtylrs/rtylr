import { describe, it, expect } from 'vitest';
import { MENU_COLORS, getMenuColor, hexToRgba, isValidHex } from './colorUtils';

describe('MENU_COLORS', () => {
  it('contains an accent color', () => {
    expect(MENU_COLORS.accent).toMatch(/^#/);
  });

  it('contains a primary color', () => {
    expect(MENU_COLORS.primary).toMatch(/^#/);
  });
});

describe('getMenuColor', () => {
  it('returns the correct hex for a known key', () => {
    expect(getMenuColor('accent')).toBe(MENU_COLORS.accent);
  });

  it('returns the primary color for the primary key', () => {
    expect(getMenuColor('primary')).toBe(MENU_COLORS.primary);
  });
});

describe('hexToRgba', () => {
  it('converts a 6-digit hex to rgba', () => {
    expect(hexToRgba('#ffffff', 1)).toBe('rgba(255, 255, 255, 1)');
  });

  it('converts a 6-digit hex without # to rgba', () => {
    expect(hexToRgba('000000', 0.5)).toBe('rgba(0, 0, 0, 0.5)');
  });

  it('converts a 3-digit hex to rgba', () => {
    expect(hexToRgba('#fff', 1)).toBe('rgba(255, 255, 255, 1)');
  });

  it('handles semi-transparent alpha values', () => {
    const result = hexToRgba('#03AED2', 0.4);
    expect(result).toMatch(/^rgba\(/);
    expect(result).toContain('0.4');
  });
});

describe('isValidHex', () => {
  it('returns true for a valid 6-digit hex', () => {
    expect(isValidHex('#03AED2')).toBe(true);
  });

  it('returns true for a valid 3-digit hex', () => {
    expect(isValidHex('#fff')).toBe(true);
  });

  it('returns false for a hex without #', () => {
    expect(isValidHex('ffffff')).toBe(false);
  });

  it('returns false for an invalid hex', () => {
    expect(isValidHex('#xyz123')).toBe(false);
  });

  it('returns false for an empty string', () => {
    expect(isValidHex('')).toBe(false);
  });
});
