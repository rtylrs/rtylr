import { describe, it, expect } from 'vitest';
import { deltaLabel, deltaDirection, kpiColor, formatKpiValue } from './kpi';

describe('deltaLabel', () => {
  it('formats a positive delta with a + sign', () => {
    expect(deltaLabel(12.3)).toBe('+12.3%');
  });

  it('formats a negative delta with a - sign', () => {
    expect(deltaLabel(-5)).toBe('-5.0%');
  });

  it('formats zero without a sign', () => {
    expect(deltaLabel(0)).toBe('0.0%');
  });

  it('uses custom decimal places', () => {
    expect(deltaLabel(10, 2)).toBe('+10.00%');
  });
});

describe('deltaDirection', () => {
  it('returns up for delta above the threshold', () => {
    expect(deltaDirection(2)).toBe('up');
  });

  it('returns down for delta below the negative threshold', () => {
    expect(deltaDirection(-2)).toBe('down');
  });

  it('returns flat for delta within the threshold', () => {
    expect(deltaDirection(0.3)).toBe('flat');
  });

  it('uses custom threshold', () => {
    expect(deltaDirection(3, 5)).toBe('flat');
    expect(deltaDirection(6, 5)).toBe('up');
  });
});

describe('kpiColor', () => {
  it('returns green when up and higher is better', () => {
    expect(kpiColor(5, true)).toBe('green');
  });

  it('returns red when down and higher is better', () => {
    expect(kpiColor(-5, true)).toBe('red');
  });

  it('returns red when up and lower is better', () => {
    expect(kpiColor(5, false)).toBe('red');
  });

  it('returns green when down and lower is better', () => {
    expect(kpiColor(-5, false)).toBe('green');
  });

  it('returns gray for flat delta', () => {
    expect(kpiColor(0.1)).toBe('gray');
  });
});

describe('formatKpiValue', () => {
  it('formats a plain number with locale separators', () => {
    expect(formatKpiValue(1000)).toBe('1,000');
  });

  it('formats minor units as currency', () => {
    const result = formatKpiValue(100000, { minorUnits: true, currency: 'USD' });
    expect(result).toContain('$');
    expect(result).toContain('1,000');
  });

  it('returns zero for zero input', () => {
    expect(formatKpiValue(0)).toBe('0');
  });
});
