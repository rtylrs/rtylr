import { describe, it, expect } from 'vitest';
import {
  netCashFlow,
  peakInflowDay,
  topExpenseCategories,
  totalExpenses,
  overdueInvoiceTotal,
  largestAgingBucket,
} from './financial';
import type { CashFlowPoint, ExpenseByCategory, InvoiceAgingBucket } from '../hooks/queries/types';

function makeCashFlow(date: string, inflow: number, outflow: number): CashFlowPoint {
  return { date, inflowMinorUnits: inflow, outflowMinorUnits: outflow };
}

function makeExpense(name: string, amount: number): ExpenseByCategory {
  return { categoryId: 1, categoryName: name, amountMinorUnits: amount };
}

function makeBucket(bucket: string, amount: number): InvoiceAgingBucket {
  return { bucket, invoiceCount: 1, amountMinorUnits: amount };
}

describe('netCashFlow', () => {
  it('sums inflow minus outflow', () => {
    const series = [
      makeCashFlow('2026-01-01', 10000, 4000),
      makeCashFlow('2026-01-02', 5000, 2000),
    ];
    expect(netCashFlow(series)).toBe(9000);
  });

  it('returns 0 for empty series', () => {
    expect(netCashFlow([])).toBe(0);
  });
});

describe('peakInflowDay', () => {
  it('returns point with highest inflow', () => {
    const series = [makeCashFlow('2026-01-01', 100, 0), makeCashFlow('2026-01-02', 500, 0)];
    expect(peakInflowDay(series)?.date).toBe('2026-01-02');
  });

  it('returns undefined for empty series', () => {
    expect(peakInflowDay([])).toBeUndefined();
  });
});

describe('topExpenseCategories', () => {
  it('returns top n by amount descending', () => {
    const cats = [makeExpense('A', 100), makeExpense('B', 500), makeExpense('C', 300)];
    const top = topExpenseCategories(cats, 2);
    expect(top[0].categoryName).toBe('B');
    expect(top).toHaveLength(2);
  });
});

describe('totalExpenses', () => {
  it('sums amountMinorUnits', () => {
    const cats = [makeExpense('A', 200), makeExpense('B', 300)];
    expect(totalExpenses(cats)).toBe(500);
  });

  it('returns 0 for empty input', () => {
    expect(totalExpenses([])).toBe(0);
  });
});

describe('overdueInvoiceTotal', () => {
  it('excludes 0-30 day bucket', () => {
    const buckets = [makeBucket('0-30 days', 1000), makeBucket('31-60 days', 500)];
    expect(overdueInvoiceTotal(buckets)).toBe(500);
  });
});

describe('largestAgingBucket', () => {
  it('returns bucket with highest amountMinorUnits', () => {
    const buckets = [makeBucket('0-30 days', 200), makeBucket('31-60 days', 800)];
    expect(largestAgingBucket(buckets)?.bucket).toBe('31-60 days');
  });

  it('returns undefined for empty input', () => {
    expect(largestAgingBucket([])).toBeUndefined();
  });
});
