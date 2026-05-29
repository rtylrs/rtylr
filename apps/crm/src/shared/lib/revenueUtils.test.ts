import { describe, it, expect } from 'vitest';
import {
  parseRevenue,
  formatRevenue,
  percentChange,
  sumRevenue,
  arpc,
  formatChange,
} from './revenueUtils';

describe('parseRevenue', () => {
  it('parses a USD string with $ and commas', () => {
    expect(parseRevenue('$1,234.56')).toBeCloseTo(1234.56);
  });

  it('returns 0 for invalid input', () => {
    expect(parseRevenue('N/A')).toBe(0);
  });
});

describe('formatRevenue', () => {
  it('formats values under 1000 with 2 decimal places', () => {
    expect(formatRevenue(250)).toBe('$250.00');
  });

  it('formats thousands as K', () => {
    expect(formatRevenue(1500)).toBe('$1.5K');
  });

  it('formats millions as M', () => {
    expect(formatRevenue(2_500_000)).toBe('$2.5M');
  });
});

describe('percentChange', () => {
  it('calculates positive change', () => {
    expect(percentChange(100, 120)).toBeCloseTo(20);
  });

  it('calculates negative change', () => {
    expect(percentChange(100, 80)).toBeCloseTo(-20);
  });

  it('returns 0 when both are 0', () => {
    expect(percentChange(0, 0)).toBe(0);
  });
});

describe('sumRevenue', () => {
  it('sums multiple revenue strings', () => {
    expect(sumRevenue(['$100', '$200.50', '$50']).toFixed(2)).toBe('350.50');
  });
});

describe('arpc', () => {
  it('calculates average revenue per customer', () => {
    expect(arpc(1000, 10)).toBe(100);
  });

  it('returns 0 when customer count is 0', () => {
    expect(arpc(500, 0)).toBe(0);
  });
});

describe('formatChange', () => {
  it('formats positive change with + prefix', () => {
    expect(formatChange(12.5)).toBe('+12.5%');
  });

  it('formats negative change', () => {
    expect(formatChange(-5.3)).toBe('-5.3%');
  });
});
