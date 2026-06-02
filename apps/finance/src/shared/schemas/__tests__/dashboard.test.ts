import { describe, it, expect } from 'vitest';
import { dashboardStatsSchema } from '../dashboard';

describe('dashboardStatsSchema', () => {
  it('defaults all numeric fields to zero', () => {
    const result = dashboardStatsSchema.parse({});
    expect(result.totalReceivable).toBe(0);
    expect(result.overdueAmount).toBe(0);
    expect(result.totalExpensesMtd).toBe(0);
    expect(result.netProfitMtd).toBe(0);
    expect(result.outstandingInvoicesCount).toBe(0);
  });

  it('defaults array fields to empty arrays', () => {
    const result = dashboardStatsSchema.parse({});
    expect(result.recentTransactions).toEqual([]);
    expect(result.overdueInvoices).toEqual([]);
    expect(result.barData).toEqual([]);
  });

  it('parses recentTransactions and coerces transaction id', () => {
    const result = dashboardStatsSchema.parse({
      recentTransactions: [
        { id: 10, description: 'Sale', amount: '500', type: 'credit', date: '2026-05-01' },
      ],
    });
    expect(result.recentTransactions[0]?.id).toBe('10');
    expect(result.recentTransactions[0]?.type).toBe('credit');
  });

  it('defaults transaction type to credit when invalid', () => {
    const result = dashboardStatsSchema.parse({
      recentTransactions: [{ id: '1', description: 'X', amount: '0', type: 'unknown', date: '' }],
    });
    expect(result.recentTransactions[0]?.type).toBe('credit');
  });

  it('parses barData correctly', () => {
    const result = dashboardStatsSchema.parse({
      barData: [{ month: 'Jan', revenue: 3000, expenses: 1500 }],
    });
    expect(result.barData[0]?.month).toBe('Jan');
    expect(result.barData[0]?.revenue).toBe(3000);
  });

  it('accepts the deployed backend payload', () => {
    const result = dashboardStatsSchema.parse({
      totalReceivable: 5000,
      overdueAmount: 1200,
      totalExpensesMtd: 20000,
      netProfitMtd: 30000,
    });
    expect(result.netProfitMtd).toBe(30000);
    expect(result.totalReceivable).toBe(5000);
    expect(result.overdueAmount).toBe(1200);
  });
});
