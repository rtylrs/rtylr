import { describe, it, expect } from 'vitest';
import {
  convertWeight,
  convertVolume,
  convertLength,
  roundTo,
  formatWeight,
  formatVolume,
  unitLabel,
  needsUnit,
} from './unitConversionUtils';

describe('convertWeight', () => {
  it('converts kg to grams', () => {
    expect(convertWeight(1, 'kg', 'g')).toBe(1000);
  });

  it('converts grams to kg', () => {
    expect(convertWeight(1000, 'g', 'kg')).toBe(1);
  });

  it('converts lb to kg', () => {
    expect(roundTo(convertWeight(1, 'lb', 'kg'), 3)).toBe(0.454);
  });

  it('identity conversion returns same value', () => {
    expect(convertWeight(5, 'kg', 'kg')).toBe(5);
  });
});

describe('convertVolume', () => {
  it('converts liters to milliliters', () => {
    expect(convertVolume(1, 'l', 'ml')).toBe(1000);
  });

  it('converts ml to liters', () => {
    expect(convertVolume(1000, 'ml', 'l')).toBe(1);
  });
});

describe('convertLength', () => {
  it('converts meters to centimeters', () => {
    expect(convertLength(1, 'm', 'cm')).toBe(100);
  });

  it('converts inches to centimeters', () => {
    expect(roundTo(convertLength(1, 'in', 'cm'), 2)).toBe(2.54);
  });
});

describe('roundTo', () => {
  it('rounds to specified decimal places', () => {
    expect(roundTo(3.14159, 2)).toBe(3.14);
    expect(roundTo(3.145, 2)).toBe(3.15);
  });

  it('rounds to 0 decimals', () => {
    expect(roundTo(3.7, 0)).toBe(4);
  });
});

describe('formatWeight', () => {
  it('formats weight with unit', () => {
    expect(formatWeight(1.5, 'kg')).toBe('1.5 kg');
    expect(formatWeight(500, 'g')).toBe('500 g');
  });

  it('respects decimal places', () => {
    expect(formatWeight(1.2345, 'kg', 2)).toBe('1.23 kg');
  });
});

describe('formatVolume', () => {
  it('formats volume with unit', () => {
    expect(formatVolume(1.5, 'l')).toBe('1.5 l');
  });

  it('replaces underscore in fl_oz', () => {
    expect(formatVolume(1, 'fl_oz')).toBe('1 fl oz');
  });
});

describe('unitLabel', () => {
  it('returns readable label for known units', () => {
    expect(unitLabel('kg')).toBe('Kilograms');
    expect(unitLabel('l')).toBe('Liters');
    expect(unitLabel('in')).toBe('Inches');
  });

  it('returns unit string for unknown units', () => {
    expect(unitLabel('pcs' as never)).toBe('pcs');
  });
});

describe('needsUnit', () => {
  it('returns true for non-pcs unit', () => {
    expect(needsUnit(5, 'kg')).toBe(true);
  });

  it('returns false for pcs unit', () => {
    expect(needsUnit(5, 'pcs')).toBe(false);
  });

  it('returns false when unit is undefined', () => {
    expect(needsUnit(5, undefined)).toBe(false);
  });
});
