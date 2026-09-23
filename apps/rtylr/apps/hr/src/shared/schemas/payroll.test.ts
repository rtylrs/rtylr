import { describe, it, expect } from 'vitest';
import {
  payrollConfigSchema,
  payrollRunEntrySchema,
  payrollRunSchema,
  paySlipSearchSchema,
} from './payroll';

// ─── payroll Zod schemas - pure parse/validation tests ───────────────────────

describe('payrollConfigSchema', () => {
  it('parses the fields supported by the live payroll config contract', () => {
    const result = payrollConfigSchema.parse({
      period: 'biweekly',
      overtimeEnabled: true,
      payDayOfMonth: 15,
    });
    expect(result.period).toBe('biweekly');
    expect(result.overtimeEnabled).toBe(true);
    expect(result.payDayOfMonth).toBe(15);
  });

  it('accepts all valid period values', () => {
    for (const period of ['weekly', 'biweekly', 'monthly'] as const) {
      const r = payrollConfigSchema.parse({
        period,
        overtimeEnabled: true,
        payDayOfMonth: 28,
      });
      expect(r.period).toBe(period);
    }
  });

  it('rejects invalid period', () => {
    expect(() =>
      payrollConfigSchema.parse({ period: 'quarterly', overtimeEnabled: true, payDayOfMonth: 28 })
    ).toThrow();
  });

  it('rejects a missing pay day', () => {
    expect(() => payrollConfigSchema.parse({ period: 'weekly', overtimeEnabled: true })).toThrow();
  });
});

describe('payrollRunEntrySchema', () => {
  it('parses a complete payroll run entry', () => {
    const result = payrollRunEntrySchema.parse({
      id: 1,
      employeeId: 10,
      employeeName: 'Alice',
      hoursWorked: 80,
      basePay: 2000,
      overtimePay: 150,
      deductions: 200,
      bonuses: 100,
      netPay: 2050,
    });
    expect(result.id).toBe(1);
    expect(result.netPay).toBe(2050);
  });

  it('handles zero values for all numeric fields', () => {
    const result = payrollRunEntrySchema.parse({
      id: 2,
      employeeId: 11,
      employeeName: 'Bob',
      hoursWorked: 0,
      basePay: 0,
      overtimePay: 0,
      deductions: 0,
      bonuses: 0,
      netPay: 0,
    });
    expect(result.hoursWorked).toBe(0);
    expect(result.netPay).toBe(0);
  });

  it('handles fractional hours and pay amounts', () => {
    const result = payrollRunEntrySchema.parse({
      id: 3,
      employeeId: 12,
      employeeName: 'Carol',
      hoursWorked: 37.5,
      basePay: 937.5,
      overtimePay: 0,
      deductions: 93.75,
      bonuses: 50,
      netPay: 893.75,
    });
    expect(result.hoursWorked).toBe(37.5);
    expect(result.basePay).toBe(937.5);
  });

  it('rejects missing required fields', () => {
    expect(() =>
      payrollRunEntrySchema.parse({ id: 4, employeeId: 13, employeeName: 'Dave' })
    ).toThrow();
  });
});

describe('payrollRunSchema', () => {
  const validEntry = {
    id: 1,
    employeeId: 10,
    employeeName: 'Alice',
    hoursWorked: 80,
    basePay: 2000,
    overtimePay: 0,
    deductions: 200,
    bonuses: 0,
    netPay: 1800,
  };

  it('parses a valid payroll run', () => {
    const result = payrollRunSchema.parse({
      id: 100,
      periodStart: '2026-04-01',
      periodEnd: '2026-04-30',
      status: 'Draft',
      totalNetPay: 50000,
      totalBasePay: 55000,
      totalOvertimePay: 2000,
      totalDeductions: 7000,
      employeeCount: 25,
      createdAt: '2026-05-01T10:00:00',
      entries: [validEntry],
    });
    expect(result.id).toBe(100);
    expect(result.status).toBe('Draft');
    expect(result.entries).toHaveLength(1);
  });

  it('allows entries to be absent', () => {
    const result = payrollRunSchema.parse({
      id: 101,
      periodStart: '2026-05-01',
      periodEnd: '2026-05-31',
      status: 'Pending',
      totalNetPay: 0,
      totalBasePay: 0,
      totalOvertimePay: 0,
      totalDeductions: 0,
      employeeCount: 0,
      createdAt: '2026-06-01T08:00:00',
    });
    expect(result.entries).toBeUndefined();
  });

  it('accepts all valid status values', () => {
    for (const status of ['Draft', 'Pending', 'Approved', 'Paid'] as const) {
      const r = payrollRunSchema.parse({
        id: 1,
        periodStart: '2026-04-01',
        periodEnd: '2026-04-30',
        status,
        totalNetPay: 100,
        totalBasePay: 100,
        totalOvertimePay: 0,
        totalDeductions: 0,
        employeeCount: 1,
        createdAt: '2026-05-01T00:00:00',
      });
      expect(r.status).toBe(status);
    }
  });

  it('rejects invalid status', () => {
    expect(() =>
      payrollRunSchema.parse({
        id: 1,
        periodStart: '2026-04-01',
        periodEnd: '2026-04-30',
        status: 'Processing',
        totalNetPay: 100,
        totalBasePay: 100,
        totalOvertimePay: 0,
        totalDeductions: 0,
        employeeCount: 1,
        createdAt: '2026-05-01T00:00:00',
      })
    ).toThrow();
  });

  it('rejects missing period dates', () => {
    expect(() =>
      payrollRunSchema.parse({
        id: 1,
        status: 'Draft',
        totalNetPay: 0,
        totalBasePay: 0,
        totalOvertimePay: 0,
        totalDeductions: 0,
        employeeCount: 0,
        createdAt: '2026-05-01T00:00:00',
      })
    ).toThrow();
  });
});

// ─── paySlipSearchSchema - route search params ──────────────────────────────
//
// These cases exist to protect the fix in #44. `validateSearch` runs before the
// route renders, so a schema that throws (or that resolves a bad `runId` to a
// truthy value) would either trip the router error boundary or send the pay
// slip view back to always rendering empty. The reference is the manual
// `URLSearchParams` parsing this schema replaced.

describe('paySlipSearchSchema', () => {
  // The router's `defaultParseSearch` JSON-parses query values, so a numeric
  // `runId` reaches the schema as a number rather than a string.
  it('passes a valid run id through unchanged', () => {
    expect(paySlipSearchSchema.parse({ runId: 12 }).runId).toBe(12);
  });

  it.each([
    ['missing', {}],
    ['empty string', { runId: '' }],
    ['non-numeric', { runId: 'abc' }],
    ['null', { runId: null }],
    ['NaN', { runId: Number(undefined) }],
    ['zero', { runId: 0 }],
    ['negative', { runId: -3 }],
    ['fractional', { runId: 12.5 }],
  ])('resolves a %s run id to undefined instead of throwing', (_label, search) => {
    expect(() => paySlipSearchSchema.parse(search)).not.toThrow();
    expect(paySlipSearchSchema.parse(search).runId).toBeUndefined();
  });

  it('drops unknown search params', () => {
    expect(paySlipSearchSchema.parse({ runId: 7, employeeId: 99 })).toEqual({ runId: 7 });
  });
});
