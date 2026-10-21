import { describe, it, expect } from 'vitest';
import { dashboardStatsSchema } from '../dashboard';

describe('dashboardStatsSchema', () => {
  it('defaults every figure to zero and every list to empty', () => {
    const result = dashboardStatsSchema.parse({});
    expect(result.totalReceivable).toBe(0);
    expect(result.netProfitMtd).toBe(0);
    expect(result.pendingExpenseCount).toBe(0);
    expect(result.monthly).toEqual([]);
    expect(result.recentActivity).toEqual([]);
    expect(result.overdueInvoices).toEqual([]);
  });

  it('reads int64 counts and ids that arrive as strings', () => {
    const result = dashboardStatsSchema.parse({
      pendingExpenses: 35,
      pendingExpenseCount: '1',
      overdueInvoiceCount: '2',
      recentActivity: [
        {
          kind: 'expense',
          id: '7',
          description: 'Phone',
          amount: 35,
          date: '2026-10-19',
          status: 'pending',
        },
      ],
      overdueInvoices: [
        { id: '4', invoiceNumber: 'INV-4', clientName: 'Acme', amountDue: 300, daysOverdue: '12' },
      ],
    });
    expect(result.pendingExpenseCount).toBe(1);
    expect(result.overdueInvoiceCount).toBe(2);
    expect(result.recentActivity[0]).toMatchObject({ id: '7', kind: 'expense', amount: 35 });
    expect(result.overdueInvoices[0]?.daysOverdue).toBe(12);
  });

  it('treats an unknown activity kind as an expense and junk numbers as zero', () => {
    const result = dashboardStatsSchema.parse({
      totalRevenueMtd: 'NaN',
      recentActivity: [{ kind: 'refund', id: 1, amount: 'x' }],
    });
    expect(result.totalRevenueMtd).toBe(0);
    expect(result.recentActivity[0]).toMatchObject({ kind: 'expense', id: '1', amount: 0 });
  });

  it('keeps the monthly series in order', () => {
    const result = dashboardStatsSchema.parse({
      monthly: [
        { month: '2026-09', revenue: 3000, expenses: 1500 },
        { month: '2026-10', revenue: 0, expenses: 35 },
      ],
    });
    expect(result.monthly.map((m) => m.month)).toEqual(['2026-09', '2026-10']);
    expect(result.monthly[1]?.expenses).toBe(35);
  });
});
