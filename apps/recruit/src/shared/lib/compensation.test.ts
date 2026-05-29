import { describe, it, expect } from 'vitest';
import {
  bandRange,
  compaRatio,
  isWithinBand,
  bandPositionPct,
  compensationTypeLabel,
  formatSalary,
} from './compensation';
import type { CompensationBand } from './compensation';

function makeBand(minSalary: number, midSalary: number, maxSalary: number): CompensationBand {
  return { role: 'Engineer', level: 'L3', minSalary, midSalary, maxSalary, currency: 'USD' };
}

const band = makeBand(80000, 100000, 120000);

describe('bandRange', () => {
  it('returns spread between max and min', () => {
    expect(bandRange(band)).toBe(40000);
  });
});

describe('compaRatio', () => {
  it('returns ratio to midpoint', () => {
    expect(compaRatio(100000, band)).toBe(1);
    expect(compaRatio(80000, band)).toBeCloseTo(0.8, 2);
  });

  it('returns 0 when midpoint is 0', () => {
    expect(compaRatio(50000, makeBand(0, 0, 100000))).toBe(0);
  });
});

describe('isWithinBand', () => {
  it('returns true when salary is within band', () => {
    expect(isWithinBand(95000, band)).toBe(true);
  });

  it('returns false when below min', () => {
    expect(isWithinBand(70000, band)).toBe(false);
  });

  it('returns false when above max', () => {
    expect(isWithinBand(130000, band)).toBe(false);
  });
});

describe('bandPositionPct', () => {
  it('returns 0 for min salary', () => {
    expect(bandPositionPct(80000, band)).toBe(0);
  });

  it('returns 100 for max salary', () => {
    expect(bandPositionPct(120000, band)).toBe(100);
  });

  it('returns 50 for midpoint', () => {
    expect(bandPositionPct(100000, band)).toBe(50);
  });
});

describe('compensationTypeLabel', () => {
  it('returns display label for known types', () => {
    expect(compensationTypeLabel('salary')).toBe('Annual Salary');
    expect(compensationTypeLabel('hourly')).toBe('Hourly Rate');
  });

  it('returns raw value for unknown type', () => {
    expect(compensationTypeLabel('equity')).toBe('equity');
  });
});

describe('formatSalary', () => {
  it('formats large salaries with K notation', () => {
    expect(formatSalary(95000)).toBe('USD 95K');
  });

  it('formats small salaries without K', () => {
    expect(formatSalary(500)).toBe('USD 500');
  });

  it('uses provided currency', () => {
    expect(formatSalary(80000, 'GBP')).toBe('GBP 80K');
  });
});
