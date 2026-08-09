import { describe, it, expect } from 'vitest';
import {
  grossMarginPct,
  netMarginPct,
  operatingExpenseRatio,
  returnOnAssets,
  debtToEquityRatio,
  currentRatio,
  yoyGrowthPct,
  ebitda,
  quickRatio,
  workingCapital,
} from './reportUtils';

describe('grossMarginPct', () => {
  it('computes gross margin percentage', () => {
    expect(grossMarginPct(100000, 40000)).toBeCloseTo(60);
  });
  it('returns 0 when revenue is 0', () => {
    expect(grossMarginPct(0, 0)).toBe(0);
  });
});
describe('netMarginPct', () => {
  it('computes net margin percentage', () => {
    expect(netMarginPct(100000, 25000)).toBeCloseTo(25);
  });
  it('returns 0 when revenue is 0', () => {
    expect(netMarginPct(0, 0)).toBe(0);
  });
});
describe('operatingExpenseRatio', () => {
  it('computes opex ratio', () => {
    expect(operatingExpenseRatio(20000, 100000)).toBeCloseTo(20);
  });
  it('returns 0 when revenue is 0', () => {
    expect(operatingExpenseRatio(5000, 0)).toBe(0);
  });
});
describe('returnOnAssets', () => {
  it('computes ROA percentage', () => {
    expect(returnOnAssets(10000, 200000)).toBeCloseTo(5);
  });
});
describe('debtToEquityRatio', () => {
  it('computes debt to equity', () => {
    expect(debtToEquityRatio(50000, 100000)).toBeCloseTo(0.5);
  });
});
describe('currentRatio', () => {
  it('computes current ratio', () => {
    expect(currentRatio(60000, 30000)).toBeCloseTo(2);
  });
});
describe('yoyGrowthPct', () => {
  it('computes year over year growth', () => {
    expect(yoyGrowthPct(110000, 100000)).toBeCloseTo(10);
  });
});
describe('ebitda', () => {
  it('sums earnings components', () => {
    expect(ebitda(100, 20, 30, 15, 5)).toBe(170);
  });
});
describe('quickRatio', () => {
  it('computes quick ratio', () => {
    expect(quickRatio(10000, 5000, 8000, 15000)).toBeCloseTo(1.533, 2);
  });
  it('returns 0 for zero liabilities', () => {
    expect(quickRatio(10000, 5000, 8000, 0)).toBe(0);
  });
});
describe('workingCapital', () => {
  it('computes working capital', () => {
    expect(workingCapital(50000, 30000)).toBe(20000);
  });
  it('returns negative when liabilities exceed assets', () => {
    expect(workingCapital(10000, 15000)).toBe(-5000);
  });
});
