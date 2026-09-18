import { describe, it, expect } from 'vitest';
import {
  kpiChange,
  kpiChangePct,
  kpiChangeDirection,
  formatKpiValue,
  kpiChangeBadge,
} from './kpiCard';
import type { KpiCardData } from './kpiCard';

function makeKpi(
  value: number,
  previousValue: number,
  prefix?: string,
  unit?: string
): KpiCardData {
  return { title: 'Revenue', value, previousValue, prefix, unit };
}

describe('kpiChange', () => {
  it('returns absolute change', () => {
    expect(kpiChange(makeKpi(120, 100))).toBe(20);
  });

  it('returns negative for decrease', () => {
    expect(kpiChange(makeKpi(80, 100))).toBe(-20);
  });
});

describe('kpiChangePct', () => {
  it('returns percentage change', () => {
    expect(kpiChangePct(makeKpi(120, 100))).toBe(20);
  });

  it('returns 0 when previous is 0', () => {
    expect(kpiChangePct(makeKpi(50, 0))).toBe(0);
  });
});

describe('kpiChangeDirection', () => {
  it('returns increase when value is higher', () => {
    expect(kpiChangeDirection(makeKpi(120, 100))).toBe('increase');
  });

  it('returns decrease when value is lower', () => {
    expect(kpiChangeDirection(makeKpi(80, 100))).toBe('decrease');
  });

  it('returns neutral when unchanged', () => {
    expect(kpiChangeDirection(makeKpi(100, 100))).toBe('neutral');
  });
});

describe('formatKpiValue', () => {
  it('formats large values with K notation', () => {
    expect(formatKpiValue(makeKpi(5000, 4000, '$'))).toBe('$5K');
  });

  it('formats small values without K', () => {
    expect(formatKpiValue(makeKpi(42, 30))).toBe('42');
  });

  it('includes unit when provided', () => {
    expect(formatKpiValue(makeKpi(95, 90, '', '%'))).toBe('95%');
  });
});

describe('kpiChangeBadge', () => {
  it('shows plus sign for positive changes', () => {
    expect(kpiChangeBadge(makeKpi(120, 100))).toBe('+20%');
  });

  it('shows minus sign for negative changes', () => {
    expect(kpiChangeBadge(makeKpi(80, 100))).toBe('-20%');
  });
});
