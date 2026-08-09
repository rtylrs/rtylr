import { describe, it, expect } from 'vitest';
import {
  weightedAverageCost,
  fifoCost,
  lifoCost,
  totalValuation,
  unitCostForMethod,
  grossProfit,
  grossProfitMarginPct,
} from './valuationUtils';
import type { CostLayer } from './valuationUtils';

const layers: CostLayer[] = [
  { quantity: 10, unitCost: 5, date: '2026-01-01' },
  { quantity: 20, unitCost: 8, date: '2026-02-01' },
  { quantity: 5, unitCost: 10, date: '2026-03-01' },
];

describe('weightedAverageCost', () => {
  it('computes weighted average correctly', () => {
    // (10*5 + 20*8 + 5*10) / 35 = (50 + 160 + 50) / 35 = 260/35 = ~7.43
    const wac = weightedAverageCost(layers);
    expect(wac).toBeCloseTo(7.43, 1);
  });

  it('returns 0 for empty layers', () => {
    expect(weightedAverageCost([])).toBe(0);
  });
});

describe('fifoCost', () => {
  it('uses oldest layers first', () => {
    // Remove 15: first 10 at $5 + 5 at $8 = 50 + 40 = 90
    expect(fifoCost(layers, 15)).toBe(90);
  });

  it('computes cost for exact layer quantity', () => {
    expect(fifoCost(layers, 10)).toBe(50);
  });
});

describe('lifoCost', () => {
  it('uses newest layers first', () => {
    // Remove 10: first 5 at $10 + 5 at $8 = 50 + 40 = 90
    expect(lifoCost(layers, 10)).toBe(90);
  });
});

describe('totalValuation', () => {
  it('uses weighted average for WAC', () => {
    const total = totalValuation(layers, 'WAC');
    // WAC * total qty = 7.43 * 35 ~ 260
    expect(total).toBeCloseTo(260, 0);
  });

  it('uses FIFO for FIFO method', () => {
    // All layers at FIFO cost = same as total purchase cost = 50+160+50 = 260
    expect(totalValuation(layers, 'FIFO')).toBe(260);
  });

  it('uses LIFO for LIFO method', () => {
    // LIFO all 35: 5@10 + 20@8 + 10@5 = 50+160+50 = 260 (all layers exhausted, same total)
    expect(totalValuation(layers, 'LIFO')).toBe(260);
  });
});

describe('unitCostForMethod', () => {
  it('returns 0 for empty layers', () => {
    expect(unitCostForMethod([], 'WAC')).toBe(0);
  });

  it('returns per-unit cost', () => {
    const cost = unitCostForMethod(layers, 'WAC');
    expect(cost).toBeGreaterThan(0);
  });
});

describe('grossProfit', () => {
  it('computes gross profit', () => {
    expect(grossProfit(1000, 600)).toBe(400);
  });

  it('returns negative for a loss', () => {
    expect(grossProfit(500, 700)).toBe(-200);
  });
});

describe('grossProfitMarginPct', () => {
  it('computes margin percentage', () => {
    expect(grossProfitMarginPct(1000, 600)).toBe(40);
  });

  it('returns 0 when revenue is 0', () => {
    expect(grossProfitMarginPct(0, 100)).toBe(0);
  });
});
