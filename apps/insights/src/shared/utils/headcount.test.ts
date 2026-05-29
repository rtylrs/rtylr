import { describe, it, expect } from 'vitest';
import {
  headcountTrend,
  headcountDeltaLabel,
  headcountGrowthRate,
  previousHeadcount,
} from './headcount';
import type { HeadcountResponse } from '../hooks/queries/types';

function makeHeadcount(overrides: Partial<HeadcountResponse> = {}): HeadcountResponse {
  return { activeStaff: 50, momDelta: 5, momDeltaPct: 11.1, ...overrides };
}

describe('headcountTrend', () => {
  it('returns up for positive delta', () => {
    expect(headcountTrend(makeHeadcount({ momDeltaPct: 5 }))).toBe('up');
  });

  it('returns down for negative delta', () => {
    expect(headcountTrend(makeHeadcount({ momDeltaPct: -3 }))).toBe('down');
  });

  it('returns flat for zero delta', () => {
    expect(headcountTrend(makeHeadcount({ momDeltaPct: 0 }))).toBe('flat');
  });
});

describe('headcountDeltaLabel', () => {
  it('includes + for positive delta', () => {
    expect(headcountDeltaLabel(makeHeadcount({ momDelta: 3 }))).toBe('+3 this month');
  });

  it('includes - for negative delta', () => {
    expect(headcountDeltaLabel(makeHeadcount({ momDelta: -2 }))).toBe('-2 this month');
  });

  it('handles zero delta', () => {
    expect(headcountDeltaLabel(makeHeadcount({ momDelta: 0 }))).toBe('+0 this month');
  });
});

describe('headcountGrowthRate', () => {
  it('formats positive growth', () => {
    expect(headcountGrowthRate(makeHeadcount({ momDeltaPct: 11.1 }))).toBe('+11.1%');
  });

  it('formats negative growth', () => {
    expect(headcountGrowthRate(makeHeadcount({ momDeltaPct: -5.5 }))).toBe('-5.5%');
  });
});

describe('previousHeadcount', () => {
  it('subtracts delta from active staff', () => {
    expect(previousHeadcount(makeHeadcount({ activeStaff: 50, momDelta: 5 }))).toBe(45);
  });

  it('does not return negative values', () => {
    expect(previousHeadcount(makeHeadcount({ activeStaff: 3, momDelta: 10 }))).toBe(0);
  });
});
