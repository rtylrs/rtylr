import { describe, it, expect } from 'vitest';
import {
  conversionRate,
  dropOff,
  dropOffPct,
  funnelRates,
  overallConversion,
} from './conversionUtils';
import type { FunnelStep } from './conversionUtils';

describe('conversionRate', () => {
  it('returns the correct fraction', () => {
    expect(conversionRate(100, 75)).toBeCloseTo(0.75);
  });

  it('returns 0 when from is 0', () => {
    expect(conversionRate(0, 50)).toBe(0);
  });

  it('caps at 1 even when to > from', () => {
    expect(conversionRate(10, 15)).toBe(1);
  });
});

describe('dropOff', () => {
  it('returns the difference between from and to', () => {
    expect(dropOff(100, 75)).toBe(25);
  });

  it('returns 0 when to >= from', () => {
    expect(dropOff(50, 60)).toBe(0);
  });
});

describe('dropOffPct', () => {
  it('calculates drop-off percentage', () => {
    expect(dropOffPct(100, 80)).toBeCloseTo(20);
  });

  it('returns 0 when from is 0', () => {
    expect(dropOffPct(0, 0)).toBe(0);
  });
});

describe('funnelRates', () => {
  const steps: FunnelStep[] = [
    { name: 'Visitors', count: 1000 },
    { name: 'Signups', count: 500 },
    { name: 'Purchases', count: 100 },
  ];

  it('returns enriched steps with rate and dropOffPct', () => {
    const result = funnelRates(steps);
    expect(result).toHaveLength(3);
    expect(result[1].rate).toBeCloseTo(0.5);
    expect(result[2].dropOffPct).toBeCloseTo(80);
  });

  it('first step has rate 1 (baseline)', () => {
    expect(funnelRates(steps)[0].rate).toBe(1);
  });
});

describe('overallConversion', () => {
  it('returns rate from first to last step', () => {
    const steps: FunnelStep[] = [
      { name: 'A', count: 1000 },
      { name: 'B', count: 500 },
      { name: 'C', count: 100 },
    ];
    expect(overallConversion(steps)).toBeCloseTo(0.1);
  });

  it('returns 0 for empty array', () => {
    expect(overallConversion([])).toBe(0);
  });

  it('returns 1 for single step', () => {
    expect(overallConversion([{ name: 'A', count: 100 }])).toBe(1);
  });
});
